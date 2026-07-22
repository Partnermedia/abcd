package changelog

import (
	"os"
	"regexp"
	"strings"
	"testing"
)

// TestLatestReleaseTagOrdersBySemverNotLexically pins the anchor's ordering: the
// newest tag is the maximum by SemVer, so v0.10.0 beats v0.9.0. A lexical
// maximum would pick v0.9.0 and derive the next release against a base two
// releases stale.
func TestLatestReleaseTagOrdersBySemverNotLexically(t *testing.T) {
	r := newFixtureRepo(t)
	r.commit("initial")
	for _, tag := range []string{"v0.1.0", "v0.9.0", "v0.10.0", "v0.2.0"} {
		r.git("tag", tag)
	}

	got, ok, err := LatestReleaseTag(r.root)
	if err != nil {
		t.Fatalf("LatestReleaseTag: %v", err)
	}
	if !ok {
		t.Fatal("expected a tag to be found")
	}
	if got.String() != "0.10.0" {
		t.Errorf("LatestReleaseTag = %s, want 0.10.0 (semver order, not lexical)", got)
	}
}

// TestLatestReleaseTagIgnoresNonRelease pins that only strict vX.Y.Z release
// tags anchor a cut: a marketing tag, a two-part tag, an unprefixed one, and a
// prerelease all render or sort as something they are not.
func TestLatestReleaseTagIgnoresNonRelease(t *testing.T) {
	r := newFixtureRepo(t)
	r.commit("initial")
	for _, tag := range []string{"v0.3.0", "v1.2", "release-9", "v9.9.9-rc1", "1.5.0", "v0.3.0.1"} {
		r.git("tag", tag)
	}

	got, ok, err := LatestReleaseTag(r.root)
	if err != nil {
		t.Fatalf("LatestReleaseTag: %v", err)
	}
	if !ok || got.String() != "0.3.0" {
		t.Errorf("LatestReleaseTag = %s (found=%v), want 0.3.0", got, ok)
	}
}

// TestLatestReleaseTagNoTags pins the untagged repo: no tag is not an error, it
// is an absent anchor the caller decides about.
func TestLatestReleaseTagNoTags(t *testing.T) {
	r := newFixtureRepo(t)
	r.commit("initial")

	if _, ok, err := LatestReleaseTag(r.root); err != nil || ok {
		t.Errorf("LatestReleaseTag = (found=%v, err=%v), want (false, nil)", ok, err)
	}
}

// TestLatestChangelogVersion reads the newest dated heading — the one
// auto-release.yml would tag — and ignores everything that is not one.
func TestLatestChangelogVersion(t *testing.T) {
	r := newFixtureRepo(t)
	r.write("CHANGELOG.md", strings.Join([]string{
		"# Changelog",
		"",
		"## [Unreleased]",
		"",
		"### Added",
		"",
		"- something",
		"",
		"## [0.4.0] - 2026-07-21",
		"",
		"## [0.3.0] - 2026-07-18",
		"",
		"## [v0.1.0] - 2026-07-07",
		"",
	}, "\n"))
	r.commit("changelog")

	got, ok, err := LatestChangelogVersion(r.root)
	if err != nil {
		t.Fatalf("LatestChangelogVersion: %v", err)
	}
	if !ok || got.String() != "0.4.0" {
		t.Errorf("LatestChangelogVersion = %s (found=%v), want 0.4.0", got, ok)
	}
}

// TestLatestChangelogVersionAbsent pins that a repo with no CHANGELOG.md yields
// no heading rather than an error.
func TestLatestChangelogVersionAbsent(t *testing.T) {
	r := newFixtureRepo(t)
	r.commit("initial")

	if _, ok, err := LatestChangelogVersion(r.root); err != nil || ok {
		t.Errorf("LatestChangelogVersion = (found=%v, err=%v), want (false, nil)", ok, err)
	}
}

// headingFixtures are the lines the two heading matchers must agree on.
var headingFixtures = []string{
	"## [0.4.0] - 2026-07-21",
	"## [v0.1.0] - 2026-07-07",
	"## [10.20.30] - 2026-01-01",
	"## [Unreleased]",
	"## [0.4.0]",
	"## [0.4.0] — 2026-07-21",
	"### [0.4.0] - 2026-07-21",
	" ## [0.4.0] - 2026-07-21",
	"## [0.4] - 2026-07-21",
	"##[0.4.0] - 2026-07-21",
	"a ## [0.4.0] - 2026-07-21",
}

// TestHeadingRegexMatchesTheWorkflow is the contract test between this package
// and .github/workflows/auto-release.yml: the workflow's grep decides which
// heading gets tagged, so a derivation that reads a DIFFERENT set of headings
// would derive against a release the tagger never sees. The workflow's own
// pattern is extracted from the file (a workflow edit therefore breaks this
// test loudly) and both are run over the same fixtures.
func TestHeadingRegexMatchesTheWorkflow(t *testing.T) {
	pattern := workflowHeadingPattern(t)
	workflowRe, err := regexp.Compile(pattern)
	if err != nil {
		t.Fatalf("workflow pattern %q does not compile: %v", pattern, err)
	}
	// The package pattern differs from the workflow's by its capture group
	// alone, which changes nothing about what matches; strip the parentheses
	// (the pattern contains no others) and the two strings must be identical.
	bare := strings.NewReplacer("(", "", ")", "").Replace(datedHeadingRe.String())
	if pattern != bare {
		t.Errorf("heading pattern drift:\n workflow: %s\n package:  %s", pattern, bare)
	}
	for _, line := range headingFixtures {
		if got, want := datedHeadingRe.MatchString(line), workflowRe.MatchString(line); got != want {
			t.Errorf("disagreement on %q: package=%v workflow=%v", line, got, want)
		}
	}
}

// workflowHeadingPattern extracts the single-quoted -E pattern from the
// auto-release workflow's grep line.
func workflowHeadingPattern(t *testing.T) string {
	t.Helper()
	raw, err := os.ReadFile("../../../.github/workflows/auto-release.yml")
	if err != nil {
		t.Fatalf("read workflow: %v", err)
	}
	for _, line := range strings.Split(string(raw), "\n") {
		if !strings.Contains(line, "grep -m1 -E '") {
			continue
		}
		rest := line[strings.Index(line, "grep -m1 -E '")+len("grep -m1 -E '"):]
		end := strings.Index(rest, "'")
		if end < 0 {
			t.Fatalf("unterminated grep pattern in workflow line: %q", line)
		}
		return rest[:end]
	}
	t.Fatal("no `grep -m1 -E` line found in auto-release.yml; the CHANGELOG contract moved")
	return ""
}
