package ahoy

import (
	"os"
	"path/filepath"
	"testing"
)

// TestExplicitBinTargetSuppressesAdoption pins the sandbox escape recorded as
// iss-2608261447262355.
//
// effectiveBinTarget answers one question — which PATH entry does a verb act on
// — and it used to answer it by preferring any owned entry it could find, only
// falling back to binTarget(). In production that is right: ABCD_BIN_TARGET is
// unset, an install that finds an abcd it already owns should UPDATE that one
// rather than leave a second copy somewhere else, and adoption is how `ahoy
// install` stays idempotent across the shapes a real machine presents.
//
// Under test it is destructive. A test that sets ABCD_BIN_TARGET to a temp path
// is asking for a sandbox, and the preference silently overrode the ask: the
// developer's own installed abcd is an owned entry, so `ahoy install` reached
// from a test adopted the REAL binary and rewrote it to point inside a temp dir
// that the test then deleted. The machines that dogfood the installer are
// exactly the machines that have an owned entry to find, so the blast radius
// tracked the population most likely to run the suite.
//
// So an explicitly set ABCD_BIN_TARGET now wins. It is an instruction, not a
// hint, and honouring it is inert in production where the variable is unset.
func TestExplicitBinTargetSuppressesAdoption(t *testing.T) {
	pluginRoot := t.TempDir()
	if err := os.WriteFile(filepath.Join(pluginRoot, binName), []byte("#!/bin/sh\n"), 0o755); err != nil {
		t.Fatal(err)
	}

	// An OWNED entry on PATH: a symlink named abcd pointing into the plugin root.
	// This is the developer's installed abcd, standing in for the thing the old
	// behaviour adopted and overwrote.
	pathDir := t.TempDir()
	realEntry := filepath.Join(pathDir, binName)
	if err := os.Symlink(filepath.Join(pluginRoot, binName), realEntry); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", pathDir)

	// The sandbox the caller asked for.
	sandbox := filepath.Join(t.TempDir(), "bin", binName)
	t.Setenv("ABCD_BIN_TARGET", sandbox)

	if got := effectiveBinTarget(pluginRoot); got != sandbox {
		t.Errorf("effectiveBinTarget = %q, want the explicit ABCD_BIN_TARGET %q\n"+
			"an explicit target is a sandbox instruction; adopting %q instead lets a test write outside its sandbox",
			got, sandbox, realEntry)
	}
}

// TestAdoptionStillWinsWhenBinTargetIsUnset is the other half, and the reason
// the fix is scoped to an EXPLICIT setting rather than removing adoption. With
// ABCD_BIN_TARGET unset — production — an owned PATH entry must still be the
// entry every verb acts on, or `ahoy install` stops being idempotent and starts
// leaving a second copy beside the one it already owns.
func TestAdoptionStillWinsWhenBinTargetIsUnset(t *testing.T) {
	pluginRoot := t.TempDir()
	if err := os.WriteFile(filepath.Join(pluginRoot, binName), []byte("#!/bin/sh\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	pathDir := t.TempDir()
	owned := filepath.Join(pathDir, binName)
	if err := os.Symlink(filepath.Join(pluginRoot, binName), owned); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", pathDir)
	t.Setenv("ABCD_BIN_TARGET", "")

	if got := effectiveBinTarget(pluginRoot); got != owned {
		t.Errorf("effectiveBinTarget = %q, want the owned PATH entry %q — adoption must survive for the unset case", got, owned)
	}
}
