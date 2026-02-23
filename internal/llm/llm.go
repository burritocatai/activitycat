package llm

import (
	"context"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/burritocatai/activitycat/internal/analytics"
	"github.com/burritocatai/activitycat/internal/config"
	"github.com/burritocatai/activitycat/internal/github"
)

// Provider is the interface for LLM report generation backends.
type Provider interface {
	GenerateReport(ctx context.Context, userMessage string) (string, error)
}

// ReportGeneratedMsg is sent when the report generation completes.
type ReportGeneratedMsg struct {
	Report string
	Error  error
}

// NewProvider creates a Provider based on the given config.
func NewProvider(cfg config.Config) (Provider, error) {
	switch cfg.Provider {
	case "claude":
		return NewClaudeProvider(cfg.Model), nil
	case "ollama":
		return NewOllamaProvider(cfg.OllamaHost, cfg.Model), nil
	default:
		return nil, fmt.Errorf("unknown LLM provider: %q (expected \"claude\" or \"ollama\")", cfg.Provider)
	}
}

// GenerateReportCmd wraps any Provider in a bubbletea Cmd.
func GenerateReportCmd(
	provider Provider,
	prs []github.PullRequest,
	issues []github.Issue,
	reviews []github.Review,
	commits []github.Commit,
	commentedItems []github.CommentedItem,
	metrics *analytics.Metrics,
	prompt string,
) tea.Cmd {
	return func() tea.Msg {
		metricsText := ""
		if metrics != nil {
			metricsText = metrics.Format()
		}
		activityData := github.FormatActivityForClaude(prs, issues, reviews, commits, commentedItems, metricsText)
		userMessage := fmt.Sprintf("%s\n\nHere is my GitHub activity data:\n\n%s", prompt, activityData)

		report, err := provider.GenerateReport(context.Background(), userMessage)
		return ReportGeneratedMsg{
			Report: report,
			Error:  err,
		}
	}
}
