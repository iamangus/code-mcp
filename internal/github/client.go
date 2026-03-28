package github

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"
)

const defaultBaseURL = "https://api.github.com"

// HTTPClient implements Client using the GitHub REST API.
type HTTPClient struct {
	token   string
	owner   string
	baseURL string
	http    *http.Client
	logger  *slog.Logger
}

// HTTPClientOption configures an HTTPClient.
type HTTPClientOption func(*HTTPClient)

// WithBaseURL overrides the GitHub API base URL (used in tests).
func WithBaseURL(url string) HTTPClientOption {
	return func(c *HTTPClient) { c.baseURL = url }
}

// NewHTTPClient creates a new GitHub API client.
func NewHTTPClient(token, owner string, logger *slog.Logger, opts ...HTTPClientOption) *HTTPClient {
	c := &HTTPClient{
		token:   token,
		owner:   owner,
		baseURL: defaultBaseURL,
		http:    &http.Client{},
		logger:  logger,
	}
	for _, o := range opts {
		o(c)
	}
	return c
}

func (c *HTTPClient) CreatePR(ctx context.Context, opts CreatePROptions) (*PR, error) {
	start := time.Now()
	path := fmt.Sprintf("/repos/%s/%s/pulls", c.owner, opts.Repo)
	payload := map[string]any{
		"title": opts.Title,
		"head":  opts.Head,
		"base":  opts.Base,
		"body":  opts.Body,
		"draft": opts.Draft,
	}

	var pr PR
	if err := c.do(ctx, http.MethodPost, path, payload, &pr); err != nil {
		c.logger.Error("github: CreatePR failed", "repo", opts.Repo, "error", err, "duration_ms", time.Since(start).Milliseconds())
		return nil, err
	}
	c.logger.Info("github: PR created", "repo", opts.Repo, "number", pr.Number, "duration_ms", time.Since(start).Milliseconds())
	return &pr, nil
}

func (c *HTTPClient) UpdatePR(ctx context.Context, repo string, number int, body string) error {
	start := time.Now()
	path := fmt.Sprintf("/repos/%s/%s/pulls/%d", c.owner, repo, number)
	payload := map[string]any{"body": body}

	if err := c.do(ctx, http.MethodPatch, path, payload, nil); err != nil {
		c.logger.Error("github: UpdatePR failed", "repo", repo, "number", number, "error", err, "duration_ms", time.Since(start).Milliseconds())
		return err
	}
	c.logger.Info("github: PR updated", "repo", repo, "number", number, "duration_ms", time.Since(start).Milliseconds())
	return nil
}

func (c *HTTPClient) PromotePR(ctx context.Context, repo string, number int) error {
	start := time.Now()
	path := fmt.Sprintf("/repos/%s/%s/pulls/%d", c.owner, repo, number)
	payload := map[string]any{"draft": false}

	if err := c.do(ctx, http.MethodPatch, path, payload, nil); err != nil {
		c.logger.Error("github: PromotePR failed", "repo", repo, "number", number, "error", err, "duration_ms", time.Since(start).Milliseconds())
		return err
	}
	c.logger.Info("github: PR promoted", "repo", repo, "number", number, "duration_ms", time.Since(start).Milliseconds())
	return nil
}

func (c *HTTPClient) do(ctx context.Context, method, path string, reqBody any, out any) error {
	var bodyReader io.Reader
	if reqBody != nil {
		data, err := json.Marshal(reqBody)
		if err != nil {
			return fmt.Errorf("marshal request: %w", err)
		}
		bodyReader = bytes.NewReader(data)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, bodyReader)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Accept", "application/vnd.github+json")
	if reqBody != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return fmt.Errorf("github API %s %s: %d %s", method, path, resp.StatusCode, string(respBody))
	}

	if out != nil && len(respBody) > 0 {
		if err := json.Unmarshal(respBody, out); err != nil {
			return fmt.Errorf("decode response: %w", err)
		}
	}
	return nil
}

// Compile-time check that HTTPClient implements Client.
var _ Client = (*HTTPClient)(nil)
