package prlist

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/burritocatai/activitycat/internal/github"
	"github.com/burritocatai/activitycat/internal/ui/styles"
)

// Model represents the PR list display screen
type Model struct {
	viewport viewport.Model
	prs      []github.PullRequest
	issues   []github.Issue
	width    int
	height   int
	ready    bool
}

// New creates a new PR list model
func New(prs []github.PullRequest, issues []github.Issue, width, height int) Model {
	m := Model{
		prs:    prs,
		issues: issues,
		width:  width,
		height: height,
		ready:  false,
	}

	// Initialize viewport immediately if dimensions are provided
	if width > 0 && height > 0 {
		m.viewport = viewport.New(width, height-5)
		m.viewport.SetContent(m.renderContent())
		m.ready = true
	}

	return m
}

// Init initializes the PR list model
func (m Model) Init() tea.Cmd {
	return nil
}

// Update handles messages for the PR list screen
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

// View renders the PR list screen
func (m Model) View() string {
	if !m.ready {
		return "\n  Initializing..."
	}

	header := styles.TitleStyle.Render(fmt.Sprintf("Found %d Pull Requests and %d Closed Issues", len(m.prs), len(m.issues)))
	footer := styles.FooterStyle.Render("↑/↓: Scroll • Enter: Continue • b: Back • q: Quit")

	return lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		m.viewport.View(),
		footer,
	)
}

// renderContent formats all PRs and issues for display
func (m Model) renderContent() string {
	if len(m.prs) == 0 && len(m.issues) == 0 {
		return styles.SubtleStyle.Render("No pull requests or closed issues found in this date range.")
	}

	var content strings.Builder

	// Display PRs first
	if len(m.prs) > 0 {
		content.WriteString(lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("4")).Render("Pull Requests"))
		content.WriteString("\n\n")
		for _, pr := range m.prs {
			content.WriteString(m.formatPR(pr))
			content.WriteString("\n")
		}
	}

	// Display issues after PRs
	if len(m.issues) > 0 {
		if len(m.prs) > 0 {
			content.WriteString("\n")
		}
		content.WriteString(lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("2")).Render("Closed Issues"))
		content.WriteString("\n\n")
		for _, issue := range m.issues {
			content.WriteString(m.formatIssue(issue))
			content.WriteString("\n")
		}
	}

	return content.String()
}

// formatPR formats a single PR as a card
func (m Model) formatPR(pr github.PullRequest) string {
	// Title with state color
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

	// Repository and author
	repo := styles.SubtleStyle.Render(pr.Repository.NameWithOwner)
	author := styles.SubtleStyle.Render(fmt.Sprintf("@%s", pr.Author.Login))

	// Dates
	dates := fmt.Sprintf("Created: %s", pr.CreatedAt.Format("2006-01-02"))
	if pr.MergedAt != nil {
		dates += fmt.Sprintf(" • Merged: %s", pr.MergedAt.Format("2006-01-02"))
	} else if pr.ClosedAt != nil {
		dates += fmt.Sprintf(" • Closed: %s", pr.ClosedAt.Format("2006-01-02"))
	}

	// Reviewers
	reviewers := ""
	if len(pr.ReviewRequests) > 0 {
		reviewerLogins := pr.Reviewers()
		reviewers = fmt.Sprintf("Reviewers: %s", strings.Join(reviewerLogins, ", "))
	}

	// Body (truncated)
	body := ""
	if pr.Body != "" {
		truncated := pr.Body
		if len(truncated) > 200 {
			truncated = truncated[:200] + "..."
		}
		// Remove newlines for compact display
		truncated = strings.ReplaceAll(truncated, "\n", " ")
		body = styles.SubtleStyle.Render(truncated)
	}

	// Build card content
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

// formatIssue formats a single issue as a card
func (m Model) formatIssue(issue github.Issue) string {
	// Title with state color
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

	// Repository and author
	repo := styles.SubtleStyle.Render(issue.Repository.NameWithOwner)
	author := styles.SubtleStyle.Render(fmt.Sprintf("@%s", issue.Author.Login))

	// Dates
	dates := fmt.Sprintf("Created: %s", issue.CreatedAt.Format("2006-01-02"))
	if issue.ClosedAt != nil {
		dates += fmt.Sprintf(" • Closed: %s", issue.ClosedAt.Format("2006-01-02"))
	}

	// Body (truncated)
	body := ""
	if issue.Body != "" {
		truncated := issue.Body
		if len(truncated) > 200 {
			truncated = truncated[:200] + "..."
		}
		// Remove newlines for compact display
		truncated = strings.ReplaceAll(truncated, "\n", " ")
		body = styles.SubtleStyle.Render(truncated)
	}

	// Build card content
	var card strings.Builder
	card.WriteString(fmt.Sprintf("%s %s\n", title, state))
	card.WriteString(fmt.Sprintf("%s • %s\n", repo, author))
	card.WriteString(fmt.Sprintf("%s\n", dates))
	if body != "" {
		card.WriteString(fmt.Sprintf("\n%s\n", body))
	}

	return styles.PRCardStyle.Render(card.String())
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
