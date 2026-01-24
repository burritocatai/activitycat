package promptselect

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/burritocatai/activitycat/internal/config"
	"github.com/burritocatai/activitycat/internal/ui/styles"
)

// Model represents the prompt selection screen
type Model struct {
	prompts  []config.Prompt
	cursor   int
	selected bool
}

// New creates a new prompt selection model
func New(prompts []config.Prompt) Model {
	return Model{
		prompts:  prompts,
		cursor:   0,
		selected: false,
	}
}

// Init initializes the prompt selection model
func (m Model) Init() tea.Cmd {
	return nil
}

// Update handles messages for the prompt selection screen
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.prompts)-1 {
				m.cursor++
			}
		case "enter":
			m.selected = true
			return m, func() tea.Msg {
				return PromptSelectedMsg{
					Prompt: m.prompts[m.cursor],
				}
			}
		case "b":
			return m, func() tea.Msg {
				return BackMsg{}
			}
		case "q", "ctrl+c":
			return m, tea.Quit
		}
	}

	return m, nil
}

// View renders the prompt selection screen
func (m Model) View() string {
	s := styles.TitleStyle.Render("Select Report Prompt")
	s += "\n\n"

	for i, prompt := range m.prompts {
		cursor := "  "
		if m.cursor == i {
			cursor = "> "
			s += styles.SelectedStyle.Render(cursor + prompt.Name)
		} else {
			s += styles.UnselectedStyle.Render(cursor + prompt.Name)
		}
		s += "\n"
	}

	s += "\n" + styles.FooterStyle.Render("↑/↓: Navigate • Enter: Select • b: Back • q: Quit")

	return s
}

// PromptSelectedMsg is sent when a prompt is selected
type PromptSelectedMsg struct {
	Prompt config.Prompt
}

// BackMsg is sent when the user wants to go back
type BackMsg struct{}
