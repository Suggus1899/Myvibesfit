package postgres

import (
	"context"

	"github.com/google/uuid"

	"myvibesfit/api/internal/domain"
	"myvibesfit/api/internal/repository/db"
)

type CoachRepository struct{ q db.Querier }

func NewCoachRepository(q db.Querier) *CoachRepository { return &CoachRepository{q: q} }

var _ domain.CoachRepository = (*CoachRepository)(nil)

func (r *CoachRepository) ListClients(ctx context.Context, orgID, coachUserID uuid.UUID) ([]domain.CoachClientRow, error) {
	rows, err := r.q.ListCoachClients(ctx, db.ListCoachClientsParams{OrgID: orgID, CoachUserID: coachUserID})
	if err != nil {
		return nil, err
	}
	out := make([]domain.CoachClientRow, len(rows))
	for i, row := range rows {
		out[i] = domain.CoachClientRow{
			ClientUserID: row.ClientUserID, FullName: row.FullName, AvatarURL: pgTextToString(row.AvatarUrl),
			AssignmentID: pgUUIDToPtr(row.AssignmentID), AssignmentName: pgTextToString(row.AssignmentName),
			AssignmentStatus: row.AssignmentStatus, LastSessionAt: row.LastSessionAt, StreakDays: row.StreakDays, RecentPRs: row.RecentPrs,
		}
	}
	return out, nil
}
