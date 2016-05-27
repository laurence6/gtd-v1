package gtd

import "time"

//
const (
	DateLayout = "2006-01-02"
	TimeLayout = "15:04"
)

// Time is a unix timestamp
type Time struct {
	sec  int64
	date string
	time string
}

// Get returns the unix timestamp
func (t *Time) Get() int64 {
	return t.sec
}

// EqualZero returns if Time == 0
func (t *Time) EqualZero() bool {
	return t.sec == 0
}

// Date returns string of the date
func (t *Time) Date() string {
	return t.date
}

// Time returns string of the time
func (t *Time) Time() string {
	return t.time
}

// Set sets the Time.sec
func (t *Time) Set(sec int64) {
	t.sec = sec
	if t.sec != 0 {
		datetime := time.Unix(t.sec, 0)
		t.date = datetime.Format(DateLayout)
		t.time = datetime.Format(TimeLayout)
	} else {
		t.date = ""
		t.time = ""
	}
	return
}

// ParseDateTimeInLocation parses date & time string
func (t *Time) ParseDateTimeInLocation(dateStr, timeStr string, location *time.Location) error {
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
