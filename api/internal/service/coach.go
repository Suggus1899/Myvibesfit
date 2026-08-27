package service

import (
	"context"
	"time"

	"github.com/google/uuid"

	"myvibesfit/api/internal/domain"
)

type CoachService struct {
	repo domain.CoachRepository
}

func NewCoachService(repo domain.CoachRepository) *CoachService {
	return &CoachService{repo: repo}
}

// attentionThreshold: sin entrenar en mas de 3 dias con asignacion activa
// es la definicion de "necesita atencion" para el overview del coach.
const attentionThreshold = 3 * 24 * time.Hour

type CoachClientOverview struct {
	ClientUserID   uuid.UUID
	FullName       string
	AvatarURL      string
	AssignmentID   *uuid.UUID
	AssignmentName string
	HasAssignment  bool
	LastSessionAt  *time.Time
	StreakDays     int
	RecentPRs      int
	NeedsAttention bool
}

func (s *CoachService) Clients(ctx context.Context, orgID, coachUserID uuid.UUID) ([]CoachClientOverview, error) {
	rows, err := s.repo.ListClients(ctx, orgID, coachUserID)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	out := make([]CoachClientOverview, len(rows))
	for i, r := range rows {
		hasAssignment := r.AssignmentStatus != ""

		var lastSession *time.Time
		if r.LastSessionAt.Unix() > 0 {
			t := r.LastSessionAt
			lastSession = &t
		}

		out[i] = CoachClientOverview{
			ClientUserID:   r.ClientUserID,
			FullName:       r.FullName,
			AvatarURL:      r.AvatarURL,
			AssignmentID:   r.AssignmentID,
			AssignmentName: r.AssignmentName,
			HasAssignment:  hasAssignment,
			LastSessionAt:  lastSession,
			StreakDays:     int(r.StreakDays),
			RecentPRs:      int(r.RecentPRs),
			NeedsAttention: hasAssignment && (lastSession == nil || now.Sub(*lastSession) > attentionThreshold),
		}
	}
	return out, nil
}
