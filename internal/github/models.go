package github

import (
	"time"
)

// PullRequest represents a GitHub pull request
type PullRequest struct {
	Number     int        `json:"number"`
	Title      string     `json:"title"`
	State      string     `json:"state"` // open, closed, merged
	Body       string     `json:"body"`
	CreatedAt  time.Time  `json:"createdAt"`
	ClosedAt   *time.Time `json:"closedAt,omitempty"`
	MergedAt   *time.Time `json:"mergedAt,omitempty"`
	Author     Author     `json:"author"`
	Repository Repository `json:"repository"`
	// ReviewRequests contains users who were requested to review
	ReviewRequests []ReviewRequest `json:"reviewRequests"`
}

// Author represents a GitHub user
type Author struct {
	Login string `json:"login"`
}

// Repository represents a GitHub repository
type Repository struct {
	Name          string `json:"name"`
	NameWithOwner string `json:"nameWithOwner"`
}

// ReviewRequest represents a review request
type ReviewRequest struct {
	Login string `json:"login"`
}

// IsOpen returns true if the PR is open
func (pr PullRequest) IsOpen() bool {
	return pr.State == "open"
}

// IsMerged returns true if the PR is merged
func (pr PullRequest) IsMerged() bool {
	return pr.MergedAt != nil
}

// IsClosed returns true if the PR is closed (but not merged)
func (pr PullRequest) IsClosed() bool {
	return pr.State == "closed" && pr.MergedAt == nil
}

// Reviewers returns a list of reviewer logins
func (pr PullRequest) Reviewers() []string {
	reviewers := make([]string, len(pr.ReviewRequests))
	for i, req := range pr.ReviewRequests {
		reviewers[i] = req.Login
	}
	return reviewers
}
