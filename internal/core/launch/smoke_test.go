package launch

import (
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestSmokeLightPassesOnCommittedPayload is the positive detector: this
// repository's real payload is installable — both manifests parse, the
// marketplace source resolves, and every declared path is carried.
func TestSmokeLightPassesOnCommittedPayload(t *testing.T) {
	report := SmokeLight(bundleTreeFor(t, repoRootForTest(t)))
	if !report.OK {
		t.Fatalf("the committed payload must pass the light smoke, got %+v", report.Findings)
	}
	if report.Checked == 0 {
		t.Error("a passing smoke that checked nothing is a vacuous pass")
	}
}

// TestSmokeLightFailsAndNamesTheMissingPath is the negative detector, one case
// per way a payload can be uninstallable. Every failure must NAME the path or
// manifest at fault — a smoke that only says "failed" cannot be acted on.
func TestSmokeLightFailsAndNamesTheMissingPath(t *testing.T) {
	cases := []struct {
		name        string
		files       map[string]string
		marketplace string
		wantKind    string
		wantNamed   string
	}{
		{
			name:      "a declared command missing from the payload",
			files:     map[string]string{"plugin_extra": `"commands": "./commands/ghost.md"`},
			wantKind:  findingMissingPath,
			wantNamed: "commands/ghost.md",
		},
		{
			name:      "a declared agent missing from the payload",
			files:     map[string]string{"plugin_extra": `"agents": ["./agents/absent.md"]`},
			wantKind:  findingMissingPath,
			wantNamed: "agents/absent.md",
		},
		{
			name:      "a declared skill missing from the payload",
			files:     map[string]string{"plugin_extra": `"skills": "./skills/nowhere"`},
			wantKind:  findingMissingPath,
			wantNamed: "skills/nowhere",
		},
		{
			name: "a hook command pointing at a file the payload excludes",
			files: map[string]string{
				"hooks/hooks.json": `{"hooks": {"SessionStart": [{"hooks": [{"type": "command", "command": "\"$CLAUDE_PLUGIN_ROOT/scripts/absent.sh\""}]}]}}`,
			},
			wantKind:  findingMissingPath,
			wantNamed: "scripts/absent.sh",
		},
		{
			name:        "a marketplace source that resolves to no plugin manifest",
			files:       map[string]string{},
			marketplace: `{"name": "m", "plugins": [{"name": "abcd", "source": "./nested"}]}`,
			wantKind:    findingSourceUnresolved,
			wantNamed:   "nested",
		},
		{
			name:        "a marketplace entry naming a plugin the manifest does not",
			files:       map[string]string{},
			marketplace: `{"name": "m", "plugins": [{"name": "other", "source": "./"}]}`,
			wantKind:    findingNameMismatch,
			wantNamed:   "other",
		},
		{
			name:      "an unparseable plugin manifest",
			files:     map[string]string{"plugin_broken": "yes"},
			wantKind:  findingManifestUnreadable,
			wantNamed: ".claude-plugin/plugin.json",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			root := t.TempDir()
			writeSurfaceFixture(t, root, tc.files)
			if tc.marketplace != "" {
				writeFile(t, root, ".claude-plugin/marketplace.json", tc.marketplace)
			}
			report := SmokeLight(bundleTreeFor(t, root))
			if report.OK {
				t.Fatalf("the smoke must FAIL: %+v", report)
			}
			var named bool
			for _, f := range report.Findings {
				if f.Kind == tc.wantKind && strings.Contains(f.Path+" "+f.Detail, tc.wantNamed) {
					named = true
				}
			}
			if !named {
				t.Errorf("expected a %q finding naming %q, got %+v", tc.wantKind, tc.wantNamed, report.Findings)
			}
		})
	}
}

// TestRenderPayloadRefusesAnUninstallablePayload is the enforcement detector: the
// render is the only step that materialises a release artefact, so it is where a
// missing declared path must stop the cut rather than publish.
func TestRenderPayloadRefusesAnUninstallablePayload(t *testing.T) {
	root := t.TempDir()
	writeSurfaceFixture(t, root, map[string]string{"plugin_extra": `"commands": "./commands/ghost.md"`})
	writeFile(t, root, ".abcd/config/version-location.json",
		`{"manifest_path": ".claude-plugin/plugin.json", "json_pointer": "/version"}`)

	dest := filepath.Join(t.TempDir(), "payload")
	_, err := RenderPayload(PayloadRenderRequest{
		RepoRoot: root, Dest: dest, Version: "1.0.0",
		Entry: ChangelogEntry{Tier: "patch", Reason: "r", Date: time.Now(), SourceSHA: "abc"},
	})
	if err == nil {
		t.Fatal("the render must refuse a payload whose declared surface is incomplete")
	}
	if !strings.Contains(err.Error(), "commands/ghost.md") {
		t.Errorf("the refusal must name the missing path, got %v", err)
	}
}

// TestDryRunReportsTheInstallabilitySmoke proves the gate is reachable from the
// preview a maintainer actually runs, and that a finding reaches WouldRefuseOn
// rather than sitting silently in the report.
func TestDryRunReportsTheInstallabilitySmoke(t *testing.T) {
	root := t.TempDir()
	writeSurfaceFixture(t, root, map[string]string{"plugin_extra": `"commands": "./commands/ghost.md"`})

	report, err := DryRun(DryRunRequest{RepoRoot: root, Version: "1.0.0"})
	if err != nil {
		t.Fatalf("dry-run must still produce a report: %v", err)
	}
	if report.Smoke.OK {
		t.Fatalf("the smoke must fail on a missing declared path: %+v", report.Smoke)
	}
	var gated bool
	for _, g := range report.Gates {
		if g.Name == "installability-smoke" && g.Status == "ran" {
			gated = true
		}
	}
	if !gated {
		t.Errorf("the installability smoke must appear as a gate that ran, got %+v", report.Gates)
	}
	var refused bool
	for _, r := range report.WouldRefuseOn {
		if strings.Contains(r, "commands/ghost.md") {
			refused = true
		}
	}
	if !refused {
		t.Errorf("WouldRefuseOn must name the missing path, got %v", report.WouldRefuseOn)
	}
}
