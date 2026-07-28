package positioning

import (
	"strings"
	"testing"
)

// canonBrief is the identity block every fixture below is held to — the same
// three bullets abcd-cli's own brief carries.
const canonBrief = `## Identity (canonical)

- **Title:** abcd — Agent-Based Configuration for Development
- **Tagline:** A host-agnostic configuration layer for intent-driven development.
- **Pitch:** A single Go binary that carries the why from idea to shipped
  reality, usable as a plugin in compatible agent harnesses.
`

// The three CLEAN surfaces: byte-for-byte the shapes abcd-cli ships today —
// an HTML strapline in a centred README header, a manifest description that
// carries tagline and pitch, and a conventions opening paragraph that wraps the
// tagline across two lines with markdown emphasis inside it.
const (
	cleanREADME = `<div align="center">

  <h1>Agent-Based Configuration for Development</h1>

  <p>A host-agnostic configuration layer for intent-driven development.</p>

  <img src="badge.svg" alt="Status">

</div>

## Install
`
	cleanManifest = `{
  "name": "abcd",
  "description": "A host-agnostic configuration layer for intent-driven development — a single Go binary that carries the why from idea to shipped reality, usable as a plugin in compatible agent harnesses.",
  "license": "MIT"
}
`
	cleanAgents = `# AGENTS.md

<!-- BEGIN ABCD -->
managed block
<!-- END ABCD -->

abcd (Agent-Based Configuration for Development) is a Go CLI and an agent-harness
plugin: a host-agnostic **configuration layer for intent-driven development**. A single
` + "`abcd`" + ` binary holds all behaviour in a transport-agnostic core.

## Build
`
)

// The iss-143 acceptance corpus: the three real drifted variants the repo was
// carrying when the drift was recorded. Every one of them must be caught.
const (
	drift143README = `<div align="center">

  <h1>Agent-Based Configuration for Development</h1>

  <p>An opinionated, intent-driven development framework for product thinkers.</p>

</div>
`
	drift143Manifest = `{
  "name": "abcd",
  "description": "A host-agnostic configuration layer for development — a single Go binary that carries the why from idea to shipped reality, usable as a plugin in compatible agent harnesses.",
  "license": "MIT"
}
`
	drift143Agents = `# AGENTS.md

<!-- BEGIN ABCD -->
managed block
<!-- END ABCD -->

abcd (Agent-Based Configuration for Development) is a Go CLI and an agent-harness
plugin: a host-agnostic **configuration layer for development**. A single
` + "`abcd`" + ` binary holds all behaviour in a transport-agnostic core.
`
)

// abcdLikeConfig is the surface registry abcd-cli's own repo declares.
func abcdLikeConfig() Config {
	return Config{
		SchemaVersion: 1,
		Block:         BlockLocation{File: "brief.md", Heading: "Identity (canonical)"},
		Surfaces: []Surface{
			{
				ID: "readme-strapline", Files: []string{"README.md"}, Kind: KindRegexp,
				Patterns: []string{`(?s)<p>(.*?)</p>`},
				Requires: []string{"tagline"}, Template: "{tagline}",
			},
			{
				ID: "manifest-description", Files: []string{"plugin.json"}, Kind: KindJSONField,
				Field: "description", Requires: []string{"tagline", "pitch"},
				Template: "{tagline} {pitch}",
			},
			{
				ID: "conventions-opening", Files: []string{"AGENTS.md"}, Kind: KindRegexp,
				Patterns: []string{`(?s)<!-- END ABCD -->\n+(.*?)(?:\n\n|\z)`},
				Requires: []string{"tagline"}, Template: "{tagline}",
			},
		},
	}
}

func cleanRepo(t *testing.T) string {
	t.Helper()
	return writeRepo(t, map[string]string{
		"brief.md":    canonBrief,
		"README.md":   cleanREADME,
		"plugin.json": cleanManifest,
		"AGENTS.md":   cleanAgents,
	})
}

func TestCheckCleanRepoHasNoDrift(t *testing.T) {
	rep, err := Check(cleanRepo(t), abcdLikeConfig())
	if err != nil {
		t.Fatalf("Check: %v", err)
	}
	if len(rep.Surfaces) != 3 {
		t.Fatalf("checked %d surfaces, want 3", len(rep.Surfaces))
	}
	for _, s := range rep.Surfaces {
		if s.Status != StatusOK {
			t.Errorf("surface %s: status %s (found %q, missing %v), want ok", s.ID, s.Status, s.Found, s.Missing)
		}
	}
	if rep.Drifted() != 0 {
		t.Errorf("Drifted() = %d, want 0", rep.Drifted())
	}
}

// The acceptance corpus, one variant at a time: each drifted surface must be
// caught, and the report must name the file, the exact drifted text, and the
// canonical line it should carry.
func TestCheckCatchesTheIss143DriftCorpus(t *testing.T) {
	cases := []struct {
		surfaceID string
		file      string
		content   string
		wantFound string
	}{
		{"readme-strapline", "README.md", drift143README, "An opinionated, intent-driven development framework for product thinkers."},
		{"manifest-description", "plugin.json", drift143Manifest, "A host-agnostic configuration layer for development — a single Go binary"},
		{"conventions-opening", "AGENTS.md", drift143Agents, "configuration layer for development"},
	}
	for _, tc := range cases {
		t.Run(tc.surfaceID, func(t *testing.T) {
			files := map[string]string{
				"brief.md": canonBrief, "README.md": cleanREADME,
				"plugin.json": cleanManifest, "AGENTS.md": cleanAgents,
			}
			files[tc.file] = tc.content
			rep, err := Check(writeRepo(t, files), abcdLikeConfig())
			if err != nil {
				t.Fatalf("Check: %v", err)
			}
			var got *SurfaceResult
			for i := range rep.Surfaces {
				if rep.Surfaces[i].ID == tc.surfaceID {
					got = &rep.Surfaces[i]
				}
			}
			if got == nil {
				t.Fatalf("surface %s missing from the report", tc.surfaceID)
			}
			if got.Status != StatusDrifted {
				t.Fatalf("status = %s, want drifted", got.Status)
			}
			if got.File != tc.file {
				t.Errorf("File = %q, want %q", got.File, tc.file)
			}
			if got.Line <= 0 {
				t.Errorf("Line = %d, want the 1-based line of the drifted text", got.Line)
			}
			if !strings.Contains(got.Found, tc.wantFound) {
				t.Errorf("Found = %q, want it to contain %q", got.Found, tc.wantFound)
			}
			// The canonical line the surface should carry is quoted back.
			if !strings.Contains(got.Canonical, "A host-agnostic configuration layer for intent-driven development.") {
				t.Errorf("Canonical = %q, want the canonical tagline", got.Canonical)
			}
			if len(got.Missing) == 0 {
				t.Errorf("Missing is empty; want the drifted field named")
			}
			// The other two surfaces stay clean — one drifted surface must not
			// smear findings across the registry.
			for _, s := range rep.Surfaces {
				if s.ID != tc.surfaceID && s.Status != StatusOK {
					t.Errorf("collateral finding on %s: %s", s.ID, s.Status)
				}
			}
		})
	}
}

// All three drifted at once — the shape iss-143 actually recorded.
func TestCheckCatchesAllThreeVariantsTogether(t *testing.T) {
	root := writeRepo(t, map[string]string{
		"brief.md": canonBrief, "README.md": drift143README,
		"plugin.json": drift143Manifest, "AGENTS.md": drift143Agents,
	})
	rep, err := Check(root, abcdLikeConfig())
	if err != nil {
		t.Fatalf("Check: %v", err)
	}
	if rep.Drifted() != 3 {
		t.Fatalf("Drifted() = %d, want 3 (all of the iss-143 corpus)", rep.Drifted())
	}
}

func TestCheckNormalisesEmphasisWhitespaceAndCase(t *testing.T) {
	root := writeRepo(t, map[string]string{
		"brief.md": canonBrief,
		"README.md": `<div>

  <p>a *host-agnostic*   configuration layer
  for INTENT-DRIVEN development</p>

</div>
`,
		"plugin.json": cleanManifest, "AGENTS.md": cleanAgents,
	})
	rep, err := Check(root, abcdLikeConfig())
	if err != nil {
		t.Fatalf("Check: %v", err)
	}
	for _, s := range rep.Surfaces {
		if s.Status != StatusOK {
			t.Errorf("surface %s: %s — normalisation should accept emphasis, wrapping, and case", s.ID, s.Status)
		}
	}
}

func TestCheckAbsentSurfaceFileIsSkippedNotDrifted(t *testing.T) {
	root := writeRepo(t, map[string]string{
		"brief.md": canonBrief, "README.md": cleanREADME, "AGENTS.md": cleanAgents,
	})
	rep, err := Check(root, abcdLikeConfig())
	if err != nil {
		t.Fatalf("Check: %v", err)
	}
	for _, s := range rep.Surfaces {
		if s.ID == "manifest-description" {
			if s.Status != StatusAbsent {
				t.Fatalf("absent manifest: status %s, want absent", s.Status)
			}
			return
		}
	}
	t.Fatal("manifest-description missing from the report")
}

// A registered surface whose locator no longer matches must be reported, not
// silently treated as conforming — a silent pass is how drift settles in.
func TestCheckUnlocatableSurfaceIsReported(t *testing.T) {
	root := writeRepo(t, map[string]string{
		"brief.md": canonBrief, "README.md": "# Title\n\nno strapline element here\n",
		"plugin.json": cleanManifest, "AGENTS.md": cleanAgents,
	})
	rep, err := Check(root, abcdLikeConfig())
	if err != nil {
		t.Fatalf("Check: %v", err)
	}
	for _, s := range rep.Surfaces {
		if s.ID == "readme-strapline" {
			if s.Status != StatusUnlocatable {
				t.Fatalf("status = %s, want unlocatable", s.Status)
			}
			return
		}
	}
	t.Fatal("readme-strapline missing from the report")
}

// The first candidate file that exists wins, so one registry entry can cover
// repos with different manifest formats.
func TestSurfaceUsesFirstExistingCandidateFile(t *testing.T) {
	cfg := abcdLikeConfig()
	cfg.Surfaces[1].Files = []string{".claude-plugin/plugin.json", "package.json"}
	root := writeRepo(t, map[string]string{
		"brief.md": canonBrief, "README.md": cleanREADME,
		"package.json": cleanManifest, "AGENTS.md": cleanAgents,
	})
	rep, err := Check(root, cfg)
	if err != nil {
		t.Fatalf("Check: %v", err)
	}
	for _, s := range rep.Surfaces {
		if s.ID == "manifest-description" {
			if s.File != "package.json" || s.Status != StatusOK {
				t.Fatalf("File = %q status = %s, want package.json/ok", s.File, s.Status)
			}
			return
		}
	}
	t.Fatal("manifest-description missing from the report")
}

func TestCheckPropagatesABlockParseFailure(t *testing.T) {
	root := writeRepo(t, map[string]string{"README.md": cleanREADME})
	if _, err := Check(root, abcdLikeConfig()); err == nil {
		t.Fatal("Check with an unreadable identity block returned nil error")
	}
}

func TestNormalize(t *testing.T) {
	cases := map[string]string{
		"A host-agnostic  layer.":         "a host-agnostic layer.",
		"**bold** and _under_ and `code`": "bold and under and code",
		"<p>wrapped\n  across lines</p>":  "wrapped across lines",
		"em—dash and en–dash and non brk": "em-dash and en-dash and non brk",
	}
	for in, want := range cases {
		if got := normalize(in); got != want {
			t.Errorf("normalize(%q) = %q, want %q", in, got, want)
		}
	}
}
