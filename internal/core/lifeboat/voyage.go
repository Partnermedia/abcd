package lifeboat

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"syscall"
	"time"

	"github.com/REPPL/abcd-cli/internal/fsutil"
)

// rootSHARe is the operator-store key: a lowercase hex commit SHA, 40 chars for
// git's SHA-1 object format or 64 for SHA-256. Accepting only 40 silently
// dropped every voyage from a SHA-256 repository. A source without a root SHA (an
// empty or non-git repo) cannot key a voyage, so its pack is not logged rather
// than logged under a forged key.
var rootSHARe = regexp.MustCompile(`^(?:[0-9a-f]{40}|[0-9a-f]{64})$`)

// voyageEntry is one line of the append-only disembark ledger. It carries no
// oracle verdict or shared_with field — nothing produces them yet, and an empty
// field would be a lie (adr-35). The manifest hash ties the line to the
// lifeboat's own _provenance.json; the full file list lives there, pinned by the
// hash, so the ledger records the count, not the list.
type voyageEntry struct {
	SchemaVersion  int    `json:"schema_version"`
	Event          string `json:"event"`
	At             string `json:"at"`
	ManifestSHA256 string `json:"manifest_sha256"`
	SourceName     string `json:"source_name"`
	SourceRootSHA  string `json:"source_root_sha"`
	Dest           string `json:"dest"`
	Files          int    `json:"files"`
	Bytes          int    `json:"bytes"`
}

// appendVoyage appends one line to the operator-level voyage ledger at
// ~/.abcd/voyage/<source-root-sha>/disembark/history.jsonl. It is genuinely
// append-only (O_APPEND, one line, no whole-file rewrite) and keyed on the
// source's root-commit SHA, mirroring the history store's per-repo scoping. It
// reports whether it appended and, if not, a short reason — a failure here never
// fails the pack, because the written lifeboat's _provenance.json is
// authoritative and the ledger is only a log.
func appendVoyage(lb Lifeboat, dest, manifestSHA string, files, bytesWritten int) (appended bool, note string) {
	rootSHA := lb.Coverage.Repo.RootSHA
	if !rootSHARe.MatchString(rootSHA) {
		return false, "skipped: source has no root-commit SHA to key the voyage"
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return false, "skipped: cannot resolve home directory"
	}

	// Create and verify the two top directories one level at a time, BEFORE any
	// deeper mkdir, so a symlinked ~/.abcd or ~/.abcd/voyage is refused rather
	// than traversed (a bare MkdirAll of the leaf would follow a symlinked base
	// and create directories under its target first).
	abcdDir := filepath.Join(home, ".abcd")
	base := filepath.Join(abcdDir, "voyage")
	if err := ensureRealDir(abcdDir); err != nil {
		return false, "failed: ~/.abcd is not a real directory (symlinked?)"
	}
	if err := ensureRealDir(base); err != nil {
		return false, "failed: voyage directory is not a real directory (symlinked?)"
	}

	// Write through an os.Root anchored at the verified base: it refuses any
	// symlinked component of <root-sha>/disembark/history.jsonl, so neither a
	// swapped <root-sha> nor a swapped disembark directory can redirect the
	// append outside base — the openat-style guard the O_NOFOLLOW leaf flag alone
	// could not give the intermediate components.
	root, err := os.OpenRoot(base)
	if err != nil {
		return false, "failed: cannot open voyage store"
	}
	defer root.Close()
	rel := filepath.ToSlash(filepath.Join(rootSHA, "disembark"))
	if err := root.MkdirAll(rel, 0o700); err != nil {
		return false, "failed: cannot create voyage directory"
	}

	entry := voyageEntry{
		SchemaVersion:  SchemaVersion,
		Event:          "disembark",
		At:             time.Now().UTC().Format(time.RFC3339),
		ManifestSHA256: manifestSHA,
		SourceName:     lb.Coverage.Repo.Name,
		SourceRootSHA:  rootSHA,
		Dest:           dest,
		Files:          files,
		Bytes:          bytesWritten,
	}
	line, err := json.Marshal(entry)
	if err != nil {
		return false, "failed: cannot encode voyage entry"
	}

	relLog := filepath.ToSlash(filepath.Join(rel, "history.jsonl"))
	f, err := root.OpenFile(relLog, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
	if err != nil {
		return false, "failed: cannot open voyage ledger"
	}
	defer f.Close()
	if _, err := f.Write(append(line, '\n')); err != nil {
		return false, "failed: cannot append to voyage ledger"
	}
	return true, ""
}

// ensureRealDir makes dir if it is absent (as a single, non-following Mkdir) and
// verifies it is a real directory — not a symlink or a file. A symlink at dir is
// refused rather than followed, so voyage never creates or writes through a
// redirected ~/.abcd or ~/.abcd/voyage.
func ensureRealDir(dir string) error {
	if err := os.Mkdir(dir, 0o700); err != nil && !os.IsExist(err) {
		return err
	}
	if !fsutil.IsRealDir(dir) {
		return &os.PathError{Op: "ensureRealDir", Path: dir, Err: syscall.ELOOP}
	}
	return nil
}
