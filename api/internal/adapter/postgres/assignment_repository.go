package postgres

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"

	"myvibesfit/api/internal/domain"
	"myvibesfit/api/internal/repository/db"
)

type AssignmentRepository struct{ q db.Querier }

func NewAssignmentRepository(q db.Querier) *AssignmentRepository { return &AssignmentRepository{q: q} }

var _ domain.AssignmentRepository = (*AssignmentRepository)(nil)

func (r *AssignmentRepository) Create(ctx context.Context, a domain.Assignment) (domain.Assignment, error) {
	row, err := r.q.CreateAssignment(ctx, db.CreateAssignmentParams{
		OrgID: a.OrgID, ProgramID: uuidPtrToPgUUID(a.ProgramID), ClientUserID: a.ClientUserID,
		CoachUserID: uuidPtrToPgUUID(a.CoachUserID), Name: a.Name, StartDate: a.StartDate,
	})
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return domain.Assignment{}, domain.ErrAlreadyExists
		}
		return domain.Assignment{}, err
	}
	return toDomainAssignment(row), nil
}

func (r *AssignmentRepository) GetByID(ctx context.Context, id uuid.UUID) (domain.Assignment, error) {
	row, err := r.q.GetAssignmentByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Assignment{}, domain.ErrNotFound
		}
		return domain.Assignment{}, err
	}
	return toDomainAssignment(row), nil
}

func (r *AssignmentRepository) GetActiveByClient(ctx context.Context, clientUserID uuid.UUID) (domain.Assignment, error) {
	row, err := r.q.GetActiveAssignmentByClient(ctx, clientUserID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Assignment{}, domain.ErrNotFound
		}
		return domain.Assignment{}, err
	}
	return toDomainAssignment(row), nil
}

func (r *AssignmentRepository) Cancel(ctx context.Context, id, orgID uuid.UUID) (domain.Assignment, error) {
	row, err := r.q.CancelAssignment(ctx, db.CancelAssignmentParams{ID: id, OrgID: orgID})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Assignment{}, domain.ErrNotFound
		}
		return domain.Assignment{}, err
	}
	return toDomainAssignment(row), nil
}

func (r *AssignmentRepository) CreateWorkout(ctx context.Context, w domain.AssignedWorkout) (domain.AssignedWorkout, error) {
	row, err := r.q.CreateAssignedWorkout(ctx, db.CreateAssignedWorkoutParams{
		AssignmentID: w.AssignmentID, SourceWorkoutID: uuidPtrToPgUUID(w.SourceWorkoutID), WeekNumber: w.WeekNumber,
		DayIndex: w.DayIndex, Name: w.Name, Note: stringToPgText(w.Note), IsDeload: w.IsDeload, ScheduledOn: w.ScheduledOn,
	})
	if err != nil {
		return domain.AssignedWorkout{}, err
	}
	return toDomainAssignedWorkout(row), nil
}

func (r *AssignmentRepository) ListWorkouts(ctx context.Context, assignmentID uuid.UUID) ([]domain.AssignedWorkout, error) {
	rows, err := r.q.ListAssignedWorkouts(ctx, assignmentID)
	if err != nil {
		return nil, err
	}
	out := make([]domain.AssignedWorkout, len(rows))
	for i, row := range rows {
		out[i] = toDomainAssignedWorkout(row)
	}
	return out, nil
}

func (r *AssignmentRepository) CreateExercise(ctx context.Context, e domain.AssignedExercise) (domain.AssignedExercise, error) {
	row, err := r.q.CreateAssignedExercise(ctx, db.CreateAssignedExerciseParams{
		AssignedWorkoutID: e.AssignedWorkoutID, ExerciseID: e.ExerciseID, OrderIndex: e.OrderIndex,
		SupersetGroup: intPtrToPgInt2(e.SupersetGroup), TargetSets: e.TargetSets, TargetRepsMin: intPtrToPgInt2(e.TargetRepsMin),
		TargetRepsMax: intPtrToPgInt2(e.TargetRepsMax), TargetRpe: e.TargetRPE, TargetWeightKg: e.TargetWeightKg,
		RestSeconds: e.RestSeconds, Tempo: stringToPgText(e.Tempo), Note: stringToPgText(e.Note),
		ProgressionRuleID: uuidPtrToPgUUID(e.ProgressionRuleID),
	})
	if err != nil {
		return domain.AssignedExercise{}, err
	}
	return toDomainAssignedExercise(row), nil
}

func (r *AssignmentRepository) ListExercises(ctx context.Context, assignedWorkoutID uuid.UUID) ([]domain.AssignedExercise, error) {
	rows, err := r.q.ListAssignedExercises(ctx, assignedWorkoutID)
	if err != nil {
		return nil, err
	}
	out := make([]domain.AssignedExercise, len(rows))
	for i, row := range rows {
		out[i] = toDomainAssignedExercise(row)
	}
	return out, nil
}

func (r *AssignmentRepository) GetOrgMembership(ctx context.Context, orgID, userID uuid.UUID) (domain.Membership, error) {
	row, err := r.q.GetOrgMembership(ctx, db.GetOrgMembershipParams{OrgID: orgID, UserID: userID})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Membership{}, domain.ErrNotFound
		}
		return domain.Membership{}, err
	}
	return toDomainMembership(row), nil
}

func (r *AssignmentRepository) GetWorkoutAssignmentID(ctx context.Context, assignedWorkoutID uuid.UUID) (uuid.UUID, error) {
	id, err := r.q.GetWorkoutAssignmentID(ctx, assignedWorkoutID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return uuid.Nil, domain.ErrNotFound
		}
		return uuid.Nil, err
	}
	return id, nil
}

func (r *AssignmentRepository) MarkWorkoutCompleted(ctx context.Context, assignedWorkoutID uuid.UUID) error {
	return r.q.MarkAssignedWorkoutCompleted(ctx, assignedWorkoutID)
}

func (r *AssignmentRepository) FindNextExerciseOccurrence(ctx context.Context, assignmentID, exerciseID, excludeWorkoutID uuid.UUID) (domain.AssignedExercise, bool, error) {
	row, err := r.q.FindNextAssignedExercise(ctx, db.FindNextAssignedExerciseParams{
		AssignmentID: assignmentID, ExerciseID: exerciseID, AssignedWorkoutID: excludeWorkoutID,
	})
	if err != nil {
		// Sin proxima ocurrencia no es un error de infraestructura: el plan
		// puede haber terminado. El caller decide que hacer.
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.AssignedExercise{}, false, nil
		}
		return domain.AssignedExercise{}, false, err
	}
	return toDomainAssignedExercise(row), true, nil
}

func (r *AssignmentRepository) UpdateExerciseTargets(ctx context.Context, id uuid.UUID, weightKg *float64, repsMin, repsMax *int, overrideSource *string) (domain.AssignedExercise, error) {
	var source pgtype.Text
	if overrideSource != nil {
		source = pgtype.Text{String: *overrideSource, Valid: true}
	}
	row, err := r.q.UpdateAssignedExerciseTargets(ctx, db.UpdateAssignedExerciseTargetsParams{
		ID: id, TargetWeightKg: weightKg, TargetRepsMin: intPtrToPgInt2(repsMin),
		TargetRepsMax: intPtrToPgInt2(repsMax), OverrideSource: source,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.AssignedExercise{}, domain.ErrNotFound
		}
		return domain.AssignedExercise{}, err
	}
	return toDomainAssignedExercise(row), nil
}

func toDomainAssignment(a db.Assignment) domain.Assignment {
	return domain.Assignment{
		ID: a.ID, OrgID: a.OrgID, ProgramID: pgUUIDToPtr(a.ProgramID), ClientUserID: a.ClientUserID,
		CoachUserID: pgUUIDToPtr(a.CoachUserID), Name: a.Name, StartDate: a.StartDate, EndDate: a.EndDate,
		Status: string(a.Status), CreatedAt: a.CreatedAt, UpdatedAt: a.UpdatedAt,
	}
}

func toDomainAssignedWorkout(w db.AssignedWorkout) domain.AssignedWorkout {
	return domain.AssignedWorkout{
		ID: w.ID, AssignmentID: w.AssignmentID, SourceWorkoutID: pgUUIDToPtr(w.SourceWorkoutID), WeekNumber: w.WeekNumber,
		DayIndex: w.DayIndex, Name: w.Name, Note: pgTextToString(w.Note), IsDeload: w.IsDeload,
		ScheduledOn: w.ScheduledOn, Status: string(w.Status),
	}
}

func toDomainAssignedExercise(e db.AssignedExercise) domain.AssignedExercise {
	return domain.AssignedExercise{
		ID: e.ID, AssignedWorkoutID: e.AssignedWorkoutID, ExerciseID: e.ExerciseID, OrderIndex: e.OrderIndex,
		SupersetGroup: pgInt2ToPtr(e.SupersetGroup), TargetSets: e.TargetSets, TargetRepsMin: pgInt2ToPtr(e.TargetRepsMin),
		TargetRepsMax: pgInt2ToPtr(e.TargetRepsMax), TargetRPE: e.TargetRpe, TargetWeightKg: e.TargetWeightKg,
		RestSeconds: e.RestSeconds, Tempo: pgTextToString(e.Tempo), Note: pgTextToString(e.Note),
		ProgressionRuleID: pgUUIDToPtr(e.ProgressionRuleID), OverrideSource: pgTextToPtr(e.OverrideSource),
	}
}
