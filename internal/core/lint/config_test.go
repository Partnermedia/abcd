package lint

import (
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"
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

// TestLoadConfigRefusesFIFO pins that a FIFO in the config's place returns
// immediately instead of blocking the open forever. The docs-lint/record-lint
// config is a committed, cross-repo-clonable trust boundary reachable through
// the session hooks; before the guarded read a planted FIFO wedged every verb
// that loads it (abcd docs lint / lint / ahoy / hook session-start / cite
// refresh) at exit 124.
func TestLoadConfigRefusesFIFO(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "docs-lint.json")
	if err := syscall.Mkfifo(path, 0o644); err != nil {
		t.Skipf("mkfifo unsupported: %v", err)
	}
	done := make(chan error, 1)
	go func() {
		_, err := LoadConfig(path)
		done <- err
	}()
	select {
	case err := <-done:
		if err == nil {
			t.Fatal("LoadConfig accepted a FIFO config")
		}
	case <-time.After(3 * time.Second):
		t.Fatal("LoadConfig blocked on a FIFO config (pre-fix behaviour)")
	}
}

// TestLoadConfigRefusesSymlinkedLeaf pins that a symlinked config leaf is refused
// rather than followed. Before the fix a committed .abcd/docs-lint.json symlink
// pointing outside the repository was followed, so the linter ran its whole
// ruleset from a file the repository does not own and git does not track.
func TestLoadConfigRefusesSymlinkedLeaf(t *testing.T) {
	dir := t.TempDir()
	outside := filepath.Join(dir, "outside.json")
	if err := os.WriteFile(outside, []byte(`{"roots":["elsewhere"]}`), 0o644); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, "docs-lint.json")
	if err := os.Symlink(outside, path); err != nil {
		t.Skipf("symlink unsupported: %v", err)
	}
	cfg, err := LoadConfig(path)
	if err == nil {
		t.Fatalf("LoadConfig followed a symlinked config leaf and returned roots %v", cfg.Roots)
	}
}

// TestLoadConfigRefusesSymlinkedDir pins that a symlinked config DIRECTORY is
// refused before the leaf is touched, so a swapped .abcd cannot redirect the
// read at a config the repository does not own.
func TestLoadConfigRefusesSymlinkedDir(t *testing.T) {
	dir := t.TempDir()
	realDir := filepath.Join(dir, "real")
	if err := os.Mkdir(realDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(realDir, "docs-lint.json"), []byte(`{"roots":["elsewhere"]}`), 0o644); err != nil {
		t.Fatal(err)
	}
	linkDir := filepath.Join(dir, ".abcd")
	if err := os.Symlink(realDir, linkDir); err != nil {
		t.Skipf("symlink unsupported: %v", err)
	}
	if _, err := LoadConfig(filepath.Join(linkDir, "docs-lint.json")); err == nil {
		t.Fatal("LoadConfig followed a symlinked config directory")
	}
}

// TestLoadConfigRefusesOversize pins that a config over the byte cap is refused
// before it is read, closing the /dev/zero-symlink OOM.
func TestLoadConfigRefusesOversize(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "docs-lint.json")
	// Valid JSON, padded past the cap, so only the size guard can refuse it —
	// a malformed payload would fail json.Unmarshal even pre-fix (a vacuous
	// test). The padding lives in an exempt_paths entry the schema tolerates.
	pad := strings.Repeat("x", maxLintConfigBytes)
	body := `{"roots":["rec"],"exempt_paths":["` + pad + `"]}`
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadConfig(path); err == nil {
		t.Fatal("LoadConfig accepted an over-cap config")
	}
}
