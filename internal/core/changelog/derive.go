package changelog

import (
	"strings"

	"github.com/REPPL/abcd-cli/internal/core/launch"
)

// Derivation is the deterministic outcome of a release cut: what the next
// version is, which records decide it, and — when it cannot be decided — why.
//
// A refusal is modelled as a VALUE, not an error, in the same shape as
// launch.RetentionPlan: "this cut cannot be derived" is a legitimate result a
// read-only preview must render, not an exceptional failure. Errors are reserved
// for "the repository could not be read at all".
//
// Three states are deliberately distinguishable, because a caller that conflated
// them would write a wrong CHANGELOG heading:
//
//	Refused          — do not derive; RefusalReason says what to fix.
//	!Refused, !Bumped — nothing to release; write no heading.
//	!Refused, Bumped  — Next/NextTag are the release.
type Derivation struct {
	// Base is the version of the anchor tag the cut is measured from.
	Base launch.Semver
	// BaseTag is Base as a git tag ("v0.3.0"), empty when no anchor resolved.
	BaseTag string
	// Records is the cut: what entered and left the terminal record folders.
	Records RecordSet
	// Bump is the strongest impact in the cut — the judgement that decides the
	// version. ImpactInternal means nothing user-facing shipped.
	Bump Impact
	// Next is the derived version; only meaningful when Bumped.
	Next launch.Semver
	// NextTag is Next as a git tag, empty when nothing is released.
	NextTag string
	// Bumped reports whether the cut moves the version at all.
	Bumped bool
	// Refused reports that the cut must not be derived.
	Refused bool
	// RefusalReason names what to fix; empty unless Refused.
	RefusalReason string
}

// Derive runs the whole deterministic release derivation over the repository at
// root and writes nothing. It is the one composition of this package's parts:
// resolve the anchor tag, refuse a release already in flight, diff the record
// end-states, then apply the version policy to the strongest impact in the cut.
//
// It refuses (rather than deriving a number that would be wrong) in three cases,
// each fail-closed:
//
//   - No release tag. There is no immutable base; inventing one would report
//     every record ever written as this release's contents.
//   - The newest CHANGELOG heading is ahead of the newest tag. auto-release.yml
//     tags AFTER the ship PR merges, so this is the post-merge/pre-tag window:
//     the heading and the tag describe different releases and the base is
//     mismatched. The next cut derives correctly once the tag lands.
//   - A record ADDED by the cut carries no valid impact. An unlabelled record
//     ranks below every real impact, so deriving over it would silently
//     under-bump a release that may contain a break. The lints gate this at the
//     record lifecycle; this is the backstop at the cut, and it names both the
//     record and the file to edit. The removed side cannot refuse — its blob is
//     read from the anchor tag's immutable tree, so an unlabelled one is
//     unfixable by definition (see RecordSet.UnlabelledAdded).
func Derive(root string) (Derivation, error) {
	var d Derivation

	base, hasTag, err := LatestReleaseTag(root)
	if err != nil {
		return Derivation{}, err
	}
	if !hasTag {
		return refuse(d, "no release tag found — a cut needs an immutable base (tag the current release first)"), nil
	}
	d.Base, d.BaseTag = base, base.Tag()

	heading, hasHeading, err := LatestChangelogVersion(root)
	if err != nil {
		return Derivation{}, err
	}
	if hasHeading && launch.CoreGreater(heading, base) {
		return refuse(d, "release "+heading.Tag()+" in flight — tag pending (the newest CHANGELOG heading is ahead of "+d.BaseTag+")"), nil
	}

	records, err := ShippedSince(root, d.BaseTag)
	if err != nil {
		return Derivation{}, err
	}
	d.Records = records

	if unlabelled := records.UnlabelledAdded(); len(unlabelled) > 0 {
		names := make([]string, 0, len(unlabelled))
		for _, rec := range unlabelled {
			names = append(names, rec.ID+" ("+rec.Path+": "+rec.ImpactErr+")")
		}
		return refuse(d, "records added by the cut carry no valid impact: "+strings.Join(names, "; ")), nil
	}

	d.Bump = records.Impact()
	next, bumped := DeriveNext(base, d.Bump)
	if bumped {
		d.Next, d.NextTag, d.Bumped = next, next.Tag(), true
	}
	return d, nil
}

// refuse stamps a refusal on a partially-filled derivation, clearing anything
// that could read as a derived release. Whatever was resolved before the refusal
// (the anchor, the records) is kept, because it is what the operator needs to
// see to fix the cut.
func refuse(d Derivation, reason string) Derivation {
	d.Refused = true
	d.RefusalReason = reason
	d.Bumped = false
	d.Next = launch.Semver{}
	d.NextTag = ""
	return d
}
