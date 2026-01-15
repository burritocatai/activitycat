package dateselect

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/burritocatai/activitycat/internal/daterange"
	"github.com/burritocatai/activitycat/internal/ui/styles"
)

// Option represents a selectable date range option
type Option struct {
	Label string
	Range daterange.Range
}

// Model represents the date selection screen
type Model struct {
	options  []Option
	cursor   int
	selected bool
}

// New creates a new date selection model
func New() Model {
	return Model{
		options: []Option{
			{Label: "Last Week", Range: daterange.LastWeek()},
			{Label: "Last Month", Range: daterange.LastMonth()},
			{Label: "Last 3 Months", Range: daterange.Last3Months()},
		},
		cursor:   0,
		selected: false,
	}
}

// Init initializes the date selection model
func (m Model) Init() tea.Cmd {
	return nil
}

// Update handles messages for the date selection screen
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.options)-1 {
				m.cursor++
			}
		case "enter":
			m.selected = true
			return m, func() tea.Msg {
				return DateSelectedMsg{
					Range: m.options[m.cursor].Range,
				}
			}
		case "q", "ctrl+c":
			return m, tea.Quit
		}
	}

	return m, nil
}

// View renders the date selection screen
func (m Model) View() string {
	s := styles.TitleStyle.Render("Select Date Range")
	s += "\n\n"

	for i, option := range m.options {
		cursor := "  "
		if m.cursor == i {
			cursor = "> "
			s += styles.SelectedStyle.Render(cursor + option.Label)
		} else {
			s += styles.UnselectedStyle.Render(cursor + option.Label)
		}
		s += "\n"
	}

	s += "\n" + styles.FooterStyle.Render("↑/↓: Navigate • Enter: Select • q: Quit")

	return s
}

// DateSelectedMsg is sent when a date range is selected
type DateSelectedMsg struct {
	Range daterange.Range
}
