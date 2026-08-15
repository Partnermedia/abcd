package ahoy

import (
	"path/filepath"
	"strings"

	"github.com/REPPL/abcd-cli/internal/core"
	"github.com/REPPL/abcd-cli/internal/core/vintage"
	"github.com/REPPL/abcd-cli/internal/fsutil"
	"github.com/REPPL/abcd-cli/internal/gitutil"
)

// VintageStatus is the assembled vintage picture for a repo: the install mode,
// the comparator's report, and the human name of the reference compared against.
// It is the single source the `version` and `ahoy` renders, the session-start
// notice, and the install refusal all consume — one comparison, many renders.
type VintageStatus struct {
	Mode     string         // install mode ("dev (tip build)" / "pinned" / "" / shadowed)
	Report   vintage.Report // outcome + current + expected
	Source   string         // reference name: "checkout tip" / "plugin manifest pin"; "" when the current vintage is undeterminable
	RepoRoot string         // the source checkout root, when the tip was the reference
}

// Vintage assembles the running binary's vintage picture for cwd. It resolves the
// install mode and the plugin root, reads the binary's own build vintage, and
// picks the applicable disk reference: the source checkout tip when cwd is
// abcd's own checkout (the dogfood case), otherwise the plugin-cache manifest
// pin. It never touches the network.
func Vintage(cwd string) VintageStatus {
	abs, err := filepath.Abs(cwd)
	if err != nil {
		abs = cwd
	}
	pluginRoot, pluginOK := resolvePluginRoot()
	mode := detectInstallMode(pluginRoot, pluginOK)
	pinTag := ""
	if pluginOK {
		pinTag = readPinnedTag(pluginRoot)
	}
	return vintageFrom(vintage.CurrentBuildVintage(), mode, abs, core.Version, pinTag)
}

// vintageFrom is the pure assembly, split from Vintage so its branch selection
// can be exercised with an explicit current vintage, a fixture checkout, and an
// explicit version/pin — no rebuild of the test binary required.
func vintageFrom(cur vintage.Current, mode, cwd, version, pinTag string) VintageStatus {
	// An undeterminable current vintage is terminal: no reference can make an
	// unknown binary fresh, so report it directly. The rebuild fix — not a disk
	// reference — is the out, so no Source is named.
	if !cur.Known {
		return VintageStatus{Mode: mode, Report: vintage.Report{Outcome: vintage.Unknown, Current: cur.Revision}}
	}
	// Dogfood: the embedded revision against the source checkout tip. The
	// provider self-verifies (by ancestry) that cwd is abcd's own checkout, so a
	// non-dogfood cwd yields Unknown here and falls through to the pinned
	// comparison below rather than reporting a spurious stale.
	if rep := vintage.Compare(cur, vintage.CheckoutTip(cwd, cur.Revision)); rep.Outcome != vintage.Unknown {
		return VintageStatus{Mode: mode, Report: rep, Source: "checkout tip", RepoRoot: cwd}
	}
	// Everywhere else: the stamped version against the plugin-cache manifest pin.
	// A "dev"/empty version is itself undeterminable as a pinned vintage.
	pinCur := vintage.Current{Revision: version, Known: version != "" && version != "dev"}
	return VintageStatus{
		Mode:   mode,
		Report: vintage.Compare(pinCur, vintage.PinnedVersion(pinTag)),
		Source: "plugin manifest pin",
	}
}

// DogfoodReport is the source-checkout-tip comparison the session-start notice
// consumes. It is scoped to the dogfood case on purpose: a pinned install in a
// user's own repo must never be nagged, so IsDogfood gates the whole notice.
type DogfoodReport struct {
	IsDogfood bool            // cwd is abcd's own source checkout (the binary was built from it)
	Outcome   vintage.Outcome // Fresh (silent) / Stale / Unknown (a dirty rebuild)
	Revision  string          // the binary's embedded revision (full)
	Tip       string          // the checkout tip it should match (full)
}

// DogfoodStaleness reports whether the running binary trails the tip of the
// source checkout it was built from, for the session-start notice. It reads the
// binary's own build vintage and compares against cwd's HEAD.
func DogfoodStaleness(cwd string) DogfoodReport {
	return dogfoodStalenessFrom(vintage.CurrentBuildVintage(), cwd)
}

// dogfoodStalenessFrom is the pure core, split so the fresh/stale/unknown and
// not-dogfood branches can be exercised with an explicit current vintage against
// a fixture checkout. IsDogfood is true only when the binary's embedded revision
// is contained in cwd's history — proof cwd is abcd's own checkout — so a pinned
// install elsewhere yields no notice.
func dogfoodStalenessFrom(cur vintage.Current, cwd string) DogfoodReport {
	if cur.Revision == "" {
		return DogfoodReport{} // no embedded revision: nothing to locate against a tip
	}
	abs, err := filepath.Abs(cwd)
	if err != nil {
		abs = cwd
	}
	head, err := gitutil.Run(abs, "rev-parse", "HEAD")
	if err != nil || head == "" {
		return DogfoodReport{}
	}
	// The binary's revision must be a commit in this checkout for the comparison
	// to be about THIS checkout at all; --is-ancestor also fixes the direction.
	if _, err := gitutil.Run(abs, "merge-base", "--is-ancestor", cur.Revision, head); err != nil {
		return DogfoodReport{}
	}
	r := DogfoodReport{IsDogfood: true, Revision: cur.Revision, Tip: head}
	switch {
	case !cur.Known:
		// A dirty rebuild atop a real commit: the revision no longer identifies
		// what is running, so the vintage is unknown and worth surfacing.
		r.Outcome = vintage.Unknown
	case cur.Revision == head:
		r.Outcome = vintage.Fresh
	default:
		r.Outcome = vintage.Stale
	}
	return r
}

// DisplayVintage renders the vintage for a one-line surface: a short revision
// for a checkout-tip comparison, the version verbatim for a pinned one, and
// "unknown" when it could not be determined.
func (v VintageStatus) DisplayVintage() string {
	if v.Report.Outcome == vintage.Unknown && v.Report.Current == "" {
		return "unknown"
	}
	if v.Report.Current == "" {
		return "unknown"
	}
	if isHexSHA(v.Report.Current) {
		return shortRev(v.Report.Current)
	}
	return v.Report.Current
}

// Staleness renders the comparator verdict in the words a surface prints.
func (v VintageStatus) Staleness() string {
	switch v.Report.Outcome {
	case vintage.Fresh:
		return "up to date"
	case vintage.Stale:
		ref := v.Report.Expected
		if isHexSHA(ref) {
			ref = shortRev(ref)
		}
		src := v.Source
		if src == "" {
			src = "reference"
		}
		return "stale — behind the " + src + " (" + ref + ")"
	default:
		return "unknown"
	}
}

// readPinnedTag reads the release tag the plugin-cache manifest recorded, or ""
// when the manifest is absent or the tag unresolved. NOTE: the same .binary-meta
// file is parsed by internal/surface/cli/skew.go's readBinaryMeta; that copy is
// the retired skew notice (iss-206) and is left untouched here. The two readers
// should be consolidated when skew.go is removed.
func readPinnedTag(pluginRoot string) string {
	const maxBytes = 4 << 10
	data, err := fsutil.ReadGuarded(filepath.Join(pluginRoot, ".binary-meta"), maxBytes)
	if err != nil {
		return ""
	}
	for _, line := range strings.Split(string(data), "\n") {
		if k, val, ok := strings.Cut(strings.TrimSpace(line), "="); ok && k == "release_tag" {
			return val
		}
	}
	return ""
}

// isHexSHA reports whether s is a full 40-character hex commit SHA — the shape
// the checkout-tip comparison keys on, distinct from a version string.
func isHexSHA(s string) bool {
	if len(s) != 40 {
		return false
	}
	for _, r := range s {
		if (r < '0' || r > '9') && (r < 'a' || r > 'f') {
			return false
		}
	}
	return true
}

// shortRev abbreviates a commit SHA for a one-line surface.
func shortRev(s string) string {
	if len(s) <= 12 {
		return s
	}
	return s[:12]
}
