package github

import "context"

// Client is the interface for GitHub API operations.
type Client interface {
	CreatePR(ctx context.Context, opts CreatePROptions) (*PR, error)
	UpdatePR(ctx context.Context, repo string, number int, body string) error
	PromotePR(ctx context.Context, repo string, number int) error
}

// CreatePROptions holds parameters for creating a pull request.
type CreatePROptions struct {
	Repo  string
	Title string
	Head  string
	Base  string
	Body  string
	Draft bool
}

// PR represents a GitHub pull request.
type PR struct {
	Number  int    `json:"number"`
	HTMLURL string `json:"html_url"`
}
