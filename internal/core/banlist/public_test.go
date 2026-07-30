package banlist

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/REPPL/abcd-cli/internal/core/lint"
)

// realConfig copies this repo's own committed docs-lint config into a temp repo.
// Editing the REAL file's bytes is the point of these tests: a surgical editor
// proven only against a synthetic two-entry fixture is not proven at all.
func realConfig(t *testing.T) (root string, original []byte) {
	t.Helper()
	src := filepath.Join("..", "..", "..", ".abcd", "docs-lint.json")
	data, err := os.ReadFile(src)
	if err != nil {
		t.Skipf("repo docs-lint config unavailable: %v", err)
	}
	root = t.TempDir()
	dst := filepath.Join(root, filepath.FromSlash(PublicConfigRelPath))
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(dst, data, 0o644); err != nil {
		t.Fatal(err)
	}
	return root, data
}

func readConfig(t *testing.T, root string) []byte {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(PublicConfigRelPath)))
	if err != nil {
		t.Fatal(err)
	}
	return data
}

// TestAddPublicIsAMinimalDiff is the HARD constraint on the public layer: adding
// an entry adds exactly one line and touches nothing else — unrelated keys,
// ordering, and formatting survive byte-for-byte — and the result still loads
// through the linter's own strict validation.
func TestAddPublicIsAMinimalDiff(t *testing.T) {
	root, original := realConfig(t)

	res, err := AddPublic(AddPublicRequest{RepoRoot: root, Key: "widgetworks", Pattern: `(?i)\bwidgetworks\b`})
	if err != nil {
		t.Fatalf("AddPublic: %v", err)
	}
	if res.Entry.ID != "names/widgetworks" {
		t.Errorf("ID = %q, want names/widgetworks", res.Entry.ID)
	}
	if res.Entry.Severity != "blocker" {
		t.Errorf("Severity = %q, want blocker by default", res.Entry.Severity)
	}

	// The minimal diff a JSON array admits: one inserted line, and the line that
	// held the previous last element gains its separating comma. Nothing else in
	// the file may move, reflow, or reorder.
	before := strings.Split(string(original), "\n")
	after := strings.Split(string(readConfig(t, root)), "\n")
	if len(after) != len(before)+1 {
		t.Fatalf("line count %d -> %d, want exactly one added line", len(before), len(after))
	}
	at := -1
	for i := range before {
		if before[i] != after[i] {
			at = i
			break
		}
	}
	if at < 0 {
		t.Fatal("no line changed; the entry was not written")
	}
	if after[at] != before[at]+"," {
		t.Errorf("line %d changed beyond its separator:\n got %q\nwant %q", at+1, after[at], before[at]+",")
	}
	if !strings.Contains(after[at+1], "names/widgetworks") {
		t.Errorf("the inserted line is not the new entry: %q", after[at+1])
	}
	if got, want := strings.Join(after[at+2:], "\n"), strings.Join(before[at+1:], "\n"); got != want {
		t.Errorf("the tail of the file changed:\n got %q\nwant %q", got, want)
	}

	cfg, err := lint.LoadConfig(filepath.Join(root, filepath.FromSlash(PublicConfigRelPath)))
	if err != nil {
		t.Fatalf("the edited config no longer loads: %v", err)
	}
	var found bool
	for _, tok := range cfg.BannedTokens {
		if tok.ID == "names/widgetworks" {
			found = true
			if len(tok.AllowContext) == 0 || tok.Successor == "" {
				t.Errorf("entry fails the strict banned_tokens schema: %+v", tok)
			}
		}
	}
	if !found {
		t.Error("the added entry is absent from the loaded banned_tokens family")
	}
}

// TestAddPublicEntryGatesUserFacingContent pins AC1 end to end: an entry the verb
// wrote is enforced by the same docs-lint family the hand-curated harness entries
// live in — blocker on a plain mention, silent on the escape line.
func TestAddPublicEntryGatesUserFacingContent(t *testing.T) {
	root, _ := realConfig(t)
	if _, err := AddPublic(AddPublicRequest{RepoRoot: root, Key: "widgetworks", Pattern: `(?i)\bwidgetworks\b`}); err != nil {
		t.Fatalf("AddPublic: %v", err)
	}
	cfg, err := lint.LoadConfig(filepath.Join(root, filepath.FromSlash(PublicConfigRelPath)))
	if err != nil {
		t.Fatal(err)
	}

	docs := t.TempDir()
	write := func(rel, body string) {
		p := filepath.Join(docs, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write("docs/named.md", "# t\n\nBuilt with widgetworks.\n")
	write("docs/allowed.md", "# t\n\n<!-- docs-lint: allow --> widgetworks is named deliberately.\n")
	write("docs/clean.md", "# t\n\nBuilt with a generic term.\n")

	findings, err := lint.Lint(cfg, docs)
	if err != nil {
		t.Fatal(err)
	}
	var hits []lint.Finding
	for _, f := range findings {
		if f.RuleID == "names/widgetworks" {
			hits = append(hits, f)
		}
	}
	if len(hits) != 1 {
		t.Fatalf("names/widgetworks findings = %+v, want exactly one (named.md only)", hits)
	}
	if hits[0].File != filepath.Join("docs", "named.md") || hits[0].Line != 3 || hits[0].Severity != "blocker" {
		t.Errorf("finding = %+v, want a blocker on docs/named.md:3", hits[0])
	}
}

// TestRemovePublicRestoresTheOriginalBytes is the strongest statement the
// surgical editor can make: add then remove is a round trip to the exact bytes
// that were there before.
func TestRemovePublicRestoresTheOriginalBytes(t *testing.T) {
	root, original := realConfig(t)
	if _, err := AddPublic(AddPublicRequest{RepoRoot: root, Key: "widgetworks", Pattern: `(?i)\bwidgetworks\b`}); err != nil {
		t.Fatalf("AddPublic: %v", err)
	}
	if _, err := RemovePublic(root, "widgetworks"); err != nil {
		t.Fatalf("RemovePublic: %v", err)
	}
	if got := readConfig(t, root); string(got) != string(original) {
		t.Errorf("round trip is not byte-identical:\n--- got\n%s\n--- want\n%s", got, original)
	}
}

// TestRemovePublicRefusesEntriesItDoesNotOwn keeps the hand-curated families
// (harness/*, present_tense/*) out of the verb's reach: the id namespace is the
// ownership boundary, and a refusal writes nothing.
func TestRemovePublicRefusesEntriesItDoesNotOwn(t *testing.T) {
	root, original := realConfig(t)

	if _, err := RemovePublic(root, "harness/gemini"); !errors.Is(err, ErrNotManaged) {
		t.Errorf("removing harness/gemini: err = %v, want ErrNotManaged", err)
	}
	if _, err := RemovePublic(root, "nosuchkey"); !errors.Is(err, ErrUnknownKey) {
		t.Errorf("removing an unknown key: err = %v, want ErrUnknownKey", err)
	}
	if got := readConfig(t, root); string(got) != string(original) {
		t.Error("a refused removal wrote to the config")
	}
}

// TestListPublicRendersTheWholeFamilyInFull pins AC6's public half: public entries
// render in full (pattern included — they are committed and reviewable), with the
// managed flag marking which the verb owns.
func TestListPublicRendersTheWholeFamilyInFull(t *testing.T) {
	root, _ := realConfig(t)
	if _, err := AddPublic(AddPublicRequest{RepoRoot: root, Key: "widgetworks", Pattern: `(?i)\bwidgetworks\b`}); err != nil {
		t.Fatalf("AddPublic: %v", err)
	}
	rep, err := ListPublic(root)
	if err != nil {
		t.Fatalf("ListPublic: %v", err)
	}
	if !rep.Present {
		t.Fatal("Present = false for a config that exists")
	}
	byID := map[string]PublicEntry{}
	for _, e := range rep.Entries {
		byID[e.ID] = e
	}
	managed, ok := byID["names/widgetworks"]
	if !ok {
		t.Fatalf("the added entry is missing from the report: %+v", rep.Entries)
	}
	if !managed.Managed || managed.Key != "widgetworks" || managed.Pattern != `(?i)\bwidgetworks\b` {
		t.Errorf("managed entry = %+v", managed)
	}
	harness, ok := byID["harness/gemini"]
	if !ok {
		t.Fatalf("the hand-curated harness family is missing from the report: %+v", rep.Entries)
	}
	if harness.Managed {
		t.Error("harness/gemini reported as verb-managed")
	}
}

// TestAddPublicRefusals covers the input contract, including the duplicate that
// would otherwise land two entries under one id.
func TestAddPublicRefusals(t *testing.T) {
	root, original := realConfig(t)
	if _, err := AddPublic(AddPublicRequest{RepoRoot: root, Key: "widgetworks", Pattern: `(?i)\bwidgetworks\b`}); err != nil {
		t.Fatalf("AddPublic: %v", err)
	}
	after := readConfig(t, root)

	cases := []struct {
		name string
		req  AddPublicRequest
		want error
	}{
		{"duplicate key", AddPublicRequest{RepoRoot: root, Key: "widgetworks", Pattern: `x`}, ErrDuplicateKey},
		{"key with a metacharacter", AddPublicRequest{RepoRoot: root, Key: "widget*", Pattern: `x`}, ErrInvalidKey},
		{"empty pattern", AddPublicRequest{RepoRoot: root, Key: "empty", Pattern: ""}, ErrInvalidPattern},
		{"uncompilable pattern", AddPublicRequest{RepoRoot: root, Key: "broken", Pattern: `[unclosed`}, ErrInvalidPattern},
		{"unknown severity", AddPublicRequest{RepoRoot: root, Key: "loud", Pattern: `x`, Severity: "shout"}, ErrInvalidSeverity},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := AddPublic(tc.req); !errors.Is(err, tc.want) {
				t.Fatalf("err = %v, want %v", err, tc.want)
			}
			if got := readConfig(t, root); string(got) != string(after) {
				t.Error("a refused add wrote to the config")
			}
		})
	}
	if _, err := AddPublic(AddPublicRequest{RepoRoot: t.TempDir(), Key: "k", Pattern: "x"}); !errors.Is(err, ErrNoStore) {
		t.Errorf("adding without a docs-lint config: err = %v, want ErrNoStore", err)
	}
	_ = original
}

// TestPublicSurgerySurvivesFormatting exercises the array-boundary cases a real
// config does not cover: a single-entry family and a pretty-printed one.
func TestPublicSurgerySurvivesFormatting(t *testing.T) {
	const single = `{
  "roots": ["docs"],
  "banned_tokens": [
    {"id":"names/only","pattern":"only","severity":"blocker","successor":"s","allow_context":["(?i)<!--\\s*docs-lint:\\s*allow\\b"],"message":"m"}
  ],
  "rules": {"links_resolve": {"enabled": true, "severity": "blocker"}}
}
`
	const pretty = `{
  "banned_tokens": [
    {
      "id": "names/first",
      "pattern": "first",
      "severity": "blocker",
      "successor": "s",
      "allow_context": ["(?i)<!--\\s*docs-lint:\\s*allow\\b"],
      "message": "m"
    }
  ]
}
`
	for _, tc := range []struct{ name, body string }{{"single entry", single}, {"pretty printed", pretty}} {
		t.Run(tc.name, func(t *testing.T) {
			root := t.TempDir()
			path := filepath.Join(root, filepath.FromSlash(PublicConfigRelPath))
			if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(path, []byte(tc.body), 0o644); err != nil {
				t.Fatal(err)
			}
			if _, err := AddPublic(AddPublicRequest{RepoRoot: root, Key: "second", Pattern: "second"}); err != nil {
				t.Fatalf("AddPublic: %v", err)
			}
			if _, err := lint.LoadConfig(path); err != nil {
				t.Fatalf("edited config does not load: %v", err)
			}
			if _, err := RemovePublic(root, "second"); err != nil {
				t.Fatalf("RemovePublic: %v", err)
			}
			if got := readConfig(t, root); string(got) != tc.body {
				t.Errorf("round trip is not byte-identical:\n--- got\n%s\n--- want\n%s", got, tc.body)
			}
			// Removing the last entry leaves a valid, loadable config.
			if _, err := RemovePublic(root, firstKey(t, root)); err != nil {
				t.Fatalf("RemovePublic (last entry): %v", err)
			}
			if _, err := lint.LoadConfig(path); err != nil {
				t.Fatalf("config with an emptied family does not load: %v", err)
			}
		})
	}
}

func firstKey(t *testing.T, root string) string {
	t.Helper()
	rep, err := ListPublic(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(rep.Entries) == 0 {
		t.Fatal("no entries to remove")
	}
	return rep.Entries[0].Key
}
