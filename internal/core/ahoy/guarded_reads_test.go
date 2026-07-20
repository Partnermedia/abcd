package ahoy

import (
	"os"
	"path/filepath"
	"syscall"
	"testing"
	"time"
)

// iss-97: ahoy.Detect's marker/config reads used bare os.ReadFile — no size cap
// and, worse, a blocking open, so a FIFO planted at a marker/config path (cwd is
// attacker-influenced for any hook that calls Detect) would hang the read
// forever. The reads route through fsutil.ReadGuarded (O_NONBLOCK|O_NOFOLLOW,
// regular-file check, byte cap), so a FIFO returns promptly as not-regular and an
// oversized file is refused. These tests arm that: pre-fix each hangs on the FIFO
// (the select times out), post-fix each returns within the deadline.

// mkfifoOrSkip plants a FIFO at path, skipping on platforms without mkfifo.
func mkfifoOrSkip(t *testing.T, path string) {
	t.Helper()
	if err := syscall.Mkfifo(path, 0o644); err != nil {
		t.Skipf("mkfifo unsupported: %v", err)
	}
}

// withinDeadline runs fn and fails if it does not return within 3s (a blocking
// open on a FIFO would otherwise hang indefinitely).
func withinDeadline(t *testing.T, what string, fn func()) {
	t.Helper()
	done := make(chan struct{})
	go func() {
		fn()
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatalf("%s hung on a FIFO (guarded open must not block)", what)
	}
}

func TestClassifyMarkerRejectsFifoPromptly(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "CLAUDE.md")
	mkfifoOrSkip(t, p)

	var got markerState
	withinDeadline(t, "classifyMarker", func() { got = classifyMarker(p) })
	if got != markerMissing {
		t.Fatalf("a FIFO marker must classify as missing, got %v", got)
	}
}

func TestReadConfigRejectsFifoPromptly(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, ".abcd"), 0o755); err != nil {
		t.Fatal(err)
	}
	mkfifoOrSkip(t, configPath(dir))

	withinDeadline(t, "readConfig", func() { _, _ = readConfig(dir) })
}
