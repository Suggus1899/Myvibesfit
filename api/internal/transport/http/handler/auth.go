package handler

import (
	"encoding/json"
	"net/http"

	"myvibesfit/api/internal/domain"
	"myvibesfit/api/internal/service"
	"myvibesfit/api/internal/transport/http/dto"
	"myvibesfit/api/internal/transport/http/middleware"
)

type AuthHandler struct {
	svc *service.AuthService
}

func NewAuthHandler(svc *service.AuthService) *AuthHandler {
	return &AuthHandler{svc: svc}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req dto.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, domain.ErrInvalidInput)
		return
	}
	if req.Email == "" || req.Password == "" || req.FullName == "" {
		writeError(w, domain.ErrInvalidInput)
		return
	}

	result, err := h.svc.Register(r.Context(), req.Email, req.Password, req.FullName)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, authResponse(result))
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req dto.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, domain.ErrInvalidInput)
		return
	}

	result, err := h.svc.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, authResponse(result))
}

func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	var req dto.RefreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, domain.ErrInvalidInput)
		return
	}

	result, err := h.svc.Refresh(r.Context(), req.RefreshToken)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, authResponse(result))
}

func (h *AuthHandler) JoinOrg(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserID(r.Context())
	if !ok {
		writeError(w, domain.ErrUnauthorized)
		return
	}

	var req dto.JoinOrgRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.JoinCode == "" {
		writeError(w, domain.ErrInvalidInput)
		return
	}

	result, err := h.svc.JoinOrganization(r.Context(), userID, req.JoinCode)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, authResponse(result))
}

func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserID(r.Context())
	if !ok {
		writeError(w, domain.ErrUnauthorized)
		return
	}

	result, err := h.svc.Me(r.Context(), userID)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, dto.UserDTO{
		ID:       result.User.ID,
		Email:    result.User.Email,
		FullName: result.User.FullName,
		OrgID:    result.OrgID,
		Role:     result.Role,
	})
}

func authResponse(r *service.AuthResult) dto.AuthResponse {
	return dto.AuthResponse{
		AccessToken:  r.AccessToken,
		RefreshToken: r.RefreshToken,
		User: dto.UserDTO{
			ID:       r.User.ID,
			Email:    r.User.Email,
			FullName: r.User.FullName,
			OrgID:    r.OrgID,
			Role:     r.Role,
		},
	}
}
