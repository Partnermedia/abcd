package capture

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

// resolveRoots resolves (repoRoot, issuesRoot) from the request fields plus git
// discovery, mirroring _issue_lib._resolve_roots (contracts A–D). repoRoot is
// canonicalised; issuesRoot is made absolute without following symlinks so the
// symlink-refusal guards stay effective.
func resolveRoots(repoRoot, issuesRoot string) (string, string, error) {
	var rr string
	switch {
	case repoRoot != "":
		abs, err := filepath.Abs(repoRoot)
		if err != nil {
			return "", "", err
		}
		rr = abs
	case issuesRoot != "":
		// Discover the repo from the explicit issuesRoot's parent.
		absIssues, err := filepath.Abs(issuesRoot)
		if err != nil {
			return "", "", err
		}
		disc := discoverRepoRoot(filepath.Dir(absIssues))
		if disc == "" {
			return "", "", fmt.Errorf("custom issuesRoot requires explicit repoRoot when not in a git repo")
		}
		rr = disc
	default:
		cwd, err := os.Getwd()
		if err != nil {
			return "", "", err
		}
		disc := discoverRepoRoot(cwd)
		if disc == "" {
			return "", "", fmt.Errorf("cannot resolve repoRoot: not in a git repo and no explicit roots given")
		}
		rr = disc
	}

	var ir string
	if issuesRoot != "" {
		abs, err := filepath.Abs(issuesRoot)
		if err != nil {
			return "", "", err
		}
		ir = abs
	} else {
		ir = filepath.Join(rr, LedgerRelPath)
	}
	return rr, ir, nil
}

// discoverRepoRoot returns the git worktree root containing start, or "".
func discoverRepoRoot(start string) string {
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	cmd.Dir = start
	out, err := cmd.Output()
	if err == nil {
		if root := strings.TrimSpace(string(out)); root != "" {
			return root
		}
	}
	dir := start
	for {
		if _, err := os.Stat(filepath.Join(dir, ".git")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return ""
		}
		dir = parent
	}
}

var reNonSlug = regexp.MustCompile(`[^a-z0-9]+`)

// normaliseSlug lowercases, collapses non-alphanumeric runs to a single hyphen,
// and trims hyphens, mirroring _normalise_slug. Empty result is an error.
func normaliseSlug(slug string) (string, error) {
	candidate := strings.Trim(reNonSlug.ReplaceAllString(strings.ToLower(slug), "-"), "-")
	if candidate == "" {
		return "", fmt.Errorf("slug normalises to empty: %q", slug)
	}
	if !reSlug.MatchString(candidate) {
		return "", fmt.Errorf("slug %q is not kebab-case", candidate)
	}
	return candidate, nil
}

var emptyChecksum = sha256Hex(nil)

func sha256Hex(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

// readWithChecksum reads a file's bytes once and hashes that same buffer.
func readWithChecksum(path string) (string, string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", "", err
	}
	return string(data), sha256Hex(data), nil
}

// writeFileAtomic writes data to path via a temp file + fsync + rename,
// preserving the existing file mode (default 0644).
func writeFileAtomic(path string, data []byte) error {
	dir := filepath.Dir(path)
	mode := os.FileMode(0o644)
	if fi, err := os.Stat(path); err == nil {
		mode = fi.Mode().Perm()
	}
	tmp, err := os.CreateTemp(dir, ".abcd-iss-tmp-*")
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
