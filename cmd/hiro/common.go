package main

import (
	"bytes"
	"fmt"
	"io"
	"regexp"
	"strings"
	"text/template"
	"time"

	"github.com/hiroapp/cli/datetime"
	"github.com/hiroapp/cli/db"
	"github.com/hiroapp/cli/table"
)

const categorySeparator = ":"

var tmpl = template.Must(template.New("entry").Funcs(template.FuncMap{
	"now":  func() time.Time { return time.Now().Truncate(time.Second) },
	"join": strings.Join,
	"format": func(t time.Time) string {
		return t.Format(timeLayout)
	},
}).Parse(strings.TrimSpace(`
Id:       {{.Entry.ID}}
Category: {{.Category}}
Start:    {{format .Entry.Start}}
{{if not .HideEnd}}End:      {{if .Entry.End.IsZero}}{{else}}{{format .Entry.End}}{{end}}
{{end}}{{if not .HideDuration}}Duration: {{.Entry.Duration now}}
{{end}}
{{if .Entry.Note}}{{.Entry.Note}}
{{end}}
`)))

func FprintEntry(w io.Writer, e *db.Entry, path db.CategoryPath, m PrintMask) error {
	return tmpl.Execute(w, map[string]interface{}{
		"Entry":        e,
		"HideDuration": m&PrintHideDuration > 0,
		"HideEnd":      m&PrintHideEnd > 0,
		"Category":     FormatCategory(path),
	})
}

func FprintIterator(w io.Writer, itr db.Iterator, categories db.CategoryMap, m PrintMask) error {
	for first := true; ; first = false {
		if entry, err := itr.Next(); err == io.EOF {
			return nil
		} else if err != nil {
			return err
		} else {
			if !first {
				separator := "\n"
				if m&PrintSeparator > 0 {
					separator += entrySeparator + "\n\n"
				}
				if _, err := fmt.Fprintf(w, "%s", separator); err != nil {
					return err
				}
			}
			if err := FprintEntry(w, entry, categories.Path(entry.CategoryID), m); err != nil {
				return err
			}
		}
	}
}

func FormatCategory(path db.CategoryPath) string {
	names := make([]string, len(path))
	for i, category := range path {
		names[i] = category.Name
	}
	return strings.Join(names, categorySeparator)
}

type PrintMask int

const (
	PrintDefault      PrintMask = 0
	PrintHideDuration PrintMask = 1 << (iota - 1)
	PrintHideEnd
	PrintSeparator
)

var entryField = regexp.MustCompile("^([^:]+):\\s*(.*?)\\s*$")

const (
	timeLayout     = "2006-01-02 15:04:05 -0700"
	entrySeparator = "8< ----- do not remove this separator ----- >8"
)

func Indent(s, indent string) string {
	lines := strings.Split(s, "\n")
	for i, line := range lines {
		lines[i] = indent + line
	}
	return strings.Join(lines, "\n")
}

// FormatDuration returns the duration as a H:MM:SS formated string, e.g.
// "1:03:03" for 1h2m3s or "123:45:56" for 123h45m56s.
func FormatDuration(d time.Duration) string {
	hours := d / time.Hour
	d -= hours * time.Hour
	minutes := d / time.Minute
	d -= minutes * time.Minute
	seconds := d / time.Second
	return fmt.Sprintf("%d:%02d:%02d", hours, minutes, seconds)
}

func PeriodHeadline(from, to time.Time, period datetime.Period) string {
	switch period {
	case datetime.Day:
		return fmt.Sprintf("%s", from.Format("2006-01-02: Monday"))
	case datetime.Week:
		_, isoWeek := from.ISOWeek()
		return fmt.Sprintf("Week %d: %s - %s", isoWeek, from.Format("2006-01-02"), to.Format("2006-01-02"))
	case datetime.Month:
		return fmt.Sprintf("%s", from.Format("January 2006"))
	case datetime.Year:
		return fmt.Sprintf("%s", from.Format("Year 2006"))
	default:
		panic("not implemeneted")
	}
}

// ParseCategory splits a colon separated category identifier into a string
// slice containing the individual parts, e.g. "Foo:Bar:Baz" into "Foo", "Bar",
// "Baz". The empty category string maps to nil.
func ParseCategory(category string) []string {
	if category == "" {
		return nil
	}
	return strings.Split(category, categorySeparator)
}

func FormatReport(r *Report) string {
	if r == nil {
		return ""
	}
	buf := &bytes.Buffer{}
	buf.WriteString(PeriodHeadline(r.From, r.To, r.Duration))
	buf.WriteString("\n\n")
	t := table.New()
	t.Add(
		table.String("DATE"),
		table.String("DAY"),
		table.String("HOURS"),
		table.String("TOTAL"),
	)
	var trackedTotal time.Duration
	for _, day := range r.Days {
		trackedTotal += day.Tracked
		trackedS := FormatDuration(day.Tracked)
		trackedTotalS := FormatDuration(trackedTotal)
		// @TODO add better padding support to table
		t.Add(
			table.String(day.From.Format("2006-01-02 ")),
			table.String(day.From.Format("Mon ")),
			table.String(trackedS+" ").Align(table.Right),
			table.String(trackedTotalS).Align(table.Right),
		)
	}
	buf.WriteString(Indent(t.String(), "  "))
	buf.WriteString("\n\n")
	for _, day := range r.Days {
		if day.Tracked == 0 {
			continue
		}
		trackedS := FormatDuration(day.Tracked)
		dayS := day.From.Format("2006-01-02 (Monday)")
		fmt.Fprintf(buf, "%s - %s\n\n", dayS, trackedS)
		buf.WriteString(Indent(strings.Join(day.Notes, "\n"), "  "))
		buf.WriteString("\n\n")
	}
	return buf.String()
}

type Report struct {
	From time.Time
	To   time.Time
	// @TODO Duration is an unfortunate name, maybe rename it
	Duration datetime.Period
	Days     []*ReportDay
}

type ReportDay struct {
	From    time.Time
	To      time.Time
	Tracked time.Duration
	Notes   []string
}
