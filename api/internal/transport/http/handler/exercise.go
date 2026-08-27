package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"myvibesfit/api/internal/domain"
	"myvibesfit/api/internal/service"
	"myvibesfit/api/internal/transport/http/dto"
	"myvibesfit/api/internal/transport/http/middleware"
)

type ExerciseHandler struct {
	svc *service.ExerciseService
}

func NewExerciseHandler(svc *service.ExerciseService) *ExerciseHandler {
	return &ExerciseHandler{svc: svc}
}

func (h *ExerciseHandler) List(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	filter := domain.ExerciseFilter{
		Pattern: q.Get("pattern"),
		Muscle:  q.Get("muscle"),
	}
	if orgID, ok := middleware.OrgID(r.Context()); ok {
		filter.OrgID = &orgID
	}
	if limit, err := strconv.Atoi(q.Get("limit")); err == nil {
		filter.Limit = int32(limit)
	}
	if offset, err := strconv.Atoi(q.Get("offset")); err == nil {
		filter.Offset = int32(offset)
	}

	exercises, err := h.svc.List(r.Context(), filter)
	if err != nil {
		writeError(w, err)
		return
	}

	out := make([]dto.ExerciseDTO, len(exercises))
	for i, e := range exercises {
		out[i] = exerciseDTO(e)
	}
	writeJSON(w, http.StatusOK, out)
}

func (h *ExerciseHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, domain.ErrInvalidInput)
		return
	}
	ex, err := h.svc.Get(r.Context(), id)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, exerciseDTO(ex))
}

func (h *ExerciseHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateExerciseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, domain.ErrInvalidInput)
		return
	}

	userID, _ := middleware.UserID(r.Context())
	in := service.CreateExerciseInput{
		Slug: req.Slug, Name: req.Name, Description: req.Description,
		Instructions: req.Instructions, Pattern: req.Pattern, Mechanic: req.Mechanic,
		PrimaryMuscle: req.PrimaryMuscle, SecondaryMuscles: req.SecondaryMuscles,
		Equipment: req.Equipment, Difficulty: req.Difficulty, Tracking: req.Tracking,
		IsUnilateral: req.IsUnilateral, VideoURL: req.VideoURL, ThumbnailURL: req.ThumbnailURL,
		CreatedBy: userID,
	}
	if orgID, ok := middleware.OrgID(r.Context()); ok {
		in.OrgID = &orgID
	}

	ex, err := h.svc.Create(r.Context(), in)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, exerciseDTO(ex))
}

func (h *ExerciseHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, domain.ErrInvalidInput)
		return
	}
	orgID, ok := middleware.OrgID(r.Context())
	if !ok {
		writeError(w, domain.ErrForbidden)
		return
	}
	var req dto.UpdateExerciseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, domain.ErrInvalidInput)
		return
	}

	ex, err := h.svc.Update(r.Context(), service.UpdateExerciseInput{
		ID: id, OrgID: orgID, Name: req.Name, Description: req.Description, Instructions: req.Instructions,
		Pattern: req.Pattern, Mechanic: req.Mechanic, PrimaryMuscle: req.PrimaryMuscle,
		SecondaryMuscles: req.SecondaryMuscles, Equipment: req.Equipment, Difficulty: req.Difficulty,
		Tracking: req.Tracking, IsUnilateral: req.IsUnilateral, VideoURL: req.VideoURL, ThumbnailURL: req.ThumbnailURL,
	})
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, exerciseDTO(ex))
}

func (h *ExerciseHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, domain.ErrInvalidInput)
		return
	}
	orgID, ok := middleware.OrgID(r.Context())
	if !ok {
		writeError(w, domain.ErrForbidden)
		return
	}
	if err := h.svc.Deactivate(r.Context(), id, orgID); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func exerciseDTO(e domain.Exercise) dto.ExerciseDTO {
	return dto.ExerciseDTO{
		ID: e.ID, OrgID: e.OrgID, Slug: e.Slug, Name: e.Name, Description: e.Description, Instructions: e.Instructions,
		Pattern: e.Pattern, Mechanic: e.Mechanic, PrimaryMuscle: e.PrimaryMuscle,
		SecondaryMuscles: e.SecondaryMuscles, Equipment: e.Equipment,
		Difficulty: e.Difficulty, Tracking: e.Tracking, IsUnilateral: e.IsUnilateral,
		VideoURL: e.VideoURL, ThumbnailURL: e.ThumbnailURL,
	}
}
