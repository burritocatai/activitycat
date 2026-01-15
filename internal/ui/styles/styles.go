package styles

import "github.com/charmbracelet/lipgloss"

var (
	// TitleStyle is used for main titles
	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("205")).
			MarginBottom(1)

	// SelectedStyle is used for selected menu items
	SelectedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("170")).
			Bold(true).
			PaddingLeft(2)

	// UnselectedStyle is used for unselected menu items
	UnselectedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			PaddingLeft(2)

	// MergedStyle is used for merged PRs
	MergedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("42")). // Green
			Bold(true)

	// OpenStyle is used for open PRs
	OpenStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("39")). // Blue
			Bold(true)

	// ClosedStyle is used for closed (not merged) PRs
	ClosedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")). // Red
			Bold(true)

	// PRCardStyle is used for PR cards
	PRCardStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("240")).
			Padding(1, 2).
			MarginBottom(1)

	// ErrorStyle is used for error messages
	ErrorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")).
			Bold(true).
			Padding(1, 2)

	// SubtleStyle is used for less important text
	SubtleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241"))

	// FooterStyle is used for footer instructions
	FooterStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			MarginTop(1)

	// ReportStyle is used for the generated report
	ReportStyle = lipgloss.NewStyle().
			Padding(1, 2)
)
