package ahoy

import (
	"bytes"
	"os"
	"path/filepath"

	"github.com/REPPL/abcd-cli/internal/fsutil"
	"strings"
)

// Canonical .gitignore fence markers and the do-not-hand-edit header.
const (
	gitignoreBegin  = "# BEGIN ABCD"
	gitignoreEnd    = "# END ABCD"
	gitignoreHeader = "# abcd-managed block — do not hand-edit. Run /abcd:ahoy to refresh."
)

// gitignoreMaxBytes caps how large a .gitignore we will round-trip.
const gitignoreMaxBytes = 256 * 1024

// visibilityEntries is the canonical abcd-managed entry set per visibility,
// per the brief's visibility table (§1). Order preserved for stable diffs.
// .work/ is always ignored regardless of visibility.
var visibilityEntries = map[string][]string{
	"private": {".work/"},
	"public":  {".abcd/", "memory/", ".work/"},
}

// canonicalGitignoreBlock returns the block lines (EOL-naive) for a visibility.
func canonicalGitignoreBlock(visibility string) []string {
	lines := []string{gitignoreBegin, gitignoreHeader}
	lines = append(lines, visibilityEntries[visibility]...)
	lines = append(lines, gitignoreEnd)
	return lines
}

// gitignoreEOL returns "\r\n" when the first newline is CRLF, else "\n".
func gitignoreEOL(raw []byte) string {
	nl := bytes.IndexByte(raw, '\n')
	if nl == -1 {
		return "\n"
	}
	if nl > 0 && raw[nl-1] == '\r' {
		return "\r\n"
	}
	return "\n"
}

// gitignoreBlockDrifts reports whether the abcd-managed block in cwd/.gitignore
// differs from the canonical entry set for visibility. Read-only; fail-closed
// (drift) on any unsafe/unreadable shape so apply is offered.
func gitignoreBlockDrifts(cwd, visibility string) bool {
	entries, ok := visibilityEntries[visibility]
	if !ok {
		return false
	}
	path := filepath.Join(cwd, ".gitignore")
	fi, err := os.Lstat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return true // persisted visibility but no policy applied
		}
		return true
	}
	if fi.Mode()&os.ModeSymlink != 0 || !fi.Mode().IsRegular() {
		return true
	}
	if fi.Size() > gitignoreMaxBytes {
		return true
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		return true
	}
	inner, count := extractGitignoreBlock(raw)
	if count == 0 {
		return true // no abcd block — apply plants it
	}
	if count > 1 {
		return true // duplicate blocks — collapse on apply
	}
	current := parseGitignoreEntries(inner)
	expected := make(map[string]bool, len(entries))
	for _, e := range entries {
		expected[e] = true
	}
	if len(current) != len(expected) {
		return true
	}
	for e := range current {
		if !expected[e] {
			return true
		}
	}
	return false
}

// extractGitignoreBlock returns the inner bytes between the single BEGIN/END
// pair and the number of pairs found. count==0 => no block; count>1 => drift.
func extractGitignoreBlock(raw []byte) (inner []byte, count int) {
	begin := []byte(gitignoreBegin)
	end := []byte(gitignoreEnd)
	beginCount := bytes.Count(raw, begin)
	endCount := bytes.Count(raw, end)
	if beginCount == 0 && endCount == 0 {
		return nil, 0
	}
	if beginCount != 1 || endCount != 1 {
		return nil, 2
	}
	bi := bytes.Index(raw, begin)
	lineEnd := bytes.IndexByte(raw[bi:], '\n')
	if lineEnd == -1 {
		return nil, 2
	}
	lineEnd += bi
	ei := bytes.Index(raw[lineEnd+1:], end)
	if ei == -1 {
		return nil, 2
	}
	ei += lineEnd + 1
	return raw[lineEnd+1 : ei], 1
}

// parseGitignoreEntries returns the entry set inside a block body, stripping
// comments and blank lines and preserving a trailing slash.
func parseGitignoreEntries(block []byte) map[string]bool {
	entries := map[string]bool{}
	for _, line := range strings.Split(string(block), "\n") {
		s := strings.TrimSpace(line)
		if s == "" || strings.HasPrefix(s, "#") {
			continue
		}
		entries[s] = true
	}
	return entries
}

// applyVisibilityBlock ensures cwd/.gitignore carries exactly one canonical
// abcd block for visibility. It strips every existing block (including
// unbalanced fragments and stray END lines) then appends one canonical block,
// preserving non-block content. Returns (wrote, err). Fail-closed on unsafe
// files: an error is returned and the user's file is preserved.
func applyVisibilityBlock(cwd, visibility string) (bool, error) {
	if _, ok := visibilityEntries[visibility]; !ok {
		return false, &ahoyError{"unknown visibility: " + visibility}
	}
	path := filepath.Join(cwd, ".gitignore")

	fi, err := os.Lstat(path)
	absent := false
	if err != nil {
		if os.IsNotExist(err) {
			absent = true
		} else {
			return false, err
		}
	}
	if !absent {
		if fi.Mode()&os.ModeSymlink != 0 {
			return false, &ahoyError{"refusing to overwrite symlinked .gitignore"}
		}
		if !fi.Mode().IsRegular() {
			return false, &ahoyError{"refusing to overwrite non-regular .gitignore"}
		}
		if fi.Size() > gitignoreMaxBytes {
			return false, &ahoyError{"refusing to overwrite oversize .gitignore"}
		}
	}

	if absent {
		eol := "\n"
		body := strings.Join(canonicalGitignoreBlock(visibility), eol) + eol
		if err := fsutil.WriteFileAtomicPreserveMode(path, []byte(body)); err != nil {
			return false, err
		}
		return true, nil
	}

	raw, err := os.ReadFile(path)
	if err != nil {
		return false, err
	}
	eol := gitignoreEOL(raw)
	withoutBlocks := removeGitignoreBlocks(string(raw), eol)
	trimmedLeft := strings.TrimRight(withoutBlocks, "\r\n")
	canonical := strings.Join(canonicalGitignoreBlock(visibility), eol) + eol

	var newText string
	if trimmedLeft != "" {
		newText = trimmedLeft + eol + eol + canonical
	} else {
		newText = canonical
	}
	newBytes := []byte(newText)
	if bytes.Equal(newBytes, raw) {
		return false, nil // byte-identical — preserve mtime
	}
	if err := fsutil.WriteFileAtomicPreserveMode(path, newBytes); err != nil {
		return false, err
	}
	return true, nil
}

// removeGitignoreBlocks strips every BEGIN..END block (inclusive) plus any
// stray END without a BEGIN, preserving all other content.
func removeGitignoreBlocks(text, eol string) string {
	if !strings.Contains(text, gitignoreBegin) && !strings.Contains(text, gitignoreEnd) {
		return text
	}
	lines := splitKeepEOL(text)
	var out []string
	var buffered []string // lines held since an as-yet-unclosed BEGIN
	inside := false
	for _, line := range lines {
		stripped := strings.TrimRight(strings.TrimRight(line, "\r\n"), " \t")
		if !inside {
			if stripped == gitignoreBegin {
				inside = true
				buffered = nil
				continue
			}
			if stripped == gitignoreEnd {
				continue // stray END — drop
			}
			out = append(out, line)
			continue
		}
		if stripped == gitignoreEnd {
			inside = false
			buffered = nil // matched BEGIN..END span removed in full
			continue
		}
		buffered = append(buffered, line) // hold until we know the block closes
	}
	// An unbalanced BEGIN with no matching END must NOT swallow everything to EOF:
	// that would silently delete the user's own ignore rules. Mirror the stray-END
	// policy — drop the orphan BEGIN line alone and preserve the content after it.
	if inside {
		out = append(out, buffered...)
	}
	return strings.Join(out, "")
}

// splitKeepEOL splits text into lines that each retain their terminator.
func splitKeepEOL(text string) []string {
	var out []string
	idx := 0
	for idx < len(text) {
		nl := strings.IndexByte(text[idx:], '\n')
		if nl == -1 {
			out = append(out, text[idx:])
			break
		}
		out = append(out, text[idx:idx+nl+1])
		idx += nl + 1
	}
	return out
}

// ahoyError is a small local error type so the package needs no fmt import here.
type ahoyError struct{ msg string }

func (e *ahoyError) Error() string { return e.msg }
