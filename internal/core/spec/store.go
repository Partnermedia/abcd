package spec

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/REPPL/abcd-cli/internal/core/frontmatter"
	"github.com/REPPL/abcd-cli/internal/fsutil"
)

// Load discovers spec files under both buckets, parses their frontmatter, and
// returns the in-memory Store. A missing specs/ directory yields an empty store
// (soft, mirroring lint's missing-dir behaviour). A present-but-malformed spec
// file is a hard, loud error.
func Load(repoRoot string) (Store, error) {
	var store Store
	for _, bucket := range []string{StatusOpen, StatusClosed} {
		specs, err := loadBucket(repoRoot, bucket)
		if err != nil {
			return Store{}, err
		}
		store.Specs = append(store.Specs, specs...)
	}
	return store, nil
}

// loadBucket reads one bucket directory. A missing directory is soft (nil, nil).
func loadBucket(repoRoot, bucket string) ([]Spec, error) {
	dir := filepath.Join(repoRoot, SpecsRelDir, bucket)
	di, err := os.Lstat(dir)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("spec: stat %s: %w", filepath.Join(SpecsRelDir, bucket), err)
	}
	if di.Mode()&os.ModeSymlink != 0 {
		return nil, fmt.Errorf("spec: %s is a symlink (refusing to follow)", filepath.Join(SpecsRelDir, bucket))
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("spec: reading %s: %w", filepath.Join(SpecsRelDir, bucket), err)
	}
	var specs []Spec
	for _, e := range entries {
		if e.IsDir() || !specFileRe.MatchString(e.Name()) {
			continue
		}
		rel := filepath.Join(SpecsRelDir, bucket, e.Name())
		data, err := readRepoFile(filepath.Join(dir, e.Name()), rel)
		if err != nil {
			return nil, err
		}
		sp, err := parseSpec(rel, string(data), bucket)
		if err != nil {
			return nil, err
		}
		specs = append(specs, sp)
	}
	return specs, nil
}

// parseSpec builds a Spec from a file's content and validates it. A file whose
// frontmatter lacks a well-formed id or intent is malformed and rejected.
func parseSpec(relPath, content, bucket string) (Spec, error) {
	fields := frontmatter.Fields(strings.Split(content, "\n"))
	sp := Spec{
		ID:     fields["id"].Value,
		Slug:   fields["slug"].Value,
		Intent: fields["intent"].Value,
		Status: bucket,
		Path:   relPath,
	}
	if err := Validate(sp); err != nil {
		return Spec{}, fmt.Errorf("spec: malformed %s: %w", relPath, err)
	}
	return sp, nil
}

// NextID mints the next spec id. The rule is:
//
//	max(N over existing spec-store files ∪ N over every intent's spec_id
//	frontmatter across .abcd/development/intents/**) + 1
//
// Scanning the intents is what keeps a freshly minted spec from colliding with
// a reservation: itd-3 shipped with spec_id: spc-1 but has no spec-store file,
// so a spec-only scan would hand out spc-1 again. Folding intent reservations in
// means the first minted id is spc-2 while that reservation stands.
func NextID(repoRoot string) (string, error) {
	max := 0
	store, err := Load(repoRoot)
	if err != nil {
		return "", err
	}
	for _, sp := range store.Specs {
		if n := specNum(sp.ID); n > max {
			max = n
		}
	}
	reserved, err := maxIntentSpecNum(repoRoot)
	if err != nil {
		return "", err
	}
	if reserved > max {
		max = reserved
	}
	return fmt.Sprintf("spc-%d", max+1), nil
}

// maxIntentSpecNum returns the highest N across every intent's spec_id
// frontmatter value in the intent lifecycle buckets, or 0 if none.
func maxIntentSpecNum(repoRoot string) (int, error) {
	max := 0
	for _, bucket := range intentBuckets {
		dir := filepath.Join(repoRoot, IntentsRelDir, bucket)
		di, err := os.Lstat(dir)
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			return 0, fmt.Errorf("spec: stat %s: %w", filepath.Join(IntentsRelDir, bucket), err)
		}
		if di.Mode()&os.ModeSymlink != 0 {
			return 0, fmt.Errorf("spec: %s is a symlink (refusing to follow)", filepath.Join(IntentsRelDir, bucket))
		}
		entries, err := os.ReadDir(dir)
		if err != nil {
			return 0, fmt.Errorf("spec: reading %s: %w", filepath.Join(IntentsRelDir, bucket), err)
		}
		for _, e := range entries {
			if e.IsDir() || !intentFileRe.MatchString(e.Name()) {
				continue
			}
			rel := filepath.Join(IntentsRelDir, bucket, e.Name())
			data, err := readRepoFile(filepath.Join(dir, e.Name()), rel)
			if err != nil {
				return 0, err
			}
			fields := frontmatter.Fields(strings.Split(string(data), "\n"))
			v := fields["spec_id"].Value
			if frontmatter.IsNull(v) {
				continue
			}
			// A non-null spec_id from which no reservation number can be parsed
			// (e.g. "spc-", "spc-abc") is fail-closed, not silently dropped:
			// dropping it would leave its reservation out of the max and let NextID
			// mint a colliding id. A "spc-N" or "spc-N-<slug>" form is fine and
			// reserves N — record-lint's planned rule (prefix ^spc-) and the
			// spec_lifecycle specNum parse both tolerate the trailing slug, so this
			// check must NOT reject that form or the two gates would disagree and a
			// lint-green record would brick the mint path.
			if !specNumRe.MatchString(v) {
				return 0, fmt.Errorf("spec: intent %s has a spec_id %q with no reservable number (must be spc-N)", rel, v)
			}
			if n := specNum(v); n > max {
				max = n
			}
		}
	}
	return max, nil
}

// Create mints an id via NextID and writes specs/open/spc-N-<slug>.md with the
// intent link in frontmatter. Both the intent id and the slug are validated
// before any path is built (the slug becomes a filename). The write is atomic.
func Create(repoRoot, intentID, slug string) (Spec, error) {
	if !intentIDRe.MatchString(intentID) {
		return Spec{}, fmt.Errorf("spec: intent id %q must match ^itd-[0-9]+$", intentID)
	}
	if !slugRe.MatchString(slug) {
		return Spec{}, fmt.Errorf("spec: slug %q must be kebab-case", slug)
	}
	id, err := NextID(repoRoot)
	if err != nil {
		return Spec{}, err
	}
	openDir := filepath.Join(repoRoot, SpecsRelDir, StatusOpen)
	if err := ensureDir(openDir, filepath.Join(SpecsRelDir, StatusOpen)); err != nil {
		return Spec{}, err
	}
	name := fmt.Sprintf("%s-%s.md", id, slug)
	// 0o644 matches the intent-side markdown writer — both write committed design-record files.
	if err := fsutil.WriteFileAtomic(filepath.Join(openDir, name), []byte(renderSpec(id, slug, intentID)), 0o644); err != nil {
		return Spec{}, fmt.Errorf("spec: writing %s: %w", filepath.Join(SpecsRelDir, StatusOpen, name), err)
	}
	sp := Spec{
		ID:     id,
		Slug:   slug,
		Intent: intentID,
		Status: StatusOpen,
		Path:   filepath.Join(SpecsRelDir, StatusOpen, name),
	}
	return sp, Validate(sp)
}

// Close moves a spec file open/ -> closed/ via os.Rename (atomic on one
// filesystem) and returns the updated Spec. It fails closed if the spec is
// missing or already closed. The linked intent is deliberately left untouched:
// moving it is a later reconcile concern that consumes Spec.Intent.
func Close(repoRoot, specID string) (Spec, error) {
	if !specIDRe.MatchString(specID) {
		return Spec{}, fmt.Errorf("spec: id %q must match ^spc-[0-9]+$", specID)
	}
	store, err := Load(repoRoot)
	if err != nil {
		return Spec{}, err
	}
	sp, ok := store.Lookup(specID)
	if !ok {
		return Spec{}, fmt.Errorf("spec: %s not found", specID)
	}
	if sp.Status == StatusClosed {
		return Spec{}, fmt.Errorf("spec: %s is already closed", specID)
	}
	name := filepath.Base(sp.Path)
	closedDir := filepath.Join(repoRoot, SpecsRelDir, StatusClosed)
	if err := ensureDir(closedDir, filepath.Join(SpecsRelDir, StatusClosed)); err != nil {
		return Spec{}, err
	}
	dstRel := filepath.Join(SpecsRelDir, StatusClosed, name)
	// Best-effort clobber guard: os.Rename would silently overwrite the destination,
	// so refuse when it already exists. This Lstat→Rename check is racy against a
	// file appearing in the window — accepted under the trusted-worktree model (only
	// the developer/agent mutates the store; there is no concurrent adversary), where
	// the atomic same-filesystem rename is preferred over a non-atomic no-clobber
	// link+remove that a crash could leave half-done.
	if _, err := os.Lstat(filepath.Join(closedDir, name)); err == nil {
		return Spec{}, fmt.Errorf("spec: refusing to overwrite existing %s", dstRel)
	}
	if err := os.Rename(filepath.Join(repoRoot, sp.Path), filepath.Join(closedDir, name)); err != nil {
		return Spec{}, fmt.Errorf("spec: closing %s: %w", specID, err)
	}
	sp.Status = StatusClosed
	sp.Path = filepath.Join(SpecsRelDir, StatusClosed, name)
	return sp, nil
}

// readRepoFile reads a repo file behind the trust-boundary guards. It opens ONCE
// with O_NOFOLLOW (refuse a symlinked leaf) and O_NONBLOCK (a FIFO/device leaf
// returns immediately instead of blocking the open), then validates the SAME file
// descriptor (regular file, size cap) before reading — so a symlink swap between
// stat and read cannot redirect it.
func readRepoFile(abs, rel string) ([]byte, error) {
	f, err := os.OpenFile(abs, os.O_RDONLY|syscall.O_NOFOLLOW|syscall.O_NONBLOCK, 0)
	if err != nil {
		if errors.Is(err, syscall.ELOOP) {
			return nil, fmt.Errorf("spec: %s is a symlink (refusing to follow)", rel)
		}
		return nil, fmt.Errorf("spec: opening %s: %w", rel, err)
	}
	defer f.Close()
	fi, err := f.Stat()
	if err != nil {
		return nil, fmt.Errorf("spec: stat %s: %w", rel, err)
	}
	if !fi.Mode().IsRegular() {
		return nil, fmt.Errorf("spec: %s is not a regular file", rel)
	}
	if fi.Size() > maxSpecFileBytes {
		return nil, fmt.Errorf("spec: %s exceeds the %d-byte cap", rel, maxSpecFileBytes)
	}
	data, err := io.ReadAll(io.LimitReader(f, maxSpecFileBytes+1))
	if err != nil {
		return nil, fmt.Errorf("spec: reading %s: %w", rel, err)
	}
	if int64(len(data)) > maxSpecFileBytes {
		return nil, fmt.Errorf("spec: %s exceeds the %d-byte cap", rel, maxSpecFileBytes)
	}
	return data, nil
}

// ensureDir creates dir if absent, refusing a symlinked leaf directory.
// NOTE: a symlinked ANCESTOR (e.g. a symlinked specs/) is not caught here — a
// low-severity follow-up under the trusted-worktree model (planting one needs
// write access equal to editing the record directly).
func ensureDir(dir, rel string) error {
	if di, err := os.Lstat(dir); err == nil && di.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("spec: %s is a symlink (refusing to follow)", rel)
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("spec: creating %s: %w", rel, err)
	}
	return nil
}
