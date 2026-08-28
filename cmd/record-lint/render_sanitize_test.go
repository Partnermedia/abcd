package main

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/intentdriven/abcd/internal/core/lint"
)

// render_sanitize_test.go — iss-264. record-lint's standalone binary prints each
// finding as `file:line: [SEV RuleID] message`. File carries a path and Message
// carries link targets, both lifted from the linted (possibly hostile-clone)
// tree; only the two error exits were path-scrubbed. A control byte in either
// field replays a terminal escape into CI logs, and an absolute File leaks the
// developer's local layout (iss-29). The findings line must mask control runes
// (termsafe.Sanitize) and strip the repo root / home, mirroring the main abcd
// CLI renderer (cli.go:496) and the error exits. Attack runes are built
// numerically so this file carries none of them.
func TestRenderFindingMasksAndScrubs(t *testing.T) {
	root := t.TempDir()
	// A finding whose File is ABSOLUTE (under root) and carries a control byte in
	// the path tail, and whose Message carries an ESC + RLO override.
	poisonedFile := filepath.Join(root, "docs", "note\x1b.md")
	f := lint.Finding{
		File:     poisonedFile,
		Line:     7,
		RuleID:   "stray_root_docs",
		Severity: "blocker",
		Message:  "link \x1btarget‮ points outside the tree",
	}

	got := renderFinding(f, root)

	for _, r := range []rune{0x1b, 0x202e} {
		if strings.ContainsRune(got, r) {
			t.Errorf("iss-264: findings line carries a raw U+%04X attack rune; it must be masked like cli.go:496\nline: %q", r, got)
		}
	}
	if strings.Contains(got, root) {
		t.Errorf("iss-264: findings line leaks the absolute repo root %q (iss-29: no absolute path in machine output)\nline: %q", root, got)
	}
	// The line structure and the legitimate content survive.
	if !strings.Contains(got, "docs/note") || !strings.Contains(got, ":7:") ||
		!strings.Contains(got, "stray_root_docs") || !strings.Contains(got, "link") {
		t.Fatalf("iss-264: scrubbing must not drop the finding's structure/content:\n%q", got)
	}
}
