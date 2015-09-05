package datetime

import (
	"testing"
	"time"

	"github.com/felixge/godebug/pretty"
)

var diffConfig = &pretty.Config{Diffable: true, PrintStringers: true}

func TestIterator(t *testing.T) {
	zone := time.FixedZone("", 3600)
	tests := []struct {
		Duration Duration
		Offset   time.Duration
		FirstDay time.Weekday
		Want     [][2]time.Time
	}{
		{
			Duration: Day,
			Offset:   5 * time.Hour,
			Want: [][2]time.Time{
				{
					time.Date(2015, 2, 27, 0, 0, 0, 0, zone),
					time.Date(2015, 2, 27, 23, 59, 59, 999999999, zone),
				},
				{
					time.Date(2015, 2, 28, 0, 0, 0, 0, zone),
					time.Date(2015, 2, 28, 23, 59, 59, 999999999, zone),
				},
				{
					time.Date(2015, 3, 1, 0, 0, 0, 0, zone),
					time.Date(2015, 3, 1, 23, 59, 59, 999999999, zone),
				},
			},
		},

		{
			Duration: Week,
			FirstDay: time.Monday,
			Offset:   4 * 24 * time.Hour,
			Want: [][2]time.Time{
				{
					time.Date(2015, 2, 16, 0, 0, 0, 0, zone),
					time.Date(2015, 2, 22, 23, 59, 59, 999999999, zone),
				},
				{
					time.Date(2015, 2, 23, 0, 0, 0, 0, zone),
					time.Date(2015, 3, 1, 23, 59, 59, 999999999, zone),
				},
				{
					time.Date(2015, 3, 2, 0, 0, 0, 0, zone),
					time.Date(2015, 3, 8, 23, 59, 59, 999999999, zone),
				},
			},
		},

		{
			Duration: Month,
			Offset:   72 * time.Hour,
			Want: [][2]time.Time{
				{
					time.Date(2015, 1, 1, 0, 0, 0, 0, zone),
					time.Date(2015, 1, 31, 23, 59, 59, 999999999, zone),
				},
				{
					time.Date(2015, 2, 1, 0, 0, 0, 0, zone),
					time.Date(2015, 2, 28, 23, 59, 59, 999999999, zone),
				},
				{
					time.Date(2015, 3, 1, 0, 0, 0, 0, zone),
					time.Date(2015, 3, 31, 23, 59, 59, 999999999, zone),
				},
			},
		},

		{
			Duration: Year,
			Offset:   100 * 24 * time.Hour,
			Want: [][2]time.Time{
				{
					time.Date(2013, 1, 1, 0, 0, 0, 0, zone),
					time.Date(2013, 12, 31, 23, 59, 59, 999999999, zone),
				},
				{
					time.Date(2014, 1, 1, 0, 0, 0, 0, zone),
					time.Date(2014, 12, 31, 23, 59, 59, 999999999, zone),
				},
				{
					time.Date(2015, 1, 1, 0, 0, 0, 0, zone),
					time.Date(2015, 12, 31, 23, 59, 59, 999999999, zone),
				},
			},
		},
	}

	for i, test := range tests {
		for _, asc := range []bool{true, false} {
			cursor := test.Want[0][0]
			if !asc {
				cursor = test.Want[len(test.Want)-1][0]
			}
			cursor = cursor.Add(test.Offset)
			var (
				itr  = NewIterator(cursor, test.Duration, asc, test.FirstDay)
				want = make([][2]time.Time, len(test.Want))
				got  [][2]time.Time
			)
			for i := 0; i < len(want); i++ {
				j := i
				if !asc {
					j = len(test.Want) - i - 1
				}
				want[j] = test.Want[i]
			}
			for _ = range want {
				from, to := itr.Next()
				got = append(got, [2]time.Time{from, to})
			}
			if diff := diffConfig.Compare(got, want); diff != "" {
				t.Errorf("test %d cursor=%s asc=%t: %s", i, cursor, asc, diff)
			}
		}
	}
}
