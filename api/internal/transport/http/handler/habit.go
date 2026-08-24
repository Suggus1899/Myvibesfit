package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"myvibesfit/api/internal/domain"
	"myvibesfit/api/internal/service"
	"myvibesfit/api/internal/transport/http/dto"
	"myvibesfit/api/internal/transport/http/middleware"
)

type HabitHandler struct {
	svc *service.HabitService
}

func NewHabitHandler(svc *service.HabitService) *HabitHandler {
	return &HabitHandler{svc: svc}
}

func (h *HabitHandler) List(w http.ResponseWriter, r *http.Request) {
	var orgID *uuid.UUID
	if id, ok := middleware.OrgID(r.Context()); ok {
		orgID = &id
	}
	habits, err := h.svc.ListHabits(r.Context(), orgID)
	if err != nil {
		writeError(w, err)
		return
	}
	out := make([]dto.HabitDTO, len(habits))
	for i, hb := range habits {
		out[i] = dto.HabitDTO{
			ID: hb.ID, Slug: hb.Slug, Name: hb.Name, Icon: hb.Icon,
			Unit: string(hb.Unit), DefaultTarget: hb.DefaultTarget, IsSystem: hb.IsSystem,
		}
	}
	writeJSON(w, http.StatusOK, out)
}

func (h *HabitHandler) MyHabits(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserID(r.Context())
	if !ok {
		writeError(w, domain.ErrUnauthorized)
		return
	}
	rows, err := h.svc.MyHabits(r.Context(), userID)
	if err != nil {
		writeError(w, err)
		return
	}
	out := make([]dto.MyHabitDTO, len(rows))
	for i, row := range rows {
		days := make([]int, len(row.DaysOfWeek))
		for j, d := range row.DaysOfWeek {
			days[j] = int(d)
		}
		out[i] = dto.MyHabitDTO{
			ID: row.ID, HabitID: row.HabitID, Name: row.HabitName, Icon: row.HabitIcon,
			Unit: string(row.HabitUnit), TargetValue: row.TargetValue, Frequency: string(row.Frequency), DaysOfWeek: days,
		}
	}
	writeJSON(w, http.StatusOK, out)
}

func (h *HabitHandler) Subscribe(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserID(r.Context())
	if !ok {
		writeError(w, domain.ErrUnauthorized)
		return
	}
	var req dto.SubscribeHabitRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.HabitID == uuid.Nil {
		writeError(w, domain.ErrInvalidInput)
		return
	}

	ch, err := h.svc.Subscribe(r.Context(), service.SubscribeHabitInput{
		UserID: userID, HabitID: req.HabitID, TargetValue: req.TargetValue,
		Frequency: req.Frequency, DaysOfWeek: req.DaysOfWeek,
	})
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"id": ch.ID, "habit_id": ch.HabitID})
}

func (h *HabitHandler) Unsubscribe(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserID(r.Context())
	if !ok {
		writeError(w, domain.ErrUnauthorized)
		return
	}
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, domain.ErrInvalidInput)
		return
	}
	if err := h.svc.Unsubscribe(r.Context(), id, userID); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *HabitHandler) Log(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserID(r.Context())
	if !ok {
		writeError(w, domain.ErrUnauthorized)
		return
	}
	clientHabitID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, domain.ErrInvalidInput)
		return
	}
	var req dto.LogHabitRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, domain.ErrInvalidInput)
		return
	}
	logDate, err := parseDate(req.LogDate)
	if err != nil {
		writeError(w, domain.ErrInvalidInput)
		return
	}
	if req.ClientLocalID == uuid.Nil {
		writeError(w, domain.ErrInvalidInput)
		return
	}

	result, err := h.svc.LogHabit(r.Context(), service.LogHabitInput{
		ClientLocalID: req.ClientLocalID, ClientHabitID: clientHabitID, UserID: userID,
		LogDate: logDate, Value: req.Value, IsCompleted: req.IsCompleted,
	})
	if err != nil {
		writeError(w, err)
		return
	}

	resp := dto.HabitLogResponse{
		Log: dto.HabitLogDTO{
			ID: result.Log.ID, LogDate: result.Log.LogDate.Format(dateLayout),
			Value: result.Log.Value, IsCompleted: result.Log.IsCompleted,
		},
	}
	for _, a := range result.UnlockedAchievements {
		resp.UnlockedAchievements = append(resp.UnlockedAchievements, achievementDTO(a))
	}
	writeJSON(w, http.StatusOK, resp)
}
