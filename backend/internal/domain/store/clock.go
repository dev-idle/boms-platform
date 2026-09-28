package store

import (
	"fmt"
	"time"
)

// ParseClock reads a wall-clock time written HH:MM (00:00-23:59) as an offset
// from midnight.
func ParseClock(s string) (time.Duration, error) {
	if len(s) != 5 || s[2] != ':' || !digits(s[:2]) || !digits(s[3:]) {
		return 0, fmt.Errorf("clock time %q is not HH:MM", s)
	}
	hour := int(s[0]-'0')*10 + int(s[1]-'0')
	minute := int(s[3]-'0')*10 + int(s[4]-'0')
	if hour > 23 || minute > 59 {
		return 0, fmt.Errorf("clock time %q is not a time of day", s)
	}
	return time.Duration(hour)*time.Hour + time.Duration(minute)*time.Minute, nil
}

// FormatClock writes an offset from midnight as HH:MM.
func FormatClock(d time.Duration) string {
	return fmt.Sprintf("%02d:%02d", int(d/time.Hour), int(d%time.Hour/time.Minute))
}

func digits(s string) bool {
	for i := range len(s) {
		if s[i] < '0' || s[i] > '9' {
			return false
		}
	}
	return true
}
