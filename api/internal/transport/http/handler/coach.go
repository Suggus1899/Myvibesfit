package handler

import (
	"net/http"

	"myvibesfit/api/internal/domain"
	"myvibesfit/api/internal/service"
	"myvibesfit/api/internal/transport/http/dto"
	"myvibesfit/api/internal/transport/http/middleware"
)

type CoachHandler struct {
	svc *service.CoachService
}

func NewCoachHandler(svc *service.CoachService) *CoachHandler {
	return &CoachHandler{svc: svc}
}

func (h *CoachHandler) Clients(w http.ResponseWriter, r *http.Request) {
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

	clients, err := h.svc.Clients(r.Context(), orgID, userID)
	if err != nil {
		writeError(w, err)
		return
	}

	out := make([]dto.CoachClientDTO, len(clients))
	for i, c := range clients {
		out[i] = dto.CoachClientDTO{
			ClientUserID:   c.ClientUserID,
			FullName:       c.FullName,
			AvatarURL:      c.AvatarURL,
			AssignmentID:   c.AssignmentID,
			AssignmentName: c.AssignmentName,
			HasAssignment:  c.HasAssignment,
			LastSessionAt:  c.LastSessionAt,
			StreakDays:     c.StreakDays,
			RecentPRs:      c.RecentPRs,
			NeedsAttention: c.NeedsAttention,
		}
	}
	writeJSON(w, http.StatusOK, out)
}
