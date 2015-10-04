package main

import (
	"time"

	"github.com/hiroapp/cli/datetime"
	"github.com/hiroapp/cli/db"
)

// NewSummaryIterator returns a new summary iterator producing summaries for
// the given period and firstDay of the week. If the period is datetime.Day,
// it is is ignored. Callers are required to call Close once they are done
// with the iterator.
func NewSummaryIterator(d db.DB, period datetime.Duration, firstDay time.Weekday, now time.Time) (*SummaryIterator, error) {
	entries, err := d.Query(db.Query{})
	if err != nil {
		return nil, err
	}
	return &SummaryIterator{
		now:      now,
		entries:  entries,
		period:   period,
		firstDay: firstDay,
	}, nil
}

// SummaryIterator implements summary iteration.
type SummaryIterator struct {
	now      time.Time
	entries  db.Iterator
	entry    *db.Entry
	periods  *datetime.Iterator
	period   datetime.Duration
	firstDay time.Weekday
}

// Next returns the next summary or an error. When there are no more summaries,
// the error io.EOF is returned.
func (s *SummaryIterator) Next() (*Summary, error) {
	// Fetch the first entry on first call to Next to create periods iterator.
	// Doing it here rather than in the constructor to avoid callers having to
	// implement additional logic when the db is empty.
	var err error
	if s.entry == nil {
		if s.entry, err = s.entries.Next(); err != nil {
			return nil, err
		}
		s.periods = datetime.NewIterator(s.entry.Start, s.period, false, s.firstDay)
	}

	summary := &Summary{Categories: make(map[string]time.Duration)}
	summary.From, summary.To = s.periods.Next()
	for {
		duration := s.entry.PartialDuration(s.now, summary.From, summary.To)
		if duration > 0 {
			summary.Categories[s.entry.CategoryID] += duration
		}
		if s.entry.Start.Before(summary.From) {
			break
		}
		if s.entry, err = s.entries.Next(); err != nil {
			return nil, err
		}
	}
	return summary, nil
}

// Close closes the iterator.
func (s *SummaryIterator) Close() error {
	return s.entries.Close()
}

// Summary stores how much time was spend in which category for a given time
// range.
type Summary struct {
	From       time.Time
	To         time.Time
	Categories map[string]time.Duration
}
