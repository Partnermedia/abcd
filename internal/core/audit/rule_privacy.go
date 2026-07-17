package audit

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/REPPL/abcd-cli/internal/gitutil"
)

// maxScanBytes caps how much of a tracked file privacy-hygiene will read. A
// committed file that leaks a path in prose or source is small; anything larger
// is a data blob, skipped so a huge (or endless, via a device) file cannot
// exhaust memory. Exposed to tests via MaxScanBytesForTest.
const maxScanBytes = 4 << 20 // 4 MiB

// MaxScanBytesForTest exposes the scan cap to the package's external tests.
func MaxScanBytesForTest() int { return maxScanBytes }

// privacyHygiene scans committed files for content that must never leave the
// machine: absolute local home paths (/Users/<name>, /home/<name>, C:\Users\).
// A line carrying the waiver escape `abcd-audit:allow` is exempt, so a
// deliberately illustrative path can be kept.
//
// v1 scope is absolute local paths — the highest-signal, lowest-false-positive
// leak. Real-email and private-repo-name detection need a configured allowlist
// and names file and are deferred to a later phase (recorded in the plan).
type privacyHygiene struct{}

// auditWaiver is the language-agnostic line-scoped escape. Unlike the docs-lint
// HTML-comment form it works in source files too, where `<!-- -->` is not valid.
const auditWaiver = "abcd-audit:allow"

var (
	// Absolute local home paths. A username segment after /Users/ or /home/ is
	// required (so a bare "/Users" mention in prose is not flagged), but a trailing
	// separator is NOT: the username itself is the leak, so "/Users/name" and abcd-audit:allow
	// "/home/name" at end-of-line (e.g. `HOME=/home/name`) must be caught. This abcd-audit:allow
	// mirrors the Windows branch, which never required a trailing separator.
	absPathRe = regexp.MustCompile(`(?:/Users/|/home/)[A-Za-z0-9._-]+|[A-Za-z]:\\Users\\[A-Za-z0-9._-]+`)
)

func (privacyHygiene) Meta() RuleMeta {
	return RuleMeta{
		ID:         "privacy-hygiene",
		Severity:   SeverityError,
		Fix:        "replace the absolute local path with a repo-relative one, or add `abcd-audit:allow` on the line if it is deliberately illustrative",
		PolicyInfo: "an absolute local path in a committed file leaks a username and machine layout; committed content must use repo-relative paths",
	}
}

func (privacyHygiene) Where(Context) bool { return true }

func (privacyHygiene) Eval(ctx Context) ([]Finding, error) {
	tracked, err := gitutil.TrackedFiles(ctx.RepoRoot)
	if err != nil {
		return nil, err
	}
	// Contain every read to the repo root. os.Root refuses any path component
	// that escapes the root — including via a symlinked intermediate directory —
	// so a hostile working tree cannot redirect the scan at a file outside the
	// audited repo. If the root itself cannot be opened while git has reported
	// tracked files, the scan cannot run over content that exists — surface the
	// error rather than reporting a clean pass (audit.go:94: "a check that cannot
	// run must not be silently reported as passing").
	root, err := os.OpenRoot(ctx.RepoRoot)
	if err != nil {
		return nil, err
	}
	defer root.Close()

	var out []Finding
	for _, rel := range tracked {
		data, ok := readTrackedFile(root, filepath.FromSlash(rel))
		if !ok {
			continue
		}
		if isBinary(data) {
			continue
		}
		lines := strings.Split(string(data), "\n")
		for i, line := range lines {
			if strings.Contains(line, auditWaiver) {
				continue
			}
			if absPathRe.MatchString(line) {
				out = append(out, Finding{
					RuleID:   "privacy-hygiene",
					Severity: SeverityError,
					File:     rel,
					Line:     i + 1,
					Message:  "committed file contains an absolute local path",
				})
			}
		}
	}
	return out, nil
}

// readTrackedFile reads a tracked path safely for scanning, relative to root
// (an os.Root scoped to the repo, so no component can escape the repo — the
// containment guarantee the leaf-only O_NOFOLLOW could not give). O_NONBLOCK
// makes opening a FIFO or device return immediately rather than block until a
// writer appears; the regular-file check then skips it before any read. It reads
// only regular files and never more than maxScanBytes, so a huge or
// device-backed file cannot exhaust memory. A file that cannot be opened, is not
// a regular file, or exceeds the cap is skipped (ok=false), not a scan failure.
func readTrackedFile(root *os.Root, rel string) ([]byte, bool) {
	f, err := root.OpenFile(rel, os.O_RDONLY|syscallNoFollow, 0)
	if err != nil {
		return nil, false // escapes the root, missing, or unreadable
	}
	defer f.Close()

	info, err := f.Stat()
	if err != nil || !info.Mode().IsRegular() {
		return nil, false // FIFO, device, directory, or vanished
	}
	if info.Size() > maxScanBytes {
		return nil, false // a data blob, not prose — skip
	}
	// LimitReader is a belt-and-suspenders cap in case Size understates (a file
	// growing during the read): never buffer past the cap.
	data, err := io.ReadAll(io.LimitReader(f, maxScanBytes))
	if err != nil {
		return nil, false
	}
	return data, true
}

// isBinary reports whether data looks non-textual: a NUL byte in the first 8 KiB
// is the standard heuristic git itself uses. A binary file cannot leak a path in
// a way this line scanner would read correctly, so it is skipped.
func isBinary(data []byte) bool {
	n := len(data)
	if n > 8192 {
		n = 8192
	}
	return bytes.IndexByte(data[:n], 0) >= 0
}
