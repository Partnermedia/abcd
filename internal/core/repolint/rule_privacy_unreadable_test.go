package repolint_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// A tracked file the scan cannot OPEN is content nobody looked at, so it must
// warn "not scanned" — never silently count as conforming (the engine
// contract, and the same fix the oversize arm got as iss-356 item 4). Runs
// only where permissions bind: root bypasses mode bits, so under euid 0 (a
// container) the EACCES this stages cannot occur.
func TestRule_PrivacyWarnsOnUnreadableTrackedFile(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("permission bits do not bind under euid 0; the EACCES cannot be staged")
	}
	b := newFixtureRepo(t).conforming()
	locked := filepath.Join(b.root, "locked.md")
	if err := os.WriteFile(locked, []byte("ordinary content\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	b.commit()
	if err := os.Chmod(locked, 0o000); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(locked, 0o644) })

	res := b.run()
	f := findingFor(res, "privacy-hygiene")
	if f == nil {
		t.Fatal("an unreadable tracked file was silently reported as conforming; want a not-scanned warn")
	}
	if f.File != "locked.md" || !strings.Contains(f.Message, "not scanned") {
		t.Fatalf("want a not-scanned warn for locked.md, got: %+v", f)
	}
}
