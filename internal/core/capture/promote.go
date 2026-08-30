package capture

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/intentdriven/abcd/internal/core/intent"
	"github.com/intentdriven/abcd/internal/core/issueschema"
	"github.com/intentdriven/abcd/internal/fsutil"
)

// PromoteRequest is the input to Promote: graduate an issue into an intent.
// LinkIntent, when non-empty, selects the stamp-only mode that links an
// EXISTING draft (itd-N) instead of minting one — the repair path after a
// stamp failure, and the "I already filed the intent by hand" path.
type PromoteRequest struct {
	RepoRoot   string
	IssuesRoot string
	// ID is the record to graduate: an iss-N, or the rdi-N of a reading item
	// that has been dispositioned.
	ID         string
	LinkIntent string // itd-N; "" mints
}

// PromoteResult is the outcome of a successful Promote. Paths are
// repo-relative. Linked reports stamp-only mode (no draft minted this call).
// MintWarning is the loud-degrade note from the intent-id refs-union scan
// (empty when the scan completed, and always empty in link mode) — the surface
// MUST render it so a degrade to working-tree-only minting is never silent.
type PromoteResult struct {
	IssueID string `json:"issue_id"`
	// IssueStatus is the source record's status. For an issue that is its status
	// directory; for a reading item it is the STANDING DISPOSITION's state,
	// because that family's status signal is the presence of the keyed
	// disposition and never folder membership.
	IssueStatus State  `json:"issue_status"`
	IssuePath   string `json:"issue_path"`
	IntentID    string `json:"intent_id"`
	IntentPath  string `json:"intent_path"`
	Linked      bool   `json:"linked"`
	MintWarning string `json:"mint_warning,omitempty"`
}

// stampWriteHook, when non-nil, replaces the atomic in-place write inside
// Promote's stamp step. It is a test-only seam (nil in production, zero
// overhead) used to force a deterministic post-mint stamp failure without
// relying on platform- or uid-dependent filesystem tricks (a chmod'd status
// dir is a no-op for root), mirroring removeSourceHook in commitTransition.
var stampWriteHook func(path string, data []byte) error

// beforeStampHook, when non-nil, fires between the pre-flight and the moment the
// stamp closure takes the ledger lock. It is a test-only seam (nil in production,
// zero overhead) that forces exactly the window a concurrent write would land in,
// so the under-lock re-checks are exercised rather than asserted. It fires
// OUTSIDE the lock on purpose: a hook that ran inside it could not write to the
// ledger it is meant to change.
var beforeStampHook func()

// Promote graduates an issue into an intent without retyping (spc-24, step 2
// of the record walk). Default mode mints an intent draft — slug reused from
// the issue, body carrying a by-id pointer to the issue rather than a copy
// (SSOT), promoted_from back-edge in the draft's frontmatter — then stamps the
// issue's promoted_to with the minted itd-N. Promotion is orthogonal to
// fix-status: the issue may sit in any status directory and never moves.
//
// Ordering + residue contract: mint first, stamp second. No cross-store lock
// is attempted — the ledger lock alone guards the stamp, exactly as transition
// does — so a failure after the mint leaves an orphan draft; the returned
// error names the draft and the stamp-only remedy
// (`capture promote <iss-N> --intent <itd-N>`).
func Promote(req PromoteRequest) (PromoteResult, error) {
	repoRoot, issuesRoot, err := resolveRoots(req.RepoRoot, req.IssuesRoot)
	if err != nil {
		return PromoteResult{}, err
	}
	if err := mutationPreamble(issuesRoot); err != nil {
		return PromoteResult{}, err
	}
	if reReadingItemID.MatchString(req.ID) {
		return promoteReadingItem(repoRoot, issuesRoot, req)
	}

	// Pre-flight outside the lock: locate and read the issue, refuse a
	// double-promote early (re-checked under the lock at stamp time).
	src, _, err := findIssue(issuesRoot, req.ID)
	if err != nil {
		return PromoteResult{}, err
	}
	content, _, err := readWithChecksum(src)
	if err != nil {
		return PromoteResult{}, err
	}
	fm, body, err := parseFrontmatterAndBody(content)
	if err != nil {
		return PromoteResult{}, err
	}
	if existing := asString(fm["promoted_to"]); existing != "" {
		return PromoteResult{}, fmt.Errorf("%s is already promoted to %s; refusing to promote twice", req.ID, existing)
	}

	var itdID, intentPath, mintWarning string
	linked := req.LinkIntent != ""
	if linked {
		// Stamp-only mode: the target intent must exist in the store (any bucket)
		// BEFORE anything is written — an unknown itd-N is a structural fault.
		// Existence is probed by filename through the shared record-id probe
		// (recordref.go), the same probe resolve's provenance flags use.
		if !reItdID.MatchString(req.LinkIntent) {
			return PromoteResult{}, fmt.Errorf("invalid itd-N identifier: %q", req.LinkIntent)
		}
		rel, ok := findRecordFile(repoRoot, intentStoreRelDirs(), req.LinkIntent)
		if !ok {
			return PromoteResult{}, fmt.Errorf("%s not found in the intent store; nothing stamped", req.LinkIntent)
		}
		itdID, intentPath = req.LinkIntent, rel
	} else {
		// Mint mode: reuse the issue's slug and seed a draft that POINTS at the
		// issue by id — never a copy of its body (the issue record stays the
		// single source of the observation).
		slug := asString(fm["slug"])
		title := issueTitleLine(body, slug)
		seed := "Graduated from `" + req.ID + "`: " + title +
			". Read that issue record for the source observation."
		it, warn, err := intent.CreateDraft(repoRoot, intent.DraftOptions{
			Slug:         slug,
			Title:        title,
			SeedBody:     seed,
			PromotedFrom: req.ID,
		})
		if err != nil {
			return PromoteResult{}, err
		}
		itdID, intentPath, mintWarning = it.ID, it.Path, warn
	}

	// Stamp second, under the ledger lock (the same flock every ledger mutation
	// takes). Re-find and checksum-re-read: the file may have transitioned
	// between the pre-flight and the lock.
	var stamped struct {
		path   string
		status State
	}
	stampErr := withLedgerLock(issuesRoot, func() error {
		src, status, err := findIssue(issuesRoot, req.ID)
		if err != nil {
			return err
		}
		content, _, err := readWithChecksum(src)
		if err != nil {
			return err
		}
		fm, _, err := parseFrontmatterAndBody(content)
		if err != nil {
			return err
		}
		if existing := asString(fm["promoted_to"]); existing != "" {
			return fmt.Errorf("%s is already promoted to %s; refusing to promote twice", req.ID, existing)
		}
		newContent, err := setScalarField(content, "promoted_to", rawScalar(itdID))
		if err != nil {
			return err
		}
		newFM, _, err := parseFrontmatterAndBody(newContent)
		if err != nil {
			return err
		}
		if err := validateStrict(newFM); err != nil {
			return err
		}
		if err := validateInvariants(newFM, status, src); err != nil {
			return err
		}
		// In place, atomic — the file keeps its status directory (promotion is
		// not resolution). The write happens under the same lock as the re-read,
		// so no checksum window exists between them.
		write := fsutil.WriteFileAtomicPreserveMode
		if stampWriteHook != nil {
			write = stampWriteHook
		}
		if err := write(src, []byte(newContent)); err != nil {
			return err
		}
		stamped.path, stamped.status = src, status
		return nil
	})
	if stampErr != nil {
		if !linked {
			// The mint already happened; report the orphan and the repair verb.
			return PromoteResult{}, fmt.Errorf(
				"%w — the minted draft %s (%s) is orphaned; complete the link with `abcd capture promote %s --intent %s`",
				stampErr, itdID, intentPath, req.ID, itdID)
		}
		return PromoteResult{}, stampErr
	}

	return PromoteResult{
		IssueID:     req.ID,
		IssueStatus: stamped.status,
		IssuePath:   fsutil.RepoRel(repoRoot, stamped.path),
		IntentID:    itdID,
		IntentPath:  intentPath,
		Linked:      linked,
		MintWarning: mintWarning,
	}, nil
}

// issueTitleLine derives the minted draft's title — the issue's one-line
// summary — from the first non-blank body line, whitespace-collapsed. A
// degenerate empty body falls back to the slug so the draft still carries an
// honest heading.
func issueTitleLine(body, fallback string) string {
	for _, line := range strings.Split(body, "\n") {
		if fields := strings.Fields(line); len(fields) > 0 {
			return strings.Join(fields, " ")
		}
	}
	return fallback
}

// promoteReadingItem graduates a DISPOSITIONED reading item into an intent
// draft. It is the same act as promoting an issue and shares its ordering
// (mint first, stamp second, no cross-store lock, the orphan named on failure)
// — with one refusal of its own in front.
//
// Item-to-intent without a disposition is the collapse this whole record family
// exists to prevent: it makes the action the answer, and leaves nothing able to
// show that the finding was ever weighed. So the disposition directory is probed
// BEFORE anything is minted, and its absence refuses. Circumventing the verb —
// writing the draft and the stamp by hand — is a lapse-log entry, not something
// this gate can see.
func promoteReadingItem(repoRoot, issuesRoot string, req PromoteRequest) (PromoteResult, error) {
	src, err := findReadingItem(issuesRoot, req.ID)
	if err != nil {
		return PromoteResult{}, err
	}
	content, _, err := readWithChecksum(src)
	if err != nil {
		return PromoteResult{}, err
	}
	fm, _, err := parseFrontmatterAndBody(content)
	if err != nil {
		return PromoteResult{}, err
	}
	if err := validateReadingStrict(fm); err != nil {
		return PromoteResult{}, err
	}
	if existing := asString(fm["promoted_to"]); existing != "" {
		return PromoteResult{}, fmt.Errorf("%s is already promoted to %s; refusing to promote twice", req.ID, existing)
	}

	standing, err := standingDispositions(filepath.Join(issuesRoot, issueschema.DispositionsDir, req.ID))
	if err != nil {
		return PromoteResult{}, err
	}
	if err := refuseUnlessAcceptedGiven(issuesRoot, req.ID, standing); err != nil {
		return PromoteResult{}, err
	}
	state := issueschema.DispositionAccepted

	var itdID, intentPath, mintWarning string
	linked := req.LinkIntent != ""
	if linked {
		if !reItdID.MatchString(req.LinkIntent) {
			return PromoteResult{}, fmt.Errorf("invalid itd-N identifier: %q", req.LinkIntent)
		}
		rel, ok := findRecordFile(repoRoot, intentStoreRelDirs(), req.LinkIntent)
		if !ok {
			return PromoteResult{}, fmt.Errorf("%s not found in the intent store; nothing stamped", req.LinkIntent)
		}
		itdID, intentPath = req.LinkIntent, rel
	} else {
		// The pattern named is the item's one durable one-liner and the only body
		// field every position carries, so it is what the draft is titled and
		// slugged from. The seed POINTS at the item rather than copying it: the
		// reading record stays the single source of what the instrument returned.
		title := asString(fm["pattern"])
		slug, err := normaliseSlug(deriveSlug(title))
		if err != nil {
			return PromoteResult{}, err
		}
		seed := "Graduated from `" + req.ID + "` (" + state + "): " + title +
			". Read that reading record for the instrument's own text."
		it, warn, err := intent.CreateDraft(repoRoot, intent.DraftOptions{
			Slug:         slug,
			Title:        title,
			SeedBody:     seed,
			PromotedFrom: req.ID,
		})
		if err != nil {
			return PromoteResult{}, err
		}
		itdID, intentPath, mintWarning = it.ID, it.Path, warn
	}

	if beforeStampHook != nil {
		beforeStampHook()
	}
	stampErr := withLedgerLock(issuesRoot, func() error {
		src, err := findReadingItem(issuesRoot, req.ID)
		if err != nil {
			return err
		}
		content, _, err := readWithChecksum(src)
		if err != nil {
			return err
		}
		fm, _, err := parseFrontmatterAndBody(content)
		if err != nil {
			return err
		}
		if existing := asString(fm["promoted_to"]); existing != "" {
			return fmt.Errorf("%s is already promoted to %s; refusing to promote twice", req.ID, existing)
		}
		// Re-read the standing answer HERE, not only in the pre-flight. A
		// disposition landing between the two — an acceptance superseded by a
		// rejection while the mint runs — would otherwise leave a standing
		// `rejected` beside a `promoted_to`, a ledger holding both a refusal and
		// the admission it refused. Nothing can land after this check, because the
		// lock is held from here to the write.
		if err := refuseUnlessAccepted(issuesRoot, req.ID); err != nil {
			return err
		}
		newContent, err := setScalarField(content, "promoted_to", rawScalar(itdID))
		if err != nil {
			return err
		}
		newFM, _, err := parseFrontmatterAndBody(newContent)
		if err != nil {
			return err
		}
		if err := validateReadingStrict(newFM); err != nil {
			return err
		}
		write := fsutil.WriteFileAtomicPreserveMode
		if stampWriteHook != nil {
			write = stampWriteHook
		}
		return write(src, []byte(newContent))
	})
	if stampErr != nil {
		if !linked {
			return PromoteResult{}, fmt.Errorf(
				"%w — the minted draft %s (%s) is orphaned; complete the link with `abcd capture promote %s --intent %s`",
				stampErr, itdID, intentPath, req.ID, itdID)
		}
		return PromoteResult{}, stampErr
	}

	return PromoteResult{
		IssueID:     req.ID,
		IssueStatus: State(state),
		IssuePath:   fsutil.RepoRel(repoRoot, src),
		IntentID:    itdID,
		IntentPath:  intentPath,
		Linked:      linked,
		MintWarning: mintWarning,
	}, nil
}

// standingDispositionState reads the state of the item's standing disposition.
// More than one standing answer is a ledger fault the write path refuses, so it
// is reported here rather than silently resolved by picking one.
func standingDispositionState(issuesRoot, item string, standing []string) (string, error) {
	if len(standing) == 0 {
		return "", fmt.Errorf("%w: %s carries no standing disposition", ErrInvariantViolation, item)
	}
	if len(standing) > 1 {
		return "", fmt.Errorf("%w: %s carries %d standing dispositions (%s); exactly one answer is in force at a time",
			ErrInvariantViolation, item, len(standing), renderList(standing))
	}
	path := filepath.Join(issuesRoot, issueschema.DispositionsDir, item, standing[0]+".md")
	// The third reader of a disposition file, and it needs the same guard as the
	// other two: the state read here is what licenses the stamp, so a symlinked
	// record would license it from outside the ledger.
	if err := refuseSymlinkedFile(path); err != nil {
		return "", err
	}
	content, _, err := readWithChecksum(path)
	if err != nil {
		return "", err
	}
	fm, _, err := parseFrontmatterAndBody(content)
	if err != nil {
		return "", err
	}
	return asString(fm["state"]), nil
}

// refuseUnlessAccepted re-reads the item's standing answer and refuses anything
// but an acceptance. It is the under-lock form of the pre-flight check, sharing
// its wording so the two can never describe the rule differently.
func refuseUnlessAccepted(issuesRoot, item string) error {
	standing, err := standingDispositions(filepath.Join(issuesRoot, issueschema.DispositionsDir, item))
	if err != nil {
		return err
	}
	return refuseUnlessAcceptedGiven(issuesRoot, item, standing)
}

// refuseUnlessAcceptedGiven is the rule itself, over a standing set the caller
// has already read.
//
// `accepted` is the one standing state a promotion follows from: acceptance is
// the record, and the action it licenses is a separate admission. An
// undispositioned item collapses the two acts into one, so nothing could show
// the finding was weighed before it was acted on. A `rejected` or `declined`
// one would let the action contradict the record it is supposed to follow from.
// A `held` one would settle by action exactly what the hold left open.
func refuseUnlessAcceptedGiven(issuesRoot, item string, standing []string) error {
	if len(standing) == 0 {
		return fmt.Errorf(
			"%s carries no disposition; an item is answered before it is acted on, and promoting an undispositioned item collapses the two acts into one — record the disposition first (abcd capture disposition %s --state <state> ...)",
			item, item)
	}
	state, err := standingDispositionState(issuesRoot, item, standing)
	if err != nil {
		return err
	}
	if state != issueschema.DispositionAccepted {
		return fmt.Errorf(
			"%s carries a standing disposition of %q, and only %q licenses an action; supersede it with a new disposition first (abcd capture disposition %s --state accepted --grounds \"...\" --supersedes %s)",
			item, state, issueschema.DispositionAccepted, item, standing[0])
	}
	return nil
}
