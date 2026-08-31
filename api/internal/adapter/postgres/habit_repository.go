package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"myvibesfit/api/internal/domain"
	"myvibesfit/api/internal/repository/db"
)

type HabitRepository struct{ q db.Querier }

func NewHabitRepository(q db.Querier) *HabitRepository { return &HabitRepository{q: q} }

var _ domain.HabitRepository = (*HabitRepository)(nil)

func (r *HabitRepository) ListHabits(ctx context.Context, orgID *uuid.UUID) ([]domain.Habit, error) {
	rows, err := r.q.ListHabits(ctx, uuidPtrToPgUUID(orgID))
	if err != nil {
		return nil, err
	}
	out := make([]domain.Habit, len(rows))
	for i, row := range rows {
		out[i] = toDomainHabit(row)
	}
	return out, nil
}

func (r *HabitRepository) GetHabitByID(ctx context.Context, id uuid.UUID) (domain.Habit, error) {
	row, err := r.q.GetHabitByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Habit{}, domain.ErrNotFound
		}
		return domain.Habit{}, err
	}
	return toDomainHabit(row), nil
}

func (r *HabitRepository) ListMyHabits(ctx context.Context, userID uuid.UUID) ([]domain.MyHabit, error) {
	rows, err := r.q.ListMyHabits(ctx, userID)
	if err != nil {
		return nil, err
	}
	out := make([]domain.MyHabit, len(rows))
	for i, row := range rows {
		out[i] = domain.MyHabit{
			ID: row.ID, UserID: row.UserID, HabitID: row.HabitID, TargetValue: row.TargetValue,
			Frequency: string(row.Frequency), DaysOfWeek: row.DaysOfWeek, StartedOn: row.StartedOn, EndedOn: row.EndedOn,
			HabitName: row.HabitName, HabitIcon: row.HabitIcon, HabitUnit: string(row.HabitUnit),
		}
	}
	return out, nil
}

func (r *HabitRepository) GetActiveClientHabit(ctx context.Context, userID, habitID uuid.UUID) (domain.ClientHabit, error) {
	row, err := r.q.GetActiveClientHabit(ctx, db.GetActiveClientHabitParams{UserID: userID, HabitID: habitID})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ClientHabit{}, domain.ErrNotFound
		}
		return domain.ClientHabit{}, err
	}
	return toDomainClientHabit(row), nil
}

func (r *HabitRepository) GetClientHabitOwner(ctx context.Context, clientHabitID uuid.UUID) (uuid.UUID, error) {
	userID, err := r.q.GetClientHabitOwner(ctx, clientHabitID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return uuid.Nil, domain.ErrNotFound
		}
		return uuid.Nil, err
	}
	return userID, nil
}

func (r *HabitRepository) SubscribeHabit(ctx context.Context, c domain.ClientHabit) (domain.ClientHabit, error) {
	row, err := r.q.SubscribeHabit(ctx, db.SubscribeHabitParams{
		UserID: c.UserID, HabitID: c.HabitID, AssignedBy: uuidPtrToPgUUID(c.AssignedBy),
		TargetValue: c.TargetValue, Frequency: db.HabitFrequency(c.Frequency), DaysOfWeek: c.DaysOfWeek,
	})
	if err != nil {
		return domain.ClientHabit{}, err
	}
	return toDomainClientHabit(row), nil
}

func (r *HabitRepository) UnsubscribeHabit(ctx context.Context, id, userID uuid.UUID) error {
	return r.q.UnsubscribeHabit(ctx, db.UnsubscribeHabitParams{ID: id, UserID: userID})
}

func (r *HabitRepository) UpsertLog(ctx context.Context, l domain.HabitLog) (domain.HabitLog, error) {
	row, err := r.q.UpsertHabitLog(ctx, db.UpsertHabitLogParams{
		ClientLocalID: l.ClientLocalID, ClientHabitID: l.ClientHabitID, UserID: l.UserID,
		LogDate: l.LogDate, Value: l.Value, IsCompleted: l.IsCompleted,
	})
	if err != nil {
		// Queda un 23505 posible: reusar un client_local_id ya gastado en
		// otro habito o dia. Es input invalido del cliente, no un 500 con el
		// error de Postgres crudo encima.
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return domain.HabitLog{}, domain.ErrAlreadyExists
		}
		return domain.HabitLog{}, err
	}
	// El upsert devuelve una fila propia (trae la bandera inserted), no
	// db.HabitLog, asi que se mapea aparte.
	return domain.HabitLog{
		ID: row.ID, ClientLocalID: row.ClientLocalID, ClientHabitID: row.ClientHabitID, UserID: row.UserID,
		LogDate: row.LogDate, Value: row.Value, IsCompleted: row.IsCompleted, LoggedAt: row.LoggedAt,
		Inserted: row.Inserted,
	}, nil
}

func (r *HabitRepository) ListLogsForDate(ctx context.Context, userID uuid.UUID, date time.Time) ([]domain.HabitLog, error) {
	rows, err := r.q.ListHabitLogsForDate(ctx, db.ListHabitLogsForDateParams{UserID: userID, LogDate: date})
	if err != nil {
		return nil, err
	}
	out := make([]domain.HabitLog, len(rows))
	for i, row := range rows {
		out[i] = toDomainHabitLog(row)
	}
	return out, nil
}

func (r *HabitRepository) CountActiveForUser(ctx context.Context, userID uuid.UUID) (int64, error) {
	return r.q.CountActiveHabitsForUser(ctx, userID)
}

func (r *HabitRepository) CountCompletedLogsForDate(ctx context.Context, userID uuid.UUID, date time.Time) (int64, error) {
	return r.q.CountCompletedHabitLogsForDate(ctx, db.CountCompletedHabitLogsForDateParams{UserID: userID, LogDate: date})
}

func toDomainHabit(h db.Habit) domain.Habit {
	return domain.Habit{
		ID: h.ID, OrgID: pgUUIDToPtr(h.OrgID), Slug: h.Slug, Name: h.Name, Icon: h.Icon,
		Unit: string(h.Unit), DefaultTarget: h.DefaultTarget, IsSystem: h.IsSystem, CreatedAt: h.CreatedAt,
	}
}

func toDomainClientHabit(c db.ClientHabit) domain.ClientHabit {
	return domain.ClientHabit{
		ID: c.ID, UserID: c.UserID, HabitID: c.HabitID, AssignedBy: pgUUIDToPtr(c.AssignedBy),
		TargetValue: c.TargetValue, Frequency: string(c.Frequency), DaysOfWeek: c.DaysOfWeek,
		StartedOn: c.StartedOn, EndedOn: c.EndedOn,
	}
}

func toDomainHabitLog(l db.HabitLog) domain.HabitLog {
	return domain.HabitLog{
		ID: l.ID, ClientLocalID: l.ClientLocalID, ClientHabitID: l.ClientHabitID, UserID: l.UserID,
		LogDate: l.LogDate, Value: l.Value, IsCompleted: l.IsCompleted, LoggedAt: l.LoggedAt,
	}
}
