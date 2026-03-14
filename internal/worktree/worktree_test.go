package worktree

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolve_ValidRelativePath(t *testing.T) {
	dir := t.TempDir()
	abs, err := Resolve(dir, "subdir/file.txt")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := filepath.Join(dir, "subdir", "file.txt")
	if abs != expected {
		t.Errorf("got %q, want %q", abs, expected)
	}
}

func TestResolve_RootPath(t *testing.T) {
	dir := t.TempDir()
	abs, err := Resolve(dir, ".")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if abs != filepath.Clean(dir) {
		t.Errorf("got %q, want %q", abs, filepath.Clean(dir))
	}
}

func TestResolve_TraversalAttack(t *testing.T) {
	dir := t.TempDir()
	_, err := Resolve(dir, "../escape")
	if err == nil {
		t.Fatal("expected error for path traversal, got nil")
	}
	te, ok := err.(*ToolError)
	if !ok {
		t.Fatalf("expected ToolError, got %T: %v", err, err)
	}
	if te.Message == "" {
		t.Error("expected non-empty ToolError message")
	}
}

func TestResolve_NullBytes(t *testing.T) {
	dir := t.TempDir()
	_, err := Resolve(dir, "file\x00name")
	if err == nil {
		t.Fatal("expected error for null bytes, got nil")
	}
	te, ok := err.(*ToolError)
	if !ok {
		t.Fatalf("expected ToolError, got %T: %v", err, err)
	}
	if te.Message != "Tool Error: path contains null bytes" {
		t.Errorf("unexpected message: %s", te.Message)
	}
}

func TestResolve_NonExistentRoot(t *testing.T) {
	_, err := Resolve("/nonexistent/path/that/does/not/exist", "file.txt")
	if err == nil {
		t.Fatal("expected error for non-existent root, got nil")
	}
	_, ok := err.(*ToolError)
	if !ok {
		t.Fatalf("expected ToolError, got %T: %v", err, err)
	}
}

func TestResolve_AbsolutePathWithinRoot(t *testing.T) {
	dir := t.TempDir()
	abs, err := Resolve(dir, filepath.Join(dir, "subfile.txt"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := filepath.Join(dir, "subfile.txt")
	if abs != expected {
		t.Errorf("got %q, want %q", abs, expected)
	}
}

func TestResolve_AbsolutePathOutsideRoot(t *testing.T) {
	dir := t.TempDir()
	// Use /tmp as an absolute path that is outside dir (unless dir happens to be /tmp)
	outsidePath := os.TempDir()
	_, err := Resolve(dir, outsidePath)
	if err == nil {
		t.Fatal("expected error for absolute path outside root, got nil")
	}
}
