package intent

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/REPPL/abcd-cli/internal/core/frontmatter"
	"github.com/REPPL/abcd-cli/internal/core/lint"
	"github.com/REPPL/abcd-cli/internal/core/spec"
)

const (
	draftsDir   = ".abcd/development/intents/drafts"
	plannedDir  = ".abcd/development/intents/planned"
	shippedDir  = ".abcd/development/intents/shipped"
	specsOpen   = ".abcd/development/specs/open"
	specsClosed = ".abcd/development/specs/closed"
)

// writeFile writes content to root/rel, creating parent directories.
func writeFile(t *testing.T, root, rel, content string) {
	t.Helper()
	abs := filepath.Join(root, rel)
	if err := os.MkdirAll(filepath.Dir(abs), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(abs, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

// draftWithAC is a draft intent carrying a non-empty ## Acceptance Criteria
// section (the itd-1 gate Plan enforces).
func draftWithAC(id, slug string) string {
	return "---\n" +
		"id: " + id + "\n" +
		"slug: " + slug + "\n" +
		"spec_id: null\n" +
		"kind: null\n" +
		"---\n" +
		"# " + slug + "\n\n" +
		"## Acceptance Criteria\n\n" +
		"- **Given** a user, **when** they act, **then** it works.\n"
}

func TestLoadCorpus(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, draftsDir+"/itd-10-alpha.md", draftWithAC("itd-10", "alpha"))
	writeFile(t, root, plannedDir+"/itd-2-beta.md",
		"---\nid: itd-2\nslug: beta\nspec_id: spc-1\nkind: standalone\n---\n# beta\n")

	c, err := Load(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(c.Intents) != 2 {
		t.Fatalf("expected 2 intents, got %d: %+v", len(c.Intents), c.Intents)
	}
	it, ok := c.Lookup("itd-2")
	if !ok || it.Bucket != "planned" || it.SpecID != "spc-1" || it.Kind != "standalone" {
		t.Fatalf("Lookup(itd-2) = %+v, %v", it, ok)
	}
	if it, ok := c.Lookup("itd-10"); !ok || it.Bucket != "drafts" {
		t.Fatalf("Lookup(itd-10) = %+v, %v", it, ok)
	}
}

func TestLoadMissingDirIsEmpty(t *testing.T) {
	c, err := Load(t.TempDir())
	if err != nil {
		t.Fatalf("Load on missing intents dir must be soft: %v", err)
	}
	if len(c.Intents) != 0 {
		t.Fatalf("expected empty corpus, got %+v", c.Intents)
	}
}

func TestLoadMalformedIsHardError(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, draftsDir+"/itd-1-broken.md", "# no frontmatter, no id\n")
	if _, err := Load(root); err == nil {
		t.Fatal("Load must hard-error on an intent file with no well-formed id")
	}
}

func TestPlanHappyPath(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, draftsDir+"/itd-10-alpha.md", draftWithAC("itd-10", "alpha"))

	res, err := Plan(root, "itd-10")
	if err != nil {
		t.Fatal(err)
	}
	if res.Spec.ID != "spc-1" || res.Spec.Intent != "itd-10" {
		t.Fatalf("Plan spec = %+v", res.Spec)
	}
	if res.Intent.Bucket != "planned" || res.Intent.SpecID != "spc-1" || res.Intent.Kind != "standalone" {
		t.Fatalf("Plan intent = %+v", res.Intent)
	}

	// The draft file is gone; the planned file exists with both link sides.
	if _, err := os.Stat(filepath.Join(root, draftsDir, "itd-10-alpha.md")); !os.IsNotExist(err) {
		t.Fatal("draft file should be gone after Plan")
	}
	body, err := os.ReadFile(filepath.Join(root, plannedDir, "itd-10-alpha.md"))
	if err != nil {
		t.Fatalf("planned file should exist: %v", err)
	}
	f := frontmatter.Fields(strings.Split(string(body), "\n"))
	if f["spec_id"].Value != "spc-1" {
		t.Fatalf("planned intent spec_id = %q, want spc-1\n%s", f["spec_id"].Value, body)
	}
	if f["kind"].Value != "standalone" {
		t.Fatalf("planned intent kind = %q, want standalone\n%s", f["kind"].Value, body)
	}
	// The spec file carries the reciprocal intent link.
	sbody, err := os.ReadFile(filepath.Join(root, specsOpen, "spc-1-alpha.md"))
	if err != nil {
		t.Fatalf("spec file should exist: %v", err)
	}
	if !strings.Contains(string(sbody), "intent: itd-10") {
		t.Fatalf("spec file missing reciprocal link:\n%s", sbody)
	}
}

// TestPlanResidualPassesRecordLint proves that a drafts->planned Plan leaves a
// frontmatter/bucket state the intent_lifecycle record-lint rule accepts — the
// DoD's "make record-lint stays green" guarantee, checked through the real lint
// engine over the fixture.
func TestPlanResidualPassesRecordLint(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, draftsDir+"/itd-10-alpha.md", draftWithAC("itd-10", "alpha"))
	if _, err := Plan(root, "itd-10"); err != nil {
		t.Fatal(err)
	}

	cfg := lint.Config{
		Roots: []string{".abcd/development"},
		Rules: map[string]lint.RuleConfig{
			"intent_lifecycle": {Enabled: true, Severity: "blocker", IntentsDir: "intents"},
		},
	}
	findings, err := lint.Lint(cfg, root)
	if err != nil {
		t.Fatal(err)
	}
	for _, fnd := range findings {
		if fnd.RuleID == "intent_lifecycle" {
			t.Fatalf("planned intent violates intent_lifecycle: %s:%d %s", fnd.File, fnd.Line, fnd.Message)
		}
	}
}

func TestPlanRefusesNoAcceptanceCriteria(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, draftsDir+"/itd-10-alpha.md",
		"---\nid: itd-10\nslug: alpha\nspec_id: null\nkind: null\n---\n# alpha\n\nNo AC section here.\n")

	if _, err := Plan(root, "itd-10"); err == nil {
		t.Fatal("Plan must refuse an intent with no Acceptance Criteria")
	}
	// Nothing moved, no spec minted.
	if _, err := os.Stat(filepath.Join(root, draftsDir, "itd-10-alpha.md")); err != nil {
		t.Fatal("draft must remain in place after refusal")
	}
	if _, err := os.Stat(filepath.Join(root, specsOpen)); !os.IsNotExist(err) {
		t.Fatal("no spec should be minted on refusal")
	}
}

func TestPlanRefusesEmptyAcceptanceCriteria(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, draftsDir+"/itd-10-alpha.md",
		"---\nid: itd-10\nslug: alpha\nspec_id: null\nkind: null\n---\n# alpha\n\n## Acceptance Criteria\n\n## Next Section\n\nbody\n")
	if _, err := Plan(root, "itd-10"); err == nil {
		t.Fatal("Plan must refuse an intent whose Acceptance Criteria section is empty")
	}
}

func TestPlanRefusesNonDraft(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, plannedDir+"/itd-10-alpha.md",
		"---\nid: itd-10\nslug: alpha\nspec_id: null\nkind: standalone\n---\n# alpha\n\n## Acceptance Criteria\n\n- ok\n")
	if _, err := Plan(root, "itd-10"); err == nil {
		t.Fatal("Plan must refuse an intent that is not in drafts/")
	}
}

// TestPlanReusesExistingSpecForIntent proves Plan is retry-safe: when a spec
// already realises the intent (e.g. a prior Plan minted it but the
// drafts->planned rename failed, leaving the intent a null-spec_id draft),
// re-running Plan reuses that spec instead of minting a duplicate.
func TestPlanReusesExistingSpecForIntent(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, draftsDir+"/itd-10-alpha.md", draftWithAC("itd-10", "alpha"))
	// Pre-create the spec for this intent; the draft is still an unlinked draft.
	sp, err := spec.Create(root, "itd-10", "alpha")
	if err != nil {
		t.Fatal(err)
	}

	res, err := Plan(root, "itd-10")
	if err != nil {
		t.Fatal(err)
	}
	if res.Spec.ID != sp.ID {
		t.Fatalf("Plan minted a new spec %q, want reuse of %q", res.Spec.ID, sp.ID)
	}
	if res.Intent.Bucket != "planned" || res.Intent.SpecID != sp.ID {
		t.Fatalf("Plan intent = %+v", res.Intent)
	}

	// Exactly one spec realises the intent (no duplicate minted).
	store, err := spec.Load(root)
	if err != nil {
		t.Fatal(err)
	}
	n := 0
	for _, s := range store.Specs {
		if s.Intent == "itd-10" {
			n++
		}
	}
	if n != 1 {
		t.Fatalf("expected exactly 1 spec for itd-10, got %d: %+v", n, store.Specs)
	}
}

// TestPlanRefusesDraftWithSpecID rejects a draft whose frontmatter already
// carries a non-null spec_id (half-planned / lint-invalid): Plan must not mint
// another spec for it.
func TestPlanRefusesDraftWithSpecID(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, draftsDir+"/itd-10-alpha.md",
		"---\nid: itd-10\nslug: alpha\nspec_id: spc-1\nkind: standalone\n---\n# alpha\n\n## Acceptance Criteria\n\n- ok\n")
	if _, err := Plan(root, "itd-10"); err == nil {
		t.Fatal("Plan must refuse a draft that already has a non-null spec_id")
	}
}

// TestPlanRefusesWhenPlannedTargetExists proves Plan fails closed rather than
// clobbering a same-name file already sitting in planned/.
func TestPlanRefusesWhenPlannedTargetExists(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, draftsDir+"/itd-10-alpha.md", draftWithAC("itd-10", "alpha"))
	writeFile(t, root, plannedDir+"/itd-10-alpha.md",
		"---\nid: itd-10\nslug: alpha\nspec_id: null\nkind: standalone\n---\n# pre-existing\n")
	if _, err := Plan(root, "itd-10"); err == nil {
		t.Fatal("Plan must refuse to overwrite an existing planned target")
	}
	// The pre-existing planned file is untouched.
	body, err := os.ReadFile(filepath.Join(root, plannedDir, "itd-10-alpha.md"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(body), "pre-existing") {
		t.Fatalf("planned target was clobbered:\n%s", body)
	}
}

func TestPlanRejectsBadID(t *testing.T) {
	root := t.TempDir()
	if _, err := Plan(root, "itd-../../etc"); err == nil {
		t.Fatal("Plan must reject a traversal id")
	}
}

func TestLinkHappyPath(t *testing.T) {
	root := t.TempDir()
	// A planned intent with no spec_id yet, and a spec that already realises it.
	writeFile(t, root, plannedDir+"/itd-10-alpha.md",
		"---\nid: itd-10\nslug: alpha\nspec_id: null\nkind: standalone\n---\n# alpha\n")
	writeFile(t, root, specsOpen+"/spc-3-alpha.md",
		"---\nid: spc-3\nslug: alpha\nintent: itd-10\n---\n# alpha\n")

	res, err := Link(root, "itd-10", "spc-3")
	if err != nil {
		t.Fatal(err)
	}
	if res.Intent.SpecID != "spc-3" {
		t.Fatalf("Link intent spec_id = %q, want spc-3", res.Intent.SpecID)
	}
	body, err := os.ReadFile(filepath.Join(root, plannedDir, "itd-10-alpha.md"))
	if err != nil {
		t.Fatal(err)
	}
	if f := frontmatter.Fields(strings.Split(string(body), "\n")); f["spec_id"].Value != "spc-3" {
		t.Fatalf("spec_id not written: %q\n%s", f["spec_id"].Value, body)
	}
}

func TestLinkMismatchFails(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, plannedDir+"/itd-10-alpha.md",
		"---\nid: itd-10\nslug: alpha\nspec_id: null\nkind: standalone\n---\n# alpha\n")
	// The spec realises a DIFFERENT intent.
	writeFile(t, root, specsOpen+"/spc-3-other.md",
		"---\nid: spc-3\nslug: other\nintent: itd-99\n---\n# other\n")
	if _, err := Link(root, "itd-10", "spc-3"); err == nil {
		t.Fatal("Link must fail closed when the spec realises a different intent")
	}
}

func TestLinkMissingSpecFails(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, plannedDir+"/itd-10-alpha.md",
		"---\nid: itd-10\nslug: alpha\nspec_id: null\nkind: standalone\n---\n# alpha\n")
	if _, err := Link(root, "itd-10", "spc-9"); err == nil {
		t.Fatal("Link must fail when the spec does not exist")
	}
}

func TestStatusCounts(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, draftsDir+"/itd-10-alpha.md", draftWithAC("itd-10", "alpha"))
	writeFile(t, root, plannedDir+"/itd-2-beta.md",
		"---\nid: itd-2\nslug: beta\nspec_id: spc-1\nkind: standalone\n---\n# beta\n")
	writeFile(t, root, specsOpen+"/spc-1-beta.md",
		"---\nid: spc-1\nslug: beta\nintent: itd-2\n---\n# beta\n")
	writeFile(t, root, specsClosed+"/spc-2-old.md",
		"---\nid: spc-2\nslug: old\nintent: itd-7\n---\n# old\n")

	v, err := Status(root)
	if err != nil {
		t.Fatal(err)
	}
	if v.Buckets["drafts"] != 1 || v.Buckets["planned"] != 1 {
		t.Fatalf("bucket counts = %+v", v.Buckets)
	}
	if v.SpecsOpen != 1 || v.SpecsClosed != 1 {
		t.Fatalf("spec counts open=%d closed=%d", v.SpecsOpen, v.SpecsClosed)
	}
	if len(v.Linked) != 1 || v.Linked[0].Intent != "itd-2" || v.Linked[0].Spec != "spc-1" {
		t.Fatalf("linked pairs = %+v", v.Linked)
	}
}
