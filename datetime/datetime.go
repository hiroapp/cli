// Package datetime provides utilties for dealing with
// dates and times.
package datetime

import "time"

type Duration int

const (
	Day Duration = iota
	Week
	Month
	Year
)

func NewIterator(cursor time.Time, duration Duration, asc bool, firstWeekday time.Weekday) *Iterator {
	return &Iterator{
		cursor:       normalizeCursor(cursor, duration, firstWeekday),
		duration:     duration,
		asc:          asc,
		firstWeekday: firstWeekday,
	}
}

func normalizeCursor(cursor time.Time, duration Duration, firstDay time.Weekday) time.Time {
	loc := cursor.Location()
	switch duration {
	case Day, Week:
		cursor = time.Date(cursor.Year(), cursor.Month(), cursor.Day(), 0, 0, 0, 0, loc)
	case Month:
		cursor = time.Date(cursor.Year(), cursor.Month(), 1, 0, 0, 0, 0, loc)
	case Year:
		cursor = time.Date(cursor.Year(), 1, 1, 0, 0, 0, 0, loc)
	}
	if duration == Week {
		for cursor.Weekday() != firstDay {
			cursor = cursor.AddDate(0, 0, -1)
		}
	}
	return cursor
}

type Iterator struct {
	cursor       time.Time
	duration     Duration
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
	switch i.duration {
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
