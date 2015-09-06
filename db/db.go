package db

import (
	"database/sql"
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
	// Save normalizes, validates and saves the given entry or returns an error.
	Save(*Entry) error
	// Query returns an Iterator that lists all entries matched by the given
	// query, or an error. Callers are required to call Close() on the iterator.
	Query(Query) (Iterator, error)
	// Remove deletes the entry with the given id from the db or returns an
	// error.
	Remove(string) error
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
	// Category returns entries with the given category name.
	Category []string
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

// Save is part of the db interface.
func (d *db) Save(e *Entry) error {
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
	var cID interface{}
	if len(e.Category) > 0 {
		cID, err = d.categoryID(e.Category, true)
		if err != nil {
			return err
		}
	}
	q := "INSERT INTO entries (id, start, end, note, category_id) VALUES (?, ?, ?, ?, ?)"
	args := []interface{}{e.ID, start, end, e.Note, cID}
	if !insert {
		q = "UPDATE entries SET id=?, start=?, end=?, note=?, category_id=? WHERE id=?"
		args = append(args, e.ID)
	}
	_, err = d.Exec(q, args...)
	return err
}

// categoryID returns the id of the given category, or an error.
func (d *db) categoryID(category []string, create bool) (string, error) {
	var parentID sql.NullString
	for _, name := range category {
		q := "SELECT id FROM categories WHERE name = ? AND parent_id "
		args := []interface{}{name}
		if parentID.Valid {
			q += "= ?"
			args = append(args, parentID)
		} else {
			q += "IS NULL"
		}
		row := d.QueryRow(q, args...)
		if err := row.Scan(&parentID); err == sql.ErrNoRows {
			if !create {
				return "", err
			}
			id := uuid.NewRandom().String()
			if _, err := d.Exec("INSERT INTO categories (id, name, parent_id) VALUES (?, ?, ?)", id, name, parentID); err != nil {
				return "", err
			}
			parentID.String = id
			parentID.Valid = true
		} else if err != nil {
			return "", err
		}
	}
	return parentID.String, nil

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
	if len(q.Category) != 0 {
		// @TODO figure out if this can be turned into a sub-query.
		categoryID, err := d.categoryID(q.Category, false)
		if err == sql.ErrNoRows {
			return EntryIterator(nil), nil
		} else if err != nil {
			return nil, err
		}
		where = append(where, "category_id = ?")
		args = append(args, categoryID)
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
	for dst, val := range map[*time.Time]sql.NullString{&entry.Start: start, &entry.End: end} {
		if !val.Valid {
			continue
		} else if t, err := time.Parse(datetimeLayout, val.String); err != nil {
			return nil, err
		} else {
			*dst = t
		}
	}
	if categoryID.Valid {
		category, err := i.category(categoryID.String)
		if err != nil {
			return nil, err
		}
		entry.Category = category
	}
	return &entry, i.rows.Err()
}

func (i *iterator) category(id string) ([]string, error) {
	rows, err := i.db.Query(`
WITH RECURSIVE
	category_ids(category_id) AS (
		VALUES(?)
		UNION ALL
		SELECT parent_id
			FROM categories, category_ids
			WHERE id=category_id
	)
SELECT categories.name
	FROM categories, category_ids
	WHERE category_ids.category_id = categories.id;
`, id)
	if err != nil {
		return nil, err
	}
	names := []string{}
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		names = append([]string{name}, names...)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	} else if err := rows.Close(); err != nil {
		return nil, err
	} else {
		return names, nil
	}
}

func (i *iterator) Close() error {
	return i.rows.Close()
}

func (d *db) Remove(id string) error {
	_, err := d.Exec("DELETE FROM entries WHERE id=?", id)
	return err
}
