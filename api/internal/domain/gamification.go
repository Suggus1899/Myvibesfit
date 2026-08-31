package domain

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

const (
	StreakKindWorkout = "workout"
	StreakKindHabit   = "habit"
	StreakKindOverall = "overall"

	XpSourceSession        = "session"
	XpSourceHabit          = "habit"
	XpSourcePersonalRecord = "personal_record"
	XpSourceStreak         = "streak"
	XpSourceAchievement    = "achievement"
)

const (
	XPPerSession        = 50
	XPPerPersonalRecord = 100
	XPPerHabitCheck     = 10
	xpPerLevel          = 500
)

type UserStat struct {
	UserID        uuid.UUID
	TotalXP       int32
	Level         int16
	TotalSessions int32
	TotalVolumeKg float64
	UpdatedAt     time.Time
}

type UserStreak struct {
	UserID       uuid.UUID
	Kind         string
	CurrentCount int32
	LongestCount int32
	LastActiveOn *time.Time
	FreezesLeft  int16
}

type XPEvent struct {
	ID          int64
	UserID      uuid.UUID
	Source      string
	Points      int32
	ReferenceID string
	CreatedAt   time.Time
}

type Achievement struct {
	ID          uuid.UUID
	Code        string
	Name        string
	Description string
	Icon        string
	Tier        int16
	XPReward    int32
	Criteria    []byte
}

type UserAchievement struct {
	UserID        uuid.UUID
	AchievementID uuid.UUID
	Progress      float64
	EarnedAt      *time.Time
}

type UserAchievementInfo struct {
	UserID        uuid.UUID
	AchievementID uuid.UUID
	Progress      float64
	EarnedAt      *time.Time
	Code          string
	Name          string
	Description   string
	Icon          string
	Tier          int16
	XPReward      int32
}

// GamificationRepository espeja gamification.sql.go (10 metodos).
type GamificationRepository interface {
	GetUserStats(ctx context.Context, userID uuid.UUID) (UserStat, error)
	UpsertUserStats(ctx context.Context, s UserStat) (UserStat, error)
	GetUserStreak(ctx context.Context, userID uuid.UUID, kind string) (UserStreak, error)

	// Las variantes ForUpdate toman lock de fila: son para el
	// read-modify-write de ApplyStatsDelta/BumpStreak dentro de una
	// transaccion. Sin el lock, dos syncs concurrentes del mismo usuario
	// (dos dispositivos, un retry solapado) pierden uno de los dos deltas.
	GetUserStatsForUpdate(ctx context.Context, userID uuid.UUID) (UserStat, error)
	GetUserStreakForUpdate(ctx context.Context, userID uuid.UUID, kind string) (UserStreak, error)
	UpsertUserStreak(ctx context.Context, s UserStreak) (UserStreak, error)
	ListUserStreaks(ctx context.Context, userID uuid.UUID) ([]UserStreak, error)
	CreateXPEvent(ctx context.Context, e XPEvent) error
	ListAchievements(ctx context.Context) ([]Achievement, error)
	GetUserAchievement(ctx context.Context, userID, achievementID uuid.UUID) (UserAchievement, error)
	UpsertUserAchievement(ctx context.Context, a UserAchievement) (UserAchievement, error)
	ListUserAchievements(ctx context.Context, userID uuid.UUID) ([]UserAchievementInfo, error)
}

// GamificationService no guarda estado propio: recibe el GamificationRepository
// (pool o tx-scoped, via TxRepos) en cada llamada para poder participar en
// la misma transaccion que el sync de entrenamiento o el check-in de
// habitos que la disparan. Vive en domain porque ya no depende de
// infraestructura, solo de otros ports.
type GamificationService struct{}

func NewGamificationService() *GamificationService {
	return &GamificationService{}
}

type StatsDelta struct {
	XP       int32
	Sessions int32
	VolumeKg float64
}

// ApplyStatsDelta suma deltas sobre el acumulado actual (no hay UPDATE
// parcial para este patron de upsert, asi que se relee y reescribe
// completo).
func (s *GamificationService) ApplyStatsDelta(ctx context.Context, repo GamificationRepository, userID uuid.UUID, d StatsDelta) (UserStat, error) {
	current, err := repo.GetUserStatsForUpdate(ctx, userID)
	if err != nil && !errors.Is(err, ErrNotFound) {
		return UserStat{}, err
	}

	xp, sessions, volume := d.XP, d.Sessions, d.VolumeKg
	if err == nil {
		xp += current.TotalXP
		sessions += current.TotalSessions
		volume += current.TotalVolumeKg
	}

	return repo.UpsertUserStats(ctx, UserStat{UserID: userID, TotalXP: xp, Level: levelForXP(xp), TotalSessions: sessions, TotalVolumeKg: volume})
}

func levelForXP(xp int32) int16 {
	return int16(1 + xp/xpPerLevel)
}

func (s *GamificationService) RecordXPEvent(ctx context.Context, repo GamificationRepository, userID uuid.UUID, source string, points int32, referenceID string) error {
	return repo.CreateXPEvent(ctx, XPEvent{UserID: userID, Source: source, Points: points, ReferenceID: referenceID})
}

// BumpStreak avanza la racha si activeDate es el dia siguiente a la ultima
// actividad, la reinicia si hay un salto, y no hace nada si ya se conto hoy.
func (s *GamificationService) BumpStreak(ctx context.Context, repo GamificationRepository, userID uuid.UUID, kind string, activeDate time.Time) (UserStreak, error) {
	current, err := repo.GetUserStreakForUpdate(ctx, userID, kind)
	if err != nil && !errors.Is(err, ErrNotFound) {
		return UserStreak{}, err
	}

	day := activeDate.UTC().Truncate(24 * time.Hour)
	newCount := int32(1)
	longest := int32(1)

	if err == nil {
		longest = current.LongestCount
		if current.LastActiveOn != nil {
			last := current.LastActiveOn.UTC().Truncate(24 * time.Hour)
			switch day.Sub(last).Hours() / 24 {
			case 0:
				return current, nil // ya contado hoy
			case 1:
				newCount = current.CurrentCount + 1
			default:
				newCount = 1
			}
		}
	}
	if newCount > longest {
		longest = newCount
	}

	return repo.UpsertUserStreak(ctx, UserStreak{UserID: userID, Kind: kind, CurrentCount: newCount, LongestCount: longest, LastActiveOn: &day})
}

type AchievementSnapshot struct {
	TotalSessions int
	WorkoutStreak int
	HasNewPR      bool
}

// achievementCriteria mapea el code de cada logro (contenido, vive en la
// tabla achievement) a la condicion que lo desbloquea (logica, vive en
// codigo). Para 6 logros un interprete generico de criteria jsonb es mas
// complejo que el problema que resuelve.
var achievementCriteria = map[string]func(AchievementSnapshot) bool{
	"first-workout":  func(s AchievementSnapshot) bool { return s.TotalSessions >= 1 },
	"ten-workouts":   func(s AchievementSnapshot) bool { return s.TotalSessions >= 10 },
	"fifty-workouts": func(s AchievementSnapshot) bool { return s.TotalSessions >= 50 },
	"streak-7":       func(s AchievementSnapshot) bool { return s.WorkoutStreak >= 7 },
	"streak-30":      func(s AchievementSnapshot) bool { return s.WorkoutStreak >= 30 },
	"first-pr":       func(s AchievementSnapshot) bool { return s.HasNewPR },
}

// GrantAchievementXP suma el xp_reward de los logros recien desbloqueados
// al acumulado del usuario. Separado de CheckAchievements porque este ultimo
// necesita ejecutarse leyendo stats ya actualizadas por la sesion/habito
// que disparo la verificacion; el reward de logro es un delta aparte.
func (s *GamificationService) GrantAchievementXP(ctx context.Context, repo GamificationRepository, userID uuid.UUID, unlocked []Achievement) error {
	if len(unlocked) == 0 {
		return nil
	}
	var total int32
	for _, a := range unlocked {
		total += a.XPReward
	}
	if err := s.RecordXPEvent(ctx, repo, userID, XpSourceAchievement, total, ""); err != nil {
		return err
	}
	_, err := s.ApplyStatsDelta(ctx, repo, userID, StatsDelta{XP: total})
	return err
}

func (s *GamificationService) Stats(ctx context.Context, repo GamificationRepository, userID uuid.UUID) (UserStat, []UserStreak, error) {
	stats, err := repo.GetUserStats(ctx, userID)
	if err != nil && !errors.Is(err, ErrNotFound) {
		return UserStat{}, nil, err
	}
	streaks, err := repo.ListUserStreaks(ctx, userID)
	if err != nil {
		return UserStat{}, nil, err
	}
	return stats, streaks, nil
}

func (s *GamificationService) MyAchievements(ctx context.Context, repo GamificationRepository, userID uuid.UUID) ([]UserAchievementInfo, error) {
	return repo.ListUserAchievements(ctx, userID)
}

func (s *GamificationService) CheckAchievements(ctx context.Context, repo GamificationRepository, userID uuid.UUID, snap AchievementSnapshot) ([]Achievement, error) {
	all, err := repo.ListAchievements(ctx)
	if err != nil {
		return nil, err
	}

	var unlocked []Achievement
	for _, a := range all {
		check, ok := achievementCriteria[a.Code]
		if !ok || !check(snap) {
			continue
		}

		existing, err := repo.GetUserAchievement(ctx, userID, a.ID)
		if err == nil && existing.EarnedAt != nil {
			continue // ya desbloqueado
		}
		if err != nil && !errors.Is(err, ErrNotFound) {
			return nil, err
		}

		now := time.Now()
		if _, err := repo.UpsertUserAchievement(ctx, UserAchievement{UserID: userID, AchievementID: a.ID, Progress: 100, EarnedAt: &now}); err != nil {
			return nil, err
		}
		unlocked = append(unlocked, a)
	}
	return unlocked, nil
}
