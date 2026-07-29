package audit

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/REPPL/abcd-cli/internal/adapter/scanner"
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
// machine: absolute local home paths (/Users/<name>, /home/<name>, C:\Users\)
// and network identifiers outside the reserved documentation ranges (addresses,
// LAN hostnames, device names). A line carrying the waiver escape
// `abcd-audit:allow` is exempt, so a deliberately illustrative value can be kept.
//
// The network half is an allowlist inversion and it does NOT live here: the
// patterns are the scanner's canonical set (internal/adapter/scanner/network.go),
// consulted directly so this rule and Stage-1 redaction can never disagree about
// what a leak is. Real-email and private-repo-name detection need a configured
// allowlist and names file and are deferred to a later phase (recorded in the plan).
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

	// The canonical network-identifier set, built once in the scanner. Compiled
	// at package init so the per-file scan below costs no rebuild.
	networkPatterns = scanner.NetworkPatterns()
)

func (privacyHygiene) Meta() RuleMeta {
	return RuleMeta{
		ID:       "privacy-hygiene",
		Severity: SeverityError,
		Fix: "replace the absolute local path with a repo-relative one, and any network identifier with a reserved documentation value " +
			"(RFC 5737/3849/2606/7042, or a persona-derived device name); or add `abcd-audit:allow` on the line if it is deliberately illustrative",
		PolicyInfo: "an absolute local path or a real network identifier in a committed file leaks a username, a machine, or a network layout; " +
			"committed content uses repo-relative paths and reserved documentation identifiers",
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
			msg, leaked := privacyLeak(line)
			if !leaked {
				continue
			}
			out = append(out, Finding{
				RuleID:   "privacy-hygiene",
				Severity: SeverityError,
				File:     rel,
				Line:     i + 1,
				Message:  msg,
			})
		}
	}
	return out, nil
}

// privacyLeak reports whether one line carries content that must never be
// committed, and the message naming the class. At most one finding per line is
// reported (the absolute-path class first), matching the rule's v1 behaviour:
// the citation points a reader at the line, and the line is what gets fixed.
func privacyLeak(line string) (string, bool) {
	if hasAbsHomePath(line) {
		return "committed file contains an absolute local path", true
	}
	for _, p := range networkPatterns {
		for _, loc := range p.Re.FindAllStringIndex(line, -1) {
			matched := line[loc[0]:loc[1]]
			if p.Skip != nil && p.Skip(matched) {
				continue
			}
			if p.SkipAt != nil && p.SkipAt(line, loc[0], loc[1]) {
				continue
			}
			return "committed file contains a " + p.Label, true
		}
	}
	return "", false
}

// hasAbsHomePath reports whether a line carries an absolute home path whose
// final segment is a real username. A segment naming a well-known system
// directory under /Users — Shared, Guest, Public — is NOT a username (iss-153),
// so product code that legitimately writes to /Users/Shared needs no waiver. The
// exemption is scoped to the /Users root: a /home/<name> segment is always a
// user, and the allowlist is a macOS convention.
func hasAbsHomePath(line string) bool {
	for _, m := range absPathRe.FindAllString(line, -1) {
		seg := m
		if i := strings.LastIndexAny(m, `/\`); i >= 0 {
			seg = m[i+1:]
		}
		if isUsersRoot(m) && scanner.IsNonUserHomeSegment(seg) {
			continue
		}
		return true
	}
	return false
}

// isUsersRoot reports whether an absolute-path match hangs off a /Users (or
// C:\Users) root rather than /home.
func isUsersRoot(m string) bool {
	l := strings.ToLower(m)
	return strings.HasPrefix(l, "/users/") || strings.Contains(l, `:\users\`)
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
