package lifeboat

import (
	"os"
	"path/filepath"
	"testing"
)

// TestIsAbcdLifeboatRefusesSymlinkedProvenance is the attack-input test for the
// guarded destination-gate read: a pack destination is attacker-influenced, so a
// symlinked _provenance.json must not be followed. Without O_NOFOLLOW the old
// os.ReadFile followed the link to a real provenance file and wrongly classified
// the directory as an abcd lifeboat (green-lighting an overwrite).
func TestIsAbcdLifeboatRefusesSymlinkedProvenance(t *testing.T) {
	base := t.TempDir()
	// A valid provenance file the symlink points at.
	// A provenance the symlink points at that WOULD satisfy isAbcdLifeboat (schema
	// >= 1 AND a non-empty manifest hash) if it were followed — so the test fails if
	// the read follows the link.
	realProv := filepath.Join(base, "real_provenance.json")
	if err := os.WriteFile(realProv, []byte(`{"schema_version":1,"manifest_sha256":"`+
		"0000000000000000000000000000000000000000000000000000000000000000"+`"}`), 0o644); err != nil {
		t.Fatal(err)
	}
	dir := filepath.Join(base, "dest")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(realProv, filepath.Join(dir, ProvenanceName)); err != nil {
		t.Skipf("symlink unsupported: %v", err)
	}
	if isAbcdLifeboat(dir) {
		t.Error("isAbcdLifeboat followed a symlinked _provenance.json; a destination symlink must be refused")
	}
}
