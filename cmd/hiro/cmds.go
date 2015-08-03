package main

import (
	"fmt"
	"io"
	"time"

	"github.com/felixge/hiro/db"
)

func cmdStart(d db.DB, groupString string) {
	group := splitGroup(groupString)
	entry := &db.Entry{Group: group, Start: time.Now()}
	if err := d.Save(entry); err != nil {
		fatal(err)
	}
}

func cmdList(d db.DB) {
	itr, err := d.Query(db.Query{})
	if err != nil {
		fatal(err)
	}
	for {
		if entry, err := itr.Next(); err == io.EOF {
			break
		} else if err != nil {
			fatal(err)
		} else {
			fmt.Printf("%s - %s\n", entry.Start, entry.Group)
		}
	}
}

func cmdVersion() {
	fmt.Printf("%s\n", version)
}
