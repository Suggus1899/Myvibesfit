package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"myvibesfit/api/internal/domain"
	"myvibesfit/api/internal/repository/db"
)

type UnitOfWork struct{ pool *pgxpool.Pool }

func NewUnitOfWork(pool *pgxpool.Pool) *UnitOfWork { return &UnitOfWork{pool: pool} }

var _ domain.UnitOfWork = (*UnitOfWork)(nil)

func (u *UnitOfWork) Execute(ctx context.Context, fn func(domain.TxRepos) error) error {
	tx, err := u.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx) //nolint:errcheck // no-op si ya hizo commit

	qtx := db.New(tx)
	repos := domain.TxRepos{
		Assignments:  NewAssignmentRepository(qtx),
		Sessions:     NewSessionRepository(qtx),
		Gamification: NewGamificationRepository(qtx),
		Habits:       NewHabitRepository(qtx),
		Coaches:      NewCoachRepository(qtx),
		Suggestions:  NewAISuggestionRepository(qtx),
	}
	if err := fn(repos); err != nil {
		return err
	}
	return tx.Commit(ctx)
}
