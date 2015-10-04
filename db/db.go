package db

import (
	"database/sql"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"code.google.com/p/go-uuid/uuid"

	_ "github.com/mattn/go-sqlite3"
)
import "path/filepath"

const (
	datetimeLayout = "2006-01-02 15:04:05 -07:00"
)

// DB defines the hiro database api.
type DB interface {
	// SaveEntry normalizes, validates and saves the given entry or returns an
	// error.
	SaveEntry(*Entry) error
	// Query returns an Iterator that lists all entries matched by the given
	// query, or an error. Callers are required to call Close() on the iterator.
	Query(Query) (Iterator, error)
	// Remove deletes the entry with the given id from the db or returns an
	// error.
	Remove(string) error
	// GetOrCreateCategoryPath returns a category path with the given names,
	// creating categories as needed.
	GetOrCreateCategoryPath([]string) (CategoryPath, error)
	// categories as needed
	// SaveCategory saves the given Category.
	SaveCategory(*Category) error
	// Categories returns all categories indexed by id an error.
	Categories() (CategoryMap, error)
	// Close closes the database.
	Close() error
}

type Query struct {
	// IDs returns entries matching the given ids if set.
	IDs []string
	// Asc returns entries in ascending order if true.
	Asc bool
	// Active returns entries without an end time if true.
	Active bool
	// Category returns entries with the given category id.
	CategoryID string
}

type Iterator interface {
	// Next returns the next entry or an error. An io.EOF is returned after the
	// last entry.
	Next() (*Entry, error)
	Close() error
}

func New(dir string) (DB, error) {
	if err := os.MkdirAll(dir, 0700); err != nil {
		return nil, err
	} else if d, err := sql.Open("sqlite3", filepath.Join(dir, "hiro.db")); err != nil {
		return nil, err
	} else {
		db := &db{DB: d}
		return db, db.init()
	}
}

// db implements the DB interface.
type db struct {
	*sql.DB
}

func (d *db) init() error {
	_, err := d.Exec(`
PRAGMA foreign_keys = ON;

CREATE TABLE IF NOT EXISTS categories (
	id TEXT PRIMARY KEY,
	name TEXT,
	parent_id TEXT REFERENCES categories
);
CREATE INDEX IF NOT EXISTS parent_id ON categories(parent_id);
CREATE UNIQUE INDEX IF NOT EXISTS name_parent_id ON categories(name, parent_id);

CREATE TABLE IF NOT EXISTS entries (
	id TEXT PRIMARY KEY,
	start TEXT,
	end TEXT,
	note TEXT,
	category_id TEXT REFERENCES categories
);
CREATE INDEX IF NOT EXISTS category_id ON entries(category_id);
`)
	return err
}

// SaveEntry is part of the DB interface.
func (d *db) SaveEntry(e *Entry) error {
	e.Start = e.Start.Truncate(time.Second)
	e.End = e.End.Truncate(time.Second)
	err := e.Valid()
	if err != nil {
		return err
	}
	var insert bool
	if insert = e.ID == ""; insert {
		e.ID = uuid.NewRandom().String()
	}
	var start, end interface{}
	start = e.Start.Format(datetimeLayout)
	if !e.End.IsZero() {
		end = e.End.Format(datetimeLayout)
	}
	categoryID := sql.NullString{String: e.CategoryID, Valid: e.CategoryID != ""}
	q := "INSERT INTO entries (id, start, end, note, category_id) VALUES (?, ?, ?, ?, ?)"
	args := []interface{}{e.ID, start, end, e.Note, categoryID}
	if !insert {
		q = "UPDATE entries SET id=?, start=?, end=?, note=?, category_id=? WHERE id=?"
		args = append(args, e.ID)
	}
	_, err = d.Exec(q, args...)
	return err
}

// @TODO test
func (d *db) Query(q Query) (Iterator, error) {
	var parts = []string{"SELECT id, start, end, note, category_id", "FROM entries"}
	var (
		args  []interface{}
		where []string
	)
	if len(q.IDs) > 0 {
		where = append(where, fmt.Sprintf("id IN (?"+strings.Repeat(", ?", len(q.IDs)-1)+") "))
		for _, id := range q.IDs {
			args = append(args, id)
		}
	}
	if q.Active {
		where = append(where, "end IS NULL")
	}
	if q.CategoryID != "" {
		where = append(where, "category_id = ?")
		args = append(args, q.CategoryID)
	}
	if len(where) > 0 {
		parts = append(parts, "WHERE "+strings.Join(where, " AND "))
	}
	order := "DESC"
	if q.Asc == true {
		order = "ASC"
	}
	parts = append(parts, "ORDER BY DATETIME(start, 'utc') "+order)
	sql := strings.Join(parts, " ")
	rows, err := d.DB.Query(sql, args...)
	return &iterator{db: d.DB, rows: rows}, err
}

// GetOrCreateCategoryPath is part of the DB interface.
func (d *db) GetOrCreateCategoryPath(names []string) (CategoryPath, error) {
	categories, err := d.Categories()
	if err != nil {
		return nil, err
	}
	node := categories.Root()
	path := make(CategoryPath, 0, len(names))
	var categoryID string
	for _, name := range names {
		var category *Category
		nodes := node.ChildrenByName(name)
		if l := len(nodes); l > 1 {
			return nil, errors.New("category exists more than once")
		} else if l == 1 {
			node = nodes[0]
			category = node.Category
		} else {
			node = nil
			category = &Category{Name: name, ParentID: categoryID}
			if err := d.SaveCategory(category); err != nil {
				return nil, err
			}
		}
		path = append(path, category)
		categoryID = category.ID
	}
	return path, nil
}

// SaveCategory is part of the DB interface.
func (d *db) SaveCategory(c *Category) error {
	var insert bool
	if insert = c.ID == ""; insert {
		c.ID = uuid.NewRandom().String()
	}
	parentID := sql.NullString{String: c.ParentID, Valid: c.ParentID != ""}
	q := "INSERT INTO categories (id, name, parent_id) VALUES (?, ?, ?)"
	args := []interface{}{c.ID, c.Name, parentID}
	if !insert {
		q = "UPDATE categories SET id=?, name=?, parent_id=? WHERE id=?"
		args = append(args, c.ID)
	}
	_, err := d.Exec(q, args...)
	return err

}

// Categories is part of the DB interface.
func (d *db) Categories() (CategoryMap, error) {
	rows, err := d.DB.Query("SELECT id, name, parent_id FROM categories")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	m := make(map[string]*Category)
	for rows.Next() {
		category := &Category{}
		var parentID sql.NullString
		if err := rows.Scan(&category.ID, &category.Name, &parentID); err != nil {
			return nil, err
		}
		category.ParentID = parentID.String
		m[category.ID] = category
	}
	return m, nil
}

func (d *db) Close() error {
	return d.DB.Close()
}

type iterator struct {
	db   *sql.DB
	rows *sql.Rows
}

func (i *iterator) Next() (*Entry, error) {
	if !i.rows.Next() {
		return nil, io.EOF
	}
	var (
		entry      Entry
		start      sql.NullString
		end        sql.NullString
		categoryID sql.NullString
	)
	if err := i.rows.Scan(&entry.ID, &start, &end, &entry.Note, &categoryID); err != nil {
		return nil, err
	}
	entry.CategoryID = categoryID.String
	for dst, val := range map[*time.Time]sql.NullString{&entry.Start: start, &entry.End: end} {
		if !val.Valid {
			continue
		} else if t, err := time.Parse(datetimeLayout, val.String); err != nil {
			return nil, err
		} else {
			*dst = t
		}
	}
	return &entry, i.rows.Err()
}

func (i *iterator) Close() error {
	return i.rows.Close()
}

func (d *db) Remove(id string) error {
	_, err := d.Exec("DELETE FROM entries WHERE id=?", id)
	return err
}
