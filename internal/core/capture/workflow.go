package capture

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/Partnermedia/abcd/internal/core/changelog"
	"github.com/Partnermedia/abcd/internal/fsutil"
)

// mutationPreamble runs the idempotent pre-mutation steps: sweep orphan
// placeholders and (re-)assert the symlink-refused directory shape. Read-only
// entry points (List, Status) deliberately skip it.
//
// The orphan sweep runs UNDER the ledger lock (iss-102): a >60s-stalled capture
// fills its reserved placeholder via commitCapture, which now also holds the
// ledger lock, so the sweep's classify-then-unlink can no longer interleave with
// a commit's fill and delete a just-committed issue file.
func mutationPreamble(issuesRoot string) error {
	if err := withLedgerLock(issuesRoot, func() error {
		return cleanOrphanPlaceholders(issuesRoot)
	}); err != nil {
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

	// The mint is timestamp-numeric (adr-45; mechanics per spc-33): it consults
	// no maximum, so the refs-union scan the max+1 allocator needed (iss-115,
	// iss-120) is gone — time orders the ids and entropy separates same-second
	// minters on branches this working tree cannot see.
	issID, placeholder, err := reservePath(issuesRoot, slugNorm, req.ForceID)
	if err != nil {
		return CaptureResult{}, err
	}

	result, err := commitCapture(issuesRoot, req, issID, slugNorm, placeholder)
	if err != nil {
		_ = cancelReservation(placeholder)
		return CaptureResult{}, err
	}
	// Machine output carries a repo-relative locator, never an absolute
	// developer-identity path (iss-81).
	result.Path = fsutil.RepoRel(repoRoot, result.Path)
	return result, nil
}

func commitCapture(issuesRoot string, req CaptureRequest, issID, slug, placeholder string) (CaptureResult, error) {
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

	// The checksum re-read + fill runs UNDER the ledger lock (iss-102): the orphan
	// sweep (mutationPreamble) also holds this lock, so a >60s-stalled commit can no
	// longer land its fill in the sweep's classify-then-unlink window and have the
	// just-committed file deleted. If the sweep reclaimed the placeholder first, the
	// re-read fails and the capture reports an error rather than a false success.
	var result CaptureResult
	err = withLedgerLock(issuesRoot, func() error {
		// Guard the overwrite: the placeholder must still be the zero-byte file we
		// reserved (expected_checksum = sha256("")).
		_, checksum, rerr := readWithChecksum(placeholder)
		if rerr != nil {
			return rerr
		}
		if checksum != emptyChecksum {
			return fmt.Errorf("%w: placeholder %s changed since reservation", ErrChecksumMismatch, placeholder)
		}
		if werr := fsutil.WriteFileAtomicPreserveMode(placeholder, []byte(content)); werr != nil {
			return werr
		}
		result = CaptureResult{ID: issID, Slug: slug, Path: placeholder, Status: StateOpen}
		return nil
	})
	if err != nil {
		return CaptureResult{}, err
	}
	return result, nil
}

// Resolve moves an open issue to resolved/, writing the resolution note and the
// product impact. resolved/ is gated by issue_impact_valid, so the impact is
// validated against the shared changelog enum up front (empty or invalid is
// refused, never defaulted) and stamped bare alongside the note — the tool's own
// resolve path can never mint a record its own blocker rejects.
//
// The optional ByIntent/BySpec/ByCommit members land as the structured
// resolved_by object (spc-25) in the same atomic transition. Ids are validated
// for existence in their record store before anything is written; the sha is
// shape-checked only (its home may be the remote, and shallow or rebased
// states must not refuse a legitimate resolution). No members → no
// resolved_by key at all: provenance is optional, never defaulted.
func Resolve(req ResolveRequest) (TransitionResult, error) {
	impact, err := changelog.ParseImpact(req.Impact)
	if err != nil {
		return TransitionResult{}, fmt.Errorf("resolve: %w", err)
	}
	rb, err := resolveProvenance(req)
	if err != nil {
		return TransitionResult{}, err
	}
	extras := []kv{{"impact", rawScalar(string(impact))}}
	if rb != nil {
		var members nested
		if rb.Intent != "" {
			members = append(members, kv{"intent", rb.Intent})
		}
		if rb.Spec != "" {
			members = append(members, kv{"spec", rb.Spec})
		}
		if rb.Commit != "" {
			members = append(members, kv{"commit", rb.Commit})
		}
		extras = append(extras, kv{"resolved_by", members})
	}
	res, err := transition(req.RepoRoot, req.IssuesRoot, req.ID, "resolution", req.Resolution,
		extras, StateResolved)
	if err != nil {
		return TransitionResult{}, err
	}
	res.ResolvedBy = rb
	return res, nil
}

// resolveProvenance validates the resolved_by members before any write: id
// shape and store existence for --intent/--spec (via the shared findRecordFile
// probe promote's link mode also uses), shape only for --commit. Returns nil
// when no member was supplied.
func resolveProvenance(req ResolveRequest) (*ResolvedBy, error) {
	if req.ByIntent == "" && req.BySpec == "" && req.ByCommit == "" {
		return nil, nil
	}
	repoRoot, _, err := resolveRoots(req.RepoRoot, req.IssuesRoot)
	if err != nil {
		return nil, err
	}
	if req.ByIntent != "" {
		if !reItdID.MatchString(req.ByIntent) {
			return nil, fmt.Errorf("resolve: --intent %q does not match ^itd-[0-9]+$; nothing written", req.ByIntent)
		}
		if _, ok := findRecordFile(repoRoot, intentStoreRelDirs(), req.ByIntent); !ok {
			return nil, fmt.Errorf("resolve: --intent %s not found in the intent store; nothing written", req.ByIntent)
		}
	}
	if req.BySpec != "" {
		if !reSpcID.MatchString(req.BySpec) {
			return nil, fmt.Errorf("resolve: --spec %q does not match ^spc-[0-9]+$; nothing written", req.BySpec)
		}
		if _, ok := findRecordFile(repoRoot, specStoreRelDirs(), req.BySpec); !ok {
			return nil, fmt.Errorf("resolve: --spec %s not found in the spec store; nothing written", req.BySpec)
		}
	}
	if req.ByCommit != "" && !reCommitSha.MatchString(req.ByCommit) {
		return nil, fmt.Errorf("resolve: --commit %q is not a 7-64 char lowercase hex sha; nothing written", req.ByCommit)
	}
	return &ResolvedBy{Intent: req.ByIntent, Spec: req.BySpec, Commit: req.ByCommit}, nil
}

// Wontfix moves an open issue to wontfix/, writing the wontfix_reason note.
// wontfix/ carries no impact (issue_impact_valid gates resolved/ only), so no
// judgement is stamped.
func Wontfix(req WontfixRequest) (TransitionResult, error) {
	return transition(req.RepoRoot, req.IssuesRoot, req.ID, "wontfix_reason", req.Reason, nil, StateWontfix)
}

// transition moves an open issue to target, setting the defining note field and
// any extra frontmatter fields (e.g. resolved/'s impact) in one atomic write.
func transition(repoRoot, issuesRoot, issID, field, note string, extra []kv, target State) (TransitionResult, error) {
	rr, ir, err := resolveRoots(repoRoot, issuesRoot)
	if err != nil {
		return TransitionResult{}, err
	}
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
		for _, f := range extra {
			if members, isNested := f.val.(nested); isNested {
				newContent, err = setMapField(newContent, f.key, members)
			} else {
				newContent, err = setScalarField(newContent, f.key, f.val)
			}
			if err != nil {
				return err
			}
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
	// Machine output carries a repo-relative locator, never an absolute
	// developer-identity path (iss-81).
	result.Path = fsutil.RepoRel(rr, result.Path)
	return result, nil
}

// removeSourceHook, when non-nil, replaces os.Remove(src) inside
// commitTransition. It is a test-only seam (nil in production, zero overhead)
// used to force a deterministic non-ENOENT remove failure without relying on
// platform-specific filesystem tricks (immutable attributes, read-only
// remounts) that a portable test cannot set up.
var removeSourceHook func(path string) error

// commitTransition re-verifies the source's checksum, writes the destination
// atomically, then removes the source. A concurrent edit surfaces as a
// checksum mismatch or transition conflict rather than a lost update.
//
// A non-ENOENT remove failure (iss-186: EPERM/EROFS/EIO on the source status
// dir) rolls the destination back rather than leaving the issue split across
// both directories — findIssue rejects any id present in more than one status
// dir, so a stranded copy could never again be transitioned without manual
// repair. Rolling back restores the pre-call state (src present, dst absent)
// so the caller can simply retry once the underlying failure clears.
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
	if err := fsutil.WriteFileAtomicPreserveMode(dst, []byte(newContent)); err != nil {
		return err
	}
	removeSrc := os.Remove
	if removeSourceHook != nil {
		removeSrc = removeSourceHook
	}
	if err := removeSrc(src); err != nil && !os.IsNotExist(err) {
		if rbErr := os.Remove(dst); rbErr != nil && !os.IsNotExist(rbErr) {
			return fmt.Errorf("%w (rollback of %s also failed: %v)", err, dst, rbErr)
		}
		return err
	}
	return nil
}

// List scans one state (or all) and returns issues sorted ascending by numeric
// N plus a roster of unparseable files. Read-only: no preamble, no dir
// creation, tolerant of a virgin/absent ledger.
func List(req ListRequest) (ListResult, error) {
	repoRoot, ir, err := resolveRoots(req.RepoRoot, req.IssuesRoot)
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
	relativiseLedgerPaths(repoRoot, issues, skipped)
	return ListResult{Issues: issues, Skipped: skipped}, nil
}

// relativiseLedgerPaths rewrites every ledger Path in place to a repo-relative
// locator, so no --json envelope echoes an absolute developer-identity path
// (iss-81). It covers the structured Path field; a SkipRecord.Error string that
// wraps an os.ReadFile failure can still carry an absolute path, which the CLI
// text surface sanitises via termsafe but the --json surface does not — tracked
// separately from this locator fix.
func relativiseLedgerPaths(repoRoot string, issues []Issue, skipped []SkipRecord) {
	for i := range issues {
		issues[i].Path = fsutil.RepoRel(repoRoot, issues[i].Path)
	}
	for i := range skipped {
		skipped[i].Path = fsutil.RepoRel(repoRoot, skipped[i].Path)
	}
}

// Status is a pure read: counts per status dir plus up to 10 most-recent open
// issues (newest first). Guaranteed no mutation.
func Status(req StatusRequest) (StatusResult, error) {
	repoRoot, ir, err := resolveRoots(req.RepoRoot, req.IssuesRoot)
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
	if open == nil {
		open = []Issue{}
	}
	res.RecentOpen = open
	relativiseLedgerPaths(repoRoot, res.RecentOpen, res.Skipped)
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
