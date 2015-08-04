package main

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/felixge/hiro/db"
	"github.com/jawher/mow.cli"
)

// version is populated when building via the Makefile
var version string = "?"

func main() {
	app := cli.App("hiro", "Command line time tracking.")
	app.Command("start", "Start a new time entry, ending the currently active one", func(cmd *cli.Cmd) {
		group := cmd.StringArg("GROUP", "", "The group to assign to the new entry")
		cmd.Action = func() { cmdStart(mustDB(), *group) }
	})
	app.Command("list", "Lists all time entries.", func(cmd *cli.Cmd) {
		cmd.Action = func() { cmdList(mustDB()) }
	})
	app.Command("edit", "Edit time entry", func(cmd *cli.Cmd) {
		id := cmd.StringsArg("ID", nil, "The id of the entry to edit")
		cmd.Action = func() { cmdEdit(mustDB(), *id...) }
		cmd.Spec = "ID..."
	})
	app.Command("version", "Prints the version", func(cmd *cli.Cmd) {
		cmd.Action = cmdVersion
	})
	app.Run(os.Args)
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

// splitGroup splits a colon separated group identifier into the names of the
// individual groups, e.g. "Foo:Bar:Baz" into "Foo", "Bar", "Baz".
func splitGroup(group string) []string {
	return strings.Split(group, ":")
}

func fatal(err error) {
	fmt.Fprintf(os.Stderr, "%s\n", err)
	os.Exit(1)
}
