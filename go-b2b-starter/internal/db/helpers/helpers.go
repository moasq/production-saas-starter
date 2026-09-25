// Package helpers provides utility functions for converting between Go types
// and PostgreSQL types (pgtype). These helpers are used by repository
// implementations across all modules.
package helpers

import (
	"github.com/jackc/pgx/v5/pgtype"
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
