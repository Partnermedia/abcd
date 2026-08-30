package capture

import (
	"fmt"
	"strings"

	"github.com/intentdriven/abcd/internal/core/intent"
	"github.com/intentdriven/abcd/internal/fsutil"
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
	// Grounds is the REQUIRED conjecture behind the promotion, in the shared
	// `<token>: <text>` grammar (core/grounds). A capture routed to an intent
	// draft is a conjecture being pursued, and there is nothing to stage here:
	// promote mints the value in the same call, so it has no corpus to fix.
	Grounds string
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
	// Redacted / Degraded mirror TransitionResult: the grounds text is free prose
	// written to the same committed ledger, so it goes through the same redactor
	// and reports the same way. Rewriting somebody's reasoning in silence is worse
	// than not recording it.
	Redacted int    `json:"redacted,omitempty"`
	Degraded string `json:"redaction_degraded,omitempty"`
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
//
// Every refusal that can be established from the bytes in hand is therefore
// raised BEFORE the mint: the grounds text, and whether the record can take the
// append at all. What is left to the stamp is what only a write under the lock
// can discover, which is the residue the remedy above is for — never a
// deterministic refusal that would leak one draft per attempt.
func Promote(req PromoteRequest) (PromoteResult, error) {
	repoRoot, issuesRoot, err := resolveRoots(req.RepoRoot, req.IssuesRoot)
	if err != nil {
		return PromoteResult{}, err
	}
	if err := mutationPreamble(issuesRoot); err != nil {
		return PromoteResult{}, err
	}
	// BEFORE anything is minted or stamped. Promote's residue contract is
	// mint-first-stamp-second, so a refusal raised any later than here would leave
	// an orphan draft behind for a missing argument — the exact residue the rest
	// of this path works to avoid.
	g, gRedacted, gDegraded, err := requireGrounds(repoRoot, "promote", req.Grounds)
	if err != nil {
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
	// Establish that the RECORD can accept the append, before anything is minted.
	// requireGrounds above already gated the grounds TEXT; what it cannot answer
	// is whether the bytes it will be appended to can hold it. A record whose body
	// leaves a fence or a comment open masks everything appended below it, so the
	// stamp's read-back refuses — permanently, and identically on every retry,
	// including the retry through the repair verb the failure message names. With
	// the mint first, that left one orphan draft per attempt and a draft counter
	// climbing behind an operator who had no way to succeed (iss-2608301803423101).
	//
	// The dry run is the real append against the pre-flight bytes, discarded. It
	// is not a substitute for the guard under the lock — the file may change
	// between the two, and the write is judged again there — but the failure it
	// removes is the deterministic one, where the record could never have taken
	// the entry in the first place.
	if _, err := appendGrounds("promote", content, g); err != nil {
		return PromoteResult{}, err
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
		newContent, err = appendGrounds("promote", newContent, g)
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
		Redacted:    gRedacted,
		Degraded:    gDegraded,
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
