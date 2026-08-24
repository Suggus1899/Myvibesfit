package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"myvibesfit/api/internal/domain"
	"myvibesfit/api/internal/repository/db"
	"myvibesfit/api/internal/service"
	"myvibesfit/api/internal/transport/http/dto"
	"myvibesfit/api/internal/transport/http/middleware"
)

type AssignmentHandler struct {
	svc *service.AssignmentService
}

func NewAssignmentHandler(svc *service.AssignmentService) *AssignmentHandler {
	return &AssignmentHandler{svc: svc}
}

func (h *AssignmentHandler) Assign(w http.ResponseWriter, r *http.Request) {
	programID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, domain.ErrInvalidInput)
		return
	}
	var req dto.AssignRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.ClientUserID == uuid.Nil {
		writeError(w, domain.ErrInvalidInput)
		return
	}
	startDate, err := parseDate(req.StartDate)
	if err != nil {
		writeError(w, domain.ErrInvalidInput)
		return
	}

	orgID, _ := middleware.OrgID(r.Context())
	coachID, _ := middleware.UserID(r.Context())

	a, err := h.svc.Assign(r.Context(), service.AssignInput{
		OrgID: orgID, ProgramID: programID, ClientUserID: req.ClientUserID,
		CoachUserID: coachID, StartDate: startDate,
	})
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, assignmentDTO(a))
}

func (h *AssignmentHandler) Cancel(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, domain.ErrInvalidInput)
		return
	}
	orgID, _ := middleware.OrgID(r.Context())

	a, err := h.svc.Cancel(r.Context(), id, orgID)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, assignmentDTO(a))
}

func (h *AssignmentHandler) CurrentForMe(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserID(r.Context())
	if !ok {
		writeError(w, domain.ErrUnauthorized)
		return
	}

	detail, err := h.svc.GetCurrentForClient(r.Context(), userID)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, assignmentDetailDTO(detail))
}

func assignmentDTO(a db.Assignment) dto.AssignmentDTO {
	return dto.AssignmentDTO{
		ID: a.ID, ProgramID: pgUUIDPtr(a.ProgramID), ClientUserID: a.ClientUserID,
		CoachUserID: pgUUIDPtr(a.CoachUserID), Name: a.Name, StartDate: a.StartDate.Format(dateLayout),
		EndDate: datePtrToString(a.EndDate), Status: string(a.Status),
	}
}

func assignmentDetailDTO(d service.AssignmentDetail) dto.AssignmentDetailDTO {
	out := dto.AssignmentDetailDTO{
		Assignment: assignmentDTO(d.Assignment),
		Workouts:   make([]dto.AssignedWorkoutDTO, len(d.Workouts)),
	}
	for i, w := range d.Workouts {
		exercises := make([]dto.AssignedExerciseDTO, len(w.Exercises))
		for j, e := range w.Exercises {
			exercises[j] = dto.AssignedExerciseDTO{
				ID: e.ID, ExerciseID: e.ExerciseID, OrderIndex: int(e.OrderIndex),
				SupersetGroup: pgInt2Ptr(e.SupersetGroup), TargetSets: int(e.TargetSets),
				TargetRepsMin: pgInt2Ptr(e.TargetRepsMin), TargetRepsMax: pgInt2Ptr(e.TargetRepsMax),
				TargetRPE: e.TargetRpe, TargetWeightKg: e.TargetWeightKg, RestSeconds: int(e.RestSeconds),
				Tempo: pgText(e.Tempo), Note: pgText(e.Note),
			}
		}
		out.Workouts[i] = dto.AssignedWorkoutDTO{
			ID: w.Workout.ID, WeekNumber: int(w.Workout.WeekNumber), DayIndex: int(w.Workout.DayIndex),
			Name: w.Workout.Name, Note: pgText(w.Workout.Note), IsDeload: w.Workout.IsDeload,
			ScheduledOn: datePtrToString(w.Workout.ScheduledOn), Status: string(w.Workout.Status),
			Exercises: exercises,
		}
	}
	return out
}
