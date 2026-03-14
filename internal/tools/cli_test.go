package tools

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/iamangus/code-mcp/internal/worktree"
)

// TestExecuteTerminalCommand_EchoStdout verifies stdout capture.
func TestExecuteTerminalCommand_EchoStdout(t *testing.T) {
	dir := t.TempDir()
	stdout, stderr, code, timedOut, err := ExecuteTerminalCommand(dir, "echo hello", 5*time.Second)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if timedOut {
		t.Fatal("should not have timed out")
	}
	if code != 0 {
		t.Errorf("expected exit code 0, got %d", code)
	}
	if !strings.Contains(stdout, "hello") {
		t.Errorf("expected 'hello' in stdout, got %q", stdout)
	}
	_ = stderr
}

// TestExecuteTerminalCommand_NonZeroExit verifies non-zero exit code capture.
func TestExecuteTerminalCommand_NonZeroExit(t *testing.T) {
	dir := t.TempDir()
	_, _, code, timedOut, err := ExecuteTerminalCommand(dir, "exit 42", 5*time.Second)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if timedOut {
		t.Fatal("should not have timed out")
	}
	if code != 42 {
		t.Errorf("expected exit code 42, got %d", code)
	}
}

// TestExecuteTerminalCommand_StderrCapture verifies stderr capture.
func TestExecuteTerminalCommand_StderrCapture(t *testing.T) {
	dir := t.TempDir()
	_, stderr, _, _, err := ExecuteTerminalCommand(dir, "echo error_msg >&2", 5*time.Second)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(stderr, "error_msg") {
		t.Errorf("expected 'error_msg' in stderr, got %q", stderr)
	}
}

// TestExecuteTerminalCommand_Timeout verifies that the timeout flag is set.
func TestExecuteTerminalCommand_Timeout(t *testing.T) {
	dir := t.TempDir()
	_, _, _, timedOut, _ := ExecuteTerminalCommand(dir, "sleep 10", 1*time.Millisecond)
	if !timedOut {
		t.Error("expected timeout flag to be true")
	}
}

// TestExecuteTerminalCommand_InvalidDir verifies ToolError for bad worktree.
func TestExecuteTerminalCommand_InvalidDir(t *testing.T) {
	_, _, _, _, err := ExecuteTerminalCommand("/nonexistent/dir/xyz", "echo hi", 5*time.Second)
	if err == nil {
		t.Fatal("expected error for invalid dir, got nil")
	}
	if _, ok := err.(*worktree.ToolError); !ok {
		t.Fatalf("expected ToolError, got %T", err)
	}
}

// TestGetGitDiff_InGitRepo verifies git diff output in a temporary git repository.
func TestGetGitDiff_InGitRepo(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}

	dir := t.TempDir()

	run := func(args ...string) {
		t.Helper()
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		cmd.Env = append(os.Environ(),
			"GIT_AUTHOR_NAME=Test",
			"GIT_AUTHOR_EMAIL=test@example.com",
			"GIT_COMMITTER_NAME=Test",
			"GIT_COMMITTER_EMAIL=test@example.com",
		)
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v failed: %v\n%s", args, err, out)
		}
	}

	run("init")
	run("config", "user.email", "test@example.com")
	run("config", "user.name", "Test")

	// Create and commit initial file
	filePath := filepath.Join(dir, "hello.txt")
	os.WriteFile(filePath, []byte("initial\n"), 0644)
	run("add", "hello.txt")
	run("commit", "-m", "initial commit")

	// Modify the file (unstaged change)
	os.WriteFile(filePath, []byte("modified\n"), 0644)

	out, err := GetGitDiff(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Should show either diff or status
	if out == "" {
		t.Error("expected non-empty output from GetGitDiff")
	}
}

// TestGetGitDiff_InvalidDir verifies ToolError for bad worktree.
func TestGetGitDiff_InvalidDir(t *testing.T) {
	_, err := GetGitDiff("/nonexistent/dir/xyz")
	if err == nil {
		t.Fatal("expected error for invalid dir, got nil")
	}
	if _, ok := err.(*worktree.ToolError); !ok {
		t.Fatalf("expected ToolError, got %T", err)
	}
}
