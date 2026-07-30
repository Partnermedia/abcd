package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/REPPL/abcd-cli/internal/core/banlist"
)

const blDocsLint = `{
  "roots": ["docs", "README.md"],
  "banned_tokens": [
    {"id":"harness/gemini","pattern":"(?i)\\bgemini\\b","severity":"blocker","successor":"a generic term","allow_context":["(?i)<!--\\s*docs-lint:\\s*allow\\b"],"message":"m"}
  ],
  "rules": {"links_resolve": {"enabled": true, "severity": "blocker"}},
  "exempt_paths": [],
  "exempt_if_status": []
}
`

// blRepo stands up a repo with the public config present and, when body != "", a
// private banlist too.
func blRepo(t *testing.T, body string) string {
	t.Helper()
	repo := t.TempDir()
	write := func(rel, content string, perm os.FileMode) {
		p := filepath.Join(repo, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(content), perm); err != nil {
			t.Fatal(err)
		}
	}
	write(banlist.PublicConfigRelPath, blDocsLint, 0o644)
	if body != "" {
		// The store's first line declares the KEYED format, exactly as abcd writes
		// it: without that line every line below is a whole-line pattern, which is a
		// different store and has its own tests in the core package.
		write(banlist.PrivateRelPath, blFormatDecl+"\n"+body, 0o600)
	}
	return repo
}

// blFormatDecl is the private store's format declaration, spelled out here rather
// than imported: the surface tests assert against the format as a USER writes it,
// so a change to the constant that broke every existing store would fail here.
const blFormatDecl = "# abcd-banlist: keyed"

// TestBanlistBareRendersPrivateKeysOnly pins AC6 at the surface: the private layer
// renders by key, and the pattern value appears nowhere in the rendered text.
func TestBanlistBareRendersPrivateKeysOnly(t *testing.T) {
	t.Chdir(blRepo(t, "widget-partner   widgetworks\nlab-host         alice-laptop\\.example\\.com\n"))

	var stdout, stderr bytes.Buffer
	if code := Run([]string{"banlist"}, &stdout, &stderr); code != 0 {
		t.Fatalf("exit = %d, stderr = %s", code, stderr.String())
	}
	out := stdout.String()
	for _, want := range []string{"widget-partner", "lab-host", "harness/gemini"} {
		if !strings.Contains(out, want) {
			t.Errorf("render missing %q:\n%s", want, out)
		}
	}
	for _, leak := range []string{"widgetworks", "alice-laptop"} {
		if strings.Contains(out, leak) {
			t.Errorf("render leaks the private pattern %q:\n%s", leak, out)
		}
	}
}

// TestBanlistRendersTheHonestReach: the private layer cannot be enforced in CI, and
// the render says so rather than leaving a reader to assume coverage.
func TestBanlistRendersTheHonestReach(t *testing.T) {
	t.Chdir(blRepo(t, "widget-partner   widgetworks\n"))

	var stdout, stderr bytes.Buffer
	if code := Run([]string{"banlist"}, &stdout, &stderr); code != 0 {
		t.Fatalf("exit = %d, stderr = %s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "CI cannot enforce") {
		t.Errorf("render does not state the private layer's local-only reach:\n%s", stdout.String())
	}
}

// TestBanlistBareWarnsWhenThePrivateLayerIsInactive mirrors the hook's AC4 contract
// on the read surface: an absent store is reported as inactive, not as clean.
func TestBanlistBareWarnsWhenThePrivateLayerIsInactive(t *testing.T) {
	t.Chdir(blRepo(t, ""))

	var stdout, stderr bytes.Buffer
	if code := Run([]string{"banlist"}, &stdout, &stderr); code != 0 {
		t.Fatalf("exit = %d, stderr = %s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "INACTIVE") {
		t.Errorf("render does not report the absent private store as inactive:\n%s", stdout.String())
	}
}

// TestBanlistJSONCarriesBothLayersWithoutThePattern is the machine-readable half of
// the same redaction: --json is one call on the same result, and the private
// pattern is absent from it too.
func TestBanlistJSONCarriesBothLayersWithoutThePattern(t *testing.T) {
	t.Chdir(blRepo(t, "widget-partner   widgetworks\n"))

	var stdout, stderr bytes.Buffer
	if code := Run([]string{"banlist", "--json"}, &stdout, &stderr); code != 0 {
		t.Fatalf("exit = %d, stderr = %s", code, stderr.String())
	}
	var rep banlist.Report
	if err := json.Unmarshal(stdout.Bytes(), &rep); err != nil {
		t.Fatalf("banlist --json is not a Report: %v\n%s", err, stdout.String())
	}
	if len(rep.Private.Entries) != 1 || rep.Private.Entries[0].Key != "widget-partner" {
		t.Errorf("private entries = %+v", rep.Private.Entries)
	}
	if len(rep.Public.Entries) != 1 || rep.Public.Entries[0].ID != "harness/gemini" {
		t.Errorf("public entries = %+v", rep.Public.Entries)
	}
	if strings.Contains(stdout.String(), "widgetworks") {
		t.Errorf("banlist --json leaks the private pattern:\n%s", stdout.String())
	}
}

// TestBanlistAddPrivateThenRemove drives the private mutations through the surface.
func TestBanlistAddPrivateThenRemove(t *testing.T) {
	repo := blRepo(t, "")
	t.Chdir(repo)

	var stdout, stderr bytes.Buffer
	if code := Run([]string{"banlist", "add", "--private", "widget-partner", `widgetworks`}, &stdout, &stderr); code != 0 {
		t.Fatalf("add exit = %d, stderr = %s", code, stderr.String())
	}
	if strings.Contains(stdout.String(), "widgetworks") {
		t.Errorf("the add confirmation echoes the pattern:\n%s", stdout.String())
	}
	store, err := os.ReadFile(filepath.Join(repo, filepath.FromSlash(banlist.PrivateRelPath)))
	if err != nil {
		t.Fatalf("store not written: %v", err)
	}
	if !strings.Contains(string(store), "widget-partner widgetworks") {
		t.Errorf("store body = %q", store)
	}

	stdout.Reset()
	if code := Run([]string{"banlist", "remove", "--private", "widget-partner"}, &stdout, &stderr); code != 0 {
		t.Fatalf("remove exit = %d, stderr = %s", code, stderr.String())
	}
	store, err = os.ReadFile(filepath.Join(repo, filepath.FromSlash(banlist.PrivateRelPath)))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(store), "widget-partner") {
		t.Errorf("entry survived removal: %q", store)
	}
}

// TestBanlistAddPublicThenRemove drives the public mutations, and pins that the
// public layer renders in full — the pattern is committed, reviewable config.
func TestBanlistAddPublicThenRemove(t *testing.T) {
	repo := blRepo(t, "")
	t.Chdir(repo)

	var stdout, stderr bytes.Buffer
	if code := Run([]string{"banlist", "add", "--public", "widgetworks", `(?i)\bwidgetworks\b`}, &stdout, &stderr); code != 0 {
		t.Fatalf("add exit = %d, stderr = %s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "names/widgetworks") {
		t.Errorf("add confirmation does not name the entry id:\n%s", stdout.String())
	}

	stdout.Reset()
	if code := Run([]string{"banlist", "list", "--public"}, &stdout, &stderr); code != 0 {
		t.Fatalf("list exit = %d, stderr = %s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), `(?i)\bwidgetworks\b`) {
		t.Errorf("public list does not render the pattern in full:\n%s", stdout.String())
	}

	stdout.Reset()
	if code := Run([]string{"banlist", "remove", "--public", "widgetworks"}, &stdout, &stderr); code != 0 {
		t.Fatalf("remove exit = %d, stderr = %s", code, stderr.String())
	}
	cfg, err := os.ReadFile(filepath.Join(repo, filepath.FromSlash(banlist.PublicConfigRelPath)))
	if err != nil {
		t.Fatal(err)
	}
	if string(cfg) != blDocsLint {
		t.Errorf("add/remove round trip is not byte-identical:\n%s", cfg)
	}
}

// TestBanlistUsageErrorsExitTwo pins the usage contract: the layer is never
// guessed, and a refusal from the core is a usage error, not a crash.
func TestBanlistUsageErrorsExitTwo(t *testing.T) {
	t.Chdir(blRepo(t, "widget-partner   widgetworks\n"))

	cases := []struct {
		name string
		args []string
	}{
		{"add without a layer", []string{"banlist", "add", "k", "p"}},
		{"add with both layers", []string{"banlist", "add", "--private", "--public", "k", "p"}},
		{"remove without a layer", []string{"banlist", "remove", "k"}},
		{"list with both layers", []string{"banlist", "list", "--private", "--public"}},
		{"duplicate private key", []string{"banlist", "add", "--private", "widget-partner", "other"}},
		{"unknown private key", []string{"banlist", "remove", "--private", "nosuchkey"}},
		{"hand-curated public entry", []string{"banlist", "remove", "--public", "harness/gemini"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			if code := Run(tc.args, &stdout, &stderr); code != 2 {
				t.Errorf("exit = %d, want 2 (stderr: %s)", code, stderr.String())
			}
		})
	}
}
