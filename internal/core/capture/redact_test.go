package capture

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestCaptureRedactsHomePathOnWrite is the regression for
// iss-2608231025198888: `abcd capture` wrote free text straight to the
// committed ledger with no detector between, so an absolute home path reached
// the repository with every lint gate green. The committed file must never
// carry the caller's own home root.
func TestCaptureRedactsHomePathOnWrite(t *testing.T) {
	repo := t.TempDir()
	home := t.TempDir()
	t.Setenv("HOME", home)

	res, err := Capture(CaptureRequest{
		RepoRoot:    repo,
		Text:        "the PATH entry is " + filepath.Join(home, ".local", "bin", "abcd") + " and it moved",
		Severity:    SeverityMinor,
		Category:    "process",
		Source:      "user-observation",
		Slug:        "home-path-in-body",
		FoundDuring: "unit-test",
	})
	if err != nil {
		t.Fatalf("Capture: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(repo, res.Path))
	if err != nil {
		// Path is repo-relative on the result; fall back to an absolute join.
		t.Fatalf("read committed issue: %v", err)
	}
	if strings.Contains(string(data), home) {
		t.Fatalf("committed issue carries the caller's home root %q:\n%s", home, data)
	}
	if res.Redacted == 0 {
		t.Fatalf("Capture reported no redaction, want a non-zero count so the surface can say so")
	}
}

// TestCaptureLeavesCleanTextAlone guards the other direction: a body with
// nothing to redact must be written byte-for-byte and report zero, so the
// notice never fires spuriously.
func TestCaptureLeavesCleanTextAlone(t *testing.T) {
	repo := t.TempDir()
	t.Setenv("HOME", t.TempDir())

	body := "the enumeration is four items long and misses the artefact class"
	res, err := Capture(CaptureRequest{
		RepoRoot:    repo,
		Text:        body,
		Severity:    SeverityMinor,
		Category:    "process",
		Source:      "user-observation",
		Slug:        "nothing-to-redact",
		FoundDuring: "unit-test",
	})
	if err != nil {
		t.Fatalf("Capture: %v", err)
	}
	if res.Redacted != 0 {
		t.Fatalf("Redacted = %d, want 0 for a clean body", res.Redacted)
	}
	data, err := os.ReadFile(filepath.Join(repo, res.Path))
	if err != nil {
		t.Fatalf("read committed issue: %v", err)
	}
	if !strings.Contains(string(data), body) {
		t.Fatalf("clean body was altered:\n%s", data)
	}
}
