package db

import (
	"testing"
	"time"
)

func TestEntry_PartialOverlap(t *testing.T) {
	now := time.Now()
	tests := []struct {
		Entry *Entry
		Now   time.Time
		From  time.Time
		To    time.Time
		Want  time.Duration
	}{
		{
			Entry: &Entry{Start: now, End: now.Add(10 * time.Second)},
			From:  now,
			To:    now.Add(10 * time.Second),
			Want:  10 * time.Second,
		},
		{
			Entry: &Entry{Start: now, End: now.Add(10 * time.Second)},
			From:  now.Add(time.Second),
			To:    now.Add(15 * time.Second),
			Want:  9 * time.Second,
		},
		{
			Entry: &Entry{Start: now, End: now.Add(10 * time.Second)},
			From:  now.Add(-3 * time.Second),
			To:    now.Add(8 * time.Second),
			Want:  8 * time.Second,
		},
		{
			Entry: &Entry{Start: now, End: now.Add(10 * time.Second)},
			From:  now.Add(1 * time.Second),
			To:    now.Add(8 * time.Second),
			Want:  7 * time.Second,
		},
		{
			Entry: &Entry{Start: now, End: now.Add(10 * time.Second)},
			From:  now.Add(10 * time.Second),
			To:    now.Add(20 * time.Second),
			Want:  0,
		},
		{
			Entry: &Entry{Start: now},
			From:  now,
			To:    now.Add(10 * time.Second),
			Now:   now.Add(10 * time.Second),
			Want:  10 * time.Second,
		},
		{
			Entry: &Entry{Start: now},
			From:  now.Add(time.Second),
			To:    now.Add(12 * time.Second),
			Now:   now.Add(10 * time.Second),
			Want:  9 * time.Second,
		},
	}
	for _, test := range tests {
		got := test.Entry.PartialDuration(test.Now, test.From, test.To)
		if got != test.Want {
			t.Errorf("got=%s want=%s\n", got, test.Want)
		}
	}
}
