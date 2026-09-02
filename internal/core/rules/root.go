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
// reached: an unbounded walk let a directory planted in the shared temp dir
// govern the injected rules and disarm the guard of every repository beneath
// it. The per-repo file is a repo-scope setting by design (the configuration
// chapter and itd-3 both reject a global rules.json), so the bound is the
// documented contract, not a new policy.
//
// The exact scope of what that closes: a plant in a shared temp directory, and
// any ancestor a session's own working tree sits inside. It does NOT close the
// user-scope ~/.abcd when the home directory is ITSELF a git working tree
// (dotfiles-in-home) — the toplevel for a session in a non-repo directory
// beneath such a home is the home, so ~/.abcd governs it. That residual is
// iss-2609020219198779: closing it needs a decision on whether a home-directory
// toplevel is a legitimate configuration scope (spc-23 plans a user layer that
// would make it one), so it is recorded rather than silently assumed shut.
//
// "Not a repository" and "a repository git will not answer for" are DIFFERENT
// outcomes and only the first resolves to cwd with no walk. abcd runs git under
// an isolated environment (GIT_CONFIG_GLOBAL=/dev/null, GIT_CONFIG_NOSYSTEM=1)
// which also discards the developer's safe.directory exceptions, so rev-parse
// fails inside a checkout owned by another uid; it fails outright when the host
// launches the hook with git off its PATH. Collapsing those onto cwd dropped
// the repository's own configuration layer with no error anywhere: from a
// subdirectory the root became the subdirectory, so guard.json's hazards
// became allows, the rules kill switch stopped applying, and the private
// banlist store went missing. A repo-shaped tree therefore falls back to the
// .git marker (gitutil.RepoShapedRoot) and is bounded there, and only a
// genuinely non-repo directory keeps the no-walk cwd: outside a working tree
// there is no boundary to stop at, and a walk that stops nowhere is the defect.
// A nested repository or submodule resolves its own toplevel either way and
// does not inherit the superproject's .abcd — correct scoping, visible to
// submodule workflows.
func ResolveRoot(cwd string) string {
	// git reports the physical path; resolve cwd the same way so the ancestor
	// chain from cwd actually passes through top (macOS /var -> /private/var).
	dir, err := filepath.EvalSymlinks(cwd)
	if err != nil {
		dir = filepath.Clean(cwd)
	}
	top, err := gitutil.Run(cwd, "rev-parse", "--show-toplevel")
	if err != nil || top == "" {
		// The marker walk runs on the symlink-resolved path, so the bound it
		// returns is already on the same chain the walk below climbs.
		top = gitutil.RepoShapedRoot(dir)
		if top == "" {
			return cwd // genuinely not a repository: no boundary, no walk.
		}
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
