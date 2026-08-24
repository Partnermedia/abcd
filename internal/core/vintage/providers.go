package vintage

import "github.com/intentdriven/abcd/internal/gitutil"

// checkoutTipProvider is the dogfood disk provider: the expected vintage is the
// tip of the source checkout the binary was built from.
type checkoutTipProvider struct {
	repoRoot   string
	currentRev string
}

// CheckoutTip is the dogfood disk provider. The expected vintage is the source
// checkout's HEAD, but only when the running binary's embedded revision is an
// ancestor of it: that both proves the binary came from THIS checkout (a random
// cwd is not abcd's source) and fixes the direction — an ancestor is behind or
// equal, which is the only case where "the binary trails the tip" is a truthful
// claim. A revision that is ahead, divergent, or foreign to the repo leaves the
// reference unresolved, so the comparator reports unknown rather than guessing a
// direction it cannot see (the same non-directional caution the retired skew
// notice embodied).
func CheckoutTip(repoRoot, currentRev string) Provider {
	return checkoutTipProvider{repoRoot: repoRoot, currentRev: currentRev}
}

func (p checkoutTipProvider) Expected() Expected {
	if p.currentRev == "" || p.repoRoot == "" {
		return Expected{Known: false}
	}
	// gitutil.Run isolates the command from an inherited GIT_DIR/GIT_WORK_TREE,
	// so a hostile environment cannot redirect the tip read at another repo.
	head, err := gitutil.Run(p.repoRoot, "rev-parse", "HEAD")
	if err != nil || head == "" {
		return Expected{Known: false}
	}
	// --is-ancestor exits zero only when currentRev is an ancestor of (or equal
	// to) HEAD. Any non-zero exit — not an ancestor, or the revision is not in
	// the repo at all — collapses to an unresolved reference: both mean "do not
	// assert staleness here".
	if _, err := gitutil.Run(p.repoRoot, "merge-base", "--is-ancestor", p.currentRev, head); err != nil {
		return Expected{Known: false}
	}
	return Expected{Revision: head, Known: true}
}

// pinnedVersionProvider is the everywhere-else disk provider: the expected
// vintage is the release version the plugin cache pinned.
type pinnedVersionProvider struct{ tag string }

// PinnedVersion is the disk provider for a pinned install: the expected vintage
// is the release tag the plugin-cache manifest recorded, compared against the
// binary's stamped version. An empty tag is an unresolved reference (unknown),
// never a silent match.
func PinnedVersion(pinnedTag string) Provider {
	return pinnedVersionProvider{tag: pinnedTag}
}

func (p pinnedVersionProvider) Expected() Expected {
	if p.tag == "" {
		return Expected{Known: false}
	}
	return Expected{Revision: p.tag, Known: true}
}
