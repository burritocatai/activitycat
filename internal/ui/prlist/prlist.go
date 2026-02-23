package prlist

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/burritocatai/activitycat/internal/analytics"
	"github.com/burritocatai/activitycat/internal/github"
	"github.com/burritocatai/activitycat/internal/ui/styles"
)

// Model represents the activity list display screen
type Model struct {
	viewport       viewport.Model
	prs            []github.PullRequest
	issues         []github.Issue
	reviews        []github.Review
	commits        []github.Commit
	commentedItems []github.CommentedItem
	metrics        *analytics.Metrics
	width          int
	height         int
	ready          bool
}

// New creates a new activity list model
func New(
	prs []github.PullRequest,
	issues []github.Issue,
	reviews []github.Review,
	commits []github.Commit,
	commentedItems []github.CommentedItem,
	metrics *analytics.Metrics,
	width, height int,
) Model {
	m := Model{
		prs:            prs,
		issues:         issues,
		reviews:        reviews,
		commits:        commits,
		commentedItems: commentedItems,
		metrics:        metrics,
		width:          width,
		height:         height,
		ready:          false,
	}

	if width > 0 && height > 0 {
		m.viewport = viewport.New(width, height-5)
		m.viewport.SetContent(m.renderContent())
		m.ready = true
	}

	return m
}

// Init initializes the model
func (m Model) Init() tea.Cmd {
	return nil
}

// Update handles messages
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "b":
			return m, func() tea.Msg {
				return BackMsg{}
			}
		case "enter":
			return m, func() tea.Msg {
				return ContinueMsg{}
			}
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		if !m.ready {
			m.viewport = viewport.New(msg.Width, msg.Height-5)
			m.viewport.SetContent(m.renderContent())
			m.ready = true
		} else {
			m.viewport.Width = msg.Width
			m.viewport.Height = msg.Height - 5
			m.viewport.SetContent(m.renderContent())
		}
	}

	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

// View renders the screen
func (m Model) View() string {
	if !m.ready {
		return "\n  Initializing..."
	}

	header := styles.TitleStyle.Render(fmt.Sprintf(
		"Activity: %d PRs, %d Issues, %d Reviews, %d Commits",
		len(m.prs), len(m.issues), len(m.reviews), len(m.commits),
	))
	footer := styles.FooterStyle.Render("↑/↓: Scroll • Enter: Continue • b: Back • q: Quit")

	return lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		m.viewport.View(),
		footer,
	)
}

// renderContent formats all activity for display
func (m Model) renderContent() string {
	if len(m.prs) == 0 && len(m.issues) == 0 && len(m.reviews) == 0 &&
		len(m.commits) == 0 && len(m.commentedItems) == 0 {
		return styles.SubtleStyle.Render("No activity found in this date range.")
	}

	var content strings.Builder

	// 1. Analytics summary box
	if m.metrics != nil {
		content.WriteString(styles.MetricsBoxStyle.Render(m.metrics.Format()))
		content.WriteString("\n\n")
	}

	// 2. Repo breakdown table
	if m.metrics != nil && len(m.metrics.RepoStats) > 0 {
		content.WriteString(lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205")).Render("Repository Breakdown"))
		content.WriteString("\n\n")
		for _, rs := range m.metrics.RepoStats {
			line := fmt.Sprintf("  %-40s  PRs:%-3d  Issues:%-3d  Reviews:%-3d  Commits:%-3d  Comments:%-3d",
				rs.Repo, rs.PRs, rs.Issues, rs.Reviews, rs.Commits, rs.CommentedItems)
			content.WriteString(styles.SubtleStyle.Render(line))
			content.WriteString("\n")
		}
		content.WriteString("\n")
	}

	// 3. Pull Requests
	if len(m.prs) > 0 {
		content.WriteString(lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("4")).Render("Pull Requests"))
		content.WriteString("\n\n")
		for _, pr := range m.prs {
			content.WriteString(formatPR(pr))
			content.WriteString("\n")
		}
	}

	// 4. Closed Issues
	if len(m.issues) > 0 {
		if content.Len() > 0 {
			content.WriteString("\n")
		}
		content.WriteString(lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("2")).Render("Closed Issues"))
		content.WriteString("\n\n")
		for _, issue := range m.issues {
			content.WriteString(formatIssue(issue))
			content.WriteString("\n")
		}
	}

	// 5. Code Reviews Given
	if len(m.reviews) > 0 {
		content.WriteString("\n")
		content.WriteString(lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("135")).Render("Code Reviews Given"))
		content.WriteString("\n\n")
		for _, r := range m.reviews {
			content.WriteString(formatReview(r))
			content.WriteString("\n")
		}
	}

	// 6. Commits
	if len(m.commits) > 0 {
		content.WriteString("\n")
		content.WriteString(lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("208")).Render("Commits"))
		content.WriteString("\n\n")
		for _, c := range m.commits {
			content.WriteString(formatCommit(c))
			content.WriteString("\n")
		}
	}

	// 7. Commented Items
	if len(m.commentedItems) > 0 {
		content.WriteString("\n")
		content.WriteString(lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("241")).Render("Commented Items"))
		content.WriteString("\n\n")
		for _, ci := range m.commentedItems {
			content.WriteString(formatCommentedItem(ci))
			content.WriteString("\n")
		}
	}

	return content.String()
}

func formatPR(pr github.PullRequest) string {
	var stateStyle lipgloss.Style
	var stateLabel string

	if pr.IsMerged() {
		stateStyle = styles.MergedStyle
		stateLabel = "MERGED"
	} else if pr.IsOpen() {
		stateStyle = styles.OpenStyle
		stateLabel = "OPEN"
	} else {
		stateStyle = styles.ClosedStyle
		stateLabel = "CLOSED"
	}

	title := lipgloss.NewStyle().Bold(true).Render(pr.Title)
	state := stateStyle.Render(fmt.Sprintf("[%s]", stateLabel))
	repo := styles.SubtleStyle.Render(pr.Repository.NameWithOwner)
	author := styles.SubtleStyle.Render(fmt.Sprintf("@%s", pr.Author.Login))

	dates := fmt.Sprintf("Created: %s", pr.CreatedAt.Format("2006-01-02"))
	if mt := pr.MergeTime(); mt != nil {
		dates += fmt.Sprintf(" • Merged: %s", mt.Format("2006-01-02"))
	} else if pr.ClosedAt != nil {
		dates += fmt.Sprintf(" • Closed: %s", pr.ClosedAt.Format("2006-01-02"))
	}

	reviewers := ""
	if len(pr.ReviewRequests) > 0 {
		reviewers = fmt.Sprintf("Reviewers: %s", strings.Join(pr.Reviewers(), ", "))
	}

	body := ""
	if pr.Body != "" {
		truncated := pr.Body
		if len(truncated) > 200 {
			truncated = truncated[:200] + "..."
		}
		truncated = strings.ReplaceAll(truncated, "\n", " ")
		body = styles.SubtleStyle.Render(truncated)
	}

	var card strings.Builder
	card.WriteString(fmt.Sprintf("%s %s\n", title, state))
	card.WriteString(fmt.Sprintf("%s • %s\n", repo, author))
	card.WriteString(fmt.Sprintf("%s\n", dates))
	if reviewers != "" {
		card.WriteString(fmt.Sprintf("%s\n", reviewers))
	}
	if body != "" {
		card.WriteString(fmt.Sprintf("\n%s\n", body))
	}

	return styles.PRCardStyle.Render(card.String())
}

func formatIssue(issue github.Issue) string {
	var stateStyle lipgloss.Style
	var stateLabel string

	if issue.IsOpen() {
		stateStyle = styles.OpenStyle
		stateLabel = "OPEN"
	} else {
		stateStyle = styles.ClosedStyle
		stateLabel = "CLOSED"
	}

	title := lipgloss.NewStyle().Bold(true).Render(issue.Title)
	state := stateStyle.Render(fmt.Sprintf("[%s]", stateLabel))
	repo := styles.SubtleStyle.Render(issue.Repository.NameWithOwner)
	author := styles.SubtleStyle.Render(fmt.Sprintf("@%s", issue.Author.Login))

	dates := fmt.Sprintf("Created: %s", issue.CreatedAt.Format("2006-01-02"))
	if issue.ClosedAt != nil {
		dates += fmt.Sprintf(" • Closed: %s", issue.ClosedAt.Format("2006-01-02"))
	}

	body := ""
	if issue.Body != "" {
		truncated := issue.Body
		if len(truncated) > 200 {
			truncated = truncated[:200] + "..."
		}
		truncated = strings.ReplaceAll(truncated, "\n", " ")
		body = styles.SubtleStyle.Render(truncated)
	}

	var card strings.Builder
	card.WriteString(fmt.Sprintf("%s %s\n", title, state))
	card.WriteString(fmt.Sprintf("%s • %s\n", repo, author))
	card.WriteString(fmt.Sprintf("%s\n", dates))
	if body != "" {
		card.WriteString(fmt.Sprintf("\n%s\n", body))
	}

	return styles.PRCardStyle.Render(card.String())
}

func formatReview(r github.Review) string {
	state := styles.ReviewStyle.Render(fmt.Sprintf("[%s]", strings.ToUpper(r.State)))
	title := lipgloss.NewStyle().Bold(true).Render(r.Title)
	repo := styles.SubtleStyle.Render(r.Repository.NameWithOwner)
	author := styles.SubtleStyle.Render(fmt.Sprintf("by @%s", r.Author.Login))
	date := r.CreatedAt.Format("2006-01-02")

	var card strings.Builder
	card.WriteString(fmt.Sprintf("%s %s\n", title, state))
	card.WriteString(fmt.Sprintf("%s • %s • %s\n", repo, author, date))

	return styles.PRCardStyle.Render(card.String())
}

func formatCommit(c github.Commit) string {
	sha := styles.CommitStyle.Render(c.SHA[:min(7, len(c.SHA))])
	msg := c.Commit.Message
	if idx := strings.Index(msg, "\n"); idx != -1 {
		msg = msg[:idx]
	}
	if len(msg) > 80 {
		msg = msg[:80] + "..."
	}
	repo := styles.SubtleStyle.Render(c.Repository.FullName)
	date := styles.SubtleStyle.Render(c.Commit.Author.Date.Format("2006-01-02"))

	return fmt.Sprintf("  %s %s  %s  %s", sha, msg, repo, date)
}

func formatCommentedItem(ci github.CommentedItem) string {
	kind := "Issue"
	if ci.IsPR {
		kind = "PR"
	}
	kindLabel := styles.SubtleStyle.Render(fmt.Sprintf("[%s]", kind))
	title := ci.Title
	repo := styles.SubtleStyle.Render(ci.Repository.NameWithOwner)
	comments := styles.SubtleStyle.Render(fmt.Sprintf("%d comments", ci.Comments))

	return fmt.Sprintf("  %s %s  %s  %s", kindLabel, title, repo, comments)
}

// SetSize updates the dimensions
func (m *Model) SetSize(width, height int) {
	m.width = width
	m.height = height
	if m.ready {
		m.viewport.Width = width
		m.viewport.Height = height - 5
		m.viewport.SetContent(m.renderContent())
	}
}

// BackMsg is sent when the user wants to go back
type BackMsg struct{}

// ContinueMsg is sent when the user wants to continue to prompt selection
type ContinueMsg struct{}
