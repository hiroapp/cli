package db

import (
	"io/ioutil"
	"os"
	"testing"
	"time"
)

func TestDB(t *testing.T) {
	dir, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)
	db := New(dir)
	now := time.Now()
	entries := make([]*Entry, 5000)
	year := 24 * time.Hour * 365
	duration := (2 * year) / time.Duration(len(entries))
	for i := range entries {
		start := now.Add(-time.Duration(i+1) * duration)
		group := []string{"A"}
		if i%2 == 0 {
			group = []string{"A", "B"}
		} else if i%3 == 0 {
			group = []string{"C"}
		}
		entries[i] = &Entry{
			Group: group,
			Start: start,
			End:   start.Add(duration),
		}
		if err := db.Save(entries[i]); err != nil {
			t.Fatal(err)
		}
	}
	itr, err := db.Query(Query{Start: now.Add(-year)})
	if err != nil {
		t.Fatal(err)
	}
	for i, o := range entries {
		if entry, err := itr.Next(); err != nil {
			t.Fatalf("entry %d: %s", i, err)
		} else if !entry.Equal(o) {
			t.Fatal("entry %d: %#v != %#v", i, entry, o)
		}
	}
}
