package ahoy

import (
	"os"
	"path/filepath"
	"testing"
)

// The update verb keys its dispatch on what actually runs: the first `abcd`
// PATH occupant, classified by the same ownership predicate as detection and
// install (spc-32). These tests pin each dispatch shape the resolver reports.

func TestUpdateTargetPluginRootSymlink(t *testing.T) {
	home, pluginRoot := setupUserScope(t)
	binDir := filepath.Join(home, ".local", "bin")
	t.Setenv("PATH", binDir)
	linkOwned(t, filepath.Join(binDir, "abcd"), pluginRoot)

	tgt := ResolveUpdateTarget()
	if tgt.Kind != UpdateTargetPluginRoot {
		t.Errorf("owned symlink into the plugin root classified %q, want %q", tgt.Kind, UpdateTargetPluginRoot)
	}
	if tgt.Path != filepath.Join(binDir, "abcd") {
		t.Errorf("target path = %q, want the first PATH entry", tgt.Path)
	}
}

func TestUpdateTargetDevShim(t *testing.T) {
	home, pluginRoot := setupUserScope(t)
	binDir := filepath.Join(home, ".local", "bin")
	t.Setenv("PATH", binDir)
	target := filepath.Join(binDir, "abcd")
	if err := os.MkdirAll(binDir, 0o755); err != nil {
		t.Fatal(err)
	}
	content := renderDevShim(pluginRoot, pluginBinaryPath(pluginRoot))
	if err := os.WriteFile(target, []byte(content), 0o755); err != nil {
		t.Fatal(err)
	}

	if tgt := ResolveUpdateTarget(); tgt.Kind != UpdateTargetDevShim {
		t.Errorf("dev shim classified %q, want %q", tgt.Kind, UpdateTargetDevShim)
	}
}

func TestUpdateTargetStrandedIsOwnedDangling(t *testing.T) {
	home, pluginRoot := setupUserScope(t)
	binDir := filepath.Join(home, ".local", "bin")
	t.Setenv("PATH", binDir)
	linkStranded(t, filepath.Join(binDir, "abcd"), pluginRoot)

	if tgt := ResolveUpdateTarget(); tgt.Kind != UpdateTargetDangling {
		t.Errorf("stranded entry classified %q, want %q", tgt.Kind, UpdateTargetDangling)
	}
}

func TestUpdateTargetRegularFileIsFile(t *testing.T) {
	home, _ := setupUserScope(t)
	binDir := filepath.Join(home, ".local", "bin")
	t.Setenv("PATH", binDir)
	if err := os.MkdirAll(binDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(binDir, "abcd"), []byte("\x7fELFfake"), 0o755); err != nil {
		t.Fatal(err)
	}

	tgt := ResolveUpdateTarget()
	if tgt.Kind != UpdateTargetFile {
		t.Errorf("regular file classified %q, want %q", tgt.Kind, UpdateTargetFile)
	}
	if tgt.ResolvedPath == "" {
		t.Errorf("a file target must carry its resolved path for the channel checks")
	}
}

func TestUpdateTargetForeignSymlink(t *testing.T) {
	home, _ := setupUserScope(t)
	binDir := filepath.Join(home, ".local", "bin")
	t.Setenv("PATH", binDir)
	other := filepath.Join(t.TempDir(), "elsewhere", "abcd")
	if err := os.MkdirAll(filepath.Dir(other), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(other, []byte("#!/bin/sh\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(binDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(other, filepath.Join(binDir, "abcd")); err != nil {
		t.Fatal(err)
	}

	if tgt := ResolveUpdateTarget(); tgt.Kind != UpdateTargetForeign {
		t.Errorf("foreign symlink classified %q, want %q", tgt.Kind, UpdateTargetForeign)
	}
}

func TestUpdateTargetAbsent(t *testing.T) {
	setupUserScope(t, filepath.Join(t.TempDir(), "empty"))

	if tgt := ResolveUpdateTarget(); tgt.Kind != UpdateTargetAbsent {
		t.Errorf("empty PATH classified %q, want %q", tgt.Kind, UpdateTargetAbsent)
	}
}

// TestUpdateTargetReportsLaterOwnedEntry: when the first occupant is foreign,
// the refusal must be able to say a working abcd sits shadowed behind it.
func TestUpdateTargetReportsLaterOwnedEntry(t *testing.T) {
	home, pluginRoot := setupUserScope(t)
	first := filepath.Join(t.TempDir(), "first")
	binDir := filepath.Join(home, ".local", "bin")
	t.Setenv("PATH", first+string(os.PathListSeparator)+binDir)
	if err := os.MkdirAll(first, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(first, "abcd"), []byte("\x7fELFfake"), 0o755); err != nil {
		t.Fatal(err)
	}
	linkOwned(t, filepath.Join(binDir, "abcd"), pluginRoot)

	tgt := ResolveUpdateTarget()
	if tgt.Kind != UpdateTargetFile {
		t.Fatalf("first occupant classified %q, want %q", tgt.Kind, UpdateTargetFile)
	}
	if tgt.LaterOwned != filepath.Join(binDir, "abcd") {
		t.Errorf("later owned entry = %q, want the shadowed install", tgt.LaterOwned)
	}
}
