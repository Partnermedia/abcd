package lint

// The outstanding-readings report (reading_outstanding): every reading item that
// carries no disposition, and every open hold with its exit condition.
//
// It is a REPORT, not a gate, and the difference is enforced in code rather than
// left to configuration. Its severity is pinned to `info` here and never read
// from RuleConfig, because a rule whose severity a config could raise to blocker
// is a gate waiting to happen — and a reading must never block a push that has
// nothing to do with it. What it protects against is quieter than a gate and
// harder to notice: nothing in this design means "already covered", so an
// unanswered item has no state to sit in, and without this report it would
// simply not appear anywhere.
//
// The status signal it reads is the presence of the KEYED disposition directory,
// never folder membership — one probe, and the reason the reading families are
// deliberately absent from issueschema.StatusDirs.

import (
	"os"
	"path/filepath"
	"regexp"
	"sort"

	"github.com/intentdriven/abcd/internal/core/issueschema"
)

const ruleReadingOutstanding = "reading_outstanding"

// severityInfo is the report tier: printed by every surface, counted by no
// gate. record-lint exits non-zero on `blocker` alone, so an info finding is
// visible and inert by construction.
const severityInfo = "info"

var (
	readingRunDirRe   = regexp.MustCompile(`^` + issueschema.ReadingRunFamily + `-[0-9]+$`)
	readingItemFileRe = regexp.MustCompile(`^(` + issueschema.ReadingItemFamily + `-[0-9]+)\.md$`)
)

// OutstandingItem is one reading item nobody has answered.
type OutstandingItem struct {
	Item string `json:"item"`
	Run  string `json:"run"`
	Path string `json:"path"`
}

// OpenHold is one standing `held` disposition, rendered with the exit condition
// that is the only thing distinguishing a hold from a parking space.
type OpenHold struct {
	Item          string `json:"item"`
	Disposition   string `json:"disposition"`
	ExitCondition string `json:"exit_condition"`
	Path          string `json:"path"`
}

// OutstandingReadings is the whole report, ordered deterministically.
type OutstandingReadings struct {
	Undispositioned []OutstandingItem `json:"undispositioned"`
	OpenHolds       []OpenHold        `json:"open_holds"`
}

// Empty reports whether there is nothing outstanding — the ordinary state of a
// repository that has commissioned no reading, and the state a surface renders
// as silence rather than as a heading with nothing under it.
func (r OutstandingReadings) Empty() bool {
	return len(r.Undispositioned) == 0 && len(r.OpenHolds) == 0
}

// ReadReadingOutstanding builds the report from the ledger at issuesDir
// (repo-relative). It is exported because the report has two surfaces — the lint
// findings and the bare capture status board — and two implementations of one
// question would be two answers to it. An absent readings tree is not an error:
// a repository that has commissioned no reading is in a state, not a fault.
func ReadReadingOutstanding(repoRoot, issuesDir string) (OutstandingReadings, error) {
	var report OutstandingReadings
	issuesRoot := filepath.Join(repoRoot, filepath.FromSlash(issuesDir))
	readingsRoot := filepath.Join(issuesRoot, issueschema.ReadingsDir)
	runs, err := os.ReadDir(readingsRoot)
	if err != nil {
		if os.IsNotExist(err) {
			return report, nil
		}
		return report, err
	}

	for _, run := range runs {
		if !run.IsDir() || !readingRunDirRe.MatchString(run.Name()) {
			continue
		}
		entries, err := os.ReadDir(filepath.Join(readingsRoot, run.Name()))
		if err != nil {
			return OutstandingReadings{}, err
		}
		for _, e := range entries {
			m := readingItemFileRe.FindStringSubmatch(e.Name())
			if e.IsDir() || m == nil {
				continue
			}
			item := m[1]
			rel := filepath.Join(issuesDir, issueschema.ReadingsDir, run.Name(), e.Name())
			standing, err := standingDisposition(issuesRoot, issuesDir, item)
			if err != nil {
				return OutstandingReadings{}, err
			}
			switch {
			case standing == nil:
				report.Undispositioned = append(report.Undispositioned, OutstandingItem{
					Item: item, Run: run.Name(), Path: filepath.ToSlash(rel),
				})
			case standing.state == issueschema.DispositionHeld:
				report.OpenHolds = append(report.OpenHolds, OpenHold{
					Item: item, Disposition: standing.id,
					ExitCondition: standing.exitCondition,
					Path:          filepath.ToSlash(standing.rel),
				})
			}
		}
	}

	sort.Slice(report.Undispositioned, func(i, j int) bool {
		return report.Undispositioned[i].Item < report.Undispositioned[j].Item
	})
	sort.Slice(report.OpenHolds, func(i, j int) bool {
		return report.OpenHolds[i].Item < report.OpenHolds[j].Item
	})
	return report, nil
}

// standingRecord is the disposition currently in force for one item.
type standingRecord struct {
	id            string
	state         string
	exitCondition string
	rel           string
}

// standingDisposition returns the disposition of item that no sibling
// supersedes, or nil when the item is unanswered.
func standingDisposition(issuesRoot, issuesDir, item string) (*standingRecord, error) {
	itemDir := filepath.Join(issuesRoot, issueschema.DispositionsDir, item)
	entries, err := os.ReadDir(itemDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var records []issueschema.DispositionRecord
	rel := map[string]string{}
	for _, e := range entries {
		id, ok := issueschema.DispositionFileID(e.Name())
		if e.IsDir() || !ok {
			continue
		}
		content, err := os.ReadFile(filepath.Join(itemDir, e.Name()))
		if err != nil {
			return nil, err
		}
		records = append(records, issueschema.ParseDisposition(id, string(content)))
		rel[id] = filepath.Join(issuesDir, issueschema.DispositionsDir, item, e.Name())
	}

	// The WALK is here; the JUDGEMENT is not. Which records stand is decided by
	// issueschema.StandingDispositionIDs, the one reader core/capture calls too:
	// a record neither can read retires nothing and does not vanish either, so
	// the board and the verb cannot disagree about what is in force.
	standing := issueschema.StandingDispositionIDs(records)
	if len(standing) == 0 {
		return nil, nil
	}
	// More than one standing answer is a ledger fault the write path refuses.
	// Reporting the FIRST by id keeps this report deterministic; naming the fault
	// belongs to the writer, which is where it can still be prevented.
	for _, r := range records {
		if r.ID != standing[0] {
			continue
		}
		return &standingRecord{
			id: r.ID, state: r.State, exitCondition: r.ExitCondition, rel: rel[r.ID],
		}, nil
	}
	return nil, nil
}

// checkReadingOutstanding renders the report as findings, every one of them at
// severityInfo whatever the configuration says.
func checkReadingOutstanding(repoRoot string, cfg RuleConfig) ([]Finding, error) {
	report, err := ReadReadingOutstanding(repoRoot, issuesDirOf(cfg))
	if err != nil {
		return nil, err
	}
	var out []Finding
	for _, o := range report.Undispositioned {
		out = append(out, Finding{
			File: o.Path, Line: 1, RuleID: ruleReadingOutstanding, Severity: severityInfo,
			Message: o.Item + " (run " + o.Run + ") carries no disposition — outstanding. " +
				"An item nobody has answered has no state to sit in, because nothing in this " +
				"vocabulary means \"already covered\"; answer it with `abcd capture disposition " +
				o.Item + " --state <state> ...`",
		})
	}
	for _, h := range report.OpenHolds {
		out = append(out, Finding{
			File: h.Path, Line: 1, RuleID: ruleReadingOutstanding, Severity: severityInfo,
			Message: h.Item + " is held (" + h.Disposition + "), exit condition: " + h.ExitCondition +
				". A hold exits only through a superseding disposition that cites it — never by expiry, and never silently",
		})
	}
	return out, nil
}
