package gitutil

import (
	"io"
	"os"
	"os/exec"
	"strings"
)

// isolatedGit builds a git command under root with global and system config
// neutralised, so a developer's environment cannot change what abcd observes —
// and with the two repo-local config knobs that can execute code on an
// otherwise read-only command forced off. The probe points git at arbitrary,
// possibly-hostile repositories, and a repo's own .git/config is fully trusted
// by git and cannot be disabled by env; core.hooksPath=/dev/null stops any hook
// firing and core.fsmonitor=false stops an fsmonitor daemon being spawned. These
// are the defence for read-only commands (log/tag/rev-list/rev-parse); a command
// that honours external-diff/textconv/pager config must not be added to the
// probe without further hardening.
func isolatedGit(root string, args ...string) *exec.Cmd {
	full := append([]string{
		"-c", "core.hooksPath=/dev/null",
		"-c", "core.fsmonitor=false",
		"-C", root,
	}, args...)
	cmd := exec.Command("git", full...)
	cmd.Env = append(os.Environ(),
		"GIT_CONFIG_GLOBAL=/dev/null",
		"GIT_CONFIG_NOSYSTEM=1",
		"GIT_OPTIONAL_LOCKS=0",
	)
	return cmd
}

// capWriter buffers at most a fixed number of bytes, silently discarding the
// rest, and never errors — so a git process writing far more than the cap is
// not blocked (no SIGPIPE) yet cannot grow abcd's memory past the cap.
type capWriter struct {
	buf       []byte
	remaining int
}

func (w *capWriter) Write(p []byte) (int, error) {
	if w.remaining > 0 {
		n := len(p)
		if n > w.remaining {
			n = w.remaining
		}
		w.buf = append(w.buf, p[:n]...)
		w.remaining -= n
	}
	return len(p), nil
}

// InRepo reports whether root is inside a git working tree. A convention rule
// uses it to tell "git says this path is not ignored" apart from "git cannot
// answer" (git absent, or not a repo) — the latter is "cannot tell", never
// "compliant".
func InRepo(root string) bool {
	out, err := isolatedGit(root, "rev-parse", "--is-inside-work-tree").Output()
	return err == nil && strings.TrimSpace(string(out)) == "true"
}

// TrackedFiles returns the repo-relative paths git tracks under root, NUL-safe
// so a filename with a newline cannot desync the list. Outside a repo (or with
// git absent) it returns no files and no error — a scan over committed files
// then degrades to "nothing to scan" rather than failing.
func TrackedFiles(root string) ([]string, error) {
	out, err := isolatedGit(root, "ls-files", "-z").Output()
	if err != nil {
		// Not a repo / git absent → nothing tracked, not an error.
		return nil, nil
	}
	parts := strings.Split(string(out), "\x00")
	files := make([]string, 0, len(parts))
	for _, p := range parts {
		if p != "" {
			files = append(files, p)
		}
	}
	return files, nil
}

// Run executes a read-only git command under root with the developer's global
// and system config neutralised, returning trimmed stdout. It is the shared
// primitive for tooling that reads git history (the lifeboat probe's Tier-0
// adapters); centralising it keeps every caller on the same isolated
// environment rather than re-deriving the exec plumbing. An error (git absent,
// not a repo, a failing subcommand) is returned verbatim so the caller can
// decide whether "git cannot answer" degrades to empty or is fatal.
func Run(root string, args ...string) (string, error) {
	out, err := isolatedGit(root, args...).Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

// RunLimited is Run with a hard cap on how much stdout is buffered. A hostile or
// archived repository can make a read-only command (a full `git log`) emit
// arbitrarily much output; the unbounded `Output()` would buffer all of it. The
// probe uses this so a giant history cannot exhaust memory — output past
// maxBytes is discarded (the last retained line may be truncated, which degrades
// a single probe rather than crashing it). Stderr is ignored; an exit error is
// still returned.
func RunLimited(root string, maxBytes int, args ...string) (string, error) {
	cmd := isolatedGit(root, args...)
	w := &capWriter{remaining: maxBytes}
	cmd.Stdout = w
	cmd.Stderr = io.Discard
	if err := cmd.Run(); err != nil {
		return "", err
	}
	return strings.TrimSpace(string(w.buf)), nil
}
