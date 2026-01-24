package app

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/burritocatai/activitycat/internal/claude"
	"github.com/burritocatai/activitycat/internal/config"
	"github.com/burritocatai/activitycat/internal/daterange"
	"github.com/burritocatai/activitycat/internal/github"
	"github.com/burritocatai/activitycat/internal/ui/dateselect"
	"github.com/burritocatai/activitycat/internal/ui/loading"
	"github.com/burritocatai/activitycat/internal/ui/prlist"
	"github.com/burritocatai/activitycat/internal/ui/promptselect"
	"github.com/burritocatai/activitycat/internal/ui/report"
	"github.com/burritocatai/activitycat/internal/ui/styles"
)

// State represents the current screen/state of the application
type State int

const (
	StateSelectDate State = iota
	StateLoading
	StatePRList
	StatePromptSelect
	StateGenerating
	StateReport
	StateError
)

// Model is the main application model
type Model struct {
	state State
	width int
	height int

	// Sub-models for each screen
	dateSelect   dateselect.Model
	loading      loading.Model
	prList       prlist.Model
	promptSelect promptselect.Model
	reportView   report.Model

	// Shared data
	selectedRange   daterange.Range
	prs             []github.PullRequest
	issues          []github.Issue
	prompts         []config.Prompt
	selectedPrompt  config.Prompt
	generatedReport string
	err             error
}

// New creates a new application model
func New() Model {
	// Load prompts early
	prompts, _ := config.LoadPrompts()

	return Model{
		state:      StateSelectDate,
		dateSelect: dateselect.New(),
		prompts:    prompts,
	}
}

// Init initializes the application
func (m Model) Init() tea.Cmd {
	return m.dateSelect.Init()
}

// Update handles all messages and state transitions
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		// Update sub-models with new size
		if m.state == StatePRList {
			m.prList.SetSize(msg.Width, msg.Height)
		}
		if m.state == StateReport {
			m.reportView.SetSize(msg.Width, msg.Height)
		}

	case tea.KeyMsg:
		// Global quit
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}

	case dateselect.DateSelectedMsg:
		// Transition from date selection to loading
		m.selectedRange = msg.Range
		m.loading = loading.New("Fetching pull requests and closed issues...")
		m.state = StateLoading
		return m, tea.Batch(
			m.loading.Init(),
			github.FetchPRsCmd(m.selectedRange),
		)

	case github.PRsLoadedMsg:
		// Transition from loading to PR list or error
		if msg.Error != nil {
			m.err = msg.Error
			m.state = StateError
			return m, nil
		}
		m.prs = msg.PRs
		m.issues = msg.Issues
		m.prList = prlist.New(m.prs, m.issues, m.width, m.height)
		m.state = StatePRList
		return m, nil

	case prlist.ContinueMsg:
		// Transition from PR list to prompt selection
		m.promptSelect = promptselect.New(m.prompts)
		m.state = StatePromptSelect
		return m, nil

	case prlist.BackMsg:
		// Go back to date selection from PR list
		m.state = StateSelectDate
		return m, nil

	case promptselect.PromptSelectedMsg:
		// Transition from prompt selection to generating
		m.selectedPrompt = msg.Prompt
		m.loading = loading.New("Generating report with Claude AI...")
		m.state = StateGenerating
		return m, tea.Batch(
			m.loading.Init(),
			claude.GenerateReportCmd(m.prs, m.issues, m.selectedPrompt.Content),
		)

	case promptselect.BackMsg:
		// Go back to PR list from prompt selection
		m.state = StatePRList
		return m, nil

	case claude.ReportGeneratedMsg:
		// Transition from generating to report view or error
		if msg.Error != nil {
			m.err = msg.Error
			m.state = StateError
			return m, nil
		}
		m.generatedReport = msg.Report
		m.reportView = report.New(m.generatedReport, m.width, m.height)
		m.state = StateReport
		return m, nil

	case report.BackMsg:
		// Go back to prompt selection from report
		m.state = StatePromptSelect
		return m, nil
	}

	// Delegate to current state's model
	return m.updateCurrentState(msg)
}

// updateCurrentState delegates update to the current state's model
func (m Model) updateCurrentState(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch m.state {
	case StateSelectDate:
		m.dateSelect, cmd = m.dateSelect.Update(msg)
	case StateLoading, StateGenerating:
		m.loading, cmd = m.loading.Update(msg)
	case StatePRList:
		m.prList, cmd = m.prList.Update(msg)
	case StatePromptSelect:
		m.promptSelect, cmd = m.promptSelect.Update(msg)
	case StateReport:
		m.reportView, cmd = m.reportView.Update(msg)
	case StateError:
		// In error state, any key returns to date selection
		if _, ok := msg.(tea.KeyMsg); ok {
			m.state = StateSelectDate
			m.err = nil
		}
	}

	return m, cmd
}

// View renders the current state's view
func (m Model) View() string {
	switch m.state {
	case StateSelectDate:
		return m.dateSelect.View()
	case StateLoading, StateGenerating:
		return m.loading.View()
	case StatePRList:
		return m.prList.View()
	case StatePromptSelect:
		return m.promptSelect.View()
	case StateReport:
		return m.reportView.View()
	case StateError:
		return m.renderError()
	default:
		return "Unknown state"
	}
}

// renderError renders the error state
func (m Model) renderError() string {
	s := styles.TitleStyle.Render("Error")
	s += "\n\n"
	s += styles.ErrorStyle.Render(m.err.Error())
	s += "\n\n"
	s += styles.FooterStyle.Render("Press any key to return to date selection")
	return s
}
