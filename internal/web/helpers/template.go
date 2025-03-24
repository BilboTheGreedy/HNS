package helpers

import (
	"time"
)

// FormatDate formats a time.Time for display in templates
func FormatDate(t time.Time) string {
	return t.Format("Jan 02, 2006 15:04")
}

// FormatDateShort formats a time.Time for compact display
func FormatDateShort(t time.Time) string {
	return t.Format("2006-01-02")
}

// FormatDateTime formats a time.Time with both date and time
func FormatDateTime(t time.Time) string {
	return t.Format("2006-01-02 15:04:05")
}

// FormatRelativeTime returns a relative time string (e.g., "2 hours ago")
func FormatRelativeTime(t time.Time) string {
	now := time.Now()
	diff := now.Sub(t)

	switch {
	case diff < time.Minute:
		return "just now"
	case diff < time.Hour:
		minutes := int(diff.Minutes())
		if minutes == 1 {
			return "1 minute ago"
		}
		return timeFormat(minutes, "minute")
	case diff < 24*time.Hour:
		hours := int(diff.Hours())
		if hours == 1 {
			return "1 hour ago"
		}
		return timeFormat(hours, "hour")
	case diff < 48*time.Hour:
		return "yesterday"
	case diff < 7*24*time.Hour:
		days := int(diff.Hours() / 24)
		return timeFormat(days, "day")
	case diff < 30*24*time.Hour:
		weeks := int(diff.Hours() / 24 / 7)
		if weeks == 1 {
			return "1 week ago"
		}
		return timeFormat(weeks, "week")
	default:
		return t.Format("Jan 02, 2006")
	}
}

// timeFormat returns a formatted time string with plural handling
func timeFormat(count int, unit string) string {
	if count == 1 {
		return "1 " + unit + " ago"
	}
	return timeFormat(count) + " " + unit + "s ago"
}
