package config

import (
	"os"
	"path/filepath"
	"strings"
)

// Prompt represents a user-defined prompt template
type Prompt struct {
	Name    string
	Content string
}

// defaultPrompt is used when no user prompts are found
var defaultPrompt = Prompt{
	Name: "Default Report",
	Content: `Please analyze my GitHub PR activity and create a concise monthly report.

Include:
1. Summary of activity (number of PRs, repositories involved)
2. Key contributions and themes
3. Notable achievements or patterns
4. Areas of focus

Keep the report professional and highlight the most important work.`,
}

// LoadPrompts reads all prompt files from $HOME/.config/activitycat/prompts/
// Returns at least one prompt (the default if no user prompts exist)
func LoadPrompts() ([]Prompt, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		// Return default prompt if we can't get home dir
		return []Prompt{defaultPrompt}, nil
	}

	promptsDir := filepath.Join(homeDir, ".config", "activitycat", "prompts")

	// Check if directory exists
	if _, err := os.Stat(promptsDir); os.IsNotExist(err) {
		// Directory doesn't exist, return default
		return []Prompt{defaultPrompt}, nil
	}

	// Read directory contents
	entries, err := os.ReadDir(promptsDir)
	if err != nil {
		// Error reading directory, return default
		return []Prompt{defaultPrompt}, nil
	}

	var prompts []Prompt
	for _, entry := range entries {
		// Skip directories and hidden files
		if entry.IsDir() || strings.HasPrefix(entry.Name(), ".") {
			continue
		}

		// Read file content
		filePath := filepath.Join(promptsDir, entry.Name())
		content, err := os.ReadFile(filePath)
		if err != nil {
			// Skip files we can't read
			continue
		}

		// Use filename without extension as name
		name := strings.TrimSuffix(entry.Name(), filepath.Ext(entry.Name()))

		prompts = append(prompts, Prompt{
			Name:    name,
			Content: string(content),
		})
	}

	// If no prompts were loaded, return default
	if len(prompts) == 0 {
		return []Prompt{defaultPrompt}, nil
	}

	return prompts, nil
}
