package service

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"myvibesfit/api/internal/repository/db"
)

type HabitService struct {
	q    db.Querier
	pool *pgxpool.Pool
	gam  *GamificationService
}

func NewHabitService(q db.Querier, pool *pgxpool.Pool, gam *GamificationService) *HabitService {
	return &HabitService{q: q, pool: pool, gam: gam}
}

func (s *HabitService) ListHabits(ctx context.Context, orgID *uuid.UUID) ([]db.Habit, error) {
	return s.q.ListHabits(ctx, uuidToPg(orgID))
}

func (s *HabitService) MyHabits(ctx context.Context, userID uuid.UUID) ([]db.ListMyHabitsRow, error) {
	return s.q.ListMyHabits(ctx, userID)
}

type SubscribeHabitInput struct {
	UserID      uuid.UUID
	HabitID     uuid.UUID
	AssignedBy  *uuid.UUID
	TargetValue *float64
	Frequency   string
	DaysOfWeek  []int
}

// Subscribe es idempotente: si ya existe una suscripcion activa a ese
// habito, la devuelve en vez de crear una segunda.
func (s *HabitService) Subscribe(ctx context.Context, in SubscribeHabitInput) (db.ClientHabit, error) {
	existing, err := s.q.GetActiveClientHabit(ctx, db.GetActiveClientHabitParams{UserID: in.UserID, HabitID: in.HabitID})
	if err == nil {
		return existing, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return db.ClientHabit{}, err
	}

	freq := db.HabitFrequency(in.Frequency)
	if freq == "" {
		freq = db.HabitFrequencyDaily
	}
	days := make([]int16, len(in.DaysOfWeek))
	for i, d := range in.DaysOfWeek {
		days[i] = int16(d)
	}

	return s.q.SubscribeHabit(ctx, db.SubscribeHabitParams{
		UserID: in.UserID, HabitID: in.HabitID, AssignedBy: uuidToPg(in.AssignedBy),
		TargetValue: in.TargetValue, Frequency: freq, DaysOfWeek: days,
	})
}

func (s *HabitService) Unsubscribe(ctx context.Context, id, userID uuid.UUID) error {
	return s.q.UnsubscribeHabit(ctx, db.UnsubscribeHabitParams{ID: id, UserID: userID})
}

type LogHabitInput struct {
	ClientLocalID uuid.UUID
	ClientHabitID uuid.UUID
	UserID        uuid.UUID
	LogDate       time.Time
	Value         float64
	IsCompleted   bool
}

type HabitLogResult struct {
	Log                  db.HabitLog
	UnlockedAchievements []db.Achievement
}

// LogHabit registra el check-in y, si con este quedan todos los habitos
// activos del dia cumplidos, otorga XP y avanza la racha de habitos. Todo
// en una transaccion.
func (s *HabitService) LogHabit(ctx context.Context, in LogHabitInput) (HabitLogResult, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return HabitLogResult{}, err
	}
	defer tx.Rollback(ctx) //nolint:errcheck
	qtx := db.New(tx)

	log, err := qtx.UpsertHabitLog(ctx, db.UpsertHabitLogParams{
		ClientLocalID: in.ClientLocalID, ClientHabitID: in.ClientHabitID, UserID: in.UserID,
		LogDate: in.LogDate, Value: in.Value, IsCompleted: in.IsCompleted,
	})
	if err != nil {
		return HabitLogResult{}, err
	}

	result := HabitLogResult{Log: log}

	if in.IsCompleted {
		if err := s.gam.RecordXPEvent(ctx, qtx, in.UserID, db.XpSourceHabit, xpPerHabitCheck, ""); err != nil {
			return HabitLogResult{}, err
		}
		if _, err := s.gam.ApplyStatsDelta(ctx, qtx, in.UserID, StatsDelta{XP: xpPerHabitCheck}); err != nil {
			return HabitLogResult{}, err
		}

		active, err := qtx.CountActiveHabitsForUser(ctx, in.UserID)
		if err != nil {
			return HabitLogResult{}, err
		}
		completedToday, err := qtx.CountCompletedHabitLogsForDate(ctx, db.CountCompletedHabitLogsForDateParams{UserID: in.UserID, LogDate: in.LogDate})
		if err != nil {
			return HabitLogResult{}, err
		}

		if active > 0 && completedToday >= active {
			if _, err := s.gam.BumpStreak(ctx, qtx, in.UserID, db.StreakKindHabit, in.LogDate); err != nil {
				return HabitLogResult{}, err
			}
			if _, err := s.gam.BumpStreak(ctx, qtx, in.UserID, db.StreakKindOverall, in.LogDate); err != nil {
				return HabitLogResult{}, err
			}
		}

		stats, err := qtx.GetUserStats(ctx, in.UserID)
		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return HabitLogResult{}, err
		}
		habitStreak, err := qtx.GetUserStreak(ctx, db.GetUserStreakParams{UserID: in.UserID, Kind: db.StreakKindWorkout})
		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return HabitLogResult{}, err
		}
		unlocked, err := s.gam.CheckAchievements(ctx, qtx, in.UserID, AchievementSnapshot{
			TotalSessions: int(stats.TotalSessions), WorkoutStreak: int(habitStreak.CurrentCount),
		})
		if err != nil {
			return HabitLogResult{}, err
		}
		if err := s.gam.GrantAchievementXP(ctx, qtx, in.UserID, unlocked); err != nil {
			return HabitLogResult{}, err
		}
		result.UnlockedAchievements = unlocked
	}

	if err := tx.Commit(ctx); err != nil {
		return HabitLogResult{}, err
	}
	return result, nil
}

func (s *HabitService) LogsForDate(ctx context.Context, userID uuid.UUID, date time.Time) ([]db.HabitLog, error) {
	return s.q.ListHabitLogsForDate(ctx, db.ListHabitLogsForDateParams{UserID: userID, LogDate: date})
}
