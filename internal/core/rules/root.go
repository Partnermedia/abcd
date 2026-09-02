package rules

import (
	"io"
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
// The exact scope of what that closes: any ancestor a session's own working
// tree sits inside, and a plant in a shared directory that is not a repository
// — the one-command `: > /tmp/.git` or `mkdir /tmp/.git` beside a planted
// /tmp/.abcd, which the fallback below refuses because a marker must look like
// a repository. Two residuals stay open, both recorded rather than silently
// assumed shut:
//
//   - iss-2609020219198779 — the user-scope ~/.abcd when the home directory is
//     ITSELF a git working tree (dotfiles-in-home). The toplevel for a session
//     in a non-repo directory beneath such a home is the home, so ~/.abcd
//     governs it. Closing it needs a decision on whether a home-directory
//     toplevel is a legitimate configuration scope (spc-23 plans a user layer
//     that would make it one).
//   - iss-2609020259564193 — a REAL repository planted in a shared ancestor by
//     another uid (`git init /tmp`). It passes every check here, because git's
//     refusal on ownership is the same signal in the attack as in the case the
//     fallback exists for: a legitimate checkout owned by another uid, or a
//     container bind mount. Closing it needs an ownership policy first.
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
// .git marker (gitutil.RepoShapedRoot), and only a directory with no repository
// above it keeps the no-walk cwd: outside a working tree there is no boundary
// to stop at, and a walk that stops nowhere is the defect.
//
// The marker walk is a NAME check that runs to the filesystem root, so it is
// not by itself a bound: an unprivileged `: > /tmp/.git` would otherwise hand
// every session in a plain directory beneath /tmp to a planted /tmp/.abcd, git
// having refused to answer for a directory that is not a repository. The
// fallback root is accepted only when its .git is a PLAUSIBLE repository — a
// directory carrying HEAD, or a regular file beginning "gitdir: "
// (plausibleRepository below, which states the shapes exactly) — and anything
// else takes the non-repo route: cwd, no walk, nothing above it consulted. That
// discriminator is what git reads, not a trust boundary: laying out a real
// repository still passes it, which is residual iss-2609020259564193 above.
//
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
		marker := gitutil.RepoShapedRoot(dir)
		if marker == "" || !plausibleRepository(marker) {
			return cwd // no repository to bound at: no boundary, no walk.
		}
		top = marker
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

// plausibleRepository reports whether root's .git entry has the shape of a
// repository rather than merely carrying the name. It is the discriminator
// ResolveRoot applies to gitutil.RepoShapedRoot's answer; the check lives here
// and not there because RepoShapedRoot's other callers want the crude marker
// (they ask "could anything here be committed?", where a name is enough).
//
// Two shapes pass, the two git itself reads: a .git DIRECTORY carrying HEAD (an
// ordinary checkout — HEAD is the first file git looks for, and no repository
// lacks it), and a regular .git FILE beginning "gitdir: " (a linked worktree or
// a submodule, whose working-tree root is the directory that file sits in).
// Everything else is not a repository for this purpose: an empty file, an empty
// or HEAD-less directory, a dangling symlink, a socket or a device. The .git
// path is stat'd, not lstat'd, so a symlink to a real repository still passes
// and a dangling one does not.
//
// Only the first eight bytes of a .git file are read, so a large file planted
// under the name cannot make the resolver buffer it.
func plausibleRepository(root string) bool {
	marker := filepath.Join(root, ".git")
	fi, err := os.Stat(marker)
	if err != nil {
		return false
	}
	if fi.IsDir() {
		_, err := os.Stat(filepath.Join(marker, "HEAD"))
		return err == nil
	}
	if !fi.Mode().IsRegular() {
		return false
	}
	f, err := os.Open(marker)
	if err != nil {
		return false
	}
	defer f.Close()
	var head [len("gitdir: ")]byte
	if _, err := io.ReadFull(f, head[:]); err != nil {
		return false
	}
	return string(head[:]) == "gitdir: "
}
