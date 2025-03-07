package model

import (
	"database/sql/driver"
	"fmt"
	"time"
)

// rfc3339Milli is like time.RFC3339Nano, but with millisecond precision, and fractional seconds do not have trailing
// zeros removed.
const rfc3339Milli = "2006-01-02T15:04:05.000Z07:00"

type Time struct {
	T time.Time
}

func (t *Time) String() string {
	if t == nil {
		return ""
	}
	return t.T.UTC().Format(rfc3339Milli)
}

func ParseTime(v string) (Time, error) {
	t, err := time.Parse(rfc3339Milli, v)
	if err != nil {
		return Time{}, err
	}
	return Time{T: t}, nil
}

// Value satisfies driver.Valuer interface.
func (t Time) Value() (driver.Value, error) {
	return t.T.UTC().Format(rfc3339Milli), nil
}

// Scan satisfies sql.Scanner interface.
func (t *Time) Scan(src any) error {
	if src == nil {
		return nil
	}

	s, ok := src.(string)
	if !ok {
		return fmt.Errorf("error scanning time, got %+v", src)
	}

	parsedT, err := time.Parse(rfc3339Milli, s)
	if err != nil {
		return err
	}

	t.T = parsedT.UTC()

	return nil
}

func Now() *Time {
	return &Time{T: time.Now()}
}

func (t *Time) MarshalText() ([]byte, error) {
	return []byte(t.String()), nil
}

func (t *Time) UnmarshalText(data []byte) error {
	if len(data) == 0 {
		return nil
	}

	parsedT, err := ParseTime(string(data))
	if err != nil {
		return err
	}

	t.T = parsedT.T.UTC()

	return nil
}
