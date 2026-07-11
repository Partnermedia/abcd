package intent

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/REPPL/abcd-cli/internal/fsutil"
)

// review.go — the intent-fidelity review outbox+inbox (itd-80 phase 4).
//
// The intent file IS the record: its `## Audit Notes` section holds one machine
// marker per review receipt, so idempotency and review state live in one
// committed place (directory/file-as-truth) with no side database.
//
// Two flows meet here:
//
//   - EMIT (emitReviewForIntent, called by Reconcile after a ship move): parks an
//     OWED stub in the shipped intent's Audit Notes and writes an ephemeral review
//     request under .abcd/.work.local/reviews/. Report-only — the caller treats a
//     failure as non-fatal (the intent still ships).
//   - INGEST (IngestVerdict): reads an untrusted verdict JSON emitted by the
//     host-delegated intent-fidelity-reviewer, validates it FAIL-CLOSED against the
//     schema and against the parked OWED receipt, then either replaces the OWED
//     stub with the rendered verdict (INGESTED) or quarantines a bad payload
//     (DEAD_LETTER) — never a partial application.
//
// Receipt digest. resource_digest = sha256 over the intent's `## Acceptance
// Criteria` section body — the authority the reviewer judges and the only intent
// text the criteria map onto. It deliberately EXCLUDES the Audit Notes section,
// so writing the marker does not change the receipt: re-emit and re-ingest stay
// idempotent. receipt_id = "rcp-" + first-12-hex of sha256(intent_id | spec_id |
// hex(resource_digest)). No timestamps feed it (deterministic).

// VerdictType is the only _type the ingest accepts.
const VerdictType = "abcd/intent-fidelity-verdict/v1"

// reviewsRelDir is the ephemeral (gitignored) review outbox/quarantine.
const reviewsRelDir = ".abcd/.work.local/reviews"

// maxVerdictBytes caps the untrusted verdict payload (trust boundary).
const maxVerdictBytes = 1 * 1024 * 1024

// verdictEnum is the closed set of acceptance verdicts.
var verdictEnum = map[string]bool{
	"MET": true, "MET_WITH_CONCERNS": true, "NOT_MET": true, "INCONCLUSIVE": true,
}

var (
	// rcpIDRe constrains a receipt id so it can never build a path that escapes
	// the reviews dir (path-traversal defence). 12 lowercase hex chars.
	rcpIDRe = regexp.MustCompile(`^rcp-[0-9a-f]{12}$`)
	// auditHeadingRe matches the `## Audit Notes` heading (any heading depth).
	auditHeadingRe = regexp.MustCompile(`^#{1,6}\s+Audit Notes\s*$`)
	// bulletRe matches a TOP-LEVEL markdown list item. Acceptance-Criteria bullets
	// are numbered positionally ac-1..ac-K, so only column-0 bullets count — an
	// indented sub-bullet is detail of its parent, not a separate criterion.
	bulletRe = regexp.MustCompile(`^[-*]\s+\S`)
	// markerRe matches a parked review marker line inside the Audit Notes.
	markerRe = regexp.MustCompile(`<!-- abcd-review: (OWED|INGESTED|DEAD_LETTER) receipt=(rcp-[0-9a-f]+) -->`)
	// criterionIDRe validates a criterion id shape before it is positionally bounded.
	criterionIDRe = regexp.MustCompile(`^ac-([0-9]+)$`)
)

// ---------------------------------------------------------------------------
// Verdict schema (hand-rolled; stdlib encoding/json only)
// ---------------------------------------------------------------------------

type verdict struct {
	Type              string             `json:"_type"`
	ReceiptID         string             `json:"receipt_id"`
	Verifier          verdictVerifier    `json:"verifier"`
	Policy            verdictPolicy      `json:"policy"`
	InputAttestations []verdictAttest    `json:"input_attestations"`
	Criteria          []verdictCriterion `json:"criteria"`
	AcceptanceRollup  map[string]int     `json:"acceptance_rollup"`
	GapAudit          verdictGapAudit    `json:"gap_audit"`
}

type verdictVerifier struct {
	ID      string `json:"id"`
	Version string `json:"version"`
}

type verdictPolicy struct {
	RubricHash string `json:"rubric_hash"`
	PromptHash string `json:"prompt_hash"`
}

type verdictAttest struct {
	Kind   string `json:"kind"`
	Ref    string `json:"ref"`
	Digest string `json:"digest"`
}

type verdictEvidence struct {
	Ref   string `json:"ref"`
	Quote string `json:"quote"`
}

type verdictCriterion struct {
	CriterionID string            `json:"criterion_id"`
	Verdict     string            `json:"verdict"`
	Rationale   string            `json:"rationale"`
	Evidence    []verdictEvidence `json:"evidence"`
}

type verdictGapEntry struct {
	Claim    string            `json:"claim"`
	Evidence []verdictEvidence `json:"evidence"`
}

type verdictGapAudit struct {
	Honoured []verdictGapEntry `json:"honoured"`
	Diverged []verdictGapEntry `json:"diverged"`
	Missing  []verdictGapEntry `json:"missing"`
}

// ---------------------------------------------------------------------------
// Results
// ---------------------------------------------------------------------------

// ReviewEmitResult reports one emit (OWED stub + request file).
type ReviewEmitResult struct {
	ReceiptID   string `json:"receipt_id"`
	IntentID    string `json:"intent_id"`
	Status      string `json:"status"` // owed | already_owed | already_ingested | already_dead_letter
	RequestPath string `json:"request_path"`
}

// IngestVerdictResult reports one verdict ingest.
type IngestVerdictResult struct {
	Status         string `json:"status"` // ingested | dead_letter | noop
	ReceiptID      string `json:"receipt_id"`
	IntentID       string `json:"intent_id"`
	Criteria       int    `json:"criteria"`
	Met            int    `json:"met"`
	MetWithConcern int    `json:"met_with_concerns"`
	NotMet         int    `json:"not_met"`
	Inconclusive   int    `json:"inconclusive"`
	DeadLetterPath string `json:"dead_letter_path,omitempty"`
	Reason         string `json:"reason,omitempty"`
}

// ---------------------------------------------------------------------------
// Emit (called by Reconcile; also the manual re-emit verb)
// ---------------------------------------------------------------------------

// receiptFor computes the deterministic receipt id for an intent/spec pair from
// the intent's Acceptance Criteria section (see the package-level digest note).
func receiptFor(intentID, specID, content string) string {
	acBody := sectionBody(content, acHeadingRe)
	rd := sha256.Sum256([]byte(acBody))
	h := sha256.Sum256([]byte(intentID + "|" + specID + "|" + hex.EncodeToString(rd[:])))
	return "rcp-" + hex.EncodeToString(h[:])[:12]
}

// emitReviewForIntent parks an OWED stub in the intent's Audit Notes and writes
// the ephemeral review request. It is idempotent: if a marker for the computed
// receipt already exists (OWED/INGESTED/DEAD_LETTER) the Audit Notes are left
// untouched. All ids are validated before any path is built.
func emitReviewForIntent(repoRoot string, it Intent) (ReviewEmitResult, error) {
	if !intentIDRe.MatchString(it.ID) {
		return ReviewEmitResult{}, fmt.Errorf("intent: id %q must match ^itd-[0-9]+$", it.ID)
	}
	if !specIDRe.MatchString(it.SpecID) {
		return ReviewEmitResult{}, fmt.Errorf("intent: spec id %q must match ^spc-[0-9]+$", it.SpecID)
	}
	abs := filepath.Join(repoRoot, it.Path)
	data, err := readRepoFile(abs, it.Path)
	if err != nil {
		return ReviewEmitResult{}, err
	}
	content := string(data)

	// Reuse an already-parked receipt rather than recomputing one. The receipt
	// digest excludes the Audit Notes section, but creating that section on the
	// first emit (when it was absent and the Acceptance Criteria was the file's
	// last section) can still shift what sectionBody reads as the AC body — so a
	// freshly recomputed receipt may disagree with the parked marker and append a
	// second stub. The parked marker is the authority the ingest resolves against.
	if rcp, state, ok := existingMarker(content); ok {
		res := ReviewEmitResult{ReceiptID: rcp, IntentID: it.ID}
		res.RequestPath = filepath.Join(reviewsRelDir, rcp+".request.md")
		switch state {
		case "INGESTED":
			res.Status = "already_ingested"
		case "DEAD_LETTER":
			res.Status = "already_dead_letter"
		default:
			res.Status = "already_owed"
			// Only an OWED receipt still awaits a verdict: ensure its ephemeral
			// request still exists (it is gitignored and may have been swept). A
			// terminal INGESTED/DEAD_LETTER receipt needs no request rewrite.
			if err := writeReviewRequest(repoRoot, it, rcp, content); err != nil {
				return res, err
			}
		}
		return res, nil
	}

	rcp := receiptFor(it.ID, it.SpecID, content)
	res := ReviewEmitResult{ReceiptID: rcp, IntentID: it.ID}
	block := owedBlock(rcp)
	updated := upsertReviewBlock(content, rcp, block)
	if err := fsutil.WriteFileAtomic(abs, []byte(updated), 0o644); err != nil {
		return ReviewEmitResult{}, fmt.Errorf("intent: writing OWED stub to %s: %w", it.Path, err)
	}
	if err := writeReviewRequest(repoRoot, it, rcp, updated); err != nil {
		return ReviewEmitResult{}, err
	}
	res.Status = "owed"
	res.RequestPath = filepath.Join(reviewsRelDir, rcp+".request.md")
	return res, nil
}

// ReEmitReview re-parks the OWED stub and request for a shipped intent (the
// manual `abcd intent review <itd-N>` verb). It resolves the intent, refuses one
// not in shipped/, and delegates to the shared emit.
func ReEmitReview(repoRoot, intentID string) (ReviewEmitResult, error) {
	if !intentIDRe.MatchString(intentID) {
		return ReviewEmitResult{}, fmt.Errorf("intent: id %q must match ^itd-[0-9]+$", intentID)
	}
	corpus, err := Load(repoRoot)
	if err != nil {
		return ReviewEmitResult{}, err
	}
	it, ok := corpus.Lookup(intentID)
	if !ok {
		return ReviewEmitResult{}, fmt.Errorf("intent: %s not found in any bucket", intentID)
	}
	if it.Bucket != BucketShipped {
		return ReviewEmitResult{}, fmt.Errorf("intent: %s is in %s, not shipped; only a shipped intent owes a fidelity review", intentID, it.Bucket)
	}
	if !specIDRe.MatchString(it.SpecID) {
		return ReviewEmitResult{}, fmt.Errorf("intent: %s has no well-formed spec_id (%q); refusing to emit a review", intentID, it.SpecID)
	}
	return emitReviewForIntent(repoRoot, it)
}

// writeReviewRequest writes the ephemeral review request markdown. The request is
// a prompt over the intent's Acceptance Criteria plus the receipt metadata; the
// host reads it, runs the reviewer, and produces the verdict JSON.
func writeReviewRequest(repoRoot string, it Intent, rcp, content string) error {
	if !rcpIDRe.MatchString(rcp) {
		return fmt.Errorf("intent: receipt id %q is malformed; refusing to build a request path", rcp)
	}
	dir := filepath.Join(repoRoot, reviewsRelDir)
	if err := ensureRealDir(dir, reviewsRelDir); err != nil {
		return err
	}
	ac := strings.TrimSpace(sectionBody(content, acHeadingRe))
	var b strings.Builder
	fmt.Fprintf(&b, "# Fidelity review request — %s\n\n", rcp)
	fmt.Fprintf(&b, "- receipt_id: %s\n", rcp)
	fmt.Fprintf(&b, "- intent: %s\n", it.Path)
	fmt.Fprintf(&b, "- spec: %s\n", it.SpecID)
	fmt.Fprintf(&b, "- delivered: the diff/commit range that realised %s (host supplies the range)\n\n", it.SpecID)
	b.WriteString("## Acceptance Criteria (authority; numbered ac-1..ac-K in order)\n\n")
	if ac == "" {
		b.WriteString("(none found)\n")
	} else {
		b.WriteString(ac + "\n")
	}
	b.WriteString("\nRun the intent-fidelity-reviewer agent over the criteria and the delivered\n")
	b.WriteString("diff, then ingest its verdict JSON:\n\n")
	fmt.Fprintf(&b, "    abcd intent review ingest --verdict-json <path>   # receipt %s\n", rcp)

	path := filepath.Join(dir, rcp+".request.md")
	if err := fsutil.WriteFileAtomic(path, []byte(b.String()), 0o644); err != nil {
		return fmt.Errorf("intent: writing review request %s: %w", filepath.Join(reviewsRelDir, rcp+".request.md"), err)
	}
	return nil
}

// ---------------------------------------------------------------------------
// Ingest (untrusted verdict -> committed Audit Notes)
// ---------------------------------------------------------------------------

// IngestVerdict reads an untrusted intent-fidelity verdict JSON and applies it to
// the committed record, FAIL-CLOSED at every step:
//
//   - malformed/oversize/unreadable payload with no resolvable receipt -> reject;
//   - receipt matching no parked marker (unsolicited) -> reject;
//   - already INGESTED for this receipt -> no-op;
//   - schema/semantic validation failure on a resolvable receipt -> DEAD_LETTER
//     (marker + INCONCLUSIVE criteria + retained raw payload), never partial;
//   - otherwise -> INGESTED (OWED stub replaced by the rendered verdict).
func IngestVerdict(repoRoot, verdictPath string) (IngestVerdictResult, error) {
	raw, err := readVerdictFile(verdictPath)
	if err != nil {
		return IngestVerdictResult{}, err
	}

	// Lenient first pass: recover _type + receipt id so we can classify and
	// resolve the payload. A payload that is not a fidelity verdict at all, or that
	// we cannot even key on, has no home and is rejected outright (not dead-lettered).
	var lenient struct {
		Type      string `json:"_type"`
		ReceiptID string `json:"receipt_id"`
	}
	if err := json.Unmarshal(raw, &lenient); err != nil {
		return IngestVerdictResult{}, fmt.Errorf("intent: verdict is not parseable JSON; refusing to ingest: %w", err)
	}
	if lenient.Type != VerdictType {
		return IngestVerdictResult{}, fmt.Errorf("intent: verdict _type %q is not %q; refusing to ingest", lenient.Type, VerdictType)
	}
	if !rcpIDRe.MatchString(lenient.ReceiptID) {
		return IngestVerdictResult{}, fmt.Errorf("intent: verdict has no resolvable receipt_id (malformed or absent); refusing to ingest")
	}
	rcp := lenient.ReceiptID

	it, content, state, ok, err := findIntentByReceipt(repoRoot, rcp)
	if err != nil {
		return IngestVerdictResult{}, err
	}
	if !ok {
		return IngestVerdictResult{}, fmt.Errorf("intent: verdict receipt %s matches no parked review marker (unsolicited); refusing to ingest", rcp)
	}
	if state == "INGESTED" {
		return IngestVerdictResult{Status: "noop", ReceiptID: rcp, IntentID: it.ID}, nil
	}

	// Full schema + semantic validation. Any failure with a resolvable receipt
	// quarantines the payload rather than corrupting the record.
	v, verr := validateVerdict(raw, rcp, content)
	if verr != nil {
		return deadLetter(repoRoot, it, content, rcp, raw, verr.Error())
	}

	rollup := countVerdicts(v)
	block := ingestedBlock(rcp, v, rollup)
	updated := upsertReviewBlock(content, rcp, block)
	if err := fsutil.WriteFileAtomic(filepath.Join(repoRoot, it.Path), []byte(updated), 0o644); err != nil {
		return IngestVerdictResult{}, fmt.Errorf("intent: writing verdict to %s: %w", it.Path, err)
	}
	return IngestVerdictResult{
		Status: "ingested", ReceiptID: rcp, IntentID: it.ID, Criteria: len(v.Criteria),
		Met: rollup["MET"], MetWithConcern: rollup["MET_WITH_CONCERNS"],
		NotMet: rollup["NOT_MET"], Inconclusive: rollup["INCONCLUSIVE"],
	}, nil
}

// readVerdictFile reads the payload behind the trust guards (regular file, no
// symlink, size cap).
func readVerdictFile(path string) ([]byte, error) {
	fi, err := os.Lstat(path)
	if err != nil {
		return nil, fmt.Errorf("intent: stat verdict %s: %w", path, err)
	}
	if fi.Mode()&os.ModeSymlink != 0 {
		return nil, fmt.Errorf("intent: verdict %s is a symlink (refusing to follow)", path)
	}
	if !fi.Mode().IsRegular() {
		return nil, fmt.Errorf("intent: verdict %s is not a regular file", path)
	}
	if fi.Size() > maxVerdictBytes {
		return nil, fmt.Errorf("intent: verdict %s exceeds the %d-byte cap", path, maxVerdictBytes)
	}
	// The Lstat->ReadFile gap is a benign TOCTOU: a swap between the two opens a
	// different regular file, not a symlink escape, and is accepted under the
	// trusted-worktree model (mirrors the ensureRealDir ancestor-symlink note).
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("intent: reading verdict %s: %w", path, err)
	}
	return data, nil
}

// validateVerdict parses and fully validates the payload against the reviewer
// contract and the intent's actual Acceptance Criteria. It returns a non-nil
// error describing the first violation (the DEAD_LETTER reason).
func validateVerdict(raw []byte, rcp, intentContent string) (verdict, error) {
	dec := json.NewDecoder(strings.NewReader(string(raw)))
	dec.DisallowUnknownFields() // reject smuggled extra fields
	var v verdict
	if err := dec.Decode(&v); err != nil {
		return verdict{}, fmt.Errorf("malformed verdict JSON: %v", err)
	}
	if v.Type != VerdictType {
		return verdict{}, fmt.Errorf("wrong _type %q (want %q)", v.Type, VerdictType)
	}
	if v.ReceiptID != rcp {
		return verdict{}, fmt.Errorf("receipt_id %q disagrees with the resolved receipt %q", v.ReceiptID, rcp)
	}
	// The attestation chain (rubric + prompt the host pinned) is what makes this a
	// VSA-shaped verdict rather than an unverifiable opinion; require both.
	if strings.TrimSpace(v.Policy.RubricHash) == "" || strings.TrimSpace(v.Policy.PromptHash) == "" {
		return verdict{}, fmt.Errorf("policy.rubric_hash and policy.prompt_hash are both required")
	}
	if len(v.Criteria) == 0 {
		return verdict{}, fmt.Errorf("criteria is empty")
	}

	k := countAcceptanceCriteria(intentContent)
	if k == 0 {
		return verdict{}, fmt.Errorf("intent has no parseable Acceptance Criteria bullets to judge")
	}
	seen := map[int]bool{}
	for i, c := range v.Criteria {
		m := criterionIDRe.FindStringSubmatch(c.CriterionID)
		if m == nil {
			return verdict{}, fmt.Errorf("criterion[%d] id %q is not ac-N", i, c.CriterionID)
		}
		n := atoiPositive(m[1])
		if n < 1 || n > k {
			return verdict{}, fmt.Errorf("criterion %q is out of range (intent has ac-1..ac-%d)", c.CriterionID, k)
		}
		// Dedup on the resolved index, so ac-1 and a zero-padded ac-01 (which map to
		// the same bullet) cannot both be applied.
		if seen[n] {
			return verdict{}, fmt.Errorf("criterion %q targets an already-judged Acceptance-Criteria bullet (ac-%d)", c.CriterionID, n)
		}
		seen[n] = true
		if !verdictEnum[c.Verdict] {
			return verdict{}, fmt.Errorf("criterion %q has out-of-enum verdict %q", c.CriterionID, c.Verdict)
		}
		if !hasCitedEvidence(c.Evidence) {
			return verdict{}, fmt.Errorf("criterion %q cites no evidence ref", c.CriterionID)
		}
	}
	// Every criterion must be judged: a verdict covering only some of ac-1..ac-K is
	// a PARTIAL judgement. Accepting it would write an incomplete INGESTED state and
	// let the idempotency short-circuit drop a later complete verdict — fail closed.
	if len(seen) != k {
		return verdict{}, fmt.Errorf("verdict judges %d of %d Acceptance-Criteria bullets (every ac-1..ac-%d must be judged exactly once)", len(seen), k, k)
	}

	// Rollup counts must sum to the number of criteria (reviewer contract rule 4).
	sum := 0
	for key, n := range v.AcceptanceRollup {
		if !verdictEnum[key] {
			return verdict{}, fmt.Errorf("acceptance_rollup has non-verdict key %q", key)
		}
		sum += n
	}
	if sum != len(v.Criteria) {
		return verdict{}, fmt.Errorf("acceptance_rollup sums to %d, not the %d criteria", sum, len(v.Criteria))
	}

	for _, bucket := range [][2]any{{"honoured", v.GapAudit.Honoured}, {"diverged", v.GapAudit.Diverged}, {"missing", v.GapAudit.Missing}} {
		name := bucket[0].(string)
		for i, e := range bucket[1].([]verdictGapEntry) {
			if !hasCitedEvidence(e.Evidence) {
				return verdict{}, fmt.Errorf("gap_audit.%s[%d] cites no evidence ref", name, i)
			}
		}
	}
	return v, nil
}

// deadLetter quarantines a bad-but-resolvable verdict: it retains the raw payload
// under the ephemeral reviews dir and replaces the parked marker with a
// DEAD_LETTER block recording all criteria INCONCLUSIVE. Never partial.
func deadLetter(repoRoot string, it Intent, content, rcp string, raw []byte, reason string) (IngestVerdictResult, error) {
	if !rcpIDRe.MatchString(rcp) {
		return IngestVerdictResult{}, fmt.Errorf("intent: receipt id %q is malformed; refusing to dead-letter", rcp)
	}
	dir := filepath.Join(repoRoot, reviewsRelDir)
	if err := ensureRealDir(dir, reviewsRelDir); err != nil {
		return IngestVerdictResult{}, err
	}
	dlRel := filepath.Join(reviewsRelDir, rcp+".deadletter.json")
	if err := fsutil.WriteFileAtomic(filepath.Join(dir, rcp+".deadletter.json"), raw, 0o644); err != nil {
		return IngestVerdictResult{}, fmt.Errorf("intent: retaining dead-letter payload %s: %w", dlRel, err)
	}
	block := deadLetterBlock(rcp, reason, dlRel)
	updated := upsertReviewBlock(content, rcp, block)
	if err := fsutil.WriteFileAtomic(filepath.Join(repoRoot, it.Path), []byte(updated), 0o644); err != nil {
		return IngestVerdictResult{}, fmt.Errorf("intent: writing dead-letter marker to %s: %w", it.Path, err)
	}
	return IngestVerdictResult{
		Status: "dead_letter", ReceiptID: rcp, IntentID: it.ID,
		DeadLetterPath: dlRel, Reason: reason,
	}, nil
}

// ---------------------------------------------------------------------------
// Receipt resolution + Audit Notes surgery
// ---------------------------------------------------------------------------

// findIntentByReceipt scans every bucket for the intent whose Audit Notes carry a
// review marker for rcp. It returns the intent, its content, and the marker state
// (OWED/INGESTED/DEAD_LETTER). ok is false when no intent claims the receipt.
func findIntentByReceipt(repoRoot, rcp string) (Intent, string, string, bool, error) {
	corpus, err := Load(repoRoot)
	if err != nil {
		return Intent{}, "", "", false, err
	}
	for _, it := range corpus.Intents {
		data, err := readRepoFile(filepath.Join(repoRoot, it.Path), it.Path)
		if err != nil {
			return Intent{}, "", "", false, err
		}
		content := string(data)
		if state, ok := markerState(content, rcp); ok {
			return it, content, state, true, nil
		}
	}
	return Intent{}, "", "", false, nil
}

// existingMarker returns the receipt id and state of the FIRST parked review
// marker in content, if any. Emit reuses this parked receipt rather than
// recomputing one (see emitReviewForIntent's receipt-shift note).
func existingMarker(content string) (string, string, bool) {
	if m := markerRe.FindStringSubmatch(content); m != nil {
		return m[2], m[1], true
	}
	return "", "", false
}

// markerState returns the state of the review marker for rcp, if present.
func markerState(content, rcp string) (string, bool) {
	for _, m := range markerRe.FindAllStringSubmatch(content, -1) {
		if m[2] == rcp {
			return m[1], true
		}
	}
	return "", false
}

// upsertReviewBlock replaces the existing review block for rcp with newBlock, or
// appends newBlock to the Audit Notes section (creating the section if absent). A
// review block runs from its marker line to the next marker, the next heading, or
// end of file.
func upsertReviewBlock(content, rcp, newBlock string) string {
	lines := strings.Split(content, "\n")
	start := -1
	for i, ln := range lines {
		m := markerRe.FindStringSubmatch(strings.TrimRight(ln, "\r"))
		if m != nil && m[2] == rcp {
			start = i
			break
		}
	}
	if start >= 0 {
		end := len(lines)
		for j := start + 1; j < len(lines); j++ {
			t := strings.TrimRight(lines[j], "\r")
			if markerRe.MatchString(t) || headingRe.MatchString(t) {
				end = j
				break
			}
		}
		out := make([]string, 0, len(lines))
		out = append(out, lines[:start]...)
		out = append(out, strings.Split(newBlock, "\n")...)
		out = append(out, lines[end:]...)
		return strings.Join(out, "\n")
	}
	return appendToAuditNotes(content, newBlock)
}

// appendToAuditNotes appends a block to the `## Audit Notes` section, creating
// the section at end of file if it is absent.
func appendToAuditNotes(content, block string) string {
	lines := strings.Split(content, "\n")
	head := -1
	for i, ln := range lines {
		if auditHeadingRe.MatchString(strings.TrimRight(ln, "\r")) {
			head = i
			break
		}
	}
	if head < 0 {
		body := strings.TrimRight(content, "\n")
		return body + "\n\n## Audit Notes\n\n" + block + "\n"
	}
	// Find the end of the Audit Notes section (next heading or EOF).
	end := len(lines)
	for j := head + 1; j < len(lines); j++ {
		if headingRe.MatchString(strings.TrimRight(lines[j], "\r")) {
			end = j
			break
		}
	}
	section := lines[head+1 : end]
	// Drop trailing blank lines inside the section, then re-add one separator.
	for len(section) > 0 && strings.TrimSpace(section[len(section)-1]) == "" {
		section = section[:len(section)-1]
	}
	rebuilt := make([]string, 0, len(lines)+8)
	rebuilt = append(rebuilt, lines[:head+1]...)
	rebuilt = append(rebuilt, "")
	rebuilt = append(rebuilt, section...)
	if len(section) > 0 {
		rebuilt = append(rebuilt, "")
	}
	rebuilt = append(rebuilt, strings.Split(block, "\n")...)
	rebuilt = append(rebuilt, "")
	rebuilt = append(rebuilt, lines[end:]...)
	return strings.Join(rebuilt, "\n")
}

// ---------------------------------------------------------------------------
// Block rendering (deterministic; no timestamps)
// ---------------------------------------------------------------------------

func owedBlock(rcp string) string {
	return fmt.Sprintf("<!-- abcd-review: OWED receipt=%s -->\nFidelity review OWED (receipt %s).", rcp, rcp)
}

func deadLetterBlock(rcp, reason, dlRel string) string {
	// reason is derived from untrusted payload content (e.g. an out-of-enum token),
	// so it is sanitised before it lands in the committed record.
	return fmt.Sprintf("<!-- abcd-review: DEAD_LETTER receipt=%s -->\n"+
		"Fidelity review DEAD_LETTER (receipt %s): %s. Raw payload retained at %s. "+
		"All criteria recorded INCONCLUSIVE.", rcp, rcp, oneLine(reason), dlRel)
}

func ingestedBlock(rcp string, v verdict, rollup map[string]int) string {
	var b strings.Builder
	fmt.Fprintf(&b, "<!-- abcd-review: INGESTED receipt=%s -->\n", rcp)
	fmt.Fprintf(&b, "Fidelity review — receipt %s (verifier %s %s).\n\n",
		rcp, orDash(v.Verifier.ID), orDash(v.Verifier.Version))
	// Pinned provenance: the verifier identity, the policy hashes it attested to,
	// and every input attestation. All fields are untrusted, so route each through
	// the oneLine neutraliser before it lands in the committed record.
	fmt.Fprintf(&b, "Provenance: %s@%s · rubric_hash %s · prompt_hash %s\n",
		orDash(v.Verifier.ID), orDash(v.Verifier.Version),
		orDash(v.Policy.RubricHash), orDash(v.Policy.PromptHash))
	if len(v.InputAttestations) > 0 {
		b.WriteString("Input attestations:")
		for _, a := range v.InputAttestations {
			fmt.Fprintf(&b, " %s:%s@%s;", orDash(a.Kind), orDash(a.Ref), orDash(a.Digest))
		}
		b.WriteString("\n")
	}
	b.WriteString("\n")
	fmt.Fprintf(&b, "Acceptance rollup: MET %d · MET_WITH_CONCERNS %d · NOT_MET %d · INCONCLUSIVE %d\n\n",
		rollup["MET"], rollup["MET_WITH_CONCERNS"], rollup["NOT_MET"], rollup["INCONCLUSIVE"])

	b.WriteString("Per-criterion verdicts:\n")
	for _, c := range v.Criteria {
		fmt.Fprintf(&b, "- %s — %s: %s\n", c.CriterionID, c.Verdict, oneLine(c.Rationale))
		for _, e := range c.Evidence {
			fmt.Fprintf(&b, "  evidence: %s\n", renderEvidence(e))
		}
	}
	b.WriteString("\nGap audit:\n")
	renderBucket(&b, "honoured", v.GapAudit.Honoured)
	renderBucket(&b, "diverged", v.GapAudit.Diverged)
	renderBucket(&b, "missing", v.GapAudit.Missing)
	return strings.TrimRight(b.String(), "\n")
}

func renderBucket(b *strings.Builder, name string, entries []verdictGapEntry) {
	if len(entries) == 0 {
		fmt.Fprintf(b, "- %s: (none)\n", name)
		return
	}
	fmt.Fprintf(b, "- %s:\n", name)
	for _, e := range entries {
		fmt.Fprintf(b, "  - %s\n", oneLine(e.Claim))
		for _, ev := range e.Evidence {
			fmt.Fprintf(b, "    evidence: %s\n", renderEvidence(ev))
		}
	}
}

func renderEvidence(e verdictEvidence) string {
	ref := oneLine(e.Ref)
	if q := oneLine(e.Quote); q != "" {
		return fmt.Sprintf("%s — %q", ref, q)
	}
	return ref
}

// ---------------------------------------------------------------------------
// Small helpers
// ---------------------------------------------------------------------------

// sectionBody returns the text of the section introduced by the first heading
// matching headRe, up to the next heading or end of file.
func sectionBody(content string, headRe *regexp.Regexp) string {
	lines := strings.Split(content, "\n")
	for i, ln := range lines {
		if !headRe.MatchString(strings.TrimRight(ln, "\r")) {
			continue
		}
		var body []string
		for _, b := range lines[i+1:] {
			if headingRe.MatchString(strings.TrimRight(b, "\r")) {
				break
			}
			body = append(body, b)
		}
		return strings.Join(body, "\n")
	}
	return ""
}

// countAcceptanceCriteria counts the top-level list bullets in the intent's
// `## Acceptance Criteria` section — the positional authority ac-1..ac-K.
func countAcceptanceCriteria(content string) int {
	n := 0
	for _, ln := range strings.Split(sectionBody(content, acHeadingRe), "\n") {
		if bulletRe.MatchString(strings.TrimRight(ln, "\r")) {
			n++
		}
	}
	return n
}

func countVerdicts(v verdict) map[string]int {
	m := map[string]int{"MET": 0, "MET_WITH_CONCERNS": 0, "NOT_MET": 0, "INCONCLUSIVE": 0}
	for _, c := range v.Criteria {
		m[c.Verdict]++
	}
	return m
}

func hasCitedEvidence(ev []verdictEvidence) bool {
	for _, e := range ev {
		if strings.TrimSpace(e.Ref) != "" {
			return true
		}
	}
	return false
}

// oneLine sanitises an untrusted verdict string before it is rendered into the
// committed Audit Notes. It collapses newlines (so injected content cannot break
// out of its line) AND neutralises HTML-comment delimiters, so untrusted content
// can never forge an `<!-- abcd-review: <STATE> receipt=<rcp> -->` marker to spoof
// review state, misroute a future ingest, or poison idempotency into a false
// no-op. Every untrusted field rendered into the record passes through here.
func oneLine(s string) string {
	s = strings.ReplaceAll(s, "\r", " ")
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.ReplaceAll(s, "<!--", "< !--")
	s = strings.ReplaceAll(s, "-->", "-- >")
	return strings.TrimSpace(s)
}

func orDash(s string) string {
	if s = oneLine(s); s == "" {
		return "-"
	}
	return s
}

// atoiPositive parses a non-negative decimal string (already ^[0-9]+$ via regex).
func atoiPositive(s string) int {
	n := 0
	for _, r := range s {
		n = n*10 + int(r-'0')
		if n > 1_000_000 {
			return n // clamp; huge indices are out-of-range anyway
		}
	}
	return n
}
