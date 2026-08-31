package service

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"

	"myvibesfit/api/internal/domain"
)

type HabitService struct {
	repo domain.HabitRepository
	uow  domain.UnitOfWork
	gam  *domain.GamificationService
}

func NewHabitService(repo domain.HabitRepository, uow domain.UnitOfWork, gam *domain.GamificationService) *HabitService {
	return &HabitService{repo: repo, uow: uow, gam: gam}
}

func (s *HabitService) ListHabits(ctx context.Context, orgID *uuid.UUID) ([]domain.Habit, error) {
	return s.repo.ListHabits(ctx, orgID)
}

func (s *HabitService) MyHabits(ctx context.Context, userID uuid.UUID) ([]domain.MyHabit, error) {
	return s.repo.ListMyHabits(ctx, userID)
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
func (s *HabitService) Subscribe(ctx context.Context, in SubscribeHabitInput) (domain.ClientHabit, error) {
	existing, err := s.repo.GetActiveClientHabit(ctx, in.UserID, in.HabitID)
	if err == nil {
		return existing, nil
	}
	if !errors.Is(err, domain.ErrNotFound) {
		return domain.ClientHabit{}, err
	}

	days := make([]int16, len(in.DaysOfWeek))
	for i, d := range in.DaysOfWeek {
		days[i] = int16(d)
	}

	return s.repo.SubscribeHabit(ctx, domain.ClientHabit{
		UserID: in.UserID, HabitID: in.HabitID, AssignedBy: in.AssignedBy,
		TargetValue: in.TargetValue, Frequency: stringOrDefault(in.Frequency, domain.DefaultHabitFrequency), DaysOfWeek: days,
	})
}

func (s *HabitService) Unsubscribe(ctx context.Context, id, userID uuid.UUID) error {
	return s.repo.UnsubscribeHabit(ctx, id, userID)
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
	Log                  domain.HabitLog
	UnlockedAchievements []domain.Achievement
}

// LogHabit registra el check-in y, si con este quedan todos los habitos
// activos del dia cumplidos, otorga XP y avanza la racha de habitos. Todo
// dentro de uow.Execute.
func (s *HabitService) LogHabit(ctx context.Context, in LogHabitInput) (HabitLogResult, error) {
	owner, err := s.repo.GetClientHabitOwner(ctx, in.ClientHabitID)
	if err != nil {
		return HabitLogResult{}, err
	}
	if owner != in.UserID {
		return HabitLogResult{}, domain.ErrNotFound
	}

	var result HabitLogResult

	err = s.uow.Execute(ctx, func(repos domain.TxRepos) error {
		log, err := repos.Habits.UpsertLog(ctx, domain.HabitLog{
			ClientLocalID: in.ClientLocalID, ClientHabitID: in.ClientHabitID, UserID: in.UserID,
			LogDate: in.LogDate, Value: in.Value, IsCompleted: in.IsCompleted,
		})
		if err != nil {
			return err
		}
		result = HabitLogResult{Log: log}

		// Solo el primer registro del dia premia. Un reenvio (reintento de
		// red, doble tap, resync al reabrir la app) actualiza la fila pero no
		// vuelve a otorgar XP ni a mover la racha.
		if !in.IsCompleted || !log.Inserted {
			return nil
		}

		if err := s.gam.RecordXPEvent(ctx, repos.Gamification, in.UserID, domain.XpSourceHabit, domain.XPPerHabitCheck, ""); err != nil {
			return err
		}
		if _, err := s.gam.ApplyStatsDelta(ctx, repos.Gamification, in.UserID, domain.StatsDelta{XP: domain.XPPerHabitCheck}); err != nil {
			return err
		}

		active, err := repos.Habits.CountActiveForUser(ctx, in.UserID)
		if err != nil {
			return err
		}
		completedToday, err := repos.Habits.CountCompletedLogsForDate(ctx, in.UserID, in.LogDate)
		if err != nil {
			return err
		}

		if active > 0 && completedToday >= active {
			if _, err := s.gam.BumpStreak(ctx, repos.Gamification, in.UserID, domain.StreakKindHabit, in.LogDate); err != nil {
				return err
			}
			if _, err := s.gam.BumpStreak(ctx, repos.Gamification, in.UserID, domain.StreakKindOverall, in.LogDate); err != nil {
				return err
			}
		}

		stats, err := repos.Gamification.GetUserStats(ctx, in.UserID)
		if err != nil && !errors.Is(err, domain.ErrNotFound) {
			return err
		}
		habitStreak, err := repos.Gamification.GetUserStreak(ctx, in.UserID, domain.StreakKindWorkout)
		if err != nil && !errors.Is(err, domain.ErrNotFound) {
			return err
		}
		unlocked, err := s.gam.CheckAchievements(ctx, repos.Gamification, in.UserID, domain.AchievementSnapshot{
			TotalSessions: int(stats.TotalSessions), WorkoutStreak: int(habitStreak.CurrentCount),
		})
		if err != nil {
			return err
		}
		if err := s.gam.GrantAchievementXP(ctx, repos.Gamification, in.UserID, unlocked); err != nil {
			return err
		}
		result.UnlockedAchievements = unlocked
		return nil
	})
	if err != nil {
		return HabitLogResult{}, err
	}
	return result, nil
}

func (s *HabitService) LogsForDate(ctx context.Context, userID uuid.UUID, date time.Time) ([]domain.HabitLog, error) {
	return s.repo.ListLogsForDate(ctx, userID, date)
}
