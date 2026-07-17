package memory

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestLoadRegistryRefusesSymlink is the attack-input test for the guarded read:
// the sources index sits in the repo working tree (a trust boundary), so a
// committed symlink (e.g. to /dev/zero) must be refused with a typed
// RegistryFormatError, not followed. A missing file still yields an empty
// registry, and a real JSON object still loads.
func TestLoadRegistryRefusesSymlink(t *testing.T) {
	dir := t.TempDir()

	// Missing -> empty registry, no error.
	if reg, err := LoadRegistry(filepath.Join(dir, "absent.json")); err != nil || len(reg) != 0 {
		t.Fatalf("missing registry: reg=%v err=%v, want empty+nil", reg, err)
	}

	// A symlinked registry must be refused (O_NOFOLLOW -> ErrNotRegular ->
	// RegistryFormatError), never followed.
	target := filepath.Join(dir, "target.json")
	if err := os.WriteFile(target, []byte(`{"ok":1}`), 0o644); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(dir, "index.json")
	if err := os.Symlink(target, link); err != nil {
		t.Skipf("symlink unsupported: %v", err)
	}
	// The security property: the symlink is REFUSED (O_NOFOLLOW), never followed —
	// no content is returned. The refusal is classified like every other guarded
	// refusal (ErrNotRegular/ErrTooBig): a typed RegistryFormatError, so a caller
	// that branches on the type handles a planted symlink identically to any other
	// malformed index, and the raw ELOOP *os.PathError (with its syscall detail)
	// never escapes.
	reg, err := LoadRegistry(link)
	if err == nil {
		t.Fatalf("LoadRegistry followed a symlinked sources index (reg=%v); a committed symlink must be refused", reg)
	}
	var rfe *RegistryFormatError
	if !errors.As(err, &rfe) {
		t.Errorf("symlinked index returned %T (%v); want a typed *RegistryFormatError like the other guarded refusals", err, err)
	}
	if strings.Contains(err.Error(), "too many levels of symbolic links") {
		t.Errorf("symlink refusal leaked the raw ELOOP syscall detail: %v", err)
	}

	// A real regular-file JSON object still loads.
	real := filepath.Join(dir, "real.json")
	if err := os.WriteFile(real, []byte(`{"k":"v"}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if reg, err := LoadRegistry(real); err != nil || reg["k"] != "v" {
		t.Fatalf("real registry: reg=%v err=%v", reg, err)
	}
}
