package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	"myvibesfit/api/internal/domain"
)

type ProfileService struct {
	repo domain.ProfileRepository
}

func NewProfileService(repo domain.ProfileRepository) *ProfileService {
	return &ProfileService{repo: repo}
}

// Get devuelve un perfil con los defaults del esquema (no un error) si el
// usuario nunca completo el onboarding: la app necesita distinguir "sin
// perfil" de "fallo la red", y el formulario tiene que abrir con valores
// usables, no con strings vacios. IsOnboarded queda en false.
func (s *ProfileService) Get(ctx context.Context, userID uuid.UUID) (domain.ClientProfile, error) {
	p, err := s.repo.Get(ctx, userID)
	if errors.Is(err, domain.ErrNotFound) {
		return domain.ClientProfile{
			UserID:             userID,
			Sex:                domain.DefaultBiologicalSex,
			Experience:         domain.DefaultExperienceLevel,
			PrimaryGoal:        domain.DefaultTrainingGoal,
			DaysPerWeek:        3,
			SessionMinutes:     60,
			AvailableEquipment: []string{},
			UnitSystem:         domain.UnitSystemMetric,
		}, nil
	}
	return p, err
}

type SaveProfileInput struct {
	UserID             uuid.UUID
	BirthDate          *time.Time
	Sex                string
	HeightCm           *float64
	Experience         string
	PrimaryGoal        string
	DaysPerWeek        int
	SessionMinutes     int
	AvailableEquipment []string
	Limitations        string
	UnitSystem         string
}

func (s *ProfileService) Save(ctx context.Context, in SaveProfileInput) (domain.ClientProfile, error) {
	unit := stringOrDefault(in.UnitSystem, domain.UnitSystemMetric)
	if unit != domain.UnitSystemMetric && unit != domain.UnitSystemImperial {
		return domain.ClientProfile{}, fmt.Errorf("%w: unit_system debe ser metric o imperial", domain.ErrInvalidInput)
	}
	if in.DaysPerWeek < 0 || in.DaysPerWeek > 7 {
		return domain.ClientProfile{}, fmt.Errorf("%w: days_per_week debe estar entre 1 y 7", domain.ErrInvalidInput)
	}

	return s.repo.Upsert(ctx, domain.ClientProfile{
		UserID:             in.UserID,
		BirthDate:          in.BirthDate,
		Sex:                stringOrDefault(in.Sex, domain.DefaultBiologicalSex),
		HeightCm:           in.HeightCm,
		Experience:         stringOrDefault(in.Experience, domain.DefaultExperienceLevel),
		PrimaryGoal:        stringOrDefault(in.PrimaryGoal, domain.DefaultTrainingGoal),
		DaysPerWeek:        int16OrDefault(in.DaysPerWeek, 3),
		SessionMinutes:     int16OrDefault(in.SessionMinutes, 60),
		AvailableEquipment: nonNilSlice(in.AvailableEquipment),
		Limitations:        in.Limitations,
		UnitSystem:         unit,
	})
}
