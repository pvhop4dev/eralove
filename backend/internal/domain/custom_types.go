package domain

import (
	"encoding/json"
	"time"
)

// Date represents a date in YYYY-MM-DD format
type Date struct {
	time.Time
}

// UnmarshalJSON implements json.Unmarshaler interface for Date
func (d *Date) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}

	if s == "" || s == "null" {
		d.Time = time.Time{}
		return nil
	}

	// Try parsing different date formats
	formats := []string{
		"2006-01-02",                 // YYYY-MM-DD
		"2006-01-02T15:04:05Z07:00",  // RFC3339
		"2006-01-02T15:04:05Z",       // RFC3339 without timezone
		"2006-01-02 15:04:05",        // DateTime
	}

	var err error
	for _, format := range formats {
		d.Time, err = time.Parse(format, s)
		if err == nil {
			return nil
		}
	}

	return err
}

// MarshalJSON implements json.Marshaler interface for Date
func (d Date) MarshalJSON() ([]byte, error) {
	if d.Time.IsZero() {
		return []byte("null"), nil
	}
	return json.Marshal(d.Time.Format("2006-01-02"))
}

// String returns the date in YYYY-MM-DD format
func (d Date) String() string {
	if d.Time.IsZero() {
		return ""
	}
	return d.Time.Format("2006-01-02")
}

// ToTimePtr converts Date to *time.Time
func (d *Date) ToTimePtr() *time.Time {
	if d == nil || d.Time.IsZero() {
		return nil
	}
	t := d.Time
	return &t
}

// DateFromTime creates a Date from time.Time
func DateFromTime(t time.Time) Date {
	return Date{Time: t}
}

// DateFromTimePtr creates a Date from *time.Time
func DateFromTimePtr(t *time.Time) *Date {
	if t == nil {
		return nil
	}
	d := Date{Time: *t}
	return &d
}
