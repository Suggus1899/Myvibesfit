package service

import (
	"context"
	"time"

	"github.com/google/uuid"

	"myvibesfit/api/internal/domain"
)

type CoachService struct {
	repo     domain.CoachRepository
	progress domain.ProgressRepository
}

func NewCoachService(repo domain.CoachRepository, progress domain.ProgressRepository) *CoachService {
	return &CoachService{repo: repo, progress: progress}
}

// ClientProgress devuelve los records de un cliente para la pantalla de
// detalle. Autoriza primero: un coach solo ve a los clientes vinculados a
// el, no a todos los de la organizacion.
func (s *CoachService) ClientProgress(ctx context.Context, orgID, coachUserID, clientUserID uuid.UUID) ([]domain.PersonalRecord, error) {
	ok, err := s.repo.IsClientOf(ctx, orgID, coachUserID, clientUserID)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, domain.ErrNotFound
	}
	return s.progress.ListPersonalRecords(ctx, clientUserID, nil)
}

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

		// La query devuelve 'epoch' cuando el cliente nunca entreno.
		var lastSession *time.Time
		if !r.LastSessionAt.IsZero() && r.LastSessionAt.Year() > 1970 {
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
			NeedsAttention: domain.NeedsAttention(hasAssignment, lastSession, now),
		}
	}
	return out, nil
}
