package main

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/felixge/hiro/db"
	"github.com/felixge/hiro/term/editor"
)

func cmdStart(d db.DB, groupString string) {
	group := splitGroup(groupString)
	entry := &db.Entry{Group: group, Start: time.Now()}
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

func cmdEdit(d db.DB, ids ...string) {
	itr, err := d.Query(db.Query{IDs: ids})
	if err != nil {
		fatal(err)
	}
	e := editor.New()
	defer e.Close()
	if err := FprintIterator(e, itr, PrintDefault); err != nil {
		fatal(err)
	} else if err := e.Run(); err != nil {
		fatal(err)
	}
	io.Copy(os.Stdout, e)
}

func cmdVersion() {
	fmt.Printf("%s\n", version)
}
