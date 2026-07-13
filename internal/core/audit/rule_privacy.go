package audit

import (
	"bytes"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/REPPL/abcd-cli/internal/gitutil"
)

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
	// Absolute local home paths. The trailing [/\\] and a name segment avoid
	// matching a bare "/Users" mention in prose.
	absPathRe = regexp.MustCompile(`(?:/Users/|/home/)[A-Za-z0-9._-]+[/\\]|[A-Za-z]:\\Users\\[A-Za-z0-9._-]+`)
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
	var out []Finding
	for _, rel := range tracked {
		abs := filepath.Join(ctx.RepoRoot, filepath.FromSlash(rel))
		data, err := os.ReadFile(abs)
		if err != nil {
			// A tracked path that cannot be read now (deleted-but-staged, a
			// permission quirk) is skipped, not a scan failure.
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
