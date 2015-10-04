package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/bradfitz/slice"
	"github.com/hiroapp/cli/datetime"
	"github.com/hiroapp/cli/db"
	"github.com/hiroapp/cli/table"
	"github.com/hiroapp/cli/term"
)

func cmdStart(d db.DB, resume bool, categoryS string) {
	entries, err := active(d)
	if err != nil {
		fatal(err)
	}
	now := time.Now()
	path, err := d.CategoryPath(ParseCategory(categoryS), true)
	if err != nil {
		fatal(err)
	}
	entry := &db.Entry{CategoryID: path.CategoryID(), Start: now}
	if resume {
		last, err := Last(d)
		if err != nil {
			fatal(err)
		}
		if !last.End.IsZero() {
			entry.Start = last.End
		}
		if entry.CategoryID == "" {
			entry.CategoryID = last.CategoryID
		}
	}
	if err := d.SaveEntry(entry); err != nil {
		fatal(err)
	}
	FprintEntry(os.Stdout, entry, path, PrintHideDuration|PrintHideEnd)
	if err := endAt(d, entries, now); err != nil {
		fatal(err)
	}
}

func cmdEnd(d db.DB) {
	if entries, err := active(d); err != nil {
		fatal(err)
	} else if err := endAt(d, entries, time.Now()); err != nil {
		fatal(err)
	}
}

// ById returns the entry with the given id, or an error.
func ById(d db.DB, id string) (*db.Entry, error) {
	itr, err := d.Query(db.Query{IDs: []string{id}})
	if err != nil {
		return nil, err
	}
	defer itr.Close()
	if entry, err := itr.Next(); err == io.EOF {
		return nil, fmt.Errorf("entry does not exist: %s", id)
	} else {
		return entry, err
	}
}

// Last returns the last entry or an error.
func Last(d db.DB) (*db.Entry, error) {
	itr, err := d.Query(db.Query{})
	if err != nil {
		return nil, err
	}
	defer itr.Close()
	if entry, err := itr.Next(); err == io.EOF {
		return nil, errors.New("db is empty")
	} else {
		return entry, err
	}
}

func active(d db.DB) ([]*db.Entry, error) {
	if itr, err := d.Query(db.Query{Active: true}); err != nil {
		return nil, err
	} else {
		return db.IteratorEntries(itr)
	}
}

func endAt(d db.DB, entries []*db.Entry, t time.Time) error {
	categories, err := d.Categories()
	if err != nil {
		return err
	}
	for _, entry := range entries {
		entry.End = t
		if err := d.SaveEntry(entry); err != nil {
			return err
		}
		FprintEntry(os.Stdout, entry, categories.Path(entry.CategoryID), PrintDefault)
	}
	return nil
}

func cmdLs(d db.DB, categoryS string, asc bool) {
	if categories, err := d.Categories(); err != nil {
		fatal(err)
	} else if path, err := d.CategoryPath(ParseCategory(categoryS), false); err != nil {
		fatal(err)
	} else if itr, err := d.Query(db.Query{Asc: asc, CategoryID: path.CategoryID()}); err != nil {
		fatal(err)
	} else {
		FprintIterator(os.Stdout, itr, categories, PrintDefault)
	}
}

func cmdEdit(d db.DB, id string) {
	var (
		entry *db.Entry
		err   error
	)
	if id != "" {
		entry, err = ById(d, id)
	} else {
		entry, err = Last(d)
	}
	if err != nil {
		fatal(err)
	}
	categories, err := d.Categories()
	if err != nil {
		fatal(err)
	}
	e := term.NewEditor()
	FprintEntry(e, entry, categories.Path(entry.CategoryID), PrintSeparator|PrintHideDuration)
	if err := e.Run(); err != nil {
		fatal(err)
	} else if doc, err := ParseEntryDocument(e); err != nil {
		fatal(err)
	} else if doc == nil {
		return
	} else {
		entry := &db.Entry{
			ID:    doc.ID,
			Start: doc.Start,
			End:   doc.End,
			Note:  doc.Note,
		}
		if path, err := d.CategoryPath(doc.Category, false); err != nil {
			fatal(err)
		} else {
			entry.CategoryID = path.CategoryID()
		}
		if err := d.SaveEntry(entry); err != nil {
			fatal(err)
		} else {
			FprintIterator(os.Stdout, db.EntryIterator([]*db.Entry{entry}), categories, PrintDefault)
		}
	}
}

func cmdRm(d db.DB, id string) {
	entry, err := ById(d, id)
	if err != nil {
		fatal(err)
	} else if err := d.Remove(id); err != nil {
		fatal(err)
	}
	categories, err := d.Categories()
	if err != nil {
		fatal(err)
	}
	FprintEntry(os.Stdout, entry, categories.Path(entry.CategoryID), PrintDefault)
}

func cmdSummary(d db.DB, periodS, firstDayS string) {
	period, err := datetime.ParseDuration(periodS)
	if err != nil {
		fatal(err)
	}
	firstDay, err := datetime.ParseWeekday(firstDayS)
	if err != nil {
		fatal(err)
	}
	categories, err := d.Categories()
	if err != nil {
		fatal(err)
	}
	itr, err := NewSummaryIterator(d, period, firstDay, time.Now())
	if err != nil {
		fatal(err)
	}
	defer itr.Close()
	for {
		if summary, err := itr.Next(); err == io.EOF {
			break
		} else if err != nil {
			fatal(err)
		} else {
			fmt.Printf("%s\n\n", PeriodHeadline(summary.From, summary.To, period))
			names := make(map[string]string)
			order := make([]string, 0, len(summary.Categories))
			for id, _ := range summary.Categories {
				names[id] = FormatCategory(categories.Path(id))
				order = append(order, id)
			}
			slice.Sort(order, func(i, j int) bool {
				return summary.Categories[order[i]] > summary.Categories[order[j]]
			})
			t := table.New().Padding(" ")
			for _, id := range order {
				d := FormatDuration(summary.Categories[id])
				t.Add(table.String(names[id]), table.String(d).Align(table.Right))
			}
			fmt.Printf("%s\n", Indent(t.String(), "  "))
		}
	}
}

func cmdReport(d db.DB, categoryS, durationS, firstDayS string) {
	duration, err := datetime.ParseDuration(durationS)
	if err != nil {
		fatal(err)
	} else if duration == datetime.Day {
		fatal(errors.New("bad duration: day"))
	}
	firstDay, err := datetime.ParseWeekday(firstDayS)
	if err != nil {
		fatal(err)
	}
	path, err := d.CategoryPath(ParseCategory(categoryS), false)
	if err != nil {
		fatal(err)
	}
	entryItr, err := d.Query(db.Query{CategoryID: path.CategoryID()})
	if err != nil {
		fatal(err)
	}
	defer entryItr.Close()
	entry, err := entryItr.Next()
	if err == io.EOF {
		return
	} else if err != nil {
		fatal(err)
	}
	// @TODO move logic into separate function
	now := time.Now()
	reportItr := datetime.NewIterator(entry.Start, duration, false, firstDay)
	report := &Report{Duration: duration}
	report.From, report.To = reportItr.Next()
	dayItr := datetime.NewIterator(report.To, datetime.Day, false, firstDay)
	day := &ReportDay{}
	day.From, day.To = dayItr.Next()
	report.Days = append([]*ReportDay{day}, report.Days...)
	noteAssigned := false

outer:
	for {
		var overlap time.Duration
		for {
			overlap = entry.PartialDuration(now, day.From, day.To)
			if overlap > 0 {
				day.Tracked += overlap
				if !noteAssigned {
					note := strings.Trim(entry.Note, "\n")
					if note != "" {
						day.Notes = append([]string{note}, day.Notes...)
					}
					noteAssigned = true
				}
			}
			if !entry.Start.Before(day.From) {
				entry, err = entryItr.Next()
				if err == io.EOF {
					fmt.Fprint(os.Stdout, FormatReport(report))
					break outer
				} else if err != nil {
					fatal(err)
				}
				noteAssigned = false
				continue
			}
			day = &ReportDay{}
			day.From, day.To = dayItr.Next()
			if day.To.Before(report.From) {
				fmt.Fprint(os.Stdout, FormatReport(report))
				report = &Report{Duration: duration}
				report.From, report.To = reportItr.Next()
			}
			report.Days = append([]*ReportDay{day}, report.Days...)
		}
	}
}

func cmdVersion() {
	fmt.Printf("%s\n", version)
}
