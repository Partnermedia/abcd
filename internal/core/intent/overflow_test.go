package intent

import (
	"os"
	"path/filepath"
	"strconv"
	"testing"
)

// TestCreateFromTextRefusesAtIntegerCeiling proves the itd-N mint refuses cleanly
// when the observed max is at the integer ceiling, instead of wrapping max+1 to a
// negative id and persisting a malformed draft (a scanned ref or a local file
// carrying itd-<MaxInt>-x.md is the attack path). It must return a loud error and
// write NO new draft.
func TestCreateFromTextRefusesAtIntegerCeiling(t *testing.T) {
	root := t.TempDir()
	maxID := "itd-" + strconv.Itoa(int(^uint(0)>>1)) // itd-<math.MaxInt>
	draftsDir := filepath.Join(root, IntentsRelDir, BucketDrafts)
	if err := os.MkdirAll(draftsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	ceiling := filepath.Join(draftsDir, maxID+"-huge.md")
	if err := os.WriteFile(ceiling, []byte("---\nid: "+maxID+"\n---\n# huge\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	_, _, err := CreateFromText(root, "a fresh intent worth shipping", "")
	if err == nil {
		t.Fatal("CreateFromText must refuse at the integer ceiling, not mint a wrapped negative id")
	}

	// No new draft may have been written — only the ceiling file remains.
	entries, rerr := os.ReadDir(draftsDir)
	if rerr != nil {
		t.Fatal(rerr)
	}
	if len(entries) != 1 || entries[0].Name() != filepath.Base(ceiling) {
		var names []string
		for _, e := range entries {
			names = append(names, e.Name())
		}
		t.Fatalf("a draft was written despite the ceiling refusal: %v", names)
	}
}
