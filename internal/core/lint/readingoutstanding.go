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
	"strings"

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
	dispositionFileRe = regexp.MustCompile(`^(` + issueschema.DispositionFamily + `-[0-9]+)\.md$`)
	dispositionIDRe   = regexp.MustCompile(`^` + issueschema.DispositionFamily + `-[0-9]+$`)
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

// StandingDispositions returns the ids of the dispositions of item that no
// sibling supersedes — the answers currently in force, which is exactly one in a
// healthy ledger.
//
// It is exported for ONE reason: core/capture asks the same question with its own
// reader (it cannot import this package — this package is imported by capture's
// own tests, so the edge back would be a cycle), and two readers of one question
// must be held to one answer by a test. The verb refusing a second disposition
// the board says is not needed, or promoting an item the board still shows as
// held, are the divergences that test exists to prevent.
func StandingDispositions(repoRoot, issuesDir, item string) ([]string, error) {
	return standingDispositionIDs(filepath.Join(repoRoot, filepath.FromSlash(issuesDir)), item)
}

// standingDisposition returns the disposition of item that no sibling
// supersedes, or nil when the item is unanswered. The superseded records stay in
// place — a hold that vanished when it was answered would take its own exit
// condition with it — so "standing" is computed from the supersession edges
// rather than from what is present.
func standingDisposition(issuesRoot, issuesDir, item string) (*standingRecord, error) {
	records, standing, err := dispositionsOf(issuesRoot, issuesDir, item)
	if err != nil || len(standing) == 0 {
		return nil, err
	}
	// More than one standing answer is a ledger fault the write path refuses.
	// Reporting the FIRST by id keeps this report deterministic; naming the fault
	// belongs to the writer, which is where it can still be prevented.
	return records[standing[0]], nil
}

// standingDispositionIDs is the standing set alone, sorted — the value both this
// package and core/capture must agree on.
func standingDispositionIDs(issuesRoot, item string) ([]string, error) {
	_, standing, err := dispositionsOf(issuesRoot, "", item)
	return standing, err
}

// dispositionsOf reads one item's dispositions and computes which of them stand.
//
// standing = present minus superseded, and the second set is where the two
// readers of this question have to agree. A record that is not well-formed cannot
// be trusted to RETIRE another — it may say anything — but it cannot be dropped
// either, because then a malformed record would silently remove the answer it
// claims to replace. So it stands, and whoever looks is told there is something
// here to deal with. core/capture reaches the same verdict through its own strict
// parser; TestStandingDispositionAgreesAcrossBothReaders holds the two together.
func dispositionsOf(issuesRoot, issuesDir, item string) (map[string]*standingRecord, []string, error) {
	itemDir := filepath.Join(issuesRoot, issueschema.DispositionsDir, item)
	entries, err := os.ReadDir(itemDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil, nil
		}
		return nil, nil, err
	}
	records := map[string]*standingRecord{}
	superseded := map[string]bool{}
	for _, e := range entries {
		m := dispositionFileRe.FindStringSubmatch(e.Name())
		if e.IsDir() || m == nil {
			continue
		}
		content, err := os.ReadFile(filepath.Join(itemDir, e.Name()))
		if err != nil {
			return nil, nil, err
		}
		lines := strings.Split(string(content), "\n")
		fields := frontmatterFields(lines)
		id := m[1]
		records[id] = &standingRecord{
			id:            id,
			state:         fieldScalar(fields, "state"),
			exitCondition: fieldScalar(fields, "exit_condition"),
			rel:           filepath.Join(issuesDir, issueschema.DispositionsDir, item, e.Name()),
		}
		// A duplicated top-level key is malformed to every record reader here: the
		// strict parser refuses the file outright while this lenient scanner keeps
		// only the first value, so a second line can hide the value the first
		// shows. Such a record retires nothing.
		if len(frontmatterDuplicates(lines)) > 0 {
			continue
		}
		if s := fieldScalar(fields, "supersedes_disposition"); dispositionIDRe.MatchString(s) {
			superseded[s] = true
		}
	}
	var standing []string
	for id := range records {
		if !superseded[id] {
			standing = append(standing, id)
		}
	}
	sort.Strings(standing)
	return records, standing, nil
}

// fieldScalar reads a frontmatter value with its surrounding quotes stripped, so
// a quoted state compares as the writer's own enum member.
func fieldScalar(fields map[string]fmField, key string) string {
	f, ok := fields[key]
	if !ok {
		return ""
	}
	return strings.Trim(strings.TrimSpace(f.value), `"'`)
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
