package intent

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/intentdriven/abcd/internal/core/changelog"
	"github.com/intentdriven/abcd/internal/core/recordid"
	"github.com/intentdriven/abcd/internal/fsutil"
)

// mintLockTimeout bounds how long CreateFromText waits for the intent-store mint
// lock. A var (not const) so a test can shorten it to exercise contention.
var mintLockTimeout = 5 * time.Second

// intentNumRe extracts N from an itd-N-<slug>.md filename (the allocator scan).
var intentNumRe = regexp.MustCompile(`^itd-([0-9]+)(?:-[a-z0-9-]+)?\.md$`)

// maxSlugLen caps a derived slug so a pathological free-text line cannot produce
// an unwieldy filename. Mirrors the capture-side derivation budget.
const maxSlugLen = 60

// CreateFromText files a new draft intent seeded from free-form text, mirroring
// the capture engine's create shape: it derives a filename-safe slug, mints the
// next itd-N under the exclusive store mint lock (so two concurrent sessions
// never mint the same id), and atomically writes drafts/itd-N-<slug>.md with the
// canonical draft frontmatter set and a minimal, honest body skeleton carrying
// the text. Empty/whitespace text is refused and nothing is written.
//
// The seeded record is lint-valid (intent_lifecycle accepts a draft whose kind is
// null and whose spec_id is null) and passes Validate; a human expands it, then
// `abcd intent plan` schedules it. This is the quoted-text create path itd-46
// delivers — the create half of what spc-6 AC3 (promote) needs.
func CreateFromText(repoRoot, text, impact string) (Intent, string, error) {
	trimmed := strings.TrimSpace(text)
	if trimmed == "" {
		return Intent{}, "", fmt.Errorf("intent: refusing to create from empty text")
	}
	// Redact the caller's text through the one canonical scanner BEFORE anything
	// derived from it is built (gh-486). The slug becomes the filename and is
	// derived straight from this text, so a secret/home-path in the prose would
	// otherwise reach the committed filename too (the capture-slug leak shape):
	// redacting first and deriving the slug from the redacted text is what keeps a
	// finding out of the path as well as the body. CreateDraft redacts title and
	// body again as its own boundary guard; that second pass is idempotent here.
	redacted, _, err := redactIntentText(repoRoot, trimmed)
	if err != nil {
		return Intent{}, "", err
	}
	slug, err := deriveIntentSlug(redacted)
	if err != nil {
		return Intent{}, "", err
	}
	return CreateDraft(repoRoot, DraftOptions{
		Slug:     slug,
		Title:    titleLine(redacted),
		SeedBody: redacted,
		Impact:   impact,
	})
}

// DraftOptions parameterises CreateDraft: the explicit slug and title, the Why
// This Matters seed body, an optional impact judgement, and — on the promote
// path (spc-24) — the iss-N the draft graduated from, written as the
// promoted_from back-edge (an iss-N, or the rdi-N of a dispositioned reading
// item).
type DraftOptions struct {
	Slug         string
	Title        string
	SeedBody     string
	Impact       string
	PromotedFrom string
}

// promotedFromRe constrains the promote back-edge to a captured id before it is
// written into frontmatter. Two families graduate into an intent: an issue
// somebody noticed (iss-N) and a reading item an instrument returned and a
// researcher dispositioned (rdi-N). Both are one act — an observation earning an
// admission — so both write the same back edge, and the forward stamp on the
// source record is what closes the join.
var promotedFromRe = regexp.MustCompile(`^(iss|rdi)-[0-9]+$`)

// CreateDraft is the one canonical draft-mint primitive: both the quoted-text
// create (CreateFromText) and the capture-promote path (capture.Promote, which
// supplies an explicit slug and seed body) route through it, so a draft can
// never be minted outside the store mint lock. It validates every option at the
// boundary — the slug becomes a filename — mints the next itd-N, and atomically
// writes drafts/itd-N-<slug>.md. On any refusal nothing is written.
func CreateDraft(repoRoot string, opts DraftOptions) (Intent, string, error) {
	if !slugRe.MatchString(opts.Slug) {
		return Intent{}, "", fmt.Errorf("intent: slug %q is not kebab-case", opts.Slug)
	}
	if strings.TrimSpace(opts.Title) == "" {
		return Intent{}, "", fmt.Errorf("intent: refusing to create a draft with an empty title")
	}
	if strings.TrimSpace(opts.SeedBody) == "" {
		return Intent{}, "", fmt.Errorf("intent: refusing to create a draft with an empty seed body")
	}
	if opts.PromotedFrom != "" && !promotedFromRe.MatchString(opts.PromotedFrom) {
		return Intent{}, "", fmt.Errorf("intent: promoted_from %q must match ^(iss|rdi)-[0-9]+$", opts.PromotedFrom)
	}
	// impact is optional on a draft (intent_impact_valid gates the move into
	// shipped/, not the seed), but when set it must be a legal, non-internal
	// judgement — the same bar the gate applies at shipped/ — so the value the
	// tool stamps travels unchanged to shipped/ and passes the blocker there. An
	// invalid or internal impact is refused up front, never absorbed.
	if opts.Impact != "" {
		imp, err := changelog.ParseImpact(opts.Impact)
		if err != nil {
			return Intent{}, "", fmt.Errorf("intent: %w", err)
		}
		if imp == changelog.ImpactInternal {
			return Intent{}, "", fmt.Errorf("intent: impact must not be internal on an intent — a press-release-first intent is user-facing by definition; declare one of additive|breaking|fix, or record the work as an issue instead")
		}
	}

	// Boundary redaction (gh-486): CreateDraft is the ONE canonical draft-mint
	// primitive, so the "no unredacted secret/home-path in a committed intent"
	// invariant lives HERE, where every caller (quoted-text create AND
	// capture-promote) passes through it — never once per caller, where a future
	// caller would reopen the leak. Title and SeedBody are the two free-text
	// members; the structural fields (slug, ids, impact) are validated/enum-
	// constrained above and carry nothing to redact. Fail-closed on a degraded
	// scanner; the pass is idempotent for text a caller already redacted.
	rTitle, _, err := redactIntentText(repoRoot, opts.Title)
	if err != nil {
		return Intent{}, "", err
	}
	rBody, _, err := redactIntentText(repoRoot, opts.SeedBody)
	if err != nil {
		return Intent{}, "", err
	}
	opts.Title = rTitle
	opts.SeedBody = rBody

	var created Intent
	var mintWarning string
	err = withIntentMintLock(repoRoot, func() error {
		id, warn, err := nextIntentID(repoRoot)
		if err != nil {
			return err
		}
		mintWarning = warn
		draftsDirAbs := filepath.Join(repoRoot, IntentsRelDir, BucketDrafts)
		if err := ensureRealDir(draftsDirAbs, filepath.Join(IntentsRelDir, BucketDrafts)); err != nil {
			return err
		}
		name := id + "-" + opts.Slug + ".md"
		rel := filepath.Join(IntentsRelDir, BucketDrafts, name)
		abs := filepath.Join(draftsDirAbs, name)
		// Refuse to clobber an existing draft (best-effort guard under the lock).
		if _, statErr := os.Lstat(abs); statErr == nil {
			return fmt.Errorf("intent: refusing to overwrite existing %s", rel)
		}
		content := seedDraft(id, opts)
		if err := fsutil.WriteFileAtomic(abs, []byte(content), 0o644); err != nil {
			return fmt.Errorf("intent: writing %s: %w", rel, err)
		}
		created = Intent{
			ID:           id,
			Slug:         opts.Slug,
			Kind:         "null",
			SpecID:       "null",
			Bucket:       BucketDrafts,
			Path:         rel,
			PromotedFrom: opts.PromotedFrom,
		}
		return nil
	})
	if err != nil {
		return Intent{}, "", err
	}
	return created, mintWarning, Validate(created)
}

// deriveIntentSlug lowercases the text, collapses non-[a-z0-9] runs to a single
// hyphen, trims leading/trailing hyphens, caps the length, and insists the result
// is kebab-case and non-empty — the slug becomes a filename, so it is validated
// before any path is built (path-traversal / filename-safety defence).
func deriveIntentSlug(text string) (string, error) {
	lowered := strings.ToLower(text)
	collapsed := strings.Trim(slugNonAlnumRe.ReplaceAllString(lowered, "-"), "-")
	if len(collapsed) > maxSlugLen {
		collapsed = strings.Trim(collapsed[:maxSlugLen], "-")
	}
	if collapsed == "" {
		return "", fmt.Errorf("intent: text %q has no slug-able characters", text)
	}
	if !slugRe.MatchString(collapsed) {
		return "", fmt.Errorf("intent: derived slug %q is not kebab-case", collapsed)
	}
	return collapsed, nil
}

var slugNonAlnumRe = regexp.MustCompile(`[^a-z0-9]+`)

// nextIntentID returns the next free itd-N: max N over every intent file in every
// bucket AND over intent filenames on every other git ref, plus one. Called under
// the mint lock so the working-tree scan and the subsequent write are one critical
// section (no two concurrent creates in the same worktree observe the same max).
// Folding in recordid.MaxAcrossRefs is what stops two parallel branches from
// re-minting the same itd-N once one has committed it (iss-115, iss-120). The
// returned mintWarning is non-empty (and MUST be surfaced) when the ref scan
// degraded to working-tree-only, so the fallback is never silent.
func nextIntentID(repoRoot string) (id, mintWarning string, err error) {
	max := 0
	for _, bucket := range Buckets {
		dir := filepath.Join(repoRoot, IntentsRelDir, bucket)
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue // absent bucket is soft
		}
		for _, e := range entries {
			if e.IsDir() {
				continue
			}
			m := intentNumRe.FindStringSubmatch(e.Name())
			if m == nil {
				continue
			}
			n, err := strconv.Atoi(m[1])
			if err != nil {
				continue
			}
			if n > max {
				max = n
			}
		}
	}
	scan := recordid.MaxAcrossRefs(repoRoot, "itd", []string{IntentsRelDir})
	if scan.Max > max {
		max = scan.Max
	}
	// Guard the max+1 below against int overflow: a hand-crafted MaxInt itd-N
	// (a local file or a fetched remote-tracking ref carrying itd-<MaxInt>-x.md)
	// parses to math.MaxInt with no error, so max+1 would wrap to math.MinInt and
	// mint itd--9223372036854775808 — a malformed draft WriteFileAtomic persists
	// before Validate runs, plus a mint DoS for the family. Refuse clearly instead,
	// mirroring the capture allocator's ceiling guard.
	if max >= math.MaxInt {
		return "", "", fmt.Errorf("intent: itd-N counter near the integer ceiling (highest observed %d); refusing to allocate", max)
	}
	return fmt.Sprintf("itd-%d", max+1), scan.Warning(), nil
}

// seedDraft renders the canonical draft skeleton: the full draft frontmatter set
// (id, slug, spec_id: null, kind: null, suggested_kind: null,
// reclassification_history: [], builds_on: [], severity: minor, plus the
// promoted_from back-edge when the draft graduated from an issue) and an honest,
// minimal body carrying the seed text under Why This Matters, with the itd-1
// discipline's Acceptance Criteria section left as a placeholder for the human to
// fill before planning.
func seedDraft(id string, opts DraftOptions) string {
	var b strings.Builder
	b.WriteString("---\n")
	b.WriteString("id: " + id + "\n")
	b.WriteString("slug: " + opts.Slug + "\n")
	b.WriteString("spec_id: null\n")
	b.WriteString("kind: null\n")
	b.WriteString("suggested_kind: null\n")
	b.WriteString("reclassification_history: []\n")
	b.WriteString("builds_on: []\n")
	b.WriteString("severity: minor\n")
	// The promote back-edge (spc-24): bare, like every id field in this store.
	// Absent on a quoted-text draft — the field exists only when the draft
	// graduated from an issue.
	if opts.PromotedFrom != "" {
		b.WriteString("promoted_from: " + opts.PromotedFrom + "\n")
	}
	// impact is written only when the caller declared one (validated in
	// CreateDraft). It is bare — the machine-read enum the shipped-intent gate
	// compares byte-for-byte — and travels unchanged to shipped/. An unset impact
	// writes no line: a draft is "not judged yet", exactly like the null fields.
	if opts.Impact != "" {
		b.WriteString("impact: " + opts.Impact + "\n")
	}
	b.WriteString("---\n\n")
	b.WriteString("# " + opts.Title + "\n\n")
	b.WriteString("## Press Release\n\n")
	b.WriteString("> " + seedNote(opts) + "\n\n")
	b.WriteString("## Why This Matters\n\n")
	b.WriteString(opts.SeedBody + "\n\n")
	b.WriteString("## Acceptance Criteria\n\n")
	b.WriteString("> _Required (the itd-1 discipline): add at least one Given-When-Then bullet describing the verifiable bar for \"shipped\" before this draft can be planned._\n\n")
	b.WriteString("## Open Questions\n\n")
	b.WriteString("_None recorded yet._\n\n")
	b.WriteString("## Audit Notes\n\n")
	b.WriteString("_Empty. Populated by intent-auditor when intent moves to shipped/._\n")
	return b.String()
}

// The two Press Release placeholders this package mints, in their parts: a
// per-path opening clause and the instruction both of them close with.
const (
	captureSeedOpening   = "Seeded from a quoted-text intent capture."
	promotionSeedOpening = "Seeded by promotion from "
	seedNoteTail         = "Expand into the full press-release narrative before planning."
)

// CaptureSeedNote is the placeholder a quoted-text capture mints, whole.
const CaptureSeedNote = captureSeedOpening + " " + seedNoteTail

// IsSeedNote reports whether a press-release body is still one of the templates
// this package writes, rather than something somebody wrote.
//
// It lives HERE, with the templates, because it is the only place that can stay
// true: a consumer comparing against its own copy of the sentence would keep
// matching the old wording the moment these are reworded, and the symptom would
// be a placeholder quoted at a reader as a testimonial. Both minted forms are
// covered and nothing else is — the promotion form's source id is the only part
// allowed to vary, and a body with a single written sentence in it fails.
//
// The caller passes the body already reduced to its words: quote markers,
// emphasis markers and whitespace runs removed.
func IsSeedNote(text string) bool {
	if text == CaptureSeedNote {
		return true
	}
	return strings.HasPrefix(text, promotionSeedOpening) && strings.HasSuffix(text, seedNoteTail)
}

// seedNote is the standard Press Release placeholder, honest about which create
// path seeded the draft.
func seedNote(opts DraftOptions) string {
	if opts.PromotedFrom != "" {
		return "_" + promotionSeedOpening + opts.PromotedFrom + ". " + seedNoteTail + "_"
	}
	return "_" + CaptureSeedNote + "_"
}

// titleLine collapses internal whitespace and trims the seed text into a single
// heading line (a multi-word free-text line becomes one clean title).
func titleLine(text string) string {
	return strings.Join(strings.Fields(text), " ")
}

// withIntentMintLock runs fn while holding an exclusive advisory lock over the
// intent store, serializing id minting across concurrent abcd processes in the
// same worktree. It flocks the intents/ directory file descriptor itself, so no
// lock artifact is left in the committed record tree (mirroring the spec store's
// mint lock). O_NOFOLLOW refuses a symlinked intents/.
func withIntentMintLock(repoRoot string, fn func() error) error {
	intentsDir := filepath.Join(repoRoot, IntentsRelDir)
	if err := ensureRealDir(intentsDir, IntentsRelDir); err != nil {
		return err
	}
	fd, err := syscall.Open(intentsDir, syscall.O_RDONLY|syscall.O_DIRECTORY|syscall.O_NOFOLLOW, 0)
	if err != nil {
		return fmt.Errorf("intent: opening mint lock on %s: %w", IntentsRelDir, err)
	}
	defer syscall.Close(fd)

	deadline := time.Now().Add(mintLockTimeout)
	for {
		lockErr := syscall.Flock(fd, syscall.LOCK_EX|syscall.LOCK_NB)
		if lockErr == nil {
			break
		}
		if lockErr != syscall.EWOULDBLOCK {
			return fmt.Errorf("intent: acquiring mint lock: %w", lockErr)
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("intent: could not acquire mint lock within %s", mintLockTimeout)
		}
		time.Sleep(10 * time.Millisecond)
	}
	defer syscall.Flock(fd, syscall.LOCK_UN)

	return fn()
}
