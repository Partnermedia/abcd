package capture

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

// mutationPreamble runs the idempotent pre-mutation steps: sweep orphan
// placeholders and (re-)assert the symlink-refused directory shape. Read-only
// entry points (List, Status) deliberately skip it.
func mutationPreamble(issuesRoot string) error {
	if err := cleanOrphanPlaceholders(issuesRoot); err != nil {
		return err
	}
	return ensureLedgerDirs(issuesRoot)
}

// Capture appends a new issue to open/ with an auto-assigned (or forced) iss-N.
// The write is transactional: a zero-byte placeholder is reserved, and on any
// failure it is swept. Returns the committed path under open/.
func Capture(req CaptureRequest) (CaptureResult, error) {
	repoRoot, issuesRoot, err := resolveRoots(req.RepoRoot, req.IssuesRoot)
	if err != nil {
		return CaptureResult{}, err
	}
	_ = repoRoot
	if err := mutationPreamble(issuesRoot); err != nil {
		return CaptureResult{}, err
	}

	if strings.TrimSpace(req.FoundDuring) == "" {
		return CaptureResult{}, fmt.Errorf("found_during must be a non-empty string")
	}
	slugNorm, err := normaliseSlug(req.Slug)
	if err != nil {
		return CaptureResult{}, err
	}

	issID, placeholder, err := reservePath(issuesRoot, slugNorm, req.ForceID)
	if err != nil {
		return CaptureResult{}, err
	}

	result, err := commitCapture(req, issID, slugNorm, placeholder)
	if err != nil {
		_ = cancelReservation(placeholder)
		return CaptureResult{}, err
	}
	return result, nil
}

func commitCapture(req CaptureRequest, issID, slug, placeholder string) (CaptureResult, error) {
	fields := []kv{
		{"schema_version", 1},
		{"id", issID},
		{"slug", slug},
		{"severity", string(req.Severity)},
		{"category", string(req.Category)},
		{"source", string(req.Source)},
		{"found_during", req.FoundDuring},
	}
	fm := map[string]any{
		"schema_version": 1,
		"id":             issID,
		"slug":           slug,
		"severity":       string(req.Severity),
		"category":       string(req.Category),
		"source":         string(req.Source),
		"found_during":   req.FoundDuring,
	}
	if req.FoundAt != "" {
		fields = append(fields, kv{"found_at", req.FoundAt})
		fm["found_at"] = req.FoundAt
	}
	if req.RelatedIntents != nil {
		fields = append(fields, kv{"related_intents", req.RelatedIntents})
		fm["related_intents"] = req.RelatedIntents
	}
	if req.RelatedSpecs != nil {
		fields = append(fields, kv{"related_specs", req.RelatedSpecs})
		fm["related_specs"] = req.RelatedSpecs
	}
	if req.BlockedBy != nil {
		fields = append(fields, kv{"blocked_by", req.BlockedBy})
		fm["blocked_by"] = req.BlockedBy
	}

	if err := validateStrict(fm); err != nil {
		return CaptureResult{}, err
	}
	if err := validateInvariants(fm, StateOpen, placeholder); err != nil {
		return CaptureResult{}, err
	}

	content, err := buildIssueText(fields, req.Text)
	if err != nil {
		return CaptureResult{}, err
	}

	// Guard the overwrite: the placeholder must still be the zero-byte file we
	// reserved (expected_checksum = sha256("")).
	_, checksum, err := readWithChecksum(placeholder)
	if err != nil {
		return CaptureResult{}, err
	}
	if checksum != emptyChecksum {
		return CaptureResult{}, fmt.Errorf("%w: placeholder %s changed since reservation", ErrChecksumMismatch, placeholder)
	}
	if err := writeFileAtomic(placeholder, []byte(content)); err != nil {
		return CaptureResult{}, err
	}
	return CaptureResult{ID: issID, Slug: slug, Path: placeholder, Status: StateOpen}, nil
}

// Resolve moves an open issue to resolved/, writing the resolution note.
func Resolve(req ResolveRequest) (TransitionResult, error) {
	return transition(req.RepoRoot, req.IssuesRoot, req.ID, "resolution", req.Resolution, StateResolved)
}

// Wontfix moves an open issue to wontfix/, writing the wontfix_reason note.
func Wontfix(req WontfixRequest) (TransitionResult, error) {
	return transition(req.RepoRoot, req.IssuesRoot, req.ID, "wontfix_reason", req.Reason, StateWontfix)
}

func transition(repoRoot, issuesRoot, issID, field, note string, target State) (TransitionResult, error) {
	rr, ir, err := resolveRoots(repoRoot, issuesRoot)
	if err != nil {
		return TransitionResult{}, err
	}
	_ = rr
	if err := mutationPreamble(ir); err != nil {
		return TransitionResult{}, err
	}
	if !reIssID.MatchString(issID) {
		return TransitionResult{}, fmt.Errorf("invalid iss-N identifier: %q", issID)
	}
	if strings.TrimSpace(note) == "" {
		return TransitionResult{}, fmt.Errorf("%s must be a non-empty string", field)
	}

	// The find→read→move critical section runs under the ledger lock, the SAME
	// lock id allocation takes, so two concurrent conflicting transitions (a
	// resolve and a wontfix on one issue) serialize: the second sees the issue
	// already moved out of open/ and conflicts, instead of both passing the
	// checksum re-read and landing the issue in two status dirs (split-brain).
	var result TransitionResult
	err = withLedgerLock(ir, func() error {
		src, status, err := findIssue(ir, issID)
		if err != nil {
			return err
		}
		if status != StateOpen {
			return fmt.Errorf("%w: %s already in %s", ErrTransitionConflict, issID, status)
		}

		content, checksum, err := readWithChecksum(src)
		if err != nil {
			return err
		}
		newContent, err := setScalarField(content, field, note)
		if err != nil {
			return err
		}

		dst := filepath.Join(ir, statusDirName[target], filepath.Base(src))
		fm, _, err := parseFrontmatterAndBody(newContent)
		if err != nil {
			return err
		}
		if err := validateStrict(fm); err != nil {
			return err
		}
		if err := validateInvariants(fm, target, dst); err != nil {
			return err
		}

		if err := commitTransition(src, dst, newContent, checksum); err != nil {
			return err
		}
		result = TransitionResult{ID: issID, Path: dst, FromStatus: StateOpen, ToStatus: target}
		return nil
	})
	if err != nil {
		return TransitionResult{}, err
	}
	return result, nil
}

// commitTransition re-verifies the source's checksum, writes the destination
// atomically, then removes the source. A concurrent edit surfaces as a
// checksum mismatch or transition conflict rather than a lost update.
func commitTransition(src, dst, newContent, expected string) error {
	_, current, err := readWithChecksum(src)
	if os.IsNotExist(err) {
		return fmt.Errorf("%w: %s move source missing", ErrTransitionConflict, src)
	}
	if err != nil {
		return err
	}
	if current != expected {
		return fmt.Errorf("%w: %s changed since it was read", ErrChecksumMismatch, src)
	}
	if err := writeFileAtomic(dst, []byte(newContent)); err != nil {
		return err
	}
	if err := os.Remove(src); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

// List scans one state (or all) and returns issues sorted ascending by numeric
// N plus a roster of unparseable files. Read-only: no preamble, no dir
// creation, tolerant of a virgin/absent ledger.
func List(req ListRequest) (ListResult, error) {
	_, ir, err := resolveRoots(req.RepoRoot, req.IssuesRoot)
	if err != nil {
		return ListResult{}, err
	}
	state := req.State
	if state == "" {
		state = StateAll
	}
	if state != StateAll && state != StateOpen && state != StateResolved && state != StateWontfix {
		return ListResult{}, fmt.Errorf("state must be all/open/resolved/wontfix, got %q", state)
	}
	issues, skipped := scanLedger(ir, state)
	sortIssues(issues)
	prioritise(issues, openIDSet(ir))
	return ListResult{Issues: issues, Skipped: skipped}, nil
}

// Status is a pure read: counts per status dir plus up to 10 most-recent open
// issues (newest first). Guaranteed no mutation.
func Status(req StatusRequest) (StatusResult, error) {
	_, ir, err := resolveRoots(req.RepoRoot, req.IssuesRoot)
	if err != nil {
		return StatusResult{}, err
	}
	var res StatusResult
	open, skOpen := scanLedger(ir, StateOpen)
	resolved, skRes := scanLedger(ir, StateResolved)
	wontfix, skWf := scanLedger(ir, StateWontfix)
	res.OpenCount = len(open)
	res.ResolvedCount = len(resolved)
	res.WontfixCount = len(wontfix)
	res.Skipped = append(append(append([]SkipRecord{}, skOpen...), skRes...), skWf...)

	openIDs := idSet(open)
	// Newest first: higher N is newer (ids are monotonic with creation).
	sort.SliceStable(open, func(i, j int) bool { return issNumber(open[i].ID) > issNumber(open[j].ID) })
	if len(open) > 10 {
		open = open[:10]
	}
	// Derived-priority view over the recent slice: unblocked first, then severity.
	prioritise(open, openIDs)
	res.RecentOpen = open
	return res, nil
}

// severityRank orders severities for the derived-priority view: higher rank
// sorts earlier (critical is most urgent, nitpick least).
var severityRank = map[Severity]int{
	SeverityCritical: 4, SeverityMajor: 3, SeverityMinor: 2, SeverityNitpick: 1,
}

// prioritise applies the read-time priority projection in place: it fills each
// issue's BlockedByOpen with the blocked_by targets still in open/ (openIDs),
// then stably orders unblocked issues ahead of blocked ones and, within each
// group, higher severity first. There is no stored priority — this is a derived
// view so the CLI and any future front door share one ordering. The caller is
// expected to have pre-sorted issues into a deterministic tiebreak order.
func prioritise(issues []Issue, openIDs map[string]bool) {
	for i := range issues {
		var stillOpen []string
		for _, dep := range issues[i].BlockedBy {
			if openIDs[dep] {
				stillOpen = append(stillOpen, dep)
			}
		}
		issues[i].BlockedByOpen = stillOpen
	}
	sort.SliceStable(issues, func(i, j int) bool {
		bi, bj := len(issues[i].BlockedByOpen) > 0, len(issues[j].BlockedByOpen) > 0
		if bi != bj {
			return !bi // unblocked (false) sorts first
		}
		return severityRank[issues[i].Severity] > severityRank[issues[j].Severity]
	})
}

// openIDSet returns the set of ids currently in open/ — the predicate a
// blocked_by target must satisfy to still count as blocking. Read-only.
func openIDSet(issuesRoot string) map[string]bool {
	open, _ := scanLedger(issuesRoot, StateOpen)
	return idSet(open)
}

// idSet collects the ids of a slice of issues into a set.
func idSet(issues []Issue) map[string]bool {
	set := make(map[string]bool, len(issues))
	for _, iss := range issues {
		set[iss.ID] = true
	}
	return set
}

// scanLedger reads issues from the requested state(s). Stray/non-matching .md
// files are silently ignored; corrupt matching files go into Skipped.
func scanLedger(issuesRoot string, state State) ([]Issue, []SkipRecord) {
	var targets []State
	if state == StateAll {
		targets = statusDirs[:]
	} else {
		targets = []State{state}
	}
	var issues []Issue
	var skipped []SkipRecord
	for _, sub := range targets {
		dir := filepath.Join(issuesRoot, statusDirName[sub])
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue // virgin/absent ledger tolerance
		}
		names := make([]string, 0, len(entries))
		for _, e := range entries {
			names = append(names, e.Name())
		}
		sort.Strings(names)
		for _, name := range names {
			if filepath.Ext(name) != ".md" || !reFilenameID.MatchString(name) {
				continue // stray .md (README, etc.) silently ignored
			}
			path := filepath.Join(dir, name)
			data, err := os.ReadFile(path)
			if err != nil {
				skipped = append(skipped, SkipRecord{Path: path, Error: err.Error()})
				continue
			}
			fm, body, err := parseFrontmatterAndBody(string(data))
			if err != nil {
				skipped = append(skipped, SkipRecord{Path: path, Error: err.Error()})
				continue
			}
			if err := validateStrict(fm); err != nil {
				skipped = append(skipped, SkipRecord{Path: path, Error: err.Error()})
				continue
			}
			if err := validateInvariants(fm, sub, path); err != nil {
				skipped = append(skipped, SkipRecord{Path: path, Error: err.Error()})
				continue
			}
			issues = append(issues, issueFromFrontmatter(fm, sub, path, body))
		}
	}
	return issues, skipped
}

// sortIssues orders issues ascending by numeric N.
func sortIssues(issues []Issue) {
	sort.SliceStable(issues, func(i, j int) bool {
		return issNumber(issues[i].ID) < issNumber(issues[j].ID)
	})
}

// issNumber extracts N from an iss-N reference; -1 for non-matching input.
func issNumber(s string) int {
	m := reSortIssID.FindStringSubmatch(s)
	if m == nil {
		return -1
	}
	n, err := strconv.Atoi(m[1])
	if err != nil {
		return -1
	}
	return n
}
