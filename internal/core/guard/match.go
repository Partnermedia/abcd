package guard

import (
	"path"
	"strings"
)

// wrappers are commands that RUN another command with the same intent, so the
// hazard sits one token further along: `sudo rm -rf *` is an `rm`. A wrapper's
// OWN arguments are stepped over with it (wrapperValueFlags, wrapperOperands),
// because a wrapper that defangs an entry as soon as it carries a flag is worse
// than one the set never knew — the registry looks armed and is not.
var wrappers = map[string]bool{
	"sudo":    true,
	"doas":    true,
	"command": true,
	"env":     true,
	"nohup":   true,
	"time":    true,
	"xargs":   true,
	"timeout": true,
	"exec":    true,
}

// wrapperValueFlags names, per wrapper, that wrapper's OWN flags which consume
// the FOLLOWING token, so the walk to command position steps over the value
// rather than reading it as the command (`sudo -u bob rm -rf *` is an `rm`).
//
// The table is deliberately explicit and small: only wrappers in the set above,
// only flags those wrappers actually document, and only the separate-token form
// — a `--flag=value` carries its value in the same token and needs no entry.
// Booleans (`sudo -n`, `env -i`, `time -p`) need no entry either; they are
// stepped over as flags. A value flag the table does not name is NOT stepped
// over, so its value is read as the command and the entry misses: a miss is a
// non-match, never a false block, the same trade operandAt already makes.
var wrapperValueFlags = map[string][]string{
	"sudo": {
		"-C", "--close-from", "-D", "--chdir", "-g", "--group", "-h", "--host",
		"-p", "--prompt", "-R", "--chroot", "-r", "--role", "-T",
		"--command-timeout", "-t", "--type", "-U", "--other-user", "-u", "--user",
	},
	"doas":    {"-a", "-C", "-u"},
	"env":     {"-u", "--unset", "-C", "--chdir", "-S", "--split-string"},
	"time":    {"-f", "--format", "-o", "--output"},
	"xargs":   {"-a", "--arg-file", "-d", "--delimiter", "-E", "-I", "-L", "-n", "--max-args", "-P", "--max-procs", "-s", "--max-chars", "--process-slot-var"},
	"timeout": {"-k", "--kill-after", "-s", "--signal"},
	"exec":    {"-a"},
	// `command` and `nohup` take no value flags at all.
}

// wrapperOperands names wrappers whose grammar puts a mandatory OPERAND between
// the flags and the command: `timeout [OPTIONS] DURATION COMMAND...`. The
// duration is not a flag, so no amount of flag stepping reaches past it — read
// as command position, `timeout 30 rm -rf /` is a command called `30`.
var wrapperOperands = map[string]int{
	"timeout": 1,
}

// reserved are shell keywords and grouping tokens that PRECEDE a command rather
// than being one: without stepping over them, `if cd scratch; then rm -rf *; fi`
// reads `then` as argv[0] and every hazard inside a conditional or a loop body
// escapes command position.
var reserved = map[string]bool{
	"if":    true,
	"then":  true,
	"elif":  true,
	"else":  true,
	"do":    true,
	"while": true,
	"until": true,
	"{":     true,
	"!":     true,
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
		if isAssignment(tok) || reserved[tok] {
			i++
			continue
		}
		if w := path.Base(tok); wrappers[w] {
			i = skipWrapperArgs(s.tokens, i+1, w)
			continue
		}
		break
	}
	if i >= len(s.tokens) {
		return "", nil
	}
	return path.Base(s.tokens[i]), s.tokens[i+1:]
}

// skipWrapperArgs advances past one wrapper's own arguments, from pos (the token
// just after the wrapper name), and returns the index where the command it
// launches begins. Without it a known wrapper turned an entry the registry does
// describe into an allow with one extra token — `sudo <hazard>` was seen,
// `sudo -u bob <hazard>` was not, because `-u` was read as the command name
// (iss-148).
func skipWrapperArgs(tokens []string, pos int, wrapper string) int {
	valueFlags := wrapperValueFlags[wrapper]
	for pos < len(tokens) {
		tok := tokens[pos]
		if tok == "--" {
			// End of the wrapper's options: everything after it is the command.
			pos++
			break
		}
		if tok == "-" || !strings.HasPrefix(tok, "-") {
			break
		}
		pos++
		if !strings.Contains(tok, "=") && containsString(valueFlags, tok) {
			pos++ // its value belongs to the wrapper, not to command position
		}
	}
	for n := wrapperOperands[wrapper]; n > 0 && pos < len(tokens); n-- {
		pos++
	}
	return pos
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
	ops := operands(args, p.ValueFlags)
	if p.Subcommand != "" && operandAt(ops, 0) != p.Subcommand {
		return false
	}
	for _, group := range p.Flags {
		if !flagGroupMatches(group, args) {
			return false
		}
	}
	for _, prefix := range p.ArgPrefixes {
		if !argPrefixMatches(prefix, ops) {
			return false
		}
	}
	return true
}

// operands returns the segment's non-flag arguments in order, stepping over the
// value of any flag listed in valueFlags (`git -C /repo push` is a push). An
// unknown value-taking flag is not stepped over — the miss is a non-match, never
// a false block. The subcommand is operand 0.
func operands(args []string, valueFlags []string) []string {
	var ops []string
	for i := 0; i < len(args); i++ {
		a := args[i]
		if !strings.HasPrefix(a, "-") {
			ops = append(ops, a)
			continue
		}
		if a == "--" {
			continue
		}
		if !strings.Contains(a, "=") && containsString(valueFlags, a) {
			i++ // its value is an option argument, not an operand
		}
	}
	return ops
}

// operandAt returns the n-th operand, or "" when the command line has no such
// argument.
func operandAt(ops []string, n int) string {
	if n < 0 || n >= len(ops) {
		return ""
	}
	return ops[n]
}

// argPrefixMatches reports whether some operand carries the prefix. Only
// operands are considered, so a prefix like "+" can never be satisfied by an
// option token: the constraint describes an argument (`git push origin
// +main:main`), not a flag.
func argPrefixMatches(prefix string, ops []string) bool {
	for _, op := range ops {
		if strings.HasPrefix(op, prefix) {
			return true
		}
	}
	return false
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
