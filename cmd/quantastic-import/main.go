// Command quantastic-import imports legacy time.json data from quantastic.
package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/felixge/hiro/db"
)

func main() {
	if err := run(); err != nil {
		fatal(err)
	}
}

func run() error {
	flag.Parse()
	d := mustDB()
	file, err := os.Open(flag.Arg(0))
	if err != nil {
		return err
	}
	defer file.Close()
	var m map[string]*Entry
	if err := json.NewDecoder(file).Decode(&m); err != nil {
		return err
	}
	for _, legacy := range m {
		entry := &db.Entry{
			Start:    legacy.Start.Time,
			End:      legacy.End.Time,
			Category: legacy.Category,
			Note:     legacy.Note,
		}
		if err := d.Save(entry); err != nil {
			return err
		}
	}
	return nil
}

type Entry struct {
	Category []string
	Start    Time
	End      Time
	Note     string
}

type Time struct {
	time.Time
}

func (t *Time) UnmarshalJSON(data []byte) error {
	s := string(data)
	if s == "null" {
		return nil
	}
	ti, err := time.Parse(`"`+time.RFC3339+`"`, s)
	if err != nil {
		return err
	}
	t.Time = ti
	return nil
}

func mustDB() db.DB {
	dir := os.Getenv("HIRO_DIR")
	if dir == "" {
		fatal(errors.New("HIRO_DIR env variable must be set"))
	} else if d, err := db.New(dir); err != nil {
		fatal(fmt.Errorf("could not open db: %s", err))
	} else {
		return d
	}
	panic("unreachable")
}

func fatal(err error) {
	fmt.Printf("%s\n", err)
	os.Exit(1)
}
