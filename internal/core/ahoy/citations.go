package ahoy

// The citation-baseline status line (spc-17's surfacing clause).
//
// The refresh verb is manual, so something has to remind a maintainer that the
// record it writes is ageing — otherwise the first anyone hears of a 400-day-old
// baseline is a release that will not cut. `ahoy` is where that belongs: it is
// the read-only "where am I" board a maintainer already looks at.
//
// It is a SIGNAL, not a gap. A stale baseline is not a discrepancy `ahoy install`
// could close — clearing it means fetching the network, and possibly a human
// opening links — so surfacing it as a resolvable gap would be a promise the
// apply path cannot keep.

import (
	"path/filepath"
	"strconv"

	"github.com/REPPL/abcd-cli/internal/core/lint"
)

// detectCitations renders the baseline's coverage and age as one line, or "" when
// this repo has not armed the citation gate.
//
// Every failure is soft. This is a read-only status board: a missing config, a
// disabled rule, or a corrupt baseline must leave the rest of `ahoy` renderable,
// so an unreadable record is REPORTED as unreadable rather than raised.
func detectCitations(cwd string) string {
	cfg, err := lint.LoadConfig(filepath.Join(cwd, ".abcd", "docs-lint.json"))
	if err != nil {
		return ""
	}
	if rc, ok := cfg.Rules["citation_baseline"]; !ok || !rc.Enabled {
		return ""
	}

	sum, err := lint.CitationAgeSummary(cfg, cwd)
	if err != nil {
		return "baseline unreadable — run `abcd docs cite refresh`"
	}
	if sum.Cited == 0 {
		return "nothing cited"
	}

	line := strconv.Itoa(sum.Cited) + " cited"
	if !sum.Present {
		return line + ", no baseline — run `abcd docs cite refresh`"
	}
	if sum.Missing > 0 {
		line += ", " + strconv.Itoa(sum.Missing) + " unreceipted"
	}
	if sum.Broken > 0 {
		line += ", " + strconv.Itoa(sum.Broken) + " broken"
	}
	if sum.Overdue > 0 {
		line += ", " + strconv.Itoa(sum.Overdue) + " overdue"
	}
	// Stale includes the overdue ones; report only the difference, so the two
	// numbers add up to what a reader expects rather than double-counting.
	if warnOnly := sum.Stale - sum.Overdue; warnOnly > 0 {
		line += ", " + strconv.Itoa(warnOnly) + " stale"
	}
	if sum.Healthy() {
		return line + ", all current"
	}
	if sum.OldestChecked != "" {
		line += " (oldest checked " + sum.OldestChecked + ")"
	}
	return line
}
