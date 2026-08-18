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
// The set is an UPGRADE, not the safety property. Since adr-42 a hazard behind a
// launcher this map does not name is a loud Tier 2 warn rather than a silent
// allow, so being incomplete here costs precision, not coverage — which is what
// makes it safe to add names without pretending the list is finished. It never
// can be: any binary that execs its arguments belongs here, and a repository adds
// one with a line in a Makefile.
//
// Every entry below is Linux/util-linux/coreutils, and every flag list is derived
// by PROBING the installed binary (wrappers_test.go), never by reading --help —
// nsenter's help contradicts its own parser on `-S`, and gh-299 shipped a list
// that a size assertion certified as complete while it was wrong three ways.
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

	// Scheduling, buffering and namespace shims: each runs the command that
	// follows it, and each was a silent allow for every bundled blocker (iss-272).
	"nice":        true,
	"setsid":      true,
	"stdbuf":      true,
	"ionice":      true,
	"eatmydata":   true,
	"proxychains": true,
	"chrt":        true,
	"taskset":     true,
	"unshare":     true,
	"nsenter":     true,
	"flock":       true,
	"chroot":      true,
	// `runuser -u <user> [--] <cmd>` is a direct-exec grammar. Its OTHER grammar,
	// `runuser -c <string>`, hands the string to a shell and belongs to the
	// execute-a-string family, not here.
	"runuser": true,
	// A multiplexer: its first operand is the applet, so stepping it leaves
	// `sh -c …` in command position and the interpreter path takes over.
	"busybox": true,
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
// non-match, never a false block, the same trade the operand walk already makes.
var wrapperValueFlags = map[string][]string{
	"sudo": {
		"-C", "--close-from", "-D", "--chdir", "-g", "--group", "-h", "--host",
		"-p", "--prompt", "-R", "--chroot", "-r", "--role", "-T",
		"--command-timeout", "-t", "--type", "-U", "--other-user", "-u", "--user",
	},
	"doas": {"-a", "-C", "-u"},
	// `env`'s -S/--split-string are deliberately NOT here: the execute-a-string
	// pre-pass (payload.go) owns them on the raw token chain, because the wrapper
	// walk would otherwise consume and discard the value it needs to inspect.
	"env":     {"-u", "--unset", "-C", "--chdir"},
	"time":    {"-f", "--format", "-o", "--output"},
	"xargs":   {"-a", "--arg-file", "-d", "--delimiter", "-E", "-I", "-L", "-n", "--max-args", "-P", "--max-procs", "-s", "--max-chars", "--process-slot-var"},
	"timeout": {"-k", "--kill-after", "-s", "--signal"},
	"exec":    {"-a"},
	// `command` and `nohup` take no value flags at all.

	// Probed on util-linux 2.39.3 / coreutils 9.4 (wrappers_test.go). Two of these
	// are traps a document would have got wrong:
	//   - `taskset -c` is a BOOLEAN format switch, not a value flag. Listing it
	//     would make the walk step over the COMMAND and create a miss that does
	//     not exist today.
	//   - `setsid -c` is `--ctty`, not a payload flag, and takes nothing.
	"nice":   {"-n"},
	"stdbuf": {"-i", "-o", "-e", "--input", "--output", "--error"},
	"ionice": {"-c", "-n", "--class", "--classdata"},
	"chrt":   {"-T", "-P", "-D", "--sched-runtime", "--sched-period", "--sched-deadline"},
	"unshare": {
		"-S", "-G", "-w", "-R", "--setuid", "--setgid", "--wd", "--root",
		"--map-user", "--map-group", "--map-users", "--map-groups",
		"--propagation", "--setgroups", "--monotonic", "--boottime",
	},
	// NOT -W/--wd: nsenter's is optional-argument and consumes nothing, unlike
	// unshare's -w/--wd, which is required-argument. Same package, same version,
	// same letter, opposite grammar — verified by probe, not by --help.
	"nsenter": {"-t", "-S", "-G", "--target", "--setuid", "--setgid"},
	"chroot":  {"--userspec", "--groups"},
	"flock":   {"-w", "-E", "--timeout", "--wait", "--conflict-exit-code"},
	"runuser": {"-u", "-g", "-G", "--user", "--group", "--supp-group"},
	// `setsid`, `eatmydata`, `proxychains` and `taskset` take no value flags.
}

// wrapperOperands names wrappers whose grammar puts a mandatory OPERAND between
// the flags and the command: `timeout [OPTIONS] DURATION COMMAND...`. The
// duration is not a flag, so no amount of flag stepping reaches past it — read
// as command position, `timeout 30 rm -rf /` is a command called `30`.
var wrapperOperands = map[string]int{
	"timeout": 1,
	// `chrt [OPTIONS] PRIORITY COMMAND`, `taskset [OPTIONS] MASK COMMAND`,
	// `flock [OPTIONS] FILE COMMAND`, `chroot [OPTIONS] DIR COMMAND`. Each was
	// verified by running it: `chrt -f /bin/echo hi` answers
	// "invalid priority argument: '/bin/echo'".
	"chrt":    1,
	"taskset": 1,
	"flock":   1,
	"chroot":  1,
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
	if p.Subcommand2 != "" && operandAt(ops, 1) != p.Subcommand2 {
		return false
	}
	for _, group := range p.Flags {
		if !flagGroupMatches(group, args) {
			return false
		}
	}
	for _, fv := range p.FlagValues {
		if !flagValueMatches(fv, args) {
			return false
		}
	}
	for _, prefix := range p.ArgPrefixes {
		if !argPrefixMatches(prefix, ops) {
			return false
		}
	}
	for _, pa := range p.ArgPaths {
		if !pathArgMatches(pa, ops) {
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

// flagValueMatches reports whether some argument SETS one of the flag
// alternatives to one of the accepted values. All three spellings a shell user
// reaches for are read — `-X DELETE`, `-XDELETE`, `--method=DELETE` — because a
// constraint another spelling of the same call steps past is not one.
func flagValueMatches(fv FlagValue, args []string) bool {
	for _, alt := range strings.Split(fv.Flag, "|") {
		if alt == "" {
			continue
		}
		for i, arg := range args {
			switch {
			case arg == alt:
				// The separate-token form: the value is the next argument.
				if i+1 < len(args) && acceptsValue(fv.Values, args[i+1]) {
					return true
				}
			case strings.HasPrefix(arg, alt+"="):
				if acceptsValue(fv.Values, arg[len(alt)+1:]) {
					return true
				}
			case isShortFlag(alt) && len(arg) > len(alt) && strings.HasPrefix(arg, alt):
				// A short flag's value may be attached with no separator at all.
				if acceptsValue(fv.Values, arg[len(alt):]) {
					return true
				}
			}
		}
	}
	return false
}

// acceptsValue reports whether a setting is one the constraint accepts, ignoring
// case: what makes `-X DELETE` destructive is the request it sends, and that
// does not turn on how the word was typed.
func acceptsValue(values []string, got string) bool {
	for _, want := range values {
		if want != "" && strings.EqualFold(want, got) {
			return true
		}
	}
	return false
}

// isShortFlag reports whether an alternative is a single-letter short flag
// (`-X`), the only shape that can carry an attached value.
func isShortFlag(alt string) bool {
	return len(alt) == 2 && alt[0] == '-' && alt[1] != '-'
}

// pathArgMatches reports whether some operand is a resource path rooted at
// pa.Root and exactly pa.Segments segments deep. The depth limit is the entry's
// scope, not a detail: `repos/{owner}/{repo}` IS the repository, while a deeper
// path under it names something inside the repository and stays allowed.
//
// A leading or trailing slash is ignored (`gh api /repos/owner/repo` is the same
// call), and every remaining segment must be non-empty. An operand written as a
// fully-qualified URL is normalised to its path first: `gh` passes an absolute
// URL through to the API unchanged, so spelling the host out is the same call
// and must not be a way around the same entry.
func pathArgMatches(pa PathArg, ops []string) bool {
	for _, op := range ops {
		segs := strings.Split(strings.Trim(pathOf(op), "/"), "/")
		if len(segs) != pa.Segments || segs[0] != pa.Root {
			continue
		}
		empty := false
		for _, seg := range segs {
			if seg == "" {
				empty = true
				break
			}
		}
		if !empty {
			return true
		}
	}
	return false
}

// pathOf reduces an operand to the path part a resource constraint reads: a
// leading `scheme://host` is dropped, and so is anything from the first `?` or
// `#`. Both are normalisation, not interpretation — `https://api.github.com/
// repos/owner/repo` and `repos/owner/repo?` are the same API call as
// `repos/owner/repo`, and an entry a change of spelling walks past is not one.
//
// The host itself is deliberately NOT inspected. Reading it would mean deciding
// which hostnames are the real API, which is a lookalike-domain problem this
// guard has no way to settle; normalising every authority away can only ever
// make the depth check see a path it would otherwise have missed.
func pathOf(op string) string {
	if q := strings.IndexAny(op, "?#"); q >= 0 {
		op = op[:q]
	}
	i := strings.Index(op, "://")
	if i <= 0 {
		return op
	}
	// A scheme is letters, digits, `+`, `-`, `.`, starting with a letter —
	// anything else before the `://` is ordinary path text, not an authority.
	if c := op[0]; !(c >= 'a' && c <= 'z') && !(c >= 'A' && c <= 'Z') {
		return op
	}
	for j := 1; j < i; j++ {
		c := op[j]
		switch {
		case c >= 'a' && c <= 'z', c >= 'A' && c <= 'Z', c >= '0' && c <= '9',
			c == '+', c == '-', c == '.':
		default:
			return op
		}
	}
	rest := op[i+3:]
	slash := strings.IndexByte(rest, '/')
	if slash < 0 {
		return "" // a bare authority names no resource path at all
	}
	return rest[slash:]
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
