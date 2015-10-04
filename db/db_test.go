package db

import (
	"database/sql"
	"testing"
	"time"

	"github.com/kylelemons/godebug/pretty"

	"code.google.com/p/go-uuid/uuid"
)

// - Transactions
// - Start and End are truncated.
// - Record is validates
// - ID is assigned on insert, kept on update
// - Category id is resolved or created (if set)
// - uuid assignment

// - update

// TestSaveEntry performs save operations and validates their results as well as the
// resulting db state using Query.
func TestSaveEntry(t *testing.T) {
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
		if err := d.SaveEntry(fixture); err != nil {
			t.Errorf("test %s: %s", test.Name, err)
			continue
		}
		if test.Update {
			test.Entry.ID = fixture.ID
		}
		var (
			err    = d.SaveEntry(test.Entry)
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

func TestGetOrCreateCategoryPath(t *testing.T) {
	db := mustDB(t)
	path, err := db.CategoryPath([]string{"a", "b", "c"}, true)
	if err != nil {
		t.Fatal(err)
	}
	wantPath := CategoryPath{
		&Category{ID: path[0].ID, Name: "a"},
		&Category{ID: path[1].ID, Name: "b", ParentID: path[0].ID},
		&Category{ID: path[2].ID, Name: "c", ParentID: path[1].ID},
	}
	if diff := pretty.Compare(path, wantPath); diff != "" {
		t.Fatal(diff)
	} else if path2, err := db.CategoryPath([]string{"a", "b", "c"}, true); err != nil {
		t.Fatal(err)
	} else if diff := pretty.Compare(path2, wantPath); diff != "" {
		t.Fatal(diff)
	}
	path3, err := db.CategoryPath([]string{"a", "d"}, true)
	if err != nil {
		t.Fatal(err)
	}
	wantPath = CategoryPath{
		&Category{ID: path[0].ID, Name: "a"},
		&Category{ID: path3[1].ID, Name: "d", ParentID: path[0].ID},
	}
	if diff := pretty.Compare(path3, wantPath); diff != "" {
		t.Fatal(diff)
	}
	categories, err := db.Categories()
	if err != nil {
		t.Fatal(err)
	}
	wantCategories := CategoryMap{
		path[0].ID:  path[0],
		path[1].ID:  path[1],
		path[2].ID:  path[2],
		path3[1].ID: path3[1],
	}
	if diff := pretty.Compare(categories, wantCategories); diff != "" {
		t.Fatal(diff)
	}
	// @TODO check actual error
	if _, err := db.CategoryPath([]string{"a", "e"}, false); err == nil {
		t.Fatal("expected does not exist error")
	}
}

func TestCategories(t *testing.T) {
	db := mustDB(t)
	a := &Category{Name: "a"}
	if err := db.SaveCategory(a); err != nil {
		t.Fatal(err)
	}
	b := &Category{Name: "b", ParentID: a.ID}
	if err := db.SaveCategory(b); err != nil {
		t.Fatal(err)
	}
	want := CategoryMap{a.ID: a, b.ID: b}
	if got, err := db.Categories(); err != nil {
		t.Error(err)
	} else if diff := pretty.Compare(got, want); diff != "" {
		t.Fatal(diff)
	}
	a.Name, b.Name = "c", "d"
	if err := db.SaveCategory(a); err != nil {
		t.Fatal(err)
	} else if err := db.SaveCategory(b); err != nil {
		t.Fatal(err)
	} else if got, err := db.Categories(); err != nil {
		t.Error(err)
	} else if diff := pretty.Compare(got, want); diff != "" {
		t.Fatal(diff)
	}
}

func TestCategoryMap_Root(t *testing.T) {
	tests := []struct {
		Name string
		Map  CategoryMap
		Root *CategoryNode
	}{
		{
			Name: "empty map",
			Map:  nil,
			Root: &CategoryNode{},
		},
		{
			Name: "flat",
			Map: CategoryMap{
				"1": &Category{ID: "1", Name: "a"},
				"2": &Category{ID: "2", Name: "b"},
				"3": &Category{ID: "3", Name: "c"},
			},
			Root: &CategoryNode{
				Children: []*CategoryNode{
					{Category: &Category{ID: "1", Name: "a"}},
					{Category: &Category{ID: "2", Name: "b"}},
					{Category: &Category{ID: "3", Name: "c"}},
				},
			},
		},
		{
			Name: "nested",
			Map: CategoryMap{
				"1": &Category{ID: "1", Name: "a"},
				"2": &Category{ID: "2", Name: "b", ParentID: "1"},
				"3": &Category{ID: "3", Name: "c", ParentID: "2"},
				"4": &Category{ID: "4", Name: "d", ParentID: "1"},
				"5": &Category{ID: "5", Name: "e"},
			},
			Root: &CategoryNode{
				Children: []*CategoryNode{
					{
						Category: &Category{ID: "1", Name: "a"},
						Children: []*CategoryNode{
							{
								Category: &Category{ID: "2", Name: "b", ParentID: "1"},
								Children: []*CategoryNode{
									{
										Category: &Category{ID: "3", Name: "c", ParentID: "2"},
									},
								},
							},
							{
								Category: &Category{ID: "4", Name: "d", ParentID: "1"},
							},
						},
					},
					{Category: &Category{ID: "5", Name: "e"}},
				},
			},
		},
	}
	for _, test := range tests {
		got := test.Map.Root()
		if diff := pretty.Compare(got, test.Root); diff != "" {
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
