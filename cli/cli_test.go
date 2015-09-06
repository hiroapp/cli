package cli

import (
	"testing"
	"time"
)

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
