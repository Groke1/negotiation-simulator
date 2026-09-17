package pgxconv

import (
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

func StringToUUID(value string) (pgtype.UUID, error) {
	trimmed := strings.ToLower(strings.TrimSpace(value))
	trimmed = strings.ReplaceAll(trimmed, "-", "")
	if len(trimmed) != 32 {
		return pgtype.UUID{}, fmt.Errorf("invalid uuid")
	}

	decoded, err := hex.DecodeString(trimmed)
	if err != nil || len(decoded) != 16 {
		return pgtype.UUID{}, fmt.Errorf("invalid uuid")
	}

	var bytes [16]byte
	copy(bytes[:], decoded)
	return pgtype.UUID{Bytes: bytes, Valid: true}, nil
}

func TimestamptzToTime(t pgtype.Timestamptz) *time.Time {
	if t.Valid {
		return &t.Time
	}
	return nil
}

func TimeToTimestamptz(t *time.Time) pgtype.Timestamptz {
	if t == nil {
		return pgtype.Timestamptz{}
	}
	return pgtype.Timestamptz{Time: *t, Valid: true}
}

func TimeToRequiredTimestamptz(t time.Time) pgtype.Timestamptz {
	return pgtype.Timestamptz{Time: t, Valid: true}
}

func PgInt4ToInt32(pgInt pgtype.Int4) *int32 {
	if !pgInt.Valid {
		return nil
	}
	return &pgInt.Int32
}

func Int32ToPgInt4(val *int32) pgtype.Int4 {
	if val != nil {
		return pgtype.Int4{Int32: *val, Valid: true}
	}
	return pgtype.Int4{}
}

func PgUUIDtoString(pgUUID pgtype.UUID) *string {
	if !pgUUID.Valid {
		return nil
	}
	value := UUIDToString(pgUUID)
	return &value
}

func UUIDToString(pgUUID pgtype.UUID) string {
	if !pgUUID.Valid {
		return ""
	}
	b := pgUUID.Bytes
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}

func StringPtrToText(value *string) pgtype.Text {
	if value == nil {
		return pgtype.Text{}
	}
	return pgtype.Text{String: *value, Valid: true}
}

func TextToStringPtr(value pgtype.Text) *string {
	if !value.Valid {
		return nil
	}
	return &value.String
}

func Int64ToPgtype(value *int64) pgtype.Int8 {
	if value == nil {
		return pgtype.Int8{}
	}
	return pgtype.Int8{Int64: *value, Valid: true}
}

func BoolToPgtype(value *bool) pgtype.Bool {
	if value == nil {
		return pgtype.Bool{}
	}
	return pgtype.Bool{Bool: *value, Valid: true}
}
