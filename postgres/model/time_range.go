package model

import (
	"database/sql/driver"
	"fmt"
	"strings"
	"time"
)

type TimeRange struct {
	Lower time.Time
	Upper time.Time
	Valid bool
}

func (r *TimeRange) Scan(value any) error {
	var raw string

	switch v := value.(type) {
	case string:
		raw = v

	case []byte:
		raw = string(v)

	default:
		return fmt.Errorf(
			"cannot scan TimeRange from %T",
			value,
		)
	}

	if raw == "" || raw == "empty" {
		r.Valid = false
		return nil
	}

	if len(raw) < 2 {
		return fmt.Errorf("invalid tstzrange: %q", raw)
	}

	lowerType := raw[0]
	upperType := raw[len(raw)-1]

	if lowerType != '[' && lowerType != '(' {
		return fmt.Errorf("invalid lower bound: %q", raw)
	}

	if upperType != ']' && upperType != ')' {
		return fmt.Errorf("invalid upper bound: %q", raw)
	}

	content := raw[1 : len(raw)-1]

	parts := strings.SplitN(content, ",", 2)

	if len(parts) != 2 {
		return fmt.Errorf("invalid tstzrange: %q", raw)
	}

	lower, err := parseRangeTime(parts[0])
	if err != nil {
		return fmt.Errorf("parse lower bound: %w", err)
	}

	upper, err := parseRangeTime(parts[1])
	if err != nil {
		return fmt.Errorf("parse upper bound: %w", err)
	}

	r.Lower = lower
	r.Upper = upper
	r.Valid = true

	return nil
}

func parseRangeTime(value string) (time.Time, error) {
	value = strings.TrimSpace(value)
	value = strings.Trim(value, `"`)

	formats := []string{
		"2006-01-02 15:04:05.999999999Z07",
		"2006-01-02 15:04:05Z07",
	}

	for _, format := range formats {
		if parsed, err := time.Parse(format, value); err == nil {
			return parsed, nil
		}
	}

	return time.Time{}, fmt.Errorf(
		"unsupported timestamp format: %q",
		value,
	)
}

func (r TimeRange) Value() (driver.Value, error) {
	if !r.Valid {
		return nil, nil
	}

	return fmt.Sprintf(
		`["%s","%s")`,
		r.Lower.Format(time.RFC3339Nano),
		r.Upper.Format(time.RFC3339Nano),
	), nil
}
