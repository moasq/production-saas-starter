// Package helpers provides utility functions for converting between Go types
// and PostgreSQL types (pgtype, pgvector). These helpers are used by repository
// implementations across all modules.
package helpers

import (
	"encoding/json"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/pgvector/pgvector-go"
)

// ToPgText converts a string to pgtype.Text
func ToPgText(s string) pgtype.Text {
	if s == "" {
		return pgtype.Text{Valid: false}
	}
	return pgtype.Text{String: s, Valid: true}
}

// FromPgText converts pgtype.Text to string
func FromPgText(t pgtype.Text) string {
	if !t.Valid {
		return ""
	}
	return t.String
}

// ToPgInt4 converts an int32 to pgtype.Int4
func ToPgInt4(i int32) pgtype.Int4 {
	return pgtype.Int4{Int32: i, Valid: true}
}

// ToPgInt4Ptr converts a pointer to int32 to pgtype.Int4
func ToPgInt4Ptr(i *int32) pgtype.Int4 {
	if i == nil {
		return pgtype.Int4{Valid: false}
	}
	return pgtype.Int4{Int32: *i, Valid: true}
}

// FromPgInt4 converts pgtype.Int4 to int32
func FromPgInt4(i pgtype.Int4) int32 {
	if !i.Valid {
		return 0
	}
	return i.Int32
}

// ToPgBool converts a bool to pgtype.Bool
func ToPgBool(b bool) pgtype.Bool {
	return pgtype.Bool{Bool: b, Valid: true}
}

// ToPgBoolPtr converts a pointer to bool to pgtype.Bool
func ToPgBoolPtr(b *bool) pgtype.Bool {
	if b == nil {
		return pgtype.Bool{Valid: false}
	}
	return pgtype.Bool{Bool: *b, Valid: true}
}

// FromPgBool converts pgtype.Bool to bool
func FromPgBool(b pgtype.Bool) bool {
	if !b.Valid {
		return false
	}
	return b.Bool
}

// ToJSONB converts a map to JSON bytes
func ToJSONB(m map[string]any) []byte {
	if m == nil {
		return []byte("{}")
	}
	data, err := json.Marshal(m)
	if err != nil {
		return []byte("{}")
	}
	return data
}

// FromJSONB converts JSON bytes to a map
func FromJSONB(b []byte) map[string]any {
	if len(b) == 0 {
		return nil
	}
	var result map[string]any
	if err := json.Unmarshal(b, &result); err != nil {
		return nil
	}
	return result
}

// ToVector converts a float64 slice to pgvector.Vector
func ToVector(embedding []float64) pgvector.Vector {
	// Convert []float64 to []float32 for pgvector
	f32 := make([]float32, len(embedding))
	for i, v := range embedding {
		f32[i] = float32(v)
	}
	return pgvector.NewVector(f32)
}

// FromVector converts pgvector.Vector to float64 slice
func FromVector(v pgvector.Vector) []float64 {
	f32 := v.Slice()
	result := make([]float64, len(f32))
	for i, val := range f32 {
		result[i] = float64(val)
	}
	return result
}
