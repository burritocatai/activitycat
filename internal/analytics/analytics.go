package analytics

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"github.com/burritocatai/activitycat/internal/daterange"
	"github.com/burritocatai/activitycat/internal/github"
)

// Metrics holds computed analytics for a date range
type Metrics struct {
	// PR lifecycle
	PRsOpened    int
	PRsMerged    int
	PRsClosed    int // closed without merge
	PRsOpen      int
	MergeRate    float64       // percentage of closed PRs that were merged
	AvgMergeTime time.Duration // average time from creation to merge
	MedMergeTime time.Duration // median time from creation to merge

	// Totals
	TotalCommits        int
	TotalReviews        int
	TotalCommentedItems int
	TotalIssuesClosed   int

	// Rates
	PRsPerWeek     float64
	CommitsPerDay  float64
	ReviewsPerWeek float64

	// Repo breakdown
	RepoStats []RepoStats

	// Most active day
	MostActiveDay   string
	MostActiveCount int

	// Date range for rate calculations
	days float64
}

// RepoStats groups all activity types for a single repository
type RepoStats struct {
	Repo           string
	PRs            int
	Issues         int
	Reviews        int
	Commits        int
	CommentedItems int
	Total          int
}

// Compute calculates metrics from all fetched data
func Compute(
	prs []github.PullRequest,
	issues []github.Issue,
	reviews []github.Review,
	commits []github.Commit,
	commentedItems []github.CommentedItem,
	dr daterange.Range,
) *Metrics {
	m := &Metrics{}

	days := dr.End.Sub(dr.Start).Hours() / 24
	if days < 1 {
		days = 1
	}
	m.days = days
	weeks := days / 7
	if weeks < 1 {
		weeks = 1
	}

	// PR lifecycle
	var mergeTimes []time.Duration
	for _, pr := range prs {
		switch {
		case pr.IsMerged():
			m.PRsMerged++
			if mt := pr.MergeTime(); mt != nil {
				mergeTimes = append(mergeTimes, mt.Sub(pr.CreatedAt))
			}
		case pr.IsOpen():
			m.PRsOpen++
		case pr.IsClosed():
			m.PRsClosed++
		default:
			m.PRsOpened++ // fallback
		}
	}
	m.PRsOpened = len(prs)

	resolved := m.PRsMerged + m.PRsClosed
	if resolved > 0 {
		m.MergeRate = float64(m.PRsMerged) / float64(resolved) * 100
	}

	if len(mergeTimes) > 0 {
		sort.Slice(mergeTimes, func(i, j int) bool { return mergeTimes[i] < mergeTimes[j] })

		var total time.Duration
		for _, d := range mergeTimes {
			total += d
		}
		m.AvgMergeTime = total / time.Duration(len(mergeTimes))
		m.MedMergeTime = mergeTimes[len(mergeTimes)/2]
	}

	// Totals
	m.TotalCommits = len(commits)
	m.TotalReviews = len(reviews)
	m.TotalCommentedItems = len(commentedItems)
	m.TotalIssuesClosed = len(issues)

	// Rates
	m.PRsPerWeek = float64(len(prs)) / weeks
	m.CommitsPerDay = float64(len(commits)) / days
	m.ReviewsPerWeek = float64(len(reviews)) / weeks

	// Repo breakdown
	repoMap := make(map[string]*RepoStats)
	getRepo := func(name string) *RepoStats {
		if rs, ok := repoMap[name]; ok {
			return rs
		}
		rs := &RepoStats{Repo: name}
		repoMap[name] = rs
		return rs
	}

	for _, pr := range prs {
		getRepo(pr.Repository.NameWithOwner).PRs++
	}
	for _, issue := range issues {
		getRepo(issue.Repository.NameWithOwner).Issues++
	}
	for _, r := range reviews {
		getRepo(r.Repository.NameWithOwner).Reviews++
	}
	for _, c := range commits {
		getRepo(c.Repository.FullName).Commits++
	}
	for _, ci := range commentedItems {
		getRepo(ci.Repository.NameWithOwner).CommentedItems++
	}

	for _, rs := range repoMap {
		rs.Total = rs.PRs + rs.Issues + rs.Reviews + rs.Commits + rs.CommentedItems
		m.RepoStats = append(m.RepoStats, *rs)
	}
	sort.Slice(m.RepoStats, func(i, j int) bool {
		return m.RepoStats[i].Total > m.RepoStats[j].Total
	})

	// Most active day
	dayCount := make(map[string]int)
	addDay := func(t time.Time) {
		key := t.Format("2006-01-02")
		dayCount[key]++
	}
	for _, pr := range prs {
		addDay(pr.CreatedAt)
	}
	for _, issue := range issues {
		if issue.ClosedAt != nil {
			addDay(*issue.ClosedAt)
		}
	}
	for _, r := range reviews {
		addDay(r.CreatedAt)
	}
	for _, c := range commits {
		addDay(c.Commit.Author.Date)
	}

	for day, count := range dayCount {
		if count > m.MostActiveCount {
			m.MostActiveDay = day
			m.MostActiveCount = count
		}
	}

	return m
}

// Format returns a human-readable summary of the metrics
func (m *Metrics) Format() string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("PRs: %d opened, %d merged, %d closed, %d open",
		m.PRsOpened, m.PRsMerged, m.PRsClosed, m.PRsOpen))

	if m.PRsMerged+m.PRsClosed > 0 {
		sb.WriteString(fmt.Sprintf("  |  Merge rate: %.0f%%", m.MergeRate))
	}
	sb.WriteString("\n")

	if m.AvgMergeTime > 0 {
		sb.WriteString(fmt.Sprintf("Time to merge: avg %s, median %s\n",
			formatDuration(m.AvgMergeTime), formatDuration(m.MedMergeTime)))
	}

	sb.WriteString(fmt.Sprintf("Commits: %d  |  Reviews: %d  |  Issues closed: %d  |  Commented on: %d\n",
		m.TotalCommits, m.TotalReviews, m.TotalIssuesClosed, m.TotalCommentedItems))

	sb.WriteString(fmt.Sprintf("Rates: %.1f PRs/wk, %.1f commits/day, %.1f reviews/wk\n",
		m.PRsPerWeek, m.CommitsPerDay, m.ReviewsPerWeek))

	if m.MostActiveDay != "" {
		sb.WriteString(fmt.Sprintf("Most active day: %s (%d activities)", m.MostActiveDay, m.MostActiveCount))
	}

	return sb.String()
}

func formatDuration(d time.Duration) string {
	hours := d.Hours()
	if hours < 1 {
		return fmt.Sprintf("%dm", int(d.Minutes()))
	}
	if hours < 24 {
		return fmt.Sprintf("%.1fh", hours)
	}
	days := math.Round(hours / 24)
	return fmt.Sprintf("%.0fd", days)
}
