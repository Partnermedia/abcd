package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

// TestReferenceMatchesCommittedPage is the drift gate for the generated CLI
// reference. It regenerates the reference from the live command tree and diffs it
// against the committed docs/reference/cli/commands.md. Any divergence — a new
// verb, a changed flag, an edited summary, or a hand-edit of the page — fails the
// build, so the reference can never silently go stale. Regenerate with
// `go generate ./internal/surface/cli` (or `go run ./cmd/abcd-gen-cli-ref`).
func TestReferenceMatchesCommittedPage(t *testing.T) {
	// The package dir is internal/surface/cli; the page lives at the repo root.
	page := filepath.Join("..", "..", "..", filepath.FromSlash(ReferencePagePath))
	committed, err := os.ReadFile(page)
	if err != nil {
		t.Fatalf("cannot read committed reference page %s: %v\n"+
			"regenerate it with `go generate ./internal/surface/cli`", ReferencePagePath, err)
	}

	want := GenerateReference()
	if string(committed) != want {
		t.Fatalf("%s is stale: the committed reference no longer matches the command tree.\n"+
			"Regenerate it with `go generate ./internal/surface/cli` and commit the result.\n"+
			"first difference at %s", ReferencePagePath, firstDiff(string(committed), want))
	}
}

// firstDiff returns a short human-readable locator (line:col) of the first byte
// at which got and want differ, so a failing drift test points at the divergence
// instead of dumping two large documents.
func firstDiff(got, want string) string {
	line, col := 1, 1
	for i := 0; i < len(got) && i < len(want); i++ {
		if got[i] != want[i] {
			return fmt.Sprintf("line %d col %d (committed=%q generated=%q)",
				line, col, string(got[i]), string(want[i]))
		}
		if got[i] == '\n' {
			line++
			col = 1
		} else {
			col++
		}
	}
	// One is a prefix of the other: the difference is trailing content.
	return fmt.Sprintf("line %d col %d (length differs — one file is a prefix of the other)", line, col)
}
