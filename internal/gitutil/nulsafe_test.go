package gitutil_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/intentdriven/abcd/internal/gitutil"
)

// TestRunCappedPreservesLeadingWhitespaceInNulLists pins the byte fidelity a
// NUL-separated listing depends on. The -z form exists so whitespace in
// filenames cannot desync the list, but a whole-buffer TrimSpace strips
// leading whitespace off the FIRST entry — so a repo whose first-sorting path
// begins with a space loses that one file from any set keyed on the returned
// names (the lifeboat not-ignored set silently classified it ignored).
// Trailing NULs are not IsSpace, so only the leading side was damaged.
func TestRunCappedPreservesLeadingWhitespaceInNulLists(t *testing.T) {
	repo := newRepo(t, "")
	// A leading-space name sorts before every printable-range name, so it is
	// the first ls-files entry — the position the trim corrupted.
	leading := " lead.txt"
	if err := os.WriteFile(filepath.Join(repo, leading), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(repo, "z.txt"), []byte("y"), 0o644); err != nil {
		t.Fatal(err)
	}
	commitAll(t, repo)

	out, err := gitutil.RunCapped(repo, 1<<20, "ls-files", "--cached", "-z")
	if err != nil {
		t.Fatal(err)
	}
	set := map[string]bool{}
	for _, f := range strings.Split(out, "\x00") {
		if f != "" {
			set[f] = true
		}
	}
	if !set[leading] {
		t.Fatalf("the first NUL-list entry lost its leading space: %q", out)
	}
	if !set["z.txt"] {
		t.Fatalf("z.txt missing from the listing: %q", out)
	}
}
