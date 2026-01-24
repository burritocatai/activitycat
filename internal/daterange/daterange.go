package daterange

import (
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

// GitHubQueryString formats the range for GitHub search queries
// Returns a string like ">=YYYY-MM-DD" for the start date
func (r Range) GitHubQueryString() string {
	return ">=" + r.Start.Format("2006-01-02")
}

// String returns a human-readable representation of the range
func (r Range) String() string {
	return r.Start.Format("2006-01-02") + " to " + r.End.Format("2006-01-02")
}
