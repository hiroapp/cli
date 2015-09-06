package main

import (
	"bufio"
	"fmt"
	"io"
	"regexp"
	"strings"
	"text/template"
	"time"

	"github.com/felixge/hiro/db"
)

var tmpl = template.Must(template.New("entry").Funcs(template.FuncMap{
	"now":  func() time.Time { return time.Now().Truncate(time.Second) },
	"join": strings.Join,
	"format": func(t time.Time) string {
		return t.Format(timeLayout)
	},
}).Parse(strings.TrimSpace(`
Id:       {{.Entry.ID}}
Category: {{join .Entry.Category ":"}}
Start:    {{format .Entry.Start}}
{{if not .HideEnd}}End:      {{if .Entry.End.IsZero}}{{else}}{{format .Entry.End}}{{end}}
{{end}}{{if not .HideDuration}}Duration: {{.Entry.Duration now}}
{{end}}
{{if .Entry.Note}}{{.Entry.Note}}
{{end}}
`)))

func FprintEntry(w io.Writer, e *db.Entry, m PrintMask) error {
	return tmpl.Execute(w, map[string]interface{}{
		"Entry":        e,
		"HideDuration": m&PrintHideDuration > 0,
		"HideEnd":      m&PrintHideEnd > 0,
	})
}

func FprintIterator(w io.Writer, itr db.Iterator, m PrintMask) error {
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
			if err := FprintEntry(w, entry, m); err != nil {
				return err
			}
		}
	}
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

// ParseEntries parses entries in a, for now, poorly specified plaintext
// format from r and returns them or an error.
//
// @TODO properly define the plaintext format and implement good error
// handling.
func ParseEntries(r io.Reader) ([]*db.Entry, error) {
	var (
		entries []*db.Entry
		entry   = &db.Entry{}
		scanner = bufio.NewScanner(r)
		isNote  = false
	)
	for {
		var (
			ok   = scanner.Scan()
			line string
		)
		if ok {
			line = scanner.Text()
		}
		if !ok || line == entrySeparator {
			isNote = false
			entry.Note = strings.TrimSpace(entry.Note)
			if !entry.Empty() {
				entries = append(entries, entry)
				entry = &db.Entry{}
			}
			if !ok {
				break
			}
			continue
		}
		matches := entryField.FindStringSubmatch(line)
		if isNote {
			entry.Note += line + "\n"
			continue
		} else if line == "" {
			if !entry.Empty() {
				isNote = true
			}
			continue
		} else if len(matches) != 3 {
			return nil, fmt.Errorf("bad line: %q", line)
		}
		field, val := matches[1], matches[2]
		switch fieldLow := strings.ToLower(field); fieldLow {
		case "id":
			entry.ID = val
		case "category":
			entry.Category = splitCategory(val)
		case "start", "end":
			if val == "" {
				continue
			}
			tVal, err := time.Parse(timeLayout, val)
			if err != nil {
				return nil, err
			}
			if fieldLow == "start" {
				entry.Start = tVal
			} else {
				entry.End = tVal
			}
		default:
			return nil, fmt.Errorf("Unknown field: %s", field)
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("Failed to scan: %s", err)
	}
	return entries, nil
}

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
