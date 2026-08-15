package cli

import (
	"strings"
	"sync/atomic"
	"testing"

	"github.com/REPPL/abcd-cli/internal/core/vintage"
)

// recordingFetcher counts every LatestTag call, so a test can assert exactly
// which command paths reach the network.
type recordingFetcher struct{ calls *int32 }

func (r recordingFetcher) LatestTag() (string, error) {
	atomic.AddInt32(r.calls, 1)
	return "v99.0.0", nil
}

// TestOnlyVersionCheckTouchesTheNetwork is the zero-network invariant (AC4,
// adr-38 tier 1): every implicit path reaches the disk only, and the network is
// touched exactly once, by the explicit --check.
func TestOnlyVersionCheckTouchesTheNetwork(t *testing.T) {
	var calls int32
	orig := newReleaseFetcher
	newReleaseFetcher = func() vintage.ReleaseFetcher { return recordingFetcher{&calls} }
	t.Cleanup(func() { newReleaseFetcher = orig })

	// Every implicit path: none may fetch.
	runCLI(t, "version")
	runCLI(t, "ahoy")
	if _, err := runCLIStdinErr(t, `{"cwd":"`+t.TempDir()+`"}`, "hook", "session-start"); err != nil {
		t.Fatalf("session-start hook errored: %v", err)
	}
	if got := atomic.LoadInt32(&calls); got != 0 {
		t.Fatalf("an implicit path fetched %d time(s); every path except --check must be disk-only", got)
	}

	// The explicit check fetches exactly once and names its source.
	out := runCLI(t, "version", "--check")
	if got := atomic.LoadInt32(&calls); got != 1 {
		t.Fatalf("version --check fetched %d time(s), want exactly 1", got)
	}
	if !strings.Contains(string(out), checkSource) {
		t.Fatalf("check output did not name its source %q:\n%s", checkSource, out)
	}
	if !strings.Contains(string(out), "v99.0.0") {
		t.Fatalf("check output did not report the latest release:\n%s", out)
	}
}
