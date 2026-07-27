package lint

// The citation baseline's health, and the release gate's promotion of it.
//
// Two surfaces need to say something about the baseline without running the
// whole lint: `abcd ahoy` reports its age at a glance, and the release preflight
// nags about entries approaching the 365-day blocker. Both read the summary
// below — one grading, so the number a maintainer sees at ahoy is by
// construction the number the release gate acts on. A second count computed
// slightly differently is how a status line starts lying.

import (
	"errors"
	"os"
	"path/filepath"
	"sort"
	"time"
)

// CitationNagWindowDays is how far ahead of the blocking threshold the release
// preflight starts naming entries. A month is enough notice to clear a manual
// queue before a release, and short enough that the nag still means something.
const CitationNagWindowDays = 30

// CitationSummary is the committed baseline's state against the repo's current
// citations. Every count is over the CITED set: an entry for a URL the docs no
// longer cite is not a health problem, it is something the next refresh drops.
type CitationSummary struct {
	// Present reports whether a baseline file exists at all.
	Present bool `json:"present"`
	// Cited is how many distinct URLs the documentation cites.
	Cited int `json:"cited"`
	// Recorded is how many of those have a baseline entry; Missing is the rest.
	Recorded int `json:"recorded"`
	Missing  int `json:"missing"`
	// Broken counts entries recorded as broken.
	Broken int `json:"broken"`
	// Manual counts entries a human established. They age on the same clock as
	// automatic ones, so this is reported, never subtracted from the stale count.
	Manual int `json:"manual"`
	// Stale counts entries past the warn threshold (which includes every
	// approaching and overdue one); Overdue counts those past the blocking one;
	// Approaching counts the band just short of blocking.
	Stale       int `json:"stale"`
	Overdue     int `json:"overdue"`
	Approaching int `json:"approaching"`
	// OldestChecked is the earliest last_checked date among cited entries.
	OldestChecked string `json:"oldest_checked,omitempty"`
	// OverdueURLs and ApproachingURLs name what a nag must list, sorted.
	OverdueURLs     []string `json:"overdue_urls,omitempty"`
	ApproachingURLs []string `json:"approaching_urls,omitempty"`
}

// Healthy reports whether the baseline needs no attention: every citation has a
// current, unbroken receipt.
func (s CitationSummary) Healthy() bool {
	return s.Present && s.Missing == 0 && s.Broken == 0 && s.Stale == 0
}

// CitationAgeSummary grades the baseline against the wall clock.
func CitationAgeSummary(cfg Config, repoRoot string) (CitationSummary, error) {
	return CitationAgeSummaryAt(cfg, repoRoot, time.Now())
}

// CitationAgeSummaryAt is CitationAgeSummary with the clock supplied, so the
// band boundaries are testable exactly.
//
// A missing baseline is reported, not an error: a repo that has not refreshed
// yet is a state a status line must be able to describe. A MALFORMED one is
// still an error — a record the gate cannot read is not a health report.
func CitationAgeSummaryAt(cfg Config, repoRoot string, now time.Time) (CitationSummary, error) {
	relPath, warnDays, blockDays, err := CitationPolicy(cfg)
	if err != nil {
		return CitationSummary{}, err
	}

	cited, err := CollectCitedURLs(cfg, repoRoot)
	if err != nil {
		return CitationSummary{}, err
	}
	sum := CitationSummary{Cited: len(cited)}

	base, err := LoadBaseline(filepath.Join(repoRoot, filepath.FromSlash(relPath)))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			sum.Missing = len(cited)
			return sum, nil
		}
		return CitationSummary{}, err
	}
	sum.Present = true

	for _, ref := range cited {
		e, ok := base.Entries[ref.URL]
		if !ok {
			sum.Missing++
			continue
		}
		sum.Recorded++
		if e.Outcome == OutcomeBroken {
			sum.Broken++
		}
		if e.Verification == VerificationManual {
			sum.Manual++
		}
		if sum.OldestChecked == "" || e.LastChecked < sum.OldestChecked {
			sum.OldestChecked = e.LastChecked
		}
		checked, perr := e.Checked()
		if perr != nil {
			continue // unreachable: LoadBaseline validated the date
		}
		age := DaysBetween(checked, now)
		if age >= warnDays {
			sum.Stale++
		}
		switch {
		case age >= blockDays:
			sum.Overdue++
			sum.OverdueURLs = append(sum.OverdueURLs, ref.URL)
		case age >= blockDays-CitationNagWindowDays:
			sum.Approaching++
			sum.ApproachingURLs = append(sum.ApproachingURLs, ref.URL)
		}
	}
	sort.Strings(sum.OverdueURLs)
	sort.Strings(sum.ApproachingURLs)
	return sum, nil
}

// ArmCitationOverdue returns cfg with the citation staleness blocker armed: an
// entry past the block threshold becomes a BLOCKER finding rather than the warn
// the commit gate emits.
//
// It follows ArmReceiptGate's shape deliberately. The promotion is the CALLER's
// decision — a release gate, not the in-tree config — so a repo cannot defang
// its own release by writing "overdue_severity": "warn" into a committed file,
// and an ordinary commit is never calendar-blocked (spc-17). What arming does
// NOT do is enable a rule the repo turned off: promoting a severity is the
// release gate's business, adopting a rule is the repo's.
func ArmCitationOverdue(cfg Config) Config {
	rules := make(map[string]RuleConfig, len(cfg.Rules))
	for k, v := range cfg.Rules {
		rules[k] = v
	}
	rc := rules[ruleCitationBaseline]
	rc.OverdueSeverity = severityBlocker
	rules[ruleCitationBaseline] = rc
	cfg.Rules = rules
	return cfg
}
