package lint

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// writeConfig writes a record-lint config JSON to a temp file and returns its path.
func writeConfig(t *testing.T, body string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "record-lint.json")
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

// TestLoadConfigRejectsMissingSuccessor asserts a banned_tokens entry without a
// successor is rejected at load — the machine-readable old->new mapping is
// mandatory, not prose-only (iss-51).
func TestLoadConfigRejectsMissingSuccessor(t *testing.T) {
	path := writeConfig(t, `{
	  "roots": ["rec"],
	  "banned_tokens": [
	    {"id":"t1","pattern":"foo","message":"no foo","severity":"blocker","allow_context":["ok"]}
	  ]
	}`)
	if _, err := LoadConfig(path); err == nil {
		t.Fatal("LoadConfig accepted a banned_tokens entry with no successor; want rejection")
	}
}

// TestLoadConfigRejectsEmptyAllowContext asserts a banned_tokens entry with an
// empty allow_context is rejected at load — every ban must declare where the
// token is legitimately allowed (iss-51).
func TestLoadConfigRejectsEmptyAllowContext(t *testing.T) {
	path := writeConfig(t, `{
	  "roots": ["rec"],
	  "banned_tokens": [
	    {"id":"t1","pattern":"foo","message":"no foo","severity":"blocker","successor":"bar","allow_context":[]}
	  ]
	}`)
	if _, err := LoadConfig(path); err == nil {
		t.Fatal("LoadConfig accepted a banned_tokens entry with empty allow_context; want rejection")
	}
}

// TestLoadConfigAcceptsWellFormedEntry asserts a fully-specified entry (successor
// present, allow_context non-empty) loads without error — the strict schema does
// not reject a valid ban.
func TestLoadConfigAcceptsWellFormedEntry(t *testing.T) {
	path := writeConfig(t, `{
	  "roots": ["rec"],
	  "banned_tokens": [
	    {"id":"t1","pattern":"foo","message":"no foo","severity":"blocker","successor":"bar","allow_context":["ok"]}
	  ]
	}`)
	if _, err := LoadConfig(path); err != nil {
		t.Fatalf("LoadConfig rejected a well-formed entry: %v", err)
	}
}

// TestBannedTokenFindingCitesSuccessor asserts the rendered finding message for a
// banned token includes its declared successor — the finding tells the reader
// what to use instead (iss-51 decision c).
func TestBannedTokenFindingCitesSuccessor(t *testing.T) {
	path := writeConfig(t, `{
	  "roots": ["rec"],
	  "banned_tokens": [
	    {"id":"t1","pattern":"oldpath/thing","message":"oldpath is retired","severity":"blocker","successor":"newpath/thing","allow_context":["historical"]}
	  ]
	}`)
	cfg, err := LoadConfig(path)
	if err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}
	root := t.TempDir()
	writeFile(t, root, "rec/bad.md", "see oldpath/thing here\n")

	fs, err := Lint(cfg, root)
	if err != nil {
		t.Fatal(err)
	}
	var msg string
	for _, f := range fs {
		if f.RuleID == "t1" {
			msg = f.Message
		}
	}
	if msg == "" {
		t.Fatalf("expected a t1 finding: %+v", fs)
	}
	if !strings.Contains(msg, "newpath/thing") {
		t.Errorf("finding message does not cite the successor 'newpath/thing': %q", msg)
	}
}
