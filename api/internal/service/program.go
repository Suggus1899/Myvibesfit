package service

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"

	"myvibesfit/api/internal/domain"
)

type ProgramService struct {
	repo  domain.ProgramRepository
	audit *AuditLogger
}

func NewProgramService(repo domain.ProgramRepository, audit *AuditLogger) *ProgramService {
	return &ProgramService{repo: repo, audit: audit}
}

// ---- reglas de progresion ----

func (s *ProgramService) ListProgressionRules(ctx context.Context, orgID *uuid.UUID) ([]domain.ProgressionRule, error) {
	return s.repo.ListProgressionRules(ctx, orgID)
}

type CreateProgressionRuleInput struct {
	OrgID  *uuid.UUID
	Name   string
	Type   string
	Params json.RawMessage
}

func (s *ProgramService) CreateProgressionRule(ctx context.Context, in CreateProgressionRuleInput) (domain.ProgressionRule, error) {
	if in.Name == "" || in.Type == "" {
		return domain.ProgressionRule{}, domain.ErrInvalidInput
	}
	params := in.Params
	if len(params) == 0 {
		params = []byte("{}")
	}
	return s.repo.CreateProgressionRule(ctx, domain.ProgressionRule{
		OrgID: in.OrgID, Name: in.Name, Type: in.Type, Params: params, IsSystem: false,
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

func (s *ProgramService) CreateProgram(ctx context.Context, in CreateProgramInput) (domain.Program, error) {
	if in.Name == "" {
		return domain.Program{}, domain.ErrInvalidInput
	}
	return s.repo.Create(ctx, domain.Program{
		OrgID: in.OrgID, CreatedBy: in.CreatedBy, Name: in.Name, Description: in.Description,
		Goal: stringOrDefault(in.Goal, domain.DefaultTrainingGoal), Level: stringOrDefault(in.Level, domain.DefaultExperienceLevel),
		TotalWeeks: int16OrDefault(in.TotalWeeks, 4), DaysPerWeek: int16OrDefault(in.DaysPerWeek, 3),
	})
}

func (s *ProgramService) GetProgram(ctx context.Context, id, orgID uuid.UUID) (domain.Program, error) {
	return s.repo.GetByID(ctx, id, orgID)
}

type ListProgramsFilter struct {
	OrgID  uuid.UUID
	Status string
	Limit  int32
	Offset int32
}

func (s *ProgramService) ListPrograms(ctx context.Context, f ListProgramsFilter) ([]domain.Program, error) {
	limit := f.Limit
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	return s.repo.ListByOrg(ctx, domain.ListProgramsFilter{OrgID: f.OrgID, Status: f.Status, Limit: limit, Offset: f.Offset})
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

func (s *ProgramService) UpdateProgram(ctx context.Context, in UpdateProgramInput) (domain.Program, error) {
	if in.Name == "" {
		return domain.Program{}, domain.ErrInvalidInput
	}
	return s.repo.Update(ctx, domain.Program{
		ID: in.ID, OrgID: in.OrgID, Name: in.Name, Description: in.Description,
		Goal: stringOrDefault(in.Goal, domain.DefaultTrainingGoal), Level: stringOrDefault(in.Level, domain.DefaultExperienceLevel),
		TotalWeeks: int16OrDefault(in.TotalWeeks, 4), DaysPerWeek: int16OrDefault(in.DaysPerWeek, 3),
	})
}

func (s *ProgramService) SetProgramStatus(ctx context.Context, id, orgID, actorID uuid.UUID, status domain.ProgramStatus) (domain.Program, error) {
	p, err := s.repo.SetStatus(ctx, id, orgID, status)
	if err != nil {
		return domain.Program{}, err
	}
	s.audit.Log(ctx, orgID, actorID, "program.set_status", "program", id.String(), map[string]any{"status": status})
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

func (s *ProgramService) CreateProgramWorkout(ctx context.Context, in CreateProgramWorkoutInput) (domain.ProgramWorkout, error) {
	if in.Name == "" || in.WeekNumber < 1 || in.DayIndex < 1 || in.DayIndex > 7 {
		return domain.ProgramWorkout{}, domain.ErrInvalidInput
	}
	if _, err := s.GetProgram(ctx, in.ProgramID, in.OrgID); err != nil {
		return domain.ProgramWorkout{}, err
	}

	return s.repo.CreateWorkout(ctx, domain.ProgramWorkout{
		ProgramID: in.ProgramID, WeekNumber: int16(in.WeekNumber), DayIndex: int16(in.DayIndex),
		Name: in.Name, Note: in.Note, IsDeload: in.IsDeload, EstimatedMinutes: in.EstimatedMinutes,
	})
}

func (s *ProgramService) ListProgramWorkouts(ctx context.Context, programID, orgID uuid.UUID) ([]domain.ProgramWorkout, error) {
	if _, err := s.GetProgram(ctx, programID, orgID); err != nil {
		return nil, err
	}
	return s.repo.ListWorkouts(ctx, programID)
}

func (s *ProgramService) workoutChecked(ctx context.Context, id, orgID uuid.UUID) error {
	workoutOrgID, err := s.repo.GetWorkoutOrgID(ctx, id)
	if err != nil {
		return err
	}
	if workoutOrgID != orgID {
		return domain.ErrNotFound
	}
	return nil
}

type UpdateProgramWorkoutInput struct {
	ID               uuid.UUID
	OrgID            uuid.UUID
	Name             string
	Note             string
	IsDeload         bool
	EstimatedMinutes *int
}

func (s *ProgramService) UpdateProgramWorkout(ctx context.Context, in UpdateProgramWorkoutInput) (domain.ProgramWorkout, error) {
	if in.Name == "" {
		return domain.ProgramWorkout{}, domain.ErrInvalidInput
	}
	if err := s.workoutChecked(ctx, in.ID, in.OrgID); err != nil {
		return domain.ProgramWorkout{}, err
	}
	return s.repo.UpdateWorkout(ctx, domain.ProgramWorkout{
		ID: in.ID, Name: in.Name, Note: in.Note, IsDeload: in.IsDeload, EstimatedMinutes: in.EstimatedMinutes,
	})
}

func (s *ProgramService) DeleteProgramWorkout(ctx context.Context, id, orgID uuid.UUID) error {
	if err := s.workoutChecked(ctx, id, orgID); err != nil {
		return err
	}
	return s.repo.DeleteWorkout(ctx, id)
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

func (s *ProgramService) CreateProgramExercise(ctx context.Context, in CreateProgramExerciseInput) (domain.ProgramExercise, error) {
	if err := s.workoutChecked(ctx, in.ProgramWorkoutID, in.OrgID); err != nil {
		return domain.ProgramExercise{}, err
	}

	return s.repo.CreateExercise(ctx, domain.ProgramExercise{
		ProgramWorkoutID: in.ProgramWorkoutID, ExerciseID: in.ExerciseID, OrderIndex: int16(in.OrderIndex),
		SupersetGroup: in.SupersetGroup, TargetSets: int16OrDefault(in.TargetSets, 3), TargetRepsMin: in.TargetRepsMin,
		TargetRepsMax: in.TargetRepsMax, TargetRPE: in.TargetRPE, TargetPct1RM: in.TargetPct1RM,
		RestSeconds: int16OrDefault(in.RestSeconds, 90), Tempo: in.Tempo, Note: in.Note, ProgressionRuleID: in.ProgressionRuleID,
	})
}

func (s *ProgramService) ListProgramExercises(ctx context.Context, workoutID, orgID uuid.UUID) ([]domain.ProgramExercise, error) {
	if err := s.workoutChecked(ctx, workoutID, orgID); err != nil {
		return nil, err
	}
	return s.repo.ListExercises(ctx, workoutID)
}

func (s *ProgramService) exerciseChecked(ctx context.Context, id, orgID uuid.UUID) error {
	exerciseOrgID, err := s.repo.GetExerciseOrgID(ctx, id)
	if err != nil {
		return err
	}
	if exerciseOrgID != orgID {
		return domain.ErrNotFound
	}
	return nil
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

func (s *ProgramService) UpdateProgramExercise(ctx context.Context, in UpdateProgramExerciseInput) (domain.ProgramExercise, error) {
	if err := s.exerciseChecked(ctx, in.ID, in.OrgID); err != nil {
		return domain.ProgramExercise{}, err
	}

	return s.repo.UpdateExercise(ctx, domain.ProgramExercise{
		ID: in.ID, OrderIndex: int16(in.OrderIndex), SupersetGroup: in.SupersetGroup, TargetSets: int16OrDefault(in.TargetSets, 3),
		TargetRepsMin: in.TargetRepsMin, TargetRepsMax: in.TargetRepsMax, TargetRPE: in.TargetRPE, TargetPct1RM: in.TargetPct1RM,
		RestSeconds: int16OrDefault(in.RestSeconds, 90), Tempo: in.Tempo, Note: in.Note, ProgressionRuleID: in.ProgressionRuleID,
	})
}

func (s *ProgramService) DeleteProgramExercise(ctx context.Context, id, orgID uuid.UUID) error {
	if err := s.exerciseChecked(ctx, id, orgID); err != nil {
		return err
	}
	return s.repo.DeleteExercise(ctx, id)
}

func int16OrDefault(v int, fallback int16) int16 {
	if v <= 0 {
		return fallback
	}
	return int16(v)
}
