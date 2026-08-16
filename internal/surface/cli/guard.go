package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/REPPL/abcd-cli/internal/core/ahoy"
	"github.com/REPPL/abcd-cli/internal/core/guard"
	"github.com/REPPL/abcd-cli/internal/termsafe"
	"github.com/spf13/cobra"
)

// maxGuardStdinBytes caps the candidate command read from stdin. A command line
// is short; anything larger is not one, and an unbounded read on a hook's stdin
// would let a pathological payload stall the session it is guarding.
const maxGuardStdinBytes = 1 << 20 // 1 MiB

// newGuardCommand builds the `guard` sub-tree: the two front doors onto
// internal/core/guard's one decision. `check` is the human/scriptable verb;
// `hook` is the host adapter that a PreToolUse-style hook runs before a shell
// command executes.
//
// Both are thin: core decides, this file only formats the decision and maps it
// onto an exit status. The two differ in exactly one respect, and it is the whole
// point of the split — `check` treats a guard it cannot evaluate as a FAULT
// (loud, non-zero, so a script never mistakes silence for clearance), while
// `hook` treats the same condition as fail-OPEN (loud, exit 0, so a broken guard
// can never brick a session). spc-16, "Fail-open-loud and health".
func newGuardCommand(asJSON *bool) *cobra.Command {
	guardCmd := &cobra.Command{
		Use:   "guard",
		Short: "Check a shell command against the hazard registry before it runs",
		Args:  cobra.NoArgs,
	}

	var command string
	checkCmd := &cobra.Command{
		Use:   "check",
		Short: "Decide whether a candidate shell command is safe to run",
		Long: "Evaluates one candidate command line against the hazard registry — the\n" +
			"bundled defaults merged with this repo's `.abcd/guard.json` — and reports\n" +
			"allow, warn, or block. A blocker exits 1 and names the safe successor; a\n" +
			"warn exits 0 with the warning rendered; an allow exits 0. A guard that\n" +
			"cannot be evaluated at all (an unparsable command line, a malformed\n" +
			"registry) exits 2, so a caller never reads silence as clearance.\n\n" +
			"Matching is shell-token-aware and applies in command position only, so a\n" +
			"hazard named inside a quoted argument never fires.\n\n" +
			"An allow means no registry entry matched — it is never a statement that a\n" +
			"command is safe. The guard reads command names it can see in command\n" +
			"position, so a hazard reached any other way is not seen: one launched\n" +
			"through a wrapper outside the known set, one launched through a known\n" +
			"wrapper carrying a value-taking flag the guard does not name (`sudo -u bob\n" +
			"<hazard>` is seen, the bundled short form `sudo -Hu bob <hazard>` is not),\n" +
			"one whose API path an entry names by its ROOT segment but the host serves\n" +
			"under a prefix (a GitHub Enterprise Server install mounts the same endpoints\n" +
			"under `/api/v3/`; the api.github.com URL form IS read), a bare `$VAR` inside\n" +
			"an interpreter payload (an execute-a-string payload IS read — `sh -c`,\n" +
			"`env -S`; one the guard cannot read is warned or, for `env -S`, blocked),\n" +
			"or a dangerous form no entry describes. Coverage is what the registry\n" +
			"names.\n\n" +
			"The candidate comes from --command, or from stdin when the flag is absent.\n" +
			"Prefer stdin for a command line you did not type yourself: the shell expands\n" +
			"a double-quoted --command argument before this verb starts, so a candidate\n" +
			"containing a command substitution would run at check time. A quoted-delimiter\n" +
			"heredoc (`abcd guard check <<'EOF'` ... `EOF`) passes it through untouched.",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			candidate, err := guardCandidate(cmd, command)
			if err != nil {
				return &exitError{Code: 2, Msg: fmt.Sprintf("guard check: %s", scrubPaths(err))}
			}
			reg, err := loadGuardRegistry()
			if err != nil {
				return &exitError{Code: 2, Msg: fmt.Sprintf("guard check: %s", scrubPaths(err))}
			}
			// A disabled registry evaluated nothing, so there is no answer to
			// render — and a bare `allow` here is indistinguishable from a real
			// clearance to the CI job or script using this verb as a gate. Same
			// category as an unparsable command line: the question could not be
			// answered, so say that rather than answer it wrongly.
			if reg.Disabled {
				return &exitError{Code: 2, Msg: fmt.Sprintf(
					"guard check: the hazard registry is switched off in %s; nothing was checked", guard.RepoRelPath)}
			}
			dec, err := reg.Check(candidate)
			if err != nil {
				return &exitError{Code: 2, Msg: fmt.Sprintf("guard check: %s", scrubPaths(err))}
			}
			if rerr := render(cmd.OutOrStdout(), *asJSON, dec, func(w io.Writer) {
				writeGuardDecision(w, dec)
			}); rerr != nil {
				return rerr
			}
			// A refusal is an EXPECTED outcome, not a fault: the render is the
			// output and the exit code the only extra signal (exit 1), matching
			// how `embark from` and `intent ready` report a refusal.
			if dec.Verdict == guard.VerdictBlock {
				return &exitError{Code: 1}
			}
			return nil
		},
	}
	checkCmd.Flags().StringVar(&command, "command", "", "the candidate command line (default: read from stdin)")
	guardCmd.AddCommand(checkCmd)

	guardCmd.AddCommand(newGuardHookCommand())
	return guardCmd
}

// newGuardHookCommand builds `guard hook`: the adapter between a host's
// pre-tool-use hook payload and the core decision.
//
// The mapping is the host's own convention for a pre-execution hook: exit 2 is
// the blocking status and the hook's stderr is what the host replays to the
// agent, so the refusal — successor and why — goes to stderr and nowhere else. An
// allow exits 0 and is silent; a warn exits 1 (loud, non-blocking) so its message
// is not discarded — neither exit status blocks, so the command still runs.
//
// Every path that is NOT a decision — an unreadable payload, a tool that is not
// the shell, a command the tokenizer cannot split, a registry that will not
// load — allows the command and says so unmissably. This is the fail-open-loud
// contract: a guard that cannot answer must never be the reason a session stops,
// and must never be silently absent either (itd-103 AC 1).
func newGuardHookCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "hook",
		Short: "Host pre-tool-use adapter: decide a shell command from a hook payload",
		Long: "Reads a host pre-tool-use hook payload on stdin and evaluates its shell\n" +
			"command against the hazard registry. A blocker exits with the host's\n" +
			"blocking status and puts the safe successor and the plain-language why on\n" +
			"stderr, which is the channel the host replays to the agent. A warn and an\n" +
			"allow both let the command run.\n\n" +
			"Anything the adapter cannot turn into a decision — an unreadable payload, a\n" +
			"tool call that is not a shell command, an unparsable command line, a\n" +
			"registry that will not load — allows the command and warns loudly on\n" +
			"stderr. A guard that cannot answer never stops a session, and is never\n" +
			"silently absent.",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			// failOpen is the single exit for every non-decision path, so the
			// contract cannot be violated by forgetting one of them.
			// Exit 1, not 0, is what makes this LOUD. A pre-tool-use hook that
			// exits 0 has its stderr discarded, so the warning would exist and
			// nobody would ever see it; a non-zero, non-blocking status both lets
			// the command run and puts the warning in front of a human. Only the
			// blocking status (2) stops anything.
			failOpen := func(format string, a ...any) error {
				fmt.Fprintf(cmd.ErrOrStderr(),
					"abcd guard: NOT CHECKED — "+format+". This command runs UNGUARDED.\n", a...)
				return &exitError{Code: 1}
			}

			raw, err := io.ReadAll(io.LimitReader(cmd.InOrStdin(), maxHookStdinBytes))
			if err != nil {
				return failOpen("the hook payload could not be read (%v)", err)
			}
			var in guardHookInput
			if err := json.Unmarshal(raw, &in); err != nil {
				return failOpen("the hook payload is not readable JSON (%v)", err)
			}
			// The manifest scopes this hook to the shell tool, so a different tool
			// name means the wiring is wrong — worth saying rather than ignoring.
			if !strings.EqualFold(in.ToolName, "Bash") {
				return failOpen("the payload is a %q tool call, not a shell command", termsafe.Sanitize(in.ToolName))
			}
			candidate := in.ToolInput.Command
			if strings.TrimSpace(candidate) == "" {
				return failOpen("the payload carries no command to check")
			}

			// The payload names the session's working directory; fall back to the
			// process cwd only when it does not, exactly as the other hooks do.
			cwd := in.Cwd
			if cwd == "" {
				if wd, err := os.Getwd(); err == nil {
					cwd = wd
				}
			}
			reg, err := guard.Load(rulesRoot(cwd))
			if err != nil {
				return failOpen("the hazard registry did not load (%s)", scrubPaths(err))
			}
			// A disabled registry allows everything, which makes it an unguarded
			// session — and it is the CHEAPEST one to reach: the other unguarded
			// states need a broken install, this one needs a single file write.
			// It is not a fault (someone chose it, and the choice sits in a file a
			// reviewer can read), but it must never pass for protection.
			if reg.Disabled {
				return failOpen("the hazard registry is switched off in %s", guard.RepoRelPath)
			}
			dec, err := reg.Check(candidate)
			if err != nil {
				return failOpen("the command line could not be parsed (%s)", scrubPaths(err))
			}

			switch dec.Verdict {
			case guard.VerdictBlock:
				// Exit 2 is the host's blocking status; stderr is the message it
				// replays to the agent, so the refusal itself is the lesson.
				fmt.Fprintln(cmd.ErrOrStderr(), termsafe.Sanitize(dec.Message))
				return &exitError{Code: 2}
			case guard.VerdictWarn:
				// Exit 1, not 0, is what makes a warn LOUD on a pre-tool-use hook.
				// Exit 0 has the hook's stderr DISCARDED — the same reason failOpen
				// uses exit 1 — so a warn that returned nil would write a message
				// nobody ever sees and run as if allowed (iss-231). A non-zero,
				// non-blocking status both lets the command run and surfaces the
				// warning. Only the blocking status (2) stops anything.
				fmt.Fprintln(cmd.ErrOrStderr(), termsafe.Sanitize(dec.Message))
				return &exitError{Code: 1}
			default:
				return nil // allow: silent, and cheap
			}
		},
	}
}

// guardHookInput is the subset of the host's pre-tool-use payload the adapter
// reads. Unknown fields are ignored: the host owns this schema and adds to it.
type guardHookInput struct {
	Cwd       string `json:"cwd"`
	ToolName  string `json:"tool_name"`
	ToolInput struct {
		Command string `json:"command"`
	} `json:"tool_input"`
}

// guardCandidate resolves the command to check: the flag when given, otherwise
// stdin. An empty candidate from either channel is a usage error, never an
// implicit allow — "nothing to check" and "checked and clear" must not look the
// same to a caller.
func guardCandidate(cmd *cobra.Command, flag string) (string, error) {
	// Whether the flag was GIVEN decides the channel, not whether its value is
	// non-empty: `--command ""` is an empty question, and falling through to stdin
	// there would leave the verb waiting on a terminal that is never going to
	// answer.
	if cmd.Flags().Changed("command") {
		if strings.TrimSpace(flag) == "" {
			return "", fmt.Errorf("--command is empty; pass a command line to check")
		}
		return flag, nil
	}
	// Read ONE byte past the cap so an overflow is detectable. Truncating to the
	// cap and answering on the prefix is the quietest possible way to hand out a
	// clearance nobody earned: pad a command past the limit and the tail — the
	// part that matters — is never tokenised.
	raw, err := io.ReadAll(io.LimitReader(cmd.InOrStdin(), maxGuardStdinBytes+1))
	if err != nil {
		return "", fmt.Errorf("reading the candidate command from stdin failed: %w", err)
	}
	if len(raw) > maxGuardStdinBytes {
		return "", fmt.Errorf("the candidate command is too long (over %d bytes); it was not checked", maxGuardStdinBytes)
	}
	candidate := strings.TrimRight(string(raw), "\r\n")
	if strings.TrimSpace(candidate) == "" {
		return "", fmt.Errorf("no command to check: pass --command \"<command line>\" or pipe one on stdin")
	}
	return candidate, nil
}

// loadGuardRegistry resolves the repo root the same way the modular-rules loader
// does — the nearest ancestor holding a .abcd directory — so `.abcd/guard.json`
// is honoured from any nested working directory, kill switch included.
func loadGuardRegistry() (guard.Registry, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return guard.Registry{}, err
	}
	return guard.Load(rulesRoot(cwd))
}

// guardHealthLine renders ahoy's one-line guard-health verdict. A guard that
// cannot check anything fails OPEN inside the session, which is invisible there
// by design — so this line is the place a broken guard becomes visible, and it
// says which of the three parts is missing rather than a bare "unhealthy".
func guardHealthLine(h ahoy.GuardHealth) string {
	if h.Healthy() {
		state := fmt.Sprintf("armed (%d hazards)", h.Entries)
		if h.Disabled {
			// Loadable and wired, but switched off in .abcd/guard.json. Not a
			// fault, and not something to report as protection either. The file is
			// read from the working tree, so this can be true before anyone has
			// reviewed the edit that made it true (iss-147).
			state = "OFF — disabled in " + guard.RepoRelPath
		}
		return state
	}
	// Without a plugin root the manifest was never opened and the binary was never
	// looked for. Saying "hook not installed" here would accuse an install that
	// may be perfectly armed, so the line reports what is unknown instead.
	if !h.PluginRootResolved {
		return "UNKNOWN — " + h.Detail
	}
	var missing []string
	if !h.HookInstalled {
		missing = append(missing, "hook not installed")
	}
	if !h.BinaryReachable {
		missing = append(missing, "binary unreachable")
	}
	if !h.RegistryLoadable {
		missing = append(missing, "registry does not load")
	}
	return "NOT ARMED — " + strings.Join(missing, ", ") + " (shell commands run unchecked)"
}

// writeGuardDecision renders one decision as the human report. Entry text is
// registry-supplied (a per-repo file can add entries), so it is sanitised before
// it reaches a terminal.
func writeGuardDecision(w io.Writer, dec guard.Decision) {
	fmt.Fprintf(w, "abcd guard — %s\n", dec.Verdict)
	if dec.Verdict == guard.VerdictAllow {
		return
	}
	fmt.Fprintf(w, "  entry:       %s (%s)\n", termsafe.Sanitize(dec.EntryID), termsafe.Sanitize(dec.Tier))
	fmt.Fprintf(w, "  why:         %s\n", termsafe.Sanitize(dec.Why))
	fmt.Fprintf(w, "  run instead: %s\n", termsafe.Sanitize(dec.Successor))
	if len(dec.Matches) > 1 {
		fmt.Fprintf(w, "  also matched: %s\n", termsafe.Sanitize(strings.Join(dec.Matches[1:], ", ")))
	}
}
