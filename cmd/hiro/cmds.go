package main

import (
	"fmt"
	"os"
	"time"

	"github.com/felixge/hiro/db"
	"github.com/felixge/hiro/term"
)

func cmdStart(d db.DB, categoryString string) {
	category := splitCategory(categoryString)
	entry := &db.Entry{Category: category, Start: time.Now()}
	if err := d.Save(entry); err != nil {
		fatal(err)
	} else if err := FprintEntry(os.Stdout, entry, PrintHideDuration|PrintHideEnd); err != nil {
		fatal(err)
	}
}

func cmdList(d db.DB) {
	itr, err := d.Query(db.Query{})
	if err != nil {
		fatal(err)
	} else if err := FprintIterator(os.Stdout, itr, PrintDefault); err != nil {
		fatal(err)
	}
}

func cmdEdit(d db.DB, id string) {
	itr, err := d.Query(db.Query{IDs: []string{id}})
	if err != nil {
		fatal(err)
	}
	e := term.NewEditor()
	if err := FprintIterator(e, itr, PrintSeparator|PrintHideDuration); err != nil {
		fatal(err)
	} else if err := e.Run(); err != nil {
		fatal(err)
	} else if entries, err := ParseEntries(e); err != nil {
		fatal(err)
	} else if l := len(entries); l == 0 {
		return
	} else if l > 1 {
		fatal(fmt.Errorf("editing multiple entries is not supported yet"))
	} else if err := d.Save(entries[0]); err != nil {
		fatal(err)
	} else if err := FprintIterator(os.Stdout, db.EntryIterator(entries), PrintDefault); err != nil {
		fatal(err)
	}
}

func cmdVersion() {
	fmt.Printf("%s\n", version)
}
