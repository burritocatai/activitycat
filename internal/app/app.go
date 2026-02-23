package app

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/burritocatai/activitycat/internal/analytics"
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
	state  State
	width  int
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
	reviews         []github.Review
	commits         []github.Commit
	commentedItems  []github.CommentedItem
	metrics         *analytics.Metrics
	prompts         []config.Prompt
	selectedPrompt  config.Prompt
	generatedReport string
	err             error
}

// New creates a new application model
func New() Model {
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

		if m.state == StatePRList {
			m.prList.SetSize(msg.Width, msg.Height)
		}
		if m.state == StateReport {
			m.reportView.SetSize(msg.Width, msg.Height)
		}

	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}

	case dateselect.DateSelectedMsg:
		m.selectedRange = msg.Range
		m.loading = loading.New("Fetching activity data (PRs, issues, reviews, commits, comments)...")
		m.state = StateLoading
		return m, tea.Batch(
			m.loading.Init(),
			github.FetchActivityCmd(m.selectedRange),
		)

	case github.ActivityLoadedMsg:
		if msg.Error != nil {
			m.err = msg.Error
			m.state = StateError
			return m, nil
		}
		m.prs = msg.PRs
		m.issues = msg.Issues
		m.reviews = msg.Reviews
		m.commits = msg.Commits
		m.commentedItems = msg.CommentedItems
		m.metrics = analytics.Compute(m.prs, m.issues, m.reviews, m.commits, m.commentedItems, m.selectedRange)
		m.prList = prlist.New(m.prs, m.issues, m.reviews, m.commits, m.commentedItems, m.metrics, m.width, m.height)
		m.state = StatePRList
		return m, nil

	case prlist.ContinueMsg:
		m.promptSelect = promptselect.New(m.prompts)
		m.state = StatePromptSelect
		return m, nil

	case prlist.BackMsg:
		m.state = StateSelectDate
		return m, nil

	case promptselect.PromptSelectedMsg:
		m.selectedPrompt = msg.Prompt
		m.loading = loading.New("Generating report with Claude AI...")
		m.state = StateGenerating
		return m, tea.Batch(
			m.loading.Init(),
			claude.GenerateReportCmd(m.prs, m.issues, m.reviews, m.commits, m.commentedItems, m.metrics, m.selectedPrompt.Content),
		)

	case promptselect.BackMsg:
		m.state = StatePRList
		return m, nil

	case claude.ReportGeneratedMsg:
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
		m.state = StatePromptSelect
		return m, nil
	}

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
