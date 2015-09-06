// Package cli provides hiro specific command line utility functions.
package cli

import (
	"fmt"
	"time"
)

// FormatDuration returns the duration as a H:MM:SS formated string, e.g.
// "1:03:03" for 1h2m3s or "123:45:56" for 123h45m56s.
func FormatDuration(d time.Duration) string {
	hours := d / time.Hour
	d -= hours * time.Hour
	minutes := d / time.Minute
	d -= minutes * time.Minute
	seconds := d / time.Second
	return fmt.Sprintf("%d:%02d:%02d", hours, minutes, seconds)
}
