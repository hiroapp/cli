package db

import (
	"database/sql"
	"fmt"
	"io"
	"os"
	"time"

	"code.google.com/p/go-uuid/uuid"

	"github.com/hashicorp/go-multierror"
	_ "github.com/mattn/go-sqlite3"

	_ "github.com/mattn/go-sqlite3"
)
import "path/filepath"

const sqliteDate = "2006-01-02 15:04:05"

// DB defines the hiro database api.
type DB interface {
	Save(*Entry) error
	Query(Query) (Iterator, error)
}

type Query struct {
	Start time.Time
	// Groups [][]string
	// Asc bool
}

type Iterator interface {
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

CREATE TABLE IF NOT EXISTS groups (
	id TEXT PRIMARY KEY,
	name TEXT,
	parent_id TEXT REFERENCES groups
);
CREATE INDEX IF NOT EXISTS parent_id ON groups(parent_id);

CREATE TABLE IF NOT EXISTS entries (
	id TEXT PRIMARY KEY,
	start TEXT,
	end TEXT,
	note TEXT,
	group_id TEXT REFERENCES groups
);
CREATE INDEX IF NOT EXISTS group_id ON entries(group_id);
`)
	return err
}

// Save is part of the db interface.
func (d *db) Save(e *Entry) error {
	e.Start = e.Start.UTC().Truncate(time.Second)
	if err := e.Valid(); err != nil {
		return err
	}
	if e.ID == "" {
		e.ID = uuid.NewRandom().String()
	}
	var start, end interface{}
	if !e.Start.IsZero() {
		start = e.Start.Format(sqliteDate)
	}
	if !e.End.IsZero() {
		end = e.End.Format(sqliteDate)
	}
	tx, err := d.Begin()
	if err != nil {
		return err
	}
	groupID, err := groupID(tx, e.Group)
	if err != nil {
		return multierror.Append(err, tx.Rollback())
	}
	if _, err := tx.Exec(
		"INSERT INTO entries (id, start, end, note, group_id) VALUES (?, ?, ?, ?, ?)",
		e.ID,
		start,
		end,
		e.Note,
		groupID,
	); err != nil {
		return multierror.Append(err, tx.Rollback())
	}
	return tx.Commit()
}

func groupID(tx *sql.Tx, names []string) (string, error) {
	var parentID sql.NullString
	for _, name := range names {
		q := "SELECT id FROM groups WHERE name = ? AND parent_id "
		args := []interface{}{name}
		if parentID.Valid {
			q += "= ?"
			args = append(args, parentID)
		} else {
			q += "IS NULL"
		}
		row := tx.QueryRow(q, args...)
		if err := row.Scan(&parentID); err == sql.ErrNoRows {
			id := uuid.NewRandom().String()
			fmt.Printf("insert\n")
			if _, err := tx.Exec("INSERT INTO groups (id, name, parent_id) VALUES (?, ?, ?)", id, name, parentID); err != nil {
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

func (d *db) Query(q Query) (Iterator, error) {
	rows, err := d.DB.Query("SELECT id, start, end, group_id FROM entries ORDER BY start DESC;")
	return &iterator{db: d.DB, rows: rows}, err
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
		entry   Entry
		start   sql.NullString
		end     sql.NullString
		groupID string
	)
	if err := i.rows.Scan(&entry.ID, &start, &end, &groupID); err != nil {
		return nil, err
	}
	for dst, val := range map[*time.Time]sql.NullString{&entry.Start: start, &entry.End: end} {
		if !val.Valid {
			continue
		} else if t, err := time.Parse(sqliteDate, val.String); err != nil {
			return nil, err
		} else {
			*dst = t
		}
	}
	group, err := i.group(groupID)
	if err != nil {
		return nil, err
	}
	entry.Group = group
	return &entry, i.rows.Err()
}

func (i *iterator) group(id string) ([]string, error) {
	rows, err := i.db.Query(`
WITH RECURSIVE
	group_ids(group_id) AS (
		VALUES(?)
		UNION ALL
		SELECT parent_id
			FROM groups, group_ids
			WHERE id=group_id
	)
SELECT groups.name
	FROM groups, group_ids
	WHERE group_ids.group_id = groups.id;
`, id)
	if err != nil {
		return nil, err
	}
	names := []string{}
	fmt.Printf("%s\n", id)
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
	return i.Close()
}
