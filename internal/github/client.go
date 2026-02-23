package github

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"sync"

	"github.com/burritocatai/activitycat/internal/daterange"
	tea "github.com/charmbracelet/bubbletea"
)

// CheckAuth verifies that gh CLI is installed and authenticated
func CheckAuth() error {
	// Check if gh is installed
	if _, err := exec.LookPath("gh"); err != nil {
		return fmt.Errorf("gh CLI not found. Please install it from https://cli.github.com/")
	}

	// Check if authenticated
	cmd := exec.Command("gh", "auth", "status")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("gh CLI not authenticated. Run 'gh auth login' first")
	}

	return nil
}

// FetchPRs executes gh search prs to fetch PRs for the authenticated user
func FetchPRs(ctx context.Context, dateRange daterange.Range) ([]PullRequest, error) {
	args := []string{
		"search", "prs",
		"--author", "@me",
		"--created", dateRange.GitHubQueryString(),
		"--json", "number,title,state,body,createdAt,closedAt,author,repository",
		"--limit", "1000",
	}

	cmd := exec.CommandContext(ctx, "gh", args...)
	output, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return nil, fmt.Errorf("gh command failed: %s", string(exitErr.Stderr))
		}
		return nil, fmt.Errorf("gh command failed: %w", err)
	}

	var prs []PullRequest
	if err := json.Unmarshal(output, &prs); err != nil {
		return nil, fmt.Errorf("failed to parse PR data: %w", err)
	}

	return prs, nil
}

// FetchIssues executes gh search issues to fetch closed issues for the authenticated user
func FetchIssues(ctx context.Context, dateRange daterange.Range) ([]Issue, error) {
	args := []string{
		"search", "issues",
		"--author", "@me",
		"--closed", dateRange.GitHubQueryString(),
		"--json", "number,title,state,body,createdAt,closedAt,author,repository",
		"--limit", "1000",
	}

	cmd := exec.CommandContext(ctx, "gh", args...)
	output, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return nil, fmt.Errorf("gh command failed: %s", string(exitErr.Stderr))
		}
		return nil, fmt.Errorf("gh command failed: %w", err)
	}

	var issues []Issue
	if err := json.Unmarshal(output, &issues); err != nil {
		return nil, fmt.Errorf("failed to parse issue data: %w", err)
	}

	return issues, nil
}

// FetchReviews fetches PRs the user reviewed
func FetchReviews(ctx context.Context, dateRange daterange.Range) ([]Review, error) {
	args := []string{
		"search", "prs",
		"--reviewed-by", "@me",
		"--created", dateRange.GitHubQueryString(),
		"--json", "number,title,state,author,repository,createdAt,closedAt",
		"--limit", "1000",
	}

	cmd := exec.CommandContext(ctx, "gh", args...)
	output, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return nil, fmt.Errorf("gh command failed: %s", string(exitErr.Stderr))
		}
		return nil, fmt.Errorf("gh command failed: %w", err)
	}

	var reviews []Review
	if err := json.Unmarshal(output, &reviews); err != nil {
		return nil, fmt.Errorf("failed to parse review data: %w", err)
	}

	return reviews, nil
}

// FetchCommits fetches commits authored by the user
func FetchCommits(ctx context.Context, dateRange daterange.Range) ([]Commit, error) {
	args := []string{
		"search", "commits",
		"--author", "@me",
		"--author-date", dateRange.GitHubQueryString(),
		"--json", "sha,commit,repository",
		"--limit", "1000",
	}

	cmd := exec.CommandContext(ctx, "gh", args...)
	output, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return nil, fmt.Errorf("gh command failed: %s", string(exitErr.Stderr))
		}
		return nil, fmt.Errorf("gh command failed: %w", err)
	}

	var commits []Commit
	if err := json.Unmarshal(output, &commits); err != nil {
		return nil, fmt.Errorf("failed to parse commit data: %w", err)
	}

	return commits, nil
}

// FetchCommentedPRs fetches PRs the user commented on
func FetchCommentedPRs(ctx context.Context, dateRange daterange.Range) ([]CommentedItem, error) {
	args := []string{
		"search", "prs",
		"--commenter", "@me",
		"--created", dateRange.GitHubQueryString(),
		"--json", "number,title,state,author,repository,commentsCount",
		"--limit", "1000",
	}

	cmd := exec.CommandContext(ctx, "gh", args...)
	output, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return nil, fmt.Errorf("gh command failed: %s", string(exitErr.Stderr))
		}
		return nil, fmt.Errorf("gh command failed: %w", err)
	}

	var items []CommentedItem
	if err := json.Unmarshal(output, &items); err != nil {
		return nil, fmt.Errorf("failed to parse commented PR data: %w", err)
	}

	for i := range items {
		items[i].IsPR = true
	}

	return items, nil
}

// FetchCommentedIssues fetches issues the user commented on
func FetchCommentedIssues(ctx context.Context, dateRange daterange.Range) ([]CommentedItem, error) {
	args := []string{
		"search", "issues",
		"--commenter", "@me",
		"--created", dateRange.GitHubQueryString(),
		"--json", "number,title,state,author,repository,commentsCount",
		"--limit", "1000",
	}

	cmd := exec.CommandContext(ctx, "gh", args...)
	output, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return nil, fmt.Errorf("gh command failed: %s", string(exitErr.Stderr))
		}
		return nil, fmt.Errorf("gh command failed: %w", err)
	}

	var items []CommentedItem
	if err := json.Unmarshal(output, &items); err != nil {
		return nil, fmt.Errorf("failed to parse commented issue data: %w", err)
	}

	for i := range items {
		items[i].IsPR = false
	}

	return items, nil
}

// ActivityLoadedMsg is sent when all activity data is loaded
type ActivityLoadedMsg struct {
	PRs            []PullRequest
	Issues         []Issue
	Reviews        []Review
	Commits        []Commit
	CommentedItems []CommentedItem
	Error          error
}

// FetchActivityCmd runs all fetch functions concurrently and returns an ActivityLoadedMsg
func FetchActivityCmd(dateRange daterange.Range) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()

		var (
			prs            []PullRequest
			issues         []Issue
			reviews        []Review
			commits        []Commit
			commentedPRs   []CommentedItem
			commentedIssue []CommentedItem
			prErr, issueErr, reviewErr, commitErr, commentPRErr, commentIssueErr error
			wg sync.WaitGroup
		)

		wg.Add(6)

		go func() {
			defer wg.Done()
			prs, prErr = FetchPRs(ctx, dateRange)
		}()

		go func() {
			defer wg.Done()
			issues, issueErr = FetchIssues(ctx, dateRange)
		}()

		go func() {
			defer wg.Done()
			reviews, reviewErr = FetchReviews(ctx, dateRange)
		}()

		go func() {
			defer wg.Done()
			commits, commitErr = FetchCommits(ctx, dateRange)
		}()

		go func() {
			defer wg.Done()
			commentedPRs, commentPRErr = FetchCommentedPRs(ctx, dateRange)
		}()

		go func() {
			defer wg.Done()
			commentedIssue, commentIssueErr = FetchCommentedIssues(ctx, dateRange)
		}()

		wg.Wait()

		// Return the first error encountered
		for _, err := range []error{prErr, issueErr, reviewErr, commitErr, commentPRErr, commentIssueErr} {
			if err != nil {
				return ActivityLoadedMsg{Error: err}
			}
		}

		// Merge commented PRs and issues
		commented := append(commentedPRs, commentedIssue...)

		return ActivityLoadedMsg{
			PRs:            prs,
			Issues:         issues,
			Reviews:        reviews,
			Commits:        commits,
			CommentedItems: commented,
		}
	}
}

// FormatActivityForClaude converts all activity data to a text format suitable for Claude API
func FormatActivityForClaude(prs []PullRequest, issues []Issue, reviews []Review, commits []Commit, commentedItems []CommentedItem, metricsText string) string {
	var sb strings.Builder

	sb.WriteString("# GitHub Activity Report\n\n")

	// Metrics summary at top
	if metricsText != "" {
		sb.WriteString("## Metrics Summary\n\n")
		sb.WriteString(metricsText)
		sb.WriteString("\n\n")
	}

	sb.WriteString(fmt.Sprintf("Total PRs: %d\n", len(prs)))
	sb.WriteString(fmt.Sprintf("Total Closed Issues: %d\n", len(issues)))
	sb.WriteString(fmt.Sprintf("Total Reviews Given: %d\n", len(reviews)))
	sb.WriteString(fmt.Sprintf("Total Commits: %d\n", len(commits)))
	sb.WriteString(fmt.Sprintf("Total Items Commented On: %d\n\n", len(commentedItems)))

	// Format PRs
	if len(prs) > 0 {
		sb.WriteString("## Pull Requests\n\n")
		for i, pr := range prs {
			sb.WriteString(fmt.Sprintf("### PR #%d: %s\n", i+1, pr.Title))
			sb.WriteString(fmt.Sprintf("- Repository: %s\n", pr.Repository.NameWithOwner))
			sb.WriteString(fmt.Sprintf("- Author: @%s\n", pr.Author.Login))
			sb.WriteString(fmt.Sprintf("- State: %s\n", pr.State))
			sb.WriteString(fmt.Sprintf("- Created: %s\n", pr.CreatedAt.Format("2006-01-02")))

			if mt := pr.MergeTime(); mt != nil {
				sb.WriteString(fmt.Sprintf("- Merged: %s\n", mt.Format("2006-01-02")))
			} else if pr.ClosedAt != nil {
				sb.WriteString(fmt.Sprintf("- Closed: %s\n", pr.ClosedAt.Format("2006-01-02")))
			}

			reviewers := pr.Reviewers()
			if len(reviewers) > 0 {
				sb.WriteString(fmt.Sprintf("- Reviewers: %s\n", strings.Join(reviewers, ", ")))
			}

			if pr.Body != "" {
				body := pr.Body
				if len(body) > 500 {
					body = body[:500] + "..."
				}
				sb.WriteString(fmt.Sprintf("- Description: %s\n", body))
			}

			sb.WriteString("\n")
		}
	}

	// Format Issues
	if len(issues) > 0 {
		sb.WriteString("## Closed Issues\n\n")
		for i, issue := range issues {
			sb.WriteString(fmt.Sprintf("### Issue #%d: %s\n", i+1, issue.Title))
			sb.WriteString(fmt.Sprintf("- Repository: %s\n", issue.Repository.NameWithOwner))
			sb.WriteString(fmt.Sprintf("- Author: @%s\n", issue.Author.Login))
			sb.WriteString(fmt.Sprintf("- State: %s\n", issue.State))
			sb.WriteString(fmt.Sprintf("- Created: %s\n", issue.CreatedAt.Format("2006-01-02")))

			if issue.ClosedAt != nil {
				sb.WriteString(fmt.Sprintf("- Closed: %s\n", issue.ClosedAt.Format("2006-01-02")))
			}

			if issue.Body != "" {
				body := issue.Body
				if len(body) > 500 {
					body = body[:500] + "..."
				}
				sb.WriteString(fmt.Sprintf("- Description: %s\n", body))
			}

			sb.WriteString("\n")
		}
	}

	// Format Reviews
	if len(reviews) > 0 {
		sb.WriteString("## Code Reviews Given\n\n")
		for i, r := range reviews {
			sb.WriteString(fmt.Sprintf("### Review #%d: %s\n", i+1, r.Title))
			sb.WriteString(fmt.Sprintf("- Repository: %s\n", r.Repository.NameWithOwner))
			sb.WriteString(fmt.Sprintf("- PR Author: @%s\n", r.Author.Login))
			sb.WriteString(fmt.Sprintf("- State: %s\n", r.State))
			sb.WriteString(fmt.Sprintf("- Created: %s\n", r.CreatedAt.Format("2006-01-02")))
			sb.WriteString("\n")
		}
	}

	// Format Commits
	if len(commits) > 0 {
		sb.WriteString("## Commits\n\n")
		for _, c := range commits {
			msg := c.Commit.Message
			// Use first line only
			if idx := strings.Index(msg, "\n"); idx != -1 {
				msg = msg[:idx]
			}
			if len(msg) > 100 {
				msg = msg[:100] + "..."
			}
			sb.WriteString(fmt.Sprintf("- %s %s (%s, %s)\n",
				c.SHA[:min(7, len(c.SHA))],
				msg,
				c.Repository.FullName,
				c.Commit.Author.Date.Format("2006-01-02"),
			))
		}
		sb.WriteString("\n")
	}

	// Format Commented Items
	if len(commentedItems) > 0 {
		sb.WriteString("## Items Commented On\n\n")
		for _, item := range commentedItems {
			kind := "Issue"
			if item.IsPR {
				kind = "PR"
			}
			sb.WriteString(fmt.Sprintf("- [%s] %s (%s, %d comments)\n",
				kind, item.Title, item.Repository.NameWithOwner, item.Comments))
		}
		sb.WriteString("\n")
	}

	return sb.String()
}
