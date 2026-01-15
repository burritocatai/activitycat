package claude

// ReportRequest represents the data needed to generate a report
type ReportRequest struct {
	PRData string
	Prompt string
}

// ReportResponse represents the generated report
type ReportResponse struct {
	Content string
	Error   error
}
