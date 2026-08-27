package service

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"

	"myvibesfit/api/internal/domain"
)

const (
	summaryPeriodDays  = 14
	suggestionTTLDays  = 7
	recentSetsLimit    = 300
	maxRecentPRsListed = 5
)

// SuggestionWorkerService arma el resumen estructurado de una asignacion y
// le pide a Claude una sugerencia (fase 9). Pensado para correr desde
// cmd/worker por cron, una vez por noche.
type SuggestionWorkerService struct {
	suggestions  domain.AISuggestionRepository
	gamification domain.GamificationRepository
	habits       domain.HabitRepository
	progress     domain.ProgressRepository
	proposer     domain.SuggestionProposer
	model        string
}

func NewSuggestionWorkerService(
	suggestions domain.AISuggestionRepository,
	gamification domain.GamificationRepository,
	habits domain.HabitRepository,
	progress domain.ProgressRepository,
	proposer domain.SuggestionProposer,
	model string,
) *SuggestionWorkerService {
	if model == "" {
		model = domain.DefaultSuggestionModel
	}
	return &SuggestionWorkerService{
		suggestions: suggestions, gamification: gamification, habits: habits,
		progress: progress, proposer: proposer, model: model,
	}
}

func (s *SuggestionWorkerService) ListActiveAssignments(ctx context.Context) ([]domain.ActiveAssignmentForWorker, error) {
	return s.suggestions.ListActiveAssignmentsForWorker(ctx)
}

// ProcessAssignment arma el resumen de la asignacion, le pide una sugerencia
// a Claude, y la guarda como pending. Nunca la aplica.
func (s *SuggestionWorkerService) ProcessAssignment(ctx context.Context, a domain.ActiveAssignmentForWorker) error {
	summary, err := s.buildSummary(ctx, a.AssignmentID, a.ClientUserID, a.ClientName)
	if err != nil {
		return err
	}

	sug, err := s.proposer.Propose(ctx, summary)
	if err != nil {
		return err
	}

	payload, err := json.Marshal(sug.Payload)
	if err != nil {
		return err
	}
	snapshot, err := json.Marshal(summary)
	if err != nil {
		return err
	}

	assignmentID := a.AssignmentID
	_, err = s.suggestions.Create(ctx, domain.AISuggestion{
		OrgID: a.OrgID, ClientUserID: a.ClientUserID, CoachUserID: a.CoachUserID, AssignmentID: &assignmentID,
		Kind: sug.Kind, Payload: payload, Rationale: sug.Rationale, Confidence: &sug.Confidence,
		Model: s.model, InputSnapshot: snapshot, ExpiresAt: time.Now().Add(suggestionTTLDays * 24 * time.Hour),
	})
	return err
}

func (s *SuggestionWorkerService) buildSummary(ctx context.Context, assignmentID, userID uuid.UUID, clientName string) (domain.AssignmentSummary, error) {
	since := time.Now().AddDate(0, 0, -summaryPeriodDays)
	sinceDate := since.Truncate(24 * time.Hour)
	today := time.Now().Truncate(24 * time.Hour)

	assignedWorkouts, err := s.suggestions.CountAssignedWorkoutsInRange(ctx, assignmentID, sinceDate, today)
	if err != nil {
		return domain.AssignmentSummary{}, err
	}

	completedSessions, err := s.suggestions.CountCompletedSessionsSince(ctx, userID, since)
	if err != nil {
		return domain.AssignmentSummary{}, err
	}

	streak, err := s.gamification.GetUserStreak(ctx, userID, domain.StreakKindWorkout)
	if err != nil && !errors.Is(err, domain.ErrNotFound) {
		streak = domain.UserStreak{} // sin fila todavia = sin racha, no es un error
	}

	activeHabits, err := s.habits.CountActiveForUser(ctx, userID)
	if err != nil {
		return domain.AssignmentSummary{}, err
	}
	completedHabitLogs, err := s.suggestions.CountCompletedHabitLogsSince(ctx, userID, sinceDate)
	if err != nil {
		return domain.AssignmentSummary{}, err
	}
	habitAdherence := 0.0
	if activeHabits > 0 {
		habitAdherence = float64(completedHabitLogs) / float64(activeHabits*summaryPeriodDays)
	}

	records, err := s.progress.ListPersonalRecords(ctx, userID, nil)
	if err != nil {
		return domain.AssignmentSummary{}, err
	}

	sets, err := s.suggestions.ListRecentWorkingSets(ctx, userID, since, recentSetsLimit)
	if err != nil {
		return domain.AssignmentSummary{}, err
	}

	exerciseNames := map[uuid.UUID]string{}
	trends := aggregateTrends(sets, exerciseNames)

	prs := make([]domain.PRSummary, 0, maxRecentPRsListed)
	for _, r := range records {
		if r.AchievedAt.Before(since) || len(prs) >= maxRecentPRsListed {
			continue
		}
		name := exerciseNames[r.ExerciseID]
		if name == "" {
			name = r.ExerciseID.String()
		}
		prs = append(prs, domain.PRSummary{
			ExerciseName: name, Type: string(r.Type), Value: r.Value,
			AchievedAt: r.AchievedAt.Format(time.RFC3339),
		})
	}

	return domain.AssignmentSummary{
		ClientName:        clientName,
		PeriodDays:        summaryPeriodDays,
		AssignedWorkouts:  int(assignedWorkouts),
		CompletedSessions: int(completedSessions),
		CurrentStreakDays: int(streak.CurrentCount),
		HabitAdherencePct: habitAdherence,
		RecentPRs:         prs,
		ExerciseTrends:    trends,
	}, nil
}

func aggregateTrends(sets []domain.RecentWorkingSet, names map[uuid.UUID]string) []domain.ExerciseTrend {
	type acc struct {
		name        string
		count       int
		first, last *float64
		rpeSum      float64
		rpeCount    int
	}
	byExercise := map[uuid.UUID]*acc{}
	order := []uuid.UUID{}

	for _, s := range sets {
		names[s.ExerciseID] = s.ExerciseName
		a, ok := byExercise[s.ExerciseID]
		if !ok {
			a = &acc{name: s.ExerciseName}
			byExercise[s.ExerciseID] = a
			order = append(order, s.ExerciseID)
		}
		a.count++
		if s.WeightKg != nil {
			if a.first == nil {
				a.first = s.WeightKg
			}
			a.last = s.WeightKg
		}
		if s.RPE != nil {
			a.rpeSum += *s.RPE
			a.rpeCount++
		}
	}

	trends := make([]domain.ExerciseTrend, 0, len(order))
	for _, id := range order {
		a := byExercise[id]
		t := domain.ExerciseTrend{ExerciseName: a.name, SetCount: a.count, FirstWeightKg: a.first, LastWeightKg: a.last}
		if a.rpeCount > 0 {
			avg := a.rpeSum / float64(a.rpeCount)
			t.AvgRPE = &avg
		}
		trends = append(trends, t)
	}
	return trends
}
