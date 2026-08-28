package history

import (
	"bytes"
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/intentdriven/abcd/internal/adapter/gitleaks"
	"github.com/intentdriven/abcd/internal/adapter/scanner"
)

// A bare, unanchored high-entropy value with a key name and delimiter — the
// exact residue iss-96 pins: the native prefix set passes it through, and it is
// what an opted-in gitleaks catches (its generic-api-key rule).
const gitleaksResidueSecret = "Zx8Kq2Lm9Pv4Wr7Tn1Bc5Yd3Hs6Jf0Gu8Ae2Qi4"

// TestCaptureDefaultOffStoresResidueVerbatim proves the default-off path is
// unchanged: with no gitleaks opt-in config in the repo, the residue value the
// native scanner misses is stored verbatim, exactly as before this adapter
// existed. scanGitleaks is NOT overridden here — the real Scan runs, finds no
// config, and returns (nil, nil).
func TestCaptureDefaultOffStoresResidueVerbatim(t *testing.T) {
	repoRoot, _ := setupStore(t)

	transcript := strings.Join([]string{
		"user: set the key",
		"api_key = " + gitleaksResidueSecret,
		"assistant: done",
	}, "\n")

	res, err := Capture(repoRoot, testRootSHA, "sess-defoff", []byte(transcript), "native")
	if err != nil {
		t.Fatalf("Capture: %v", err)
	}
	if !res.Wrote {
		t.Fatal("expected Wrote=true")
	}
	onDisk, err := os.ReadFile(res.Record.Path)
	if err != nil {
		t.Fatalf("read record: %v", err)
	}
	// Default-off: the native scanner does not catch it, and no adapter ran, so
	// the value is present verbatim — the behaviour the iss-96 pin records.
	if !bytes.Contains(onDisk, []byte(gitleaksResidueSecret)) {
		t.Error("default-off path changed behaviour: the residue value was redacted with no opt-in")
	}
}

// TestCaptureFoldsGitleaksFindings proves the augmentation path end to end: with
// the seam injected to stand in for an opted-in gitleaks, the residue value is
// redacted out of the stored record and counted in the audit buckets.
func TestCaptureFoldsGitleaksFindings(t *testing.T) {
	repoRoot, _ := setupStore(t)

	restore := scanGitleaks
	t.Cleanup(func() { scanGitleaks = restore })
	scanGitleaks = func(_, text, logical string) ([]scanner.Finding, error) {
		// Locate the residue value as the real adapter would and emit a finding.
		lines := strings.Split(text, "\n")
		var out []scanner.Finding
		for i, ln := range lines {
			if col := strings.Index(ln, gitleaksResidueSecret); col >= 0 {
				out = append(out, scanner.Finding{
					File:     logical,
					Line:     i + 1,
					Column:   col + 1,
					Kind:     "gitleaks:generic-api-key",
					Severity: scanner.SeverityHardFail,
					Matched:  gitleaksResidueSecret,
					Snippet:  ln,
				})
			}
		}
		return out, nil
	}

	transcript := strings.Join([]string{
		"user: set the key",
		"api_key = " + gitleaksResidueSecret,
		"assistant: done",
	}, "\n")

	res, err := Capture(repoRoot, testRootSHA, "sess-fold", []byte(transcript), "native")
	if err != nil {
		t.Fatalf("Capture: %v", err)
	}
	onDisk, err := os.ReadFile(res.Record.Path)
	if err != nil {
		t.Fatalf("read record: %v", err)
	}
	if bytes.Contains(onDisk, []byte(gitleaksResidueSecret)) {
		t.Errorf("the gitleaks-augmented secret leaked into the stored record:\n%s", onDisk)
	}
	if res.Record.Secrets < 1 {
		t.Errorf("expected the gitleaks finding counted in Secrets, got %d", res.Record.Secrets)
	}
}

// TestCaptureGitleaksLoudStagePropagates proves the loud-stage reaches the
// caller: an opted-in-but-absent binary makes Capture fail closed and write
// nothing, naming the opt-in.
func TestCaptureGitleaksLoudStagePropagates(t *testing.T) {
	repoRoot, _ := setupStore(t)

	restore := scanGitleaks
	t.Cleanup(func() { scanGitleaks = restore })
	scanGitleaks = func(_, _, _ string) ([]scanner.Finding, error) {
		return nil, gitleaks.ErrConfiguredNotFound
	}

	_, err := Capture(repoRoot, testRootSHA, "sess-loud", []byte("user: hi\n"), "native")
	if err == nil {
		t.Fatal("expected Capture to fail closed on an armed-but-absent gitleaks")
	}
	if !errors.Is(err, gitleaks.ErrConfiguredNotFound) {
		t.Fatalf("error is not ErrConfiguredNotFound: %v", err)
	}
	if !strings.Contains(err.Error(), "gitleaks configured but not found") {
		t.Fatalf("error does not name the opt-in: %q", err.Error())
	}
	// Nothing was written.
	recs, lerr := List(testRootSHA)
	if lerr != nil {
		t.Fatalf("List: %v", lerr)
	}
	for _, r := range recs {
		if r.SessionID == "sess-loud" {
			t.Error("a record was written despite the loud-stage failure")
		}
	}
}
