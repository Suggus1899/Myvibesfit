package service

import (
	"context"

	"github.com/google/uuid"

	"myvibesfit/api/internal/domain"
)

type AISuggestionService struct {
	repo  domain.AISuggestionRepository
	audit *AuditLogger
}

func NewAISuggestionService(repo domain.AISuggestionRepository, audit *AuditLogger) *AISuggestionService {
	return &AISuggestionService{repo: repo, audit: audit}
}

func (s *AISuggestionService) ListPending(ctx context.Context, coachID, orgID uuid.UUID) ([]domain.PendingSuggestion, error) {
	return s.repo.ListPendingForCoach(ctx, coachID, orgID)
}

// Review aprueba o rechaza una sugerencia pendiente. Nunca la aplica al
// plan — eso queda para un caso de uso posterior; aqui solo se registra la
// decision humana, que es el requisito de fase 9 ("nunca se aplica sin que
// el coach pulse Aprobar").
func (s *AISuggestionService) Review(ctx context.Context, id, orgID, reviewerID uuid.UUID, approve bool) (domain.AISuggestion, error) {
	status := domain.SuggestionStatusRejected
	if approve {
		status = domain.SuggestionStatusApproved
	}
	sug, err := s.repo.Review(ctx, id, orgID, status, reviewerID)
	if err != nil {
		return domain.AISuggestion{}, err
	}
	s.audit.Log(ctx, orgID, reviewerID, "ai_suggestion.review", "ai_suggestion", id.String(), map[string]any{"status": status})
	return sug, nil
}
