package handler

import "time"

const dateLayout = "2006-01-02"

func datePtrToString(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.Format(dateLayout)
}

func parseDate(s string) (time.Time, error) {
	if s == "" {
		return time.Now().UTC().Truncate(24 * time.Hour), nil
	}
	return time.Parse(dateLayout, s)
}
