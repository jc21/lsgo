package fsx

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
)

// GitFileStatus is a single file's status within a Git repository, as
// shown in one of the two Git columns in the long view.
type GitFileStatus int

const (
	GitNotModified GitFileStatus = iota
	GitNew
	GitModified
	GitDeleted
	GitRenamed
	GitTypeChange
	GitIgnored
	GitConflicted
)

// GitStatus is a file's complete Git status: separate flags for the
// staged (index) and unstaged (working tree) state, since a file can have
// changes in both at once.
type GitStatus struct {
	Staged   GitFileStatus
	Unstaged GitFileStatus
}

// GitCache discovers and queries Git repositories for the paths lsgo is
// about to list. Repository status is looked up lazily and cached, since
// running "git status" is the most expensive part of a --git listing.
//
// Rather than linking against libgit2, this shells out to the git binary,
// which keeps lsgo a dependency-free build while still giving accurate,
// porcelain-compatible results.
type GitCache struct {
	mu    sync.Mutex
	repos map[string]*gitRepo // keyed by repo root
	// misses records paths already confirmed to have no repository, so we
	// don't repeat the (relatively expensive) discovery walk.
	misses map[string]bool
}

type gitRepo struct {
	root string

	once     sync.Once
	statuses map[string]GitStatus // path -> status, keyed by absolute path
}

// NewGitCache builds an (initially empty) cache. Repositories are
// discovered on demand as paths are queried.
func NewGitCache() *GitCache {
	return &GitCache{
		repos:  make(map[string]*gitRepo),
		misses: make(map[string]bool),
	}
}

// Prime discovers the Git repositories (if any) containing the given
// starting paths, so that later Status/HasAnythingFor calls are cheap.
func (g *GitCache) Prime(paths []string) {
	for _, p := range paths {
		g.repoFor(p)
	}
}

// HasAnythingFor reports whether a repository is known to cover the given
// path, used to decide whether it's worth adding a Git column at all.
func (g *GitCache) HasAnythingFor(path string) bool {
	return g.repoFor(path) != nil
}

// Status returns the Git status for a file or directory. When
// prefixLookup is true (used for directories), the result aggregates the
// status of every tracked path underneath it.
func (g *GitCache) Status(path string, prefixLookup bool) GitStatus {
	repo := g.repoFor(path)
	if repo == nil {
		return GitStatus{}
	}

	abs, err := filepath.Abs(path)
	if err != nil {
		return GitStatus{}
	}
	abs = filepath.Clean(abs)

	repo.load()

	if !prefixLookup {
		return repo.statuses[abs]
	}

	var combined GitStatus
	prefix := abs + string(filepath.Separator)
	for p, s := range repo.statuses {
		if p == abs || strings.HasPrefix(p, prefix) {
			combined = mergeGitStatus(combined, s)
		}
	}
	return combined
}

func mergeGitStatus(a, b GitStatus) GitStatus {
	if a.Staged == GitNotModified {
		a.Staged = b.Staged
	}
	if a.Unstaged == GitNotModified {
		a.Unstaged = b.Unstaged
	}
	return a
}

// repoFor finds (discovering and caching as necessary) the repository that
// owns path, or nil if it isn't in one.
func (g *GitCache) repoFor(path string) *gitRepo {
	abs, err := filepath.Abs(path)
	if err != nil {
		return nil
	}

	g.mu.Lock()
	defer g.mu.Unlock()

	for root, repo := range g.repos {
		if abs == root || strings.HasPrefix(abs, root+string(filepath.Separator)) {
			return repo
		}
	}
	if g.misses[abs] {
		return nil
	}

	root, ok := discoverRepoRoot(abs)
	if !ok {
		g.misses[abs] = true
		return nil
	}

	if repo, exists := g.repos[root]; exists {
		return repo
	}

	repo := &gitRepo{root: root}
	g.repos[root] = repo
	return repo
}

// discoverRepoRoot runs "git rev-parse --show-toplevel" from the directory
// containing path, returning the repository's working directory.
func discoverRepoRoot(path string) (string, bool) {
	dir := path
	if info, err := os.Stat(path); err == nil && !info.IsDir() {
		dir = filepath.Dir(path)
	}

	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		return "", false
	}

	root := strings.TrimSpace(string(out))
	if root == "" {
		return "", false
	}
	return filepath.Clean(root), true
}

// load runs "git status" once for the whole repository and parses its
// porcelain output into a path -> status map.
func (r *gitRepo) load() {
	r.once.Do(func() {
		r.statuses = make(map[string]GitStatus)

		cmd := exec.Command("git", "status", "--porcelain=v1", "--ignored=matching", "-z")
		cmd.Dir = r.root
		out, err := cmd.Output()
		if err != nil {
			return
		}

		for _, entry := range strings.Split(string(out), "\x00") {
			if len(entry) < 4 {
				continue
			}

			indexStatus := entry[0]
			workStatus := entry[1]
			name := entry[3:]

			// Renames report "old -> new"; git -z actually emits the old
			// path as a second NUL-separated field, but we only need the
			// new name here, so take the text after the last path
			// separator marker if present.
			if idx := strings.Index(name, "\x00"); idx >= 0 {
				name = name[idx+1:]
			}

			abs := filepath.Clean(filepath.Join(r.root, name))
			r.statuses[abs] = GitStatus{
				Staged:   fileStatusFromChar(indexStatus, true),
				Unstaged: fileStatusFromChar(workStatus, false),
			}
		}
	})
}

func fileStatusFromChar(c byte, staged bool) GitFileStatus {
	switch c {
	case 'A':
		return GitNew
	case 'M':
		return GitModified
	case 'D':
		return GitDeleted
	case 'R':
		return GitRenamed
	case 'T':
		return GitTypeChange
	case '?':
		if !staged {
			return GitNew
		}
		return GitNotModified
	case '!':
		return GitIgnored
	case 'U':
		return GitConflicted
	default:
		return GitNotModified
	}
}
