package model

import (
	"database/sql/driver"
	"strconv"
	"time"
)

const (
	DateLayout = "2006-01-02"
	TimeLayout = "15:04"
)

// Time is a unix timestamp.
type Time struct {
	sec int64
}

func NewTime(sec int64) Time {
	return Time{
		sec: sec,
	}
}

// Set sets Time.sec.
func (t *Time) Set(sec int64) {
	t.sec = sec
}

// Get returns Time.sec.
func (t Time) Get() int64 {
	return t.sec
}

// Date returns string of the date.
func (t Time) Date() string {
	if t.sec == 0 {
		return ""
	}

	datetime := time.Unix(t.sec, 0)
	return datetime.Format(DateLayout)
}

// Time returns string of the time.
func (t Time) Time() string {
	if t.sec == 0 {
		return ""
	}

	datetime := time.Unix(t.sec, 0)
	return datetime.Format(TimeLayout)
}

// EqualZero returns if Time == 0.
func (t Time) EqualZero() bool {
	return t.sec == 0
}

// ParseDateTimeInLocation parses date & time string and sets Time.sec. If date is empty, Time.sec = 0. If date is not empty and time is empty, only date will be parsed.
func (t *Time) ParseDateTimeInLocation(dateStr string, timeStr string, location *time.Location) error {
	if dateStr != "" {
		var datetime time.Time
		var err error

		if timeStr != "" {
			datetime, err = time.ParseInLocation(DateLayout+TimeLayout, dateStr+timeStr, location)
		} else {
			datetime, err = time.ParseInLocation(DateLayout, dateStr, location)
		}
		if err != nil {
			return err
		}

		t.Set(datetime.Unix())
		return nil
	}
	t.Set(0)
	return nil
}

// Scan implements sql.Scanner interface.
func (t *Time) Scan(val interface{}) error {
	secStr, ok := val.([]uint8)
	if !ok {
		return nil
	}

	sec, err := strconv.ParseInt(string(secStr), 10, 64)
	if err != nil {
		return err
	}

	t.Set(sec)
	return nil
}

// Value implements sql.Valuer interface.
func (t Time) Value() (driver.Value, error) {
	return t.sec, nil
}

// String converts Time to string.
func (t Time) String() string {
	return strconv.FormatInt(t.sec, 10)
}
