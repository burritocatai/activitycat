package daterange

import (
	"fmt"
	"time"
)

// Range represents a date range with start and end times
type Range struct {
	Start time.Time
	End   time.Time
}

// LastWeek returns a range covering the last 7 days
func LastWeek() Range {
	now := time.Now()
	return Range{
		Start: now.AddDate(0, 0, -7),
		End:   now,
	}
}

// LastMonth returns a range covering the last 30 days
func LastMonth() Range {
	now := time.Now()
	return Range{
		Start: now.AddDate(0, 0, -30),
		End:   now,
	}
}

// Last3Months returns a range covering the last 90 days
func Last3Months() Range {
	now := time.Now()
	return Range{
		Start: now.AddDate(0, 0, -90),
		End:   now,
	}
}

// Parse parses a date string in YYYY-MM-DD format
func Parse(s string) (time.Time, error) {
	return time.Parse("2006-01-02", s)
}

// Custom creates a Range from two date strings in YYYY-MM-DD format
func Custom(startStr, endStr string) (Range, error) {
	start, err := Parse(startStr)
	if err != nil {
		return Range{}, fmt.Errorf("invalid start date: %w", err)
	}
	end, err := Parse(endStr)
	if err != nil {
		return Range{}, fmt.Errorf("invalid end date: %w", err)
	}
	if end.Before(start) {
		return Range{}, fmt.Errorf("end date must not be before start date")
	}
	return Range{Start: start, End: end}, nil
}

// GitHubQueryString formats the range for GitHub search queries
// Returns a string like ">=YYYY-MM-DD" for the start date
func (r Range) GitHubQueryString() string {
	return ">=" + r.Start.Format("2006-01-02")
}

// String returns a human-readable representation of the range
func (r Range) String() string {
	return r.Start.Format("2006-01-02") + " to " + r.End.Format("2006-01-02")
}
