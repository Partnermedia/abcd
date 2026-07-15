// Package history is abcd's native session-transcript store: the write/read/
// redact engine that populates ~/.abcd/history/<root-sha>/transcripts/ and
// retires the specstory shim (adr-29). It is transport-agnostic — no stdout, no
// os.Exit, no CLI knowledge — so any surface can drive it and marshal its
// structured results.
//
// The index.json registry and per-repo meta.json (the store's substrate) are
// owned by internal/core/ahoy and created at install time; this package only
// writes transcript records into an already-bootstrapped transcripts/ dir.
//
// Redaction is NOT reimplemented here. Every transcript is sanitised through
// internal/adapter/scanner — the same detector and masking discipline the
// launch path uses — in a two-stage, fail-closed capture: sanitise on write,
// then re-scan and refuse to write if any hard_fail secret or self home path
// survived. A stored record can never contain a live secret or an absolute home
// path.
package history

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/REPPL/abcd-cli/internal/adapter/scanner"
	"github.com/REPPL/abcd-cli/internal/fsutil"
)

// recordSchemaVersion is the frontmatter schema stamped into every record.
const recordSchemaVersion = 1

// Record is one stored transcript's metadata (its frontmatter). It never
// carries raw content — the redacted body is fetched separately via Read.
type Record struct {
	SessionID    string    `json:"session_id"`
	RootCommit   string    `json:"root_commit"`
	CapturedAt   time.Time `json:"captured_at"`
	SourceKind   string    `json:"source_kind"`
	SourceSHA256 string    `json:"source_sha256"`
	Path         string    `json:"path"`
	Secrets      int       `json:"redacted_secrets"`
	HomePaths    int       `json:"redacted_home_paths"`
}

// CaptureResult reports the outcome of one capture.
type CaptureResult struct {
	Record   Record            `json:"record"`
	Wrote    bool              `json:"wrote"`    // false on an idempotent no-op (source unchanged)
	Residual []scanner.Finding `json:"residual"` // populated only alongside RedactionResidualError
}

// RedactionResidualError is returned by Capture when the stage-two re-scan finds
// a hard_fail span that survived redaction. NO file is written. It carries the
// surviving findings' kinds/locations only — the scanner has already masked
// their Matched fields, so no raw secret material is exposed.
type RedactionResidualError struct {
	Residual []scanner.Finding
}

func (e *RedactionResidualError) Error() string {
	kinds := make([]string, 0, len(e.Residual))
	for _, f := range e.Residual {
		kinds = append(kinds, f.Kind)
	}
	return fmt.Sprintf("history: redaction left %d hard_fail span(s) unresolved [%s]; refusing to write",
		len(e.Residual), strings.Join(kinds, ", "))
}

// Capture reads a raw session transcript, redacts it through the scanner
// (two-stage, fail-closed), and writes a record into
// ~/.abcd/history/<rootSHA>/transcripts/.
//
// It is idempotent on the source's sha256: an identical source already stored
// is a no-op (Wrote=false, existing record returned, mtime preserved). It is
// fail-closed: if a hard_fail span survives redaction it returns a
// *RedactionResidualError and writes nothing.
//
// Precondition: the transcripts/ dir must already exist (abcd install created
// it). Capture re-validates that the store's owned dirs are real directories;
// it never creates the index or meta.
func Capture(repoRoot, rootSHA, sessionID string, raw []byte, kind string) (CaptureResult, error) {
	// Boundary validation — external inputs.
	if !rootSHARe.MatchString(rootSHA) {
		return CaptureResult{}, fmt.Errorf("history: rootSHA must be a 40-char lowercase hex SHA")
	}
	if !sessionIDRe.MatchString(sessionID) {
		return CaptureResult{}, fmt.Errorf("history: sessionID must be non-empty and match [A-Za-z0-9._-]+")
	}
	if _, ok := validKinds[kind]; !ok {
		return CaptureResult{}, fmt.Errorf("history: source kind %q is not one of native, specstory-import", kind)
	}

	tdir, err := ownedDirsReal(rootSHA)
	if err != nil {
		return CaptureResult{}, err
	}

	release, err := repoLock(tdir)
	if err != nil {
		return CaptureResult{}, err
	}
	defer release()

	sum := sha256.Sum256(raw)
	sourceSHA := hex.EncodeToString(sum[:])

	// Idempotency: re-capturing the SAME source for the SAME session and kind is a
	// no-op. Keying on the source SHA alone would silently attribute a second,
	// distinct session that happens to produce byte-identical bytes to the first
	// session's record — the second session would then have no record at all while
	// Capture reports success. So the no-op requires the session id and kind to
	// match too; an identical source under a new session id writes a new record.
	existing, err := listRecords(tdir)
	if err != nil {
		return CaptureResult{}, err
	}
	for _, r := range existing {
		if r.SourceSHA256 == sourceSHA && r.SessionID == sessionID && r.SourceKind == kind {
			return CaptureResult{Record: r, Wrote: false}, nil
		}
	}

	// Stage one — sanitise on write, using the per-repo merged scanner.
	sc, err := scanner.New(repoRoot)
	if err != nil {
		return CaptureResult{}, fmt.Errorf("history: scanner init: %w", err)
	}
	// Fail closed on a degraded scanner. ScanText/Redact cannot signal the
	// unavailable state in-band (only ScanBundle does), so without this guard a
	// broken per-repo pii.json would silently redact with a weakened pattern set
	// and still report the write as clean — the exact fail-open this store forbids.
	if unavail, reason := sc.Unavailable(); unavail {
		return CaptureResult{}, fmt.Errorf("history: refusing to capture with a degraded scanner: %s", reason)
	}
	text := string(raw)
	findings := sc.ScanText(text, "transcript")
	redacted, _ := scanner.Redact(text, findings)

	// Stage one-and-a-half — deterministic literal home-path backstop, wholly
	// INDEPENDENT of the scanner heuristic (defence in depth on this trust
	// boundary). The scanner's stage-two re-scan below uses the same detector
	// that produced `findings`, so a span that detector's trailing-boundary
	// heuristic dropped would slip through both stages. This literal sweep
	// collapses every remaining occurrence of the resolved $HOME to "~", then
	// fails closed if any absolute path still reveals the caller's own home.
	if home := callerHome(); home != "" {
		redacted = strings.ReplaceAll(redacted, home, "~")
		if resid := survivingCallerHome(redacted, home); len(resid) > 0 {
			return CaptureResult{Residual: resid}, &RedactionResidualError{Residual: resid}
		}
	}

	// Stage two — verify. Re-scan the redacted text; a surviving hard_fail
	// blocks the write (fail-closed).
	residual := hardFails(sc.ScanText(redacted, "transcript"))
	if len(residual) > 0 {
		return CaptureResult{Residual: residual}, &RedactionResidualError{Residual: residual}
	}

	secrets, homePaths := countBuckets(findings)
	capturedAt := time.Now().UTC()
	name := recordFilename(capturedAt, sessionID)
	path := filepath.Join(tdir, name)

	// Refuse a pre-planted symlink at the leaf record path.
	if fi, err := os.Lstat(path); err == nil && fi.Mode()&os.ModeSymlink != 0 {
		return CaptureResult{}, &StorePathError{Path: path, Msg: "record path is a symlink; refusing"}
	}

	rec := Record{
		SessionID:    sessionID,
		RootCommit:   rootSHA,
		CapturedAt:   capturedAt,
		SourceKind:   kind,
		SourceSHA256: sourceSHA,
		Path:         path,
		Secrets:      secrets,
		HomePaths:    homePaths,
	}
	if err := fsutil.WriteFileAtomic(path, marshalRecord(rec, redacted), 0o644); err != nil {
		return CaptureResult{}, fmt.Errorf("history: write record: %w", err)
	}
	return CaptureResult{Record: rec, Wrote: true}, nil
}

// List returns the records under <rootSHA>/transcripts/, newest first. It reads
// frontmatter only, never bodies. An absent transcripts dir returns no records
// and no error (the store is simply not populated for this repo yet); a
// symlinked owned dir is refused with an error.
func List(rootSHA string) ([]Record, error) {
	if !rootSHARe.MatchString(rootSHA) {
		return nil, fmt.Errorf("history: rootSHA must be a 40-char lowercase hex SHA")
	}
	tdir, err := transcriptsDir(rootSHA)
	if err != nil {
		return nil, err
	}
	if _, err := os.Lstat(tdir); err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	if !fsutil.IsRealDir(tdir) {
		return nil, &StorePathError{Path: tdir, Msg: "transcripts dir is a symlink; refusing"}
	}
	return listRecords(tdir)
}

// Read returns the metadata and full redacted body of one record, matched by
// session id (newest when a session has several records) or by the record
// filename. It never un-redacts; the stored bytes are already sanitised.
func Read(rootSHA, sessionOrFile string) (Record, []byte, error) {
	records, err := List(rootSHA)
	if err != nil {
		return Record{}, nil, err
	}
	var match *Record
	for i := range records {
		if records[i].SessionID == sessionOrFile || filepath.Base(records[i].Path) == sessionOrFile {
			match = &records[i]
			break // List is newest-first, so the first hit is the newest
		}
	}
	if match == nil {
		return Record{}, nil, fmt.Errorf("history: no record for %q under %s", sessionOrFile, rootSHA)
	}
	data, err := os.ReadFile(match.Path)
	if err != nil {
		return Record{}, nil, err
	}
	rec, body, err := parseRecord(data)
	if err != nil {
		return Record{}, nil, err
	}
	rec.Path = match.Path
	return rec, []byte(body), nil
}

// hardFails filters findings to the hard_fail severity.
func hardFails(findings []scanner.Finding) []scanner.Finding {
	var out []scanner.Finding
	for _, f := range findings {
		if f.Severity == scanner.SeverityHardFail {
			out = append(out, f)
		}
	}
	return out
}

// countBuckets rolls the redacted findings into the two audit counters stamped
// into the record frontmatter: home paths (self + third-party) and everything
// else (secret tokens plus real-name/email/username identity spans).
func countBuckets(findings []scanner.Finding) (secrets, homePaths int) {
	for _, f := range findings {
		switch f.Kind {
		case "home_path_self", "home_path_other":
			homePaths++
		default:
			secrets++
		}
	}
	return secrets, homePaths
}
