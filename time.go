package hiro

import (
	"fmt"
	"io"
	"strings"
	"time"
)

type Interval int

const (
	All Interval = iota
	Day
	Week
	Month
	Year
)

func ParseInterval(s string) (Interval, error) {
	switch strings.ToLower(s) {
	case "":
		return All, nil
	case "day":
		return Day, nil
	case "week":
		return Week, nil
	case "month":
		return Month, nil
	case "year":
		return Year, nil
	default:
		return All, fmt.Errorf("bad interval: %s", s)
	}
}

// NewIntervalIterator returns a new IntervalIterator that returns all
// TimeRanges of the given interval between from and to. The last TimeRange
// includes the to time.
func NewIntervalIterator(from, to time.Time, interval Interval) *IntervalIterator {
	return &IntervalIterator{from: from, to: to, interval: interval}
}

// IntervalIterator iterates over time interval ranges.
type IntervalIterator struct {
	from     time.Time
	to       time.Time
	eof      bool
	interval Interval
}

// Next returns the next time range, or an io.EOF error after all ranges have
// been returned.
func (i *IntervalIterator) Next() (TimeRange, error) {
	if i.eof {
		return TimeRange{}, io.EOF
	} else if !i.from.Before(i.to) {
		i.eof = true
	}
	var (
		zone = time.FixedZone(i.from.Zone())
		from = time.Date(i.from.Year(), i.from.Month(), i.from.Day(), 0, 0, 0, 0, zone)
		to   time.Time
	)
	switch i.interval {
	case Day:
		to = from.AddDate(0, 0, 1)
	case Week:
		to = from.AddDate(0, 0, 7)
	case Month:
		to = from.AddDate(0, 1, 0)
	case Year:
		to = from.AddDate(1, 0, 0)
	default:
		panic("unreachable")
	}
	i.from = to
	return TimeRange{From: from, To: to.Add(-time.Nanosecond)}, nil
}

// TimeRange holds a time range.
type TimeRange struct {
	From time.Time
	To   time.Time
}
