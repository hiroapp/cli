// Package table provides a simple interface for rendering ascii tables.
package table

import (
	"bytes"
	"strings"
)

// New returns a new table.
func New() *Table {
	return &Table{}
}

// Table holds tabular data and formatting options.
type Table struct {
	padding string
	rows    [][]*Cell
}

// Padding sets the padding s for the Table and returns it.
func (t *Table) Padding(s string) *Table {
	t.padding = s
	return t
}

// Add adds the given row of cells to the table and returns it.
func (t *Table) Add(cells ...*Cell) *Table {
	t.rows = append(t.rows, cells)
	return t
}

// Strings renders the table and returns it as a string.
func (t *Table) String() string {
	var (
		widths = t.widths()
		buf    = &bytes.Buffer{}
	)
	for _, row := range t.rows {
		for i, cell := range row {
			s := cell.s
			padding := strings.Repeat(" ", widths[i]-len(s))
			last := i == len(row)-1
			if cell.align == Right {
				s = padding + s
			} else if !last {
				s += padding
			}
			if !last {
				s += " "
			}
			buf.WriteString(s)
		}
		buf.WriteString("\n")
	}
	return buf.String()
}

func (t *Table) widths() []int {
	var widths []int
	for _, row := range t.rows {
		for i, cell := range row {
			if i == len(widths) {
				widths = append(widths, 0)
			}
			if l := len(cell.s); l > widths[i] {
				widths[i] = l
			}
		}
	}
	return widths
}

// String returns a cell holding the given string value.
func String(s string) *Cell {
	return &Cell{s: s}
}

// Cell holds a table cell.
type Cell struct {
	s     string
	align Alignment
}

// Align sets the alignment of the cell and returns it.
func (c *Cell) Align(a Alignment) *Cell {
	c.align = a
	return c
}

// Alignment holds a cell alignment.
type Alignment int

const (
	Left Alignment = iota
	Right
)
