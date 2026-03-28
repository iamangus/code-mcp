package github

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHTTPClient_CreatePR(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/repos/owner/myrepo/pulls" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer test-token" {
			t.Errorf("wrong auth header: %s", r.Header.Get("Authorization"))
		}

		var body map[string]any
		json.NewDecoder(r.Body).Decode(&body)
		if body["title"] != "Test PR" {
			t.Errorf("unexpected title: %v", body["title"])
		}
		if body["draft"] != true {
			t.Errorf("expected draft=true, got %v", body["draft"])
		}

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]any{
			"number":   42,
			"html_url": "https://github.com/owner/myrepo/pull/42",
		})
	}))
	defer srv.Close()

	c := NewHTTPClient("test-token", "owner", slog.Default(), WithBaseURL(srv.URL))
	pr, err := c.CreatePR(context.Background(), CreatePROptions{
		Repo:  "myrepo",
		Title: "Test PR",
		Head:  "feature",
		Base:  "main",
		Body:  "description",
		Draft: true,
	})
	if err != nil {
		t.Fatalf("CreatePR: %v", err)
	}
	if pr.Number != 42 {
		t.Errorf("expected PR #42, got #%d", pr.Number)
	}
	if pr.HTMLURL != "https://github.com/owner/myrepo/pull/42" {
		t.Errorf("unexpected URL: %s", pr.HTMLURL)
	}
}

func TestHTTPClient_CreatePR_NotDraft(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		json.NewDecoder(r.Body).Decode(&body)
		if body["draft"] != false {
			t.Errorf("expected draft=false, got %v", body["draft"])
		}
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]any{"number": 1, "html_url": "https://example.com/1"})
	}))
	defer srv.Close()

	c := NewHTTPClient("tok", "owner", slog.Default(), WithBaseURL(srv.URL))
	_, err := c.CreatePR(context.Background(), CreatePROptions{Repo: "r", Title: "t", Head: "h", Base: "b", Draft: false})
	if err != nil {
		t.Fatalf("CreatePR: %v", err)
	}
}

func TestHTTPClient_UpdatePR(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			t.Errorf("expected PATCH, got %s", r.Method)
		}
		if r.URL.Path != "/repos/owner/myrepo/pulls/42" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		var body map[string]any
		json.NewDecoder(r.Body).Decode(&body)
		if body["body"] != "new body" {
			t.Errorf("unexpected body: %v", body["body"])
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{}`))
	}))
	defer srv.Close()

	c := NewHTTPClient("test-token", "owner", slog.Default(), WithBaseURL(srv.URL))
	if err := c.UpdatePR(context.Background(), "myrepo", 42, "new body"); err != nil {
		t.Fatalf("UpdatePR: %v", err)
	}
}

func TestHTTPClient_PromotePR(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			t.Errorf("expected PATCH, got %s", r.Method)
		}
		var body map[string]any
		json.NewDecoder(r.Body).Decode(&body)
		if body["draft"] != false {
			t.Errorf("expected draft=false, got %v", body["draft"])
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{}`))
	}))
	defer srv.Close()

	c := NewHTTPClient("test-token", "owner", slog.Default(), WithBaseURL(srv.URL))
	if err := c.PromotePR(context.Background(), "myrepo", 42); err != nil {
		t.Fatalf("PromotePR: %v", err)
	}
}

func TestHTTPClient_ErrorResponse(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnprocessableEntity)
		w.Write([]byte(`{"message":"Validation Failed"}`))
	}))
	defer srv.Close()

	c := NewHTTPClient("test-token", "owner", slog.Default(), WithBaseURL(srv.URL))
	_, err := c.CreatePR(context.Background(), CreatePROptions{
		Repo: "myrepo", Title: "t", Head: "h", Base: "b",
	})
	if err == nil {
		t.Fatal("expected error for 422 response")
	}
}
