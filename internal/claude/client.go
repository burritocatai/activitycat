package claude

import (
	"context"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/anthropics/anthropic-sdk-go"
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

// GenerateReport calls the Claude API to generate a report from PR data
func GenerateReport(ctx context.Context, prs []github.PullRequest, prompt string) (string, error) {
	// Check for API key
	if err := CheckAPIKey(); err != nil {
		return "", err
	}

	// Create client (will automatically use ANTHROPIC_API_KEY from environment)
	client := anthropic.NewClient()

	// Format PR data
	prData := github.FormatPRsForClaude(prs)

	// Build the message
	userMessage := fmt.Sprintf("%s\n\nHere is my GitHub PR activity data:\n\n%s", prompt, prData)

	// Call the API
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

	// Extract text content from response
	if len(message.Content) == 0 {
		return "", fmt.Errorf("no content in Claude API response")
	}

	// Get the text from the first content block
	// The Content field is a slice of ContentBlockUnion which has a Text field
	contentBlock := message.Content[0]
	return contentBlock.Text, nil
}

// GenerateReportCmd wraps GenerateReport in a Bubbletea Cmd
func GenerateReportCmd(prs []github.PullRequest, prompt string) tea.Cmd {
	return func() tea.Msg {
		report, err := GenerateReport(context.Background(), prs, prompt)
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
