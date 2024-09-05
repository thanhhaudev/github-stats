package clock

import "time"

type ClockKey struct{}

const DateTimeFormatWithTimezone = "2006-01-02 15:04:05 -0700"

type Clock interface {
	SetLocation(loc string) error
	ToClockTz(t time.Time) time.Time
	Now() time.Time
	After(d time.Duration) <-chan time.Time
	FromUnix(u int64) time.Time
	FromString(l, s string) (time.Time, error)
}

type clock struct {
	loc *time.Location
}

func (c *clock) Now() time.Time {
	return time.Now().In(c.loc)
}

func (c *clock) After(d time.Duration) <-chan time.Time {
	return time.After(d)
}

func (c *clock) FromUnix(u int64) time.Time {
	return time.Unix(u, 0).In(c.loc)
}

func (c *clock) FromString(l, s string) (time.Time, error) {
	val, err := time.Parse(l, s)
	if err != nil {
		return time.Time{}, err
	}

	return val.In(c.loc), nil
}

func (c *clock) ToClockTz(t time.Time) time.Time {
	return t.In(c.loc)
}

func (c *clock) SetLocation(loc string) error {
	val, err := time.LoadLocation(loc)
	if err != nil {
		return err
	}

	c.loc = val

	return nil
}

// NewClock creates a new clock
func NewClock() Clock {
	return &clock{
		loc: time.UTC,
	}
}
