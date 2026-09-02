package reading

// candidates.go derives the widening run a comparative reading characterises,
// and holds the two refusals that derivation carries (adr-2609021016272867;
// spc-2609020626039834, "The derived run" and "The ordering guard").
//
// No operand names the run. The design fixes the invocation at a position and a
// target state (framework v4 section 8.2 and ruling M8; companion v4 section
// 4.1), brief invariant 15's operand enumeration names two, and
// adr-2609021016286571 restores that letter — so the run comes from the record.
// The rule is the ADR's: the one committed widening run at the target whose
// items carry no disposition and no admission. None, or more than one, refuses
// and lists the widening runs at the target, because the operator's next act —
// dispositioning one run's items — is what the design sequences after the
// comparative reading anyway, and they can only perform it if they are told what
// there is to disposition.

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/intentdriven/abcd/internal/core/capture"
	"github.com/intentdriven/abcd/internal/core/issueschema"
	"github.com/intentdriven/abcd/internal/core/recordid"
	"github.com/intentdriven/abcd/internal/fsutil"
)

// WideningRun is one widening run at a target, as the derivation sees it: what
// it is, how much it holds, and whether anything has happened to what it holds.
//
// Every field is on the LISTING a refusal renders, because the operator's next
// move depends on all of them: a run that did not qualify is shown rather than
// hidden, with the reason it did not.
type WideningRun struct {
	// ID is the run identifier (rdg-N).
	ID string `json:"id"`
	// Items is how many reading records the run holds in the ledger.
	Items int `json:"items"`
	// Dispositioned reports that at least one item carries a standing
	// disposition; Admitted, that at least one carries an admission. Either is a
	// fate, and a candidate set is defined as pre-admission.
	Dispositioned bool `json:"dispositioned"`
	Admitted      bool `json:"admitted"`
	// Committed reports that the run reached its commit marker
	// (`run.json`). A run without one never happened and its records are the
	// next ingest sweep's to roll back, so it is listed and never selected.
	Committed bool `json:"committed"`
	// Fated names what stands over the items that carry a fate, one rendered
	// phrase per item, so a refusal names THE ITEM AND THE DISPOSITION rather
	// than only the run. The two booleans above answer "did this run qualify";
	// this answers "which item stopped it", which is what the ordering guard
	// owes an operator (spc-2609020626039834, "The ordering guard").
	Fated []string `json:"fated,omitempty"`
}

// Qualifies reports whether this run is the candidate set the ADR describes: a
// committed widening run holding items, none of which carries a fate.
func (r WideningRun) Qualifies() bool {
	return r.Committed && r.Items > 0 && !r.Dispositioned && !r.Admitted && len(r.Fated) == 0
}

// Line renders one run for the refusal listing: one run per line, saying what it
// holds and why it did or did not qualify.
func (r WideningRun) Line() string {
	var b strings.Builder
	fmt.Fprintf(&b, "%s — %d item(s)", r.ID, r.Items)
	if !r.Committed {
		b.WriteString(", NOT committed (no ")
		b.WriteString(issueschema.RunRecordFileName)
		b.WriteString(" commit marker, so the run never happened)")
	}
	switch {
	case len(r.Fated) > 0:
		b.WriteString(", ")
		b.WriteString(strings.Join(r.Fated, "; "))
	case r.Items == 0:
		b.WriteString(", nothing to characterise")
	case r.Committed:
		b.WriteString(", no disposition and no admission")
	}
	return b.String()
}

// renderRuns renders the whole listing, one run per line, or says there is none.
func renderRuns(runs []WideningRun) string {
	if len(runs) == 0 {
		return "  (no widening run is committed at this target)"
	}
	lines := make([]string, 0, len(runs))
	for _, r := range runs {
		lines = append(lines, "  "+r.Line())
	}
	return strings.Join(lines, "\n")
}

// NoCandidateRun is the refusal when no widening run at the target qualifies. It
// carries the whole listing, because the remedy depends on which of the reasons
// applies: no run at all, a run that never committed, a run holding nothing, or
// a run whose items already carry a fate.
type NoCandidateRun struct {
	Target string
	Runs   []WideningRun
}

func (e *NoCandidateRun) Error() string {
	return fmt.Sprintf("the %s assembly derives its candidate set from the record, and no committed "+
		"widening run at %s has every item free of a disposition and an admission, so there is "+
		"nothing to characterise. The candidate set is defined as PRE-ADMISSION: characterisation "+
		"precedes admission, and a candidate whose fate is already recorded is not one "+
		"(adr-2609021016272867). The widening runs at this target:\n%s",
		PositionComparative, shortTarget(e.Target), renderRuns(e.Runs))
}

// AmbiguousCandidateRun is the refusal when more than one qualifies. The remedy
// is stated because the design already sequences it: dispositioning one run's
// items is the act that follows a comparative reading, and performing it makes
// the selection unambiguous.
type AmbiguousCandidateRun struct {
	Target string
	Runs   []WideningRun
}

func (e *AmbiguousCandidateRun) Error() string {
	qualifying := make([]string, 0, len(e.Runs))
	for _, r := range e.Runs {
		if r.Qualifies() {
			qualifying = append(qualifying, r.ID)
		}
	}
	return fmt.Sprintf("the %s assembly derives its candidate set from the record, and %d widening "+
		"runs at %s have every item free of a disposition and an admission (%s); the invocation is "+
		"a position and a target state, so nothing names which. Disposition one run's items — the "+
		"act the design places after the comparative reading in any case — and the selection is "+
		"unambiguous. The widening runs at this target:\n%s",
		PositionComparative, len(qualifying), shortTarget(e.Target),
		strings.Join(qualifying, ", "), renderRuns(e.Runs))
}

// shortTarget abbreviates a commit for a refusal message.
func shortTarget(sha string) string {
	if len(sha) > 12 {
		return sha[:12]
	}
	return sha
}

// WideningRuns lists every widening run at the target, committed or not, with
// its item count and the fate of its items.
//
// It reads the DURABLE run family for what a run is — its position, its target,
// and whether it committed — and the LEDGER for what it holds. Those are the two
// places the two facts actually live: a run's identity is in
// `.abcd/development/readings/<run>/`, and its items are reading records under
// the issue ledger.
//
// A directory carrying neither a run record nor a manifest is not a run and is
// skipped: it says nothing about a position or a target, so listing it would
// name a run this function cannot describe.
func WideningRuns(repoRoot, target string) ([]WideningRun, error) {
	runsRoot := filepath.Join(repoRoot, filepath.FromSlash(issueschema.ReadingsRecordDir))
	if err := refuseSymlinkedRunDir(runsRoot); err != nil {
		return nil, err
	}
	entries, err := os.ReadDir(runsRoot)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("reading: listing the committed runs: %w", err)
	}
	names := make([]string, 0, len(entries))
	for _, e := range entries {
		if !e.IsDir() || !recordid.ValidReadingRunID(e.Name()) {
			continue
		}
		names = append(names, e.Name())
	}
	sort.Strings(names)

	var out []WideningRun
	for _, name := range names {
		dir := filepath.Join(runsRoot, name)
		if err := refuseSymlinkedRunDir(dir); err != nil {
			return nil, err
		}
		head, committed, err := runHeadOf(dir)
		if err != nil {
			return nil, err
		}
		if head.position != string(PositionWidening) || head.target != target {
			continue
		}
		run := WideningRun{ID: name, Committed: committed}
		items, err := ledgerItemsOf(repoRoot, name)
		if err != nil {
			return nil, err
		}
		run.Items = len(items)
		for _, item := range items {
			fate, err := capture.ItemFate(repoRoot, name, item)
			if err != nil {
				return nil, fmt.Errorf("reading: reading the fate of %s in run %s: %w", item, name, err)
			}
			switch {
			case fate.Cyclic:
				run.Dispositioned = true
				run.Fated = append(run.Fated, fmt.Sprintf(
					"%s carries a supersession cycle, so its fate cannot be read and it is not "+
						"demonstrably uncommitted", item))
			case fate.Contested():
				run.Dispositioned = true
				run.Fated = append(run.Fated, fmt.Sprintf(
					"%s carries %d standing dispositions (%s), so no reader may say which is in force",
					item, len(fate.Dispositions), strings.Join(fate.Dispositions, ", ")))
			case len(fate.Dispositions) > 0:
				run.Dispositioned = true
				run.Fated = append(run.Fated, fmt.Sprintf(
					"%s carries the standing disposition %s", item, strings.Join(fate.Dispositions, ", ")))
			}
			if len(fate.Admissions) > 0 {
				run.Admitted = true
				run.Fated = append(run.Fated, fmt.Sprintf(
					"%s carries the admission %s", item, strings.Join(fate.Admissions, ", ")))
			}
		}
		out = append(out, run)
	}
	return out, nil
}

// DeriveCandidateRun applies the ADR's rule over the runs at the target: exactly
// one must qualify.
//
// The two refusals carry the whole listing rather than only what went wrong,
// because the operator resolves either of them by looking at the runs: with none
// they need to see whether a run is uncommitted, empty or already dispositioned,
// and with more than one they need to see which to disposition.
func DeriveCandidateRun(repoRoot, target string) (WideningRun, error) {
	runs, err := WideningRuns(repoRoot, target)
	if err != nil {
		return WideningRun{}, err
	}
	var qualifying []WideningRun
	for _, r := range runs {
		if r.Qualifies() {
			qualifying = append(qualifying, r)
		}
	}
	switch len(qualifying) {
	case 1:
		return qualifying[0], nil
	case 0:
		return WideningRun{}, &NoCandidateRun{Target: target, Runs: runs}
	default:
		return WideningRun{}, &AmbiguousCandidateRun{Target: target, Runs: runs}
	}
}

// runHead is the subset of a run's own artefacts the derivation reads: which
// position it read at, and which commit it read.
type runHead struct {
	position string
	target   string
}

// runHeadOf reads one run directory's head, and reports whether the run
// COMMITTED. The run record is the commit marker and is read first; a directory
// without one is described from its parked manifest instead, so an uncommitted
// run is listed with its position and target rather than silently absent.
func runHeadOf(dir string) (runHead, bool, error) {
	var record struct {
		Position     string `json:"position"`
		TargetCommit string `json:"target_commit"`
	}
	err := readRunJSON(filepath.Join(dir, issueschema.RunRecordFileName), &record)
	switch {
	case err == nil:
		return runHead{position: record.Position, target: record.TargetCommit}, true, nil
	case !os.IsNotExist(err):
		return runHead{}, false, err
	}
	var parked struct {
		Position     string `json:"position"`
		TargetCommit string `json:"target_commit"`
	}
	if err := readRunJSON(filepath.Join(dir, ManifestFileName), &parked); err != nil {
		if os.IsNotExist(err) {
			return runHead{}, false, nil
		}
		return runHead{}, false, err
	}
	return runHead{position: parked.Position, target: parked.TargetCommit}, false, nil
}

// readRunJSON decodes one of a run's own artefacts, leaving os.ErrNotExist
// unwrapped so the caller can tell "absent" from "unreadable".
func readRunJSON(path string, into any) error {
	raw, err := fsutil.ReadGuarded(path, MaxFileBytes)
	if err != nil {
		if os.IsNotExist(err) {
			return err
		}
		return fmt.Errorf("reading: %s: %w", path, err)
	}
	if err := json.Unmarshal(raw, into); err != nil {
		return fmt.Errorf("reading: %s does not decode: %w", path, err)
	}
	return nil
}

// ledgerItemsOf lists the reading-record ids one run holds in the ledger, sorted.
// An absent directory is a run with no items, which is a state and not a fault.
func ledgerItemsOf(repoRoot, run string) ([]string, error) {
	dir := filepath.Join(repoRoot, filepath.FromSlash(CandidateSource), run)
	if err := refuseSymlinkedRunDir(filepath.Dir(dir)); err != nil {
		return nil, err
	}
	if err := refuseSymlinkedRunDir(dir); err != nil {
		return nil, err
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("reading: listing the items of run %s: %w", run, err)
	}
	var out []string
	for _, e := range entries {
		id, ok := strings.CutSuffix(e.Name(), ".md")
		if e.IsDir() || !ok || !recordid.ValidReadingItemID(id) {
			continue
		}
		out = append(out, id)
	}
	sort.Strings(out)
	return out, nil
}

// refuseSymlinkedRunDir refuses a path that exists and is not a real directory.
// It is the same stance core/capture takes on the same ledger tree: a run's
// directory is a run's directory, never a pointer at somebody else's — and here
// the consequence is what a reading is HANDED, so following one would let a link
// decide the candidate set.
func refuseSymlinkedRunDir(dir string) error {
	fi, err := os.Lstat(dir)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("reading: %s could not be examined: %w", dir, err)
	}
	if fi.Mode()&os.ModeSymlink != 0 || !fi.IsDir() {
		return fmt.Errorf("reading: %s is not a real directory (a symlink is not followed); the "+
			"candidate set is decided by the record, never by a link", dir)
	}
	return nil
}

// refuseNonWideningCandidate is the record-level check on every file the
// candidate row enumerates: it must be a well-formed reading record, and it must
// have been returned at the WIDENING position.
//
// A run at any other position is not a candidate set — the design fixes the
// comparative object as the widening reading's returned configurations and
// admits no other source (companion 7.2, R4) — and the row's own path cannot
// establish that, because every position's records live in the same family.
func refuseNonWideningCandidate(rel string, raw []byte) error {
	fm, err := capture.ValidateReadingRecord(string(raw))
	if err != nil {
		return fmt.Errorf("reading: %s is admitted as a candidate and is not a well-formed reading "+
			"record: %w", rel, err)
	}
	if got, _ := fm["position"].(string); got != string(PositionWidening) {
		return fmt.Errorf("reading: %s was returned at the %s position, and a candidate set is one "+
			"WIDENING run's returned configurations; a run at any other position is not a candidate "+
			"set (adr-2609021016272867)", rel, got)
	}
	return nil
}

// assertCandidateProjection is the fail-closed half of the readings-store
// exclusion the manifest declares at the comparative position.
//
// The directory rows of the floor are enforced by path, and a path needs no
// parse; the readings row is a SIGNAL row instead, because a directory row there
// would be false — one run's items do travel. So what the manifest asserts is
// the narrower promise, and this is what makes it a promise rather than a
// disclosure: no candidate item may be emitted from outside the derived run's
// directory, or under a field outside the two projected, and nothing else may be
// emitted from the readings store at all (adr-56; brief invariant 16).
func assertCandidateProjection(cands []candidate, run string) error {
	prefix := CandidateSource + "/" + run + "/"
	fields := map[string]bool{}
	for _, f := range CandidateFields {
		fields[f] = true
	}
	for _, c := range cands {
		underStore := strings.HasPrefix(c.path, CandidateSource+"/")
		if c.kind != KindCandidate {
			if underStore {
				return fmt.Errorf("reading: item %s comes from the readings store as a %s item; the "+
					"only thing that store supplies is the derived run's candidates, which the "+
					"manifest asserts", c.path, c.kind)
			}
			continue
		}
		if !strings.HasPrefix(c.path, prefix) {
			return fmt.Errorf("reading: candidate item %s lies outside the derived run %s, whose "+
				"exclusion the manifest asserts: every run other than the candidate run stays behind",
				c.path, run)
		}
		if !fields[c.field] {
			return fmt.Errorf("reading: candidate item %s projects the field %q, and the manifest "+
				"asserts that every field of the named run's items other than %s stays behind",
				c.path, c.field, strings.Join(CandidateFields, " and "))
		}
	}
	return nil
}

// candidateItems reports how many DISTINCT reading records the collected
// candidates came from, so the count the manifest records as `candidates` can be
// checked against what the row actually enumerated.
func candidateItems(cands []candidate) []string {
	seen := map[string]bool{}
	var out []string
	for _, c := range cands {
		if c.kind != KindCandidate || seen[c.path] {
			continue
		}
		seen[c.path] = true
		out = append(out, candidateIDOf(c.path))
	}
	sort.Strings(out)
	return out
}

// candidateIDOf is the item id a candidate's path names. The record's file is
// named for its id, which is what lets the bundle tell a reading which rdi-N
// each text belongs to without carrying the path (brief invariant 15).
func candidateIDOf(rel string) string {
	return strings.TrimSuffix(filepath.Base(rel), ".md")
}
