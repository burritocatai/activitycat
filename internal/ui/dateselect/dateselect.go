package dateselect

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/burritocatai/activitycat/internal/daterange"
	"github.com/burritocatai/activitycat/internal/ui/styles"
)

// Option represents a selectable date range option
type Option struct {
	Label string
	Range daterange.Range
}

// activeField tracks which textinput is focused in custom mode
type activeField int

const (
	fieldStart activeField = iota
	fieldEnd
)

// Model represents the date selection screen
type Model struct {
	options  []Option
	cursor   int
	selected bool

	// Custom range mode
	customMode  bool
	activeField activeField
	startInput  textinput.Model
	endInput    textinput.Model
	inputError  string
}

// New creates a new date selection model
func New() Model {
	si := textinput.New()
	si.Placeholder = "YYYY-MM-DD"
	si.CharLimit = 10
	si.Focus()

	ei := textinput.New()
	ei.Placeholder = "YYYY-MM-DD"
	ei.CharLimit = 10

	return Model{
		options: []Option{
			{Label: "Last Week", Range: daterange.LastWeek()},
			{Label: "Last Month", Range: daterange.LastMonth()},
			{Label: "Last 3 Months", Range: daterange.Last3Months()},
			{Label: "Custom Range..."},
		},
		cursor:     0,
		selected:   false,
		startInput: si,
		endInput:   ei,
	}
}

// Init initializes the date selection model
func (m Model) Init() tea.Cmd {
	return nil
}

// Update handles messages for the date selection screen
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	if m.customMode {
		return m.updateCustom(msg)
	}
	return m.updatePreset(msg)
}

// updatePreset handles the preset menu mode
func (m Model) updatePreset(msg tea.Msg) (Model, tea.Cmd) {
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
			// Last option is "Custom Range..."
			if m.cursor == len(m.options)-1 {
				m.customMode = true
				m.activeField = fieldStart
				m.startInput.SetValue("")
				m.endInput.SetValue("")
				m.startInput.Focus()
				m.endInput.Blur()
				m.inputError = ""
				return m, m.startInput.Cursor.BlinkCmd()
			}
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

// updateCustom handles the custom date input mode
func (m Model) updateCustom(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "esc":
			if m.activeField == fieldEnd {
				m.activeField = fieldStart
				m.startInput.Focus()
				m.endInput.Blur()
				m.inputError = ""
				return m, m.startInput.Cursor.BlinkCmd()
			}
			// Esc on start field → back to preset menu
			m.customMode = false
			m.inputError = ""
			return m, nil
		case "enter":
			if m.activeField == fieldStart {
				// Validate start date before advancing
				_, err := daterange.Parse(m.startInput.Value())
				if err != nil {
					m.inputError = "Invalid start date (use YYYY-MM-DD)"
					return m, nil
				}
				m.activeField = fieldEnd
				m.startInput.Blur()
				m.endInput.Focus()
				m.inputError = ""
				return m, m.endInput.Cursor.BlinkCmd()
			}
			// On end field: validate both and emit
			r, err := daterange.Custom(m.startInput.Value(), m.endInput.Value())
			if err != nil {
				m.inputError = err.Error()
				return m, nil
			}
			m.selected = true
			m.customMode = false
			return m, func() tea.Msg {
				return DateSelectedMsg{Range: r}
			}
		}
	}

	// Delegate to active textinput
	var cmd tea.Cmd
	if m.activeField == fieldStart {
		m.startInput, cmd = m.startInput.Update(msg)
	} else {
		m.endInput, cmd = m.endInput.Update(msg)
	}
	return m, cmd
}

// View renders the date selection screen
func (m Model) View() string {
	if m.customMode {
		return m.viewCustom()
	}
	return m.viewPreset()
}

// viewPreset renders the preset menu
func (m Model) viewPreset() string {
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

// viewCustom renders the custom date input form
func (m Model) viewCustom() string {
	s := styles.TitleStyle.Render("Enter Custom Date Range")
	s += "\n\n"

	startLabel := "  Start date: "
	endLabel := "  End date:   "

	if m.activeField == fieldStart {
		s += styles.SelectedStyle.Render("> Start date: ") + m.startInput.View() + "\n"
		s += styles.UnselectedStyle.Render(endLabel+m.endInput.View()) + "\n"
	} else {
		s += styles.UnselectedStyle.Render(startLabel+m.startInput.View()) + "\n"
		s += styles.SelectedStyle.Render("> End date:   ") + m.endInput.View() + "\n"
	}

	if m.inputError != "" {
		s += "\n" + styles.ErrorStyle.Render(m.inputError)
	}

	s += "\n" + styles.FooterStyle.Render("Enter: Confirm • Esc: Back • ctrl+c: Quit")

	return s
}

// DateSelectedMsg is sent when a date range is selected
type DateSelectedMsg struct {
	Range daterange.Range
}
