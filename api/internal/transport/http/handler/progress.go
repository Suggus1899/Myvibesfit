package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"myvibesfit/api/internal/domain"
	"myvibesfit/api/internal/repository/db"
	"myvibesfit/api/internal/service"
	"myvibesfit/api/internal/transport/http/dto"
	"myvibesfit/api/internal/transport/http/middleware"
)

type ProgressHandler struct {
	svc *service.ProgressService
}

func NewProgressHandler(svc *service.ProgressService) *ProgressHandler {
	return &ProgressHandler{svc: svc}
}

func (h *ProgressHandler) ExerciseHistory(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserID(r.Context())
	if !ok {
		writeError(w, domain.ErrUnauthorized)
		return
	}
	exerciseID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, domain.ErrInvalidInput)
		return
	}

	since := sinceFromQuery(r, 180)
	var limit int32
	if v, err := strconv.Atoi(r.URL.Query().Get("limit")); err == nil {
		limit = int32(v)
	}

	logs, err := h.svc.ExerciseHistory(r.Context(), userID, exerciseID, since, limit)
	if err != nil {
		writeError(w, err)
		return
	}
	out := make([]dto.SetLogDTO, len(logs))
	for i, l := range logs {
		out[i] = dto.SetLogDTO{
			ID: l.ID, ExerciseID: l.ExerciseID, SetNumber: int(l.SetNumber), Type: string(l.Type),
			WeightKg: l.WeightKg, Reps: pgInt2Ptr(l.Reps), RPE: l.Rpe, PerformedAt: l.PerformedAt,
		}
	}
	writeJSON(w, http.StatusOK, out)
}

func (h *ProgressHandler) Records(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserID(r.Context())
	if !ok {
		writeError(w, domain.ErrUnauthorized)
		return
	}
	var exerciseID *uuid.UUID
	if raw := r.URL.Query().Get("exercise_id"); raw != "" {
		id, err := uuid.Parse(raw)
		if err != nil {
			writeError(w, domain.ErrInvalidInput)
			return
		}
		exerciseID = &id
	}

	records, err := h.svc.Records(r.Context(), userID, exerciseID)
	if err != nil {
		writeError(w, err)
		return
	}
	out := make([]dto.PersonalRecordDTO, len(records))
	for i, rec := range records {
		out[i] = dto.PersonalRecordDTO{
			ID: rec.ID, ExerciseID: rec.ExerciseID, Type: string(rec.Type), Value: rec.Value, AchievedAt: rec.AchievedAt,
		}
	}
	writeJSON(w, http.StatusOK, out)
}

func (h *ProgressHandler) Volume(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserID(r.Context())
	if !ok {
		writeError(w, domain.ErrUnauthorized)
		return
	}
	since := sinceFromQuery(r, 90)

	sessions, err := h.svc.Volume(r.Context(), userID, since)
	if err != nil {
		writeError(w, err)
		return
	}
	out := make([]dto.VolumePointDTO, len(sessions))
	for i, s := range sessions {
		out[i] = dto.VolumePointDTO{
			SessionID: s.ID, StartedAt: s.StartedAt, TotalVolumeKg: s.TotalVolumeKg, DurationSeconds: pgInt4Ptr(s.DurationSeconds),
		}
	}
	writeJSON(w, http.StatusOK, out)
}

func (h *ProgressHandler) LogBodyMetric(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserID(r.Context())
	if !ok {
		writeError(w, domain.ErrUnauthorized)
		return
	}
	var req dto.LogBodyMetricRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, domain.ErrInvalidInput)
		return
	}
	measuredOn, err := parseDate(req.MeasuredOn)
	if err != nil {
		writeError(w, domain.ErrInvalidInput)
		return
	}

	m, err := h.svc.LogBodyMetric(r.Context(), service.BodyMetricInput{
		UserID: userID, MeasuredOn: measuredOn, WeightKg: req.WeightKg, BodyFatPct: req.BodyFatPct, Note: req.Note,
	})
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, bodyMetricDTO(m))
}

func (h *ProgressHandler) ListBodyMetrics(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserID(r.Context())
	if !ok {
		writeError(w, domain.ErrUnauthorized)
		return
	}
	var limit int32
	if v, err := strconv.Atoi(r.URL.Query().Get("limit")); err == nil {
		limit = int32(v)
	}

	metrics, err := h.svc.BodyMetrics(r.Context(), userID, limit)
	if err != nil {
		writeError(w, err)
		return
	}
	out := make([]dto.BodyMetricDTO, len(metrics))
	for i, m := range metrics {
		out[i] = bodyMetricDTO(m)
	}
	writeJSON(w, http.StatusOK, out)
}

func bodyMetricDTO(m db.BodyMetric) dto.BodyMetricDTO {
	return dto.BodyMetricDTO{
		MeasuredOn: m.MeasuredOn.Format(dateLayout), WeightKg: m.WeightKg, BodyFatPct: m.BodyFatPct, Note: pgText(m.Note),
	}
}

func sinceFromQuery(r *http.Request, defaultDays int) time.Time {
	days := defaultDays
	if v, err := strconv.Atoi(r.URL.Query().Get("days")); err == nil && v > 0 {
		days = v
	}
	return time.Now().AddDate(0, 0, -days)
}
