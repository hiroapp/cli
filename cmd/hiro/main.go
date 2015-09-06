package main

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/hiroapp/cli/db"
	"github.com/jawher/mow.cli"
)

// version is populated when building via the Makefile
var version string = "?"

func main() {
	app := cli.App("hiro", "Command line time tracking.")
	app.Command("start", "Start a new time entry, ending the currently active one", func(cmd *cli.Cmd) {
		resume := cmd.BoolOpt("resume", false, "Default end time and category of previous entry")
		category := cmd.StringArg("CATEGORY", "", "The category to assign to the new entry")
		cmd.Spec = "CATEGORY | --resume [CATEGORY]"
		cmd.Action = func() { cmdStart(mustDB(), *resume, *category) }
	})
	app.Command("end", "End the currently active entry", func(cmd *cli.Cmd) {
		cmd.Action = func() { cmdEnd(mustDB()) }
	})
	app.Command("ls", "Lists all time entries.", func(cmd *cli.Cmd) {
		cmd.Action = func() { cmdLs(mustDB()) }
	})
	app.Command("edit", "Edit time entry", func(cmd *cli.Cmd) {
		id := cmd.StringArg("ID", "", "The id of the entry to edit")
		cmd.Action = func() { cmdEdit(mustDB(), *id) }
	})
	app.Command("summary", "Summarize time entries", func(cmd *cli.Cmd) {
		duration := cmd.StringOpt("duration", "day", "Summary period")
		asc := cmd.BoolOpt("asc", false, "Summary order")
		firstDay := cmd.StringOpt("firstDay", "Monday", "First day of the week")
		cmd.Action = func() { cmdSummary(mustDB(), *duration, *firstDay, *asc) }
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

// splitCategory splits a colon separated category identifier into the names of
// the individual categories, e.g. "Foo:Bar:Baz" into "Foo", "Bar", "Baz".
func splitCategory(category string) []string {
	return strings.Split(category, ":")
}

func fatal(err error) {
	fmt.Fprintf(os.Stderr, "%s\n", err)
	os.Exit(1)
}
