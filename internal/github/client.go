package github

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"

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
	// Build the command
	args := []string{
		"search", "prs",
		"--author", "@me",
		"--created", dateRange.GitHubQueryString(),
		"--json", "number,title,state,body,createdAt,closedAt,author,repository",
		"--limit", "1000",
	}

	cmd := exec.CommandContext(ctx, "gh", args...)

	// Execute and capture output
	output, err := cmd.Output()
	if err != nil {
		// Try to get stderr for more context
		if exitErr, ok := err.(*exec.ExitError); ok {
			stderr := string(exitErr.Stderr)
			return nil, fmt.Errorf("gh command failed: %s", stderr)
		}
		return nil, fmt.Errorf("gh command failed: %w", err)
	}

	// Parse JSON output
	var prs []PullRequest
	if err := json.Unmarshal(output, &prs); err != nil {
		return nil, fmt.Errorf("failed to parse PR data: %w", err)
	}

	return prs, nil
}

// FetchPRsCmd wraps FetchPRs in a Bubbletea Cmd
func FetchPRsCmd(dateRange daterange.Range) tea.Cmd {
	return func() tea.Msg {
		prs, err := FetchPRs(context.Background(), dateRange)
		return PRsLoadedMsg{
			PRs:   prs,
			Error: err,
		}
	}
}

// PRsLoadedMsg is sent when PRs are loaded
type PRsLoadedMsg struct {
	PRs   []PullRequest
	Error error
}

// FormatPRsForClaude converts PR list to a text format suitable for Claude API
func FormatPRsForClaude(prs []PullRequest) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("# GitHub Pull Request Activity Report\n\n"))
	sb.WriteString(fmt.Sprintf("Total PRs: %d\n\n", len(prs)))

	for i, pr := range prs {
		sb.WriteString(fmt.Sprintf("## PR #%d: %s\n", i+1, pr.Title))
		sb.WriteString(fmt.Sprintf("- Repository: %s\n", pr.Repository.NameWithOwner))
		sb.WriteString(fmt.Sprintf("- Author: @%s\n", pr.Author.Login))
		sb.WriteString(fmt.Sprintf("- State: %s\n", pr.State))
		sb.WriteString(fmt.Sprintf("- Created: %s\n", pr.CreatedAt.Format("2006-01-02")))

		if pr.MergedAt != nil {
			sb.WriteString(fmt.Sprintf("- Merged: %s\n", pr.MergedAt.Format("2006-01-02")))
		} else if pr.ClosedAt != nil {
			sb.WriteString(fmt.Sprintf("- Closed: %s\n", pr.ClosedAt.Format("2006-01-02")))
		}

		reviewers := pr.Reviewers()
		if len(reviewers) > 0 {
			sb.WriteString(fmt.Sprintf("- Reviewers: %s\n", strings.Join(reviewers, ", ")))
		}

		if pr.Body != "" {
			// Truncate body if too long
			body := pr.Body
			if len(body) > 500 {
				body = body[:500] + "..."
			}
			sb.WriteString(fmt.Sprintf("- Description: %s\n", body))
		}

		sb.WriteString("\n")
	}

	return sb.String()
}
