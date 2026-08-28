package intent

import (
	"strings"
	"testing"
)

// TestAppendToAuditNotesInsertsAboveTrailingLinkRefs pins the ordering fix for
// iss-2608210737265820: when the Audit Notes section ends with markdown
// link-reference definitions (e.g. itd-114's `[iss-80]:` ref parked at EOF), a new
// review block must land ABOVE them, inside the section it belongs to, not below
// them where it reads as detached.
func TestAppendToAuditNotesInsertsAboveTrailingLinkRefs(t *testing.T) {
	content := strings.Join([]string{
		"---",
		"id: itd-114",
		"---",
		"# alpha",
		"",
		"## Audit Notes",
		"",
		"An earlier note about the review.",
		"",
		"[iss-80]: https://example.com/issues/iss-80",
		"[spc-28]: https://example.com/specs/spc-28",
		"",
	}, "\n")
	block := "<!-- abcd-review: INGESTED receipt=rcp-abc123def456 -->\nFidelity review — receipt rcp-abc123def456."

	out := appendToAuditNotes(content, block)
	lines := strings.Split(out, "\n")

	blockIdx, refIdx := -1, -1
	for i, ln := range lines {
		if strings.Contains(ln, "abcd-review: INGESTED") {
			blockIdx = i
		}
		if strings.HasPrefix(ln, "[iss-80]:") {
			refIdx = i
		}
	}
	if blockIdx < 0 {
		t.Fatalf("review block not found in output:\n%s", out)
	}
	if refIdx < 0 {
		t.Fatalf("trailing link ref not preserved in output:\n%s", out)
	}
	if blockIdx > refIdx {
		t.Fatalf("review block (line %d) landed BELOW the trailing link ref (line %d); it must sit above:\n%s", blockIdx, refIdx, out)
	}
	// The prose that preceded the refs must stay above the block too.
	proseIdx := -1
	for i, ln := range lines {
		if strings.HasPrefix(ln, "An earlier note") {
			proseIdx = i
		}
	}
	if proseIdx < 0 || proseIdx > blockIdx {
		t.Fatalf("existing prose must remain above the new block (prose=%d block=%d):\n%s", proseIdx, blockIdx, out)
	}
	// Both link refs must survive exactly once.
	if strings.Count(out, "[spc-28]: https://example.com/specs/spc-28") != 1 {
		t.Fatalf("second link ref not preserved exactly once:\n%s", out)
	}
}
