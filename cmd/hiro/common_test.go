package main

import (
	"strings"
	"testing"
	"time"

	"github.com/felixge/godebug/pretty"
	"github.com/felixge/hiro/db"
)

func TestParseEntries(t *testing.T) {
	var zoneCEST = time.FixedZone("CEST", int((2 * time.Hour).Seconds()))
	tests := []struct {
		Text    string
		Want    []*db.Entry
		WantErr string
	}{
		// empty entry
		{},
		// simple entry, empty end
		{
			Text: `
Id:       3847c8e1-8d25-46e5-a09e-36c40266c871
Category: Work:Hiro
Start:    2015-08-04 10:13:51 +0000 UTC
End:

Hello World
`,
			Want: []*db.Entry{
				&db.Entry{
					ID:       "3847c8e1-8d25-46e5-a09e-36c40266c871",
					Category: []string{"Work", "Hiro"},
					Start:    time.Date(2015, 8, 4, 10, 13, 51, 0, time.UTC),
					Note:     "Hello World",
				},
			},
		},
		// multiple multi-line entries, one with end, one without
		{

			Text: `
Id:       0cd7ca90-c19a-4b11-8d98-cd29d467f9f8
Category: Misc
Start:    2015-08-04 10:19:51 +0200 CEST

This is a time entry.

With a newline.

8< ----- do not remove this separator ----- >8

Id:       9fc585e2-0314-4816-95fc-70e0e75f3b1d
Category: Another Category
Start:    2015-08-04 10:13:51 +0000 UTC
End:      2015-08-04 10:19:51 +0200 CEST

And this is another time entry.

With a newline.
`,
			Want: []*db.Entry{
				&db.Entry{
					ID:       "0cd7ca90-c19a-4b11-8d98-cd29d467f9f8",
					Category: []string{"Misc"},
					Start:    time.Date(2015, 8, 4, 10, 19, 51, 0, zoneCEST),
					Note:     "This is a time entry.\n\nWith a newline.",
				},
				&db.Entry{
					ID:       "9fc585e2-0314-4816-95fc-70e0e75f3b1d",
					Category: []string{"Another Category"},
					Start:    time.Date(2015, 8, 4, 10, 13, 51, 0, time.UTC),
					End:      time.Date(2015, 8, 4, 10, 19, 51, 0, zoneCEST),
					Note:     "And this is another time entry.\n\nWith a newline.",
				},
			},
		},
	}
	for i, test := range tests {
		entries, err := ParseEntries(strings.NewReader(test.Text))
		var gotErr string
		if err != nil {
			gotErr = err.Error()
		}
		if gotErr != test.WantErr {
			t.Errorf("test %d: got=%q want=%q", i, gotErr, test.WantErr)
		}
		if diff := (&pretty.Config{Diffable: true, PrintStringers: true}).Compare(entries, test.Want); diff != "" {
			t.Errorf("test %d: %s", i, diff)
		}
	}
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		Duration time.Duration
		Want     string
	}{
		{
			Want: "0:00:00",
		},
		{
			Duration: 1 * time.Second,
			Want:     "0:00:01",
		},
		{
			Duration: 1*time.Minute + 23*time.Second,
			Want:     "0:01:23",
		},
		{
			Duration: 1*time.Hour + 23*time.Minute + 45*time.Second,
			Want:     "1:23:45",
		},
		{
			Duration: 12*time.Hour + 34*time.Minute + 56*time.Second,
			Want:     "12:34:56",
		},
		{
			Duration: 123*time.Hour + 34*time.Minute + 56*time.Second,
			Want:     "123:34:56",
		},
	}
	for _, test := range tests {
		got := FormatDuration(test.Duration)
		if got != test.Want {
			t.Errorf("got=%s want=%s", got, test.Want)
		}
	}
}
