package github

import (
	"context"
	"sync"
)

// FakeClient is a test double for Client.
type FakeClient struct {
	mu    sync.Mutex
	Calls []FakeCall

	CreatePRResult *PR
	CreatePRError  error
	UpdatePRError  error
	PromotePRError error
}

// FakeCall records a method invocation.
type FakeCall struct {
	Method string
	Args   []any
}

// NewFakeClient creates a FakeClient with sensible defaults.
func NewFakeClient() *FakeClient {
	return &FakeClient{
		CreatePRResult: &PR{Number: 1, HTMLURL: "https://github.com/test/test/pull/1"},
	}
}

func (f *FakeClient) CreatePR(_ context.Context, opts CreatePROptions) (*PR, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.Calls = append(f.Calls, FakeCall{Method: "CreatePR", Args: []any{opts}})
	return f.CreatePRResult, f.CreatePRError
}

func (f *FakeClient) UpdatePR(_ context.Context, repo string, number int, body string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.Calls = append(f.Calls, FakeCall{Method: "UpdatePR", Args: []any{repo, number, body}})
	return f.UpdatePRError
}

func (f *FakeClient) PromotePR(_ context.Context, repo string, number int) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.Calls = append(f.Calls, FakeCall{Method: "PromotePR", Args: []any{repo, number}})
	return f.PromotePRError
}

// Compile-time check that FakeClient implements Client.
var _ Client = (*FakeClient)(nil)
