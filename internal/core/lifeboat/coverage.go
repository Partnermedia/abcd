package lifeboat

import (
	"fmt"
	"strings"
)

// Coverage is one repository's probe result: which brief sections a lifeboat
// could ground from it, at what confidence, citing what. It is the M2
// experiment's per-repo readout, and it aggregates across repositories (see
// Aggregate) — that cross-repo table is what answers "is the brief structure
// sound."
type Coverage struct {
	SchemaVersion int               `json:"schema_version"`
	Repo          RepoInfo          `json:"repo"`
	TiersPresent  []Tier            `json:"tiers_present"`
	Sections      []SectionCoverage `json:"sections"`
	Summary       Summary           `json:"summary"`
}

// RepoInfo identifies the probed repository.
type RepoInfo struct {
	Name    string `json:"name"`
	RootSHA string `json:"root_sha,omitempty"`
	Commits int    `json:"commits"`
}

// SectionCoverage is the probe result for one brief section. A grounded or
// partial row cites evidence; a blank row carries what was searched and the
// question a human must answer.
type SectionCoverage struct {
	Name       Section    `json:"name"`
	Status     Status     `json:"status"`
	Confidence Confidence `json:"confidence,omitempty"`
	Tier       Tier       `json:"tier,omitempty"`
	Evidence   []string   `json:"evidence,omitempty"`
	Searched   []string   `json:"searched,omitempty"`
	Question   string     `json:"question,omitempty"`
}

// Summary counts sections by status. Blank is counted, not hidden — a blank is
// a result.
type Summary struct {
	Grounded int `json:"grounded"`
	Partial  int `json:"partial"`
	Blank    int `json:"blank"`
}

// Render returns the human-readable per-repo coverage report.
func (c Coverage) Render() string {
	var b strings.Builder
	tiers := make([]string, len(c.TiersPresent))
	for i, t := range c.TiersPresent {
		tiers[i] = string(t)
	}
	fmt.Fprintf(&b, "coverage for %s", sanitize(c.Repo.Name))
	if c.Repo.Commits > 0 {
		fmt.Fprintf(&b, " (%d commits)", c.Repo.Commits)
	}
	b.WriteString("\n")
	fmt.Fprintf(&b, "tiers present: %s\n", tiersOrNone(tiers))
	fmt.Fprintf(&b, "grounded %d · partial %d · blank %d  (of %d sections)\n\n",
		c.Summary.Grounded, c.Summary.Partial, c.Summary.Blank, len(c.Sections))

	for _, s := range c.Sections {
		mark := statusGlyph(s.Status)
		fmt.Fprintf(&b, "%s %-32s %s", mark, s.Name, s.Status)
		if s.Confidence != "" {
			fmt.Fprintf(&b, " (%s, %s)", s.Tier, s.Confidence)
		}
		b.WriteString("\n")
		if len(s.Evidence) > 0 {
			fmt.Fprintf(&b, "    evidence: %s\n", strings.Join(sanitizeAll(s.Evidence), ", "))
		}
		if s.Status == StatusBlank {
			if len(s.Searched) > 0 {
				fmt.Fprintf(&b, "    searched: %s\n", strings.Join(sanitizeAll(s.Searched), ", "))
			}
			if s.Question != "" {
				fmt.Fprintf(&b, "    ? %s\n", sanitize(s.Question))
			}
		}
	}
	return b.String()
}

func statusGlyph(s Status) string {
	switch s {
	case StatusGrounded:
		return "+"
	case StatusPartial:
		return "~"
	default:
		return "-"
	}
}

func tiersOrNone(t []string) string {
	if len(t) == 0 {
		return "(none)"
	}
	return strings.Join(t, ", ")
}

// sanitize strips terminal control characters from a string before it is
// rendered to a terminal. Evidence, searched entries, questions, and the repo
// name are built from repository content — commit subjects, file paths, tag
// names — which a hostile or archived repo controls. Left raw, an ANSI escape
// in a commit subject could spoof or corrupt the human report. C0 controls and
// DEL are replaced with a visible caret; tab is kept as a space. (The JSON
// output is unaffected — encoding/json escapes control characters itself.)
func sanitize(s string) string {
	return strings.Map(func(r rune) rune {
		switch {
		case r == '\t':
			return ' '
		case r < 0x20 || r == 0x7f:
			return '?'
		}
		return r
	}, s)
}

// sanitizeAll sanitizes every member of a slice.
func sanitizeAll(in []string) []string {
	if len(in) == 0 {
		return nil
	}
	out := make([]string, len(in))
	for i, s := range in {
		out[i] = sanitize(s)
	}
	return out
}

// AggregateReport is the cross-repo readout: one row per brief section, one
// column per probed repository, each cell the section's status in that repo.
// This is the artefact the M2 gate reads to decide which brief sections survive.
type AggregateReport struct {
	SchemaVersion int              `json:"schema_version"`
	Repos         []AggregateRepo  `json:"repos"`
	Sections      []AggregateRow   `json:"sections"`
	Verdict       []SectionVerdict `json:"verdict"`
}

// AggregateRepo is a probed repository's identity in the aggregate header.
type AggregateRepo struct {
	Name         string `json:"name"`
	Commits      int    `json:"commits"`
	TiersPresent []Tier `json:"tiers_present"`
}

// AggregateRow is one brief section across every repository.
type AggregateRow struct {
	Section Section           `json:"section"`
	Cells   map[string]Status `json:"cells"` // repo name -> status
}

// SectionVerdict summarises a section across the corpus: how the record-rich
// repos fared versus the record-poor ones. It is the quantified answer to "what
// is the record worth" for that section.
type SectionVerdict struct {
	Section       Section `json:"section"`
	GroundedCount int     `json:"grounded_count"`
	PartialCount  int     `json:"partial_count"`
	BlankCount    int     `json:"blank_count"`
	// AlwaysBlank is true when no probed repo grounded or partially grounded the
	// section — evidence the section may not be derivable from a repository at
	// all, and belongs to a human rather than an extraction.
	AlwaysBlank bool `json:"always_blank"`
}

// Aggregate reduces per-repo coverage reports to the cross-repo table. Repos
// keep the order given; sections keep the mapping's canonical order.
func Aggregate(covs []Coverage) AggregateReport {
	repos := make([]AggregateRepo, 0, len(covs))
	names := make([]string, 0, len(covs))
	seen := map[string]bool{}
	for _, c := range covs {
		// Disambiguate a duplicate repo name so its column is not overwritten,
		// probing past a suffix that itself collides with another repo's real
		// name (e.g. "foo", "foo#2", "foo") — the cells map is keyed by name, so
		// any collision would silently drop a column.
		name := sanitize(c.Repo.Name)
		if seen[name] {
			for n := 2; ; n++ {
				cand := fmt.Sprintf("%s#%d", name, n)
				if !seen[cand] {
					name = cand
					break
				}
			}
		}
		seen[name] = true
		repos = append(repos, AggregateRepo{
			Name: name, Commits: c.Repo.Commits, TiersPresent: c.TiersPresent,
		})
		names = append(names, name)
	}

	rows := make([]AggregateRow, 0, len(Table))
	verdicts := make([]SectionVerdict, 0, len(Table))
	for _, m := range Table {
		cells := map[string]Status{}
		v := SectionVerdict{Section: m.Section}
		for i, c := range covs {
			st := statusOf(c, m.Section)
			cells[names[i]] = st
			switch st {
			case StatusGrounded:
				v.GroundedCount++
			case StatusPartial:
				v.PartialCount++
			default:
				v.BlankCount++
			}
		}
		v.AlwaysBlank = len(covs) > 0 && v.GroundedCount == 0 && v.PartialCount == 0
		rows = append(rows, AggregateRow{Section: m.Section, Cells: cells})
		verdicts = append(verdicts, v)
	}
	return AggregateReport{
		SchemaVersion: SchemaVersion,
		Repos:         repos,
		Sections:      rows,
		Verdict:       verdicts,
	}
}

func statusOf(c Coverage, section Section) Status {
	for _, s := range c.Sections {
		if s.Name == section {
			return s.Status
		}
	}
	return StatusBlank
}

// Render returns the human-readable cross-repo aggregate table.
func (a AggregateReport) Render() string {
	var b strings.Builder
	b.WriteString("cross-repo brief coverage\n\n")
	for _, r := range a.Repos {
		tiers := make([]string, len(r.TiersPresent))
		for i, t := range r.TiersPresent {
			tiers[i] = string(t)
		}
		fmt.Fprintf(&b, "  %s — %d commits, tiers: %s\n", r.Name, r.Commits, tiersOrNone(tiers))
	}
	b.WriteString("\n")

	// Column widths.
	nameW := len("brief section")
	for _, row := range a.Sections {
		if len(string(row.Section)) > nameW {
			nameW = len(string(row.Section))
		}
	}
	colW := make([]int, len(a.Repos))
	for i, r := range a.Repos {
		colW[i] = len(r.Name)
		if colW[i] < 8 {
			colW[i] = 8
		}
	}

	fmt.Fprintf(&b, "%-*s", nameW, "brief section")
	for i, r := range a.Repos {
		fmt.Fprintf(&b, "  %-*s", colW[i], r.Name)
	}
	b.WriteString("  verdict\n")

	for ri, row := range a.Sections {
		fmt.Fprintf(&b, "%-*s", nameW, row.Section)
		for i, r := range a.Repos {
			fmt.Fprintf(&b, "  %-*s", colW[i], row.Cells[r.Name])
		}
		v := a.Verdict[ri]
		if v.AlwaysBlank {
			b.WriteString("  always-blank")
		}
		b.WriteString("\n")
	}

	alwaysBlank := 0
	for _, v := range a.Verdict {
		if v.AlwaysBlank {
			alwaysBlank++
		}
	}
	fmt.Fprintf(&b, "\n%d of %d sections are blank in every probed repo.\n", alwaysBlank, len(a.Sections))
	return b.String()
}

// AlwaysBlankSections returns the sections no probed repo could ground, in the
// mapping's canonical order.
func (a AggregateReport) AlwaysBlankSections() []Section {
	var out []Section
	for _, v := range a.Verdict {
		if v.AlwaysBlank {
			out = append(out, v.Section)
		}
	}
	return out
}
