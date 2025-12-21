package postgres

import (
	"encoding/hex"
	"fmt"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/shopspring/decimal"

	geomPkg "github.com/twpayne/go-geom"
	"github.com/twpayne/go-geom/encoding/ewkb"
	"github.com/twpayne/go-geom/encoding/wkb"
)

// Int16Ptr converts pgtype.Int2 to *int16
func Int16Ptr(i pgtype.Int2) *int16 {
	if !i.Valid {
		return nil
	}
	return &i.Int16
}

// Int32Ptr converts pgtype.Int4 to *int32
func Int32Ptr(i pgtype.Int4) *int32 {
	if !i.Valid {
		return nil
	}
	return &i.Int32
}

// Int64Ptr converts pgtype.Int8 to *int64
func Int64Ptr(i pgtype.Int8) *int64 {
	if !i.Valid {
		return nil
	}
	return &i.Int64
}

// Float32Ptr converts pgtype.Float4 to *float32
func Float32Ptr(f pgtype.Float4) *float32 {
	if !f.Valid {
		return nil
	}
	return &f.Float32
}

// Float64Ptr converts pgtype.Float8 to *float64
func Float64Ptr(f pgtype.Float8) *float64 {
	if !f.Valid {
		return nil
	}
	return &f.Float64
}

// StringPtr converts pgtype.Text to *string
func StringPtr(t pgtype.Text) *string {
	if !t.Valid {
		return nil
	}
	return &t.String
}

// TimeStampPtr converts pgtype.Timestamp to *time.Time
func TimeStampPtr(t pgtype.Timestamp) *time.Time {
	if !t.Valid {
		return nil
	}
	return &t.Time
}

// time tampz
// TimeStampTzPtr converts pgtype.Timestamptz to *time.Time
func TimeStampTzPtr(t pgtype.Timestamptz) *time.Time {
	if !t.Valid {
		return nil
	}
	return &t.Time
}

// time ptr
// TimePtr converts pgtype.Time to *time.Time
func TimePtr(t pgtype.Time) *time.Time {
	if !t.Valid {
		return nil
	}
	// Convert microseconds to time.Time
	seconds := t.Microseconds / 1_000_000        // Convert to seconds
	nanos := (t.Microseconds % 1_000_000) * 1000 // Convert remaining microseconds to nanoseconds
	timeVal := time.Unix(seconds, nanos)
	return &timeVal
}

// BoolPtr converts pgtype.Bool to *bool
func BoolPtr(b pgtype.Bool) *bool {
	if !b.Valid {
		return nil
	}
	return &b.Bool
}

// from go types to pg types
// PgInt2 converts *int16 to pgtype.PgInt2
func PgInt2(i *int16) pgtype.Int2 {
	if i == nil {
		return pgtype.Int2{Valid: false}
	}
	return pgtype.Int2{Int16: *i, Valid: true}
}

// PgInt4 converts *int32 to pgtype.PgInt4
func PgInt4(i *int32) pgtype.Int4 {
	if i == nil {
		return pgtype.Int4{Valid: false}
	}
	return pgtype.Int4{Int32: *i, Valid: true}
}

// PgInt8 converts *int64 to pgtype.PgInt8
func PgInt8(i *int64) pgtype.Int8 {
	if i == nil {
		return pgtype.Int8{Valid: false}
	}
	return pgtype.Int8{Int64: *i, Valid: true}
}

// PgFloat4 converts *float32 to pgtype.PgFloat4
func PgFloat4(f *float32) pgtype.Float4 {
	if f == nil {
		return pgtype.Float4{Valid: false}
	}
	return pgtype.Float4{Float32: *f, Valid: true}
}

// PgFloat8 converts *float64 to pgtype.PgFloat8
func PgFloat8(f *float64) pgtype.Float8 {
	if f == nil {
		return pgtype.Float8{Valid: false}
	}
	return pgtype.Float8{Float64: *f, Valid: true}
}

// PgText converts *string to pgtype.PgText
func PgText(s *string) pgtype.Text {
	if s == nil {
		return pgtype.Text{Valid: false}
	}
	return pgtype.Text{String: *s, Valid: true}
}

// PgTimestamp converts *time.Time to pgtype.PgTimestamp
func PgTimestamp(t *time.Time) pgtype.Timestamp {
	if t == nil {
		return pgtype.Timestamp{Valid: false}
	}
	return pgtype.Timestamp{Time: *t, Valid: true}
}

// PgTimestamptz converts *time.Time to pgtype.PgTimestamptz
func PgTimestamptz(t *time.Time) pgtype.Timestamptz {
	if t == nil {
		return pgtype.Timestamptz{Valid: false}
	}
	return pgtype.Timestamptz{Time: *t, Valid: true}
}

// PgTime converts *time.Time to pgtype.PgTime
func PgTime(t *time.Time) pgtype.Time {
	if t == nil {
		return pgtype.Time{Valid: false}
	}
	// Convert time.Time to microseconds
	microseconds := t.Unix()*1_000_000 + int64(t.Nanosecond())/1000
	return pgtype.Time{
		Microseconds: microseconds,
		Valid:        true,
	}
}

// PgBool converts *bool to pgtype.PgBool
func PgBool(b *bool) pgtype.Bool {
	if b == nil {
		return pgtype.Bool{Valid: false}
	}
	return pgtype.Bool{Bool: *b, Valid: true}
}

// UUIDPtr converts pgtype.UUID to *uuid.UUID
func UUIDPtr(u pgtype.UUID) *uuid.UUID {
	if !u.Valid {
		return nil
	}
	id := uuid.UUID(u.Bytes)
	return &id
}

// PgUUID converts *uuid.UUID to pgtype.UUID
func PgUUID(u *uuid.UUID) pgtype.UUID {
	if u == nil {
		return pgtype.UUID{Valid: false}
	}
	return pgtype.UUID{
		Bytes: [16]byte(*u),
		Valid: true,
	}
}

// NumericPtr converts pgtype.Numeric to *float64 with improved error handling
func NumericPtr(n pgtype.Numeric) *float64 {
	if !n.Valid {
		return nil
	}

	// Use Value() method for safe conversion
	val, err := n.Value()
	if err != nil {
		return nil
	}

	// Handle different return types from PostgreSQL numeric
	switch v := val.(type) {
	case float64:
		return &v
	case string:
		// Parse string representation like "1315.0000"
		if f64, err := strconv.ParseFloat(v, 64); err == nil {
			return &f64
		}
	case int64:
		f64 := float64(v)
		return &f64
	case int32:
		f64 := float64(v)
		return &f64
	case int:
		f64 := float64(v)
		return &f64
	}
	return nil
}

// Numeric converts *float64 to pgtype.Numeric with improved accuracy
func Numeric(f *float64) pgtype.Numeric {
	if f == nil {
		return pgtype.Numeric{Valid: false}
	}

	// Use Scan for proper conversion
	numeric := pgtype.Numeric{}
	err := numeric.Scan(fmt.Sprintf("%.6f", *f))
	if err != nil {
		return pgtype.Numeric{Valid: false}
	}
	return numeric
}

// numeric from decimal
// NumericFromDecimal converts decimal.Decimal to pgtype.Numeric
func NumericFromDecimal(d *decimal.Decimal) pgtype.Numeric {
	if d == nil {
		return pgtype.Numeric{Valid: false}
	}
	numeric := pgtype.Numeric{}
	err := numeric.Scan(d.String())
	if err != nil {
		return pgtype.Numeric{Valid: false}
	}
	return numeric
}

// NumericFromFloat32 converts float32 to pgtype.Numeric
func NumericFromFloat32(f float32) pgtype.Numeric {
	fmt.Printf("[NUMERIC-DEBUG] Converting float32 to Numeric: %f\n", f)
	numeric := pgtype.Numeric{}
	err := numeric.Scan(fmt.Sprintf("%.6f", f))
	if err != nil {
		fmt.Printf("[NUMERIC-DEBUG] Failed to convert float32 to Numeric: %v\n", err)
		return pgtype.Numeric{Valid: false}
	}
	fmt.Printf("[NUMERIC-DEBUG] Numeric conversion successful, Valid: %v\n", numeric.Valid)
	return numeric
}

// Float32FromNumeric converts pgtype.Numeric to float32
func Float32FromNumeric(n pgtype.Numeric) float32 {
	fmt.Printf("[NUMERIC-DEBUG] Converting Numeric to float32, Valid: %v\n", n.Valid)
	if !n.Valid {
		fmt.Printf("[NUMERIC-DEBUG] Numeric is invalid, returning 0\n")
		return 0
	}

	val, err := n.Value()
	if err != nil {
		fmt.Printf("[NUMERIC-DEBUG] Failed to get value from Numeric: %v\n", err)
		return 0
	}

	fmt.Printf("[NUMERIC-DEBUG] Numeric value: %v (type: %T)\n", val, val)

	// Handle different return types from PostgreSQL numeric
	switch v := val.(type) {
	case float64:
		result := float32(v)
		fmt.Printf("[NUMERIC-DEBUG] Converted from float64: %f\n", result)
		return result
	case string:
		// Parse string representation like "0.9000"
		if f64, err := strconv.ParseFloat(v, 64); err == nil {
			result := float32(f64)
			fmt.Printf("[NUMERIC-DEBUG] Converted from string: %f\n", result)
			return result
		}
		fmt.Printf("[NUMERIC-DEBUG] Failed to parse string: %s\n", v)
	case int64:
		result := float32(v)
		fmt.Printf("[NUMERIC-DEBUG] Converted from int64: %f\n", result)
		return result
	case int32:
		result := float32(v)
		fmt.Printf("[NUMERIC-DEBUG] Converted from int32: %f\n", result)
		return result
	case int:
		result := float32(v)
		fmt.Printf("[NUMERIC-DEBUG] Converted from int: %f\n", result)
		return result
	}
	fmt.Printf("[NUMERIC-DEBUG] No matching type conversion, returning 0\n")
	return 0
}

// DatePtr converts pgtype.Date to *time.Time
func DatePtr(d pgtype.Date) *time.Time {
	if !d.Valid {
		return nil
	}
	return &d.Time
}

// PgDate converts *time.Time to pgtype.Date
func PgDate(t *time.Time) pgtype.Date {
	if t == nil {
		return pgtype.Date{Valid: false}
	}
	return pgtype.Date{Time: *t, Valid: true}
}

// PgTextFromString converts string to pgtype.Text (for non-empty strings)
func PgTextFromString(s string) pgtype.Text {
	if s == "" {
		return pgtype.Text{Valid: false}
	}
	return pgtype.Text{String: s, Valid: true}
}

// StringFromPgText converts pgtype.Text to string (empty string if invalid)
func StringFromPgText(t pgtype.Text) string {
	if !t.Valid {
		return ""
	}
	return t.String
}

// PgInt4FromInt32 converts int32 to pgtype.Int4
func PgInt4FromInt32(i int32) pgtype.Int4 {
	return pgtype.Int4{Int32: i, Valid: true}
}

// Int32FromPgInt4 converts pgtype.Int4 to int32 (0 if invalid)
func Int32FromPgInt4(i pgtype.Int4) int32 {
	if !i.Valid {
		return 0
	}
	return i.Int32
}

func ConvertWKBToPoint(wkbHex string) (*geomPkg.Point, error) {
	// Decode the hex string to bytes
	wkbBytes, err := hex.DecodeString(wkbHex)
	if err != nil {
		return nil, fmt.Errorf("failed to decode WKB hex: %w", err)
	}

	// Parse the WKB bytes
	geom, err := wkb.Unmarshal(wkbBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal WKB: %w", err)
	}

	// Assert that it's a Point
	geom.Bounds()
	point, ok := geom.(*geomPkg.Point)
	if !ok {
		return nil, fmt.Errorf("geometry is not a Point")
	}

	return point, nil
}

func ConvertWKBToPointString(wkbHex string) (string, error) {
	// Decode the hex string to bytes
	wkbBytes, err := hex.DecodeString(wkbHex)
	if err != nil {
		return "", fmt.Errorf("failed to decode WKB hex: %w", err)
	}

	// Parse the EWKB bytes
	g, err := ewkb.Unmarshal(wkbBytes)
	if err != nil {
		return "", fmt.Errorf("failed to unmarshal EWKB: %w", err)
	}

	// Get the coordinates
	coords := g.FlatCoords()
	if len(coords) < 2 {
		return "", fmt.Errorf("invalid point data")
	}

	// Convert to "POINT(longitude latitude)" format
	return fmt.Sprintf("POINT(%.6f %.6f)", coords[0], coords[1]), nil
}
