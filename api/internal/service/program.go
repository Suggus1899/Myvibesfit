package service

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"myvibesfit/api/internal/domain"
	"myvibesfit/api/internal/repository/db"
)

type ProgramService struct {
	q db.Querier
}

func NewProgramService(q db.Querier) *ProgramService {
	return &ProgramService{q: q}
}

// ---- reglas de progresion ----

func (s *ProgramService) ListProgressionRules(ctx context.Context, orgID *uuid.UUID) ([]db.ProgressionRule, error) {
	return s.q.ListProgressionRules(ctx, uuidToPg(orgID))
}

type CreateProgressionRuleInput struct {
	OrgID  *uuid.UUID
	Name   string
	Type   string
	Params json.RawMessage
}

func (s *ProgramService) CreateProgressionRule(ctx context.Context, in CreateProgressionRuleInput) (db.ProgressionRule, error) {
	if in.Name == "" || in.Type == "" {
		return db.ProgressionRule{}, domain.ErrInvalidInput
	}
	params := in.Params
	if len(params) == 0 {
		params = []byte("{}")
	}
	return s.q.CreateProgressionRule(ctx, db.CreateProgressionRuleParams{
		OrgID:    uuidToPg(in.OrgID),
		Name:     in.Name,
		Type:     db.ProgressionType(in.Type),
		Params:   params,
		IsSystem: false,
	})
}

// ---- programas ----

type CreateProgramInput struct {
	OrgID       uuid.UUID
	CreatedBy   uuid.UUID
	Name        string
	Description string
	Goal        string
	Level       string
	TotalWeeks  int
	DaysPerWeek int
}

func (s *ProgramService) CreateProgram(ctx context.Context, in CreateProgramInput) (db.Program, error) {
	if in.Name == "" {
		return db.Program{}, domain.ErrInvalidInput
	}
	return s.q.CreateProgram(ctx, db.CreateProgramParams{
		OrgID:       in.OrgID,
		CreatedBy:   in.CreatedBy,
		Name:        in.Name,
		Description: textToPg(in.Description),
		Goal:        goalOrDefault(in.Goal),
		Level:       levelOrDefault(in.Level),
		TotalWeeks:  int16OrDefault(in.TotalWeeks, 4),
		DaysPerWeek: int16OrDefault(in.DaysPerWeek, 3),
	})
}

func (s *ProgramService) GetProgram(ctx context.Context, id, orgID uuid.UUID) (db.Program, error) {
	p, err := s.q.GetProgramByID(ctx, db.GetProgramByIDParams{ID: id, OrgID: orgID})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return db.Program{}, domain.ErrNotFound
		}
		return db.Program{}, err
	}
	return p, nil
}

type ListProgramsFilter struct {
	OrgID  uuid.UUID
	Status string
	Limit  int32
	Offset int32
}

func (s *ProgramService) ListPrograms(ctx context.Context, f ListProgramsFilter) ([]db.Program, error) {
	limit := f.Limit
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	params := db.ListProgramsByOrgParams{OrgID: f.OrgID, LimitCount: limit, OffsetCount: f.Offset}
	if f.Status != "" {
		params.Status = db.NullProgramStatus{ProgramStatus: db.ProgramStatus(f.Status), Valid: true}
	}
	return s.q.ListProgramsByOrg(ctx, params)
}

type UpdateProgramInput struct {
	ID          uuid.UUID
	OrgID       uuid.UUID
	Name        string
	Description string
	Goal        string
	Level       string
	TotalWeeks  int
	DaysPerWeek int
}

func (s *ProgramService) UpdateProgram(ctx context.Context, in UpdateProgramInput) (db.Program, error) {
	if in.Name == "" {
		return db.Program{}, domain.ErrInvalidInput
	}
	p, err := s.q.UpdateProgram(ctx, db.UpdateProgramParams{
		ID:          in.ID,
		OrgID:       in.OrgID,
		Name:        in.Name,
		Description: textToPg(in.Description),
		Goal:        goalOrDefault(in.Goal),
		Level:       levelOrDefault(in.Level),
		TotalWeeks:  int16OrDefault(in.TotalWeeks, 4),
		DaysPerWeek: int16OrDefault(in.DaysPerWeek, 3),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return db.Program{}, domain.ErrNotFound
		}
		return db.Program{}, err
	}
	return p, nil
}

func (s *ProgramService) SetProgramStatus(ctx context.Context, id, orgID uuid.UUID, status db.ProgramStatus) (db.Program, error) {
	p, err := s.q.SetProgramStatus(ctx, db.SetProgramStatusParams{ID: id, OrgID: orgID, Status: status})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return db.Program{}, domain.ErrNotFound
		}
		return db.Program{}, err
	}
	return p, nil
}

// ---- dias del programa (program_workout) ----

type CreateProgramWorkoutInput struct {
	ProgramID        uuid.UUID
	OrgID            uuid.UUID
	WeekNumber       int
	DayIndex         int
	Name             string
	Note             string
	IsDeload         bool
	EstimatedMinutes *int
}

func (s *ProgramService) CreateProgramWorkout(ctx context.Context, in CreateProgramWorkoutInput) (db.ProgramWorkout, error) {
	if in.Name == "" || in.WeekNumber < 1 || in.DayIndex < 1 || in.DayIndex > 7 {
		return db.ProgramWorkout{}, domain.ErrInvalidInput
	}
	if _, err := s.GetProgram(ctx, in.ProgramID, in.OrgID); err != nil {
		return db.ProgramWorkout{}, err
	}

	return s.q.CreateProgramWorkout(ctx, db.CreateProgramWorkoutParams{
		ProgramID:        in.ProgramID,
		WeekNumber:       int16(in.WeekNumber),
		DayIndex:         int16(in.DayIndex),
		Name:             in.Name,
		Note:             textToPg(in.Note),
		IsDeload:         in.IsDeload,
		EstimatedMinutes: intToPgInt2(in.EstimatedMinutes),
	})
}

func (s *ProgramService) ListProgramWorkouts(ctx context.Context, programID, orgID uuid.UUID) ([]db.ProgramWorkout, error) {
	if _, err := s.GetProgram(ctx, programID, orgID); err != nil {
		return nil, err
	}
	return s.q.ListProgramWorkouts(ctx, programID)
}

func (s *ProgramService) workoutChecked(ctx context.Context, id, orgID uuid.UUID) (db.GetProgramWorkoutWithOrgRow, error) {
	row, err := s.q.GetProgramWorkoutWithOrg(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return db.GetProgramWorkoutWithOrgRow{}, domain.ErrNotFound
		}
		return db.GetProgramWorkoutWithOrgRow{}, err
	}
	if row.ProgramOrgID != orgID {
		return db.GetProgramWorkoutWithOrgRow{}, domain.ErrNotFound
	}
	return row, nil
}

type UpdateProgramWorkoutInput struct {
	ID               uuid.UUID
	OrgID            uuid.UUID
	Name             string
	Note             string
	IsDeload         bool
	EstimatedMinutes *int
}

func (s *ProgramService) UpdateProgramWorkout(ctx context.Context, in UpdateProgramWorkoutInput) (db.ProgramWorkout, error) {
	if in.Name == "" {
		return db.ProgramWorkout{}, domain.ErrInvalidInput
	}
	if _, err := s.workoutChecked(ctx, in.ID, in.OrgID); err != nil {
		return db.ProgramWorkout{}, err
	}
	return s.q.UpdateProgramWorkout(ctx, db.UpdateProgramWorkoutParams{
		ID:               in.ID,
		Name:             in.Name,
		Note:             textToPg(in.Note),
		IsDeload:         in.IsDeload,
		EstimatedMinutes: intToPgInt2(in.EstimatedMinutes),
	})
}

func (s *ProgramService) DeleteProgramWorkout(ctx context.Context, id, orgID uuid.UUID) error {
	if _, err := s.workoutChecked(ctx, id, orgID); err != nil {
		return err
	}
	return s.q.DeleteProgramWorkout(ctx, id)
}

// ---- ejercicios del dia (program_exercise) ----

type CreateProgramExerciseInput struct {
	ProgramWorkoutID  uuid.UUID
	OrgID             uuid.UUID
	ExerciseID        uuid.UUID
	OrderIndex        int
	SupersetGroup     *int
	TargetSets        int
	TargetRepsMin     *int
	TargetRepsMax     *int
	TargetRPE         *float64
	TargetPct1RM      *float64
	RestSeconds       int
	Tempo             string
	Note              string
	ProgressionRuleID *uuid.UUID
}

func (s *ProgramService) CreateProgramExercise(ctx context.Context, in CreateProgramExerciseInput) (db.ProgramExercise, error) {
	if _, err := s.workoutChecked(ctx, in.ProgramWorkoutID, in.OrgID); err != nil {
		return db.ProgramExercise{}, err
	}

	return s.q.CreateProgramExercise(ctx, db.CreateProgramExerciseParams{
		ProgramWorkoutID:  in.ProgramWorkoutID,
		ExerciseID:        in.ExerciseID,
		OrderIndex:        int16(in.OrderIndex),
		SupersetGroup:     intToPgInt2(in.SupersetGroup),
		TargetSets:        int16OrDefault(in.TargetSets, 3),
		TargetRepsMin:     intToPgInt2(in.TargetRepsMin),
		TargetRepsMax:     intToPgInt2(in.TargetRepsMax),
		TargetRpe:         in.TargetRPE,
		TargetPct1rm:      in.TargetPct1RM,
		RestSeconds:       int16OrDefault(in.RestSeconds, 90),
		Tempo:             textToPg(in.Tempo),
		Note:              textToPg(in.Note),
		ProgressionRuleID: uuidToPg(in.ProgressionRuleID),
	})
}

func (s *ProgramService) ListProgramExercises(ctx context.Context, workoutID, orgID uuid.UUID) ([]db.ProgramExercise, error) {
	if _, err := s.workoutChecked(ctx, workoutID, orgID); err != nil {
		return nil, err
	}
	return s.q.ListProgramExercises(ctx, workoutID)
}

func (s *ProgramService) exerciseChecked(ctx context.Context, id, orgID uuid.UUID) (db.GetProgramExerciseWithOrgRow, error) {
	row, err := s.q.GetProgramExerciseWithOrg(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return db.GetProgramExerciseWithOrgRow{}, domain.ErrNotFound
		}
		return db.GetProgramExerciseWithOrgRow{}, err
	}
	if row.ProgramOrgID != orgID {
		return db.GetProgramExerciseWithOrgRow{}, domain.ErrNotFound
	}
	return row, nil
}

type UpdateProgramExerciseInput struct {
	ID                uuid.UUID
	OrgID             uuid.UUID
	OrderIndex        int
	SupersetGroup     *int
	TargetSets        int
	TargetRepsMin     *int
	TargetRepsMax     *int
	TargetRPE         *float64
	TargetPct1RM      *float64
	RestSeconds       int
	Tempo             string
	Note              string
	ProgressionRuleID *uuid.UUID
}

func (s *ProgramService) UpdateProgramExercise(ctx context.Context, in UpdateProgramExerciseInput) (db.ProgramExercise, error) {
	if _, err := s.exerciseChecked(ctx, in.ID, in.OrgID); err != nil {
		return db.ProgramExercise{}, err
	}

	return s.q.UpdateProgramExercise(ctx, db.UpdateProgramExerciseParams{
		ID:                in.ID,
		OrderIndex:        int16(in.OrderIndex),
		SupersetGroup:     intToPgInt2(in.SupersetGroup),
		TargetSets:        int16OrDefault(in.TargetSets, 3),
		TargetRepsMin:     intToPgInt2(in.TargetRepsMin),
		TargetRepsMax:     intToPgInt2(in.TargetRepsMax),
		TargetRpe:         in.TargetRPE,
		TargetPct1rm:      in.TargetPct1RM,
		RestSeconds:       int16OrDefault(in.RestSeconds, 90),
		Tempo:             textToPg(in.Tempo),
		Note:              textToPg(in.Note),
		ProgressionRuleID: uuidToPg(in.ProgressionRuleID),
	})
}

func (s *ProgramService) DeleteProgramExercise(ctx context.Context, id, orgID uuid.UUID) error {
	if _, err := s.exerciseChecked(ctx, id, orgID); err != nil {
		return err
	}
	return s.q.DeleteProgramExercise(ctx, id)
}

func goalOrDefault(v string) db.TrainingGoal {
	if v == "" {
		return db.TrainingGoalGeneralHealth
	}
	return db.TrainingGoal(v)
}

func levelOrDefault(v string) db.ExperienceLevel {
	if v == "" {
		return db.ExperienceLevelBeginner
	}
	return db.ExperienceLevel(v)
}

func int16OrDefault(v int, fallback int16) int16 {
	if v <= 0 {
		return fallback
	}
	return int16(v)
}

func intToPgInt2(v *int) pgtype.Int2 {
	if v == nil {
		return pgtype.Int2{}
	}
	return pgtype.Int2{Int16: int16(*v), Valid: true}
}
