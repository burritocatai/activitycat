package claude

import (
	"context"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/burritocatai/activitycat/internal/analytics"
	"github.com/burritocatai/activitycat/internal/github"
)

// CheckAPIKey verifies that the ANTHROPIC_API_KEY environment variable is set
func CheckAPIKey() error {
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		return fmt.Errorf("ANTHROPIC_API_KEY environment variable not set")
	}
	return nil
}

// GenerateReport calls the Claude API to generate a report from all activity data
func GenerateReport(
	ctx context.Context,
	prs []github.PullRequest,
	issues []github.Issue,
	reviews []github.Review,
	commits []github.Commit,
	commentedItems []github.CommentedItem,
	metrics *analytics.Metrics,
	prompt string,
) (string, error) {
	if err := CheckAPIKey(); err != nil {
		return "", err
	}

	client := anthropic.NewClient()

	// Format all activity data including metrics
	metricsText := ""
	if metrics != nil {
		metricsText = metrics.Format()
	}
	activityData := github.FormatActivityForClaude(prs, issues, reviews, commits, commentedItems, metricsText)

	userMessage := fmt.Sprintf("%s\n\nHere is my GitHub activity data:\n\n%s", prompt, activityData)

	message, err := client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     anthropic.ModelClaudeSonnet4_5_20250929,
		MaxTokens: 4096,
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(userMessage)),
		},
	})

	if err != nil {
		return "", fmt.Errorf("Claude API error: %w", err)
	}

	if len(message.Content) == 0 {
		return "", fmt.Errorf("no content in Claude API response")
	}

	contentBlock := message.Content[0]
	return contentBlock.Text, nil
}

// GenerateReportCmd wraps GenerateReport in a Bubbletea Cmd
func GenerateReportCmd(
	prs []github.PullRequest,
	issues []github.Issue,
	reviews []github.Review,
	commits []github.Commit,
	commentedItems []github.CommentedItem,
	metrics *analytics.Metrics,
	prompt string,
) tea.Cmd {
	return func() tea.Msg {
		report, err := GenerateReport(context.Background(), prs, issues, reviews, commits, commentedItems, metrics, prompt)
		return ReportGeneratedMsg{
			Report: report,
			Error:  err,
		}
	}
}

// ReportGeneratedMsg is sent when the report is generated
type ReportGeneratedMsg struct {
	Report string
	Error  error
}
