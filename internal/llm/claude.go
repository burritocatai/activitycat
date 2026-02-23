package llm

import (
	"context"
	"fmt"
	"os"

	"github.com/anthropics/anthropic-sdk-go"
)

// ClaudeProvider implements Provider using the Anthropic API.
type ClaudeProvider struct {
	model string
}

// NewClaudeProvider creates a ClaudeProvider with the given model name.
func NewClaudeProvider(model string) *ClaudeProvider {
	return &ClaudeProvider{model: model}
}

// CheckAPIKey verifies that the ANTHROPIC_API_KEY environment variable is set.
func CheckAPIKey() error {
	if os.Getenv("ANTHROPIC_API_KEY") == "" {
		return fmt.Errorf("ANTHROPIC_API_KEY environment variable not set")
	}
	return nil
}

// GenerateReport sends the user message to Claude and returns the response.
func (c *ClaudeProvider) GenerateReport(ctx context.Context, userMessage string) (string, error) {
	if err := CheckAPIKey(); err != nil {
		return "", err
	}

	client := anthropic.NewClient()

	message, err := client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     anthropic.Model(c.model),
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

	return message.Content[0].Text, nil
}
