package manager

import (
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	"github.com/iamangus/code-mcp/internal/gitops"
)

func newTestManager(t *testing.T) (*Manager, *gitops.Fake) {
	t.Helper()
	dir := t.TempDir()
	fake := gitops.NewFake()
	mgr, err := New(dir, fake, slog.Default())
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	return mgr, fake
}

// createFakeRepo creates the bare repo directory on disk so the manager finds it.
func createFakeRepo(t *testing.T, mgr *Manager, name string) string {
	t.Helper()
	repoDir := mgr.RepoDir(name)
	if err := os.MkdirAll(repoDir, 0755); err != nil {
		t.Fatal(err)
	}
	return repoDir
}

func TestNew_CreatesMissingDir(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "sub", "repos")
	fake := gitops.NewFake()
	_, err := New(dir, fake, slog.Default())
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if _, err := os.Stat(dir); err != nil {
		t.Fatalf("dir not created: %v", err)
	}
}

func TestRepoDir(t *testing.T) {
	mgr, _ := newTestManager(t)
	got := mgr.RepoDir("myapp")
	want := filepath.Join(mgr.ReposDir(), "myapp.git")
	if got != want {
		t.Errorf("RepoDir = %q, want %q", got, want)
	}
}

func TestBranchWorktreeDir(t *testing.T) {
	mgr, _ := newTestManager(t)
	got := mgr.BranchWorktreeDir("myapp", "feature")
	want := filepath.Join(mgr.ReposDir(), "myapp+feature")
	if got != want {
		t.Errorf("BranchWorktreeDir = %q, want %q", got, want)
	}
}

func TestSyncRepo_Clone(t *testing.T) {
	mgr, fake := newTestManager(t)
	if err := mgr.SyncRepo("https://github.com/test/repo.git", "repo"); err != nil {
		t.Fatalf("SyncRepo: %v", err)
	}
	if !fake.HasCall("Clone") {
		t.Error("expected Clone to be called")
	}
}

func TestSyncRepo_FetchExisting(t *testing.T) {
	mgr, fake := newTestManager(t)
	createFakeRepo(t, mgr, "repo")
	if err := mgr.SyncRepo("https://github.com/test/repo.git", "repo"); err != nil {
		t.Fatalf("SyncRepo: %v", err)
	}
	if !fake.HasCall("Fetch") {
		t.Error("expected Fetch to be called")
	}
	if fake.HasCall("Clone") {
		t.Error("should not Clone existing repo")
	}
}

func TestRemoveRepo(t *testing.T) {
	mgr, _ := newTestManager(t)
	createFakeRepo(t, mgr, "repo")
	wtDir := mgr.BranchWorktreeDir("repo", "feat")
	os.MkdirAll(wtDir, 0755)

	if err := mgr.RemoveRepo("repo"); err != nil {
		t.Fatalf("RemoveRepo: %v", err)
	}
	if _, err := os.Stat(mgr.RepoDir("repo")); !os.IsNotExist(err) {
		t.Error("repo dir should be removed")
	}
	if _, err := os.Stat(wtDir); !os.IsNotExist(err) {
		t.Error("worktree dir should be removed")
	}
}

func TestRemoveRepo_NotFound(t *testing.T) {
	mgr, _ := newTestManager(t)
	if err := mgr.RemoveRepo("nope"); err == nil {
		t.Fatal("expected error")
	}
}

func TestCreateWorktree_NewBranch(t *testing.T) {
	mgr, fake := newTestManager(t)
	createFakeRepo(t, mgr, "repo")
	fake.BoolReturns["BranchExists"] = false

	dir, err := mgr.CreateWorktree("repo", "feature", "")
	if err != nil {
		t.Fatalf("CreateWorktree: %v", err)
	}
	if dir == "" {
		t.Error("expected non-empty dir")
	}
	if !fake.HasCall("CreateBranch") {
		t.Error("expected CreateBranch for new branch")
	}
	if !fake.HasCall("WorktreeAdd") {
		t.Error("expected WorktreeAdd")
	}
}

func TestCreateWorktree_DefaultBranch(t *testing.T) {
	mgr, fake := newTestManager(t)
	repoDir := createFakeRepo(t, mgr, "repo")
	fake.StringReturns["DefaultBranch"] = "main"

	dir, err := mgr.CreateWorktree("repo", "main", "")
	if err != nil {
		t.Fatalf("CreateWorktree: %v", err)
	}
	if dir != repoDir {
		t.Errorf("expected repo dir %q for default branch, got %q", repoDir, dir)
	}
	if fake.HasCall("WorktreeAdd") {
		t.Error("should not call WorktreeAdd for default branch")
	}
}

func TestCreateWorktree_InvalidBranch(t *testing.T) {
	mgr, _ := newTestManager(t)
	createFakeRepo(t, mgr, "repo")
	if _, err := mgr.CreateWorktree("repo", "bad branch!", ""); err == nil {
		t.Fatal("expected error for invalid branch name")
	}
}

func TestCreateWorktree_AlreadyExists(t *testing.T) {
	mgr, _ := newTestManager(t)
	createFakeRepo(t, mgr, "repo")
	wtDir := mgr.BranchWorktreeDir("repo", "feat")
	os.MkdirAll(wtDir, 0755)

	dir, err := mgr.CreateWorktree("repo", "feat", "")
	if err != nil {
		t.Fatalf("CreateWorktree: %v", err)
	}
	if dir == "" {
		t.Error("expected non-empty dir for existing worktree")
	}
}

func TestMergeBranch_Conflict(t *testing.T) {
	mgr, fake := newTestManager(t)
	createFakeRepo(t, mgr, "repo")
	fake.StringReturns["DefaultBranch"] = "main"
	fake.Errors["Merge"] = &MergeConflictError{Output: "CONFLICT (content): Merge conflict in file.go"}

	targetDir := mgr.BranchWorktreeDir("repo", "main")
	os.MkdirAll(targetDir, 0755)

	err := mgr.MergeBranch("repo", "feature", "main")
	if err == nil {
		t.Fatal("expected merge conflict error")
	}
	if _, ok := err.(*MergeConflictError); !ok {
		t.Errorf("expected MergeConflictError, got %T: %v", err, err)
	}
}

func TestGetCommits(t *testing.T) {
	mgr, fake := newTestManager(t)
	createFakeRepo(t, mgr, "repo")
	fake.StringReturns["DefaultBranch"] = "main"
	fake.StringReturns["CommitLog"] = "abc1234 first commit\ndef5678 second commit"

	wtDir := mgr.BranchWorktreeDir("repo", "feature")
	os.MkdirAll(wtDir, 0755)

	commits, err := mgr.GetCommits("repo", "feature")
	if err != nil {
		t.Fatalf("GetCommits: %v", err)
	}
	if len(commits) != 2 {
		t.Fatalf("expected 2 commits, got %d", len(commits))
	}
	if commits[0].Hash != "abc1234" {
		t.Errorf("commit[0].Hash = %q, want abc1234", commits[0].Hash)
	}
	if commits[0].Subject != "first commit" {
		t.Errorf("commit[0].Subject = %q, want 'first commit'", commits[0].Subject)
	}
}

func TestScan_EmptyDir(t *testing.T) {
	mgr, _ := newTestManager(t)
	repos, err := mgr.Scan()
	if err != nil {
		t.Fatalf("Scan: %v", err)
	}
	if len(repos) != 0 {
		t.Errorf("expected 0 repos, got %d", len(repos))
	}
}

func TestScan_DiscoverRepo(t *testing.T) {
	mgr, fake := newTestManager(t)
	createFakeRepo(t, mgr, "myapp")
	fake.StringReturns["DefaultBranch"] = "main"

	repos, err := mgr.Scan()
	if err != nil {
		t.Fatalf("Scan: %v", err)
	}
	if len(repos) != 1 {
		t.Fatalf("expected 1 repo, got %d", len(repos))
	}
	if repos[0].Name != "myapp" {
		t.Errorf("repo name = %q, want myapp", repos[0].Name)
	}
}
