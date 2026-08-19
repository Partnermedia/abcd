// Package recordid is abcd's canonical support for minting sequential record
// ids (iss-N, itd-N, spc-N) that do not collide across parallel branches.
//
// Every ledger mints "max observed N, plus one". Historically each ledger looked
// only at ITS OWN WORKING TREE, so two branches cut from the same base silently
// minted the same next id — invisible on each branch, surfacing only at merge
// (iss-115, iss-120: two programmes both minted spc-10/spc-11, iss-110/iss-111).
//
// MaxAcrossRefs is the one home for the git-ref side of that maximum: it scans
// every local branch and remote-tracking ref for a family's highest id, so a
// branch that has ALREADY committed a higher id is seen even though its file is
// absent from the current working tree. Each ledger keeps its own working-tree
// scan (which alone sees uncommitted mints) and folds this in — union, then +1.
//
// RESIDUAL WINDOW (accepted, documented): a ref only carries a mint once it is
// committed. Two branches that BOTH mint before either commits still collide —
// the ref scan cannot see an uncommitted file on another branch. That window is
// left to the armed record-lint uniqueness rules (issue_id_unique,
// intent_lifecycle, spec_id_unique), which run on the merged PR's union in CI and
// flag every colliding claimant. Refs-union minting closes the common
// commit-then-branch case; the detectors close the rest.
package recordid

import (
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/Partnermedia/abcd/internal/gitutil"
)

// refScanMaxBytes caps the stdout buffered from a single git query. A hostile or
// pathological repository cannot make one for-each-ref / ls-tree exhaust memory;
// a truncated tail can only UNDER-count the max, which the loud degrade covers.
const refScanMaxBytes = 8 << 20 // 8 MiB

// maxRefs bounds how many distinct ref tips the scan visits, keeping the cost
// O(branches) even in a repo with a pathological number of refs. Hitting the cap
// degrades loudly rather than scanning unboundedly.
const maxRefs = 4096

// RefScan is the outcome of scanning git refs for a family's highest id N.
//
// Max is the best-effort highest N observed (0 when none, or when the scan could
// not run). Degraded is true when a repository is present but git could not read
// it (git absent/broken over a real repo) or a ref query failed — in which case
// Max reflects only what was seen before the failure and the caller mints from
// the working tree essentially alone, so it MUST surface Warning() rather than
// fall back silently. A directory that is simply NOT a repository is not degraded
// (there are no refs to collide against): Max is 0 and Warning() is empty.
type RefScan struct {
	Max      int
	Degraded bool
	Reason   string
}

// Warning renders the loud-degrade note for a surface to print (stderr / result
// envelope), or "" when the scan completed. It never leaks a path — Reason is a
// short, path-free cause.
func (s RefScan) Warning() string {
	if !s.Degraded {
		return ""
	}
	return "record-id minting saw the working tree only (" + s.Reason +
		"): a parallel branch that has already committed a higher id can collide, " +
		"and record-lint on the merged PR is the backstop"
}

// MaxAcrossRefs returns the highest N in filenames of the form
// <prefix>-<N>[-<slug>].md under any of dirs (repo-relative, slash-separated),
// across every local branch (refs/heads) and remote-tracking ref (refs/remotes)
// of the repository at repoRoot.
//
// It NEVER returns an error: a git-absent host, a repoRoot outside a work tree,
// or a failed ref query degrade LOUDLY (Degraded=true, a path-free Reason) and
// the best-effort Max seen so far is returned. The caller then mints
// max(workingTreeMax, RefScan.Max)+1 and surfaces Warning() — so a degraded scan
// never silently narrows the id space it can see.
//
// Only branch TIPS are scanned (a committed id lives in the tip's tree), so the
// cost is O(distinct branch tips), not O(history). Uncommitted files on other
// branches are deliberately out of scope — see the package doc's residual window.
func MaxAcrossRefs(repoRoot, prefix string, dirs []string) RefScan {
	if len(dirs) == 0 {
		return RefScan{}
	}
	if !gitutil.InRepo(repoRoot) {
		// A directory with no repository at all has no refs to collide against, so
		// working-tree-only minting is COMPLETE, not degraded — stay quiet (the
		// common non-repo / test case). But a .git that git could NOT read (git
		// absent or broken over a real repo that may hold parallel branches) is the
		// dangerous silent-collision case the loud degrade exists to catch: warn.
		if gitDirPresent(repoRoot) {
			return RefScan{Degraded: true, Reason: "git could not read the repository (git absent or unreadable)"}
		}
		return RefScan{}
	}

	out, err := gitutil.RunLimited(repoRoot, refScanMaxBytes,
		"for-each-ref", "--format=%(objectname)", "refs/heads/", "refs/remotes/")
	if err != nil {
		return RefScan{Degraded: true, Reason: "git for-each-ref failed"}
	}

	re := idRe(prefix)
	scan := RefScan{}
	seen := map[string]bool{}
	count := 0
	for _, obj := range strings.Fields(out) {
		if seen[obj] {
			continue
		}
		seen[obj] = true
		count++
		if count > maxRefs {
			scan.Degraded = true
			scan.Reason = "more than " + strconv.Itoa(maxRefs) + " refs; scan truncated"
			break
		}
		args := make([]string, 0, 5+len(dirs))
		args = append(args, "ls-tree", "-r", "--name-only", obj, "--")
		args = append(args, dirs...)
		treeOut, terr := gitutil.RunLimited(repoRoot, refScanMaxBytes, args...)
		if terr != nil {
			// A single bad ref (e.g. a tag ref that is not a commit, an object-store
			// race) degrades loudly but does not abort the scan: the maxima from the
			// other refs still tighten the floor.
			scan.Degraded = true
			scan.Reason = "git ls-tree failed for a ref"
			continue
		}
		for _, line := range strings.Split(treeOut, "\n") {
			if line == "" {
				continue
			}
			m := re.FindStringSubmatch(path.Base(line))
			if m == nil {
				continue
			}
			n, err := strconv.Atoi(m[1])
			if err != nil {
				continue
			}
			if n > scan.Max {
				scan.Max = n
			}
		}
	}
	return scan
}

// gitDirPresent reports whether repoRoot carries a .git entry (a directory for a
// normal clone, a file for a worktree/submodule). It distinguishes "no repo here,
// nothing to scan" from "a repo is here but git could not read it", so the loud
// degrade fires only for the latter.
func gitDirPresent(repoRoot string) bool {
	_, err := os.Stat(filepath.Join(repoRoot, ".git"))
	return err == nil
}

// idRe matches a record filename <prefix>-<N>[-<slug>].md and captures N. The
// prefix is a fixed family tag (iss/itd/spc); it is regexp-quoted defensively so
// a caller can never inject metacharacters through it.
func idRe(prefix string) *regexp.Regexp {
	return regexp.MustCompile(`^` + regexp.QuoteMeta(prefix) + `-([0-9]+)(?:-[a-z0-9-]+)?\.md$`)
}
