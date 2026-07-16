package ahoy

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

// TestEnsureMarkerDryRunPredictsWithoutWriting covers the probe path: dryRun
// classifies every state and writes nothing.
func TestEnsureMarkerDryRunPredictsWithoutWriting(t *testing.T) {
	t.Run("absent -> would change", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "CLAUDE.md")
		changed, err := EnsureMarker(path, true)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !changed {
			t.Errorf("absent file: changed=false, want true (an install would happen)")
		}
		if _, statErr := os.Stat(path); !os.IsNotExist(statErr) {
			t.Errorf("dryRun created the file; it must write nothing")
		}
	})

	t.Run("current -> no change", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "CLAUDE.md")
		if _, err := EnsureMarker(path, false); err != nil {
			t.Fatalf("seed install: %v", err)
		}
		before, _ := os.ReadFile(path)
		changed, err := EnsureMarker(path, true)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if changed {
			t.Errorf("current block: changed=true, want false")
		}
		after, _ := os.ReadFile(path)
		if !bytes.Equal(before, after) {
			t.Errorf("dryRun mutated a current file")
		}
	})

	t.Run("outdated -> would change", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "CLAUDE.md")
		// A stale, non-canonical block.
		stale := "# Title\n\n<!-- BEGIN ABCD -->\nold loader text\n<!-- END ABCD -->\n"
		if err := os.WriteFile(path, []byte(stale), 0o644); err != nil {
			t.Fatal(err)
		}
		changed, err := EnsureMarker(path, true)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !changed {
			t.Errorf("outdated block: changed=false, want true (a refresh would happen)")
		}
		got, _ := os.ReadFile(path)
		if string(got) != stale {
			t.Errorf("dryRun rewrote an outdated file; it must write nothing")
		}
	})

	t.Run("symlink -> error, no change", func(t *testing.T) {
		dir := t.TempDir()
		real := filepath.Join(dir, "real.md")
		if err := os.WriteFile(real, []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
		link := filepath.Join(dir, "CLAUDE.md")
		if err := os.Symlink(real, link); err != nil {
			t.Skipf("cannot symlink: %v", err)
		}
		changed, err := EnsureMarker(link, true)
		if err == nil {
			t.Errorf("symlinked target: err=nil, want a refusal")
		}
		if changed {
			t.Errorf("symlinked target: changed=true, want false")
		}
	})
}

// TestEnsureMarkerWriteInstallsAndIsIdempotent covers the embark write path.
func TestEnsureMarkerWriteInstallsAndIsIdempotent(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "CLAUDE.md")

	changed, err := EnsureMarker(path, false)
	if err != nil {
		t.Fatalf("install: %v", err)
	}
	if !changed {
		t.Errorf("first install: changed=false, want true")
	}
	if classifyMarker(path) != markerCurrent {
		t.Errorf("after install state = %q, want current", classifyMarker(path))
	}

	// Idempotent: a second write leaves it current and unchanged.
	before, _ := os.ReadFile(path)
	changed, err = EnsureMarker(path, false)
	if err != nil {
		t.Fatalf("re-run: %v", err)
	}
	if changed {
		t.Errorf("idempotent re-run: changed=true, want false")
	}
	after, _ := os.ReadFile(path)
	if !bytes.Equal(before, after) {
		t.Errorf("idempotent re-run mutated the file")
	}
}

// TestEnsureMarkerWriteRefusesSymlinkedLeaf: a planted symlink cannot redirect
// the write.
func TestEnsureMarkerWriteRefusesSymlinkedLeaf(t *testing.T) {
	dir := t.TempDir()
	real := filepath.Join(dir, "real.md")
	if err := os.WriteFile(real, []byte("original\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(dir, "CLAUDE.md")
	if err := os.Symlink(real, link); err != nil {
		t.Skipf("cannot symlink: %v", err)
	}
	changed, err := EnsureMarker(link, false)
	if err == nil {
		t.Errorf("symlinked leaf: err=nil, want a refusal")
	}
	if changed {
		t.Errorf("symlinked leaf: changed=true, want false")
	}
	// The symlink target must not have been written through.
	got, _ := os.ReadFile(real)
	if string(got) != "original\n" {
		t.Errorf("write leaked through the symlink to its target: %q", got)
	}
}
