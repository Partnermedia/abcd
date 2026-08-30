package intent

import (
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/intentdriven/abcd/internal/core/frontmatter"
	"github.com/intentdriven/abcd/internal/core/recordid"
	"github.com/intentdriven/abcd/internal/core/spec"
	"github.com/intentdriven/abcd/internal/fsutil"
)

// Load discovers intent files across every lifecycle bucket, parses their
// frontmatter, and returns the in-memory Corpus. A missing intents/ directory
// (or a missing individual bucket) yields no records for it (soft, mirroring
// spec.Load). A present-but-malformed intent file — one whose frontmatter lacks
// a well-formed id — is a hard, loud error.
func Load(repoRoot string) (Corpus, error) {
	var c Corpus
	for _, bucket := range Buckets {
		intents, err := loadBucket(repoRoot, bucket)
		if err != nil {
			return Corpus{}, err
		}
		c.Intents = append(c.Intents, intents...)
	}
	return c, nil
}

// loadBucket reads one bucket directory. A missing directory is soft (nil, nil).
func loadBucket(repoRoot, bucket string) ([]Intent, error) {
	dir := filepath.Join(repoRoot, IntentsRelDir, bucket)
	relDir := filepath.Join(IntentsRelDir, bucket)
	di, err := os.Lstat(dir)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("intent: stat %s: %w", relDir, err)
	}
	if di.Mode()&os.ModeSymlink != 0 {
		return nil, fmt.Errorf("intent: %s is a symlink (refusing to follow)", relDir)
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("intent: reading %s: %w", relDir, err)
	}
	var intents []Intent
	for _, e := range entries {
		if e.IsDir() || !intentFileRe.MatchString(e.Name()) {
			continue
		}
		rel := filepath.Join(relDir, e.Name())
		data, err := readRepoFile(filepath.Join(dir, e.Name()), rel)
		if err != nil {
			return nil, err
		}
		it, err := parseIntent(rel, string(data), bucket)
		if err != nil {
			return nil, err
		}
		intents = append(intents, it)
	}
	return intents, nil
}

// parseIntent builds an Intent from a file's content and validates it. A file
// whose frontmatter lacks a well-formed id is malformed and rejected.
func parseIntent(relPath, content, bucket string) (Intent, error) {
	fields := frontmatter.Fields(strings.Split(content, "\n"))
	it := Intent{
		ID:           fields["id"].Value,
		Slug:         fields["slug"].Value,
		Kind:         fields["kind"].Value,
		SpecID:       fields["spec_id"].Value,
		Bucket:       bucket,
		Path:         relPath,
		PromotedFrom: fields["promoted_from"].Value,
	}
	if err := Validate(it); err != nil {
		return Intent{}, fmt.Errorf("intent: malformed %s: %w", relPath, err)
	}
	return it, nil
}

// Plan is the load-bearing verb. For a draft intent carrying a non-empty
// `## Acceptance Criteria` section, it mints a native spec, writes the intent's
// derived side of the bidirectional link (spec_id + a default kind), and moves
// the intent drafts/ → planned/.
//
// It is fail-closed: every intermediate on-disk state satisfies the
// intent_lifecycle record-lint rule, so a failure at any step leaves a
// consistent, lint-valid record rather than a half-written link. The order —
// create spec, set kind while still a draft, move to planned, then write
// spec_id — is chosen so that (kind=standalone, spec_id=null) is the only
// transient frontmatter, and that shape is valid in BOTH drafts and planned.
func Plan(repoRoot, intentID string) (PlanResult, error) {
	if !recordid.ValidIntentID(intentID) {
		return PlanResult{}, fmt.Errorf("intent: id %q must match ^itd-[0-9]+$", intentID)
	}
	corpus, err := Load(repoRoot)
	if err != nil {
		return PlanResult{}, err
	}
	it, ok := corpus.Lookup(intentID)
	if !ok {
		return PlanResult{}, fmt.Errorf("intent: %s not found in any bucket", intentID)
	}
	// A record already in planned/ takes the identity step alone. Conditions get
	// written after planning — the elicitation is a human conversation, not a
	// one-shot — and Plan is the only writer of a marker there is, so refusing the
	// re-run would leave such a record permanently unable to satisfy the gate that
	// demands the marker (iss-2608300210588874).
	if it.Bucket == BucketPlanned {
		return stampPlanned(repoRoot, it)
	}
	if it.Bucket != BucketDrafts {
		return PlanResult{}, fmt.Errorf("intent: %s is in %s, not drafts; only a draft can be planned", intentID, it.Bucket)
	}
	if !slugRe.MatchString(it.Slug) {
		return PlanResult{}, fmt.Errorf("intent: %s has slug %q which must be kebab-case", intentID, it.Slug)
	}

	draftRel := it.Path
	draftAbs := filepath.Join(repoRoot, draftRel)
	data, err := readRepoFile(draftAbs, draftRel)
	if err != nil {
		return PlanResult{}, err
	}
	content := string(data)
	if !hasAcceptanceCriteria(content) {
		return PlanResult{}, fmt.Errorf("intent: %s has no non-empty '## Acceptance Criteria' section (itd-1 discipline); refusing to plan", intentID)
	}
	// A draft that already carries a non-null spec_id is half-planned (and
	// lint-invalid): refuse rather than mint a second spec for it.
	if !frontmatter.IsNull(it.SpecID) {
		return PlanResult{}, fmt.Errorf("intent: %s is a draft with spec_id %q already set (half-planned); refusing to plan", intentID, it.SpecID)
	}

	// 1. Reuse the spec already realising this intent, or mint one. Reusing makes
	// Plan retry-safe: a re-run after a failed drafts->planned rename completes the
	// operation instead of duplicating the spec. Both branches write the reciprocal
	// intent: itd-N side (Create writes it; a reused spec already carries it).
	store, err := spec.Load(repoRoot)
	if err != nil {
		return PlanResult{}, err
	}
	sp, ok := store.ByIntent(intentID)

	// 1a. The draft face grows the record THREE times — the identity stamp, the
	// kind rewrite, the spec_id rewrite — and the largest of them is knowable
	// before any of them is written. Checking per write let a record pass the kind
	// write, MOVE to planned, and only then fail the spec_id write, leaving a
	// half-planned record neither Plan nor Link repairs; and it left a freshly
	// minted spec behind with nothing pointing at it (iss-2608300335369473). The
	// check runs first, before the spec is minted, over the id that mint WOULD
	// produce — and again below with the id it actually produced, because NextID
	// is predicted outside the mint lock and a concurrent create can push the id
	// to a wider number. Both run before the first write, so neither refusal
	// leaves a half-planned record; a spec minted and then refused is reusable,
	// since the next run finds it through ByIntent.
	specID := sp.ID
	if !ok {
		if specID, _, err = spec.NextID(repoRoot); err != nil {
			return PlanResult{}, err
		}
	}
	if err := checkDraftFaceSize(content, it, specID, draftRel); err != nil {
		return PlanResult{}, err
	}

	var mintWarning string
	if !ok {
		sp, mintWarning, err = spec.Create(repoRoot, intentID, it.Slug)
		if err != nil {
			return PlanResult{}, err
		}
		// Re-judge against the id the mint actually chose, still before any write.
		if err := checkDraftFaceSize(content, it, sp.ID, draftRel); err != nil {
			return PlanResult{}, err
		}
	}

	// 2. Stamp an identity onto every unmarked scope-condition bullet, and set the
	// binding kind (default standalone), while still in drafts. Plan is the
	// write-capable verb of the lifecycle, so it is where the identities are
	// minted: the readiness gate reports a missing marker but never writes one, a
	// reporter that writes being a reporter whose output depends on who ran it.
	// A draft with (kind=standalone, spec_id=null) stays lint-valid, so a failure
	// here leaves a consistent record (the spec exists but the intent is unlinked).
	stampedContent, conditionsStamped, err := stampScopeConditions(content, recordid.Minter{})
	if err != nil {
		return PlanResult{}, err
	}
	kind := it.Kind
	if frontmatter.IsNull(kind) {
		kind = KindStandalone
	}
	withKind, err := setFrontmatterFields(stampedContent, map[string]string{"kind": kind})
	if err != nil {
		return PlanResult{}, err
	}
	if err := writeIntentFile(draftAbs, draftRel, withKind); err != nil {
		return PlanResult{}, err
	}

	// 3. Move drafts/ → planned/ via the shared, trust-guarded move. The moved
	// file's (kind=standalone, spec_id=null) shape is a valid planned intent, so a
	// rename failure leaves a consistent state either side.
	plannedRel, err := moveIntentToBucket(repoRoot, draftRel, BucketPlanned)
	if err != nil {
		return PlanResult{}, err
	}
	plannedAbs := filepath.Join(repoRoot, plannedRel)

	// 4. Write the derived link (spec_id) now that the file is in planned. A
	// planned intent with spec_id=null is still lint-valid, so a failure here is
	// consistent too.
	withSpec, err := setFrontmatterFields(withKind, map[string]string{"spec_id": sp.ID})
	if err != nil {
		return PlanResult{}, err
	}
	if err := writeIntentFile(plannedAbs, plannedRel, withSpec); err != nil {
		return PlanResult{}, err
	}

	it.Kind = kind
	it.SpecID = sp.ID
	it.Bucket = BucketPlanned
	it.Path = plannedRel
	return PlanResult{Intent: it, Spec: sp, MintWarning: mintWarning, ConditionsStamped: conditionsStamped}, nil
}

// checkDraftFaceSize refuses a draft whose planned form would not fit under the
// cap its own reader enforces, BEFORE the first write and before the bucket
// move. It reproduces the three growth steps in order and judges the largest.
func checkDraftFaceSize(content string, it Intent, specID, rel string) error {
	// Every native id is the same width, so a fixed clock and an ADVANCING suffix
	// measure exactly what the real mint will write. The entropy has to advance:
	// a constant source hands the second bullet the id the first already used, the
	// redraw loop exhausts, and the whole judgement is lost behind a mint error on
	// every record with more than one condition (iss-2608300352403199). The source
	// is per-call, so two concurrent plans never share a counter.
	probe := recordid.Minter{
		Now:     func() time.Time { return time.Unix(0, 0).UTC() },
		Entropy: &probeEntropy{},
	}
	stamped, _, err := stampScopeConditions(content, probe)
	if err != nil {
		// Including a structural refusal: reporting it here, before the spec is
		// minted, is strictly better than letting the real stamp reach it later.
		return err
	}
	kind := it.Kind
	if frontmatter.IsNull(kind) {
		kind = KindStandalone
	}
	withKind, err := setFrontmatterFields(stamped, map[string]string{"kind": kind})
	if err != nil {
		return err
	}
	withSpec, err := setFrontmatterFields(withKind, map[string]string{"spec_id": specID})
	if err != nil {
		return err
	}
	largest := len(stamped)
	for _, n := range []int{len(withKind), len(withSpec)} {
		if n > largest {
			largest = n
		}
	}
	if largest > maxIntentFileBytes {
		return fmt.Errorf("intent: planning %s would produce %d bytes, past the %d-byte cap its own reader enforces; refusing before any write", rel, largest, maxIntentFileBytes)
	}
	return nil
}

// probeSuffixSpan keeps every probe draw below the mint's rejection band, so no
// draw is discarded and the counter advances one id per bullet. Suffixes repeat
// after 10,000 bullets — a record that long fails the size cap many times over.
const probeSuffixSpan = 50000

// probeEntropy is the size probe's entropy: an advancing counter, never
// crypto/rand. It exists so the probe's ids are distinct and sixteen digits
// wide; nothing it produces is ever stored.
type probeEntropy struct{ n uint16 }

func (e *probeEntropy) Read(p []byte) (int, error) {
	var b [2]byte
	for i := 0; i < len(p); i += 2 {
		e.n = (e.n + 1) % probeSuffixSpan
		binary.BigEndian.PutUint16(b[:], e.n)
		copy(p[i:], b[:])
	}
	return len(p), nil
}

// stampPlanned is Plan's idempotent second face// stampPlanned is Plan's idempotent second face: it mints an identity for every
// unmarked scope-condition bullet of an already-planned record and writes it
// back, touching nothing else — no spec, no frontmatter, no bucket move. An
// already-marked bullet is left byte-identical, so re-running after an edit
// stamps only what is new.
//
// A run with nothing to stamp is a refusal, not a quiet success: the caller
// asked for work to be done, and a verb that exits 0 having done none of it
// teaches its user that the command is a no-op.
func stampPlanned(repoRoot string, it Intent) (PlanResult, error) {
	rel := it.Path
	abs := filepath.Join(repoRoot, rel)
	var stampedCount int
	// The read, the mint and the write are one critical section under the store's
	// existing advisory lock: two sessions stamping the same record would
	// otherwise each write the file they read, and the later write would drop the
	// earlier one's identities (iss-2608300235388164).
	err := withIntentMintLock(repoRoot, func() error {
		data, err := readRepoFile(abs, rel)
		if err != nil {
			return err
		}
		stamped, n, err := stampScopeConditions(string(data), recordid.Minter{})
		if err != nil {
			return err
		}
		if n == 0 {
			return fmt.Errorf("intent: %s is already planned and carries no unmarked scope condition; nothing to stamp", it.ID)
		}
		if err := writeIntentFile(abs, rel, stamped); err != nil {
			return err
		}
		stampedCount = n
		return nil
	})
	if err != nil {
		return PlanResult{}, err
	}
	res := PlanResult{Intent: it, ConditionsStamped: stampedCount, StampOnly: true}
	// The stamp mints no spec, but the intent has one, and an empty spec object in
	// the result reads as "this intent has no spec" to anything consuming it. The
	// lookup is lenient: a broken link is the readiness gate's finding to report,
	// not a reason to fail a write that already succeeded.
	if store, lerr := spec.Load(repoRoot); lerr == nil {
		if sp, ok := store.Lookup(it.SpecID); ok {
			res.Spec = sp
		}
	}
	return res, nil
}

// Link retroactively writes the derived spec_id link on an existing planned
// intent for an existing spec. It validates both ids, that the intent is in
// planned/, and that the spec exists AND already declares this intent (the
// reciprocal intent: itd-N side); a spec that realises a different intent is a
// mismatch and fails closed rather than forging a one-sided link.
func Link(repoRoot, intentID, specID string) (LinkResult, error) {
	if !recordid.ValidIntentID(intentID) {
		return LinkResult{}, fmt.Errorf("intent: id %q must match ^itd-[0-9]+$", intentID)
	}
	if !recordid.ValidSpecID(specID) {
		return LinkResult{}, fmt.Errorf("intent: spec id %q must match ^spc-[0-9]+$", specID)
	}
	corpus, err := Load(repoRoot)
	if err != nil {
		return LinkResult{}, err
	}
	it, ok := corpus.Lookup(intentID)
	if !ok {
		return LinkResult{}, fmt.Errorf("intent: %s not found in any bucket", intentID)
	}
	if it.Bucket != BucketPlanned {
		return LinkResult{}, fmt.Errorf("intent: %s is in %s, not planned; only a planned intent can be linked", intentID, it.Bucket)
	}
	store, err := spec.Load(repoRoot)
	if err != nil {
		return LinkResult{}, err
	}
	sp, ok := store.Lookup(specID)
	if !ok {
		return LinkResult{}, fmt.Errorf("intent: spec %s not found", specID)
	}
	if sp.Intent != intentID {
		return LinkResult{}, fmt.Errorf("intent: spec %s realises %s, not %s (mismatch); refusing to link", specID, sp.Intent, intentID)
	}

	rel := it.Path
	abs := filepath.Join(repoRoot, rel)
	data, err := readRepoFile(abs, rel)
	if err != nil {
		return LinkResult{}, err
	}
	updated, err := setFrontmatterFields(string(data), map[string]string{"spec_id": specID})
	if err != nil {
		return LinkResult{}, err
	}
	if err := writeIntentFile(abs, rel, updated); err != nil {
		return LinkResult{}, err
	}

	it.SpecID = specID
	return LinkResult{Intent: it, Spec: sp}, nil
}

// Reconcile is the deterministic half of `abcd spec close`: it advances the
// intent a spec realises, then closes the spec, so one command marks the spec
// done AND ships its linked intent.
//
// Ordering is intent-first, spec-last, so a partial failure is recoverable by
// re-running: the intent moves planned/ → shipped/ before spec.Close runs, so a
// failure at the move leaves the spec OPEN (retry-safe), never a closed spec with
// a still-planned intent. It is idempotent: an already-shipped intent is not
// re-moved, and a re-run on an already-closed spec is a clean no-op/complete
// rather than an error.
//
// It fails closed with NO partial move when: the spec has no/empty intent link;
// the named intent does not exist; the link is ambiguous (more than one spec
// realises the intent); the intent's spec_id disagrees with this spec
// (bidirectional drift); or the intent is in an unexpected bucket (e.g. still in
// drafts — it was never planned). Every id is validated against the ^spc-/^itd-
// regexes before any path is built. The intent's `## Audit Notes` are left
// untouched (the fidelity audit is a later phase; the intent ships with them empty).
func Reconcile(repoRoot, specID string) (ReconcileResult, error) {
	if !recordid.ValidSpecID(specID) {
		return ReconcileResult{}, fmt.Errorf("intent: spec id %q must match ^spc-[0-9]+$", specID)
	}
	store, err := spec.Load(repoRoot)
	if err != nil {
		return ReconcileResult{}, err
	}
	sp, ok := store.Lookup(specID)
	if !ok {
		return ReconcileResult{}, fmt.Errorf("intent: spec %s not found", specID)
	}

	// Resolve the linked intent from the spec's intent: field, validated before it
	// is ever used to build a path.
	intentID := sp.Intent
	if !recordid.ValidIntentID(intentID) {
		return ReconcileResult{}, fmt.Errorf("intent: spec %s has no well-formed intent link (got %q); refusing to reconcile", specID, intentID)
	}
	// Ambiguity guard: cross-check the spec's link against the whole store. If more
	// than one spec claims this intent, the link is ambiguous and we refuse rather
	// than ship an intent whose realising spec is undetermined.
	var claimers []string
	for _, s := range store.Specs {
		if s.Intent == intentID {
			claimers = append(claimers, s.ID)
		}
	}
	if len(claimers) > 1 {
		return ReconcileResult{}, fmt.Errorf("intent: link ambiguous — %d specs realise %s (%s); refusing to reconcile", len(claimers), intentID, strings.Join(claimers, ", "))
	}

	corpus, err := Load(repoRoot)
	if err != nil {
		return ReconcileResult{}, err
	}
	it, ok := corpus.Lookup(intentID)
	if !ok {
		return ReconcileResult{}, fmt.Errorf("intent: %s (linked by spec %s) not found in any bucket; refusing to reconcile", intentID, specID)
	}
	// Bidirectional agreement: the intent must point back at THIS spec. A null or
	// mismatched spec_id is drift (a one-sided link) — fail closed rather than ship
	// an intent that names a different, or no, spec.
	// The comparison is canonical (spec.SameNum), not literal: record-lint matches
	// a spec_id on its NUMBER, so a slug-suffixed or zero-padded value is
	// lint-green and this verb must not refuse what the lint accepts.
	if !spec.SameNum(it.SpecID, specID) {
		return ReconcileResult{}, fmt.Errorf("intent: %s spec_id is %q but spec %s claims it (bidirectional link disagrees); refusing to reconcile", intentID, it.SpecID, specID)
	}
	// Bucket guard runs BEFORE any move, so an unexpected bucket (drafts,
	// disciplines, superseded) yields no partial move.
	switch it.Bucket {
	case BucketPlanned, BucketShipped:
		// planned → advance; shipped → idempotent (already advanced).
	default:
		return ReconcileResult{}, fmt.Errorf("intent: %s is in %s (linked by spec %s); expected planned or shipped — refusing to reconcile", intentID, it.Bucket, specID)
	}

	res := ReconcileResult{Spec: sp, Intent: it, From: it.Bucket, To: it.Bucket}
	// 1. Advance the intent planned/ → shipped/ FIRST. Its (kind, spec_id) are
	// already set (Plan wrote them), so the shipped record is lint-valid without
	// touching frontmatter. If this fails, the spec stays open — the whole
	// operation retries cleanly.
	if it.Bucket == BucketPlanned {
		dstRel, err := moveIntentToBucket(repoRoot, it.Path, BucketShipped)
		if err != nil {
			return ReconcileResult{}, err
		}
		it.Bucket = BucketShipped
		it.Path = dstRel
		res.Intent = it
		res.IntentMoved = true
		res.To = BucketShipped
	}

	// 2. Close the spec, but only if still open — a re-run on an already-closed
	// spec is a clean completion, not the "already closed" error spec.Close raises.
	if sp.Status == spec.StatusOpen {
		closed, err := spec.Close(repoRoot, specID)
		if err != nil {
			return ReconcileResult{}, err
		}
		res.Spec = closed
	}

	// 3. Emit the fidelity-review OWED stub + ephemeral request over the shipped
	// intent. The review is REPORT-ONLY: a failure here is captured, not raised —
	// the intent has already shipped and must not be un-shipped by a review-emit
	// error. The surface prints AuditEmitError loudly.
	if res.Intent.Bucket == BucketShipped {
		emit, err := emitAuditForIntent(repoRoot, res.Intent)
		if err != nil {
			res.AuditEmitError = err.Error()
		} else {
			res.ReceiptID = emit.ReceiptID
		}
	}
	return res, nil
}

// writeIntentFile is the one way this package writes an intent record — the
// seed, the lifecycle rewrites and the audit-block upserts all go through it,
// so the claim is a fact rather than a description of most of them. It caps
// the FINAL bytes before writing them, because the cap belongs at the write and
// not at any one producer: readRepoFile refuses a record over the cap, and every
// verb here loads the WHOLE corpus first, so one oversized record makes every
// intent command refuse every record until a human trims the file by hand. Each
// write is preceded by a growth step — an identity stamp, a `kind` rewrite, a
// `spec_id` rewrite — and guarding only the first of them left the others free to
// carry a record over the line (iss-2608300318192814).
//
// rel is repo-relative: an intent error never carries an absolute local path.
func writeIntentFile(abs, rel, content string) error {
	if len(content) > maxIntentFileBytes {
		return fmt.Errorf("intent: writing %s would produce %d bytes, past the %d-byte cap its own reader enforces; refusing", rel, len(content), maxIntentFileBytes)
	}
	if err := fsutil.WriteFileAtomic(abs, []byte(content), 0o644); err != nil {
		return fmt.Errorf("intent: writing %s: %w", rel, err)
	}
	return nil
}

// moveIntentToBucket moves the intent file at srcRel into dstBucket via os.Rename
// (atomic on one filesystem), behind the store's trust guards: it refuses to
// follow a symlinked destination directory and refuses to clobber an existing
// destination file. It returns the new repo-relative path. This is the single
// canonical intent move, shared by Plan (drafts → planned) and Reconcile
// (planned → shipped).
func moveIntentToBucket(repoRoot, srcRel, dstBucket string) (string, error) {
	name := filepath.Base(srcRel)
	dstRelDir := filepath.Join(IntentsRelDir, dstBucket)
	dstRel := filepath.Join(dstRelDir, name)
	dstAbs := filepath.Join(repoRoot, dstRel)
	if err := ensureRealDir(filepath.Join(repoRoot, dstRelDir), dstRelDir); err != nil {
		return "", err
	}
	if _, err := os.Lstat(dstAbs); err == nil {
		return "", fmt.Errorf("intent: refusing to overwrite existing %s", dstRel)
	}
	srcBucket := filepath.Base(filepath.Dir(srcRel))
	if err := os.Rename(filepath.Join(repoRoot, srcRel), dstAbs); err != nil {
		return "", fmt.Errorf("intent: moving %s %s->%s: %w", name, srcBucket, dstBucket, err)
	}
	return dstRel, nil
}

// Status builds the read-only lifecycle summary: intent counts by bucket, spec
// counts by status, and the intent↔spec links (every intent whose spec_id is
// non-null). Linked pairs are ordered by the corpus load order (bucket, then
// directory), which is deterministic.
func Status(repoRoot string) (StatusView, error) {
	corpus, err := Load(repoRoot)
	if err != nil {
		return StatusView{}, err
	}
	store, err := spec.Load(repoRoot)
	if err != nil {
		return StatusView{}, err
	}

	v := StatusView{Buckets: map[string]int{}, Linked: []LinkedPair{}}
	for _, b := range Buckets {
		v.Buckets[b] = 0
	}
	for _, it := range corpus.Intents {
		v.Buckets[it.Bucket]++
		if !frontmatter.IsNull(it.SpecID) {
			v.Linked = append(v.Linked, LinkedPair{Intent: it.ID, Spec: it.SpecID})
		}
	}
	for _, sp := range store.Specs {
		if sp.Status == spec.StatusClosed {
			v.SpecsClosed++
		} else {
			v.SpecsOpen++
		}
	}
	return v, nil
}

// readRepoFile reads a repo file behind the trust-boundary guards: refuse a
// symlinked leaf, require a regular file, and cap the size. (Mirrors the spec
// store's private guard; a shared read-guard is a flagged consolidation target
// alongside ensureRealDir.)
func readRepoFile(abs, rel string) ([]byte, error) {
	fi, err := os.Lstat(abs)
	if err != nil {
		return nil, fmt.Errorf("intent: stat %s: %w", rel, err)
	}
	if fi.Mode()&os.ModeSymlink != 0 {
		return nil, fmt.Errorf("intent: %s is a symlink (refusing to follow)", rel)
	}
	if !fi.Mode().IsRegular() {
		return nil, fmt.Errorf("intent: %s is not a regular file", rel)
	}
	if fi.Size() > maxIntentFileBytes {
		return nil, fmt.Errorf("intent: %s exceeds the %d-byte cap", rel, maxIntentFileBytes)
	}
	data, err := os.ReadFile(abs)
	if err != nil {
		return nil, fmt.Errorf("intent: reading %s: %w", rel, err)
	}
	return data, nil
}

// ensureRealDir creates dir if absent, refusing a symlinked leaf directory.
// NOTE: a symlinked ANCESTOR (e.g. a symlinked intents/) is not caught here — a
// low-severity follow-up under the trusted-worktree model (planting one needs
// write access equal to editing the record directly).
func ensureRealDir(dir, rel string) error {
	if di, err := os.Lstat(dir); err == nil && di.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("intent: %s is a symlink (refusing to follow)", rel)
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("intent: creating %s: %w", rel, err)
	}
	return nil
}
