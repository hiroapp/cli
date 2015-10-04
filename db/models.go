package db

import (
	"errors"
	"io"
	"time"

	"github.com/bradfitz/slice"
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
	ID         string
	CategoryID string
	Start      time.Time
	End        time.Time
	Note       string
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

// PartialDuration returns the duration of the entry that
// overlaps with the given from and to time.
func (e Entry) PartialDuration(now, from, to time.Time) time.Duration {
	if from.Before(e.Start) {
		from = e.Start
	}
	end := e.End
	if end.IsZero() {
		end = now
	}
	if to.After(end) {
		to = end
	}
	if from.After(to) {
		return 0
	}
	return to.Sub(from)
}

// @TODO can this be removed?
func (e *Entry) Equal(o *Entry) bool {
	return e == o ||
		(e.ID == o.ID &&
			e.Start.Equal(o.Start) &&
			e.End.Equal(o.End) &&
			e.Note == o.Note &&
			e.CategoryID == o.CategoryID)
}

// @TODO can this be removed?
func (e *Entry) Empty() bool {
	return e.Equal(&Entry{})
}

type Category struct {
	ID       string
	Name     string
	ParentID string
}

// CategoryNode is a node in a category tree.
type CategoryNode struct {
	// Category is the category for the current node, or nil for the root node.
	*Category
	// Children holds the child nodes of the current category, ordered by name.
	Children []*CategoryNode
}

// ChildrenByName returns the child nodes matching the given name, if any.
func (c *CategoryNode) ChildrenByName(name string) []*CategoryNode {
	if c == nil {
		return nil
	}
	var nodes []*CategoryNode
	for _, node := range c.Children {
		if node.Name == name {
			nodes = append(nodes, node)
		}
	}
	return nodes
}

// CategoryPath holds a path of the category tree.
type CategoryPath []*Category

// CategoryID returns the id of the last category in the path or an empty
// string if the path is empty.
func (c CategoryPath) CategoryID() string {
	if len(c) == 0 {
		return ""
	}
	return c[len(c)-1].ID
}

// CategoryMap is a set of categories indexed by id.
type CategoryMap map[string]*Category

// Tree returns the root node of the category tree.
func (c CategoryMap) Root() *CategoryNode {
	index := map[string]*CategoryNode{"": &CategoryNode{}}
	for _, category := range c {
		parent := index[category.ParentID]
		if parent == nil {
			parent = &CategoryNode{}
			index[category.ParentID] = parent
		}
		node := index[category.ID]
		if node == nil {
			node = &CategoryNode{}
			index[category.ID] = node
		}
		node.Category = category
		parent.Children = append(parent.Children, node)
		slice.Sort(parent.Children, func(i, j int) bool {
			return parent.Children[i].Name < parent.Children[j].Name
		})
	}
	return index[""]
}

// Path returns the path for the given category, or nil if it can't be
// resolved.
func (c CategoryMap) Path(id string) CategoryPath {
	var path CategoryPath
	for id != "" {
		if category := c[id]; category == nil {
			return nil
		} else {
			path = append([]*Category{category}, path...)
			id = category.ParentID
		}
	}
	return path
}
