package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/burritocatai/activitycat/internal/app"
	"github.com/burritocatai/activitycat/internal/claude"
	"github.com/burritocatai/activitycat/internal/github"
)

func main() {
	// Check prerequisites before starting TUI
	if err := github.CheckAuth(); err != nil {
		fmt.Fprintf(os.Stderr, "GitHub authentication error: %v\n", err)
		fmt.Fprintf(os.Stderr, "\nPlease install and authenticate the GitHub CLI:\n")
		fmt.Fprintf(os.Stderr, "  1. Install: https://cli.github.com/\n")
		fmt.Fprintf(os.Stderr, "  2. Authenticate: gh auth login\n")
		os.Exit(1)
	}

	if err := claude.CheckAPIKey(); err != nil {
		fmt.Fprintf(os.Stderr, "Claude API error: %v\n", err)
		fmt.Fprintf(os.Stderr, "\nPlease set your Anthropic API key:\n")
		fmt.Fprintf(os.Stderr, "  export ANTHROPIC_API_KEY=your_key_here\n")
		fmt.Fprintf(os.Stderr, "\nGet your API key at: https://console.anthropic.com/\n")
		os.Exit(1)
	}

	// Start the TUI application
	p := tea.NewProgram(
		app.New(),
		tea.WithAltScreen(),
	)

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running application: %v\n", err)
		os.Exit(1)
	}
}
