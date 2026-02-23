# activitycat 🐱

A beautiful terminal UI application that fetches your GitHub PR activity and generates AI-powered monthly reports using Claude or Ollama.

## Features

- 📊 **Interactive TUI** - Built with [Bubbletea](https://github.com/charmbracelet/bubbletea) for a smooth terminal experience
- 🔍 **GitHub Integration** - Fetches your PRs using the GitHub CLI (`gh`)
- 🤖 **AI-Powered Reports** - Generates insightful reports with Claude AI or local Ollama models
- 📅 **Flexible Date Ranges** - Last Week, Last Month, or Last 3 Months
- 🎨 **Beautiful Display** - Color-coded PR states and scrollable views
- ⚙️ **Custom Prompts** - Define your own report templates

## Prerequisites

1. **GitHub CLI** - Install and authenticate:
   ```bash
   # Install (macOS)
   brew install gh

   # Authenticate
   gh auth login
   ```

2. **Anthropic API Key** (only if using Claude) - Get one from [console.anthropic.com](https://console.anthropic.com/):
   ```bash
   export ANTHROPIC_API_KEY=your_key_here
   ```

3. **Ollama** (optional) - Install from [ollama.com](https://ollama.com/) if you want to use local models instead of Claude.

## Installation

### From Release

Download the latest binary for your platform from [Releases](https://github.com/burritocatai/activitycat/releases).

### From Source

```bash
go install github.com/burritocatai/activitycat@latest
```

## Usage

Simply run:
```bash
activitycat
```

The app will guide you through:
1. **Select Date Range** - Choose from preset options
2. **View PRs** - Review your pull requests with details
3. **Select Report Prompt** - Choose how you want your report generated
4. **View Generated Report** - Read your AI-generated activity report

### Keyboard Controls

- `↑/↓` or `j/k` - Navigate menus and scroll
- `Enter` - Select/Continue
- `b` - Go back to previous screen
- `q` or `Ctrl+C` - Quit

## Configuration

By default, activitycat uses Claude for report generation. To switch providers or customize the model, create a config file at:

```
$HOME/.config/activitycat/config.toml
```

### Using Ollama

```toml
provider = "ollama"
model = "llama3"
ollama_host = "http://localhost:11434"
```

### Using Claude with a specific model

```toml
provider = "claude"
model = "claude-sonnet-4-5-20250929"
```

All fields are optional — defaults are `provider = "claude"`, `model = "claude-sonnet-4-5-20250929"`, and `ollama_host = "http://localhost:11434"`.

## Custom Prompts

Create custom report prompts by adding text files to:
```
$HOME/.config/activitycat/prompts/
```

### Example Prompts

**detailed-summary.txt**:
```
Please create a detailed monthly report of my GitHub activity.

For each repository, provide:
- Number of PRs opened, merged, and closed
- Major features or improvements implemented
- Code review participation
- Overall impact and achievements

Format as a professional summary suitable for performance reviews.
```

**team-update.txt**:
```
Create a concise team update based on my GitHub PR activity.

Highlight:
- Key accomplishments this period
- Current work in progress
- Blockers or challenges faced
- Upcoming priorities

Keep it brief and actionable for team standup sharing.
```

**executive-brief.txt**:
```
Generate an executive summary of my development work.

Focus on:
- High-level impact and business value delivered
- Cross-functional collaboration
- Technical leadership and mentorship
- Strategic contributions

Use business-friendly language, avoid technical jargon.
```

## How It Works

1. **Fetch PRs** - Uses `gh search prs --author "@me"` to get your PRs across all repositories
2. **Display Activity** - Shows PR details including title, repository, state, dates, reviewers, and descriptions
3. **Generate Report** - Sends PR data + your selected prompt to your configured LLM provider (Claude or Ollama)
4. **Show Results** - Displays the AI-generated report in a scrollable view

## Development

### Project Structure

```
activitycat/
├── main.go                          # Entry point
├── internal/
│   ├── app/                         # Main application orchestrator
│   ├── ui/                          # UI components
│   │   ├── dateselect/             # Date selection screen
│   │   ├── loading/                # Loading spinner
│   │   ├── prlist/                 # PR list display
│   │   ├── promptselect/           # Prompt selection
│   │   ├── report/                 # Report display
│   │   └── styles/                 # Lipgloss styles
│   ├── github/                      # GitHub CLI integration
│   ├── llm/                         # LLM provider integration (Claude, Ollama)
│   ├── config/                      # Configuration management
│   └── daterange/                   # Date range utilities
└── README.md
```

### Building

```bash
go build -o activitycat
```

### Testing

```bash
go test ./...
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

MIT License - See LICENSE file for details

## Credits

Built with:
- [Bubbletea](https://github.com/charmbracelet/bubbletea) - TUI framework
- [Lipgloss](https://github.com/charmbracelet/lipgloss) - Style definitions
- [Bubbles](https://github.com/charmbracelet/bubbles) - TUI components
- [Anthropic SDK](https://github.com/anthropics/anthropic-sdk-go) - Claude AI integration
- [Ollama](https://ollama.com/) - Local LLM support
- [GitHub CLI](https://cli.github.com/) - GitHub API access
