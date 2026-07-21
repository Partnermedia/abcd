package history

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/REPPL/abcd-cli/internal/adapter/scanner"
	"github.com/REPPL/abcd-cli/internal/fsutil"
)

// rootSHARe is the immutable repo key: a lowercase hex commit SHA, 40 chars for
// git's SHA-1 object format or 64 for SHA-256. Accepting only 40 made every
// history verb (Capture/List/Read) fail for a SHA-256 repo, whose root SHA the
// ahoy layer derives at 64 chars — mirrors the voyage-ledger key fix.
var rootSHARe = regexp.MustCompile(`^(?:[0-9a-f]{40}|[0-9a-f]{64})$`)

// sessionIDRe restricts a vendor session id to filesystem-safe characters so it
// can be embedded verbatim in a record filename with no path-traversal or
// separator surprises.
var sessionIDRe = regexp.MustCompile(`^[A-Za-z0-9._-]+$`)

// validKinds are the accepted source_kind values.
var validKinds = map[string]struct{}{
	"native":           {},
	"specstory-import": {},
}

// historyRoot returns ~/.abcd/history. HOME is respected so tests can redirect.
//
// NOTE: internal/core/ahoy defines an identical unexported historyRoot for the
// index/meta layer of the same store. The store root belongs in one place;
// consolidating the two onto a shared definition is a flagged follow-up.
func historyRoot() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".abcd", "history"), nil
}

// transcriptsDir returns ~/.abcd/history/<rootSHA>/transcripts.
func transcriptsDir(rootSHA string) (string, error) {
	root, err := historyRoot()
	if err != nil {
		return "", err
	}
	return filepath.Join(root, rootSHA, "transcripts"), nil
}

// ownedDirsReal verifies every owned directory on the store path is a real
// directory (not a symlink) before a mutating call touches the leaf. Ports the
// _ensure_history_root_owned / _ensure_root_sha_dir_owned discipline: a swapped
// parent can redirect an O_NOFOLLOW leaf open, so the parents are re-checked on
// every call, not just at bootstrap.
func ownedDirsReal(rootSHA string) (string, error) {
	root, err := historyRoot()
	if err != nil {
		return "", err
	}
	repoDir := filepath.Join(root, rootSHA)
	tdir := filepath.Join(repoDir, "transcripts")
	for _, d := range []string{root, repoDir, tdir} {
		if !fsutil.IsRealDir(d) {
			return "", &StorePathError{Path: d, Msg: "not a real directory (absent or symlink); run `abcd ahoy install` to bootstrap the store"}
		}
	}
	return tdir, nil
}

// repoLock takes a per-<rootSHA> advisory lock on transcripts/.lock, disjoint
// from ahoy's index lock. The lock file is opened O_NOFOLLOW mode 0o600 so a
// pre-planted lock-file symlink is refused. The returned release closes the fd
// (which drops the flock). Ports the two-domain lock model from
// history_store.py.
func repoLock(tdir string) (func(), error) {
	lockPath := filepath.Join(tdir, ".lock")
	f, err := os.OpenFile(lockPath, os.O_RDWR|os.O_CREATE|syscall.O_NOFOLLOW, 0o600)
	if err != nil {
		return nil, &StorePathError{Path: lockPath, Msg: "lock file open refused (symlinked or unwritable): " + err.Error()}
	}
	if err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX); err != nil {
		f.Close()
		return nil, fmt.Errorf("history: acquire lock %s: %w", lockPath, err)
	}
	return func() { f.Close() }, nil
}

// recordFilename is <compact-utc>-<session-id>.md — sorts chronologically and,
// with nanosecond precision, does not collide within a session.
func recordFilename(capturedAt time.Time, sessionID string) string {
	return capturedAt.UTC().Format("20060102T150405.000000000Z") + "-" + sessionID + ".md"
}

// callerHome resolves the caller's home directory exactly as the scanner's
// ProbeIdentity does — the HOME env first (so tests and redirected runs agree),
// then os.UserHomeDir — trimmed of any trailing slash. Empty when neither
// resolves.
func callerHome() string {
	home := os.Getenv("HOME")
	if home == "" {
		if h, err := os.UserHomeDir(); err == nil {
			home = h
		}
	}
	return strings.TrimRight(home, "/")
}

// survivingCallerHome reports any absolute path in text that still reveals the
// caller's OWN home after the literal $HOME sweep: the $HOME literal itself
// (defensive — the sweep should have removed it), or a "/Users/<user>" /
// "/home/<user>" segment for the caller's local username (basename of $HOME),
// regardless of the character that follows it (trailing punctuation must never
// excuse a leak). It is a deterministic substring check with no dependency on
// the scanner heuristic. Returned findings carry only the kind (masked Matched),
// enough for RedactionResidualError to report without exposing raw material.
func survivingCallerHome(text, home string) []scanner.Finding {
	var out []scanner.Finding
	if home != "" && strings.Contains(text, home) {
		out = append(out, scanner.Finding{Kind: "home_path_self", Matched: "~"})
	}
	user := home
	if i := strings.LastIndex(home, "/"); i >= 0 {
		user = home[i+1:]
	}
	if user != "" {
		for _, prefix := range []string{"/Users/", "/home/"} {
			if containsUserSegment(text, prefix+user) {
				out = append(out, scanner.Finding{Kind: "home_path_self", Matched: "~"})
			}
		}
	}
	return out
}

// containsUserSegment reports whether needle ("/Users/<user>" or "/home/<user>")
// appears in text as a complete path segment: the rune following it must not be
// a username-continuation rune ([A-Za-z0-9._-]), so "/Users/me" does not falsely abcd-audit:allow
// match "/Users/metoo" (a different, longer username). abcd-audit:allow
func containsUserSegment(text, needle string) bool {
	from := 0
	for {
		i := strings.Index(text[from:], needle)
		if i < 0 {
			return false
		}
		end := from + i + len(needle)
		if end >= len(text) || !isPathUserByte(text[end]) {
			return true
		}
		from = from + i + 1
	}
}

func isPathUserByte(b byte) bool {
	return b == '.' || b == '_' || b == '-' ||
		(b >= 'A' && b <= 'Z') || (b >= 'a' && b <= 'z') || (b >= '0' && b <= '9')
}

// frontmatter fields (flat, one scalar per line) — a small fixed schema parsed
// by a line reader, so no YAML dependency is pulled in.
const (
	fmSchema      = "schema"
	fmSessionID   = "session_id"
	fmRootCommit  = "root_commit"
	fmCapturedAt  = "captured_at"
	fmSourceKind  = "source_kind"
	fmSourceSHA   = "source_sha256"
	fmRedSecrets  = "redacted_secrets"
	fmRedHomePath = "redacted_home_paths"
)

// marshalRecord renders a record file: YAML frontmatter then the redacted body.
func marshalRecord(r Record, body string) []byte {
	var b strings.Builder
	b.WriteString("---\n")
	fmt.Fprintf(&b, "%s: %d\n", fmSchema, recordSchemaVersion)
	fmt.Fprintf(&b, "%s: %s\n", fmSessionID, r.SessionID)
	fmt.Fprintf(&b, "%s: %s\n", fmRootCommit, r.RootCommit)
	fmt.Fprintf(&b, "%s: %s\n", fmCapturedAt, r.CapturedAt.UTC().Format(time.RFC3339Nano))
	fmt.Fprintf(&b, "%s: %s\n", fmSourceKind, r.SourceKind)
	fmt.Fprintf(&b, "%s: %s\n", fmSourceSHA, r.SourceSHA256)
	fmt.Fprintf(&b, "%s: %d\n", fmRedSecrets, r.Secrets)
	fmt.Fprintf(&b, "%s: %d\n", fmRedHomePath, r.HomePaths)
	b.WriteString("---\n")
	b.WriteString(body)
	if !strings.HasSuffix(body, "\n") {
		b.WriteString("\n")
	}
	return []byte(b.String())
}

// parseRecord splits a record file into its metadata and redacted body. The
// Path field is set by the caller. Returns an error when the frontmatter fence
// is missing or a required field is malformed.
func parseRecord(data []byte) (Record, string, error) {
	text := string(data)
	if !strings.HasPrefix(text, "---\n") {
		return Record{}, "", fmt.Errorf("history: record missing frontmatter fence")
	}
	rest := text[len("---\n"):]
	end := strings.Index(rest, "\n---\n")
	if end < 0 {
		return Record{}, "", fmt.Errorf("history: record frontmatter not terminated")
	}
	head := rest[:end]
	body := rest[end+len("\n---\n"):]

	fields := map[string]string{}
	for _, line := range strings.Split(head, "\n") {
		if line == "" {
			continue
		}
		i := strings.Index(line, ": ")
		if i < 0 {
			continue
		}
		fields[line[:i]] = line[i+2:]
	}

	var r Record
	r.SessionID = fields[fmSessionID]
	r.RootCommit = fields[fmRootCommit]
	r.SourceKind = fields[fmSourceKind]
	r.SourceSHA256 = fields[fmSourceSHA]
	if r.SessionID == "" || r.RootCommit == "" || r.SourceSHA256 == "" {
		return Record{}, "", fmt.Errorf("history: record frontmatter missing a required field")
	}
	if ts := fields[fmCapturedAt]; ts != "" {
		t, err := time.Parse(time.RFC3339Nano, ts)
		if err != nil {
			return Record{}, "", fmt.Errorf("history: record captured_at unparseable: %w", err)
		}
		r.CapturedAt = t.UTC()
	}
	r.Secrets, _ = strconv.Atoi(fields[fmRedSecrets])
	r.HomePaths, _ = strconv.Atoi(fields[fmRedHomePath])
	return r, body, nil
}

// listRecords reads every *.md record under tdir, newest first. A record file
// that fails to parse is skipped (an individual corrupt transcript is not
// fatal to the rest), mirroring ScanBundle's per-file tolerance.
func listRecords(tdir string) ([]Record, error) {
	entries, err := os.ReadDir(tdir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var out []Record
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") {
			continue
		}
		p := filepath.Join(tdir, e.Name())
		data, err := os.ReadFile(p)
		if err != nil {
			continue
		}
		r, _, err := parseRecord(data)
		if err != nil {
			continue
		}
		r.Path = p
		out = append(out, r)
	}
	sort.SliceStable(out, func(i, j int) bool {
		if !out[i].CapturedAt.Equal(out[j].CapturedAt) {
			return out[i].CapturedAt.After(out[j].CapturedAt)
		}
		return out[i].Path > out[j].Path
	})
	return out, nil
}

// StorePathError is a preflight fault: an owned store path is absent, a symlink,
// or otherwise unsafe. It is returned (never panicked) so the caller can surface
// a clean diagnostic and refuse the operation.
type StorePathError struct {
	Path string
	Msg  string
}

func (e *StorePathError) Error() string { return "history: " + e.Msg + " (" + e.Path + ")" }
