package postgres

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"myvibesfit/api/internal/domain"
	"myvibesfit/api/internal/repository/db"
)

type ProfileRepository struct{ q db.Querier }

func NewProfileRepository(q db.Querier) *ProfileRepository { return &ProfileRepository{q: q} }

var _ domain.ProfileRepository = (*ProfileRepository)(nil)

func (r *ProfileRepository) Get(ctx context.Context, userID uuid.UUID) (domain.ClientProfile, error) {
	row, err := r.q.GetClientProfile(ctx, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ClientProfile{}, domain.ErrNotFound
		}
		return domain.ClientProfile{}, err
	}
	return toDomainProfile(row), nil
}

func (r *ProfileRepository) Upsert(ctx context.Context, p domain.ClientProfile) (domain.ClientProfile, error) {
	row, err := r.q.UpsertClientProfile(ctx, db.UpsertClientProfileParams{
		UserID: p.UserID, BirthDate: p.BirthDate, Sex: db.BiologicalSex(p.Sex), HeightCm: p.HeightCm,
		Experience: db.ExperienceLevel(p.Experience), PrimaryGoal: db.TrainingGoal(p.PrimaryGoal),
		DaysPerWeek: p.DaysPerWeek, SessionMinutes: p.SessionMinutes,
		AvailableEquipment: p.AvailableEquipment, Limitations: stringToPgText(p.Limitations),
		UnitSystem: p.UnitSystem,
	})
	if err != nil {
		return domain.ClientProfile{}, err
	}
	return toDomainProfile(row), nil
}

func toDomainProfile(p db.ClientProfile) domain.ClientProfile {
	return domain.ClientProfile{
		UserID: p.UserID, BirthDate: p.BirthDate, Sex: string(p.Sex), HeightCm: p.HeightCm,
		Experience: string(p.Experience), PrimaryGoal: string(p.PrimaryGoal),
		DaysPerWeek: p.DaysPerWeek, SessionMinutes: p.SessionMinutes,
		AvailableEquipment: p.AvailableEquipment, Limitations: pgTextToString(p.Limitations),
		UnitSystem: p.UnitSystem, OnboardedAt: p.OnboardedAt,
	}
}
