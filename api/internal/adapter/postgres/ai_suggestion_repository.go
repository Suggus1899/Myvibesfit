package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"myvibesfit/api/internal/domain"
	"myvibesfit/api/internal/repository/db"
)

type AISuggestionRepository struct{ q db.Querier }

func NewAISuggestionRepository(q db.Querier) *AISuggestionRepository {
	return &AISuggestionRepository{q: q}
}

var _ domain.AISuggestionRepository = (*AISuggestionRepository)(nil)

func (r *AISuggestionRepository) ListActiveAssignmentsForWorker(ctx context.Context) ([]domain.ActiveAssignmentForWorker, error) {
	rows, err := r.q.ListActiveAssignmentsForWorker(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]domain.ActiveAssignmentForWorker, len(rows))
	for i, row := range rows {
		out[i] = domain.ActiveAssignmentForWorker{
			AssignmentID: row.AssignmentID, OrgID: row.OrgID, ClientUserID: row.ClientUserID,
			CoachUserID: pgUUIDToPtr(row.CoachUserID), ClientName: row.ClientName,
		}
	}
	return out, nil
}

func (r *AISuggestionRepository) CountAssignedWorkoutsInRange(ctx context.Context, assignmentID uuid.UUID, from, to time.Time) (int64, error) {
	return r.q.CountAssignedWorkoutsInRange(ctx, db.CountAssignedWorkoutsInRangeParams{
		AssignmentID: assignmentID, ScheduledOn: &from, ScheduledOn_2: &to,
	})
}

func (r *AISuggestionRepository) CountCompletedSessionsSince(ctx context.Context, userID uuid.UUID, since time.Time) (int64, error) {
	return r.q.CountCompletedSessionsSince(ctx, db.CountCompletedSessionsSinceParams{UserID: userID, StartedAt: since})
}

func (r *AISuggestionRepository) CountCompletedHabitLogsSince(ctx context.Context, userID uuid.UUID, since time.Time) (int64, error) {
	return r.q.CountCompletedHabitLogsSince(ctx, db.CountCompletedHabitLogsSinceParams{UserID: userID, LogDate: since})
}

func (r *AISuggestionRepository) ListRecentWorkingSets(ctx context.Context, userID uuid.UUID, since time.Time, limit int32) ([]domain.RecentWorkingSet, error) {
	rows, err := r.q.ListRecentWorkingSets(ctx, db.ListRecentWorkingSetsParams{UserID: userID, PerformedAt: since, Limit: limit})
	if err != nil {
		return nil, err
	}
	out := make([]domain.RecentWorkingSet, len(rows))
	for i, row := range rows {
		out[i] = domain.RecentWorkingSet{
			ExerciseID: row.ExerciseID, ExerciseName: row.ExerciseName, WeightKg: row.WeightKg,
			Reps: pgInt2ToPtr(row.Reps), RPE: row.Rpe, PerformedAt: row.PerformedAt,
		}
	}
	return out, nil
}

func (r *AISuggestionRepository) Create(ctx context.Context, s domain.AISuggestion) (domain.AISuggestion, error) {
	row, err := r.q.CreateAISuggestion(ctx, db.CreateAISuggestionParams{
		OrgID: s.OrgID, ClientUserID: s.ClientUserID, CoachUserID: uuidPtrToPgUUID(s.CoachUserID),
		AssignmentID: uuidPtrToPgUUID(s.AssignmentID), Kind: db.SuggestionKind(s.Kind), Payload: s.Payload,
		Rationale: s.Rationale, Confidence: s.Confidence, Model: s.Model, InputSnapshot: s.InputSnapshot, ExpiresAt: s.ExpiresAt,
	})
	if err != nil {
		return domain.AISuggestion{}, err
	}
	return toDomainAISuggestion(row), nil
}

func (r *AISuggestionRepository) GetByID(ctx context.Context, id, orgID uuid.UUID) (domain.AISuggestion, error) {
	row, err := r.q.GetAISuggestionByID(ctx, db.GetAISuggestionByIDParams{ID: id, OrgID: orgID})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.AISuggestion{}, domain.ErrNotFound
		}
		return domain.AISuggestion{}, err
	}
	return toDomainAISuggestion(row), nil
}

func (r *AISuggestionRepository) ListPendingForCoach(ctx context.Context, coachID, orgID uuid.UUID) ([]domain.PendingSuggestion, error) {
	coach := coachID
	rows, err := r.q.ListPendingSuggestionsForCoach(ctx, db.ListPendingSuggestionsForCoachParams{
		CoachUserID: uuidPtrToPgUUID(&coach), OrgID: orgID,
	})
	if err != nil {
		return nil, err
	}
	out := make([]domain.PendingSuggestion, len(rows))
	for i, row := range rows {
		out[i] = domain.PendingSuggestion{
			AISuggestion: domain.AISuggestion{
				ID: row.ID, OrgID: row.OrgID, ClientUserID: row.ClientUserID, CoachUserID: pgUUIDToPtr(row.CoachUserID),
				AssignmentID: pgUUIDToPtr(row.AssignmentID), Kind: string(row.Kind), Payload: row.Payload,
				Rationale: row.Rationale, Confidence: row.Confidence, Model: row.Model, InputSnapshot: row.InputSnapshot,
				Status: domain.SuggestionStatus(row.Status), ReviewedBy: pgUUIDToPtr(row.ReviewedBy), ReviewedAt: row.ReviewedAt,
				AppliedAt: row.AppliedAt, ExpiresAt: row.ExpiresAt, CreatedAt: row.CreatedAt,
			},
			ClientName: row.ClientName,
		}
	}
	return out, nil
}

func (r *AISuggestionRepository) Review(ctx context.Context, id, orgID uuid.UUID, status domain.SuggestionStatus, reviewedBy uuid.UUID) (domain.AISuggestion, error) {
	row, err := r.q.ReviewAISuggestion(ctx, db.ReviewAISuggestionParams{
		ID: id, OrgID: orgID, Status: db.SuggestionStatus(status), ReviewedBy: uuidPtrToPgUUID(&reviewedBy),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.AISuggestion{}, domain.ErrNotFound
		}
		return domain.AISuggestion{}, err
	}
	return toDomainAISuggestion(row), nil
}

func toDomainAISuggestion(s db.AiSuggestion) domain.AISuggestion {
	return domain.AISuggestion{
		ID: s.ID, OrgID: s.OrgID, ClientUserID: s.ClientUserID, CoachUserID: pgUUIDToPtr(s.CoachUserID),
		AssignmentID: pgUUIDToPtr(s.AssignmentID), Kind: string(s.Kind), Payload: s.Payload, Rationale: s.Rationale,
		Confidence: s.Confidence, Model: s.Model, InputSnapshot: s.InputSnapshot, Status: domain.SuggestionStatus(s.Status),
		ReviewedBy: pgUUIDToPtr(s.ReviewedBy), ReviewedAt: s.ReviewedAt, AppliedAt: s.AppliedAt, ExpiresAt: s.ExpiresAt, CreatedAt: s.CreatedAt,
	}
}
