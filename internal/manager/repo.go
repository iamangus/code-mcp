package manager

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// SyncRepo clones the repo if it doesn't exist, or fetches if it does.
func (m *Manager) SyncRepo(repoURL, name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	ctx := context.Background()
	repoDir := m.RepoDir(name)

	if _, err := os.Stat(repoDir); err == nil {
		m.logger.Info("repo sync: fetching", "repo", name)
		if err := m.git.Fetch(ctx, repoDir); err != nil {
			m.logger.Warn("repo sync: fetch failed (non-fatal)", "repo", name, "error", err)
		}
		return nil
	}

	m.logger.Info("repo sync: cloning", "repo", name, "url", repoURL)
	return m.git.Clone(ctx, repoURL, repoDir)
}

// RemoveRepo deletes the main clone and all worktree directories.
func (m *Manager) RemoveRepo(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	repoDir := m.RepoDir(name)
	if _, err := os.Stat(repoDir); os.IsNotExist(err) {
		return fmt.Errorf("repo %q not found", name)
	}

	entries, _ := os.ReadDir(m.reposDir)
	prefix := name + "+"
	for _, e := range entries {
		if e.IsDir() && strings.HasPrefix(e.Name(), prefix) {
			wtPath := filepath.Join(m.reposDir, e.Name())
			m.logger.Info("removing worktree dir", "repo", name, "path", wtPath)
			os.RemoveAll(wtPath)
		}
	}

	m.logger.Info("removing repo", "repo", name)
	return os.RemoveAll(repoDir)
}

// Scan discovers all repos and worktrees on disk.
func (m *Manager) Scan() ([]RepoInfo, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.scan()
}

func (m *Manager) scan() ([]RepoInfo, error) {
	entries, err := os.ReadDir(m.reposDir)
	if err != nil {
		return nil, err
	}

	ctx := context.Background()
	var repos []RepoInfo

	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		name := e.Name()
		if !strings.HasSuffix(name, ".git") {
			continue
		}
		repoName := strings.TrimSuffix(name, ".git")
		repoDir := filepath.Join(m.reposDir, name)

		defaultBranch, err := m.git.DefaultBranch(ctx, repoDir)
		if err != nil {
			m.logger.Warn("scan: could not determine default branch", "repo", repoName, "error", err)
			defaultBranch = "main"
		}

		branches := m.listWorktrees(repoName, defaultBranch)

		repos = append(repos, RepoInfo{
			Name:          repoName,
			Dir:           repoDir,
			DefaultBranch: defaultBranch,
			Branches:      branches,
		})
	}
	return repos, nil
}

func (m *Manager) listWorktrees(repoName, defaultBranch string) []BranchInfo {
	entries, err := os.ReadDir(m.reposDir)
	if err != nil {
		return nil
	}

	prefix := repoName + "+"
	var branches []BranchInfo

	for _, e := range entries {
		if !e.IsDir() || !strings.HasPrefix(e.Name(), prefix) {
			continue
		}
		branchName := strings.TrimPrefix(e.Name(), prefix)
		branches = append(branches, BranchInfo{
			Name: branchName,
			Dir:  filepath.Join(m.reposDir, e.Name()),
		})
	}
	return branches
}
