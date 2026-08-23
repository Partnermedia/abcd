package launch

import (
	"strings"
	"testing"
)

// gateNamed finds one gate row by name, or fails.
func gateNamed(t *testing.T, gates []GateSummary, name string) GateSummary {
	t.Helper()
	for _, g := range gates {
		if g.Name == name {
			return g
		}
	}
	t.Fatalf("no %q row in the gates list; got %v", name, gateNames(gates))
	return GateSummary{}
}

func gateNames(gates []GateSummary) []string {
	out := make([]string, 0, len(gates))
	for _, g := range gates {
		out = append(out, g.Name)
	}
	return out
}

// TestReceiptGateRowIsAlwaysPresent is the loud-staging guarantee that the
// v0.6.2 release failure turned into a requirement.
//
// The preview used to list five gates and omit receipt_gate, so it answered
// cleanly — bundle sized, scan clean, smoke ok — while the gate that actually
// refused the release went unmentioned. An absent row reads as "no such gate",
// which is the one thing it is not: the gate is fully implemented and armed at
// release time. The row must therefore be present in every state, including the
// state where nothing was measured at all.
func TestReceiptGateRowIsAlwaysPresent(t *testing.T) {
	cases := []struct {
		name string
		pre  *ReceiptPreflight
		want string // a substring the detail must carry
	}{
		{"not measured", nil, "not measured"},
		{"unreadable", &ReceiptPreflight{Unreadable: "receipts directory unreadable"}, "cannot read"},
		{"none recorded", &ReceiptPreflight{Commit: "abcdef1234567890"}, "no receipts recorded"},
		{"some recorded", &ReceiptPreflight{
			Commit:   "abcdef1234567890",
			Recorded: []string{"docs-currency-reviewer", "iss35-brief-surface-crosscheck"},
		}, "2 receipt(s) recorded"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			g := receiptGate(tc.pre)
			if g.Name != "semantic-receipts" {
				t.Errorf("Name = %q, want semantic-receipts", g.Name)
			}
			if g.Status != "host-run" {
				t.Errorf("Status = %q, want host-run — the preview must never report this gate as run or passed", g.Status)
			}
			if !strings.Contains(g.Detail, tc.want) {
				t.Errorf("Detail = %q, want it to contain %q", g.Detail, tc.want)
			}
			if !strings.Contains(g.Detail, "release-gate/README.md") {
				t.Errorf("Detail = %q must point at the runbook in every state", g.Detail)
			}
		})
	}
}

// TestReceiptGateNeverClaimsAPass pins the trust boundary. release.yml owns the
// required-gates list and judges receipt validity; a preview that claimed the
// gate satisfied would be a second, drifting copy of a security decision — and
// would restore exactly the false confidence that let a one-commit release
// branch reach a tag.
func TestReceiptGateNeverClaimsAPass(t *testing.T) {
	g := receiptGate(&ReceiptPreflight{
		Commit:   "abcdef1234567890",
		Recorded: []string{"docs-currency-reviewer", "iss35-brief-surface-crosscheck"},
	})
	lower := strings.ToLower(g.Detail)
	for _, banned := range []string{"satisfied", "all required", "passes", "ready to publish"} {
		if strings.Contains(lower, banned) {
			t.Errorf("Detail claims a verdict it cannot make (%q): %q", banned, g.Detail)
		}
	}
	if !strings.Contains(lower, "release.yml judges") {
		t.Errorf("Detail must defer the verdict to release.yml, got %q", g.Detail)
	}
}
