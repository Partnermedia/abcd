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
	"strconv"
	"strings"

	"github.com/intentdriven/abcd/internal/core/issueschema"
	"github.com/intentdriven/abcd/internal/fsutil"
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
	// Unsafe names the repo-relative directories the walk declined to enter
	// because they are not real directories. core/capture REFUSES these outright,
	// because its read is followed by a write; this walk is genuinely read-only,
	// so it declines and SAYS SO instead. Going quiet is the thing it must not do:
	// a tree nobody walked looks exactly like a tree with nothing in it, and "no
	// outstanding items" is the one answer this report must never give by
	// accident.
	Unsafe []string `json:"unsafe,omitempty"`
	// Cyclic names items whose dispositions supersede one another, so every
	// answer is retired and none stands. It is a ledger fault, not an unanswered
	// item: reporting it as outstanding would be a confident wrong statement about
	// an item that has been answered twice, and would invite exactly the fresh
	// uncited answer the write path refuses.
	Cyclic []OutstandingItem `json:"cyclic,omitempty"`
	// Contested names items on which more than one disposition stands. It is not
	// an exotic state: two branches each answer an item, both merge without
	// conflict, and neither cites the other. Picking the first by id resolved that
	// silently — an `accepted` sorting first hid a `held` and its exit condition,
	// and no line said two answers stood. The report names every standing id and
	// resolves nothing, because which answer is in force is the researcher's
	// judgement and there is nothing here to make it from.
	Contested []ContestedItem `json:"contested,omitempty"`
	// Unreadable names items whose SOLE standing answer is a record no reader can
	// read. Such a record has no state, so it matched none of the report's cases
	// and the item simply vanished from the board — and an item whose only answer
	// is unreadable is the case most in need of a line, not least.
	Unreadable []UnreadableAnswer `json:"unreadable,omitempty"`
}

// UnreadableAnswer is one item whose standing disposition cannot be read.
type UnreadableAnswer struct {
	Item        string `json:"item"`
	Run         string `json:"run"`
	Path        string `json:"path"`
	Disposition string `json:"disposition"`
}

// ContestedItem is one item with more than one standing answer.
type ContestedItem struct {
	Item string `json:"item"`
	Run  string `json:"run"`
	Path string `json:"path"`
	// Standing is every standing id, sorted — the whole fault, not a sample.
	Standing []string `json:"standing"`
}

// Empty reports whether there is nothing outstanding — the ordinary state of a
// repository that has commissioned no reading, and the state a surface renders
// as silence rather than as a heading with nothing under it.
func (r OutstandingReadings) Empty() bool {
	return len(r.Undispositioned) == 0 && len(r.OpenHolds) == 0 &&
		len(r.Unsafe) == 0 && len(r.Cyclic) == 0 && len(r.Contested) == 0 &&
		len(r.Unreadable) == 0
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
	if !realDir(readingsRoot) {
		report.Unsafe = append(report.Unsafe, filepath.ToSlash(filepath.Join(issuesDir, issueschema.ReadingsDir)))
		return report, nil
	}
	runs, err := os.ReadDir(readingsRoot)
	if err != nil {
		if os.IsNotExist(err) {
			return report, nil
		}
		return report, err
	}
	// The dispositions family root answers for every item below, so a link there
	// silently empties the standing set of ALL of them — every item would read as
	// unanswered. It is checked once, before any item is judged.
	dispositionsRoot := filepath.Join(issuesRoot, issueschema.DispositionsDir)
	dispositionsReadable := realDir(dispositionsRoot)
	if !dispositionsReadable {
		report.Unsafe = append(report.Unsafe, filepath.ToSlash(filepath.Join(issuesDir, issueschema.DispositionsDir)))
	}

	for _, run := range runs {
		if !readingRunDirRe.MatchString(run.Name()) {
			continue
		}
		runDir := filepath.Join(readingsRoot, run.Name())
		// A symlink is a directory to ReadDir, so the check is on the entry
		// itself, not on whether the walk would descend into it.
		if !realDir(runDir) {
			report.Unsafe = append(report.Unsafe, filepath.ToSlash(filepath.Join(issuesDir, issueschema.ReadingsDir, run.Name())))
			continue
		}
		entries, err := os.ReadDir(runDir)
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
			if !dispositionsReadable {
				// The item's answer is unreadable, which is not the same fact as
				// "unanswered" — reporting it outstanding would be a confident
				// wrong statement, and the Unsafe line above already says why.
				continue
			}
			answer, err := standingDisposition(issuesRoot, issuesDir, item)
			if err != nil {
				return OutstandingReadings{}, err
			}
			switch {
			case len(answer.unsafe) > 0:
				// A path the walk could not read supports no claim about whether the
				// item has been answered. Saying "nobody has answered it" here would
				// invite the answer to be written a second time.
				report.Unsafe = append(report.Unsafe, answer.unsafe...)
			case answer.cyclic:
				report.Cyclic = append(report.Cyclic, OutstandingItem{
					Item: item, Run: run.Name(), Path: filepath.ToSlash(rel),
				})
			case len(answer.contested) > 1:
				report.Contested = append(report.Contested, ContestedItem{
					Item: item, Run: run.Name(), Path: filepath.ToSlash(rel),
					Standing: answer.contested,
				})
			case answer.standing == nil:
				report.Undispositioned = append(report.Undispositioned, OutstandingItem{
					Item: item, Run: run.Name(), Path: filepath.ToSlash(rel),
				})
			case !answer.standing.wellFormed:
				report.Unreadable = append(report.Unreadable, UnreadableAnswer{
					Item: item, Run: run.Name(), Path: filepath.ToSlash(answer.standing.rel),
					Disposition: answer.standing.id,
				})
			}
			// Every standing hold is rendered, whatever else is true of the item —
			// including an item the walk has just called contested. Naming the
			// contest and then withholding the one thing a hold exists to publish
			// would trade one silence for another. It sits outside the switch
			// because a hold is not an alternative to the facts above; it is an
			// additional one.
			report.OpenHolds = append(report.OpenHolds, answer.holds...)
		}
	}

	sort.Slice(report.Undispositioned, func(i, j int) bool {
		return report.Undispositioned[i].Item < report.Undispositioned[j].Item
	})
	sort.Slice(report.OpenHolds, func(i, j int) bool {
		return report.OpenHolds[i].Item < report.OpenHolds[j].Item
	})
	sort.Slice(report.Cyclic, func(i, j int) bool { return report.Cyclic[i].Item < report.Cyclic[j].Item })
	sort.Slice(report.Contested, func(i, j int) bool { return report.Contested[i].Item < report.Contested[j].Item })
	sort.Slice(report.Unreadable, func(i, j int) bool { return report.Unreadable[i].Item < report.Unreadable[j].Item })
	sort.Strings(report.Unsafe)
	return report, nil
}

// realDir reports whether path is a directory the walk may enter — present, and
// not a symlink. An ABSENT path is not unsafe: an unpopulated tree is a state,
// and the caller distinguishes the two by probing separately.
func realDir(path string) bool {
	fi, err := os.Lstat(path)
	if err != nil {
		return os.IsNotExist(err)
	}
	return fi.IsDir() && fi.Mode()&os.ModeSymlink == 0
}

// standingRecord is the disposition currently in force for one item.
type standingRecord struct {
	id            string
	state         string
	exitCondition string
	rel           string
	// wellFormed carries issueschema's verdict forward, so a sole standing record
	// nobody can read is reported as unreadable rather than falling through every
	// case on the strength of its empty state.
	wellFormed bool
}

// itemAnswer is everything the walk learned about one item's dispositions. It is
// a struct rather than a tuple because the facts are not alternatives to be
// squeezed into one return value: an item can be contested AND carry a hold, and
// collapsing that pair is exactly how the hold went missing.
type itemAnswer struct {
	// standing is the single record in force, or nil when none is.
	standing *standingRecord
	// contested is every standing id when more than one stands.
	contested []string
	// cyclic reports records present with none standing — a supersession cycle.
	cyclic bool
	// holds is every standing record that is a hold, so an exit condition is
	// published whatever else is true of the item.
	holds []OpenHold
	// unsafe names the repo-relative paths the walk declined to read. Any one of
	// them means the item's answer is unknown, not absent.
	unsafe []string
}

// standingDisposition reads one item's dispositions and says what stands.
func standingDisposition(issuesRoot, issuesDir, item string) (itemAnswer, error) {
	var answer itemAnswer
	itemDir := filepath.Join(issuesRoot, issueschema.DispositionsDir, item)
	if !realDir(itemDir) {
		answer.unsafe = append(answer.unsafe,
			filepath.ToSlash(filepath.Join(issuesDir, issueschema.DispositionsDir, item)))
		return answer, nil
	}
	entries, err := os.ReadDir(itemDir)
	if err != nil {
		if os.IsNotExist(err) {
			return answer, nil
		}
		return answer, err
	}
	var records []issueschema.DispositionRecord
	byID := map[string]issueschema.DispositionRecord{}
	rel := map[string]string{}
	for _, e := range entries {
		id, ok := issueschema.DispositionFileID(e.Name())
		if !ok {
			continue
		}
		// ReadGuarded rather than Lstat-then-ReadFile: it opens once with
		// O_NOFOLLOW and validates on the SAME descriptor, so no symlink swap fits
		// between the check and the read. A record it refuses makes the item's
		// answer UNKNOWN, which is not the same fact as unanswered.
		content, err := fsutil.ReadGuarded(filepath.Join(itemDir, e.Name()), citationPageSizeLimit)
		if err != nil {
			answer.unsafe = append(answer.unsafe,
				filepath.ToSlash(filepath.Join(issuesDir, issueschema.DispositionsDir, item, e.Name())))
			continue
		}
		r := issueschema.ParseDisposition(id, string(content))
		records = append(records, r)
		byID[id] = r
		rel[id] = filepath.Join(issuesDir, issueschema.DispositionsDir, item, e.Name())
	}

	// The WALK is here; the JUDGEMENT is not. Which records stand is decided by
	// issueschema.StandingDispositionIDs, the one reader core/capture calls too:
	// a record neither can read retires nothing and does not vanish either, so
	// the board and the verb cannot disagree about what is in force.
	standing := issueschema.StandingDispositionIDs(records)
	if len(standing) == 0 {
		// Records present with nothing standing is a supersession CYCLE; no
		// records at all is simply an item nobody has answered. Rendering the two
		// the same way would announce an item answered twice as unanswered.
		answer.cyclic = len(records) > 0
		return answer, nil
	}

	for _, id := range standing {
		r := byID[id]
		if r.State != issueschema.DispositionHeld {
			continue
		}
		answer.holds = append(answer.holds, OpenHold{
			Item: item, Disposition: id, ExitCondition: r.ExitCondition,
			Path: filepath.ToSlash(rel[id]),
		})
	}
	if len(standing) > 1 {
		// No first-by-id tie-break. Which answer is in force is the researcher's
		// judgement, and there is nothing here to make it from; choosing one would
		// publish a verdict the ledger does not contain.
		answer.contested = standing
		return answer, nil
	}

	r := byID[standing[0]]
	answer.standing = &standingRecord{
		id: r.ID, state: r.State, exitCondition: r.ExitCondition, rel: rel[r.ID],
		wellFormed: r.WellFormed,
	}
	return answer, nil
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
	for _, u := range report.Unreadable {
		out = append(out, Finding{
			File: u.Path, Line: 1, RuleID: ruleReadingOutstanding, Severity: severityInfo,
			Message: u.Item + " (run " + u.Run + ") stands on " + u.Disposition +
				", which no reader of this ledger can read — its frontmatter is malformed, so the record carries no state. " +
				"The item is answered and the answer is illegible; repair the record rather than writing a second one",
		})
	}
	for _, u := range report.Unsafe {
		out = append(out, Finding{
			File: u, Line: 1, RuleID: ruleReadingOutstanding, Severity: severityInfo,
			Message: "this is not a real directory (a symlink, or not a directory at all), so the reading walk did not enter it — " +
				"the items under it are neither reported outstanding nor confirmed answered. " +
				"`abcd capture` refuses to read the reading trees through a link, because its read is followed by a write",
		})
	}
	for _, c := range report.Contested {
		out = append(out, Finding{
			File: c.Path, Line: 1, RuleID: ruleReadingOutstanding, Severity: severityInfo,
			Message: c.Item + " (run " + c.Run + ") has " + strconv.Itoa(len(c.Standing)) +
				" standing answers, none superseding another: " + strings.Join(c.Standing, ", ") +
				". Which one is in force is a judgement the ledger does not contain, so nothing here picks one — " +
				"supersede the answers that are no longer meant to stand",
		})
	}
	for _, c := range report.Cyclic {
		out = append(out, Finding{
			File: c.Path, Line: 1, RuleID: ruleReadingOutstanding, Severity: severityInfo,
			Message: c.Item + " (run " + c.Run + ") carries dispositions that supersede one another, so none stands — " +
				"a ledger fault, not an unanswered item. No write path can produce it and only a hand edit can repair it; " +
				"`abcd capture disposition " + c.Item + "` refuses until it is untied",
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
