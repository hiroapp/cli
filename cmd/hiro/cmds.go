package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/bradfitz/slice"
	"github.com/hiroapp/cli"
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
	category := ParseCategory(categoryS)
	entry := &db.Entry{Category: category, Start: now}
	if resume {
		last, err := Last(d)
		if err != nil {
			fatal(err)
		}
		if !last.End.IsZero() {
			entry.Start = last.End
		}
		if categoryS == "" {
			entry.Category = last.Category
		}
	}
	if err := d.Save(entry); err != nil {
		fatal(err)
	}
	FprintEntry(os.Stdout, entry, PrintHideDuration|PrintHideEnd)
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
	for _, entry := range entries {
		entry.End = t
		if err := d.Save(entry); err != nil {
			return err
		}
		FprintEntry(os.Stdout, entry, PrintDefault)
	}
	return nil
}

func cmdLs(d db.DB, categoryS string, asc bool) {
	itr, err := d.Query(db.Query{Asc: asc, Category: ParseCategory(categoryS)})
	if err != nil {
		fatal(err)
	}
	FprintIterator(os.Stdout, itr, PrintDefault)
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
	e := term.NewEditor()
	FprintEntry(e, entry, PrintSeparator|PrintHideDuration)
	if err := e.Run(); err != nil {
		fatal(err)
	} else if entries, err := ParseEntries(e); err != nil {
		fatal(err)
	} else if l := len(entries); l == 0 {
		return
	} else if l > 1 {
		fatal(fmt.Errorf("editing multiple entries is not supported yet"))
	} else if err := d.Save(entries[0]); err != nil {
		fatal(err)
	} else {
		FprintIterator(os.Stdout, db.EntryIterator(entries), PrintDefault)
	}
}

func cmdRm(d db.DB, id string) {
	entry, err := ById(d, id)
	if err != nil {
		fatal(err)
	} else if err := d.Remove(id); err != nil {
		fatal(err)
	}
	FprintEntry(os.Stdout, entry, PrintDefault)
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
	h := hiro.NewHiro(d)
	itr, err := h.SummaryIterator(period, firstDay, time.Now())
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
			names := make([]string, 0, len(summary.Categories))
			for name, _ := range summary.Categories {
				names = append(names, name)
			}
			slice.Sort(names, func(i, j int) bool {
				return summary.Categories[names[i]] > summary.Categories[names[j]]
			})
			t := table.New().Padding(" ")
			for _, name := range names {
				d := FormatDuration(summary.Categories[name])
				t.Add(table.String(name), table.String(d).Align(table.Right))
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
	entryItr, err := d.Query(db.Query{Category: ParseCategory(categoryS)})
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
