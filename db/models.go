package db

import (
	"errors"
	"time"
)

type Entry struct {
	ID       string
	Category []string
	Start    time.Time
	End      time.Time
	Note     string
}

func (e Entry) Valid() error {
	if e.Start.IsZero() {
		return errors.New("start time is required")
	} else if !e.End.IsZero() && e.End.Before(e.Start) {
		return errors.New("end time must be after start time")
	}
	return nil
}

func (e Entry) Duration(now time.Time) time.Duration {
	end := e.End
	if end.IsZero() {
		end = now
	}
	return end.Sub(e.Start)
}

func (e *Entry) Equal(o *Entry) bool {
	return e == o ||
		(e.ID == o.ID &&
			e.Start.Equal(o.Start) &&
			e.End.Equal(o.End) &&
			e.Note == o.Note &&
			stringSliceEqual(e.Category, o.Category))
}

func stringSliceEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i, av := range a {
		if b[i] != av {
			return false
		}
	}
	return true
}
