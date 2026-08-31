package handler

import (
	"encoding/json"
	"net/http"

	"myvibesfit/api/internal/domain"
	"myvibesfit/api/internal/service"
	"myvibesfit/api/internal/transport/http/dto"
	"myvibesfit/api/internal/transport/http/middleware"
)

type DeviceTokenHandler struct {
	svc *service.NotificationService
}

func NewDeviceTokenHandler(svc *service.NotificationService) *DeviceTokenHandler {
	return &DeviceTokenHandler{svc: svc}
}

func (h *DeviceTokenHandler) Register(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserID(r.Context())
	if !ok {
		writeError(w, domain.ErrUnauthorized)
		return
	}
	var req dto.RegisterDeviceTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, domain.ErrInvalidInput)
		return
	}

	t, err := h.svc.Register(r.Context(), userID, req.Token, req.Platform)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, dto.DeviceTokenDTO{
		Token: t.Token, Platform: t.Platform, LastSeenAt: t.LastSeenAt,
	})
}

func (h *DeviceTokenHandler) Unregister(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserID(r.Context())
	if !ok {
		writeError(w, domain.ErrUnauthorized)
		return
	}
	var req dto.UnregisterDeviceTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, domain.ErrInvalidInput)
		return
	}

	if err := h.svc.Unregister(r.Context(), userID, req.Token); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
