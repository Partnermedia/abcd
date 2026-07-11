package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestVersionJSON proves the CLI -> core -> JSON round-trip the Phase 0 exit
// criterion requires.
func TestVersionJSON(t *testing.T) {
	out := runCLI(t, "version", "--json")

	var got map[string]any
	if err := json.Unmarshal(out, &got); err != nil {
		t.Fatalf("output is not JSON: %v\n%s", err, out)
	}
	if got["name"] != "abcd" {
		t.Fatalf("name = %v, want abcd", got["name"])
	}
	if got["version"] == "" || got["version"] == nil {
		t.Fatalf("version missing: %v", got)
	}
}

func TestVersionText(t *testing.T) {
	out := runCLI(t, "version")
	if !strings.HasPrefix(string(out), "abcd ") {
		t.Fatalf("text output = %q, want it to start with \"abcd \"", out)
	}
}

func TestBareStatusJSON(t *testing.T) {
	out := runCLI(t, "--json")
	var got map[string]any
	if err := json.Unmarshal(out, &got); err != nil {
		t.Fatalf("bare status output is not JSON: %v\n%s", err, out)
	}
	if _, ok := got["dir"]; !ok {
		t.Fatalf("status JSON missing dir: %v", got)
	}
}

func TestRulesBareText(t *testing.T) {
	out := string(runCLI(t, "rules"))
	for _, want := range []string{"COMMITTING", "PII"} {
		if !strings.Contains(out, want) {
			t.Fatalf("bare `rules` missing %q:\n%s", want, out)
		}
	}
}

func TestRulesBareJSON(t *testing.T) {
	out := runCLI(t, "rules", "--json")
	var got struct {
		Disabled bool             `json:"disabled"`
		Domains  []map[string]any `json:"domains"`
	}
	if err := json.Unmarshal(out, &got); err != nil {
		t.Fatalf("rules --json not JSON: %v\n%s", err, out)
	}
	if len(got.Domains) == 0 {
		t.Fatalf("rules --json returned no domains: %s", out)
	}
	found := false
	for _, d := range got.Domains {
		if d["name"] == "COMMITTING" {
			found = true
		}
	}
	if !found {
		t.Fatalf("rules --json missing COMMITTING: %s", out)
	}
}

func TestRulesScopedUppercasesArg(t *testing.T) {
	out := string(runCLI(t, "rules", "committing"))
	if !strings.Contains(out, "COMMITTING") {
		t.Fatalf("scoped `rules committing` missing COMMITTING:\n%s", out)
	}
	if strings.Contains(out, "## PII") {
		t.Fatalf("scoped render leaked another domain:\n%s", out)
	}
}

func TestRulesUnknownDomainErrors(t *testing.T) {
	if _, err := runCLIErr(t, "rules", "nosuch"); err == nil {
		t.Fatal("unknown domain must exit non-zero")
	}
}

// validHooksJSON is a structurally-sound plugin hook manifest for the hermetic
// plugin root, so the install path's hook-manifest verification passes.
const validHooksJSON = `{
  "hooks": {
    "UserPromptSubmit": [{"hooks": [{"type": "command", "command": "$CLAUDE_PLUGIN_ROOT/hooks/prompt_router_hook"}]}],
    "SessionStart":     [{"hooks": [{"type": "command", "command": "$CLAUDE_PLUGIN_ROOT/hooks/prompt_router_reset"}]}],
    "PreCompact":       [{"hooks": [{"type": "command", "command": "$CLAUDE_PLUGIN_ROOT/hooks/prompt_router_reset"}]}]
  }
}`

// hermeticRepo redirects HOME, the plugin root and the PATH symlink target to
// temp locations, chdirs into a fresh adoptable repo, and returns its path.
func hermeticRepo(t *testing.T) string {
	t.Helper()
	home := t.TempDir()
	pluginRoot := t.TempDir()
	if err := os.MkdirAll(filepath.Join(pluginRoot, "hooks"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(pluginRoot, "hooks", "hooks.json"), []byte(validHooksJSON), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(pluginRoot, "abcd"), []byte("#!/bin/sh\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("HOME", home)
	t.Setenv("ABCD_PLUGIN_ROOT", pluginRoot)
	t.Setenv("CLAUDE_PLUGIN_ROOT", "")
	t.Setenv("ABCD_BIN_TARGET", filepath.Join(t.TempDir(), "bin", "abcd"))

	repo := t.TempDir()
	if err := os.Mkdir(filepath.Join(repo, ".git"), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Chdir(repo)
	return repo
}

// TestAhoyInstallWiredAndIdempotent proves `abcd ahoy install` reaches the core
// engine from the CLI front door (the Phase 1 install milestone) and that a
// re-run is an exact no-op.
func TestAhoyInstallWiredAndIdempotent(t *testing.T) {
	repo := hermeticRepo(t)

	out := runCLI(t, "ahoy", "install", "--yes", "--adopt",
		"--visibility", "private", "--docs-target", "both",
		"--oracle-backend", "host-delegated", "--scan-deep", "false", "--json")
	var res struct {
		Status string   `json:"status"`
		Writes []string `json:"writes"`
	}
	if err := json.Unmarshal(out, &res); err != nil {
		t.Fatalf("install output not JSON: %v\n%s", err, out)
	}
	if res.Status != "clean" {
		t.Fatalf("install status = %q, want clean\n%s", res.Status, out)
	}
	// The marker block reached disk via the CLI path.
	body, err := os.ReadFile(filepath.Join(repo, "CLAUDE.md"))
	if err != nil {
		t.Fatalf("CLAUDE.md not written: %v", err)
	}
	if !strings.Contains(string(body), "<!-- BEGIN ABCD -->") {
		t.Fatalf("CLAUDE.md has no marker block:\n%s", body)
	}

	// Second run is an exact no-op.
	out2 := runCLI(t, "ahoy", "install", "--yes", "--adopt",
		"--visibility", "private", "--docs-target", "both",
		"--oracle-backend", "host-delegated", "--scan-deep", "false", "--json")
	var res2 struct {
		Status string   `json:"status"`
		Writes []string `json:"writes"`
	}
	if err := json.Unmarshal(out2, &res2); err != nil {
		t.Fatalf("re-install output not JSON: %v\n%s", err, out2)
	}
	if res2.Status != "already_up_to_date" {
		t.Fatalf("re-install status = %q, want already_up_to_date", res2.Status)
	}
	if len(res2.Writes) != 0 {
		t.Fatalf("re-install wrote files: %v", res2.Writes)
	}
}

func runCLI(t *testing.T, args ...string) []byte {
	t.Helper()
	return runCLIStdin(t, "", args...)
}

// runCLIStdin runs the CLI with stdin bound to `stdin`, so commands that read a
// payload from "-" (e.g. `history capture -`) can be exercised end-to-end.
func runCLIStdin(t *testing.T, stdin string, args ...string) []byte {
	t.Helper()
	cmd := NewRootCommand()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetIn(strings.NewReader(stdin))
	cmd.SetArgs(args)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute %v: %v\n%s", args, err, out.String())
	}
	return out.Bytes()
}

func gitCmd(t *testing.T, repo string, args ...string) string {
	t.Helper()
	full := append([]string{"-C", repo}, args...)
	out, err := exec.Command("git", full...).CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
	return strings.TrimSpace(string(out))
}

// TestHistoryCaptureWiredAndRedacts proves the Finding-B wiring: `abcd history
// capture` reaches history.Capture from the CLI front door, redacts a planted
// secret, and stores the record on disk. Before this change history.Capture was
// dead scaffolding — no CLI subverb reached it.
func TestHistoryCaptureWiredAndRedacts(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	repo := t.TempDir()
	gitCmd(t, repo, "init")
	gitCmd(t, repo, "config", "user.email", "test@example.com")
	gitCmd(t, repo, "config", "user.name", "Test User")
	if err := os.WriteFile(filepath.Join(repo, "f.txt"), []byte("x\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	gitCmd(t, repo, "add", ".")
	gitCmd(t, repo, "commit", "-m", "init")
	t.Chdir(repo)

	// Create the store dir exactly as `abcd install` would (Capture never
	// bootstraps it).
	rootSHA := gitCmd(t, repo, "rev-list", "--max-parents=0", "HEAD")
	tdir := filepath.Join(home, ".abcd", "history", rootSHA, "transcripts")
	if err := os.MkdirAll(tdir, 0o755); err != nil {
		t.Fatal(err)
	}

	pat := "ghp_" + strings.Repeat("b", 40)
	transcript := "user: deploy with token " + pat + "\nassistant: done\n"

	out := runCLIStdin(t, transcript, "history", "capture", "--session", "sess-wired", "--json")

	var res struct {
		Record struct {
			Path      string `json:"path"`
			SessionID string `json:"session_id"`
			Secrets   int    `json:"redacted_secrets"`
		} `json:"record"`
		Wrote bool `json:"wrote"`
	}
	if err := json.Unmarshal(out, &res); err != nil {
		t.Fatalf("capture output not JSON: %v\n%s", err, out)
	}
	if !res.Wrote {
		t.Fatalf("expected Wrote=true on first capture\n%s", out)
	}
	if res.Record.SessionID != "sess-wired" {
		t.Errorf("session id = %q, want sess-wired", res.Record.SessionID)
	}
	if res.Record.Secrets < 1 {
		t.Errorf("expected >=1 secret redaction counted, got %d", res.Record.Secrets)
	}
	body, err := os.ReadFile(res.Record.Path)
	if err != nil {
		t.Fatalf("stored record unreadable: %v", err)
	}
	if bytes.Contains(body, []byte(pat)) {
		t.Errorf("planted secret leaked into the stored record:\n%s", body)
	}
}

// TestCaptureBlockedByWiredAndAnnotated proves the --blocked-by flag reaches
// capture.Capture from the CLI (writing the dependency edge), that an invalid
// token is rejected at the boundary, and that the derived-priority view renders
// unblocked-first with a [blocked-by …] annotation on the blocked row.
func TestCaptureBlockedByWiredAndAnnotated(t *testing.T) {
	repo := t.TempDir()
	t.Chdir(repo)

	// iss-1: the blocker target (minor, unblocked).
	out := runCLI(t, "capture", "root cause", "--slug", "root", "--json")
	var r1 struct {
		ID   string `json:"id"`
		Path string `json:"path"`
	}
	if err := json.Unmarshal(out, &r1); err != nil {
		t.Fatalf("capture output not JSON: %v\n%s", err, out)
	}
	if r1.ID != "iss-1" {
		t.Fatalf("first id = %q want iss-1", r1.ID)
	}

	// iss-2: critical but blocked by the still-open iss-1.
	out2 := runCLI(t, "capture", "dependent thing", "--slug", "dep",
		"--severity", "critical", "--blocked-by", "iss-1", "--json")
	var r2 struct {
		ID   string `json:"id"`
		Path string `json:"path"`
	}
	if err := json.Unmarshal(out2, &r2); err != nil {
		t.Fatalf("blocked capture output not JSON: %v\n%s", err, out2)
	}
	if r2.ID != "iss-2" {
		t.Fatalf("second id = %q want iss-2", r2.ID)
	}
	// The edge reached disk.
	body, err := os.ReadFile(r2.Path)
	if err != nil {
		t.Fatalf("iss-2 unreadable: %v", err)
	}
	if !strings.Contains(string(body), "blocked_by: [iss-1]") {
		t.Fatalf("blocked_by not written to iss-2:\n%s", body)
	}

	// Derived view: unblocked iss-1 ahead of the blocked, annotated iss-2.
	list := string(runCLI(t, "capture", "list", "--open"))
	i1 := strings.Index(list, "iss-1")
	i2 := strings.Index(list, "iss-2")
	if i1 < 0 || i2 < 0 || i1 > i2 {
		t.Fatalf("expected iss-1 before iss-2 (unblocked-first):\n%s", list)
	}
	if !strings.Contains(list, "[blocked-by iss-1]") {
		t.Fatalf("expected [blocked-by iss-1] annotation:\n%s", list)
	}

	// An invalid --blocked-by token is rejected at the boundary.
	if _, err := runCLIErr(t, "capture", "bad edge", "--blocked-by", "bogus"); err == nil {
		t.Fatalf("expected error for invalid --blocked-by token")
	}
}

// runCLIErr executes the command tree and returns its stdout/stderr plus the
// error, so a gate's non-zero exit can be asserted rather than fataled on.
func runCLIErr(t *testing.T, args ...string) ([]byte, error) {
	t.Helper()
	cmd := NewRootCommand()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs(args)
	err := cmd.Execute()
	return out.Bytes(), err
}

const docsLintConfig = `{
  "roots": ["docs"],
  "banned_tokens": [
    {"id":"present_tense/previously","pattern":"(?i)\\bpreviously\\b","severity":"blocker","message":"change-narration"}
  ],
  "rules": {
    "links_resolve": {"enabled": true, "severity": "blocker"},
    "stray_root_docs": {"enabled": true, "severity": "blocker",
      "allowlist": ["README","AGENTS","CHANGELOG","CONTRIBUTING","SECURITY","LICENSE"]}
  }
}`

// TestDocsLintRootFlagFlagsDrift proves `docs lint --root/--config` runs layer 1
// over an arbitrary tree: a change-narration token, a broken cross-link, and a
// stray top-level markdown each surface as a blocker and drive a non-zero exit.
func TestDocsLintRootFlagFlagsDrift(t *testing.T) {
	root := t.TempDir()
	cfg := filepath.Join(root, "docs-lint.json")
	if err := os.WriteFile(cfg, []byte(docsLintConfig), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(root, "docs"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "docs", "bad.md"),
		[]byte("# Bad\n\nThis was previously X, now Y.\n\nSee [gone](./missing.md).\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "FOO.md"), []byte("# Stray\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	out, err := runCLIErr(t, "docs", "lint", "--json", "--config", cfg, "--root", root)
	if err == nil {
		t.Fatalf("expected non-zero exit on blockers, got nil\n%s", out)
	}
	var res struct {
		Findings []struct {
			RuleID   string `json:"RuleID"`
			Severity string `json:"Severity"`
		} `json:"findings"`
		Blockers int `json:"blockers"`
	}
	if err := json.Unmarshal(out, &res); err != nil {
		t.Fatalf("output not JSON: %v\n%s", err, out)
	}
	if res.Blockers < 3 {
		t.Fatalf("blockers = %d, want >= 3\n%s", res.Blockers, out)
	}
	rules := map[string]bool{}
	for _, f := range res.Findings {
		rules[f.RuleID] = true
	}
	for _, want := range []string{"present_tense/previously", "links_resolve", "stray_root_docs"} {
		if !rules[want] {
			t.Fatalf("expected a %s blocker, got findings %v", want, rules)
		}
	}
}

// TestDocsLintCleanTreePasses proves a present-tense, well-linked tree with no
// stray root docs exits zero under the same layer-1 config.
func TestDocsLintCleanTreePasses(t *testing.T) {
	root := t.TempDir()
	cfg := filepath.Join(root, "docs-lint.json")
	if err := os.WriteFile(cfg, []byte(docsLintConfig), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(root, "docs"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "docs", "peer.md"), []byte("# Peer\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "docs", "good.md"),
		[]byte("# Good\n\nThe pass grades docs against reality.\n\nSee [peer](./peer.md).\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	out, err := runCLIErr(t, "docs", "lint", "--json", "--config", cfg, "--root", root)
	if err != nil {
		t.Fatalf("clean tree should pass, got error: %v\n%s", err, out)
	}
	var res struct {
		Blockers int `json:"blockers"`
	}
	if err := json.Unmarshal(out, &res); err != nil {
		t.Fatalf("output not JSON: %v\n%s", err, out)
	}
	if res.Blockers != 0 {
		t.Fatalf("clean tree blockers = %d, want 0\n%s", res.Blockers, out)
	}
}
