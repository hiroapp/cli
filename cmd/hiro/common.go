package main

import (
	"fmt"
	"io"
	"strings"
	"text/template"
	"time"

	"github.com/felixge/hiro/db"
)

var tmpl = template.Must(template.New("entry").Funcs(template.FuncMap{
	"now":  func() time.Time { return time.Now().Truncate(time.Second) },
	"join": strings.Join,
}).Parse(strings.TrimSpace(`
Id:       {{.Entry.ID}}
Category: {{join .Entry.Category ":"}}
Start:    {{.Entry.Start}}
{{if not .HideEnd}}End:      {{if .Entry.End.IsZero}}{{else}}{{.End}}{{end}}
{{end}}{{if not .HideDuration}}Duration: {{.Entry.Duration now}}
{{end}}
`)))

func FprintEntry(w io.Writer, e *db.Entry, m PrintMask) error {
	return tmpl.Execute(w, map[string]interface{}{
		"Entry":        e,
		"HideDuration": m & PrintHideDuration,
		"HideEnd":      m & PrintHideEnd,
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
				if _, err := fmt.Fprintf(w, "\n"); err != nil {
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
)

func ParseEntries(r io.Reader) ([]*db.Entry, error) {
	return nil, nil
}
