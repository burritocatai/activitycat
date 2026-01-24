package loading

import (
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Model represents the loading screen with a spinner
type Model struct {
	spinner spinner.Model
	message string
}

// New creates a new loading model
func New(message string) Model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	return Model{
		spinner: s,
		message: message,
	}
}

// Init initializes the loading model
func (m Model) Init() tea.Cmd {
	return m.spinner.Tick
}

// Update handles messages for the loading screen
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	var cmd tea.Cmd
	m.spinner, cmd = m.spinner.Update(msg)
	return m, cmd
}

// View renders the loading screen
func (m Model) View() string {
	return lipgloss.JoinVertical(
		lipgloss.Center,
		"",
		m.spinner.View()+" "+m.message,
		"",
	)
}

// SetMessage updates the loading message
func (m *Model) SetMessage(message string) {
	m.message = message
}
