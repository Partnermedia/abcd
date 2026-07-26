package spec

import (
	"os"
	"path/filepath"
	"strconv"
	"testing"
)

// TestCreateRefusesAtIntegerCeiling proves the spc-N mint refuses cleanly when
// the observed max is at the integer ceiling, instead of wrapping max+1 to a
// negative id and persisting a malformed record (a scanned ref or a local file
// carrying spc-<MaxInt>-x.md is the attack path). It must return a loud error and
// write NO new spec file.
func TestCreateRefusesAtIntegerCeiling(t *testing.T) {
	root := t.TempDir()
	maxID := "spc-" + strconv.Itoa(int(^uint(0)>>1)) // spc-<math.MaxInt>
	openDir := filepath.Join(root, SpecsRelDir, StatusOpen)
	if err := os.MkdirAll(openDir, 0o755); err != nil {
		t.Fatal(err)
	}
	ceiling := filepath.Join(openDir, maxID+"-huge.md")
	if err := os.WriteFile(ceiling, []byte("---\nid: "+maxID+"\nslug: huge\nintent: itd-1\n---\n# huge\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	_, _, err := Create(root, "itd-2", "new-one")
	if err == nil {
		t.Fatal("Create must refuse at the integer ceiling, not mint a wrapped negative id")
	}

	// No new spec file may have been written — only the ceiling file remains.
	entries, rerr := os.ReadDir(openDir)
	if rerr != nil {
		t.Fatal(rerr)
	}
	if len(entries) != 1 || entries[0].Name() != filepath.Base(ceiling) {
		var names []string
		for _, e := range entries {
			names = append(names, e.Name())
		}
		t.Fatalf("a spec file was written despite the ceiling refusal: %v", names)
	}
}
