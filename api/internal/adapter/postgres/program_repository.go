package postgres

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"myvibesfit/api/internal/domain"
	"myvibesfit/api/internal/repository/db"
)

type ProgramRepository struct{ q db.Querier }

func NewProgramRepository(q db.Querier) *ProgramRepository { return &ProgramRepository{q: q} }

var _ domain.ProgramRepository = (*ProgramRepository)(nil)

func (r *ProgramRepository) Create(ctx context.Context, p domain.Program) (domain.Program, error) {
	row, err := r.q.CreateProgram(ctx, db.CreateProgramParams{
		OrgID: p.OrgID, CreatedBy: p.CreatedBy, Name: p.Name, Description: stringToPgText(p.Description),
		Goal: db.TrainingGoal(p.Goal), Level: db.ExperienceLevel(p.Level), TotalWeeks: p.TotalWeeks, DaysPerWeek: p.DaysPerWeek,
	})
	if err != nil {
		return domain.Program{}, err
	}
	return toDomainProgram(row), nil
}

func (r *ProgramRepository) GetByID(ctx context.Context, id, orgID uuid.UUID) (domain.Program, error) {
	row, err := r.q.GetProgramByID(ctx, db.GetProgramByIDParams{ID: id, OrgID: orgID})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Program{}, domain.ErrNotFound
		}
		return domain.Program{}, err
	}
	return toDomainProgram(row), nil
}

func (r *ProgramRepository) ListByOrg(ctx context.Context, f domain.ListProgramsFilter) ([]domain.Program, error) {
	params := db.ListProgramsByOrgParams{OrgID: f.OrgID, LimitCount: f.Limit, OffsetCount: f.Offset}
	if f.Status != "" {
		params.Status = db.NullProgramStatus{ProgramStatus: db.ProgramStatus(f.Status), Valid: true}
	}
	rows, err := r.q.ListProgramsByOrg(ctx, params)
	if err != nil {
		return nil, err
	}
	out := make([]domain.Program, len(rows))
	for i, row := range rows {
		out[i] = toDomainProgram(row)
	}
	return out, nil
}

func (r *ProgramRepository) Update(ctx context.Context, p domain.Program) (domain.Program, error) {
	row, err := r.q.UpdateProgram(ctx, db.UpdateProgramParams{
		ID: p.ID, OrgID: p.OrgID, Name: p.Name, Description: stringToPgText(p.Description),
		Goal: db.TrainingGoal(p.Goal), Level: db.ExperienceLevel(p.Level), TotalWeeks: p.TotalWeeks, DaysPerWeek: p.DaysPerWeek,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Program{}, domain.ErrNotFound
		}
		return domain.Program{}, err
	}
	return toDomainProgram(row), nil
}

func (r *ProgramRepository) SetStatus(ctx context.Context, id, orgID uuid.UUID, status domain.ProgramStatus) (domain.Program, error) {
	row, err := r.q.SetProgramStatus(ctx, db.SetProgramStatusParams{ID: id, OrgID: orgID, Status: db.ProgramStatus(status)})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Program{}, domain.ErrNotFound
		}
		return domain.Program{}, err
	}
	return toDomainProgram(row), nil
}

func (r *ProgramRepository) CreateWorkout(ctx context.Context, w domain.ProgramWorkout) (domain.ProgramWorkout, error) {
	row, err := r.q.CreateProgramWorkout(ctx, db.CreateProgramWorkoutParams{
		ProgramID: w.ProgramID, WeekNumber: w.WeekNumber, DayIndex: w.DayIndex, Name: w.Name,
		Note: stringToPgText(w.Note), IsDeload: w.IsDeload, EstimatedMinutes: intPtrToPgInt2(w.EstimatedMinutes),
	})
	if err != nil {
		return domain.ProgramWorkout{}, err
	}
	return toDomainProgramWorkout(row), nil
}

func (r *ProgramRepository) ListWorkouts(ctx context.Context, programID uuid.UUID) ([]domain.ProgramWorkout, error) {
	rows, err := r.q.ListProgramWorkouts(ctx, programID)
	if err != nil {
		return nil, err
	}
	out := make([]domain.ProgramWorkout, len(rows))
	for i, row := range rows {
		out[i] = toDomainProgramWorkout(row)
	}
	return out, nil
}

func (r *ProgramRepository) UpdateWorkout(ctx context.Context, w domain.ProgramWorkout) (domain.ProgramWorkout, error) {
	row, err := r.q.UpdateProgramWorkout(ctx, db.UpdateProgramWorkoutParams{
		ID: w.ID, Name: w.Name, Note: stringToPgText(w.Note), IsDeload: w.IsDeload, EstimatedMinutes: intPtrToPgInt2(w.EstimatedMinutes),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ProgramWorkout{}, domain.ErrNotFound
		}
		return domain.ProgramWorkout{}, err
	}
	return toDomainProgramWorkout(row), nil
}

func (r *ProgramRepository) DeleteWorkout(ctx context.Context, id uuid.UUID) error {
	return r.q.DeleteProgramWorkout(ctx, id)
}

func (r *ProgramRepository) GetWorkoutOrgID(ctx context.Context, workoutID uuid.UUID) (uuid.UUID, error) {
	row, err := r.q.GetProgramWorkoutWithOrg(ctx, workoutID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return uuid.UUID{}, domain.ErrNotFound
		}
		return uuid.UUID{}, err
	}
	return row.ProgramOrgID, nil
}

func (r *ProgramRepository) CreateExercise(ctx context.Context, e domain.ProgramExercise) (domain.ProgramExercise, error) {
	row, err := r.q.CreateProgramExercise(ctx, db.CreateProgramExerciseParams{
		ProgramWorkoutID: e.ProgramWorkoutID, ExerciseID: e.ExerciseID, OrderIndex: e.OrderIndex,
		SupersetGroup: intPtrToPgInt2(e.SupersetGroup), TargetSets: e.TargetSets, TargetRepsMin: intPtrToPgInt2(e.TargetRepsMin),
		TargetRepsMax: intPtrToPgInt2(e.TargetRepsMax), TargetRpe: e.TargetRPE, TargetPct1rm: e.TargetPct1RM,
		RestSeconds: e.RestSeconds, Tempo: stringToPgText(e.Tempo), Note: stringToPgText(e.Note),
		ProgressionRuleID: uuidPtrToPgUUID(e.ProgressionRuleID),
	})
	if err != nil {
		return domain.ProgramExercise{}, err
	}
	return toDomainProgramExercise(row), nil
}

func (r *ProgramRepository) ListExercises(ctx context.Context, workoutID uuid.UUID) ([]domain.ProgramExercise, error) {
	rows, err := r.q.ListProgramExercises(ctx, workoutID)
	if err != nil {
		return nil, err
	}
	out := make([]domain.ProgramExercise, len(rows))
	for i, row := range rows {
		out[i] = toDomainProgramExercise(row)
	}
	return out, nil
}

func (r *ProgramRepository) UpdateExercise(ctx context.Context, e domain.ProgramExercise) (domain.ProgramExercise, error) {
	row, err := r.q.UpdateProgramExercise(ctx, db.UpdateProgramExerciseParams{
		ID: e.ID, OrderIndex: e.OrderIndex, SupersetGroup: intPtrToPgInt2(e.SupersetGroup), TargetSets: e.TargetSets,
		TargetRepsMin: intPtrToPgInt2(e.TargetRepsMin), TargetRepsMax: intPtrToPgInt2(e.TargetRepsMax),
		TargetRpe: e.TargetRPE, TargetPct1rm: e.TargetPct1RM, RestSeconds: e.RestSeconds,
		Tempo: stringToPgText(e.Tempo), Note: stringToPgText(e.Note), ProgressionRuleID: uuidPtrToPgUUID(e.ProgressionRuleID),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ProgramExercise{}, domain.ErrNotFound
		}
		return domain.ProgramExercise{}, err
	}
	return toDomainProgramExercise(row), nil
}

func (r *ProgramRepository) DeleteExercise(ctx context.Context, id uuid.UUID) error {
	return r.q.DeleteProgramExercise(ctx, id)
}

func (r *ProgramRepository) GetExerciseOrgID(ctx context.Context, exerciseID uuid.UUID) (uuid.UUID, error) {
	row, err := r.q.GetProgramExerciseWithOrg(ctx, exerciseID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return uuid.UUID{}, domain.ErrNotFound
		}
		return uuid.UUID{}, err
	}
	return row.ProgramOrgID, nil
}

func (r *ProgramRepository) ListProgressionRules(ctx context.Context, orgID *uuid.UUID) ([]domain.ProgressionRule, error) {
	rows, err := r.q.ListProgressionRules(ctx, uuidPtrToPgUUID(orgID))
	if err != nil {
		return nil, err
	}
	out := make([]domain.ProgressionRule, len(rows))
	for i, row := range rows {
		out[i] = toDomainProgressionRule(row)
	}
	return out, nil
}

func (r *ProgramRepository) CreateProgressionRule(ctx context.Context, rule domain.ProgressionRule) (domain.ProgressionRule, error) {
	row, err := r.q.CreateProgressionRule(ctx, db.CreateProgressionRuleParams{
		OrgID: uuidPtrToPgUUID(rule.OrgID), Name: rule.Name, Type: db.ProgressionType(rule.Type), Params: rule.Params, IsSystem: rule.IsSystem,
	})
	if err != nil {
		return domain.ProgressionRule{}, err
	}
	return toDomainProgressionRule(row), nil
}

func (r *ProgramRepository) GetProgressionRuleByID(ctx context.Context, id uuid.UUID) (domain.ProgressionRule, error) {
	row, err := r.q.GetProgressionRuleByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ProgressionRule{}, domain.ErrNotFound
		}
		return domain.ProgressionRule{}, err
	}
	return toDomainProgressionRule(row), nil
}

func toDomainProgram(p db.Program) domain.Program {
	return domain.Program{
		ID: p.ID, OrgID: p.OrgID, CreatedBy: p.CreatedBy, Name: p.Name, Description: pgTextToString(p.Description),
		Goal: string(p.Goal), Level: string(p.Level), TotalWeeks: p.TotalWeeks, DaysPerWeek: p.DaysPerWeek,
		Status: domain.ProgramStatus(p.Status), IsShared: p.IsShared, CreatedAt: p.CreatedAt, UpdatedAt: p.UpdatedAt,
	}
}

func toDomainProgramWorkout(w db.ProgramWorkout) domain.ProgramWorkout {
	return domain.ProgramWorkout{
		ID: w.ID, ProgramID: w.ProgramID, WeekNumber: w.WeekNumber, DayIndex: w.DayIndex, Name: w.Name,
		Note: pgTextToString(w.Note), IsDeload: w.IsDeload, EstimatedMinutes: pgInt2ToPtr(w.EstimatedMinutes),
	}
}

func toDomainProgramExercise(e db.ProgramExercise) domain.ProgramExercise {
	return domain.ProgramExercise{
		ID: e.ID, ProgramWorkoutID: e.ProgramWorkoutID, ExerciseID: e.ExerciseID, OrderIndex: e.OrderIndex,
		SupersetGroup: pgInt2ToPtr(e.SupersetGroup), TargetSets: e.TargetSets, TargetRepsMin: pgInt2ToPtr(e.TargetRepsMin),
		TargetRepsMax: pgInt2ToPtr(e.TargetRepsMax), TargetRPE: e.TargetRpe, TargetPct1RM: e.TargetPct1rm,
		RestSeconds: e.RestSeconds, Tempo: pgTextToString(e.Tempo), Note: pgTextToString(e.Note),
		ProgressionRuleID: pgUUIDToPtr(e.ProgressionRuleID),
	}
}

func toDomainProgressionRule(r db.ProgressionRule) domain.ProgressionRule {
	return domain.ProgressionRule{
		ID: r.ID, OrgID: pgUUIDToPtr(r.OrgID), Name: r.Name, Type: string(r.Type), Params: r.Params,
		IsSystem: r.IsSystem, CreatedAt: r.CreatedAt,
	}
}
