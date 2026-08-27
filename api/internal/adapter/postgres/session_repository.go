package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"myvibesfit/api/internal/domain"
	"myvibesfit/api/internal/repository/db"
)

type SessionRepository struct{ q db.Querier }

func NewSessionRepository(q db.Querier) *SessionRepository { return &SessionRepository{q: q} }

var _ domain.SessionRepository = (*SessionRepository)(nil)

func (r *SessionRepository) UpsertWorkoutSession(ctx context.Context, s domain.WorkoutSession) (domain.WorkoutSession, error) {
	row, err := r.q.UpsertWorkoutSession(ctx, db.UpsertWorkoutSessionParams{
		ClientLocalID: s.ClientLocalID, UserID: s.UserID, OrgID: uuidPtrToPgUUID(s.OrgID),
		AssignedWorkoutID: uuidPtrToPgUUID(s.AssignedWorkoutID), Name: s.Name, Status: db.SessionStatus(s.Status),
		StartedAt: s.StartedAt, EndedAt: s.EndedAt, DurationSeconds: intPtrToPgInt4(s.DurationSeconds),
		TotalVolumeKg: s.TotalVolumeKg, PerceivedEffort: intPtrToPgInt2(s.PerceivedEffort),
		Mood: intPtrToPgInt2(s.Mood), Notes: stringToPgText(s.Notes),
	})
	if err != nil {
		return domain.WorkoutSession{}, err
	}
	return toDomainWorkoutSession(row), nil
}

func (r *SessionRepository) UpsertSessionExercise(ctx context.Context, e domain.SessionExercise) (domain.SessionExercise, error) {
	row, err := r.q.UpsertSessionExercise(ctx, db.UpsertSessionExerciseParams{
		SessionID: e.SessionID, ExerciseID: e.ExerciseID, AssignedExerciseID: uuidPtrToPgUUID(e.AssignedExerciseID),
		OrderIndex: e.OrderIndex, SupersetGroup: intPtrToPgInt2(e.SupersetGroup), Note: stringToPgText(e.Note),
	})
	if err != nil {
		return domain.SessionExercise{}, err
	}
	return toDomainSessionExercise(row), nil
}

func (r *SessionRepository) UpsertSetLog(ctx context.Context, s domain.SetLog) (domain.SetLog, error) {
	row, err := r.q.UpsertSetLog(ctx, db.UpsertSetLogParams{
		ClientLocalID: s.ClientLocalID, SessionExerciseID: s.SessionExerciseID, UserID: s.UserID, ExerciseID: s.ExerciseID,
		SetNumber: s.SetNumber, Type: db.SetType(s.Type), WeightKg: s.WeightKg, Reps: intPtrToPgInt2(s.Reps),
		Rpe: s.RPE, Rir: intPtrToPgInt2(s.RIR), DurationSeconds: intPtrToPgInt4(s.DurationSeconds), DistanceM: s.DistanceM,
		RestTakenSeconds: intPtrToPgInt4(s.RestTakenSeconds), IsCompleted: s.IsCompleted, PerformedAt: s.PerformedAt,
	})
	if err != nil {
		return domain.SetLog{}, err
	}
	return toDomainSetLog(row), nil
}

func (r *SessionRepository) GetPersonalRecord(ctx context.Context, userID, exerciseID uuid.UUID, recordType domain.RecordType) (domain.PersonalRecord, error) {
	row, err := r.q.GetPersonalRecord(ctx, db.GetPersonalRecordParams{UserID: userID, ExerciseID: exerciseID, Type: db.RecordType(recordType)})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.PersonalRecord{}, domain.ErrNotFound
		}
		return domain.PersonalRecord{}, err
	}
	return toDomainPersonalRecord(row), nil
}

func (r *SessionRepository) UpsertPersonalRecord(ctx context.Context, rec domain.PersonalRecord) (domain.PersonalRecord, error) {
	row, err := r.q.UpsertPersonalRecord(ctx, db.UpsertPersonalRecordParams{
		UserID: rec.UserID, ExerciseID: rec.ExerciseID, Type: db.RecordType(rec.Type), Value: rec.Value,
		SetLogID: int64PtrToPgInt8(rec.SetLogID), AchievedAt: rec.AchievedAt,
	})
	if err != nil {
		return domain.PersonalRecord{}, err
	}
	return toDomainPersonalRecord(row), nil
}

func (r *SessionRepository) CountCompletedSessionsOnDate(ctx context.Context, userID uuid.UUID, date time.Time) (int64, error) {
	return r.q.CountCompletedSessionsOnDate(ctx, db.CountCompletedSessionsOnDateParams{UserID: userID, Column2: date})
}

func toDomainWorkoutSession(s db.WorkoutSession) domain.WorkoutSession {
	return domain.WorkoutSession{
		ID: s.ID, ClientLocalID: s.ClientLocalID, UserID: s.UserID, OrgID: pgUUIDToPtr(s.OrgID),
		AssignedWorkoutID: pgUUIDToPtr(s.AssignedWorkoutID), Name: s.Name, Status: domain.SessionStatus(s.Status),
		StartedAt: s.StartedAt, EndedAt: s.EndedAt, DurationSeconds: pgInt4ToPtr(s.DurationSeconds),
		TotalVolumeKg: s.TotalVolumeKg, PerceivedEffort: pgInt2ToPtr(s.PerceivedEffort), Mood: pgInt2ToPtr(s.Mood),
		Notes: pgTextToString(s.Notes), SyncedAt: s.SyncedAt, CreatedAt: s.CreatedAt,
	}
}

func toDomainSessionExercise(e db.SessionExercise) domain.SessionExercise {
	return domain.SessionExercise{
		ID: e.ID, SessionID: e.SessionID, ExerciseID: e.ExerciseID, AssignedExerciseID: pgUUIDToPtr(e.AssignedExerciseID),
		OrderIndex: e.OrderIndex, SupersetGroup: pgInt2ToPtr(e.SupersetGroup), Note: pgTextToString(e.Note),
	}
}

func toDomainSetLog(s db.SetLog) domain.SetLog {
	return domain.SetLog{
		ID: s.ID, ClientLocalID: s.ClientLocalID, SessionExerciseID: s.SessionExerciseID, UserID: s.UserID,
		ExerciseID: s.ExerciseID, SetNumber: s.SetNumber, Type: domain.SetType(s.Type), WeightKg: s.WeightKg,
		Reps: pgInt2ToPtr(s.Reps), RPE: s.Rpe, RIR: pgInt2ToPtr(s.Rir), DurationSeconds: pgInt4ToPtr(s.DurationSeconds),
		DistanceM: s.DistanceM, RestTakenSeconds: pgInt4ToPtr(s.RestTakenSeconds), IsCompleted: s.IsCompleted, PerformedAt: s.PerformedAt,
	}
}

func toDomainPersonalRecord(r db.PersonalRecord) domain.PersonalRecord {
	return domain.PersonalRecord{
		ID: r.ID, UserID: r.UserID, ExerciseID: r.ExerciseID, Type: domain.RecordType(r.Type), Value: r.Value,
		SetLogID: pgInt8ToPtr(r.SetLogID), AchievedAt: r.AchievedAt,
	}
}
