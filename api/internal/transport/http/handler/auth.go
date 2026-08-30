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

func (h *AuthHandler) CreateOrg(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserID(r.Context())
	if !ok {
		writeError(w, domain.ErrUnauthorized)
		return
	}

	var req dto.CreateOrgRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, domain.ErrInvalidInput)
		return
	}

	result, err := h.svc.CreateOrganization(r.Context(), userID, req.Name)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, authResponse(result))
}

func (h *AuthHandler) GetOrg(w http.ResponseWriter, r *http.Request) {
	orgID, ok := middleware.OrgID(r.Context())
	if !ok {
		writeError(w, domain.ErrForbidden)
		return
	}
	org, err := h.svc.GetOrganization(r.Context(), orgID)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, dto.OrgDTO{
		ID: org.ID, Name: org.Name, Slug: org.Slug, JoinCode: org.JoinCode, BrandColor: org.BrandColor,
	})
}

func (h *AuthHandler) ListMembers(w http.ResponseWriter, r *http.Request) {
	orgID, ok := middleware.OrgID(r.Context())
	if !ok {
		writeError(w, domain.ErrForbidden)
		return
	}
	members, err := h.svc.ListOrgMembers(r.Context(), orgID)
	if err != nil {
		writeError(w, err)
		return
	}
	out := make([]dto.OrgMemberDTO, len(members))
	for i, m := range members {
		out[i] = dto.OrgMemberDTO{
			MembershipID: m.MembershipID, UserID: m.UserID, FullName: m.FullName, Email: m.Email,
			AvatarURL: m.AvatarURL, Role: m.Role, Status: m.Status, JoinedAt: m.JoinedAt,
		}
	}
	writeJSON(w, http.StatusOK, out)
}

func (h *AuthHandler) UpdateMemberRole(w http.ResponseWriter, r *http.Request) {
	orgID, ok := middleware.OrgID(r.Context())
	if !ok {
		writeError(w, domain.ErrForbidden)
		return
	}
	membershipID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, domain.ErrInvalidInput)
		return
	}
	var req dto.UpdateMemberRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, domain.ErrInvalidInput)
		return
	}

	m, err := h.svc.UpdateMemberRole(r.Context(), orgID, membershipID, req.Role)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, dto.OrgMemberDTO{
		MembershipID: m.ID, UserID: m.UserID, Role: m.Role, Status: m.Status, JoinedAt: m.JoinedAt,
	})
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
