package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// writeRepoFile writes content to root/rel, creating parent directories.
func writeRepoFile(t *testing.T, root, rel, content string) {
	t.Helper()
	abs := filepath.Join(root, rel)
	if err := os.MkdirAll(filepath.Dir(abs), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(abs, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

const (
	cliDrafts    = ".abcd/development/intents/drafts"
	cliPlanned   = ".abcd/development/intents/planned"
	cliSpecsOpen = ".abcd/development/specs/open"
)

func cliDraftWithAC(id, slug string) string {
	return "---\nid: " + id + "\nslug: " + slug + "\nspec_id: null\nkind: null\n---\n# " + slug +
		"\n\n## Acceptance Criteria\n\n- **Given** x, **when** y, **then** z.\n"
}

func TestIntentBareText(t *testing.T) {
	repo := t.TempDir()
	t.Chdir(repo)
	writeRepoFile(t, repo, cliDrafts+"/itd-10-alpha.md", cliDraftWithAC("itd-10", "alpha"))
	writeRepoFile(t, repo, cliPlanned+"/itd-2-beta.md",
		"---\nid: itd-2\nslug: beta\nspec_id: spc-1\nkind: standalone\n---\n# beta\n")

	out := string(runCLI(t, "intent"))
	if !strings.Contains(out, "drafts 1") || !strings.Contains(out, "planned 1") {
		t.Fatalf("bare intent status missing counts:\n%s", out)
	}
}

func TestIntentBareJSON(t *testing.T) {
	repo := t.TempDir()
	t.Chdir(repo)
	writeRepoFile(t, repo, cliDrafts+"/itd-10-alpha.md", cliDraftWithAC("itd-10", "alpha"))

	out := runCLI(t, "intent", "--json")
	var got struct {
		Buckets map[string]int `json:"buckets"`
	}
	if err := json.Unmarshal(out, &got); err != nil {
		t.Fatalf("intent --json not JSON: %v\n%s", err, out)
	}
	if got.Buckets["drafts"] != 1 {
		t.Fatalf("intent --json drafts = %d, want 1\n%s", got.Buckets["drafts"], out)
	}
}

func TestIntentPlanHappy(t *testing.T) {
	repo := t.TempDir()
	t.Chdir(repo)
	writeRepoFile(t, repo, cliDrafts+"/itd-10-alpha.md", cliDraftWithAC("itd-10", "alpha"))

	out := runCLI(t, "intent", "plan", "itd-10", "--json")
	var got struct {
		Intent struct {
			Bucket string `json:"bucket"`
			SpecID string `json:"spec_id"`
		} `json:"intent"`
		Spec struct {
			ID string `json:"id"`
		} `json:"spec"`
	}
	if err := json.Unmarshal(out, &got); err != nil {
		t.Fatalf("plan --json not JSON: %v\n%s", err, out)
	}
	if got.Intent.Bucket != "planned" || got.Spec.ID != "spc-1" || got.Intent.SpecID != "spc-1" {
		t.Fatalf("plan result = %+v", got)
	}
	// The draft moved and the spec landed on disk.
	if _, err := os.Stat(filepath.Join(repo, cliPlanned, "itd-10-alpha.md")); err != nil {
		t.Fatalf("planned file missing: %v", err)
	}
	if _, err := os.Stat(filepath.Join(repo, cliSpecsOpen, "spc-1-alpha.md")); err != nil {
		t.Fatalf("spec file missing: %v", err)
	}
}

func TestIntentPlanRefusesNoAC(t *testing.T) {
	repo := t.TempDir()
	t.Chdir(repo)
	writeRepoFile(t, repo, cliDrafts+"/itd-10-alpha.md",
		"---\nid: itd-10\nslug: alpha\nspec_id: null\nkind: null\n---\n# alpha\n\nno criteria\n")
	if _, err := runCLIErr(t, "intent", "plan", "itd-10"); err == nil {
		t.Fatal("plan without Acceptance Criteria must exit non-zero")
	}
}

func TestIntentPlanRefusesNonDraft(t *testing.T) {
	repo := t.TempDir()
	t.Chdir(repo)
	writeRepoFile(t, repo, cliPlanned+"/itd-10-alpha.md",
		"---\nid: itd-10\nslug: alpha\nspec_id: null\nkind: standalone\n---\n# alpha\n\n## Acceptance Criteria\n\n- ok\n")
	if _, err := runCLIErr(t, "intent", "plan", "itd-10"); err == nil {
		t.Fatal("plan on a non-draft intent must exit non-zero")
	}
}

func TestIntentLinkHappy(t *testing.T) {
	repo := t.TempDir()
	t.Chdir(repo)
	writeRepoFile(t, repo, cliPlanned+"/itd-10-alpha.md",
		"---\nid: itd-10\nslug: alpha\nspec_id: null\nkind: standalone\n---\n# alpha\n")
	writeRepoFile(t, repo, cliSpecsOpen+"/spc-3-alpha.md",
		"---\nid: spc-3\nslug: alpha\nintent: itd-10\n---\n# alpha\n")

	out := runCLI(t, "intent", "link", "itd-10", "spc-3", "--json")
	var got struct {
		Intent struct {
			SpecID string `json:"spec_id"`
		} `json:"intent"`
	}
	if err := json.Unmarshal(out, &got); err != nil {
		t.Fatalf("link --json not JSON: %v\n%s", err, out)
	}
	if got.Intent.SpecID != "spc-3" {
		t.Fatalf("link spec_id = %q, want spc-3", got.Intent.SpecID)
	}
}

func TestIntentLinkMismatchErrors(t *testing.T) {
	repo := t.TempDir()
	t.Chdir(repo)
	writeRepoFile(t, repo, cliPlanned+"/itd-10-alpha.md",
		"---\nid: itd-10\nslug: alpha\nspec_id: null\nkind: standalone\n---\n# alpha\n")
	writeRepoFile(t, repo, cliSpecsOpen+"/spc-3-other.md",
		"---\nid: spc-3\nslug: other\nintent: itd-99\n---\n# other\n")
	if _, err := runCLIErr(t, "intent", "link", "itd-10", "spc-3"); err == nil {
		t.Fatal("link with a spec realising a different intent must exit non-zero")
	}
}

func TestSpecBareText(t *testing.T) {
	repo := t.TempDir()
	t.Chdir(repo)
	writeRepoFile(t, repo, cliSpecsOpen+"/spc-1-alpha.md",
		"---\nid: spc-1\nslug: alpha\nintent: itd-10\n---\n# alpha\n")

	out := string(runCLI(t, "spec"))
	if !strings.Contains(out, "open 1") || !strings.Contains(out, "spc-1") {
		t.Fatalf("bare spec status missing spec:\n%s", out)
	}
}

func TestSpecCloseHappy(t *testing.T) {
	repo := t.TempDir()
	t.Chdir(repo)
	// spec close now reconciles the linked intent, so the intent must exist and
	// be planned+linked back to this spec.
	writeRepoFile(t, repo, cliPlanned+"/itd-10-alpha.md",
		"---\nid: itd-10\nslug: alpha\nspec_id: spc-1\nkind: standalone\n---\n# alpha\n\n## Acceptance Criteria\n\n- ok\n")
	writeRepoFile(t, repo, cliSpecsOpen+"/spc-1-alpha.md",
		"---\nid: spc-1\nslug: alpha\nintent: itd-10\n---\n# alpha\n")

	out := runCLI(t, "spec", "close", "spc-1", "--json")
	var got struct {
		Spec struct {
			Status string `json:"status"`
			Path   string `json:"path"`
		} `json:"spec"`
		Intent struct {
			Bucket string `json:"bucket"`
		} `json:"intent"`
		IntentMoved bool   `json:"intent_moved"`
		From        string `json:"from"`
		To          string `json:"to"`
	}
	if err := json.Unmarshal(out, &got); err != nil {
		t.Fatalf("spec close --json not JSON: %v\n%s", err, out)
	}
	if got.Spec.Status != "closed" {
		t.Fatalf("spec close status = %q, want closed", got.Spec.Status)
	}
	if !got.IntentMoved || got.From != "planned" || got.To != "shipped" || got.Intent.Bucket != "shipped" {
		t.Fatalf("reconcile envelope = %+v", got)
	}
	if _, err := os.Stat(filepath.Join(repo, ".abcd/development/specs/closed", "spc-1-alpha.md")); err != nil {
		t.Fatalf("closed spec file missing: %v", err)
	}
	if _, err := os.Stat(filepath.Join(repo, ".abcd/development/intents/shipped", "itd-10-alpha.md")); err != nil {
		t.Fatalf("shipped intent file missing: %v", err)
	}
}

// TestSpecCloseReconcileText checks the human-readable close render names the
// intent that moved and its from->to.
func TestSpecCloseReconcileText(t *testing.T) {
	repo := t.TempDir()
	t.Chdir(repo)
	writeRepoFile(t, repo, cliPlanned+"/itd-10-alpha.md",
		"---\nid: itd-10\nslug: alpha\nspec_id: spc-1\nkind: standalone\n---\n# alpha\n\n## Acceptance Criteria\n\n- ok\n")
	writeRepoFile(t, repo, cliSpecsOpen+"/spc-1-alpha.md",
		"---\nid: spc-1\nslug: alpha\nintent: itd-10\n---\n# alpha\n")

	out := string(runCLI(t, "spec", "close", "spc-1"))
	if !strings.Contains(out, "itd-10") || !strings.Contains(out, "planned") || !strings.Contains(out, "shipped") {
		t.Fatalf("close text missing reconcile detail:\n%s", out)
	}
}

func TestSpecCloseMissingErrors(t *testing.T) {
	repo := t.TempDir()
	t.Chdir(repo)
	if _, err := runCLIErr(t, "spec", "close", "spc-99"); err == nil {
		t.Fatal("closing a missing spec must exit non-zero")
	}
}

// runCLISplit executes the command tree with stdout and stderr captured
// separately, so a deprecation warning routed to stderr can be asserted distinct
// from the stdout artefact.
func runCLISplit(t *testing.T, args ...string) (string, string, error) {
	t.Helper()
	cmd := NewRootCommand()
	var out, errb bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&errb)
	cmd.SetArgs(args)
	err := cmd.Execute()
	return out.String(), errb.String(), err
}

// TestIntentQuotedTextCreates is itd-46 AC1 at the CLI: `abcd intent "<text>"`
// files a new drafts/itd-N-<slug>.md seeded from the text — no `new` sub-verb.
func TestIntentQuotedTextCreates(t *testing.T) {
	repo := t.TempDir()
	t.Chdir(repo)

	out := runCLI(t, "intent", "I want users to feel the card respects their time", "--json")
	var got struct {
		ID     string `json:"id"`
		Slug   string `json:"slug"`
		Bucket string `json:"bucket"`
		Path   string `json:"path"`
	}
	if err := json.Unmarshal(out, &got); err != nil {
		t.Fatalf("create --json not JSON: %v\n%s", err, out)
	}
	if got.ID != "itd-1" || got.Bucket != "drafts" {
		t.Fatalf("create result = %+v, want itd-1 in drafts", got)
	}
	body, err := os.ReadFile(filepath.Join(repo, got.Path))
	if err != nil {
		t.Fatalf("created draft unreadable: %v", err)
	}
	if !strings.Contains(string(body), "I want users to feel the card respects their time") {
		t.Fatalf("seeded body missing the text:\n%s", body)
	}
}

// TestIntentNewAliasWarnsAndCreates is itd-46 AC2 (lean a): `abcd intent new
// "<text>"` routes to the same create path and prints a deprecation warning on
// stderr naming the new shape; the stdout artefact matches the sub-verb-free form.
func TestIntentNewAliasWarnsAndCreates(t *testing.T) {
	repo := t.TempDir()
	t.Chdir(repo)

	stdout, stderr, err := runCLISplit(t, "intent", "new", "a symmetric create path", "--json")
	if err != nil {
		t.Fatalf("intent new alias errored: %v\nstderr: %s", err, stderr)
	}
	var got struct {
		ID     string `json:"id"`
		Bucket string `json:"bucket"`
		Path   string `json:"path"`
	}
	if err := json.Unmarshal([]byte(stdout), &got); err != nil {
		t.Fatalf("alias stdout not JSON: %v\n%s", err, stdout)
	}
	if got.ID != "itd-1" || got.Bucket != "drafts" {
		t.Fatalf("alias create result = %+v, want itd-1 in drafts", got)
	}
	if !strings.Contains(stderr, "deprecat") {
		t.Fatalf("alias must warn on stderr about deprecation, got: %q", stderr)
	}
	if !strings.Contains(stderr, `intent "`) {
		t.Fatalf("deprecation warning must name the new quoted-text shape, got: %q", stderr)
	}
	// The warning is on stderr only — stdout stays the clean artefact.
	if strings.Contains(stdout, "deprecat") {
		t.Fatalf("deprecation warning leaked into stdout:\n%s", stdout)
	}
}

// TestIntentBareCreatesNothing is itd-46 AC3: bare `abcd intent` renders status +
// help and mutates nothing — no drafts file appears.
func TestIntentBareCreatesNothing(t *testing.T) {
	repo := t.TempDir()
	t.Chdir(repo)

	out := string(runCLI(t, "intent"))
	if !strings.Contains(out, "abcd intent") {
		t.Fatalf("bare intent missing status render:\n%s", out)
	}
	if entries, _ := os.ReadDir(filepath.Join(repo, cliDrafts)); len(entries) != 0 {
		t.Fatalf("bare intent created %d drafts files, want 0", len(entries))
	}
}

// TestBareHelpsCarryDecisionRule is itd-46 AC5: both bare-form outputs carry the
// one-line capture-vs-intent decision rule so a user knows which ledger to reach.
func TestBareHelpsCarryDecisionRule(t *testing.T) {
	repo := t.TempDir()
	t.Chdir(repo)

	intentOut := string(runCLI(t, "intent"))
	if !strings.Contains(intentOut, "user-facing change") || !strings.Contains(intentOut, "nitpick") {
		t.Fatalf("bare intent help missing decision rule:\n%s", intentOut)
	}
	captureOut := string(runCLI(t, "capture"))
	if !strings.Contains(captureOut, "user-facing change") || !strings.Contains(captureOut, "nitpick") {
		t.Fatalf("bare capture help missing decision rule:\n%s", captureOut)
	}
}
