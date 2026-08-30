package service

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"

	"myvibesfit/api/internal/domain"
)

// Los fakes embeben la interfaz del port: solo implementan lo que este flujo
// llama, y cualquier metodo no previsto revienta en el test en vez de pasar
// desapercibido.

type fakeAssignments struct {
	domain.AssignmentRepository

	assignmentID uuid.UUID
	byWorkout    map[uuid.UUID][]domain.AssignedExercise
	next         *domain.AssignedExercise

	completedWorkouts []uuid.UUID
	targetWrites      []targetWrite
}

type targetWrite struct {
	id             uuid.UUID
	weightKg       *float64
	repsMin        *int
	repsMax        *int
	overrideSource *string
}

func (f *fakeAssignments) GetWorkoutAssignmentID(context.Context, uuid.UUID) (uuid.UUID, error) {
	return f.assignmentID, nil
}

func (f *fakeAssignments) ListExercises(_ context.Context, workoutID uuid.UUID) ([]domain.AssignedExercise, error) {
	return f.byWorkout[workoutID], nil
}

func (f *fakeAssignments) MarkWorkoutCompleted(_ context.Context, id uuid.UUID) error {
	f.completedWorkouts = append(f.completedWorkouts, id)
	return nil
}

func (f *fakeAssignments) FindNextExerciseOccurrence(context.Context, uuid.UUID, uuid.UUID, uuid.UUID) (domain.AssignedExercise, bool, error) {
	if f.next == nil {
		return domain.AssignedExercise{}, false, nil
	}
	return *f.next, true, nil
}

func (f *fakeAssignments) UpdateExerciseTargets(_ context.Context, id uuid.UUID, weightKg *float64, repsMin, repsMax *int, overrideSource *string) (domain.AssignedExercise, error) {
	f.targetWrites = append(f.targetWrites, targetWrite{id, weightKg, repsMin, repsMax, overrideSource})
	return domain.AssignedExercise{ID: id}, nil
}

type fakePrograms struct {
	domain.ProgramRepository
	rule domain.ProgressionRule
}

func (f *fakePrograms) GetProgressionRuleByID(context.Context, uuid.UUID) (domain.ProgressionRule, error) {
	return f.rule, nil
}

type fakeSessions struct {
	domain.SessionRepository
	nextSetLogID int64
}

func (f *fakeSessions) UpsertWorkoutSession(_ context.Context, s domain.WorkoutSession) (domain.WorkoutSession, error) {
	s.ID = uuid.New()
	return s, nil
}

func (f *fakeSessions) UpsertSessionExercise(_ context.Context, e domain.SessionExercise) (domain.SessionExercise, error) {
	e.ID = uuid.New()
	return e, nil
}

func (f *fakeSessions) UpsertSetLog(_ context.Context, l domain.SetLog) (domain.SetLog, error) {
	f.nextSetLogID++
	l.ID = f.nextSetLogID
	return l, nil
}

func (f *fakeSessions) GetPersonalRecord(context.Context, uuid.UUID, uuid.UUID, domain.RecordType) (domain.PersonalRecord, error) {
	return domain.PersonalRecord{}, domain.ErrNotFound
}

func (f *fakeSessions) UpsertPersonalRecord(_ context.Context, p domain.PersonalRecord) (domain.PersonalRecord, error) {
	return p, nil
}

type fakeGamification struct {
	domain.GamificationRepository
}

func (f *fakeGamification) GetUserStats(context.Context, uuid.UUID) (domain.UserStat, error) {
	return domain.UserStat{}, domain.ErrNotFound
}

func (f *fakeGamification) UpsertUserStats(_ context.Context, s domain.UserStat) (domain.UserStat, error) {
	return s, nil
}

func (f *fakeGamification) GetUserStreak(context.Context, uuid.UUID, string) (domain.UserStreak, error) {
	return domain.UserStreak{}, domain.ErrNotFound
}

func (f *fakeGamification) UpsertUserStreak(_ context.Context, s domain.UserStreak) (domain.UserStreak, error) {
	return s, nil
}

func (f *fakeGamification) CreateXPEvent(context.Context, domain.XPEvent) error { return nil }

func (f *fakeGamification) ListAchievements(context.Context) ([]domain.Achievement, error) {
	return nil, nil
}

// fakeUoW corre la funcion con los repos falsos, sin transaccion real.
type fakeUoW struct{ repos domain.TxRepos }

func (u *fakeUoW) Execute(ctx context.Context, fn func(domain.TxRepos) error) error {
	return fn(u.repos)
}

type syncFixture struct {
	svc         *SyncService
	assignments *fakeAssignments
	workoutID   uuid.UUID
	nextID      uuid.UUID
}

// newSyncFixture arma un plan con el ejercicio en dos dias: el que se
// entrena ahora y la proxima ocurrencia que deberia recibir el nuevo target.
func newSyncFixture(t *testing.T, rule domain.ProgressionRule, trained domain.AssignedExercise, next domain.AssignedExercise) syncFixture {
	t.Helper()
	workoutID := uuid.New()
	trained.AssignedWorkoutID = workoutID

	assignments := &fakeAssignments{
		assignmentID: uuid.New(),
		byWorkout:    map[uuid.UUID][]domain.AssignedExercise{workoutID: {trained}},
		next:         &next,
	}
	repos := domain.TxRepos{
		Assignments:  assignments,
		Sessions:     &fakeSessions{},
		Gamification: &fakeGamification{},
	}
	svc := NewSyncService(&fakeUoW{repos: repos}, domain.NewGamificationService(), assignments, &fakePrograms{rule: rule})

	return syncFixture{svc: svc, assignments: assignments, workoutID: workoutID, nextID: next.ID}
}

func completedSession(workoutID, exerciseID uuid.UUID, reps []int, weightKg float64) SyncSessionInput {
	sets := make([]SyncSetInput, len(reps))
	for i, r := range reps {
		w, rr := weightKg, r
		sets[i] = SyncSetInput{
			ClientLocalID: uuid.New(), SetNumber: i + 1, Type: "working",
			WeightKg: &w, Reps: &rr, IsCompleted: true, PerformedAt: time.Now(),
		}
	}
	return SyncSessionInput{
		ClientLocalID: uuid.New(), Name: "Dia A", Status: string(domain.SessionStatusCompleted),
		AssignedWorkoutID: &workoutID, StartedAt: time.Now(),
		Exercises: []SyncExerciseInput{{ExerciseID: exerciseID, OrderIndex: 1, Sets: sets}},
	}
}

func ptrInt(v int) *int           { return &v }
func ptrFloat(v float64) *float64 { return &v }

func TestSyncSessionsAplicaProgresionALaProximaOcurrencia(t *testing.T) {
	ruleID := uuid.New()
	exerciseID := uuid.New()
	rule := domain.ProgressionRule{
		ID: ruleID, Type: "double_progression",
		Params: json.RawMessage(`{"increment_kg": 2.5, "round_to_kg": 2.5}`),
	}
	trained := domain.AssignedExercise{
		ID: uuid.New(), ExerciseID: exerciseID, ProgressionRuleID: &ruleID,
		TargetWeightKg: ptrFloat(60), TargetRepsMin: ptrInt(8), TargetRepsMax: ptrInt(10),
	}
	next := domain.AssignedExercise{ID: uuid.New(), ExerciseID: exerciseID, ProgressionRuleID: &ruleID}

	f := newSyncFixture(t, rule, trained, next)

	// Tres series a 60 kg tocando el techo del rango: doble progresion sube peso.
	_, err := f.svc.SyncSessions(context.Background(), uuid.New(), nil,
		[]SyncSessionInput{completedSession(f.workoutID, exerciseID, []int{10, 10, 10}, 60)})
	if err != nil {
		t.Fatalf("sync: %v", err)
	}

	if len(f.assignments.completedWorkouts) != 1 || f.assignments.completedWorkouts[0] != f.workoutID {
		t.Fatalf("el dia entrenado deberia quedar completado, fue %v", f.assignments.completedWorkouts)
	}
	if len(f.assignments.targetWrites) != 1 {
		t.Fatalf("esperaba un solo write de target, fueron %d", len(f.assignments.targetWrites))
	}
	w := f.assignments.targetWrites[0]
	if w.id != f.nextID {
		t.Fatalf("el target se escribio en %v, deberia ir a la proxima ocurrencia %v", w.id, f.nextID)
	}
	if w.weightKg == nil || *w.weightKg != 62.5 {
		t.Fatalf("esperaba 62.5 kg (60 + incremento), fue %v", w.weightKg)
	}
	if w.overrideSource != nil {
		t.Fatalf("el motor no deberia marcar override, fue %v", *w.overrideSource)
	}
}

func TestSyncSessionsNoPisaOverrideDeIAYLoConsume(t *testing.T) {
	ruleID := uuid.New()
	exerciseID := uuid.New()
	rule := domain.ProgressionRule{ID: ruleID, Type: "double_progression"}
	trained := domain.AssignedExercise{
		ID: uuid.New(), ExerciseID: exerciseID, ProgressionRuleID: &ruleID,
		TargetWeightKg: ptrFloat(60), TargetRepsMin: ptrInt(8), TargetRepsMax: ptrInt(10),
	}
	// La proxima ocurrencia ya tiene un target puesto por una sugerencia de
	// IA que el coach aprobo: la progresion no debe pisarlo.
	override := domain.OverrideSourceAISuggestion
	next := domain.AssignedExercise{
		ID: uuid.New(), ExerciseID: exerciseID, ProgressionRuleID: &ruleID,
		TargetWeightKg: ptrFloat(80), TargetRepsMin: ptrInt(5), TargetRepsMax: ptrInt(5),
		OverrideSource: &override,
	}

	f := newSyncFixture(t, rule, trained, next)

	_, err := f.svc.SyncSessions(context.Background(), uuid.New(), nil,
		[]SyncSessionInput{completedSession(f.workoutID, exerciseID, []int{10, 10, 10}, 60)})
	if err != nil {
		t.Fatalf("sync: %v", err)
	}

	if len(f.assignments.targetWrites) != 1 {
		t.Fatalf("esperaba un write (el que consume el flag), fueron %d", len(f.assignments.targetWrites))
	}
	w := f.assignments.targetWrites[0]
	if w.weightKg == nil || *w.weightKg != 80 {
		t.Fatalf("el peso de la sugerencia debia sobrevivir en 80, fue %v", w.weightKg)
	}
	if w.repsMin == nil || *w.repsMin != 5 {
		t.Fatalf("las reps de la sugerencia debian sobrevivir en 5, fue %v", w.repsMin)
	}
	if w.overrideSource != nil {
		t.Fatal("el flag deberia quedar consumido (nil) para que la proxima sesion progrese normal")
	}
}

func TestSyncSessionsSinProximaOcurrenciaNoEscribeNada(t *testing.T) {
	ruleID := uuid.New()
	exerciseID := uuid.New()
	rule := domain.ProgressionRule{ID: ruleID, Type: "double_progression"}
	trained := domain.AssignedExercise{
		ID: uuid.New(), ExerciseID: exerciseID, ProgressionRuleID: &ruleID,
		TargetWeightKg: ptrFloat(60), TargetRepsMin: ptrInt(8), TargetRepsMax: ptrInt(10),
	}

	f := newSyncFixture(t, rule, trained, domain.AssignedExercise{})
	f.assignments.next = nil // ultima vez que el ejercicio aparece en el plan

	_, err := f.svc.SyncSessions(context.Background(), uuid.New(), nil,
		[]SyncSessionInput{completedSession(f.workoutID, exerciseID, []int{10, 10, 10}, 60)})
	if err != nil {
		t.Fatalf("terminar el plan no deberia ser un error: %v", err)
	}
	if len(f.assignments.targetWrites) != 0 {
		t.Fatalf("no habia donde escribir, hubo %d writes", len(f.assignments.targetWrites))
	}
	if len(f.assignments.completedWorkouts) != 1 {
		t.Fatal("el dia entrenado igual deberia quedar completado")
	}
}

func TestSyncSessionsSinReglaNoProgresa(t *testing.T) {
	exerciseID := uuid.New()
	trained := domain.AssignedExercise{ID: uuid.New(), ExerciseID: exerciseID} // sin ProgressionRuleID
	next := domain.AssignedExercise{ID: uuid.New(), ExerciseID: exerciseID}

	f := newSyncFixture(t, domain.ProgressionRule{}, trained, next)

	_, err := f.svc.SyncSessions(context.Background(), uuid.New(), nil,
		[]SyncSessionInput{completedSession(f.workoutID, exerciseID, []int{10, 10, 10}, 60)})
	if err != nil {
		t.Fatalf("sync: %v", err)
	}
	if len(f.assignments.targetWrites) != 0 {
		t.Fatalf("un ejercicio sin regla no deberia progresar, hubo %d writes", len(f.assignments.targetWrites))
	}
}
