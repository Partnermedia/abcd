package guard

import (
	"path"
	"strings"
)

// wrappers are commands that RUN another command with the same intent, so the
// hazard sits one token further along: `sudo rm -rf *` is an `rm`. Only the
// wrapper NAME is stepped over — a wrapper's own flags are not parsed (`sudo -u
// bob rm -rf *` reads `-u` as the command), a narrow v1 limitation of the same
// family as the eval gap.
var wrappers = map[string]bool{
	"sudo":    true,
	"doas":    true,
	"command": true,
	"env":     true,
	"nohup":   true,
	"time":    true,
}

// matchesAny reports whether the pattern fires on any segment of the candidate.
func matchesAny(p Pattern, segs []segment) bool {
	for i, s := range segs {
		if matchSegment(p, s) && (p.AfterCD == nil || !*p.AfterCD || precededByCD(segs[:i], s.chain)) {
			return true
		}
	}
	return false
}

// precededByCD reports whether an earlier command in the SAME chain is a `cd`.
// A cd on a previous logical line does not chain: a new line is a new shell
// command, and its failure cannot redirect this one.
func precededByCD(before []segment, chain int) bool {
	for _, s := range before {
		if s.chain != chain {
			continue
		}
		if cmd, _ := commandOf(s); cmd == "cd" {
			return true
		}
	}
	return false
}

// commandOf returns the segment's command name (basename, wrappers and
// environment-assignment prefixes stepped over) and the arguments that follow
// it. An empty name means the segment holds no command (assignments only).
func commandOf(s segment) (string, []string) {
	i := 0
	for i < len(s.tokens) {
		tok := s.tokens[i]
		if isAssignment(tok) || wrappers[path.Base(tok)] {
			i++
			continue
		}
		break
	}
	if i >= len(s.tokens) {
		return "", nil
	}
	return path.Base(s.tokens[i]), s.tokens[i+1:]
}

// isAssignment reports whether a token is a NAME=VALUE environment prefix,
// which precedes the command rather than being one.
func isAssignment(tok string) bool {
	eq := strings.IndexByte(tok, '=')
	if eq <= 0 {
		return false
	}
	for i := 0; i < eq; i++ {
		c := tok[i]
		switch {
		case c >= 'a' && c <= 'z', c >= 'A' && c <= 'Z', c >= '0' && c <= '9', c == '_':
		default:
			return false
		}
	}
	return !(tok[0] >= '0' && tok[0] <= '9')
}

// matchSegment applies the pattern's command, subcommand, and flag constraints
// to one command-position segment.
func matchSegment(p Pattern, s segment) bool {
	cmd, args := commandOf(s)
	if cmd == "" || cmd != p.Command {
		return false
	}
	if p.Subcommand != "" && subcommandOf(args, p.ValueFlags) != p.Subcommand {
		return false
	}
	for _, group := range p.Flags {
		if !flagGroupMatches(group, args) {
			return false
		}
	}
	return true
}

// subcommandOf returns the first non-flag argument, stepping over the value of
// any flag listed in valueFlags (`git -C /repo push` is a push). An unknown
// value-taking flag is not stepped over — the miss is a non-match, never a
// false block.
func subcommandOf(args []string, valueFlags []string) string {
	for i := 0; i < len(args); i++ {
		a := args[i]
		if !strings.HasPrefix(a, "-") {
			return a
		}
		if a == "--" {
			continue
		}
		if !strings.Contains(a, "=") && containsString(valueFlags, a) {
			i++ // its value is not the subcommand
		}
	}
	return ""
}

// flagGroupMatches reports whether any alternative in one "a|b" group is present
// among the argument tokens.
func flagGroupMatches(group string, args []string) bool {
	for _, alt := range strings.Split(group, "|") {
		if alt == "" {
			continue
		}
		for _, arg := range args {
			if flagMatches(alt, arg) {
				return true
			}
		}
	}
	return false
}

// flagMatches compares one flag alternative with one argument token. A long
// flag also matches its --flag=value form; a single-letter short flag also
// matches inside a bundled cluster, so -rf satisfies both -r and -f.
func flagMatches(alt, arg string) bool {
	if alt == arg {
		return true
	}
	if strings.HasPrefix(alt, "--") {
		return strings.HasPrefix(arg, alt+"=")
	}
	if len(alt) == 2 && alt[0] == '-' {
		return isShortCluster(arg) && strings.ContainsRune(arg[1:], rune(alt[1]))
	}
	return false
}

// isShortCluster reports whether a token is a bundled short-flag cluster
// (-rf, -xfd): a single leading dash followed by letters or digits only.
func isShortCluster(arg string) bool {
	if len(arg) < 2 || arg[0] != '-' || arg[1] == '-' {
		return false
	}
	for i := 1; i < len(arg); i++ {
		c := arg[i]
		switch {
		case c >= 'a' && c <= 'z', c >= 'A' && c <= 'Z', c >= '0' && c <= '9':
		default:
			return false
		}
	}
	return true
}

func containsString(xs []string, want string) bool {
	for _, x := range xs {
		if x == want {
			return true
		}
	}
	return false
}
