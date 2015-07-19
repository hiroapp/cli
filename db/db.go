package db

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"regexp"
	"sort"
	"time"
)
import "path/filepath"

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
}

func New(dir string) DB {
	return &db{dir: dir}
}

// db implements the DB interface.
type db struct {
	dir string
}

// Save is part of the db interface.
func (d *db) Save(e *Entry) error {
	if err := e.Valid(); err != nil {
		return err
	}
	e.Start = e.Start.UTC().Truncate(time.Second)
	dir := filepath.Join(append(
		append([]string{d.dir}, e.Group...),
		e.Start.Format("Y2006"),
		e.Start.Format("M01"),
		e.Start.Format("D02"),
	)...)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}
	name := e.Start.UTC().Format("15-04-05.json")
	path := filepath.Join(dir, name)
	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	defer file.Close()
	return json.NewEncoder(file).Encode(dbEntry{End: e.End, Note: e.Note})
}

type dbEntry struct {
	End  time.Time
	Note string
}

func (d *db) Query(q Query) (Iterator, error) {
	return &iterator{dir: d.dir, offset: q.Start}, nil
}

type iterator struct {
	dir    string
	groups map[*iteratorGroup]*Entry
	offset time.Time
}

type iteratorGroup struct {
	dir    string
	offset time.Time
	name   []string
	years  []string
	months []string
	days   []string
	files  []string
}

func (g *iteratorGroup) Next() (*Entry, error) {
	for len(g.years) > 0 {
		year := g.years[0]
		if year > g.offset.Format("2006") {
			g.years = g.years[1:]
			continue
		}
		if len(g.months) == 0 {
			infos, err := ioutil.ReadDir(filepath.Join(g.dir, "Y"+year))
			if err != nil {
				return nil, err
			}
			g.months = make([]string, 0, len(infos))
			for _, info := range infos {
				if !info.IsDir() {
					continue
				} else if name := info.Name(); !monthDir.MatchString(name) {
					continue
				} else if month := name[1:]; month > g.offset.Format("01") {
					continue
				} else {
					g.months = append(g.months, month)
				}
			}
			if len(g.months) == 0 {
				g.years = g.years[1:]
				continue
			}
			sort.Sort(sort.Reverse(sort.StringSlice(g.months)))
		}
		month := g.months[0]
		fmt.Printf("%s - %s\n", year, month)
		break
	}
	return nil, io.EOF
}

func (itr *iterator) Next() (*Entry, error) {
	if itr.groups == nil {
		itr.groups = make(map[*iteratorGroup]*Entry)
		if err := itr.open(itr.dir, nil); err != nil {
			return nil, err
		}
	}
	var (
		bestEntry *Entry
		bestGroup *iteratorGroup
		err       error
	)
	for group, entry := range itr.groups {
		if entry == nil {
			if entry, err = group.Next(); err == io.EOF {
				delete(itr.groups, group)
				continue
			} else if err != nil {
				return nil, err
			}
			itr.groups[group] = entry
		}
		if bestEntry == nil || entry.Start.Before(bestEntry.Start) {
			bestEntry = entry
			bestGroup = group
		}
	}
	if bestEntry != nil {
		itr.groups[bestGroup] = nil
		return bestEntry, nil
	}
	return nil, io.EOF
}

var (
	yearDir  = regexp.MustCompile("^Y\\d{4}$")
	monthDir = regexp.MustCompile("^M\\d{2}$")
	dayDir   = regexp.MustCompile("^D\\d{2}$")
)

func (itr *iterator) open(dir string, groupName []string) error {
	file, err := os.Open(dir)
	if err != nil {
		return err
	}
	defer file.Close()
	if children, err := file.Readdir(0); err != nil {
		return err
	} else {
		group := &iteratorGroup{dir: dir, name: groupName, offset: itr.offset}
		if len(groupName) > 0 {
			itr.groups[group] = nil
		}
		for _, child := range children {
			if !child.IsDir() {
				continue
			} else if childName := child.Name(); yearDir.MatchString(childName) {
				group.years = append(group.years, childName[1:])
			} else {
				itr.open(filepath.Join(dir, child.Name()), append(groupName, childName))
			}
		}
		sort.Sort(sort.Reverse(sort.StringSlice(group.years)))
	}
	return nil
}
