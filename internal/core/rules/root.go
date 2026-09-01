package rules

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/intentdriven/abcd/internal/gitutil"
)

// ResolveRoot resolves the repo root the per-repo configuration under .abcd/ is
// read from — rules.json and config.json here, guard.json for the shell guard,
// which shares this resolver so a session's rules and its guard can never come
// from two different places.
//
// The resolution is bounded at the git working tree (GHSA-vvqc-3mv2-5p49). The
// toplevel git reports for cwd is resolved FIRST; the walk then runs from cwd
// upward but stops at that toplevel inclusive, so the nearest .abcd directory
// INSIDE the tree wins (a repo that disabled a domain, or the whole loader,
// stays disabled from any nested directory — iss-66/B12) and a tree with no
// .abcd resolves to its own toplevel. A .abcd above the working tree is never
// reached: an unbounded walk let a directory planted in the shared temp dir, or
// the user-scope ~/.abcd for any session under the home directory, govern the
// injected rules and disarm the guard of every repository beneath it. The
// per-repo file is a repo-scope setting by design (the configuration chapter
// and itd-3 both reject a global rules.json), so the bound is the documented
// contract, not a new policy.
//
// When git cannot answer (not a repository, git absent) the root is cwd itself,
// with no walk at all: outside a working tree there is no boundary to stop at,
// and a walk that stops nowhere is the defect. A nested repository or submodule
// therefore resolves its own toplevel and does not inherit the superproject's
// .abcd — correct scoping, visible to submodule workflows.
func ResolveRoot(cwd string) string {
	top, err := gitutil.Run(cwd, "rev-parse", "--show-toplevel")
	if err != nil || top == "" {
		return cwd
	}
	// git reports the physical path; resolve cwd the same way so the ancestor
	// chain from cwd actually passes through top (macOS /var -> /private/var).
	dir, err := filepath.EvalSymlinks(cwd)
	if err != nil {
		dir = filepath.Clean(cwd)
	}
	if real, err := filepath.EvalSymlinks(top); err == nil {
		top = real
	}
	for inside(dir, top) {
		if fi, err := os.Stat(filepath.Join(dir, ".abcd")); err == nil && fi.IsDir() {
			return dir
		}
		if dir == top {
			break
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return top
}

// inside reports whether dir is top or lies beneath it.
func inside(dir, top string) bool {
	if dir == top {
		return true
	}
	prefix := top
	if !strings.HasSuffix(prefix, string(filepath.Separator)) {
		prefix += string(filepath.Separator)
	}
	return strings.HasPrefix(dir, prefix)
}
