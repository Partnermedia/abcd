package history

// Staging is the write-ahead half of transcript capture (iss-2608230817034768).
//
// SessionEnd cannot afford redaction. It costs roughly 0.7s per MB, and the host
// cancels a shutdown hook rather than wait for it, so every transcript past a
// couple of MB was dropped silently and permanently enough to matter: nine of
// this repo's own ended sessions were absent from its store before this existed.
//
// Staging splits the work across the two hooks that can each afford their half.
// SessionEnd copies the raw bytes into ~/.abcd/history/<rootSHA>/staging/ at
// write speed and returns; the next SessionStart drains that directory through
// the same fail-closed Capture path, where there is a real time budget.
//
// The store's invariant is untouched: every transcript in transcripts/ is still
// redacted on write, because staging is NOT the store. Staged bytes are raw, so
// this directory is the one place in abcd that holds unredacted transcript text
// on purpose. It is created 0o700 and its files 0o600, it holds each transcript
// only until the next session starts, and nothing reads it but Drain.
//
// A staged file is also the outcome record the store never had. Before this,
// "absent from the store" spanned never-ended, ended-before-the-store-existed,
// ended-and-captured and ended-and-lost, and no caller could tell them apart —
// which is why the loss went unnoticed for a week. A file in staging says
// exactly one thing: this session ended and its capture has not completed yet.

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/Partnermedia/abcd/internal/fsutil"
)

// stagedSuffix marks a staged raw transcript. The extension is deliberately not
// .md: a staged file is unredacted and must never be mistaken for a record.
const stagedSuffix = ".raw"

// Staged is one raw transcript awaiting redaction.
type Staged struct {
	SessionID string    `json:"session_id"`
	StagedAt  time.Time `json:"staged_at"`
	Path      string    `json:"path"`
	Bytes     int64     `json:"bytes"`
}

// StageResult reports the outcome of one stage.
type StageResult struct {
	Staged Staged `json:"staged"`
	Wrote  bool   `json:"wrote"` // false when this session is already staged
}

// DrainFailure is one staged transcript that could not be captured. The staged
// file is deliberately LEFT in place: a failure here is recoverable by hand, and
// deleting the only copy abcd holds would convert a reported problem into the
// silent permanent loss this whole mechanism exists to end.
type DrainFailure struct {
	SessionID string `json:"session_id"`
	Path      string `json:"path"`
	Err       string `json:"error"`
}

// DrainResult reports one drain pass.
type DrainResult struct {
	Captured  []Record       `json:"captured"`
	Failed    []DrainFailure `json:"failed"`
	Remaining int            `json:"remaining"` // staged entries not attempted, budget exhausted
}

// stagingDirPath returns ~/.abcd/history/<rootSHA>/staging.
func stagingDirPath(rootSHA string) (string, error) {
	root, err := historyRoot()
	if err != nil {
		return "", err
	}
	return filepath.Join(root, rootSHA, "staging"), nil
}

// stagingDirReal verifies the owned path down to staging/ and creates the leaf
// if absent. The parents are NOT created: the store proper is bootstrapped by
// `abcd ahoy install`, and staging into a repo that was never installed would
// accumulate raw transcripts nothing would ever drain.
func stagingDirReal(rootSHA string) (string, error) {
	root, err := historyRoot()
	if err != nil {
		return "", err
	}
	repoDir := filepath.Join(root, rootSHA)
	for _, d := range []string{root, repoDir} {
		if !fsutil.IsRealDir(d) {
			return "", &StorePathError{Path: d, Msg: "not a real directory (absent or symlink); run `abcd ahoy install` to bootstrap the store"}
		}
	}
	sdir := filepath.Join(repoDir, "staging")
	// 0o700: staged transcripts are unredacted.
	if err := os.Mkdir(sdir, 0o700); err != nil && !errors.Is(err, os.ErrExist) {
		return "", &StorePathError{Path: sdir, Msg: "cannot create staging dir: " + err.Error()}
	}
	if !fsutil.IsRealDir(sdir) {
		return "", &StorePathError{Path: sdir, Msg: "staging path is not a real directory (symlink?); refusing"}
	}
	return sdir, nil
}

// stagedFilename is <compact-utc>-<session-id>.raw, matching recordFilename's
// shape so staging and transcripts sort and read alike.
func stagedFilename(at time.Time, sessionID string) string {
	return at.UTC().Format("20060102T150405.000000000Z") + "-" + sessionID + stagedSuffix
}

// sessionIDFromStaged recovers the session id from a staged filename. The stamp
// is fixed-width and session ids cannot contain "-"... except they can, so the
// split is on the FIRST "-" after the stamp, which is a fixed offset.
func sessionIDFromStaged(name string) string {
	base := strings.TrimSuffix(name, stagedSuffix)
	// Stamp is "20060102T150405.000000000Z" — 26 chars — then "-".
	const stampLen = 26
	if len(base) <= stampLen+1 || base[stampLen] != '-' {
		return ""
	}
	return base[stampLen+1:]
}

// Stage writes raw transcript bytes into the staging area without redacting
// them. It is the SessionEnd half of capture and must stay cheap: no scanner, no
// lock, no read of the existing store. Cost is one write, so the hook's runtime
// is independent of transcript size — which is the entire point.
//
// It is idempotent per session: a session already staged is a no-op, so a
// double-fired SessionEnd cannot stage two copies of the same transcript.
func Stage(rootSHA, sessionID string, raw []byte) (StageResult, error) {
	if !rootSHARe.MatchString(rootSHA) {
		return StageResult{}, errors.New(rootSHAErrMsg)
	}
	if !sessionIDRe.MatchString(sessionID) {
		return StageResult{}, fmt.Errorf("history: sessionID must be non-empty and match [A-Za-z0-9._-]+")
	}
	if len(raw) == 0 {
		return StageResult{}, errors.New("history: refusing to stage an empty transcript")
	}
	sdir, err := stagingDirReal(rootSHA)
	if err != nil {
		return StageResult{}, err
	}
	existing, err := listStaged(sdir)
	if err != nil {
		return StageResult{}, err
	}
	for _, s := range existing {
		if s.SessionID == sessionID {
			return StageResult{Staged: s, Wrote: false}, nil
		}
	}
	at := time.Now().UTC()
	path := filepath.Join(sdir, stagedFilename(at, sessionID))
	if fi, err := os.Lstat(path); err == nil && fi.Mode()&os.ModeSymlink != 0 {
		return StageResult{}, &StorePathError{Path: path, Msg: "staged path is a symlink; refusing"}
	}
	// 0o600: unredacted.
	if err := fsutil.WriteFileAtomic(path, raw, 0o600); err != nil {
		return StageResult{}, fmt.Errorf("history: stage transcript: %w", err)
	}
	return StageResult{
		Staged: Staged{SessionID: sessionID, StagedAt: at, Path: path, Bytes: int64(len(raw))},
		Wrote:  true,
	}, nil
}

// listStaged reads the staging dir, oldest first so a drain processes sessions
// in the order they ended.
func listStaged(sdir string) ([]Staged, error) {
	entries, err := os.ReadDir(sdir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, fmt.Errorf("history: read staging dir: %w", err)
	}
	var out []Staged
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), stagedSuffix) {
			continue
		}
		id := sessionIDFromStaged(e.Name())
		if id == "" || !sessionIDRe.MatchString(id) {
			continue
		}
		info, err := e.Info()
		if err != nil {
			continue
		}
		out = append(out, Staged{
			SessionID: id,
			StagedAt:  info.ModTime().UTC(),
			Path:      filepath.Join(sdir, e.Name()),
			Bytes:     info.Size(),
		})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Path < out[j].Path })
	return out, nil
}

// ListStaged returns the transcripts awaiting redaction for this repo, oldest
// first. An absent staging dir is not an error: it means nothing is pending.
func ListStaged(rootSHA string) ([]Staged, error) {
	if !rootSHARe.MatchString(rootSHA) {
		return nil, errors.New(rootSHAErrMsg)
	}
	sdir, err := stagingDirPath(rootSHA)
	if err != nil {
		return nil, err
	}
	return listStaged(sdir)
}

// Drain captures every staged transcript into the store and removes the ones it
// stored. It is the SessionStart half of capture.
//
// budget bounds how many entries one pass attempts; <= 0 means all of them. The
// bound exists because a backlog drains at redaction speed and SessionStart is
// interactive: a repo with a dozen missed sessions would otherwise stall the
// user's first prompt. Whatever the budget leaves is reported in Remaining
// rather than dropped, so a partial pass is loud and the caller can say so.
//
// A staged file is deleted ONLY when its transcript is in the store — either
// captured now or already present. Any failure leaves the file where it is.
func Drain(repoRoot, rootSHA string, budget int) (DrainResult, error) {
	if !rootSHARe.MatchString(rootSHA) {
		return DrainResult{}, errors.New(rootSHAErrMsg)
	}
	sdir, err := stagingDirPath(rootSHA)
	if err != nil {
		return DrainResult{}, err
	}
	staged, err := listStaged(sdir)
	if err != nil {
		return DrainResult{}, err
	}
	var res DrainResult
	for i, s := range staged {
		if budget > 0 && i >= budget {
			res.Remaining = len(staged) - i
			break
		}
		raw, err := fsutil.ReadGuarded(s.Path, maxTranscriptBytes)
		if err != nil {
			res.Failed = append(res.Failed, DrainFailure{SessionID: s.SessionID, Path: s.Path,
				Err: fmt.Sprintf("cannot read staged transcript: %v", err)})
			continue
		}
		cr, err := Capture(repoRoot, rootSHA, s.SessionID, raw, "native")
		if err != nil {
			res.Failed = append(res.Failed, DrainFailure{SessionID: s.SessionID, Path: s.Path, Err: err.Error()})
			continue
		}
		// Stored (or already stored): the staged copy has done its job, and it is
		// unredacted, so it goes now rather than lingering.
		if err := os.Remove(s.Path); err != nil && !errors.Is(err, os.ErrNotExist) {
			res.Failed = append(res.Failed, DrainFailure{SessionID: s.SessionID, Path: s.Path,
				Err: fmt.Sprintf("captured but could not remove staged copy: %v", err)})
			continue
		}
		if cr.Wrote {
			res.Captured = append(res.Captured, cr.Record)
		}
	}
	return res, nil
}
