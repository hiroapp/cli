package db

import (
	"errors"
	"io"
	"time"
)

func EntryIterator(entries []*Entry) Iterator {
	i := entryIterator(entries)
	return &i
}

type entryIterator []*Entry

func (m *entryIterator) Next() (*Entry, error) {
	if len(*m) == 0 {
		return nil, io.EOF
	}
	e := (*m)[0]
	*m = (*m)[1:]
	return e, nil
}
func (m entryIterator) Close() error { return nil }

func IteratorEntries(itr Iterator) ([]*Entry, error) {
	var entries []*Entry
	defer itr.Close()
	for {
		if entry, err := itr.Next(); err == io.EOF {
			return entries, nil
		} else if err != nil {
			return entries, err
		} else {
			entries = append(entries, entry)
		}
	}
}

type Entry struct {
	ID       string
	Category []string
	Start    time.Time
	End      time.Time
	Note     string
}

func (e Entry) Valid() error {
	if e.Start.IsZero() {
		return errors.New("start is required")
	} else if !e.End.IsZero() && !e.End.After(e.Start) {
		return errors.New("end must be after start")
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

func (e *Entry) Empty() bool {
	return e.Equal(&Entry{})
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
