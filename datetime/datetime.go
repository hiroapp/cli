// Package datetime provides utilties for dealing with
// dates and times.
package datetime

import (
	"fmt"
	"strings"
	"time"
)

// ParsePeriod returns the Period for s, or an error.
func ParsePeriod(s string) (Period, error) {
	switch strings.ToLower(s) {
	case "day":
		return Day, nil
	case "week":
		return Week, nil
	case "month":
		return Month, nil
	case "year":
		return Year, nil
	default:
		return 0, fmt.Errorf("bad period: %s", s)
	}
}

// Period holds a calendar period unit which does not have a fixed
// time.Period due to leap seconds, leap years, days per month, etc.
type Period int

const (
	// Day represents a calendar day.
	Day Period = iota
	// Day represents a calendar week.
	Week
	// Month represents a calendar month.
	Month
	// Year represents a calendar year.
	Year
)

func NewIterator(cursor time.Time, period Period, asc bool, firstWeekday time.Weekday) *Iterator {
	return &Iterator{
		cursor:       normalizeCursor(cursor, period, firstWeekday),
		period:       period,
		asc:          asc,
		firstWeekday: firstWeekday,
	}
}

func normalizeCursor(cursor time.Time, period Period, firstDay time.Weekday) time.Time {
	loc := cursor.Location()
	switch period {
	case Day, Week:
		cursor = time.Date(cursor.Year(), cursor.Month(), cursor.Day(), 0, 0, 0, 0, loc)
	case Month:
		cursor = time.Date(cursor.Year(), cursor.Month(), 1, 0, 0, 0, 0, loc)
	case Year:
		cursor = time.Date(cursor.Year(), 1, 1, 0, 0, 0, 0, loc)
	}
	if period == Week {
		for cursor.Weekday() != firstDay {
			cursor = cursor.AddDate(0, 0, -1)
		}
	}
	return cursor
}

type Iterator struct {
	cursor       time.Time
	period       Period
	asc          bool
	firstWeekday time.Weekday
}

func (i *Iterator) Next() (time.Time, time.Time) {
	var (
		next, from, to time.Time
		m              = 1
	)
	if !i.asc {
		m = -1
	}
	switch i.period {
	case Day:
		next = i.cursor.AddDate(0, 0, m*1)
		from = i.cursor
		to = from.AddDate(0, 0, 1).Add(-time.Nanosecond)
	case Week:
		next = i.cursor.AddDate(0, 0, m*7)
		from = i.cursor
		to = from.AddDate(0, 0, 7).Add(-time.Nanosecond)
	case Month:
		next = i.cursor.AddDate(0, m*1, 0)
		from = i.cursor
		to = from.AddDate(0, 1, 0).Add(-time.Nanosecond)
	case Year:
		next = i.cursor.AddDate(m*1, 0, 0)
		from = i.cursor
		to = from.AddDate(1, 0, 0).Add(-time.Nanosecond)
	}
	i.cursor = next
	return from, to
}

func ParseWeekday(s string) (time.Weekday, error) {
	switch strings.ToLower(s) {
	case "monday":
		return time.Monday, nil
	case "tuesday":
		return time.Tuesday, nil
	case "wednesday":
		return time.Wednesday, nil
	case "thursday":
		return time.Thursday, nil
	case "friday":
		return time.Friday, nil
	case "saturday":
		return time.Saturday, nil
	case "sunday":
		return time.Sunday, nil
	default:
		return 0, fmt.Errorf("bad weekday: %s", s)
	}

}
