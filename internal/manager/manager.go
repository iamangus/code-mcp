// Package manager handles discovery, cloning, and worktree management for
// multiple Git repositories stored under a single root directory.
//
// Directory layout
//
//	/repos/<name>.git           ← bare clone (default branch)
//	/repos/<name>+<branch>      ← git worktree for <branch>
package manager

import (
	"fmt"
	"log/slog"
	"os"
	"sync"

	"github.com/iamangus/code-mcp/internal/gitops"
)

// RepoInfo describes a synced repository and its worktrees.
type RepoInfo struct {
	Name          string       `json:"name"`
	Dir           string       `json:"dir"`
	DefaultBranch string       `json:"default_branch"`
	Branches      []BranchInfo `json:"branches"`
}

// BranchInfo describes a worktree branch.
type BranchInfo struct {
	Name string `json:"name"`
	Dir  string `json:"dir"`
}

// CommitInfo describes a single commit.
type CommitInfo struct {
	Hash    string `json:"hash"`
	Subject string `json:"subject"`
}

// MergeConflictError is returned when a merge produces conflicts.
type MergeConflictError struct {
	Output string
}

func (e *MergeConflictError) Error() string {
	return fmt.Sprintf("merge conflict:\n%s", e.Output)
}

// Manager manages repositories and worktrees on disk.
type Manager struct {
	reposDir string
	git      gitops.GitOps
	logger   *slog.Logger
	mu       sync.RWMutex
}

// New creates a Manager, creating reposDir if it doesn't exist.
func New(reposDir string, git gitops.GitOps, logger *slog.Logger) (*Manager, error) {
	if err := os.MkdirAll(reposDir, 0755); err != nil {
		return nil, err
	}
	return &Manager{
		reposDir: reposDir,
		git:      git,
		logger:   logger,
	}, nil
}

// ReposDir returns the root directory for all repos.
func (m *Manager) ReposDir() string {
	return m.reposDir
}

// RepoDir returns the filesystem path for a bare repo clone.
func (m *Manager) RepoDir(repo string) string {
	return fmt.Sprintf("%s/%s.git", m.reposDir, repo)
}

// BranchWorktreeDir returns the filesystem path for a branch worktree.
func (m *Manager) BranchWorktreeDir(repo, branch string) string {
	return fmt.Sprintf("%s/%s+%s", m.reposDir, repo, branch)
}
