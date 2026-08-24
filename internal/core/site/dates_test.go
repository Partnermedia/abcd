package site

import (
	"testing"

	"github.com/intentdriven/abcd/internal/gittest"
)

// TestLoadHistoryDatesNonASCIIPath proves the git history walk reads paths
// verbatim, not git's default C-quoted form for non-ASCII bytes: a record whose
// path carries a non-ASCII character must still be placed in time, not seated in
// the undated bucket because its quoted key never matched the filesystem path.
func TestLoadHistoryDatesNonASCIIPath(t *testing.T) {
	r := gittest.NewRepo(t)
	const rel = ".abcd/development/principles/café-rule.md"
	r.Write(rel, "# Café rule\n")
	r.Commit("add café rule")

	h, err := LoadHistory(r.Root())
	if err != nil {
		t.Fatalf("LoadHistory: %v", err)
	}
	d, ok := h.Files[rel]
	if !ok {
		t.Fatalf("history has no entry for %q (a quoted key would not match)", rel)
	}
	if d.Created == "" || d.Touched == "" {
		t.Errorf("non-ASCII path left undated: %+v", d)
	}
}
