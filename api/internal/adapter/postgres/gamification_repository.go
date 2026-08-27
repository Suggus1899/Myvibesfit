package postgres

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"myvibesfit/api/internal/domain"
	"myvibesfit/api/internal/repository/db"
)

type GamificationRepository struct{ q db.Querier }

func NewGamificationRepository(q db.Querier) *GamificationRepository {
	return &GamificationRepository{q: q}
}

var _ domain.GamificationRepository = (*GamificationRepository)(nil)

func (r *GamificationRepository) GetUserStats(ctx context.Context, userID uuid.UUID) (domain.UserStat, error) {
	row, err := r.q.GetUserStats(ctx, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.UserStat{}, domain.ErrNotFound
		}
		return domain.UserStat{}, err
	}
	return toDomainUserStat(row), nil
}

func (r *GamificationRepository) UpsertUserStats(ctx context.Context, s domain.UserStat) (domain.UserStat, error) {
	row, err := r.q.UpsertUserStats(ctx, db.UpsertUserStatsParams{
		UserID: s.UserID, TotalXp: s.TotalXP, Level: s.Level, TotalSessions: s.TotalSessions, TotalVolumeKg: s.TotalVolumeKg,
	})
	if err != nil {
		return domain.UserStat{}, err
	}
	return toDomainUserStat(row), nil
}

func (r *GamificationRepository) GetUserStreak(ctx context.Context, userID uuid.UUID, kind string) (domain.UserStreak, error) {
	row, err := r.q.GetUserStreak(ctx, db.GetUserStreakParams{UserID: userID, Kind: db.StreakKind(kind)})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.UserStreak{}, domain.ErrNotFound
		}
		return domain.UserStreak{}, err
	}
	return toDomainUserStreak(row), nil
}

func (r *GamificationRepository) UpsertUserStreak(ctx context.Context, s domain.UserStreak) (domain.UserStreak, error) {
	row, err := r.q.UpsertUserStreak(ctx, db.UpsertUserStreakParams{
		UserID: s.UserID, Kind: db.StreakKind(s.Kind), CurrentCount: s.CurrentCount, LongestCount: s.LongestCount, LastActiveOn: s.LastActiveOn,
	})
	if err != nil {
		return domain.UserStreak{}, err
	}
	return toDomainUserStreak(row), nil
}

func (r *GamificationRepository) ListUserStreaks(ctx context.Context, userID uuid.UUID) ([]domain.UserStreak, error) {
	rows, err := r.q.ListUserStreaks(ctx, userID)
	if err != nil {
		return nil, err
	}
	out := make([]domain.UserStreak, len(rows))
	for i, row := range rows {
		out[i] = toDomainUserStreak(row)
	}
	return out, nil
}

func (r *GamificationRepository) CreateXPEvent(ctx context.Context, e domain.XPEvent) error {
	_, err := r.q.CreateXPEvent(ctx, db.CreateXPEventParams{
		UserID: e.UserID, Source: db.XpSource(e.Source), Points: e.Points, ReferenceID: stringToPgText(e.ReferenceID),
	})
	return err
}

func (r *GamificationRepository) ListAchievements(ctx context.Context) ([]domain.Achievement, error) {
	rows, err := r.q.ListAchievements(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]domain.Achievement, len(rows))
	for i, row := range rows {
		out[i] = toDomainAchievement(row)
	}
	return out, nil
}

func (r *GamificationRepository) GetUserAchievement(ctx context.Context, userID, achievementID uuid.UUID) (domain.UserAchievement, error) {
	row, err := r.q.GetUserAchievement(ctx, db.GetUserAchievementParams{UserID: userID, AchievementID: achievementID})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.UserAchievement{}, domain.ErrNotFound
		}
		return domain.UserAchievement{}, err
	}
	return domain.UserAchievement{UserID: row.UserID, AchievementID: row.AchievementID, Progress: row.Progress, EarnedAt: row.EarnedAt}, nil
}

func (r *GamificationRepository) UpsertUserAchievement(ctx context.Context, a domain.UserAchievement) (domain.UserAchievement, error) {
	row, err := r.q.UpsertUserAchievement(ctx, db.UpsertUserAchievementParams{
		UserID: a.UserID, AchievementID: a.AchievementID, Progress: a.Progress, EarnedAt: a.EarnedAt,
	})
	if err != nil {
		return domain.UserAchievement{}, err
	}
	return domain.UserAchievement{UserID: row.UserID, AchievementID: row.AchievementID, Progress: row.Progress, EarnedAt: row.EarnedAt}, nil
}

func (r *GamificationRepository) ListUserAchievements(ctx context.Context, userID uuid.UUID) ([]domain.UserAchievementInfo, error) {
	rows, err := r.q.ListUserAchievements(ctx, userID)
	if err != nil {
		return nil, err
	}
	out := make([]domain.UserAchievementInfo, len(rows))
	for i, row := range rows {
		out[i] = domain.UserAchievementInfo{
			UserID: row.UserID, AchievementID: row.AchievementID, Progress: row.Progress, EarnedAt: row.EarnedAt,
			Code: row.Code, Name: row.Name, Description: row.Description, Icon: row.Icon, Tier: row.Tier, XPReward: row.XpReward,
		}
	}
	return out, nil
}

func toDomainUserStat(s db.UserStat) domain.UserStat {
	return domain.UserStat{
		UserID: s.UserID, TotalXP: s.TotalXp, Level: s.Level, TotalSessions: s.TotalSessions,
		TotalVolumeKg: s.TotalVolumeKg, UpdatedAt: s.UpdatedAt,
	}
}

func toDomainUserStreak(s db.UserStreak) domain.UserStreak {
	return domain.UserStreak{
		UserID: s.UserID, Kind: string(s.Kind), CurrentCount: s.CurrentCount, LongestCount: s.LongestCount,
		LastActiveOn: s.LastActiveOn, FreezesLeft: s.FreezesLeft,
	}
}

func toDomainAchievement(a db.Achievement) domain.Achievement {
	return domain.Achievement{
		ID: a.ID, Code: a.Code, Name: a.Name, Description: a.Description, Icon: a.Icon,
		Tier: a.Tier, XPReward: a.XpReward, Criteria: a.Criteria,
	}
}
