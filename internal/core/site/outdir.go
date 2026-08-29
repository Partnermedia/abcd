package site

// The output-directory gate: what `--out` may name, and what makes a directory
// the build's own (GHSA-fpf2-pg82-72rj, CWE-59).
//
// The build EMPTIES its output directory and rewrites it, so the path it is
// handed decides what gets deleted. Two things could make that path lie. A
// symlink — at the leaf, or at any ancestor — redirects every operation on it
// to wherever the link points, and a checkout can carry one as a committed
// file (git mode 120000) named `site`, the default --out. And the ownership
// marker, if it were recognised by name, could be committed too: any directory
// later passed to --out would read as "ours" and have its other entries
// removed — with `--out .`, the working tree and `.git` among them.
//
// resolveOutDir is the one gate. Every touch of the output directory — the
// inspection, the purge, the MkdirAll and the os.Root the writes go through,
// the check's reads and the status board's count — takes the path it returns,
// and the purge additionally requires the ownership rule below. There is
// exactly one place to get this right, and the callers cannot skip it because
// they have no other path.

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/intentdriven/abcd/internal/fsutil"
	"github.com/intentdriven/abcd/internal/gitutil"
)

// The marker that makes the output directory identifiable as this build's own.
//
// The build REWRITES its output tree rather than adding to it, so it has to
// empty the directory first — and emptying a directory that a person named on a
// command line is not something to do on a guess. The marker is the claim: the
// build clears a directory only when it finds its own marker there, and refuses
// a non-empty directory without one.
//
// The marker's content names the REPOSITORY whose build wrote it, keyed on the
// root-commit SHA the way the history store keys its records: a directory that
// another repository's build filled is foreign, so two projects sharing one
// --out cannot purge each other's site. That is a correctness rule, not a
// security boundary — the root commit is public, and a forger copies it as
// easily as the file name — so the content is deliberately not what stands
// between a committed forgery and the purge. What stands there is the ownership
// rule in refuseTrackedOutDir, together with resolveOutDir: a directory git
// tracks anything in, the repository root or an ancestor of it, and a directory
// holding `.git` are never a build output, whatever marker they carry. One
// schema line and one identity line are the whole format; there is nothing
// else a marker could say that the purge should believe.
//
// The name begins with a dot so it reads as tooling metadata rather than
// content. Static-asset hosts conventionally skip dotfiles, but this build does
// not depend on that and does not assert it: the marker names no path, carries
// no secret and describes only itself and the repository's first commit, so a
// host that did serve it would leak nothing. It is excluded from the
// header-coverage walk as build metadata, on the same footing as `_headers` and
// `_redirects`.
const (
	siteMarkerName    = ".abcd-site-build"
	siteMarkerSchema  = "abcd-site-build: 1"
	siteMarkerRepoKey = "repository: "
	siteMarkerBody    = "Output of `abcd site build`. This directory is rewritten on every build;" +
		" the presence of this file is what permits that instead of a refusal.\n"
	maxSiteMarkerBytes = 4 * 1024
)

// siteMarker renders the marker a build of the repository with the given root
// commit writes: the schema line, the identity line, then the explanation.
func siteMarker(rootCommit string) []byte {
	return []byte(siteMarkerSchema + "\n" + siteMarkerRepoKey + markerIdentity(rootCommit) + "\n\n" + siteMarkerBody)
}

// markerIdentity is the repository's name in the marker. A source tree with no
// root commit (no git, none yet) has no identity to write and says so with a
// word — and that word is never ACCEPTED: two identity-less trees would write
// the same marker and purge each other's output, so inspectOutDir classifies a
// non-empty directory as foreign whenever the repository has no root commit,
// and the operator empties it by hand. The word is documentation, not a key.
func markerIdentity(rootCommit string) string {
	if rootCommit == "" {
		return "unknown"
	}
	return rootCommit
}

// markerIdentifies reports whether data is the marker a build of this
// repository writes: the schema line and the identity line, exactly. The
// explanation below them is prose and is not compared.
func markerIdentifies(data []byte, rootCommit string) bool {
	if rootCommit == "" {
		return false
	}
	lines := strings.SplitN(string(data), "\n", 3)
	return len(lines) >= 2 && lines[0] == siteMarkerSchema && lines[1] == siteMarkerRepoKey+markerIdentity(rootCommit)
}

// resolveOutDir is the gate every use of the output directory goes through. It
// returns the absolute path the callers operate on, or the refusal.
//
// The symlink rule is over the WHOLE path, not the leaf: a leaf-only lstat
// cannot see `--out parent-link/site` where `parent-link` is a committed link
// and `site` does not exist yet, which is precisely the shape MkdirAll then
// creates on the far side of the link. Every existing component is inspected.
// A symlink at the leaf is refused outright. A symlink at an ancestor is
// refused when it sits inside the checkout — the git toplevel, not the
// directory the build was invoked from, so a committed link above a
// subdirectory cwd still counts — because the checkout is the one place an
// untrusted commit can plant one; a link the operating system keeps
// outside it (/var -> /private/var on macOS, an automounted home) is followed,
// since refusing it would refuse every temporary directory and protect
// nothing — under the trusted-worktree model, what is outside the checkout is
// the operator's own.
//
// The location rule is over the RESOLVED path: the repository root, any
// ancestor of it, any path with a `.git` component, and any directory holding a
// `.git` entry are refused whatever they contain. The build empties its output
// directory; none of these is ever a build output, and a marker cannot make one
// so.
func resolveOutDir(repoRoot, outDir string) (string, error) {
	if outDir == "" {
		outDir = DefaultOutDir
	}
	abs, err := filepath.Abs(outDir)
	if err != nil {
		return "", err
	}
	repoReal := fsutil.RealExistingPath(checkoutRoot(repoRoot))
	fold := fsutil.CaseFoldingFS()

	if err := refuseSymlinkComponents(abs, repoReal, fold); err != nil {
		return "", err
	}

	outReal := fsutil.RealExistingPath(abs)
	if fsutil.PathWithin(repoReal, outReal, fold) {
		return "", fmt.Errorf("site: %s is the repository root or a directory containing it; the build empties its output directory, so it never writes into one that holds the sources", abs)
	}
	// Folded like every other location rule: on a case-folding filesystem
	// `.GIT` names `.git`.
	dotGit := fsutil.FoldPath(".git", fold)
	for _, seg := range strings.Split(filepath.ToSlash(outReal), "/") {
		if fsutil.FoldPath(seg, fold) == dotGit {
			return "", fmt.Errorf("site: %s is inside a .git directory; refusing", abs)
		}
	}
	if _, err := os.Lstat(filepath.Join(outReal, ".git")); err == nil {
		return "", fmt.Errorf("site: %s holds a .git entry, so it is a repository checkout rather than a build output; refusing to empty it", abs)
	}
	return abs, nil
}

// checkoutRoot is the repository boundary the symlink rule is keyed on: the
// git toplevel of repoRoot, so that a build run from a subdirectory still
// treats a committed link above it as inside the checkout. Where git cannot
// answer (absent, not a repository) repoRoot itself is the boundary — there is
// no checkout above it to defend.
func checkoutRoot(repoRoot string) string {
	top, err := gitutil.Run(repoRoot, "rev-parse", "--path-format=absolute", "--show-toplevel")
	if err != nil || top == "" {
		return repoRoot
	}
	return top
}

// openOutDir is the handle the purge and every write go through. It opens the
// path as it is NOW, and only when it is a real directory: a leaf that became
// a link after the gate ran would otherwise be opened at its target. The
// caller closes it.
func openOutDir(outDir string) (*os.Root, error) {
	if !fsutil.IsRealDir(outDir) {
		return nil, fmt.Errorf("site: %s is not a real directory; refusing to write", outDir)
	}
	return os.OpenRoot(outDir)
}

// refuseSymlinkComponents walks every existing component of abs, shortest
// prefix first, and applies the symlink rule resolveOutDir states. The walk
// stops at the first absent component: everything below it is absent too, and
// there is nothing left to follow.
func refuseSymlinkComponents(abs, repoReal string, fold bool) error {
	var prefixes []string
	for p := abs; ; p = filepath.Dir(p) {
		prefixes = append(prefixes, p)
		if filepath.Dir(p) == p {
			break
		}
	}
	for i := len(prefixes) - 1; i >= 0; i-- {
		p := prefixes[i]
		fi, err := os.Lstat(p)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) || errors.Is(err, syscall.ENOTDIR) {
				return nil
			}
			return err
		}
		if fi.Mode()&os.ModeSymlink == 0 {
			continue
		}
		if p == abs {
			return fmt.Errorf("site: %s is a symlink; the build empties and rewrites its output directory, so it operates only on a real directory and will not follow the link", abs)
		}
		// The link's own real location: its parent is a real directory (or a
		// link outside the repository that was followed above), so resolving
		// the parent and re-adding the name places the link itself.
		linkReal := filepath.Join(fsutil.RealExistingPath(filepath.Dir(p)), filepath.Base(p))
		if fsutil.PathWithin(linkReal, repoReal, fold) {
			return fmt.Errorf("site: %s is a symlink inside the repository, so %s does not name the directory it appears to; refusing to follow it", p, abs)
		}
	}
	return nil
}

// refuseTrackedOutDir is the ownership rule the purge requires beyond the
// marker: a directory git tracks anything in is never a build output. The
// marker is a claim a commit can forge; whether git tracks the directory's
// contents is a fact the same commit cannot hide, and it is what keeps a
// committed marker from turning `--out <tracked directory>` into a purge of
// that directory. The question is asked of whatever repository contains
// outDir, which is the one whose commit could have planted the marker.
//
// It fails closed: a directory that is repo-shaped but that git cannot answer
// for might hold tracked files, and a purge that cannot rule that out does not
// run.
func refuseTrackedOutDir(outDir string) error {
	tracked, err := gitutil.TrackedFiles(outDir)
	if err != nil {
		return fmt.Errorf("site: cannot tell whether git tracks anything in %s (%v); the build empties its output directory only when it can, so refusing", outDir, err)
	}
	if len(tracked) > 0 {
		return fmt.Errorf("site: %s holds %d file(s) git tracks (%s among them); a build output is never tracked, and a %s marker cannot make one so — refusing to empty it", outDir, len(tracked), tracked[0], siteMarkerName)
	}
	return nil
}
