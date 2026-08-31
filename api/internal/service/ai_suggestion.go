package service

import (
	"context"
	"encoding/json"
	"fmt"
	"math"

	"github.com/google/uuid"

	"myvibesfit/api/internal/domain"
)

type AISuggestionService struct {
	repo  domain.AISuggestionRepository
	uow   domain.UnitOfWork
	audit *AuditLogger
}

func NewAISuggestionService(repo domain.AISuggestionRepository, uow domain.UnitOfWork, audit *AuditLogger) *AISuggestionService {
	return &AISuggestionService{repo: repo, uow: uow, audit: audit}
}

func (s *AISuggestionService) ListPending(ctx context.Context, coachID, orgID uuid.UUID) ([]domain.PendingSuggestion, error) {
	return s.repo.ListPendingForCoach(ctx, coachID, orgID)
}

// Review registra la decision del coach y, si aprueba un kind aplicable,
// escribe el ajuste en la proxima ocurrencia del ejercicio. Ambas cosas
// corren en la misma transaccion: marcarla aprobada sin aplicarla le haria
// creer al coach que el ajuste esta puesto cuando no lo esta.
//
// El target queda marcado con override_source para que el motor de
// progresion (que corre al sincronizar la sesion siguiente) lo respete una
// vez en lugar de recalcularlo encima.
func (s *AISuggestionService) Review(ctx context.Context, id, orgID, reviewerID uuid.UUID, approve bool) (domain.AISuggestion, error) {
	status := domain.SuggestionStatusRejected
	if approve {
		status = domain.SuggestionStatusApproved
	}

	var reviewed domain.AISuggestion
	err := s.uow.Execute(ctx, func(repos domain.TxRepos) error {
		sug, err := repos.Suggestions.Review(ctx, id, orgID, status, reviewerID)
		if err != nil {
			return err
		}
		reviewed = sug

		if !approve || !domain.AppliableKinds[sug.Kind] {
			return nil
		}
		if err := applySuggestion(ctx, repos, sug); err != nil {
			return err
		}
		return repos.Suggestions.MarkApplied(ctx, id, orgID)
	})
	if err != nil {
		return domain.AISuggestion{}, err
	}

	s.audit.Log(ctx, orgID, reviewerID, "ai_suggestion.review", "ai_suggestion", id.String(), map[string]any{"status": status})
	return reviewed, nil
}

// applySuggestion muta la proxima ocurrencia del ejercicio objetivo. Falla
// con un error explicito (no en silencio) cuando ya no hay donde aplicar:
// el coach tiene que enterarse de que su aprobacion no tuvo efecto.
func applySuggestion(ctx context.Context, repos domain.TxRepos, sug domain.AISuggestion) error {
	if sug.AssignmentID == nil {
		return fmt.Errorf("%w: la sugerencia no apunta a ninguna asignacion", domain.ErrInvalidInput)
	}

	var payload domain.SuggestionPayload
	if err := json.Unmarshal(sug.Payload, &payload); err != nil {
		return fmt.Errorf("%w: payload ilegible", domain.ErrInvalidInput)
	}
	exerciseID, err := uuid.Parse(payload.TargetExerciseID)
	if err != nil {
		return fmt.Errorf("%w: la sugerencia no identifica un ejercicio", domain.ErrInvalidInput)
	}

	// uuid.Nil como "dia a excluir": no se esta excluyendo ninguno, se busca
	// la primera ocurrencia pendiente.
	next, found, err := repos.Assignments.FindNextExerciseOccurrence(ctx, *sug.AssignmentID, exerciseID, uuid.Nil)
	if err != nil {
		return err
	}
	if !found {
		return fmt.Errorf("%w: el ejercicio ya no aparece en lo que queda del plan, la sugerencia no se pudo aplicar", domain.ErrInvalidInput)
	}

	factor := 1 + payload.DeltaPercent/100
	override := domain.OverrideSourceAISuggestion

	switch sug.Kind {
	case "load_adjust":
		if next.TargetWeightKg == nil {
			return fmt.Errorf("%w: el ejercicio no tiene peso objetivo sobre el que ajustar", domain.ErrInvalidInput)
		}
		weight := roundToNearest(*next.TargetWeightKg*factor, 2.5)
		_, err = repos.Assignments.UpdateExerciseTargets(ctx, next.ID, &weight, next.TargetRepsMin, next.TargetRepsMax, &override)

	case "volume_adjust":
		// El volumen se ajusta sobre el rango de repeticiones: cambiar series
		// altera la duracion de la sesion, y en incrementos chicos el redondeo
		// lo vuelve todo-o-nada.
		if next.TargetRepsMin == nil && next.TargetRepsMax == nil {
			return fmt.Errorf("%w: el ejercicio no tiene rango de repeticiones sobre el que ajustar", domain.ErrInvalidInput)
		}
		repsMin := scaleReps(next.TargetRepsMin, factor)
		repsMax := scaleReps(next.TargetRepsMax, factor)
		_, err = repos.Assignments.UpdateExerciseTargets(ctx, next.ID, next.TargetWeightKg, repsMin, repsMax, &override)
	}
	return err
}

// scaleReps nunca baja de 1: una serie de cero repeticiones no es una serie.
func scaleReps(reps *int, factor float64) *int {
	if reps == nil {
		return nil
	}
	scaled := int(math.Round(float64(*reps) * factor))
	if scaled < 1 {
		scaled = 1
	}
	return &scaled
}

func roundToNearest(value, step float64) float64 {
	if step <= 0 {
		return value
	}
	return math.Round(value/step) * step
}
