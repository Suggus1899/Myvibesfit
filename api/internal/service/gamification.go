package service

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"myvibesfit/api/internal/repository/db"
)

const (
	xpPerSession        = 50
	xpPerPersonalRecord = 100
	xpPerHabitCheck     = 10
	xpPerLevel          = 500
)

// GamificationService no guarda estado propio: recibe el db.Querier (pool o
// tx) en cada llamada para poder participar en la misma transaccion que el
// sync de entrenamiento o el check-in de habitos que la disparan.
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
// parcial en sqlc para este patron de upsert, asi que se relee y reescribe
// completo).
func (s *GamificationService) ApplyStatsDelta(ctx context.Context, q db.Querier, userID uuid.UUID, d StatsDelta) (db.UserStat, error) {
	current, err := q.GetUserStats(ctx, userID)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return db.UserStat{}, err
	}

	xp, sessions, volume := d.XP, d.Sessions, d.VolumeKg
	if err == nil {
		xp += current.TotalXp
		sessions += current.TotalSessions
		volume += current.TotalVolumeKg
	}

	return q.UpsertUserStats(ctx, db.UpsertUserStatsParams{
		UserID: userID, TotalXp: xp, Level: levelForXP(xp), TotalSessions: sessions, TotalVolumeKg: volume,
	})
}

func levelForXP(xp int32) int16 {
	return int16(1 + xp/xpPerLevel)
}

func (s *GamificationService) RecordXPEvent(ctx context.Context, q db.Querier, userID uuid.UUID, source db.XpSource, points int32, referenceID string) error {
	_, err := q.CreateXPEvent(ctx, db.CreateXPEventParams{
		UserID: userID, Source: source, Points: points, ReferenceID: textToPg(referenceID),
	})
	return err
}

// BumpStreak avanza la racha si activeDate es el dia siguiente a la ultima
// actividad, la reinicia si hay un salto, y no hace nada si ya se conto hoy.
func (s *GamificationService) BumpStreak(ctx context.Context, q db.Querier, userID uuid.UUID, kind db.StreakKind, activeDate time.Time) (db.UserStreak, error) {
	current, err := q.GetUserStreak(ctx, db.GetUserStreakParams{UserID: userID, Kind: kind})
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return db.UserStreak{}, err
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

	return q.UpsertUserStreak(ctx, db.UpsertUserStreakParams{
		UserID: userID, Kind: kind, CurrentCount: newCount, LongestCount: longest, LastActiveOn: &day,
	})
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
func (s *GamificationService) GrantAchievementXP(ctx context.Context, q db.Querier, userID uuid.UUID, unlocked []db.Achievement) error {
	if len(unlocked) == 0 {
		return nil
	}
	var total int32
	for _, a := range unlocked {
		total += a.XpReward
	}
	if err := s.RecordXPEvent(ctx, q, userID, db.XpSourceAchievement, total, ""); err != nil {
		return err
	}
	_, err := s.ApplyStatsDelta(ctx, q, userID, StatsDelta{XP: total})
	return err
}

func (s *GamificationService) Stats(ctx context.Context, q db.Querier, userID uuid.UUID) (db.UserStat, []db.UserStreak, error) {
	stats, err := q.GetUserStats(ctx, userID)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return db.UserStat{}, nil, err
	}
	streaks, err := q.ListUserStreaks(ctx, userID)
	if err != nil {
		return db.UserStat{}, nil, err
	}
	return stats, streaks, nil
}

func (s *GamificationService) MyAchievements(ctx context.Context, q db.Querier, userID uuid.UUID) ([]db.ListUserAchievementsRow, error) {
	return q.ListUserAchievements(ctx, userID)
}

func (s *GamificationService) CheckAchievements(ctx context.Context, q db.Querier, userID uuid.UUID, snap AchievementSnapshot) ([]db.Achievement, error) {
	all, err := q.ListAchievements(ctx)
	if err != nil {
		return nil, err
	}

	var unlocked []db.Achievement
	for _, a := range all {
		check, ok := achievementCriteria[a.Code]
		if !ok || !check(snap) {
			continue
		}

		existing, err := q.GetUserAchievement(ctx, db.GetUserAchievementParams{UserID: userID, AchievementID: a.ID})
		if err == nil && existing.EarnedAt != nil {
			continue // ya desbloqueado
		}
		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return nil, err
		}

		now := time.Now()
		if _, err := q.UpsertUserAchievement(ctx, db.UpsertUserAchievementParams{
			UserID: userID, AchievementID: a.ID, Progress: 100, EarnedAt: &now,
		}); err != nil {
			return nil, err
		}
		unlocked = append(unlocked, a)
	}
	return unlocked, nil
}
