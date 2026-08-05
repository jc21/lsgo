package fsx

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func runGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
}

func requireGit(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git binary not available")
	}
}

func TestGitCacheOutsideRepo(t *testing.T) {
	requireGit(t)
	dir := t.TempDir() // not a git repo

	g := NewGitCache()
	g.Prime([]string{dir})
	if g.HasAnythingFor(dir) {
		t.Error("expected no repository to be found outside a git repo")
	}
	if status := g.Status(dir, true); status != (GitStatus{}) {
		t.Errorf("expected empty status outside a repo, got %+v", status)
	}
}

func TestGitCacheDiscoversRepoAndStatus(t *testing.T) {
	requireGit(t)
	dir := t.TempDir()

	runGit(t, dir, "init", "-q")
	runGit(t, dir, "config", "user.email", "test@example.com")
	runGit(t, dir, "config", "user.name", "Test")

	committed := filepath.Join(dir, "committed.txt")
	must(t, os.WriteFile(committed, []byte("v1"), 0o644))
	runGit(t, dir, "add", "committed.txt")
	runGit(t, dir, "commit", "-q", "-m", "initial")

	// Modify the committed file (unstaged change), and add a new untracked
	// file, to exercise more than one status branch.
	must(t, os.WriteFile(committed, []byte("v2"), 0o644))
	untracked := filepath.Join(dir, "untracked.txt")
	must(t, os.WriteFile(untracked, []byte("new"), 0o644))

	g := NewGitCache()
	g.Prime([]string{dir})
	if !g.HasAnythingFor(dir) {
		t.Fatal("expected a repository to be found")
	}

	modified := g.Status(committed, false)
	if modified.Unstaged != GitModified {
		t.Errorf("expected committed.txt to show as modified, got %+v", modified)
	}

	added := g.Status(untracked, false)
	if added.Unstaged != GitNew {
		t.Errorf("expected untracked.txt to show as new, got %+v", added)
	}

	// A directory-prefix lookup should aggregate the status of files
	// beneath it.
	combined := g.Status(dir, true)
	if combined.Unstaged == GitNotModified {
		t.Errorf("expected prefix lookup over the repo root to find some status, got %+v", combined)
	}

	// A path with no status at all reports the zero value.
	clean := filepath.Join(dir, "does-not-exist.txt")
	if status := g.Status(clean, false); status != (GitStatus{}) {
		t.Errorf("expected empty status for an untouched path, got %+v", status)
	}
}

func TestMergeGitStatus(t *testing.T) {
	cases := []struct {
		a, b, want GitStatus
	}{
		{GitStatus{}, GitStatus{Staged: GitNew, Unstaged: GitModified}, GitStatus{Staged: GitNew, Unstaged: GitModified}},
		{GitStatus{Staged: GitDeleted}, GitStatus{Staged: GitNew}, GitStatus{Staged: GitDeleted}},
		{GitStatus{Unstaged: GitRenamed}, GitStatus{Unstaged: GitModified}, GitStatus{Unstaged: GitRenamed}},
	}
	for _, c := range cases {
		if got := mergeGitStatus(c.a, c.b); got != c.want {
			t.Errorf("mergeGitStatus(%+v, %+v) = %+v, want %+v", c.a, c.b, got, c.want)
		}
	}
}

func TestFileStatusFromChar(t *testing.T) {
	cases := []struct {
		c      byte
		staged bool
		want   GitFileStatus
	}{
		{'A', true, GitNew},
		{'M', true, GitModified},
		{'D', true, GitDeleted},
		{'R', true, GitRenamed},
		{'T', true, GitTypeChange},
		{'?', false, GitNew},
		{'?', true, GitNotModified},
		{'!', true, GitIgnored},
		{'U', true, GitConflicted},
		{' ', true, GitNotModified},
	}
	for _, c := range cases {
		if got := fileStatusFromChar(c.c, c.staged); got != c.want {
			t.Errorf("fileStatusFromChar(%q, %v) = %v, want %v", c.c, c.staged, got, c.want)
		}
	}
}
