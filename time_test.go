package hiro

import (
	"io"
	"testing"
	"time"

	"github.com/felixge/godebug/pretty"
)

func TestIntervalIterator(t *testing.T) {
	diffConfig := &pretty.Config{Diffable: true, PrintStringers: true}
	zone := time.FixedZone("", 3600)
	tests := []struct {
		From     time.Time
		To       time.Time
		Interval Interval
		Want     []TimeRange
	}{
		{
			From:     time.Date(2015, 2, 27, 0, 0, 0, 0, zone),
			To:       time.Date(2015, 3, 1, 0, 0, 0, 0, zone),
			Interval: Day,
			Want: []TimeRange{
				{
					From: time.Date(2015, 2, 27, 0, 0, 0, 0, zone),
					To:   time.Date(2015, 2, 27, 23, 59, 59, 999999999, zone),
				},
				{
					From: time.Date(2015, 2, 28, 0, 0, 0, 0, zone),
					To:   time.Date(2015, 2, 28, 23, 59, 59, 999999999, zone),
				},
				{
					From: time.Date(2015, 3, 1, 0, 0, 0, 0, zone),
					To:   time.Date(2015, 3, 1, 23, 59, 59, 999999999, zone),
				},
			},
		},
		{
			From:     time.Date(2015, 1, 1, 0, 0, 0, 0, zone),
			To:       time.Date(2015, 3, 1, 0, 0, 0, 0, zone),
			Interval: Month,
			Want: []TimeRange{
				{
					From: time.Date(2015, 1, 1, 0, 0, 0, 0, zone),
					To:   time.Date(2015, 1, 31, 23, 59, 59, 999999999, zone),
				},
				{
					From: time.Date(2015, 2, 1, 0, 0, 0, 0, zone),
					To:   time.Date(2015, 2, 28, 23, 59, 59, 999999999, zone),
				},
				{
					From: time.Date(2015, 3, 1, 0, 0, 0, 0, zone),
					To:   time.Date(2015, 3, 31, 23, 59, 59, 999999999, zone),
				},
			},
		},
		{
			From:     time.Date(2013, 1, 1, 0, 0, 0, 0, zone),
			To:       time.Date(2015, 1, 1, 0, 0, 0, 0, zone),
			Interval: Year,
			Want: []TimeRange{
				{
					From: time.Date(2013, 1, 1, 0, 0, 0, 0, zone),
					To:   time.Date(2013, 12, 31, 23, 59, 59, 999999999, zone),
				},
				{
					From: time.Date(2014, 1, 1, 0, 0, 0, 0, zone),
					To:   time.Date(2014, 12, 31, 23, 59, 59, 999999999, zone),
				},
				{
					From: time.Date(2015, 1, 1, 0, 0, 0, 0, zone),
					To:   time.Date(2015, 12, 31, 23, 59, 59, 999999999, zone),
				},
			},
		},
		{
			From:     time.Date(2015, 8, 31, 0, 0, 0, 0, zone),
			To:       time.Date(2015, 9, 14, 0, 0, 0, 0, zone),
			Interval: Week,
			Want: []TimeRange{
				{
					From: time.Date(2015, 8, 31, 0, 0, 0, 0, zone),
					To:   time.Date(2015, 9, 6, 23, 59, 59, 999999999, zone),
				},
				{
					From: time.Date(2015, 9, 7, 0, 0, 0, 0, zone),
					To:   time.Date(2015, 9, 13, 23, 59, 59, 999999999, zone),
				},
				{
					From: time.Date(2015, 9, 14, 0, 0, 0, 0, zone),
					To:   time.Date(2015, 9, 20, 23, 59, 59, 999999999, zone),
				},
			},
		},
	}
	for i, test := range tests {
		var (
			itr = NewIntervalIterator(test.From, test.To, test.Interval)
			got []TimeRange
		)
		for {
			if tr, err := itr.Next(); err == io.EOF {
				break
			} else if err != nil {
				t.Errorf("test %d: %s", i, err)
			} else {
				got = append(got, tr)
			}
		}
		if diff := diffConfig.Compare(got, test.Want); diff != "" {
			t.Errorf("test %d: %s", i, diff)
		}
	}
}
