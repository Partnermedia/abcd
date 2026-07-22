package changelog

import "github.com/REPPL/abcd-cli/internal/core/launch"

// DeriveNext is the whole version policy in one pure function: given the
// previous release and the strongest impact in the cut, it returns the next
// version. It is deliberately free of git, files, and clocks so the policy can
// be walked cell by cell in a table test rather than inferred from a pipeline.
//
// The arithmetic is over launch.Semver — the repo's one SemVer type — and the
// result is always a release CORE: any prerelease/build metadata on prev is
// dropped, because the derived number becomes a git tag and a manifest value,
// and carrying "rc1" forward would publish a release that claims to be a
// pre-release of itself.
//
// The policy, per plan §3:
//
//	prev >= 1.0.0   breaking -> major++, minor = patch = 0
//	                additive -> minor++, patch = 0
//	                fix      -> patch++
//	prev is 0.x     breaking -> minor++, patch = 0
//	                additive -> patch++
//	                fix      -> patch++
//
// The pre-1.0 row is load-bearing, not a shortcut. While abcd is at 0.x it has
// declared no stable surface, so a break bumps the minor (ADR-37: "pre-1.0, a
// minor may break, called out under Breaking"). The consequence is deliberate:
// NO input can derive 1.0.0 from a 0.x base. The first 1.0.0 is a human's
// explicit override, because declaring stability is a product decision no set of
// records can make on the maintainer's behalf.
//
// The second return is the "nothing to release" signal, and it is a distinct
// value rather than a silent `return prev`: an all-internal or empty cut moves
// no version, and a caller that could not tell the two apart would write a
// duplicate dated heading — which auto-release.yml would then try to tag against
// an already-tagged version.
func DeriveNext(prev launch.Semver, bump Impact) (launch.Semver, bool) {
	if !bump.DrivesBump() {
		return launch.Semver{}, false
	}
	// Copy the core only: Prerelease/Build are deliberately left zero.
	next := launch.Semver{Major: prev.Major, Minor: prev.Minor, Patch: prev.Patch}
	preStable := prev.Major == 0
	switch {
	case bump == ImpactBreaking && preStable:
		next.Minor++
		next.Patch = 0
	case bump == ImpactBreaking:
		next.Major++
		next.Minor, next.Patch = 0, 0
	case bump == ImpactAdditive && preStable:
		next.Patch++
	case bump == ImpactAdditive:
		next.Minor++
		next.Patch = 0
	default: // ImpactFix, the only remaining bump-driving member.
		next.Patch++
	}
	return next, true
}
