package handler

import (
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

const dateLayout = "2006-01-02"

func pgText(t pgtype.Text) string {
	if !t.Valid {
		return ""
	}
	return t.String
}

func pgInt2Ptr(i pgtype.Int2) *int {
	if !i.Valid {
		return nil
	}
	v := int(i.Int16)
	return &v
}

func pgInt4Ptr(i pgtype.Int4) *int {
	if !i.Valid {
		return nil
	}
	v := int(i.Int32)
	return &v
}

func pgUUIDPtr(u pgtype.UUID) *uuid.UUID {
	if !u.Valid {
		return nil
	}
	id := uuid.UUID(u.Bytes)
	return &id
}

func datePtrToString(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.Format(dateLayout)
}

func parseDate(s string) (time.Time, error) {
	if s == "" {
		return time.Now().UTC().Truncate(24 * time.Hour), nil
	}
	return time.Parse(dateLayout, s)
}
