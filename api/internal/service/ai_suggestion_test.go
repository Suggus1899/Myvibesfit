package service

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/google/uuid"

	"myvibesfit/api/internal/domain"
)

// El AuditLogger nunca propaga sus propios errores, pero si desreferencia su
// repo: un nil aqui revienta el test por el motivo equivocado.
type fakeAudit struct{ entries []domain.AuditEntry }

func (f *fakeAudit) Insert(_ context.Context, e domain.AuditEntry) error {
	f.entries = append(f.entries, e)
	return nil
}

type fakeSuggestions struct {
	domain.AISuggestionRepository

	sug     domain.AISuggestion
	applied []uuid.UUID
	status  domain.SuggestionStatus
}

func (f *fakeSuggestions) Review(_ context.Context, _, _ uuid.UUID, status domain.SuggestionStatus, _ uuid.UUID) (domain.AISuggestion, error) {
	f.status = status
	f.sug.Status = status
	return f.sug, nil
}

func (f *fakeSuggestions) MarkApplied(_ context.Context, id, _ uuid.UUID) error {
	f.applied = append(f.applied, id)
	return nil
}

type suggestionFixture struct {
	svc         *AISuggestionService
	suggestions *fakeSuggestions
	assignments *fakeAssignments
	nextID      uuid.UUID
}

func newSuggestionFixture(kind string, deltaPercent float64, exerciseID uuid.UUID, next domain.AssignedExercise) suggestionFixture {
	assignmentID := uuid.New()
	payload, _ := json.Marshal(domain.SuggestionPayload{
		TargetExerciseID: exerciseID.String(), DeltaPercent: deltaPercent,
	})

	suggestions := &fakeSuggestions{sug: domain.AISuggestion{
		ID: uuid.New(), AssignmentID: &assignmentID, Kind: kind, Payload: payload,
	}}
	assignments := &fakeAssignments{assignmentID: assignmentID, next: &next}

	repos := domain.TxRepos{Assignments: assignments, Suggestions: suggestions}
	svc := NewAISuggestionService(suggestions, &fakeUoW{repos: repos}, NewAuditLogger(&fakeAudit{}))

	return suggestionFixture{svc: svc, suggestions: suggestions, assignments: assignments, nextID: next.ID}
}

func TestReviewAplicaLoadAdjust(t *testing.T) {
	exerciseID := uuid.New()
	next := domain.AssignedExercise{
		ID: uuid.New(), ExerciseID: exerciseID,
		TargetWeightKg: ptrFloat(100), TargetRepsMin: ptrInt(5), TargetRepsMax: ptrInt(8),
	}
	f := newSuggestionFixture("load_adjust", -5, exerciseID, next)

	if _, err := f.svc.Review(context.Background(), uuid.New(), uuid.New(), uuid.New(), true); err != nil {
		t.Fatalf("review: %v", err)
	}

	if len(f.assignments.targetWrites) != 1 {
		t.Fatalf("esperaba un write, fueron %d", len(f.assignments.targetWrites))
	}
	w := f.assignments.targetWrites[0]
	if w.weightKg == nil || *w.weightKg != 95 {
		t.Fatalf("-5%% sobre 100 kg deberia dar 95, fue %v", w.weightKg)
	}
	if w.overrideSource == nil || *w.overrideSource != domain.OverrideSourceAISuggestion {
		t.Fatal("el target debe quedar marcado para que la progresion lo respete una vez")
	}
	if len(f.suggestions.applied) != 1 {
		t.Fatal("la sugerencia deberia quedar sellada como aplicada")
	}
}

func TestReviewVolumeAdjustEscalaLasReps(t *testing.T) {
	exerciseID := uuid.New()
	next := domain.AssignedExercise{
		ID: uuid.New(), ExerciseID: exerciseID,
		TargetWeightKg: ptrFloat(60), TargetRepsMin: ptrInt(8), TargetRepsMax: ptrInt(12),
	}
	f := newSuggestionFixture("volume_adjust", 25, exerciseID, next)

	if _, err := f.svc.Review(context.Background(), uuid.New(), uuid.New(), uuid.New(), true); err != nil {
		t.Fatalf("review: %v", err)
	}

	w := f.assignments.targetWrites[0]
	if w.repsMin == nil || *w.repsMin != 10 {
		t.Fatalf("+25%% sobre 8 reps deberia dar 10, fue %v", w.repsMin)
	}
	if w.repsMax == nil || *w.repsMax != 15 {
		t.Fatalf("+25%% sobre 12 reps deberia dar 15, fue %v", w.repsMax)
	}
	if w.weightKg == nil || *w.weightKg != 60 {
		t.Fatalf("volume_adjust no deberia tocar el peso, fue %v", w.weightKg)
	}
}

func TestReviewRechazarNoAplicaNada(t *testing.T) {
	exerciseID := uuid.New()
	next := domain.AssignedExercise{ID: uuid.New(), ExerciseID: exerciseID, TargetWeightKg: ptrFloat(100)}
	f := newSuggestionFixture("load_adjust", 10, exerciseID, next)

	sug, err := f.svc.Review(context.Background(), uuid.New(), uuid.New(), uuid.New(), false)
	if err != nil {
		t.Fatalf("review: %v", err)
	}
	if sug.Status != domain.SuggestionStatusRejected {
		t.Fatalf("esperaba rejected, fue %v", sug.Status)
	}
	if len(f.assignments.targetWrites) != 0 {
		t.Fatal("rechazar no deberia tocar el plan")
	}
	if len(f.suggestions.applied) != 0 {
		t.Fatal("rechazar no deberia sellar applied_at")
	}
}

func TestReviewKindNoAplicableSoloCambiaEstado(t *testing.T) {
	exerciseID := uuid.New()
	next := domain.AssignedExercise{ID: uuid.New(), ExerciseID: exerciseID, TargetWeightKg: ptrFloat(100)}

	// exercise_swap viaja con un nombre en texto libre: aprobarlo registra la
	// decision pero no muta el plan (decision de producto pendiente).
	for _, kind := range []string{"exercise_swap", "deload", "rest_day", "habit_nudge"} {
		f := newSuggestionFixture(kind, 10, exerciseID, next)
		sug, err := f.svc.Review(context.Background(), uuid.New(), uuid.New(), uuid.New(), true)
		if err != nil {
			t.Fatalf("%s: %v", kind, err)
		}
		if sug.Status != domain.SuggestionStatusApproved {
			t.Fatalf("%s: esperaba approved, fue %v", kind, sug.Status)
		}
		if len(f.assignments.targetWrites) != 0 {
			t.Fatalf("%s no deberia mutar assigned_exercise", kind)
		}
		if len(f.suggestions.applied) != 0 {
			t.Fatalf("%s no deberia sellar applied_at", kind)
		}
	}
}

func TestReviewFallaExplicitoSiNoHayProximaOcurrencia(t *testing.T) {
	exerciseID := uuid.New()
	f := newSuggestionFixture("load_adjust", 5, exerciseID, domain.AssignedExercise{})
	f.assignments.next = nil // el plan termino o se cancelo entre la sugerencia y la aprobacion

	_, err := f.svc.Review(context.Background(), uuid.New(), uuid.New(), uuid.New(), true)
	if err == nil {
		t.Fatal("el coach tiene que enterarse de que su aprobacion no tuvo efecto")
	}
	if !errors.Is(err, domain.ErrInvalidInput) {
		t.Fatalf("esperaba ErrInvalidInput (400 para el panel), fue %v", err)
	}
}

func TestReviewFallaSiNoHayPesoSobreElQueAjustar(t *testing.T) {
	exerciseID := uuid.New()
	next := domain.AssignedExercise{ID: uuid.New(), ExerciseID: exerciseID} // sin TargetWeightKg
	f := newSuggestionFixture("load_adjust", 5, exerciseID, next)

	_, err := f.svc.Review(context.Background(), uuid.New(), uuid.New(), uuid.New(), true)
	if !errors.Is(err, domain.ErrInvalidInput) {
		t.Fatalf("esperaba ErrInvalidInput, fue %v", err)
	}
	if len(f.assignments.targetWrites) != 0 {
		t.Fatal("no deberia escribir nada si no hay sobre que ajustar")
	}
}
