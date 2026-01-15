package report

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/burritocatai/activitycat/internal/ui/styles"
)

// Model represents the report display screen
type Model struct {
	viewport  viewport.Model
	textInput textinput.Model
	report    string
	width     int
	height    int
	ready     bool
	saveMode  bool
	saveError string
	saved     bool
}

// New creates a new report model
func New(report string, width, height int) Model {
	// Initialize text input for filename
	ti := textinput.New()
	ti.Placeholder = "report.md"
	ti.CharLimit = 250
	ti.Width = 50

	m := Model{
		report:    report,
		textInput: ti,
		width:     width,
		height:    height,
		ready:     false,
		saveMode:  false,
	}

	// Initialize viewport immediately if dimensions are provided
	if width > 0 && height > 0 {
		m.viewport = viewport.New(width, height-4)
		m.viewport.SetContent(styles.ReportStyle.Render(report))
		m.ready = true
	}

	return m
}

// Init initializes the report model
func (m Model) Init() tea.Cmd {
	return nil
}

// Update handles messages for the report screen
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		// If in save mode, handle text input
		if m.saveMode {
			switch msg.String() {
			case "enter":
				// Save the file
				filename := strings.TrimSpace(m.textInput.Value())
				if filename == "" {
					filename = "report.md"
				}
				if err := m.saveReport(filename); err != nil {
					m.saveError = err.Error()
				} else {
					m.saved = true
					m.saveError = ""
				}
				m.saveMode = false
				m.textInput.SetValue("")
				return m, nil
			case "esc":
				// Cancel save
				m.saveMode = false
				m.saveError = ""
				m.textInput.SetValue("")
				return m, nil
			default:
				// Update text input
				m.textInput, cmd = m.textInput.Update(msg)
				return m, cmd
			}
		}

		// Normal mode key handling
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "b":
			return m, func() tea.Msg {
				return BackMsg{}
			}
		case "s":
			// Enter save mode
			m.saveMode = true
			m.saved = false
			m.saveError = ""
			m.textInput.Focus()
			return m, textinput.Blink
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		if !m.ready {
			m.viewport = viewport.New(msg.Width, msg.Height-4)
			m.viewport.SetContent(styles.ReportStyle.Render(m.report))
			m.ready = true
		} else {
			m.viewport.Width = msg.Width
			m.viewport.Height = msg.Height - 4
			m.viewport.SetContent(styles.ReportStyle.Render(m.report))
		}
	}

	// Only update viewport if not in save mode
	if !m.saveMode {
		m.viewport, cmd = m.viewport.Update(msg)
	}
	return m, cmd
}

// View renders the report screen
func (m Model) View() string {
	if !m.ready {
		return "\n  Initializing..."
	}

	header := styles.TitleStyle.Render("Generated Report")

	var footer string
	if m.saveMode {
		// Show save prompt
		footer = lipgloss.JoinVertical(
			lipgloss.Left,
			"",
			"Save report to file:",
			m.textInput.View(),
			"",
			styles.SubtleStyle.Render("Enter: Save • Esc: Cancel"),
		)
	} else if m.saved {
		footer = lipgloss.JoinVertical(
			lipgloss.Left,
			styles.MergedStyle.Render("✓ Report saved successfully!"),
			styles.FooterStyle.Render("↑/↓: Scroll • s: Save • b: Back • q: Quit"),
		)
	} else if m.saveError != "" {
		footer = lipgloss.JoinVertical(
			lipgloss.Left,
			styles.ErrorStyle.Render("✗ Error: "+m.saveError),
			styles.FooterStyle.Render("↑/↓: Scroll • s: Save • b: Back • q: Quit"),
		)
	} else {
		footer = styles.FooterStyle.Render("↑/↓: Scroll • s: Save • b: Back • q: Quit")
	}

	return lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		m.viewport.View(),
		footer,
	)
}

// SetSize updates the dimensions
func (m *Model) SetSize(width, height int) {
	m.width = width
	m.height = height
	if m.ready {
		m.viewport.Width = width
		m.viewport.Height = height - 4
		m.viewport.SetContent(styles.ReportStyle.Render(m.report))
	}
}

// saveReport saves the report to a file
func (m *Model) saveReport(filename string) error {
	// Expand home directory if needed
	if strings.HasPrefix(filename, "~") {
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("could not get home directory: %w", err)
		}
		filename = filepath.Join(home, filename[1:])
	}

	// Make sure the filename has an extension
	if filepath.Ext(filename) == "" {
		filename += ".md"
	}

	// Create parent directories if needed
	dir := filepath.Dir(filename)
	if dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("could not create directory: %w", err)
		}
	}

	// Write the file
	if err := os.WriteFile(filename, []byte(m.report), 0644); err != nil {
		return fmt.Errorf("could not write file: %w", err)
	}

	// Get absolute path for display
	absPath, _ := filepath.Abs(filename)
	m.saveError = "" // Clear any previous error
	m.saved = true

	// Store the saved path for display (optional enhancement)
	_ = absPath

	return nil
}

// BackMsg is sent when the user wants to go back
type BackMsg struct{}
