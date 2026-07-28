package audit_test

import (
	"strings"
	"testing"

	"github.com/REPPL/abcd-cli/internal/core/audit"
)

const positioningBrief = `## Identity (canonical)

- **Title:** abcd — Agent-Based Configuration for Development
- **Tagline:** A host-agnostic configuration layer for intent-driven development.
- **Pitch:** A single Go binary that carries the why from idea to shipped
  reality, usable as a plugin in compatible agent harnesses.
`

const positioningConfig = `{
  "schema_version": 1,
  "block": {"file": ".abcd/development/brief.md", "heading": "Identity (canonical)"},
  "surfaces": [
    {"id": "readme-strapline", "files": ["README.md"], "kind": "regexp",
     "patterns": ["(?s)<p>(.*?)</p>"], "requires": ["tagline"], "template": "{tagline}"},
    {"id": "manifest-description", "files": [".claude-plugin/plugin.json"],
     "kind": "json_field", "field": "description", "requires": ["tagline"], "template": "{tagline}"}
  ]
}`

const cleanStrapline = "<div>\n\n  <h1>abcd</h1>\n\n  <p>A host-agnostic configuration layer for intent-driven development.</p>\n\n</div>\n"

// The iss-143 README variant, verbatim in shape.
const driftedStrapline = "<div>\n\n  <h1>abcd</h1>\n\n  <p>An opinionated, intent-driven development framework for product thinkers.</p>\n\n</div>\n"

const cleanManifestJSON = `{"name":"abcd","description":"A host-agnostic configuration layer for intent-driven development."}`

// withPositioning lays the identity block, the registry, and conforming
// surfaces into a fixture repo.
func (b *repoBuilder) withPositioning() *repoBuilder {
	return b.
		file(".abcd/development/brief.md", positioningBrief).
		file(".abcd/positioning.json", positioningConfig).
		file("README.md", cleanStrapline).
		file(".claude-plugin/plugin.json", cleanManifestJSON)
}

// A repo that has not adopted the check is skipped, not failed: the absence of
// .abcd/positioning.json is opting out, and must never fabricate findings.
func TestPositioningSkippedWithoutAConfig(t *testing.T) {
	res := newFixtureRepo(t).conforming().commit().run()
	if f := findingFor(res, "identity-positioning"); f != nil {
		t.Fatalf("un-adopted repo got a positioning finding: %+v", f)
	}
	var skipped bool
	for _, id := range res.Skipped {
		if id == "identity-positioning" {
			skipped = true
		}
	}
	if !skipped {
		t.Errorf("identity-positioning is neither skipped nor evaluated; Skipped = %v", res.Skipped)
	}
}

func TestPositioningCleanRepoHasNoFinding(t *testing.T) {
	res := newFixtureRepo(t).conforming().withPositioning().commit().run()
	if f := findingFor(res, "identity-positioning"); f != nil {
		t.Fatalf("conforming surfaces produced a finding: %+v", f)
	}
}

// AC2: a diverging surface yields a warn-tier finding naming the file, the
// exact drifted line, and the canonical line it should carry.
func TestPositioningDriftIsAWarnFindingNamingTheExactLine(t *testing.T) {
	res := newFixtureRepo(t).conforming().withPositioning().
		file("README.md", driftedStrapline).commit().run()

	f := findingFor(res, "identity-positioning")
	if f == nil {
		t.Fatal("drifted strapline produced no finding")
	}
	if f.Severity != audit.SeverityWarn {
		t.Errorf("severity = %s, want warn (the family highlights, it does not gate)", f.Severity)
	}
	if f.File != "README.md" {
		t.Errorf("File = %q, want README.md", f.File)
	}
	if f.Line != 5 {
		t.Errorf("Line = %d, want 5 (the strapline's line)", f.Line)
	}
	if !strings.Contains(f.Message, "An opinionated, intent-driven development framework for product thinkers.") {
		t.Errorf("message does not quote the exact drifted line: %s", f.Message)
	}
	if !strings.Contains(f.Message, "A host-agnostic configuration layer for intent-driven development.") {
		t.Errorf("message does not quote the canonical line: %s", f.Message)
	}
	if res.ExitCode != 1 {
		t.Errorf("exit = %d, want 1 (warnings only)", res.ExitCode)
	}
}

// AC2, second half: per-repo configuration upgrades the whole family to blocker.
func TestPositioningSeverityUpgradesToBlocker(t *testing.T) {
	res := newFixtureRepo(t).conforming().withPositioning().
		file("README.md", driftedStrapline).
		file(".abcd/positioning.json", strings.Replace(positioningConfig,
			`"schema_version": 1,`, `"schema_version": 1, "severity": "blocker",`, 1)).
		commit().run()

	f := findingFor(res, "identity-positioning")
	if f == nil {
		t.Fatal("drifted strapline produced no finding")
	}
	if f.Severity != audit.SeverityError {
		t.Fatalf("severity = %s, want error after the per-repo upgrade", f.Severity)
	}
	if res.ExitCode != 2 {
		t.Errorf("exit = %d, want 2", res.ExitCode)
	}
}

// A registered surface whose locator stopped matching must surface, not pass.
func TestPositioningUnlocatableSurfaceIsAFinding(t *testing.T) {
	res := newFixtureRepo(t).conforming().withPositioning().
		file("README.md", "# abcd\n\nno strapline element\n").commit().run()

	f := findingFor(res, "identity-positioning")
	if f == nil {
		t.Fatal("unlocatable surface produced no finding")
	}
	if !strings.Contains(f.Message, "could not be located") {
		t.Errorf("message does not say the surface could not be located: %s", f.Message)
	}
}

// A malformed registry must be reported, never read as "no drift".
func TestPositioningMalformedConfigIsAFinding(t *testing.T) {
	res := newFixtureRepo(t).conforming().withPositioning().
		file(".abcd/positioning.json", "{ not json").commit().run()

	f := findingFor(res, "identity-positioning")
	if f == nil {
		t.Fatal("malformed registry produced no finding")
	}
	if f.File != ".abcd/positioning.json" {
		t.Errorf("File = %q, want the registry", f.File)
	}
}

// A registry pointing at a block that is not there is the same class of problem.
func TestPositioningMissingBlockIsAFinding(t *testing.T) {
	res := newFixtureRepo(t).conforming().withPositioning().
		file(".abcd/development/brief.md", "# no identity block here\n").commit().run()

	f := findingFor(res, "identity-positioning")
	if f == nil {
		t.Fatal("missing identity block produced no finding")
	}
	if !strings.Contains(f.Message, "heading") && !strings.Contains(f.Message, "block") {
		t.Errorf("message does not explain the block could not be read: %s", f.Message)
	}
}

// Findings must never carry an absolute developer path, on any error path.
func TestPositioningFindingsCarryNoAbsolutePath(t *testing.T) {
	cases := map[string]func(*repoBuilder) *repoBuilder{
		"malformed registry": func(b *repoBuilder) *repoBuilder {
			return b.file(".abcd/positioning.json", "{ not json")
		},
		"block file absent": func(b *repoBuilder) *repoBuilder {
			return b.file(".abcd/positioning.json", strings.Replace(positioningConfig,
				".abcd/development/brief.md", ".abcd/development/nowhere.md", 1))
		},
		"block heading absent": func(b *repoBuilder) *repoBuilder {
			return b.file(".abcd/development/brief.md", "# nothing here\n")
		},
	}
	for name, mutate := range cases {
		t.Run(name, func(t *testing.T) {
			b := mutate(newFixtureRepo(t).conforming().withPositioning()).commit()
			res := b.run()
			if findingFor(res, "identity-positioning") == nil {
				t.Fatal("no positioning finding to inspect")
			}
			for _, f := range res.Findings {
				if strings.Contains(f.Message, b.root) || strings.Contains(f.Fix, b.root) {
					t.Errorf("finding leaks the absolute repo root: %s / %s", f.Message, f.Fix)
				}
			}
		})
	}
}
