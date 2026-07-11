package intent

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/REPPL/abcd-cli/internal/core/frontmatter"
	"github.com/REPPL/abcd-cli/internal/core/spec"
	"github.com/REPPL/abcd-cli/internal/fsutil"
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
		ID:     fields["id"].Value,
		Slug:   fields["slug"].Value,
		Kind:   fields["kind"].Value,
		SpecID: fields["spec_id"].Value,
		Bucket: bucket,
		Path:   relPath,
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
	if !intentIDRe.MatchString(intentID) {
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
	if !ok {
		sp, err = spec.Create(repoRoot, intentID, it.Slug)
		if err != nil {
			return PlanResult{}, err
		}
	}

	// 2. Set the binding kind (default standalone) while still in drafts. A draft
	// with (kind=standalone, spec_id=null) stays lint-valid, so a failure here
	// leaves a consistent record (the spec exists but the intent is unlinked).
	kind := it.Kind
	if frontmatter.IsNull(kind) {
		kind = KindStandalone
	}
	withKind, err := setFrontmatterFields(content, map[string]string{"kind": kind})
	if err != nil {
		return PlanResult{}, err
	}
	if err := fsutil.WriteFileAtomic(draftAbs, []byte(withKind), 0o644); err != nil {
		return PlanResult{}, fmt.Errorf("intent: writing kind to %s: %w", draftRel, err)
	}

	// 3. Move drafts/ → planned/. The moved file's (kind=standalone,
	// spec_id=null) shape is a valid planned intent, so a rename failure leaves a
	// consistent state either side.
	name := filepath.Base(draftRel)
	plannedRel := filepath.Join(IntentsRelDir, BucketPlanned, name)
	plannedAbs := filepath.Join(repoRoot, plannedRel)
	if err := ensureRealDir(filepath.Join(repoRoot, IntentsRelDir, BucketPlanned), filepath.Join(IntentsRelDir, BucketPlanned)); err != nil {
		return PlanResult{}, err
	}
	if _, err := os.Lstat(plannedAbs); err == nil {
		return PlanResult{}, fmt.Errorf("intent: refusing to overwrite existing %s", plannedRel)
	}
	if err := os.Rename(draftAbs, plannedAbs); err != nil {
		return PlanResult{}, fmt.Errorf("intent: moving %s drafts->planned: %w", intentID, err)
	}

	// 4. Write the derived link (spec_id) now that the file is in planned. A
	// planned intent with spec_id=null is still lint-valid, so a failure here is
	// consistent too.
	withSpec, err := setFrontmatterFields(withKind, map[string]string{"spec_id": sp.ID})
	if err != nil {
		return PlanResult{}, err
	}
	if err := fsutil.WriteFileAtomic(plannedAbs, []byte(withSpec), 0o644); err != nil {
		return PlanResult{}, fmt.Errorf("intent: writing spec_id to %s: %w", plannedRel, err)
	}

	it.Kind = kind
	it.SpecID = sp.ID
	it.Bucket = BucketPlanned
	it.Path = plannedRel
	return PlanResult{Intent: it, Spec: sp}, nil
}

// Link retroactively writes the derived spec_id link on an existing planned
// intent for an existing spec. It validates both ids, that the intent is in
// planned/, and that the spec exists AND already declares this intent (the
// reciprocal intent: itd-N side); a spec that realises a different intent is a
// mismatch and fails closed rather than forging a one-sided link.
func Link(repoRoot, intentID, specID string) (LinkResult, error) {
	if !intentIDRe.MatchString(intentID) {
		return LinkResult{}, fmt.Errorf("intent: id %q must match ^itd-[0-9]+$", intentID)
	}
	if !specIDRe.MatchString(specID) {
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
	if err := fsutil.WriteFileAtomic(abs, []byte(updated), 0o644); err != nil {
		return LinkResult{}, fmt.Errorf("intent: writing spec_id to %s: %w", rel, err)
	}

	it.SpecID = specID
	return LinkResult{Intent: it, Spec: sp}, nil
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

	v := StatusView{Buckets: map[string]int{}}
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

// ensureRealDir creates dir if absent, refusing to follow a symlinked directory.
func ensureRealDir(dir, rel string) error {
	if di, err := os.Lstat(dir); err == nil && di.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("intent: %s is a symlink (refusing to follow)", rel)
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("intent: creating %s: %w", rel, err)
	}
	return nil
}
