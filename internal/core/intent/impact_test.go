package intent

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/intentdriven/abcd/internal/core/lint"
)

// lintIntentImpact runs only the intent_impact_valid record-lint blocker over the
// intent tree under root and returns its findings.
func lintIntentImpact(t *testing.T, root string) []lint.Finding {
	t.Helper()
	cfg := lint.Config{
		Roots: []string{".abcd/development"},
		Rules: map[string]lint.RuleConfig{
			"intent_impact_valid": {Enabled: true, Severity: "blocker", IntentsDir: "intents"},
		},
	}
	fs, err := lint.Lint(cfg, root)
	if err != nil {
		t.Fatalf("lint: %v", err)
	}
	return fs
}

// TestCreateFromTextStampsImpact is the iss-117 intent half: the seed-draft path
// must be able to stamp a valid impact so the record survives — unchanged by the
// impact-preserving planned->shipped move — into shipped/, where
// intent_impact_valid REQUIRES one. A seed path that cannot set impact produces an
// intent the tool can never ship past its own blocker.
func TestCreateFromTextStampsImpact(t *testing.T) {
	root := t.TempDir()
	it, _, err := CreateFromText(root, "a user-facing improvement worth shipping", "additive")
	if err != nil {
		t.Fatalf("CreateFromText: %v", err)
	}
	// The seeded draft carries the impact bare (impact: additive), matching the
	// machine-read enum the shipped-intent gate compares byte-for-byte.
	raw, err := os.ReadFile(filepath.Join(root, it.Path))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(raw), "\nimpact: additive\n") {
		t.Fatalf("seeded draft must carry bare impact, got:\n%s", raw)
	}

	// Simulate the planned->shipped move (moveIntentToBucket preserves frontmatter):
	// relocate the file into shipped/ and assert intent_impact_valid is satisfied.
	shipped := filepath.Join(root, shippedDir, filepath.Base(it.Path))
	if err := os.MkdirAll(filepath.Dir(shipped), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Rename(filepath.Join(root, it.Path), shipped); err != nil {
		t.Fatal(err)
	}
	if fs := lintIntentImpact(t, root); len(fs) != 0 {
		t.Fatalf("shipped record must satisfy intent_impact_valid, got %d finding(s): %+v", len(fs), fs)
	}
}

// TestCreateFromTextRejectsBadImpact proves the seed boundary fails closed on an
// impact the intent gate would reject: a misspelling, and — because an intent is
// press-release-first — the otherwise-legal `internal`.
func TestCreateFromTextRejectsBadImpact(t *testing.T) {
	root := t.TempDir()
	for _, bad := range []string{"additiv", "Additive", "internal", `"fix"`} {
		if _, _, err := CreateFromText(root, "some intent text", bad); err == nil {
			t.Fatalf("CreateFromText with impact %q must be refused", bad)
		}
	}
	// No drafts file appeared for any refused create.
	if entries, _ := os.ReadDir(filepath.Join(root, draftsDir)); len(entries) != 0 {
		t.Fatalf("a refused create wrote %d files, want 0", len(entries))
	}
}

// TestCreateFromTextImpactOptional keeps the no-impact path intact: an empty
// impact seeds a draft with no impact line (drafts are not gated), exactly as
// before, so the existing capture->intent promotion flow is unchanged.
func TestCreateFromTextImpactOptional(t *testing.T) {
	root := t.TempDir()
	it, _, err := CreateFromText(root, "seeded without a judgement yet", "")
	if err != nil {
		t.Fatalf("CreateFromText: %v", err)
	}
	raw, err := os.ReadFile(filepath.Join(root, it.Path))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(raw), "impact:") {
		t.Fatalf("an unset impact must write no impact line, got:\n%s", raw)
	}
}
