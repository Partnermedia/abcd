package intent

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/REPPL/abcd-cli/internal/fsutil"
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
func CreateFromText(repoRoot, text string) (Intent, error) {
	trimmed := strings.TrimSpace(text)
	if trimmed == "" {
		return Intent{}, fmt.Errorf("intent: refusing to create from empty text")
	}
	slug, err := deriveIntentSlug(trimmed)
	if err != nil {
		return Intent{}, err
	}

	var created Intent
	err = withIntentMintLock(repoRoot, func() error {
		id, err := nextIntentID(repoRoot)
		if err != nil {
			return err
		}
		draftsDirAbs := filepath.Join(repoRoot, IntentsRelDir, BucketDrafts)
		if err := ensureRealDir(draftsDirAbs, filepath.Join(IntentsRelDir, BucketDrafts)); err != nil {
			return err
		}
		name := id + "-" + slug + ".md"
		rel := filepath.Join(IntentsRelDir, BucketDrafts, name)
		abs := filepath.Join(draftsDirAbs, name)
		// Refuse to clobber an existing draft (best-effort guard under the lock).
		if _, statErr := os.Lstat(abs); statErr == nil {
			return fmt.Errorf("intent: refusing to overwrite existing %s", rel)
		}
		content := seedDraft(id, slug, trimmed)
		if err := fsutil.WriteFileAtomic(abs, []byte(content), 0o644); err != nil {
			return fmt.Errorf("intent: writing %s: %w", rel, err)
		}
		created = Intent{
			ID:     id,
			Slug:   slug,
			Kind:   "null",
			SpecID: "null",
			Bucket: BucketDrafts,
			Path:   rel,
		}
		return nil
	})
	if err != nil {
		return Intent{}, err
	}
	return created, Validate(created)
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
// bucket, plus one. Called under the mint lock so the scan and the subsequent
// write are one critical section (no two concurrent creates observe the same max).
func nextIntentID(repoRoot string) (string, error) {
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
	return fmt.Sprintf("itd-%d", max+1), nil
}

// seedDraft renders the canonical draft skeleton: the full draft frontmatter set
// (id, slug, spec_id: null, kind: null, suggested_kind: null,
// reclassification_history: [], builds_on: [], severity: minor) and an honest,
// minimal body carrying the seed text under Why This Matters, with the itd-1
// discipline's Acceptance Criteria section left as a placeholder for the human to
// fill before planning.
func seedDraft(id, slug, text string) string {
	var b strings.Builder
	b.WriteString("---\n")
	b.WriteString("id: " + id + "\n")
	b.WriteString("slug: " + slug + "\n")
	b.WriteString("spec_id: null\n")
	b.WriteString("kind: null\n")
	b.WriteString("suggested_kind: null\n")
	b.WriteString("reclassification_history: []\n")
	b.WriteString("builds_on: []\n")
	b.WriteString("severity: minor\n")
	b.WriteString("---\n\n")
	b.WriteString("# " + titleLine(text) + "\n\n")
	b.WriteString("## Press Release\n\n")
	b.WriteString("> _Seeded from a quoted-text intent capture. Expand into the full press-release narrative before planning._\n\n")
	b.WriteString("## Why This Matters\n\n")
	b.WriteString(text + "\n\n")
	b.WriteString("## Acceptance Criteria\n\n")
	b.WriteString("> _Required (the itd-1 discipline): add at least one Given-When-Then bullet describing the verifiable bar for \"shipped\" before this draft can be planned._\n\n")
	b.WriteString("## Open Questions\n\n")
	b.WriteString("_None recorded yet._\n\n")
	b.WriteString("## Audit Notes\n\n")
	b.WriteString("_Empty. Populated by intent-fidelity-reviewer when intent moves to shipped/._\n")
	return b.String()
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
