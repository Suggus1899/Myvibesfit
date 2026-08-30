package domain

import "context"

// TxRepos agrupa los repos que participan en las 3 transacciones
// multi-agregado del backend (Assign, y en fase 3 SyncSessions/LogHabit).
// Se extiende agregando campos, no creando UnitOfWork[T] por flujo.
type TxRepos struct {
	Assignments  AssignmentRepository
	Sessions     SessionRepository
	Gamification GamificationRepository
	Habits       HabitRepository
	Coaches      CoachRepository
}

type UnitOfWork interface {
	Execute(ctx context.Context, fn func(TxRepos) error) error
}
