package ahoy

import (
	"bytes"
	_ "embed"
	"os"
	"path/filepath"
	"regexp"
)

// markerInner is the canonical inner content of the marker block (the rule
// loader text). Drift detection is only meaningful because there is one
// canonical source; it is embedded so the binary is self-contained.
//
//go:embed defaults/claude-md-marker-block.md
var markerInner []byte

// Canonical marker fences. Kept as byte constants for byte-level idempotency.
var (
	markerBegin = []byte("<!-- BEGIN ABCD -->")
	markerEnd   = []byte("<!-- END ABCD -->")
)

// markerBlockRe matches one fenced block, non-greedy across newlines. It only
// matches a balanced BEGIN...END pair, so an unbalanced fence reads as absent.
var markerBlockRe = regexp.MustCompile(`(?s)` + regexp.QuoteMeta("<!-- BEGIN ABCD -->") + `(.*?)` + regexp.QuoteMeta("<!-- END ABCD -->"))

// frontmatterRe matches a leading YAML frontmatter block (CRLF-aware, empty ok).
var frontmatterRe = regexp.MustCompile(`(?s)\A---(?:\r?\n)(?:.*?(?:\r?\n))?---(?:\r?\n)`)

// h1Re matches the first ATX-1 heading line (CRLF-aware).
var h1Re = regexp.MustCompile(`(?m)^# .*(?:\r?\n)`)

// detectEOL recovers a file's newline flavour. Absent/no-newline defaults to LF.
func detectEOL(existing []byte) []byte {
	if bytes.Contains(existing, []byte("\r\n")) {
		return []byte("\r\n")
	}
	return []byte("\n")
}

// synthesizeMarker builds the canonical wrapped block for a target EOL. The
// inner template's newlines are normalised to the target EOL first so a CRLF
// file never ends up with mixed EOLs.
func synthesizeMarker(inner, eol []byte) []byte {
	innerNorm := bytes.ReplaceAll(inner, []byte("\r\n"), []byte("\n"))
	innerNorm = bytes.ReplaceAll(innerNorm, []byte("\n"), eol)
	out := make([]byte, 0, len(markerBegin)+len(innerNorm)+len(markerEnd)+2*len(eol))
	out = append(out, markerBegin...)
	out = append(out, eol...)
	out = append(out, innerNorm...)
	out = append(out, eol...)
	out = append(out, markerEnd...)
	return out
}

// markerState is the three-state classifier result.
type markerState string

const (
	markerCurrent  markerState = "current"
	markerOutdated markerState = "outdated"
	markerMissing  markerState = "missing"
)

// classifyMarker reads targetPath and classifies its marker block. A read error
// or absent file is observably equivalent to "missing".
func classifyMarker(targetPath string) markerState {
	existing, err := os.ReadFile(targetPath)
	if err != nil {
		return markerMissing
	}
	matches := markerBlockRe.FindAllIndex(existing, -1)
	if len(matches) == 0 {
		return markerMissing
	}
	if len(matches) > 1 {
		return markerOutdated
	}
	eol := detectEOL(existing)
	expected := synthesizeMarker(markerInner, eol)
	first := matches[0]
	if bytes.Equal(existing[first[0]:first[1]], expected) {
		return markerCurrent
	}
	return markerOutdated
}

// installMarkerFile plants, updates, or leaves-current the block in one target.
// It returns (wrote, ok): ok=false means a per-target failure that leaves the
// file untouched. Byte-stable: a current block is not rewritten.
func installMarkerFile(targetPath string) (wrote bool, ok bool) {
	// Reject a symlinked leaf so a planted symlink cannot redirect the write.
	if fi, err := os.Lstat(targetPath); err == nil && fi.Mode()&os.ModeSymlink != 0 {
		return false, false
	}
	existing, err := os.ReadFile(targetPath)
	absent := false
	if err != nil {
		if os.IsNotExist(err) {
			absent = true
		} else {
			return false, false
		}
	}
	eol := detectEOL(existing)
	synth := synthesizeMarker(markerInner, eol)

	if absent {
		body := append(append([]byte{}, synth...), eol...)
		if err := writeFileAtomic(targetPath, body); err != nil {
			return false, false
		}
		return true, true
	}

	matches := markerBlockRe.FindAllIndex(existing, -1)
	if len(matches) == 0 {
		body := composeMarkerInsertion(existing, synth, eol)
		if err := writeFileAtomic(targetPath, body); err != nil {
			return false, false
		}
		return true, true
	}
	first := matches[0]
	if len(matches) == 1 && bytes.Equal(existing[first[0]:first[1]], synth) {
		return false, true // current — no write, mtime preserved
	}
	body := composeMarkerReplacement(existing, matches, synth)
	if err := writeFileAtomic(targetPath, body); err != nil {
		return false, false
	}
	return true, true
}

// composeMarkerInsertion inserts synth into a file with no block: after
// frontmatter, else after the first H1, else appended at EOF.
func composeMarkerInsertion(existing, synth, eol []byte) []byte {
	if loc := frontmatterRe.FindIndex(existing); loc != nil {
		head, tail := existing[:loc[1]], existing[loc[1]:]
		return concat(head, eol, synth, trailingSep(tail, eol), tail)
	}
	if loc := h1Re.FindIndex(existing); loc != nil {
		head, tail := existing[:loc[1]], existing[loc[1]:]
		return concat(head, eol, synth, trailingSep(tail, eol), tail)
	}
	if len(existing) > 0 && !bytes.HasSuffix(existing, eol) && !bytes.HasSuffix(existing, []byte("\n")) {
		existing = append(existing, eol...)
	}
	if len(existing) == 0 {
		return concat(synth, eol)
	}
	return concat(existing, eol, synth, eol)
}

// trailingSep is the separator between the block and following text: at least
// one EOL, two when the tail is non-empty and does not start with an EOL.
func trailingSep(tail, eol []byte) []byte {
	if len(tail) == 0 || tail[0] == '\n' || tail[0] == '\r' {
		return eol
	}
	return concat(eol, eol)
}

// composeMarkerReplacement replaces the first block with synth and drops later
// blocks (collapse-to-one), preserving text between blocks.
func composeMarkerReplacement(existing []byte, matches [][]int, synth []byte) []byte {
	first := matches[0]
	out := make([]byte, 0, len(existing))
	out = append(out, existing[:first[0]]...)
	out = append(out, synth...)
	cursor := first[1]
	for _, m := range matches[1:] {
		out = append(out, existing[cursor:m[0]]...)
		cursor = m[1]
	}
	out = append(out, existing[cursor:]...)
	return out
}

// removeMarkerFile strips every abcd block from one target, collapsing the EOLs
// install introduced so install->uninstall round-trips. Returns (wrote, ok).
// A symlinked leaf or non-regular file is skipped (ok=false).
func removeMarkerFile(targetPath string) (wrote bool, ok bool) {
	fi, err := os.Lstat(targetPath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, true // absent — nothing to remove
		}
		return false, false
	}
	if fi.Mode()&os.ModeSymlink != 0 || !fi.Mode().IsRegular() {
		return false, false
	}
	existing, err := os.ReadFile(targetPath)
	if err != nil {
		return false, false
	}
	matches := markerBlockRe.FindAllIndex(existing, -1)
	if len(matches) == 0 {
		return false, true // no block — untouched
	}
	body := composeMarkerRemoval(existing, matches)
	if err := writeFileAtomic(targetPath, body); err != nil {
		return false, false
	}
	return true, true
}

// composeMarkerRemoval deletes each block plus its surrounding EOL run and
// reinserts one blank-line separator when both sides survive.
func composeMarkerRemoval(existing []byte, matches [][]int) []byte {
	eol := detectEOL(existing)
	var out []byte
	cursor := 0
	n := len(existing)
	for _, m := range matches {
		start, end := m[0], m[1]
		adjStart := start
		for adjStart > 0 {
			if adjStart >= 2 && bytes.Equal(existing[adjStart-2:adjStart], []byte("\r\n")) {
				adjStart -= 2
				continue
			}
			if c := existing[adjStart-1]; c == '\n' || c == '\r' {
				adjStart--
				continue
			}
			break
		}
		adjEnd := end
		for adjEnd < n {
			if adjEnd+2 <= n && bytes.Equal(existing[adjEnd:adjEnd+2], []byte("\r\n")) {
				adjEnd += 2
				continue
			}
			if c := existing[adjEnd]; c == '\n' || c == '\r' {
				adjEnd++
				continue
			}
			break
		}
		head := existing[cursor:adjStart]
		tail := existing[adjEnd:]
		out = append(out, head...)
		switch {
		case len(head) > 0 && len(tail) > 0:
			out = append(out, eol...)
			out = append(out, eol...)
		case len(head) > 0 && len(tail) == 0:
			if !bytes.HasSuffix(head, eol) && !bytes.HasSuffix(head, []byte("\n")) {
				out = append(out, eol...)
			}
		}
		cursor = adjEnd
	}
	out = append(out, existing[cursor:]...)
	return out
}

// markerTargets maps a docs.target value to the files that host the block.
func markerTargets(docsTarget string) []string {
	switch docsTarget {
	case "claude_md":
		return []string{"CLAUDE.md"}
	case "agents_md":
		return []string{"AGENTS.md"}
	case "skip":
		return nil
	case "both":
		return []string{"CLAUDE.md", "AGENTS.md"}
	default:
		// Absent/malformed during detection: provisionally check both.
		return []string{"CLAUDE.md", "AGENTS.md"}
	}
}

// markerFileHasBlock reports whether a target file contains at least one BEGIN
// fence (used as a strong managed-repo signal during classification).
func markerFileHasBlock(targetPath string) bool {
	fi, err := os.Lstat(targetPath)
	if err != nil || fi.Mode()&os.ModeSymlink != 0 {
		return false
	}
	data, err := os.ReadFile(targetPath)
	if err != nil {
		return false
	}
	return bytes.Contains(data, markerBegin)
}

func concat(parts ...[]byte) []byte {
	var out []byte
	for _, p := range parts {
		out = append(out, p...)
	}
	return out
}

// writeFileAtomic writes data to path via a temp file + rename, preserving the
// existing file's mode when present. Parent dirs are created as needed.
func writeFileAtomic(path string, data []byte) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	mode := os.FileMode(0o644)
	if fi, err := os.Stat(path); err == nil {
		mode = fi.Mode().Perm()
	}
	tmp, err := os.CreateTemp(dir, ".abcd-tmp-*")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		os.Remove(tmpName)
		return err
	}
	if err := tmp.Sync(); err != nil {
		tmp.Close()
		os.Remove(tmpName)
		return err
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmpName)
		return err
	}
	if err := os.Chmod(tmpName, mode); err != nil {
		os.Remove(tmpName)
		return err
	}
	if err := os.Rename(tmpName, path); err != nil {
		os.Remove(tmpName)
		return err
	}
	return nil
}
