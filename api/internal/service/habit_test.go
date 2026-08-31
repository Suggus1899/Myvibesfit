package service

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"myvibesfit/api/internal/domain"
)

var habitOwnerID = uuid.New()

type fakeHabits struct {
	domain.HabitRepository

	// alreadyLoggedToday simula el reenvio: el upsert actualiza la fila del
	// dia en vez de insertarla.
	alreadyLoggedToday bool
	activeCount        int64
	completedToday     int64
}

func (f *fakeHabits) UpsertLog(_ context.Context, l domain.HabitLog) (domain.HabitLog, error) {
	l.ID = 1
	l.Inserted = !f.alreadyLoggedToday
	return l, nil
}

func (f *fakeHabits) CountActiveForUser(context.Context, uuid.UUID) (int64, error) {
	return f.activeCount, nil
}

func (f *fakeHabits) CountCompletedLogsForDate(context.Context, uuid.UUID, time.Time) (int64, error) {
	return f.completedToday, nil
}

func (f *fakeHabits) GetClientHabitOwner(context.Context, uuid.UUID) (uuid.UUID, error) {
	return habitOwnerID, nil
}

func logHabitOnce(t *testing.T, habits *fakeHabits) *fakeGamification {
	t.Helper()
	gam := &fakeGamification{}
	repos := domain.TxRepos{Habits: habits, Gamification: gam}
	svc := NewHabitService(habits, &fakeUoW{repos: repos}, domain.NewGamificationService())

	if _, err := svc.LogHabit(context.Background(), LogHabitInput{
		ClientLocalID: uuid.New(), ClientHabitID: uuid.New(), UserID: habitOwnerID,
		LogDate: time.Now(), Value: 1, IsCompleted: true,
	}); err != nil {
		t.Fatalf("log habit: %v", err)
	}
	return gam
}

func habitXP(gam *fakeGamification) int32 {
	var total int32
	for _, e := range gam.xpEvents {
		if e.Source == domain.XpSourceHabit {
			total += e.Points
		}
	}
	return total
}

func TestLogHabitPremiaElPrimerRegistroDelDia(t *testing.T) {
	gam := logHabitOnce(t, &fakeHabits{activeCount: 1, completedToday: 1})

	if got := habitXP(gam); got != domain.XPPerHabitCheck {
		t.Fatalf("esperaba %d XP, fue %d", domain.XPPerHabitCheck, got)
	}
}

func TestLogHabitNoRepiteRecompensaEnUnReenvio(t *testing.T) {
	gam := logHabitOnce(t, &fakeHabits{alreadyLoggedToday: true, activeCount: 1, completedToday: 1})

	if got := habitXP(gam); got != 0 {
		t.Fatalf("un reenvio no deberia otorgar XP, fueron %d", got)
	}
}

func TestLogHabitRechazaHabitoAjeno(t *testing.T) {
	habits := &fakeHabits{activeCount: 1}
	repos := domain.TxRepos{Habits: habits, Gamification: &fakeGamification{}}
	svc := NewHabitService(habits, &fakeUoW{repos: repos}, domain.NewGamificationService())

	// habitOwnerID es el dueno del client_habit; este usuario es otro.
	_, err := svc.LogHabit(context.Background(), LogHabitInput{
		ClientLocalID: uuid.New(), ClientHabitID: uuid.New(), UserID: uuid.New(),
		LogDate: time.Now(), Value: 1, IsCompleted: true,
	})
	if err == nil {
		t.Fatal("marcar el habito de otro usuario deberia fallar")
	}
}
