package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"myvibesfit/api/internal/domain"
	"myvibesfit/api/internal/service"
	"myvibesfit/api/internal/transport/http/dto"
	"myvibesfit/api/internal/transport/http/middleware"
)

type ProfileHandler struct {
	svc *service.ProfileService
}

func NewProfileHandler(svc *service.ProfileService) *ProfileHandler {
	return &ProfileHandler{svc: svc}
}

func (h *ProfileHandler) Get(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserID(r.Context())
	if !ok {
		writeError(w, domain.ErrUnauthorized)
		return
	}
	p, err := h.svc.Get(r.Context(), userID)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, profileDTO(p))
}

func (h *ProfileHandler) Save(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserID(r.Context())
	if !ok {
		writeError(w, domain.ErrUnauthorized)
		return
	}
	var req dto.SaveProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, domain.ErrInvalidInput)
		return
	}

	in := service.SaveProfileInput{
		UserID: userID, Sex: req.Sex, HeightCm: req.HeightCm, Experience: req.Experience,
		PrimaryGoal: req.PrimaryGoal, DaysPerWeek: req.DaysPerWeek, SessionMinutes: req.SessionMinutes,
		AvailableEquipment: req.AvailableEquipment, Limitations: req.Limitations, UnitSystem: req.UnitSystem,
	}
	if req.BirthDate != "" {
		birth, err := time.Parse(dateLayout, req.BirthDate)
		if err != nil {
			writeError(w, domain.ErrInvalidInput)
			return
		}
		in.BirthDate = &birth
	}

	p, err := h.svc.Save(r.Context(), in)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, profileDTO(p))
}

func profileDTO(p domain.ClientProfile) dto.ClientProfileDTO {
	return dto.ClientProfileDTO{
		BirthDate: datePtrToString(p.BirthDate), Sex: p.Sex, HeightCm: p.HeightCm,
		Experience: p.Experience, PrimaryGoal: p.PrimaryGoal, DaysPerWeek: int(p.DaysPerWeek),
		SessionMinutes: int(p.SessionMinutes), AvailableEquipment: p.AvailableEquipment,
		Limitations: p.Limitations, UnitSystem: p.UnitSystem, IsOnboarded: p.IsOnboarded(),
	}
}
