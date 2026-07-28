package positioning

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadConfigAbsentIsNotAnError(t *testing.T) {
	cfg, ok, err := LoadConfig(t.TempDir())
	if err != nil || ok {
		t.Fatalf("LoadConfig(bare repo) = %+v, %v, %v; want zero, false, nil", cfg, ok, err)
	}
}

func TestLoadConfigRoundTrip(t *testing.T) {
	root := writeRepo(t, map[string]string{
		ConfigRelPath: `{
  "schema_version": 1,
  "block": {"file": "brief.md", "heading": "Identity (canonical)"},
  "severity": "blocker",
  "surfaces": [
    {"id": "readme-strapline", "files": ["README.md"], "kind": "regexp",
     "patterns": ["(?s)<p>(.*?)</p>"], "requires": ["tagline"], "template": "{tagline}"}
  ]
}`,
	})
	cfg, ok, err := LoadConfig(root)
	if err != nil || !ok {
		t.Fatalf("LoadConfig = %v, %v", ok, err)
	}
	if cfg.Block.File != "brief.md" || cfg.severity() != SeverityBlocker || len(cfg.Surfaces) != 1 {
		t.Fatalf("round-trip lost data: %+v", cfg)
	}
}

func TestLoadConfigRejectsMalformedRegistries(t *testing.T) {
	cases := map[string]string{
		"not json":            `{`,
		"unknown field":       `{"schema_version":1,"block":{"file":"b.md","heading":"H"},"surfaces":[],"typo":1}`,
		"wrong schema":        `{"schema_version":2,"block":{"file":"b.md","heading":"H"}}`,
		"escaping block path": `{"schema_version":1,"block":{"file":"../outside.md","heading":"H"}}`,
		"absolute block path": `{"schema_version":1,"block":{"file":"/etc/passwd","heading":"H"}}`,
		"empty heading":       `{"schema_version":1,"block":{"file":"b.md","heading":"  "}}`,
		"bad severity":        `{"schema_version":1,"block":{"file":"b.md","heading":"H"},"severity":"loud"}`,
		"surface no id": `{"schema_version":1,"block":{"file":"b.md","heading":"H"},
			"surfaces":[{"files":["README.md"],"kind":"regexp","patterns":["(x)"],"requires":["tagline"]}]}`,
		"surface no files": `{"schema_version":1,"block":{"file":"b.md","heading":"H"},
			"surfaces":[{"id":"s","files":[],"kind":"regexp","patterns":["(x)"],"requires":["tagline"]}]}`,
		"surface escaping file": `{"schema_version":1,"block":{"file":"b.md","heading":"H"},
			"surfaces":[{"id":"s","files":["../x"],"kind":"regexp","patterns":["(x)"],"requires":["tagline"]}]}`,
		"unknown kind": `{"schema_version":1,"block":{"file":"b.md","heading":"H"},
			"surfaces":[{"id":"s","files":["README.md"],"kind":"toml","requires":["tagline"]}]}`,
		"pattern does not compile": `{"schema_version":1,"block":{"file":"b.md","heading":"H"},
			"surfaces":[{"id":"s","files":["README.md"],"kind":"regexp","patterns":["("],"requires":["tagline"]}]}`,
		"pattern without a capture group": `{"schema_version":1,"block":{"file":"b.md","heading":"H"},
			"surfaces":[{"id":"s","files":["README.md"],"kind":"regexp","patterns":["strapline"],"requires":["tagline"]}]}`,
		"pattern with two capture groups": `{"schema_version":1,"block":{"file":"b.md","heading":"H"},
			"surfaces":[{"id":"s","files":["README.md"],"kind":"regexp","patterns":["(a)(b)"],"requires":["tagline"]}]}`,
		"json_field without a field": `{"schema_version":1,"block":{"file":"b.md","heading":"H"},
			"surfaces":[{"id":"s","files":["p.json"],"kind":"json_field","requires":["tagline"]}]}`,
		"no requires": `{"schema_version":1,"block":{"file":"b.md","heading":"H"},
			"surfaces":[{"id":"s","files":["README.md"],"kind":"regexp","patterns":["(x)"],"requires":[]}]}`,
		"unknown required field": `{"schema_version":1,"block":{"file":"b.md","heading":"H"},
			"surfaces":[{"id":"s","files":["README.md"],"kind":"regexp","patterns":["(x)"],"requires":["slogan"]}]}`,
		"unknown template placeholder": `{"schema_version":1,"block":{"file":"b.md","heading":"H"},
			"surfaces":[{"id":"s","files":["README.md"],"kind":"regexp","patterns":["(x)"],"requires":["tagline"],"template":"{slogan}"}]}`,
		"duplicate surface id": `{"schema_version":1,"block":{"file":"b.md","heading":"H"},
			"surfaces":[
			 {"id":"s","files":["README.md"],"kind":"regexp","patterns":["(x)"],"requires":["tagline"]},
			 {"id":"s","files":["AGENTS.md"],"kind":"regexp","patterns":["(x)"],"requires":["tagline"]}]}`,
	}
	for name, body := range cases {
		t.Run(name, func(t *testing.T) {
			root := writeRepo(t, map[string]string{ConfigRelPath: body})
			_, ok, err := LoadConfig(root)
			if err == nil {
				t.Fatalf("LoadConfig accepted a malformed registry (ok=%v)", ok)
			}
			if !errors.Is(err, ErrConfigInvalid) {
				t.Fatalf("err = %v, want ErrConfigInvalid", err)
			}
			// A refusal must never leak the absolute repo path.
			if strings.Contains(err.Error(), filepath.ToSlash(root)) {
				t.Errorf("error leaks the absolute repo root: %v", err)
			}
		})
	}
}

func TestEmptySurfaceListFallsBackToTheThreeDefaults(t *testing.T) {
	cfg := Config{SchemaVersion: 1, Block: BlockLocation{File: "b.md", Heading: "H"}}
	got := cfg.surfaces()
	if len(got) != 3 {
		t.Fatalf("default registry has %d surfaces, want 3", len(got))
	}
	want := map[string]bool{"readme-strapline": true, "manifest-description": true, "conventions-opening": true}
	for _, s := range got {
		if !want[s.ID] {
			t.Errorf("unexpected default surface %q", s.ID)
		}
	}
	if err := (Config{SchemaVersion: 1, Block: BlockLocation{File: "b.md", Heading: "H"}, Surfaces: got}).Validate(); err != nil {
		t.Fatalf("the default surfaces do not validate: %v", err)
	}
}

// The defaults must locate abcd-cli's own surface shapes out of the box, so a
// repo that adopts the check without a bespoke registry is still held to it.
func TestDefaultSurfacesLocateTheCanonicalShapes(t *testing.T) {
	root := writeRepo(t, map[string]string{
		"brief.md": canonBrief, "README.md": cleanREADME,
		".claude-plugin/plugin.json": cleanManifest, "AGENTS.md": cleanAgents,
	})
	cfg := Config{SchemaVersion: 1, Block: BlockLocation{File: "brief.md", Heading: "Identity (canonical)"}}
	rep, err := Check(root, cfg)
	if err != nil {
		t.Fatalf("Check: %v", err)
	}
	for _, s := range rep.Surfaces {
		if s.Status != StatusOK {
			t.Errorf("default surface %s: status %s (file %q, found %q)", s.ID, s.Status, s.File, s.Found)
		}
	}
}

// A conventions file whose opening paragraph runs to end-of-file must still
// locate; a locator that needs a trailing blank line silently reports
// "unlocatable" on a perfectly conforming repo.
func TestConventionsOpeningLocatesAtEndOfFile(t *testing.T) {
	root := writeRepo(t, map[string]string{
		"brief.md": canonBrief, "README.md": cleanREADME,
		".claude-plugin/plugin.json": cleanManifest,
		"AGENTS.md": `# AGENTS.md

<!-- BEGIN ABCD -->
managed
<!-- END ABCD -->

A host-agnostic configuration layer for intent-driven development.`,
	})
	rep, err := Check(root, Config{SchemaVersion: 1, Block: BlockLocation{File: "brief.md", Heading: "Identity (canonical)"}})
	if err != nil {
		t.Fatalf("Check: %v", err)
	}
	for _, s := range rep.Surfaces {
		if s.ID == "conventions-opening" && s.Status != StatusOK {
			t.Fatalf("conventions-opening at EOF: status %s, found %q", s.Status, s.Found)
		}
	}
}

func TestSurfaceRenderTemplate(t *testing.T) {
	b := Block{Title: "T", Tagline: "TL.", Pitch: "P."}
	if got := (Surface{Requires: []string{"tagline"}, Template: "{tagline}"}).render(b); got != "TL." {
		t.Errorf("render = %q", got)
	}
	if got := (Surface{Requires: []string{"tagline", "pitch"}}).render(b); got != "TL. P." {
		t.Errorf("default render = %q, want the required fields joined", got)
	}
	if got := (Surface{Requires: []string{"title"}, Template: "{title} — {tagline}"}).render(b); got != "T — TL." {
		t.Errorf("render = %q", got)
	}
}
