package domain

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
)

// fakeGamificationRepo implementa el port en memoria: alcanza para ejercitar
// la logica de GamificationService sin Postgres.
type fakeGamificationRepo struct {
	stats        map[uuid.UUID]UserStat
	streaks      map[string]UserStreak
	achievements []Achievement
	earned       map[string]UserAchievement
	xpEvents     []XPEvent
}

func newFakeRepo() *fakeGamificationRepo {
	return &fakeGamificationRepo{
		stats:   map[uuid.UUID]UserStat{},
		streaks: map[string]UserStreak{},
		earned:  map[string]UserAchievement{},
	}
}

func streakKey(u uuid.UUID, kind string) string { return u.String() + "|" + kind }
func earnedKey(u, a uuid.UUID) string           { return u.String() + "|" + a.String() }

func (f *fakeGamificationRepo) GetUserStats(_ context.Context, userID uuid.UUID) (UserStat, error) {
	s, ok := f.stats[userID]
	if !ok {
		return UserStat{}, ErrNotFound
	}
	return s, nil
}

// En memoria no hay locks: ForUpdate se comporta igual que la lectura suelta.
func (f *fakeGamificationRepo) GetUserStatsForUpdate(ctx context.Context, userID uuid.UUID) (UserStat, error) {
	return f.GetUserStats(ctx, userID)
}

func (f *fakeGamificationRepo) GetUserStreakForUpdate(ctx context.Context, userID uuid.UUID, kind string) (UserStreak, error) {
	return f.GetUserStreak(ctx, userID, kind)
}

func (f *fakeGamificationRepo) UpsertUserStats(_ context.Context, s UserStat) (UserStat, error) {
	f.stats[s.UserID] = s
	return s, nil
}

func (f *fakeGamificationRepo) GetUserStreak(_ context.Context, userID uuid.UUID, kind string) (UserStreak, error) {
	s, ok := f.streaks[streakKey(userID, kind)]
	if !ok {
		return UserStreak{}, ErrNotFound
	}
	return s, nil
}

func (f *fakeGamificationRepo) UpsertUserStreak(_ context.Context, s UserStreak) (UserStreak, error) {
	f.streaks[streakKey(s.UserID, s.Kind)] = s
	return s, nil
}

func (f *fakeGamificationRepo) ListUserStreaks(_ context.Context, userID uuid.UUID) ([]UserStreak, error) {
	var out []UserStreak
	for _, s := range f.streaks {
		if s.UserID == userID {
			out = append(out, s)
		}
	}
	return out, nil
}

func (f *fakeGamificationRepo) CreateXPEvent(_ context.Context, e XPEvent) error {
	f.xpEvents = append(f.xpEvents, e)
	return nil
}

func (f *fakeGamificationRepo) ListAchievements(_ context.Context) ([]Achievement, error) {
	return f.achievements, nil
}

func (f *fakeGamificationRepo) GetUserAchievement(_ context.Context, userID, achievementID uuid.UUID) (UserAchievement, error) {
	a, ok := f.earned[earnedKey(userID, achievementID)]
	if !ok {
		return UserAchievement{}, ErrNotFound
	}
	return a, nil
}

func (f *fakeGamificationRepo) UpsertUserAchievement(_ context.Context, a UserAchievement) (UserAchievement, error) {
	f.earned[earnedKey(a.UserID, a.AchievementID)] = a
	return a, nil
}

func (f *fakeGamificationRepo) ListUserAchievements(_ context.Context, _ uuid.UUID) ([]UserAchievementInfo, error) {
	return nil, nil
}

func TestApplyStatsDeltaAcumulaYCalculaNivel(t *testing.T) {
	ctx := context.Background()
	repo := newFakeRepo()
	svc := NewGamificationService()
	user := uuid.New()

	// Usuario nuevo: no hay fila de stats todavia (ErrNotFound tolerado).
	got, err := svc.ApplyStatsDelta(ctx, repo, user, StatsDelta{XP: 120, Sessions: 1, VolumeKg: 1000})
	if err != nil {
		t.Fatalf("primer delta: %v", err)
	}
	if got.TotalXP != 120 || got.TotalSessions != 1 {
		t.Fatalf("primer delta: xp=%d sessions=%d, esperaba 120/1", got.TotalXP, got.TotalSessions)
	}
	if got.Level != 1 {
		t.Fatalf("120 XP deberia ser nivel 1, fue %d", got.Level)
	}

	// Segundo delta acumula sobre el anterior y cruza a nivel 2 (500 XP).
	got, err = svc.ApplyStatsDelta(ctx, repo, user, StatsDelta{XP: 400, Sessions: 2, VolumeKg: 500})
	if err != nil {
		t.Fatalf("segundo delta: %v", err)
	}
	if got.TotalXP != 520 || got.TotalSessions != 3 || got.TotalVolumeKg != 1500 {
		t.Fatalf("acumulado incorrecto: %+v", got)
	}
	if got.Level != 2 {
		t.Fatalf("520 XP deberia ser nivel 2, fue %d", got.Level)
	}
}

func TestBumpStreak(t *testing.T) {
	ctx := context.Background()
	svc := NewGamificationService()
	day := func(d int) time.Time { return time.Date(2026, 3, 10+d, 12, 0, 0, 0, time.UTC) }

	t.Run("primera actividad arranca en 1", func(t *testing.T) {
		repo := newFakeRepo()
		got, err := svc.BumpStreak(ctx, repo, uuid.New(), StreakKindWorkout, day(0))
		if err != nil {
			t.Fatal(err)
		}
		if got.CurrentCount != 1 || got.LongestCount != 1 {
			t.Fatalf("esperaba 1/1, fue %d/%d", got.CurrentCount, got.LongestCount)
		}
	})

	t.Run("dia consecutivo suma", func(t *testing.T) {
		repo := newFakeRepo()
		user := uuid.New()
		if _, err := svc.BumpStreak(ctx, repo, user, StreakKindWorkout, day(0)); err != nil {
			t.Fatal(err)
		}
		got, err := svc.BumpStreak(ctx, repo, user, StreakKindWorkout, day(1))
		if err != nil {
			t.Fatal(err)
		}
		if got.CurrentCount != 2 {
			t.Fatalf("esperaba 2, fue %d", got.CurrentCount)
		}
	})

	t.Run("mismo dia no cuenta dos veces", func(t *testing.T) {
		repo := newFakeRepo()
		user := uuid.New()
		if _, err := svc.BumpStreak(ctx, repo, user, StreakKindWorkout, day(0)); err != nil {
			t.Fatal(err)
		}
		// Otra hora del mismo dia: es el caso real de entrenar dos veces.
		got, err := svc.BumpStreak(ctx, repo, user, StreakKindWorkout, day(0).Add(6*time.Hour))
		if err != nil {
			t.Fatal(err)
		}
		if got.CurrentCount != 1 {
			t.Fatalf("esperaba seguir en 1, fue %d", got.CurrentCount)
		}
	})

	t.Run("salto de dias reinicia pero conserva el record", func(t *testing.T) {
		repo := newFakeRepo()
		user := uuid.New()
		for i := range 3 {
			if _, err := svc.BumpStreak(ctx, repo, user, StreakKindWorkout, day(i)); err != nil {
				t.Fatal(err)
			}
		}
		got, err := svc.BumpStreak(ctx, repo, user, StreakKindWorkout, day(10))
		if err != nil {
			t.Fatal(err)
		}
		if got.CurrentCount != 1 {
			t.Fatalf("racha rota deberia reiniciar a 1, fue %d", got.CurrentCount)
		}
		if got.LongestCount != 3 {
			t.Fatalf("longest deberia quedar en 3, fue %d", got.LongestCount)
		}
	})
}

func TestCheckAchievementsNoRedesbloquea(t *testing.T) {
	ctx := context.Background()
	repo := newFakeRepo()
	svc := NewGamificationService()
	user := uuid.New()

	first := Achievement{ID: uuid.New(), Code: "first-workout", XPReward: 50}
	ten := Achievement{ID: uuid.New(), Code: "ten-workouts", XPReward: 100}
	repo.achievements = []Achievement{first, ten}

	unlocked, err := svc.CheckAchievements(ctx, repo, user, AchievementSnapshot{TotalSessions: 1})
	if err != nil {
		t.Fatal(err)
	}
	if len(unlocked) != 1 || unlocked[0].Code != "first-workout" {
		t.Fatalf("esperaba solo first-workout, fue %+v", unlocked)
	}

	// Segunda pasada con el mismo snapshot: ya esta ganado, no se repite.
	unlocked, err = svc.CheckAchievements(ctx, repo, user, AchievementSnapshot{TotalSessions: 1})
	if err != nil {
		t.Fatal(err)
	}
	if len(unlocked) != 0 {
		t.Fatalf("no deberia redesbloquear, fue %+v", unlocked)
	}

	// Al cruzar el umbral del segundo logro, solo desbloquea ese.
	unlocked, err = svc.CheckAchievements(ctx, repo, user, AchievementSnapshot{TotalSessions: 10})
	if err != nil {
		t.Fatal(err)
	}
	if len(unlocked) != 1 || unlocked[0].Code != "ten-workouts" {
		t.Fatalf("esperaba solo ten-workouts, fue %+v", unlocked)
	}
}
