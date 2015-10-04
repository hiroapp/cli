package main

import (
	"bufio"
	"fmt"
	"io"
	"strings"
	"time"
)

// parseDocument parses a document from r, returning its fields and remainder
// or an error if a field occured more than once, or r returned an error.
//
// Example Document:
//
//     Field1: val1
//     Field2: val2
//
//     multi line
//     remainder text
//
// @TODO: Define Document ABNF, see test case for now.
func parseDocument(r io.Reader) (fields map[string]string, remainder string, err error) {
	scanner := bufio.NewScanner(r)
	isRemainder := false
	fields = map[string]string{}
	for scanner.Scan() {
		line := scanner.Text()
		if !isRemainder {
			if line == "" {
				isRemainder = true
				continue
			}
			pair := strings.SplitN(line, ":", 2)
			for i, val := range pair {
				pair[i] = strings.TrimSpace(val)
			}
			field, val := pair[0], pair[1]
			if _, ok := fields[field]; ok {
				err = fmt.Errorf("duplicate field: %q", field)
				break
			} else {
				fields[field] = val
			}
		} else {
			remainder += line + "\n"
		}
	}
	remainder = strings.TrimRight(remainder, " \n\r")
	if err == nil {
		err = scanner.Err()
	}
	if err != nil {
		fields = nil
		remainder = ""
	}
	return
}

// ParseEntryDocument parses a entry document from r or returns an error. If
// there entry document is empty, nil is returned for it.
func ParseEntryDocument(r io.Reader) (*EntryDocument, error) {
	fields, note, err := parseDocument(r)
	if err != nil {
		return nil, err
	} else if len(fields) == 0 && len(note) == 0 {
		return nil, nil
	}
	entry := &EntryDocument{Note: note}
	for field, val := range fields {
		if field == "Start" || field == "End" {
			if val == "" {
				continue
			}
			tVal, err := time.Parse(timeLayout, val)
			if err != nil {
				return nil, err
			}
			// time.Parse will add the local zone name if the offset matches it, but
			// we're not interested in the name, so we drop it.
			_, offset := tVal.Zone()
			tVal = tVal.In(time.FixedZone("", offset))
			if field == "Start" {
				entry.Start = tVal
			} else {
				entry.End = tVal
			}
		} else {
			switch field {
			case "Id":
				entry.ID = val
			case "Category":
				entry.Category = ParseCategory(val)
			}
		}
	}
	return entry, nil
}

// EntryDocument holds a entry document used for editing entries.
type EntryDocument struct {
	ID       string
	Category []string
	Start    time.Time
	End      time.Time
	Note     string
}
