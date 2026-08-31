package postgres

import (
	"context"

	"github.com/google/uuid"

	"myvibesfit/api/internal/domain"
	"myvibesfit/api/internal/repository/db"
)

type DeviceTokenRepository struct{ q db.Querier }

func NewDeviceTokenRepository(q db.Querier) *DeviceTokenRepository {
	return &DeviceTokenRepository{q: q}
}

var _ domain.DeviceTokenRepository = (*DeviceTokenRepository)(nil)

func (r *DeviceTokenRepository) Upsert(ctx context.Context, userID uuid.UUID, token, platform string) (domain.DeviceToken, error) {
	row, err := r.q.UpsertDeviceToken(ctx, db.UpsertDeviceTokenParams{
		UserID: userID, Token: token, Platform: platform,
	})
	if err != nil {
		return domain.DeviceToken{}, err
	}
	return toDomainDeviceToken(row), nil
}

func (r *DeviceTokenRepository) ListByUser(ctx context.Context, userID uuid.UUID) ([]domain.DeviceToken, error) {
	rows, err := r.q.ListDeviceTokensByUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	out := make([]domain.DeviceToken, 0, len(rows))
	for _, row := range rows {
		out = append(out, toDomainDeviceToken(row))
	}
	return out, nil
}

func (r *DeviceTokenRepository) Delete(ctx context.Context, userID uuid.UUID, token string) (bool, error) {
	n, err := r.q.DeleteDeviceToken(ctx, db.DeleteDeviceTokenParams{Token: token, UserID: userID})
	if err != nil {
		return false, err
	}
	return n > 0, nil
}

func toDomainDeviceToken(t db.DeviceToken) domain.DeviceToken {
	return domain.DeviceToken{
		ID: t.ID, UserID: t.UserID, Token: t.Token, Platform: t.Platform,
		LastSeenAt: t.LastSeenAt, CreatedAt: t.CreatedAt,
	}
}
