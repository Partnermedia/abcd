package gitutil

import (
	"os"
	"os/exec"
	"strings"
)

// isolatedGit builds a git command under root with global and system config
// neutralised, so a developer's environment cannot change what abcd observes.
func isolatedGit(root string, args ...string) *exec.Cmd {
	cmd := exec.Command("git", append([]string{"-C", root}, args...)...)
	cmd.Env = append(os.Environ(),
		"GIT_CONFIG_GLOBAL=/dev/null",
		"GIT_CONFIG_NOSYSTEM=1",
		"GIT_OPTIONAL_LOCKS=0",
	)
	return cmd
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
