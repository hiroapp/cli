package db

import (
	"database/sql"
	"testing"
	"time"

	"github.com/felixge/godebug/pretty"

	"code.google.com/p/go-uuid/uuid"
)

// - Transactions
// - Start and End are truncated.
// - Record is validates
// - ID is assigned on insert, kept on update
// - Category id is resolved or created (if set)
// - uuid assignment

// - update

// TestSave performs save operations and validates their results as well as the
// resulting db state using Query.
func TestSave(t *testing.T) {
	diffConfig := &pretty.Config{Diffable: true, PrintStringers: true}
	zone := time.FixedZone("", 3600)
	start := time.Date(2015, 9, 2, 15, 36, 13, 123, zone)
	tests := []struct {
		Name    string
		Entry   *Entry
		WantErr string
		Update  bool
	}{
		{
			Name:    "validation is performed",
			Entry:   &Entry{},
			WantErr: "start is required",
		},
		{
			Name:  "minimal valid entry",
			Entry: &Entry{Start: start},
		},
		{
			Name:  "start and end",
			Entry: &Entry{Start: start, End: start.Add(time.Second)},
		},
		{
			Name:   "update",
			Entry:  &Entry{Start: start, End: start.Add(time.Second)},
			Update: true,
		},
	}
	for _, test := range tests {
		var (
			fixture = &Entry{Start: start.Add(-time.Second), End: start}
			d       = mustDB(t)
		)
		if err := d.Save(fixture); err != nil {
			t.Errorf("test %s: %s", test.Name, err)
			continue
		}
		if test.Update {
			test.Entry.ID = fixture.ID
		}
		var (
			err    = d.Save(test.Entry)
			gotErr string
		)
		if err != nil {
			gotErr = err.Error()
		}
		if gotErr != test.WantErr {
			t.Errorf("test %q: got=%q want=%q", test.Name, gotErr, test.WantErr)
		} else if test.WantErr != "" {
			continue
		}
		if uuid.Parse(test.Entry.ID) == nil {
			t.Errorf("test %q: got=%q want uuid", test.Name, test.Entry.ID)
		}
		if got, want := test.Entry.Start, test.Entry.Start.Truncate(time.Second); !got.Equal(want) {
			t.Errorf("test %q: got=%q want=%q", test.Name, got, want)
		}
		if got, want := test.Entry.End, test.Entry.End.Truncate(time.Second); !got.Equal(want) {
			t.Errorf("test %q: got=%q want=%q", test.Name, got, want)
		}
		wantEntries := []*Entry{test.Entry}
		if !test.Update {
			wantEntries = append(wantEntries, fixture)
		}
		if itr, err := d.Query(Query{}); err != nil {
			t.Errorf("test %q: %s", test.Name, err)
		} else if entries, err := IteratorEntries(itr); err != nil {
			t.Errorf("test %q: %s", test.Name, err)
		} else if diff := diffConfig.Compare(entries, wantEntries); diff != "" {
			t.Errorf("test %q: %s", test.Name, diff)
		}
	}
}

func mustDB(t *testing.T) DB {
	sqlLite, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	db := &db{DB: sqlLite}
	if err := db.init(); err != nil {
		t.Fatal(err)
	}
	return db
}
