package capture

import (
	"fmt"
	"strings"

	"github.com/REPPL/abcd-cli/internal/core/intent"
	"github.com/REPPL/abcd-cli/internal/fsutil"
)

// PromoteRequest is the input to Promote: graduate an issue into an intent.
// LinkIntent, when non-empty, selects the stamp-only mode that links an
// EXISTING draft (itd-N) instead of minting one — the repair path after a
// stamp failure, and the "I already filed the intent by hand" path.
type PromoteRequest struct {
	RepoRoot   string
	IssuesRoot string
	ID         string // iss-N
	LinkIntent string // itd-N; "" mints
}

// PromoteResult is the outcome of a successful Promote. Paths are
// repo-relative. Linked reports stamp-only mode (no draft minted this call).
// MintWarning is the loud-degrade note from the intent-id refs-union scan
// (empty when the scan completed, and always empty in link mode) — the surface
// MUST render it so a degrade to working-tree-only minting is never silent.
type PromoteResult struct {
	IssueID     string `json:"issue_id"`
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
