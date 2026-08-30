// Package postgres implementa los ports de internal/domain contra
// PostgreSQL via el codigo generado por sqlc en internal/repository/db.
// Es el unico lugar del backend (fuera de internal/repository/db) que puede
// importar pgx/pgtype.
package postgres

import (
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

func pgTextToString(t pgtype.Text) string {
	if !t.Valid {
		return ""
	}
	return t.String
}

func stringToPgText(s string) pgtype.Text {
	if s == "" {
		return pgtype.Text{}
	}
	return pgtype.Text{String: s, Valid: true}
}

// pgTextToPtr conserva la diferencia entre NULL y cadena vacia, a diferencia
// de pgTextToString. Para columnas donde "no hay valor" es semanticamente
// distinto de "" — override_source, por ejemplo.
func pgTextToPtr(t pgtype.Text) *string {
	if !t.Valid {
		return nil
	}
	s := t.String
	return &s
}

func pgUUIDToPtr(u pgtype.UUID) *uuid.UUID {
	if !u.Valid {
		return nil
	}
	id := uuid.UUID(u.Bytes)
	return &id
}

func uuidPtrToPgUUID(id *uuid.UUID) pgtype.UUID {
	if id == nil {
		return pgtype.UUID{}
	}
	return pgtype.UUID{Bytes: *id, Valid: true}
}

func pgInt2ToPtr(i pgtype.Int2) *int {
	if !i.Valid {
		return nil
	}
	v := int(i.Int16)
	return &v
}

func intPtrToPgInt2(v *int) pgtype.Int2 {
	if v == nil {
		return pgtype.Int2{}
	}
	return pgtype.Int2{Int16: int16(*v), Valid: true}
}

func pgInt4ToPtr(i pgtype.Int4) *int {
	if !i.Valid {
		return nil
	}
	v := int(i.Int32)
	return &v
}

func intPtrToPgInt4(v *int) pgtype.Int4 {
	if v == nil {
		return pgtype.Int4{}
	}
	return pgtype.Int4{Int32: int32(*v), Valid: true}
}

func pgInt8ToPtr(i pgtype.Int8) *int64 {
	if !i.Valid {
		return nil
	}
	v := i.Int64
	return &v
}

func int64PtrToPgInt8(v *int64) pgtype.Int8 {
	if v == nil {
		return pgtype.Int8{}
	}
	return pgtype.Int8{Int64: *v, Valid: true}
}
