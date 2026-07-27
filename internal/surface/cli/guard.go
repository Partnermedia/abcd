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
			"hazard named inside a quoted argument never fires. Command strings passed\n" +
			"to `eval` or `sh -c` are not parsed: a hazard hidden inside one is not\n" +
			"seen. That is a known limit of this version, not a claim of coverage.\n\n" +
			"The candidate comes from --command, or from stdin when the flag is absent.",
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
// agent, so the refusal — successor and why — goes to stderr and nowhere else. A
// warn and an allow both exit 0, which is what lets the command proceed; the
// warning still goes to stderr so it is on the record.
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
			failOpen := func(format string, a ...any) error {
				fmt.Fprintf(cmd.ErrOrStderr(),
					"abcd guard: NOT CHECKED — "+format+". This command runs UNGUARDED.\n", a...)
				return nil
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
				fmt.Fprintln(cmd.ErrOrStderr(), termsafe.Sanitize(dec.Message))
				return nil
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
	if strings.TrimSpace(flag) != "" {
		return flag, nil
	}
	if flag != "" {
		return "", fmt.Errorf("--command is empty; pass a command line to check")
	}
	raw, err := io.ReadAll(io.LimitReader(cmd.InOrStdin(), maxGuardStdinBytes))
	if err != nil {
		return "", fmt.Errorf("reading the candidate command from stdin failed: %w", err)
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
			// Loadable and wired, but switched off by a committed override. Not a
			// fault, and not something to report as protection either.
			state = "OFF — disabled in " + guard.RepoRelPath
		}
		return state
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
