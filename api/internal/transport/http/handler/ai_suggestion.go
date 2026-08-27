package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"myvibesfit/api/internal/domain"
	"myvibesfit/api/internal/service"
	"myvibesfit/api/internal/transport/http/dto"
	"myvibesfit/api/internal/transport/http/middleware"
)

type AISuggestionHandler struct {
	svc *service.AISuggestionService
}

func NewAISuggestionHandler(svc *service.AISuggestionService) *AISuggestionHandler {
	return &AISuggestionHandler{svc: svc}
}

func (h *AISuggestionHandler) ListPending(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserID(r.Context())
	if !ok {
		writeError(w, domain.ErrUnauthorized)
		return
	}
	orgID, ok := middleware.OrgID(r.Context())
	if !ok {
		writeError(w, domain.ErrUnauthorized)
		return
	}

	rows, err := h.svc.ListPending(r.Context(), userID, orgID)
	if err != nil {
		writeError(w, err)
		return
	}
	out := make([]dto.AISuggestionDTO, len(rows))
	for i, s := range rows {
		out[i] = dto.AISuggestionDTO{
			ID: s.ID, ClientUserID: s.ClientUserID, ClientName: s.ClientName,
			AssignmentID: s.AssignmentID, Kind: s.Kind, Payload: s.Payload,
			Rationale: s.Rationale, Confidence: s.Confidence, Model: s.Model,
			Status: string(s.Status), CreatedAt: s.CreatedAt, ExpiresAt: s.ExpiresAt,
		}
	}
	writeJSON(w, http.StatusOK, out)
}

func (h *AISuggestionHandler) Approve(w http.ResponseWriter, r *http.Request) {
	h.review(w, r, true)
}

func (h *AISuggestionHandler) Reject(w http.ResponseWriter, r *http.Request) {
	h.review(w, r, false)
}

func (h *AISuggestionHandler) review(w http.ResponseWriter, r *http.Request, approve bool) {
	userID, ok := middleware.UserID(r.Context())
	if !ok {
		writeError(w, domain.ErrUnauthorized)
		return
	}
	orgID, ok := middleware.OrgID(r.Context())
	if !ok {
		writeError(w, domain.ErrUnauthorized)
		return
	}
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, domain.ErrInvalidInput)
		return
	}

	sug, err := h.svc.Review(r.Context(), id, orgID, userID, approve)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, dto.AISuggestionDTO{
		ID: sug.ID, ClientUserID: sug.ClientUserID, AssignmentID: sug.AssignmentID,
		Kind: sug.Kind, Payload: sug.Payload, Rationale: sug.Rationale,
		Confidence: sug.Confidence, Model: sug.Model, Status: string(sug.Status),
		CreatedAt: sug.CreatedAt, ExpiresAt: sug.ExpiresAt,
	})
}
