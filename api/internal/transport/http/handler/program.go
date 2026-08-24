package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"myvibesfit/api/internal/domain"
	"myvibesfit/api/internal/repository/db"
	"myvibesfit/api/internal/service"
	"myvibesfit/api/internal/transport/http/dto"
	"myvibesfit/api/internal/transport/http/middleware"
)

type ProgramHandler struct {
	svc *service.ProgramService
}

func NewProgramHandler(svc *service.ProgramService) *ProgramHandler {
	return &ProgramHandler{svc: svc}
}

// ---- reglas de progresion ----

func (h *ProgramHandler) ListProgressionRules(w http.ResponseWriter, r *http.Request) {
	var orgID *uuid.UUID
	if id, ok := middleware.OrgID(r.Context()); ok {
		orgID = &id
	}
	rules, err := h.svc.ListProgressionRules(r.Context(), orgID)
	if err != nil {
		writeError(w, err)
		return
	}
	out := make([]dto.ProgressionRuleDTO, len(rules))
	for i, rule := range rules {
		out[i] = progressionRuleDTO(rule)
	}
	writeJSON(w, http.StatusOK, out)
}

func (h *ProgramHandler) CreateProgressionRule(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateProgressionRuleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, domain.ErrInvalidInput)
		return
	}
	orgID, _ := middleware.OrgID(r.Context())

	rule, err := h.svc.CreateProgressionRule(r.Context(), service.CreateProgressionRuleInput{
		OrgID: &orgID, Name: req.Name, Type: req.Type, Params: req.Params,
	})
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, progressionRuleDTO(rule))
}

func progressionRuleDTO(r db.ProgressionRule) dto.ProgressionRuleDTO {
	return dto.ProgressionRuleDTO{
		ID: r.ID, OrgID: pgUUIDPtr(r.OrgID), Name: r.Name, Type: string(r.Type),
		Params: json.RawMessage(r.Params), IsSystem: r.IsSystem,
	}
}

// ---- programas ----

func (h *ProgramHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateProgramRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, domain.ErrInvalidInput)
		return
	}
	orgID, _ := middleware.OrgID(r.Context())
	userID, _ := middleware.UserID(r.Context())

	p, err := h.svc.CreateProgram(r.Context(), service.CreateProgramInput{
		OrgID: orgID, CreatedBy: userID, Name: req.Name, Description: req.Description,
		Goal: req.Goal, Level: req.Level, TotalWeeks: req.TotalWeeks, DaysPerWeek: req.DaysPerWeek,
	})
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, programDTO(p))
}

func (h *ProgramHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, domain.ErrInvalidInput)
		return
	}
	orgID, _ := middleware.OrgID(r.Context())

	p, err := h.svc.GetProgram(r.Context(), id, orgID)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, programDTO(p))
}

func (h *ProgramHandler) List(w http.ResponseWriter, r *http.Request) {
	orgID, _ := middleware.OrgID(r.Context())
	q := r.URL.Query()

	filter := service.ListProgramsFilter{OrgID: orgID, Status: q.Get("status")}
	if limit, err := strconv.Atoi(q.Get("limit")); err == nil {
		filter.Limit = int32(limit)
	}
	if offset, err := strconv.Atoi(q.Get("offset")); err == nil {
		filter.Offset = int32(offset)
	}

	programs, err := h.svc.ListPrograms(r.Context(), filter)
	if err != nil {
		writeError(w, err)
		return
	}
	out := make([]dto.ProgramDTO, len(programs))
	for i, p := range programs {
		out[i] = programDTO(p)
	}
	writeJSON(w, http.StatusOK, out)
}

func (h *ProgramHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, domain.ErrInvalidInput)
		return
	}
	var req dto.UpdateProgramRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, domain.ErrInvalidInput)
		return
	}
	orgID, _ := middleware.OrgID(r.Context())

	p, err := h.svc.UpdateProgram(r.Context(), service.UpdateProgramInput{
		ID: id, OrgID: orgID, Name: req.Name, Description: req.Description,
		Goal: req.Goal, Level: req.Level, TotalWeeks: req.TotalWeeks, DaysPerWeek: req.DaysPerWeek,
	})
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, programDTO(p))
}

func (h *ProgramHandler) Publish(w http.ResponseWriter, r *http.Request) {
	h.setStatus(w, r, db.ProgramStatusPublished)
}

func (h *ProgramHandler) Archive(w http.ResponseWriter, r *http.Request) {
	h.setStatus(w, r, db.ProgramStatusArchived)
}

func (h *ProgramHandler) setStatus(w http.ResponseWriter, r *http.Request, status db.ProgramStatus) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, domain.ErrInvalidInput)
		return
	}
	orgID, _ := middleware.OrgID(r.Context())

	p, err := h.svc.SetProgramStatus(r.Context(), id, orgID, status)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, programDTO(p))
}

func programDTO(p db.Program) dto.ProgramDTO {
	return dto.ProgramDTO{
		ID: p.ID, OrgID: p.OrgID, Name: p.Name, Description: pgText(p.Description),
		Goal: string(p.Goal), Level: string(p.Level), TotalWeeks: int(p.TotalWeeks),
		DaysPerWeek: int(p.DaysPerWeek), Status: string(p.Status),
	}
}

// ---- dias del programa ----

func (h *ProgramHandler) CreateWorkout(w http.ResponseWriter, r *http.Request) {
	programID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, domain.ErrInvalidInput)
		return
	}
	var req dto.CreateProgramWorkoutRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, domain.ErrInvalidInput)
		return
	}
	orgID, _ := middleware.OrgID(r.Context())

	pw, err := h.svc.CreateProgramWorkout(r.Context(), service.CreateProgramWorkoutInput{
		ProgramID: programID, OrgID: orgID, WeekNumber: req.WeekNumber, DayIndex: req.DayIndex,
		Name: req.Name, Note: req.Note, IsDeload: req.IsDeload, EstimatedMinutes: req.EstimatedMinutes,
	})
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, programWorkoutDTO(pw))
}

func (h *ProgramHandler) ListWorkouts(w http.ResponseWriter, r *http.Request) {
	programID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, domain.ErrInvalidInput)
		return
	}
	orgID, _ := middleware.OrgID(r.Context())

	workouts, err := h.svc.ListProgramWorkouts(r.Context(), programID, orgID)
	if err != nil {
		writeError(w, err)
		return
	}
	out := make([]dto.ProgramWorkoutDTO, len(workouts))
	for i, pw := range workouts {
		out[i] = programWorkoutDTO(pw)
	}
	writeJSON(w, http.StatusOK, out)
}

func (h *ProgramHandler) UpdateWorkout(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, domain.ErrInvalidInput)
		return
	}
	var req dto.UpdateProgramWorkoutRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, domain.ErrInvalidInput)
		return
	}
	orgID, _ := middleware.OrgID(r.Context())

	pw, err := h.svc.UpdateProgramWorkout(r.Context(), service.UpdateProgramWorkoutInput{
		ID: id, OrgID: orgID, Name: req.Name, Note: req.Note,
		IsDeload: req.IsDeload, EstimatedMinutes: req.EstimatedMinutes,
	})
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, programWorkoutDTO(pw))
}

func (h *ProgramHandler) DeleteWorkout(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, domain.ErrInvalidInput)
		return
	}
	orgID, _ := middleware.OrgID(r.Context())

	if err := h.svc.DeleteProgramWorkout(r.Context(), id, orgID); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func programWorkoutDTO(pw db.ProgramWorkout) dto.ProgramWorkoutDTO {
	return dto.ProgramWorkoutDTO{
		ID: pw.ID, ProgramID: pw.ProgramID, WeekNumber: int(pw.WeekNumber), DayIndex: int(pw.DayIndex),
		Name: pw.Name, Note: pgText(pw.Note), IsDeload: pw.IsDeload, EstimatedMinutes: pgInt2Ptr(pw.EstimatedMinutes),
	}
}

// ---- ejercicios del dia ----

func (h *ProgramHandler) CreateExercise(w http.ResponseWriter, r *http.Request) {
	workoutID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, domain.ErrInvalidInput)
		return
	}
	var req dto.CreateProgramExerciseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, domain.ErrInvalidInput)
		return
	}
	orgID, _ := middleware.OrgID(r.Context())

	pe, err := h.svc.CreateProgramExercise(r.Context(), service.CreateProgramExerciseInput{
		ProgramWorkoutID: workoutID, OrgID: orgID, ExerciseID: req.ExerciseID, OrderIndex: req.OrderIndex,
		SupersetGroup: req.SupersetGroup, TargetSets: req.TargetSets, TargetRepsMin: req.TargetRepsMin,
		TargetRepsMax: req.TargetRepsMax, TargetRPE: req.TargetRPE, TargetPct1RM: req.TargetPct1RM,
		RestSeconds: req.RestSeconds, Tempo: req.Tempo, Note: req.Note, ProgressionRuleID: req.ProgressionRuleID,
	})
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, programExerciseDTO(pe))
}

func (h *ProgramHandler) ListExercises(w http.ResponseWriter, r *http.Request) {
	workoutID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, domain.ErrInvalidInput)
		return
	}
	orgID, _ := middleware.OrgID(r.Context())

	exercises, err := h.svc.ListProgramExercises(r.Context(), workoutID, orgID)
	if err != nil {
		writeError(w, err)
		return
	}
	out := make([]dto.ProgramExerciseDTO, len(exercises))
	for i, pe := range exercises {
		out[i] = programExerciseDTO(pe)
	}
	writeJSON(w, http.StatusOK, out)
}

func (h *ProgramHandler) UpdateExercise(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, domain.ErrInvalidInput)
		return
	}
	var req dto.UpdateProgramExerciseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, domain.ErrInvalidInput)
		return
	}
	orgID, _ := middleware.OrgID(r.Context())

	pe, err := h.svc.UpdateProgramExercise(r.Context(), service.UpdateProgramExerciseInput{
		ID: id, OrgID: orgID, OrderIndex: req.OrderIndex, SupersetGroup: req.SupersetGroup,
		TargetSets: req.TargetSets, TargetRepsMin: req.TargetRepsMin, TargetRepsMax: req.TargetRepsMax,
		TargetRPE: req.TargetRPE, TargetPct1RM: req.TargetPct1RM, RestSeconds: req.RestSeconds,
		Tempo: req.Tempo, Note: req.Note, ProgressionRuleID: req.ProgressionRuleID,
	})
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, programExerciseDTO(pe))
}

func (h *ProgramHandler) DeleteExercise(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, domain.ErrInvalidInput)
		return
	}
	orgID, _ := middleware.OrgID(r.Context())

	if err := h.svc.DeleteProgramExercise(r.Context(), id, orgID); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func programExerciseDTO(pe db.ProgramExercise) dto.ProgramExerciseDTO {
	return dto.ProgramExerciseDTO{
		ID: pe.ID, ProgramWorkoutID: pe.ProgramWorkoutID, ExerciseID: pe.ExerciseID,
		OrderIndex: int(pe.OrderIndex), SupersetGroup: pgInt2Ptr(pe.SupersetGroup),
		TargetSets: int(pe.TargetSets), TargetRepsMin: pgInt2Ptr(pe.TargetRepsMin),
		TargetRepsMax: pgInt2Ptr(pe.TargetRepsMax), TargetRPE: pe.TargetRpe, TargetPct1RM: pe.TargetPct1rm,
		RestSeconds: int(pe.RestSeconds), Tempo: pgText(pe.Tempo), Note: pgText(pe.Note),
		ProgressionRuleID: pgUUIDPtr(pe.ProgressionRuleID),
	}
}
