package main

import (
	"errors"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/kylelemons/godebug/pretty"
)

var diffConfig = &pretty.Config{
	Diffable:       true,
	PrintStringers: true,
}

func Test_parseDocument(t *testing.T) {
	tests := []struct {
		Name      string
		R         io.Reader
		Fields    map[string]string
		Remainder string
		Err       error
	}{
		{
			Name: "empty document",
			R:    strings.NewReader(""),
		},
		{
			Name:   "simple field",
			R:      strings.NewReader("foo:bar"),
			Fields: map[string]string{"foo": "bar"},
		},
		{
			Name:   "simple field with trim",
			R:      strings.NewReader("   foo  :   bar   "),
			Fields: map[string]string{"foo": "bar"},
		},
		{
			Name:   "simple field with colon value",
			R:      strings.NewReader("foo: bar:baz"),
			Fields: map[string]string{"foo": "bar:baz"},
		},
		{
			Name:   "multiple fields",
			R:      strings.NewReader("a: b\nc: d\ne: f"),
			Fields: map[string]string{"a": "b", "c": "d", "e": "f"},
		},
		{
			Name:      "multiple fields with remainder",
			R:         strings.NewReader("a: b\nc: d\ne: f\n\nsome remainder"),
			Fields:    map[string]string{"a": "b", "c": "d", "e": "f"},
			Remainder: "some remainder",
		},
		{
			Name:      "multiple fields with remainder trim",
			R:         strings.NewReader("a: b\nc: d\ne: f\n\n some remainder\r\n\ntext\n\r \n"),
			Fields:    map[string]string{"a": "b", "c": "d", "e": "f"},
			Remainder: " some remainder\n\ntext",
		},
		{
			Name: "duplicate field",
			R:    strings.NewReader("a: b\na: c"),
			Err:  errors.New("duplicate field: \"a\""),
		},
		{
			Name:      "just remainder",
			R:         strings.NewReader("\nfoo"),
			Remainder: "foo",
		},
	}
	for _, test := range tests {
		fields, remainder, err := parseDocument(test.R)
		got := []interface{}{fields, remainder, err}
		want := []interface{}{test.Fields, test.Remainder, test.Err}
		if diff := pretty.Compare(got, want); diff != "" {
			t.Errorf("test %q: %s", test.Name, diff)
		}
	}
}

func TestParseEntryDocument(t *testing.T) {
	tests := []struct {
		Name  string
		R     io.Reader
		Entry *EntryDocument
		Err   error
	}{
		{
			Name: "empty entry",
			R:    strings.NewReader(""),
		},
		{
			Name: "full entry",
			R: strings.NewReader(`Id: 1
Category: Work:Hiro
Start: 2015-10-04 12:59:17 +0200
End: 2015-10-04 13:23:04 +0100

The cake is a lie!`),
			Entry: &EntryDocument{
				ID:       "1",
				Category: []string{"Work", "Hiro"},
				Start:    time.Date(2015, 10, 04, 12, 59, 17, 0, time.FixedZone("", 2*60*60)),
				End:      time.Date(2015, 10, 04, 13, 23, 4, 0, time.FixedZone("", 1*60*60)),
				Note:     "The cake is a lie!",
			},
		},
		{
			Name: "empty end",
			R: strings.NewReader(`Id: 1
Category: Work:Hiro
Start: 2015-10-04 12:59:17 +0200
End:

The cake is a lie!`),
			Entry: &EntryDocument{
				ID:       "1",
				Category: []string{"Work", "Hiro"},
				Start:    time.Date(2015, 10, 04, 12, 59, 17, 0, time.FixedZone("", 2*60*60)),
				Note:     "The cake is a lie!",
			},
		},
	}
	for _, test := range tests {
		entry, err := ParseEntryDocument(test.R)
		got := []interface{}{entry, err}
		want := []interface{}{test.Entry, test.Err}
		if diff := diffConfig.Compare(got, want); diff != "" {
			t.Errorf("test %q: %s", test.Name, diff)
		}
	}
}
