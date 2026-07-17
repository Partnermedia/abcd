package cli

import (
	"bytes"
	"strings"
	"testing"

	"github.com/REPPL/abcd-cli/internal/core/audit"
)

// TestRenderAuditHumanSanitizesUntrustedFields proves the human audit report
// neutralises terminal-display attack characters in the repo-derived File,
// Message, and Fix fields (`abcd audit` runs over any repo, so those are untrusted
// terminal output). A raw ESC/CSI/bidi rune in a finding could recolour, move the
// cursor, or visually reorder the report; none may reach the writer.
func TestRenderAuditHumanSanitizesUntrustedFields(t *testing.T) {
	res := audit.Result{
		Findings: []audit.Finding{{
			RuleID:   "privacy-hygiene",
			Severity: audit.SeverityError,
			File:     "a\u001b[31m.md", // ESC in a path
			Line:     3,
			Message:  "leak \u009b2K spoof",  // C1 CSI (2-byte encoded U+009B, not a raw 0x9b byte)
			Fix:      "edit \u202e reversed", // bidi override
		}},
		Blockers: 1,
	}
	var buf bytes.Buffer
	renderAuditHuman(&buf, res)
	out := buf.String()
	for _, r := range out {
		if r == 0x1b || (r >= 0x80 && r <= 0x9f) || (r >= 0x202A && r <= 0x202E) {
			t.Fatalf("unsanitised attack rune %U reached the audit report: %q", r, out)
		}
	}
	if !strings.Contains(out, "privacy-hygiene") {
		t.Errorf("expected the finding to still render: %q", out)
	}
}
