package lint

// The harness-leak family (harness_leak): the two shapes an agent harness stamps
// onto text the repository did not ask it to stamp — a live agent-session URL and
// a tool's own "generated with …" attribution footer.
//
// The class is DEFINED once, in the scanner's canonical pattern set
// (internal/adapter/scanner/harnessleak.go), and consulted here. That is the
// whole design of iss-178's second remedy: the leak reached a public artefact
// through the harness's append and, separately, through an agent reasoning its
// way around the prose convention, so the two surfaces that judge text — the
// outbound scrub before posting, and this lint over committed prose — must be
// unable to disagree about what a leak is. A second copy of a regex is how one of
// them comes to mean something weaker.
//
// It rides the same per-file walk as the banned-token family and honours the same
// two escapes: a fenced block is quoted material (a page must be able to SHOW the
// banned shape), and a line carrying the `abcd-lint:allow` waiver is deliberately
// illustrative — the spelling the privacy rule's own fix hint teaches, so the
// corpus converges on one token.

import (
	"strings"

	"github.com/intentdriven/abcd/internal/adapter/scanner"
)

const ruleHarnessLeak = "harness_leak"

// The line-scoped waiver escapes, in both the current and the pre-spc-29
// spellings the privacy rule honours forever (the token lives in committed
// content, so retiring one would silently re-flag every existing line).
const (
	harnessLeakWaiver      = "abcd-lint:allow"
	harnessLeakAuditWaiver = "abcd-audit:allow"
)

// checkHarnessLeak scans one file's lines for the harness-leak class.
func checkHarnessLeak(rel string, lines []string, mask []bool, cfg RuleConfig) []Finding {
	var out []Finding
	for i, line := range lines {
		if mask[i] || strings.Contains(line, harnessLeakWaiver) || strings.Contains(line, harnessLeakAuditWaiver) {
			continue
		}
		for _, p := range scanner.HarnessLeakPatterns() {
			loc := p.Re.FindStringIndex(line)
			if loc == nil {
				continue
			}
			if p.Skip != nil && p.Skip(line[loc[0]:loc[1]]) {
				continue
			}
			if p.SkipAt != nil && p.SkipAt(line, loc[0], loc[1]) {
				continue
			}
			out = append(out, Finding{
				File: rel, Line: i + 1, RuleID: ruleHarnessLeak, Severity: cfg.Severity,
				Message: "committed text carries a " + p.Label + "; " + scanner.OutboundPolicy +
					" (add `" + harnessLeakWaiver + "` on the line if it is deliberately illustrative)",
			})
			break // one finding per line: the citation points at the line, and the line is what gets fixed
		}
	}
	return out
}
