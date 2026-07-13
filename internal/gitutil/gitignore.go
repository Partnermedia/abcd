// Package gitutil holds the shared, isolated git queries. It is
// transport-agnostic: no stdout, no os.Exit, no CLI knowledge.
//
// Every git invocation here neutralises global and system config, so a
// developer's ~/.gitconfig or ~/.gitignore can never change what abcd reports
// about a repository. Queries fail open — when git is unavailable or the
// directory is not a repository, nothing is claimed rather than an error
// raised, so a convention check degrades to "cannot tell" instead of asserting
// a violation it has no evidence for.
package gitutil

import (
	"os"
	"os/exec"
	"strings"
)

// CheckIgnored returns the subset of repo-relative candidates that git ignores,
// via a single isolated `git check-ignore -z --no-index -v --stdin`. Batching
// keeps one subprocess for an arbitrary number of paths.
//
// A negation record (a pattern beginning `!`) un-ignores its path and so does
// not count as ignored. When git is unavailable or root is not a repository the
// result is empty (fail open).
func CheckIgnored(root string, candidates []string) map[string]struct{} {
	out := map[string]struct{}{}
	if len(candidates) == 0 {
		return out
	}
	cmd := exec.Command("git", "-C", root, "-c", "core.excludesFile=",
		"check-ignore", "-z", "--no-index", "-v", "--stdin")
	cmd.Env = append(os.Environ(),
		"GIT_CONFIG_GLOBAL=/dev/null",
		"GIT_CONFIG_NOSYSTEM=1",
		"GIT_OPTIONAL_LOCKS=0",
	)
	cmd.Stdin = strings.NewReader(strings.Join(candidates, "\x00") + "\x00")
	data, err := cmd.Output()
	if err != nil {
		// exit 1 == no candidate is ignored; anything else (git absent, not a
		// repo) is likewise reported as "nothing ignored".
		return out
	}
	fields := strings.Split(string(data), "\x00")
	if len(fields) > 0 && fields[len(fields)-1] == "" {
		fields = fields[:len(fields)-1]
	}
	// -v -z emits four fields per record: source, linenum, pattern, pathname.
	for i := 0; i+3 < len(fields); i += 4 {
		pattern := fields[i+2]
		pathname := fields[i+3]
		if strings.HasPrefix(pattern, "!") {
			continue // negation → the path is NOT ignored
		}
		out[pathname] = struct{}{}
	}
	return out
}

// IsIgnored reports whether a single repo-relative path is ignored by git. It is
// CheckIgnored for the one-path case; prefer CheckIgnored when asking about
// several paths, to keep it to one subprocess.
func IsIgnored(root, path string) bool {
	_, ok := CheckIgnored(root, []string{path})[path]
	return ok
}
