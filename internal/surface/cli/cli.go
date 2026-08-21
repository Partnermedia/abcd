// Package cli is abcd's default front door: a Cobra command tree that marshals
// internal/core results to the terminal (human text or, with --json, machine
// output). It holds no business logic — every command delegates to core and
// only formats the result, so an MCP or other front door can expose the same
// core verbs without duplicating behaviour.
package cli

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"maps"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"syscall"

	"github.com/Partnermedia/abcd/internal/adapter/scanner"
	"github.com/Partnermedia/abcd/internal/core"
	"github.com/Partnermedia/abcd/internal/core/ahoy"
	"github.com/Partnermedia/abcd/internal/core/capture"
	"github.com/Partnermedia/abcd/internal/core/history"
	"github.com/Partnermedia/abcd/internal/core/identity"
	"github.com/Partnermedia/abcd/internal/core/intent"
	"github.com/Partnermedia/abcd/internal/core/launch"
	"github.com/Partnermedia/abcd/internal/core/lifeboat"
	"github.com/Partnermedia/abcd/internal/core/lint"
	"github.com/Partnermedia/abcd/internal/core/memory"
	"github.com/Partnermedia/abcd/internal/core/record"
	"github.com/Partnermedia/abcd/internal/core/rules"
	"github.com/Partnermedia/abcd/internal/core/spec"
	"github.com/Partnermedia/abcd/internal/fsutil"
	"github.com/Partnermedia/abcd/internal/gitutil"
	"github.com/Partnermedia/abcd/internal/termsafe"
	"github.com/spf13/cobra"
)

// exitError carries a specific process exit code out of a command. The root
// command sets SilenceErrors, so main inspects this to choose the exit code and
// (when Msg is non-empty) print a single diagnostic line. An empty Msg means the
// command already rendered its output and only the exit code should propagate.
type exitError struct {
	Code int
	Msg  string
}

func (e *exitError) Error() string { return e.Msg }
func (e *exitError) ExitCode() int { return e.Code }

// helpRunE is the RunE every sub-verb parent carries. A cobra parent with no RunE
// is not Runnable, so cobra prints help and exits 0 WITHOUT running the command's
// Args validator: an unknown sub-verb — a typo, or a retired spelling like the
// pre-spc-30 `disembark oracle` — then reads as SUCCESS to a script (iss-266).
// Giving the parent a RunE keeps bare invocation showing help while letting its
// declared Args validator run and refuse a stray token at the usual exit 2. A
// parent must therefore have BOTH: with Args nil, cobra's legacyArgs falls
// through to ArbitraryArgs and a Runnable parent still exits 0. Most parents
// declare cobra.NoArgs; `banlist` declares a custom validator, and `capture` and
// `intent` are deliberately cobra.ArbitraryArgs because their positional is free
// text (they guard typos themselves). TestEveryParentRefusesAnUnknownSubverb
// holds both halves tree-wide, so a parent added later cannot reintroduce the
// hole by missing either one.
func helpRunE(cmd *cobra.Command, _ []string) error { return cmd.Help() }

// failOpenNoArgs is cobra.NoArgs for the two parents a HOST HOOK reaches: `guard`
// (PreToolUse runs `guard hook` before every shell command) and `hook`
// (UserPromptSubmit runs `hook prompt-router` before every prompt).
//
// On the hook plane an exit status is not a diagnostic, it is an INSTRUCTION: the
// host reads 2 as "block this action". Cobra's usage error exits 2, which is
// right in a terminal and wrong here — it makes abcd answer a question it did not
// evaluate. `guard hook`'s contract (spc-16, itd-103 AC 1) is fail-open-loud:
// exit 2 means "the guard decided to block", and every path that is NOT a
// decision exits 1 so the command still runs and the warning is still seen. An
// unknown sub-verb is not a decision.
//
// This is reachable because the manifest and the binary can skew — hooks/hooks.json
// ships with the plugin git clone while hooks/bootstrap.sh fetches the binary from
// the latest release — so a renamed hook sub-verb would otherwise block every
// shell command in the session, and the PreToolUse wrapper cannot rescue it: that
// wrapper treats 2 as a recognised code, so its "FAILED TO RUN … UNGUARDED" net
// never fires (iss-267).
//
// The refusal stays loud and non-zero, so iss-266's guarantee is intact: a
// mistyped sub-verb never reads as success. Only the code moves, from the host's
// blocking status to its non-blocking one.
//
// SCOPE, stated plainly because the gap matters more than the fix: this covers
// the unknown-SUB-VERB path under these two parents, and nothing else. Three
// neighbouring paths still exit 2, deliberately — an unknown TOP-LEVEL token hits
// the root's validator (root's exit-2 contract is pinned by three other tests and
// inverting it is a larger change than this), a stray positional on a LEAF such
// as `guard hook` exits 2, and any unknown FLAG exits 2 through FlagErrorFunc.
// None is reachable from today's manifest. What keeps them unreachable is not
// this function but the doctrine the manifest test enforces: hooks.json's
// spellings are frozen, and a rename is absorbed by an alias in the binary. That
// is iss-269.
// failOpenFlagError is FlagErrorFunc for the hook plane, for the same reason as
// failOpenNoArgs: an unknown flag is a usage error abcd cannot answer, and on this
// plane cobra's exit 2 is the host's instruction to BLOCK. This is the half
// iss-267 left behind, and it is the one a future manifest change makes
// reachable — add a flag to a hooks.json invocation and it skews against every
// older binary that has never heard of it (iss-269).
func failOpenFlagError(_ *cobra.Command, err error) error {
	return &exitError{Code: 1, Msg: err.Error() + hookPlaneSkewNote}
}

// hookPlaneSkewNote is the second line both hook-plane refusals carry: what the
// exit code means here, and the one thing that actually fixes it.
const hookPlaneSkewNote = "\nabcd: refusing at exit 1, not the host's blocking status — a usage error abcd" +
	" cannot answer is not a decision to block. If a hook invoked this, the plugin manifest and the" +
	" binary have skewed; re-run hooks/bootstrap.sh or reinstall the plugin."

// applyHookPlaneFailOpen installs the fail-open usage handling on every command a
// host hook can reach — the paths named in hooks/hooks.json, plus the parents on
// the way to them. It runs AFTER markUsageErrorsExitTwo, which sets a
// FlagErrorFunc on every command and would otherwise replace this one; the same
// ordering applyBanlistFlagErrors needs, and for the same reason.
//
// The set is spelled out rather than "everything under guard and hook" because
// `guard check` sits under the same parent and its contract is the OPPOSITE: it
// is the human/scriptable verb, where a fault exits 2 so a caller never reads
// silence as clearance (spc-16). Sweeping the subtree would quietly invert it.
// TestHookPlaneFailsOpenOnEveryUsageError derives the same set from the manifest
// and would fail if this list drifted from it.
func applyHookPlaneFailOpen(root *cobra.Command) {
	for _, path := range [][]string{
		{"guard"}, {"guard", "hook"},
		{"hook"}, {"hook", "prompt-router"}, {"hook", "prompt-router-reset"},
		{"hook", "session-start"}, {"hook", "session-end"},
	} {
		if cmd := findByPath(root, path); cmd != nil {
			cmd.SetFlagErrorFunc(failOpenFlagError)
			cmd.Args = failOpenNoArgs
		}
	}
}

// findByPath walks the tree by name. Deliberately NOT cobra's Find: that calls
// stripFlags, which calls mergePersistentFlags, which makes HasAvailableFlags()
// true — so merely LOOKING UP a command at construction time appends " [flags]"
// to its UseLine and silently rewrites the generated CLI reference. The drift
// gate caught it; this walk has no side effect at all.
func findByPath(root *cobra.Command, path []string) *cobra.Command {
	cur := root
	for _, name := range path {
		var next *cobra.Command
		for _, sub := range cur.Commands() {
			if sub.Name() == name {
				next = sub
				break
			}
		}
		if next == nil {
			return nil
		}
		cur = next
	}
	return cur
}

func failOpenNoArgs(cmd *cobra.Command, args []string) error {
	if err := cobra.NoArgs(cmd, args); err != nil {
		return &exitError{Code: 1, Msg: err.Error() + hookPlaneSkewNote}
	}
	return nil
}

// NewRootCommand builds the abcd command tree. Bare `abcd` renders a read-only
// status board (abcd's convention: bare invocation never mutates); subcommands
// carry the actions.
func NewRootCommand() *cobra.Command {
	var asJSON bool

	root := &cobra.Command{
		Use:   "abcd [<record-id>]",
		Short: "Agent-based configuration for development",
		Long: "Agent-based configuration for development.\n\n" +
			"Bare `abcd` renders the read-only status board — what can I do. A single\n" +
			"positional matching a record id (`iss-N`, `itd-N`, `spc-N`, `adr-N`) instead\n" +
			"reports what that record is, where it lives, and the next move for its\n" +
			"lifecycle state — what is this. Both forms are strictly read-only; any other\n" +
			"positional is refused as an unknown command.",
		SilenceUsage:  true,
		SilenceErrors: true,
		// Bare answers "what can I do"; `abcd <id>` answers "what is this, and
		// what is my next move" (spc-26). The positional is accepted iff it
		// matches the record-id shape — any other positional reproduces
		// cobra.NoArgs' unknown-command error byte-for-byte, so the id gate
		// never widens the root's surface.
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) == 1 && record.IDRe.MatchString(args[0]) {
				return nil
			}
			return cobra.NoArgs(cmd, args)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			cwd, err := os.Getwd()
			if err != nil {
				return err
			}
			// Record-id dispatch: read-only describe, never a write.
			if len(args) == 1 {
				d, err := record.Describe(cwd, args[0])
				if err != nil {
					return err
				}
				return render(cmd.OutOrStdout(), asJSON, d, func(w io.Writer) {
					// Title and link values come from record files a hostile
					// clone can shape — sanitise before the terminal.
					fmt.Fprintf(w, "%s (%s, %s) — %s\n", d.ID, d.Family, d.Status, termsafe.Sanitize(d.Title))
					fmt.Fprintf(w, "  path: %s\n", termsafe.Sanitize(d.Path))
					for _, k := range slices.Sorted(maps.Keys(d.Links)) {
						fmt.Fprintf(w, "  %s: %s\n", k, termsafe.Sanitize(d.Links[k]))
					}
					for _, m := range d.NextMoves {
						fmt.Fprintf(w, "  next: %s\n", termsafe.Sanitize(m))
					}
				})
			}
			st, err := core.Status(cwd)
			if err != nil {
				return err
			}
			return render(cmd.OutOrStdout(), asJSON, st, func(w io.Writer) {
				fmt.Fprintf(w, "abcd — %s\n", st.Dir)
				fmt.Fprintf(w, "  git repo:   %v\n", st.IsGitRepo)
				fmt.Fprintf(w, "  record:     %v\n", st.HasRecord)
				fmt.Fprintf(w, "  work tiers: %v\n", st.WorkTiers)
			})
		},
	}
	root.PersistentFlags().BoolVar(&asJSON, "json", false, "emit machine-readable JSON")

	root.AddCommand(newVersionCommand(&asJSON))
	root.AddCommand(newUpdateCommand(&asJSON))

	root.AddCommand(newAhoyCommand(&asJSON))
	root.AddCommand(newLintCommand(&asJSON))
	root.AddCommand(newGuardCommand(&asJSON))
	root.AddCommand(newIdentityCommand(&asJSON))

	var launchDryRun bool
	launchCmd := &cobra.Command{
		Use:   "launch",
		Short: "Preview the public launch bundle and release gates (--dry-run required; read-only)",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			cwd, err := os.Getwd()
			if err != nil {
				return err
			}
			if !launchDryRun {
				return fmt.Errorf("abcd launch: pass --dry-run to preview the bundle (publishing is not wired at this stage)")
			}
			rep, err := launch.DryRun(launch.DryRunRequest{
				RepoRoot: cwd,
				Version:  publishedVersion(cwd),
				// Grading the citation baseline needs the lint engine, which
				// imports launch for its semver — so the measurement is taken
				// HERE, where both are already in scope, and handed in as data.
				Citations: citationPreflight(cwd),
			})
			if err != nil {
				return err
			}
			return render(cmd.OutOrStdout(), asJSON, rep, func(w io.Writer) {
				fmt.Fprintf(w, "abcd launch (dry-run) — version %s\n", rep.Version)
				fmt.Fprintf(w, "  files bundled:  %d\n", len(rep.Bundle.Included))
				fmt.Fprintf(w, "  scan hardfails: %d\n", rep.Scan.HardFails)
				for _, g := range rep.Gates {
					if g.Name == "citation-baseline" && g.Status == "ran" {
						fmt.Fprintf(w, "  citations:      %s\n", termsafe.Sanitize(g.Detail))
					}
				}
				fmt.Fprintf(w, "  would publish:  %v\n", rep.WouldPublish)
				for _, reason := range rep.WouldRefuseOn {
					// Each reason embeds a raw repo filename (a control-char-rejected
					// path carries the offending bytes), so it is untrusted terminal
					// output and passes through the canonical sanitiser, matching the
					// citation line above.
					fmt.Fprintf(w, "  would refuse on: %s\n", termsafe.Sanitize(reason))
				}
			})
		},
	}
	launchCmd.Flags().BoolVar(&launchDryRun, "dry-run", false, "preview the launch bundle and gates without publishing")
	// `ship` is the release-cut verb: the dry-run above previews the launch
	// BUNDLE, this cuts the RELEASE (version + changelog record set). They hang
	// off one command because they gate the same event.
	launchCmd.AddCommand(newLaunchShipCommand(&asJSON))
	// `scaffold` writes the changelog-driven release machinery (release.yml,
	// auto-release.yml, runbook) into a managed repo that lacks it (itd-93). It
	// extends 04-launch because launch already owns how a release is cut and gated.
	launchCmd.AddCommand(newLaunchScaffoldCommand(&asJSON))
	root.AddCommand(launchCmd)

	root.AddCommand(newChangelogCommand(&asJSON))

	root.AddCommand(newCaptureCommand(&asJSON))
	root.AddCommand(newBanlistCommand(&asJSON))
	root.AddCommand(newMemoryCommand(&asJSON))
	root.AddCommand(newRulesCommand(&asJSON))
	root.AddCommand(newHookCommand())
	root.AddCommand(newHistoryCommand(&asJSON))
	root.AddCommand(newDocsCommand(&asJSON))
	root.AddCommand(newIntentCommand(&asJSON))
	root.AddCommand(newIdeateCommand(&asJSON))
	root.AddCommand(newSpecCommand(&asJSON))
	root.AddCommand(newDisembarkCommand(&asJSON))
	root.AddCommand(newEmbarkCommand(&asJSON))

	// A cobra usage error (unknown flag, unknown subcommand, stray positional
	// argument) is a plain error with no ExitCode(), so Run() would map it to
	// exit 1 — but `abcd lint` documents Conftest's tri-state where exit 1 means
	// "warnings only" (lint.go). A mistyped invocation must not masquerade as a
	// clean-ish gate pass: usage errors exit 2, like every usage error abcd raises
	// itself. Flag-parse errors route through FlagErrorFunc; argument errors come
	// from each command's Args validator — wrap both across the whole tree (B13).
	markUsageErrorsExitTwo(root)
	// AFTER the generic tagging, which sets a FlagErrorFunc on every command: the
	// banlist verbs need one that does NOT quote the offending token, because for
	// them the token may be a private pattern. Applied here rather than in the verb
	// so the ordering is explicit — the generic pass would otherwise overwrite it.
	applyBanlistFlagErrors(root)
	// Also after the generic tagging, and last: on the hook plane exit 2 is the
	// host's instruction to BLOCK, so every usage error a hook can provoke refuses
	// at exit 1 instead (iss-269).
	applyHookPlaneFailOpen(root)

	return root
}

// markUsageErrorsExitTwo walks the command tree and tags every cobra usage error
// with exit code 2, so a parse/usage failure never lands on the ambiguous exit 1
// that the audit tri-state reserves for "warnings only" (B13). Flag-parse errors
// are routed through FlagErrorFunc (inherited by children, but set on each for
// clarity); argument-validation errors (cobra.NoArgs violations, unknown
// subcommands) surface from each command's Args validator, which is wrapped.
//
// A validator that ALREADY chose an exit code keeps it. Exit 2 is the right
// default for a usage error, but it is not universal: on the hook plane 2 is the
// host's blocking status, so the `guard` and `hook` parents refuse at 1 instead
// (failOpenNoArgs, iss-267). Re-stamping every validator error here would silently
// undo that — which it did, until this branch.
func markUsageErrorsExitTwo(c *cobra.Command) {
	c.SetFlagErrorFunc(func(_ *cobra.Command, err error) error {
		return &exitError{Code: 2, Msg: err.Error()}
	})
	if validate := c.Args; validate != nil {
		c.Args = func(cmd *cobra.Command, args []string) error {
			if err := validate(cmd, args); err != nil {
				var coded *exitError
				if errors.As(err, &coded) {
					return err
				}
				return &exitError{Code: 2, Msg: err.Error()}
			}
			return nil
		}
	}
	for _, sub := range c.Commands() {
		markUsageErrorsExitTwo(sub)
	}
}

// docsLintResult is the machine-readable envelope for `abcd docs lint`: the
// findings plus the blocker count that decides the exit status.
type docsLintResult struct {
	Findings []lint.Finding `json:"findings"`
	Blockers int            `json:"blockers"`
}

// newDocsCommand builds the `docs` sub-tree. Its `lint` verb is the docs-currency
// drift gate: it loads .abcd/docs-lint.json (or --config), runs the shared
// internal/core/lint engine over the repo, renders the findings (text or --json),
// and exits non-zero when any blocker survives — the same engine record-lint uses.
func newDocsCommand(asJSON *bool) *cobra.Command {
	docsCmd := &cobra.Command{
		Use:   "docs",
		Short: "Documentation-currency checks for this repo",
		Args:  cobra.NoArgs,
		RunE:  helpRunE,
	}

	var configPath string
	var rootDir string
	var releaseGate bool
	lintCmd := &cobra.Command{
		Use:   "lint",
		Short: "Lint docs for change-narration, broken links, and stray root markdown",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			root := rootDir
			if root == "" {
				cwd, err := os.Getwd()
				if err != nil {
					return err
				}
				root = cwd
			}
			root, err := filepath.Abs(root)
			if err != nil {
				return err
			}
			cfgPath := configPath
			if cfgPath == "" {
				cfgPath = filepath.Join(root, ".abcd", "docs-lint.json")
			}
			cfg, err := lint.LoadConfig(cfgPath)
			if err != nil {
				// Surface config-load failures as clean, repo-relative
				// diagnostics — never a raw os.Open/os.ReadFile error, whose
				// *PathError embeds the absolute path (iss-29: no absolute path
				// in machine output). Reference what the user typed when they
				// passed --config, else the relative default.
				ref := filepath.Join(".abcd", "docs-lint.json")
				if configPath != "" {
					ref = configPath
				}
				if os.IsNotExist(err) {
					return &exitError{Code: 2, Msg: fmt.Sprintf(
						"docs lint: config not found at %s — run in a prepared repo or pass --config", ref)}
				}
				// Strip the path-bearing wrapper: a *PathError's inner Err is the
				// bare cause ("is a directory", "permission denied"), no path.
				detail := err.Error()
				var pe *os.PathError
				if errors.As(err, &pe) {
					detail = pe.Err.Error()
				}
				return &exitError{Code: 2, Msg: fmt.Sprintf("docs lint: cannot read config %s: %s", ref, detail)}
			}
			// --release-gate promotes the citation staleness finding from the
			// commit gate's warn to a blocker (spc-17: commits are never
			// calendar-blocked; a release is). The FLAG is the trust root, the
			// way `record-lint --release-gate` arms the receipt gate: a repo
			// must not be able to defang its own release by editing the
			// committed config.
			if releaseGate {
				cfg = lint.ArmCitationOverdue(cfg)
			}
			findings, err := lint.Lint(cfg, root)
			if err != nil {
				return err
			}
			blockers := 0
			for _, f := range findings {
				if f.Severity == "blocker" {
					blockers++
				}
			}
			res := docsLintResult{Findings: findings, Blockers: blockers}
			if err := render(cmd.OutOrStdout(), *asJSON, res, func(w io.Writer) {
				for _, f := range findings {
					// File and Message embed untrusted repo content (paths, link targets);
					// Severity/RuleID are enum-constrained.
					fmt.Fprintf(w, "%s:%d: [%s %s] %s\n",
						termsafe.Sanitize(f.File), f.Line, strings.ToUpper(f.Severity), f.RuleID, termsafe.Sanitize(f.Message))
				}
				fmt.Fprintf(w, "abcd docs lint — %d finding(s), %d blocker(s)\n", len(findings), blockers)
			}); err != nil {
				return err
			}
			if blockers > 0 {
				return fmt.Errorf("docs lint: %d blocker finding(s)", blockers)
			}
			return nil
		},
	}
	lintCmd.Flags().StringVar(&configPath, "config", "", "path to docs-lint.json (default: <root>/.abcd/docs-lint.json)")
	lintCmd.Flags().StringVar(&rootDir, "root", "", "repo root to lint (default: current working directory)")
	lintCmd.Flags().BoolVar(&releaseGate, "release-gate", false,
		"run as the release gate: a citation past its staleness threshold blocks instead of warning (release-time only)")
	docsCmd.AddCommand(lintCmd)
	// `cite` maintains the baseline `lint` enforces: the refresh does the live
	// fetching the gate refuses to do, and confirm closes the manual queue.
	docsCmd.AddCommand(newCiteCommand(asJSON))

	return docsCmd
}

// newDisembarkCommand builds the operator `disembark` sub-tree:
//
//   - `probe <repo>` walks a repository read-only and reports, per brief
//     section, whether a lifeboat could ground it, at what tier and confidence,
//     citing the evidence — and, for a blank, what was searched and the question
//     a human must answer.
//   - `coverage <report.json>...` reduces several probe reports to the cross-repo
//     table (section × repo) that answers whether the brief structure is sound.
//   - `plan <repo>` shows the full file set a pack would write, without writing.
//   - `pack <repo> <dest>` writes that file set to <dest> — never to the source.
//
// `pack` is the packer M3b ships, backed by the `/abcd:disembark` command
// surface (`commands/disembark.md`), so the surface-registry row is
// `shipped`. probe/coverage/plan are read-only; pack writes only to <dest>,
// behind a destination safety gate, and never mutates the source repository.
func newDisembarkCommand(asJSON *bool) *cobra.Command {
	disembarkCmd := &cobra.Command{
		Use:   "disembark",
		Short: "Lifeboat tooling: coverage probe, pack dry-run, and out-of-tree pack",
		Args:  cobra.NoArgs,
		RunE:  helpRunE,
	}

	probeCmd := &cobra.Command{
		Use:   "probe [repo]",
		Short: "Report which brief sections a lifeboat could ground from a repository (read-only)",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			target := "."
			if len(args) == 1 {
				target = args[0]
			}
			abs, err := filepath.Abs(target)
			if err != nil {
				return err
			}
			if info, err := os.Stat(abs); err != nil || !info.IsDir() {
				return &exitError{Code: 2, Msg: fmt.Sprintf("disembark probe: %s is not a directory", target)}
			}
			cov, err := lifeboat.Probe(abs)
			if err != nil {
				return &exitError{Code: 2, Msg: fmt.Sprintf("disembark probe: %s", scrubPaths(err))}
			}
			return render(cmd.OutOrStdout(), *asJSON, cov, func(w io.Writer) {
				fmt.Fprint(w, cov.Render())
			})
		},
	}

	coverageCmd := &cobra.Command{
		Use:   "coverage <report.json>...",
		Short: "Aggregate probe reports into the cross-repo section×repo coverage table",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			covs := make([]lifeboat.Coverage, 0, len(args))
			for _, path := range args {
				// A probe report is a cross-repo artifact (produced by `probe` on
				// other repos), so its content is untrusted: read it behind the same
				// guards as every other operand (O_NOFOLLOW, regular-file, size cap),
				// never a raw os.ReadFile that follows a symlink or reads unbounded.
				data, err := fsutil.ReadGuarded(path, maxOperandJSONBytes)
				if err != nil {
					// Reference the path the user typed, not an absolute PathError.
					detail := err.Error()
					switch {
					case errors.Is(err, fsutil.ErrNotRegular) || errors.Is(err, syscall.ELOOP):
						detail = "not a readable regular file (a symlink or non-regular operand is refused)"
					case errors.Is(err, fsutil.ErrTooBig):
						detail = fmt.Sprintf("exceeds the %d-byte cap", maxOperandJSONBytes)
					default:
						var pe *os.PathError
						if errors.As(err, &pe) {
							detail = pe.Err.Error()
						}
					}
					return &exitError{Code: 2, Msg: fmt.Sprintf("disembark coverage: cannot read %s: %s", path, detail)}
				}
				var cov lifeboat.Coverage
				if err := json.Unmarshal(data, &cov); err != nil {
					return &exitError{Code: 2, Msg: fmt.Sprintf("disembark coverage: %s is not a coverage report: %s", path, err)}
				}
				// A probe report always stamps schema_version >= 1; json.Unmarshal
				// of any other JSON object succeeds with the zero value (schema
				// version 0), which would otherwise sail past the guard as an
				// all-blank phantom repo (B38). Reject it as not a coverage report,
				// mirroring the type-mismatch message above.
				if cov.SchemaVersion < 1 {
					return &exitError{Code: 2, Msg: fmt.Sprintf(
						"disembark coverage: %s is not a coverage report: missing schema_version", path)}
				}
				if cov.SchemaVersion > lifeboat.SchemaVersion {
					return &exitError{Code: 2, Msg: fmt.Sprintf(
						"disembark coverage: %s is schema v%d; this abcd knows up to v%d — upgrade abcd",
						path, cov.SchemaVersion, lifeboat.SchemaVersion)}
				}
				covs = append(covs, cov)
			}
			agg := lifeboat.Aggregate(covs)
			return render(cmd.OutOrStdout(), *asJSON, agg, func(w io.Writer) {
				fmt.Fprint(w, agg.Render())
			})
		},
	}

	planCmd := &cobra.Command{
		Use:   "plan [repo]",
		Short: "Show the full lifeboat file set a pack would write, without writing anything (dry run)",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			target := "."
			if len(args) == 1 {
				target = args[0]
			}
			abs, err := filepath.Abs(target)
			if err != nil {
				return err
			}
			if info, err := os.Stat(abs); err != nil || !info.IsDir() {
				return &exitError{Code: 2, Msg: fmt.Sprintf("disembark plan: %s is not a directory", target)}
			}
			lb, err := lifeboat.Plan(abs)
			if err != nil {
				return &exitError{Code: 2, Msg: fmt.Sprintf("disembark plan: %s", scrubPaths(err))}
			}
			return render(cmd.OutOrStdout(), *asJSON, lb.Manifest(), func(w io.Writer) {
				fmt.Fprint(w, lb.RenderManifest())
			})
		},
	}

	packCmd := &cobra.Command{
		Use:   "pack <repo> <dest>",
		Short: "Pack a lifeboat from a repository into a destination directory (writes <dest>, never the source)",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			repoAbs, err := filepath.Abs(args[0])
			if err != nil {
				return err
			}
			if info, err := os.Stat(repoAbs); err != nil || !info.IsDir() {
				return &exitError{Code: 2, Msg: fmt.Sprintf("disembark pack: %s is not a directory", args[0])}
			}
			// Build the source repo's secret scanner and fail closed if its config
			// is degraded — a pack must not ship secrets under a weakened ruleset.
			sc, err := scanner.New(repoAbs)
			if err != nil {
				return &exitError{Code: 2, Msg: fmt.Sprintf("disembark pack: %s", scrubPaths(err))}
			}
			if bad, reason := sc.Unavailable(); bad {
				return &exitError{Code: 2, Msg: fmt.Sprintf("disembark pack: secret scanner unavailable, refusing: %s", reason)}
			}
			scan := func(files []lifeboat.PlannedFile) error {
				hard, first := 0, ""
				for _, f := range files {
					for _, fnd := range sc.ScanText(string(f.Content), f.Path) {
						if fnd.Severity == scanner.SeverityHardFail {
							hard++
							if first == "" {
								first = fmt.Sprintf("%s (%s)", f.Path, fnd.Kind)
							}
						}
					}
				}
				if hard > 0 {
					return fmt.Errorf("%d hard-fail secret(s) in planned content (first: %s); fix at source, not in the lifeboat", hard, first)
				}
				return nil
			}
			res, err := lifeboat.Pack(repoAbs, args[1], scan)
			if err != nil {
				return &exitError{Code: 2, Msg: fmt.Sprintf("disembark pack: %s", scrubPaths(err))}
			}
			return render(cmd.OutOrStdout(), *asJSON, res, func(w io.Writer) {
				fmt.Fprint(w, res.Render())
			})
		},
	}

	var lessonsJSON string
	graveyardCmd := &cobra.Command{
		Use:   "graveyard <lifeboat-dir> --lessons-json <file|->",
		Short: "Validate host-produced lesson JSON against a packed lifeboat and write the survivors (cite-or-be-dropped)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if lessonsJSON == "" {
				return &exitError{Code: 2, Msg: "disembark graveyard: --lessons-json <file|-> is required"}
			}
			dirAbs, err := filepath.Abs(args[0])
			if err != nil {
				return err
			}
			raw, err := readLessonsPayload(cmd, lessonsJSON)
			if err != nil {
				return &exitError{Code: 2, Msg: "disembark graveyard: " + scrubPaths(err)}
			}
			// Exit 0 even when every entry was dropped or nothing was written:
			// a drop is reported honestly, not a failure. Only structural faults
			// (not a lifeboat, unreadable graveyard, unparseable/oversize payload)
			// return an error and exit 2.
			res, err := lifeboat.IngestLessons(dirAbs, raw)
			if err != nil {
				return &exitError{Code: 2, Msg: "disembark graveyard: " + scrubPaths(err)}
			}
			return render(cmd.OutOrStdout(), *asJSON, res, func(w io.Writer) {
				fmt.Fprint(w, res.Render())
			})
		},
	}
	graveyardCmd.Flags().StringVar(&lessonsJSON, "lessons-json", "", "path to the host-produced lesson JSON (or - for stdin)")

	// The three M6 synthesis verbs (itd-88) share one dual-mode shape: WITHOUT the
	// --*-json flag they run the core's deterministic evidence-only fallback (raw ==
	// nil); WITH the flag they validate an untrusted host-delegated payload. Every
	// core error — including the press-release whole-document refusal
	// (ErrPressReleaseUncited) — is a scrubbed exit 2; a per-entry drop is reported
	// honestly and stays exit 0.
	var principlesJSON string
	principlesCmd := &cobra.Command{
		Use:   "principles <lifeboat-dir> [--principles-json <file|->]",
		Short: "Distil principles from a packed lifeboat (deterministic from the ADRs, or validate host-produced principle JSON)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			dirAbs, err := filepath.Abs(args[0])
			if err != nil {
				return err
			}
			raw, err := readSynthesisPayload(cmd, principlesJSON)
			if err != nil {
				return &exitError{Code: 2, Msg: "disembark principles: " + scrubPaths(err)}
			}
			res, err := lifeboat.SynthesizePrinciples(dirAbs, raw)
			if err != nil {
				return &exitError{Code: 2, Msg: "disembark principles: " + scrubPaths(err)}
			}
			return render(cmd.OutOrStdout(), *asJSON, res, func(w io.Writer) {
				fmt.Fprint(w, res.Render())
			})
		},
	}
	principlesCmd.Flags().StringVar(&principlesJSON, "principles-json", "", "path to host-produced principle JSON (or - for stdin); absent runs deterministic mode")

	var pressReleaseJSON string
	pressReleaseCmd := &cobra.Command{
		Use:   "press-release <lifeboat-dir> [--press-release-json <file|->]",
		Short: "Compose the lifeboat's press release (deterministic from the brief/spine, or validate host-produced press-release JSON)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			dirAbs, err := filepath.Abs(args[0])
			if err != nil {
				return err
			}
			raw, err := readSynthesisPayload(cmd, pressReleaseJSON)
			if err != nil {
				return &exitError{Code: 2, Msg: "disembark press-release: " + scrubPaths(err)}
			}
			// A delegated press release citing nothing resolvable is a whole-document
			// refusal (ErrPressReleaseUncited): exit 2, the derived file left untouched
			// (design §5 exception). It flows through the generic exit-2 wrapping like
			// every other structural fault.
			res, err := lifeboat.ComposePressRelease(dirAbs, raw)
			if err != nil {
				return &exitError{Code: 2, Msg: "disembark press-release: " + scrubPaths(err)}
			}
			return render(cmd.OutOrStdout(), *asJSON, res, func(w io.Writer) {
				fmt.Fprint(w, res.Render())
			})
		},
	}
	pressReleaseCmd.Flags().StringVar(&pressReleaseJSON, "press-release-json", "", "path to host-produced press-release JSON (or - for stdin); absent runs deterministic mode")

	var reviewJSON string
	reviewCmd := &cobra.Command{
		Use:   "review <lifeboat-dir> <source-repo> [--review-json <file|->]",
		Short: "Review a packed lifeboat against its source repo — a registered verdict and cited findings (deterministic, or validate a host-produced verdict JSON)",
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) != 2 {
				return &exitError{Code: 2, Msg: "disembark review: <lifeboat-dir> <source-repo> are both required"}
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			dirAbs, err := filepath.Abs(args[0])
			if err != nil {
				return err
			}
			raw, err := readSynthesisPayload(cmd, reviewJSON)
			if err != nil {
				return &exitError{Code: 2, Msg: "disembark review: " + scrubPaths(err)}
			}
			// The source repo's content is never read (the core gates it as a real dir
			// only); a manifest failure is a MAJOR_RETHINK verdict input, exit 0.
			res, err := lifeboat.ReviewLifeboat(dirAbs, args[1], raw)
			if err != nil {
				return &exitError{Code: 2, Msg: "disembark review: " + scrubPaths(err)}
			}
			return render(cmd.OutOrStdout(), *asJSON, res, func(w io.Writer) {
				fmt.Fprint(w, res.Render())
			})
		},
	}
	reviewCmd.Flags().StringVar(&reviewJSON, "review-json", "", "path to the host-produced review verdict JSON (or - for stdin); absent runs deterministic mode")

	disembarkCmd.AddCommand(probeCmd)
	disembarkCmd.AddCommand(coverageCmd)
	disembarkCmd.AddCommand(planCmd)
	disembarkCmd.AddCommand(packCmd)
	disembarkCmd.AddCommand(graveyardCmd)
	disembarkCmd.AddCommand(principlesCmd)
	disembarkCmd.AddCommand(pressReleaseCmd)
	disembarkCmd.AddCommand(reviewCmd)
	return disembarkCmd
}

// newEmbarkCommand builds the operator `embark` sub-tree — the write half of the
// M5 record round-trip (itd-88, adr-35), the inverse of `disembark`:
//
//   - `probe <lifeboat-dir> [target-dir]` inspects a packed lifeboat against a
//     target read-only and reports what would land where, what conflicts would
//     block a write, which files are not embarked, the marker action, and the
//     coverage handoff. A plan WITH conflicts is a success (a report, exit 0).
//   - `from <lifeboat-dir> [target-dir]` writes the record families back into the
//     target through two-layer containment; on ANY conflict it refuses and writes
//     nothing (exit 1, one bulk report), re-injecting the current marker block
//     (never foreign prose) into the target CLAUDE.md.
//
// The target defaults to the working directory. Structural faults (not a lifeboat,
// schema too new, failed manifest verification, bad target) exit 2 with a scrubbed
// diagnostic; the conflict refusal exits 1 after rendering the report. `embark`
// backs the `/abcd:embark` command surface (commands/embark.md), so its
// surface-registry row is shipped.
func newEmbarkCommand(asJSON *bool) *cobra.Command {
	embarkCmd := &cobra.Command{
		Use:   "embark",
		Short: "Unpack a lifeboat's record families back into a target repo (probe read-only; from writes)",
		Args:  cobra.NoArgs,
		RunE:  helpRunE,
	}

	// resolveDirs turns the args into absolute lifeboat + target dirs; the target
	// defaults to the working directory when omitted.
	resolveDirs := func(args []string) (lbAbs, tgtAbs string, err error) {
		lbAbs, err = filepath.Abs(args[0])
		if err != nil {
			return "", "", err
		}
		target := "."
		if len(args) == 2 {
			target = args[1]
		}
		tgtAbs, err = filepath.Abs(target)
		return lbAbs, tgtAbs, err
	}

	probeCmd := &cobra.Command{
		Use:   "probe <lifeboat-dir> [target-dir]",
		Short: "Report what a lifeboat would write into a target, read-only (coverage blanks first)",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			lbAbs, tgtAbs, err := resolveDirs(args)
			if err != nil {
				return err
			}
			plan, err := lifeboat.EmbarkProbe(lbAbs, tgtAbs)
			if err != nil {
				return &exitError{Code: 2, Msg: fmt.Sprintf("embark probe: %s", scrubPaths(err))}
			}
			return render(cmd.OutOrStdout(), *asJSON, plan, func(w io.Writer) {
				fmt.Fprint(w, plan.Render())
			})
		},
	}

	fromCmd := &cobra.Command{
		Use:   "from <lifeboat-dir> [target-dir]",
		Short: "Write a lifeboat's record families into a target repo; refuses on any conflict",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			lbAbs, tgtAbs, err := resolveDirs(args)
			if err != nil {
				return err
			}
			res, err := lifeboat.EmbarkFrom(lbAbs, tgtAbs)
			if err != nil {
				// A conflict refusal is an EXPECTED outcome, not a fault: render the
				// bulk report, then propagate exit 1 with an empty message (the report
				// is the output, the exit code the only extra signal). Every other
				// error is a structural fault: exit 2, scrubbed to one line.
				if errors.Is(err, lifeboat.ErrEmbarkConflicts) {
					if rerr := render(cmd.OutOrStdout(), *asJSON, res, func(w io.Writer) {
						fmt.Fprint(w, res.Render())
					}); rerr != nil {
						return rerr
					}
					return &exitError{Code: 1}
				}
				return &exitError{Code: 2, Msg: fmt.Sprintf("embark from: %s", scrubPaths(err))}
			}
			return render(cmd.OutOrStdout(), *asJSON, res, func(w io.Writer) {
				fmt.Fprint(w, res.Render())
			})
		},
	}

	embarkCmd.AddCommand(probeCmd)
	embarkCmd.AddCommand(fromCmd)
	return embarkCmd
}

// readLessonsPayload reads the untrusted lesson JSON behind the trust guards,
// mirroring the intent verdict reader: a file must be a regular, non-symlink,
// size-capped file; "-" reads stdin bounded to the same cap. The cap is the
// exported lifeboat.MaxLessonsBytes. The file read uses fsutil.ReadGuarded so
// the symlink refusal, regular-file check, and size cap all happen on the open
// fd in one call — no lstat→ReadFile swap window.
func readLessonsPayload(cmd *cobra.Command, spec string) ([]byte, error) {
	if spec == "-" {
		return readCappedStdin(cmd, lifeboat.MaxLessonsBytes)
	}
	return readGuardedOperand(spec, lifeboat.MaxLessonsBytes)
}

// readSynthesisPayload reads an untrusted synthesis payload (principles,
// press-release, or review verdict JSON) behind the same trust guards as
// readLessonsPayload, capped at the exported lifeboat.MaxSynthesisBytes. An EMPTY
// spec (the flag absent) returns a nil slice — the sentinel the dual-mode cores
// read as "run deterministic mode", never a delegated payload. A "-" spec reads
// stdin bounded to the cap; a file must be regular, non-symlink, and under the cap.
func readSynthesisPayload(cmd *cobra.Command, spec string) ([]byte, error) {
	if spec == "" {
		return nil, nil // flag absent → deterministic mode (nil raw)
	}
	if spec == "-" {
		return readCappedStdin(cmd, lifeboat.MaxSynthesisBytes)
	}
	return readGuardedOperand(spec, lifeboat.MaxSynthesisBytes)
}

// readGuardedOperand reads an untrusted operand file behind fsutil.ReadGuarded
// (O_NOFOLLOW + regular-file on the open fd + size cap, one call, no lstat→read
// TOCTOU) and maps the guard's sentinels to clean, path-scrubbed messages —
// the shared body behind readLessonsPayload/readSynthesisPayload's file branch.
func readGuardedOperand(spec string, cap int64) ([]byte, error) {
	data, err := fsutil.ReadGuarded(spec, cap)
	if err != nil {
		switch {
		case errors.Is(err, fsutil.ErrNotRegular) || errors.Is(err, syscall.ELOOP):
			return nil, fmt.Errorf("%s is not a readable regular file (a symlink or non-regular operand is refused)", spec)
		case errors.Is(err, fsutil.ErrTooBig):
			return nil, fmt.Errorf("%s exceeds the %d-byte cap", spec, cap)
		default:
			return nil, err
		}
	}
	return data, nil
}

// maxHookStdinBytes caps the hook payload read from stdin (trust boundary).
const maxHookStdinBytes = 1 << 20 // 1 MiB

// hookInput is the subset of the Claude Code hook stdin payload the hook
// entrypoints read. Unknown fields are ignored.
type hookInput struct {
	SessionID string `json:"session_id"`
	Cwd       string `json:"cwd"`
	Prompt    string `json:"prompt"`
	Source    string `json:"source"`
	Event     string `json:"hook_event_name"`
	// TranscriptPath is supplied by the Stop hook; it names the session
	// transcript on disk. Read by `hook session-end` only.
	TranscriptPath string `json:"transcript_path"`
}

// readCappedStdin reads a "-" operand one byte past the cap so an over-cap
// payload is refused whole rather than truncated into a severed prefix — the
// same refuse-whole guarantee readGuardedOperand gives the file transport
// (iss-201's class; spc-4's refuse-whole invariant on the transcript path). A
// bare LimitReader(cap) reads exactly cap bytes, so an over-cap payload is
// silently cut and its length-cap refusal never fires; the cap+1 probe closes
// that.
func readCappedStdin(cmd *cobra.Command, cap int64) ([]byte, error) {
	raw, err := io.ReadAll(io.LimitReader(cmd.InOrStdin(), cap+1))
	if err != nil {
		return nil, err
	}
	if int64(len(raw)) > cap {
		return nil, fmt.Errorf("stdin payload exceeds the %d-byte cap", cap)
	}
	return raw, nil
}

// readHookInput reads and size-caps the hook stdin payload. It reads one byte
// past the cap so an over-cap payload is reported as such rather than
// truncated into a severed prefix that json.Unmarshal misblames as malformed
// host JSON (iss-201's class; guardCandidate is the pattern).
func readHookInput(cmd *cobra.Command) (hookInput, error) {
	raw, err := io.ReadAll(io.LimitReader(cmd.InOrStdin(), maxHookStdinBytes+1))
	if err != nil {
		return hookInput{}, err
	}
	if len(raw) > maxHookStdinBytes {
		return hookInput{}, fmt.Errorf("payload is over the %d-byte cap; it was discarded unparsed", maxHookStdinBytes)
	}
	var in hookInput
	if err := json.Unmarshal(raw, &in); err != nil {
		return hookInput{}, err
	}
	return in, nil
}

// hookSession returns a stable session key, defaulting when the harness omits
// the id (the hash in the state layer neutralises any hostile value). The
// harness supplies session_id in practice; the "default" fallback means two
// concurrent id-less sessions would share one dedup ledger — an accepted
// edge-case degradation, never a correctness or safety issue.
func hookSession(in hookInput) string {
	if in.SessionID == "" {
		return "default"
	}
	return in.SessionID
}

// newHookCommand builds the operator-internal `hook` sub-tree: the Claude Code
// prompt-router entrypoints (itd-3). These are NOT a user surface — they are the
// injection transport, one front door onto internal/core/rules alongside the
// `abcd rules` verb. Every path is fail-closed and NON-blocking: a malformed
// payload, an unreadable rules.json, or a state error injects nothing, logs a
// diagnostic to stderr (out-of-band, per D3), and exits 0 so it can never wedge
// a session.
func newHookCommand() *cobra.Command {
	hookCmd := &cobra.Command{
		Use:    "hook",
		Short:  "Claude Code hook entrypoints (operator-internal)",
		Hidden: true,
		Args:   failOpenNoArgs,
		RunE:   helpRunE,
	}

	// prompt-router — UserPromptSubmit: recall-match, dedup, inject.
	hookCmd.AddCommand(&cobra.Command{
		Use:   "prompt-router",
		Short: "UserPromptSubmit: inject the rules matching the prompt",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			in, err := readHookInput(cmd)
			if err != nil {
				fmt.Fprintf(cmd.ErrOrStderr(), "abcd rules: unreadable hook payload (%v); injecting nothing\n", err)
				return nil
			}
			cwd := in.Cwd
			if cwd == "" {
				if wd, err := os.Getwd(); err == nil {
					cwd = wd
				}
			}
			root := rulesRoot(cwd)
			rs, err := rules.Load(root)
			if err != nil {
				fmt.Fprintf(cmd.ErrOrStderr(), "abcd rules: %v; injecting nothing\n", err)
				return nil
			}
			session := hookSession(in)
			// The fixed-N backstop comes from the repo's config (default 15 when
			// unset); event-driven reset is the primary refresh (D1).
			res := rules.Inject(rs, in.Prompt, rules.LoadState(session), rules.LoadBackstop(root))
			if err := rules.SaveState(session, res.State); err != nil {
				fmt.Fprintf(cmd.ErrOrStderr(), "abcd rules: state save failed (%v)\n", err)
			}
			fmt.Fprintf(cmd.ErrOrStderr(), "abcd rules: turn %d, injected %d domain(s) %v, %d bytes\n",
				res.State.Count, len(res.Injected), res.Injected, len(res.Text))
			if res.Text != "" {
				fmt.Fprint(cmd.OutOrStdout(), res.Text)
			}
			return nil
		},
	})

	// prompt-router-reset — SessionStart / PreCompact: clear the dedup ledger so
	// the next prompt re-injects (the event-driven refresh, D1/B2).
	hookCmd.AddCommand(&cobra.Command{
		Use:   "prompt-router-reset",
		Short: "SessionStart/PreCompact: clear the dedup ledger",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			in, err := readHookInput(cmd)
			if err != nil {
				fmt.Fprintf(cmd.ErrOrStderr(), "abcd rules: unreadable reset payload (%v)\n", err)
				return nil
			}
			session := hookSession(in)
			if err := rules.ResetState(session); err != nil {
				fmt.Fprintf(cmd.ErrOrStderr(), "abcd rules: reset failed (%v)\n", err)
				return nil
			}
			// SessionStart is a natural sweep point for stale ledgers.
			rules.PruneState(rules.StateTTL)
			// %q quotes the untrusted hook_event_name so an embedded newline or
			// ANSI escape cannot spoof the operator's diagnostic stream.
			fmt.Fprintf(cmd.ErrOrStderr(), "abcd rules: reset session (%q)\n", in.Event)
			return nil
		},
	})

	// session-end — SessionEnd: redact and store the session transcript (adr-29).
	//
	// Wired to SessionEnd, NOT Stop. The plan said Stop, but Stop fires once per
	// assistant *turn*: a 40-turn session would store 40 growing supersets of one
	// transcript, since Capture's sha256 dedup only collapses byte-identical
	// re-captures and a live transcript grows between turns. SessionEnd fires once
	// when the session terminates, which is the session-granular record the gate
	// asks for. SessionEnd also ignores exit code and stdout by contract, which
	// matches this verb's fail-closed, non-blocking shape exactly.
	//
	// This is a new verb because `history capture` cannot be wired to a hook: from
	// stdin it *requires* --session <id>, and the hook delivers its session id
	// inside a JSON payload, not as a flag.
	//
	// It is the only irreversible thing abcd does. A session that ends without
	// being captured is gone: no later code can reconstruct a transcript that was
	// never stored. That asymmetry — a missed capture is permanent, a failed
	// capture is merely a lost session — is why every path here degrades to "log
	// and exit 0" rather than surfacing an error to the host.
	hookCmd.AddCommand(&cobra.Command{
		Use:   "session-end",
		Short: "SessionEnd: redact and store the session transcript",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			// Diagnostics go to stderr, out of band; stdout stays empty, since a
			// Stop hook's stdout is not a place to speak to the model.
			warn := func(format string, a ...any) error {
				fmt.Fprintf(cmd.ErrOrStderr(), "abcd history: "+format+"\n", a...)
				return nil // never non-zero: a Stop hook must not wedge the session
			}

			in, err := readHookInput(cmd)
			if err != nil {
				return warn("unreadable Stop payload (%v); capturing nothing", err)
			}
			if in.TranscriptPath == "" {
				return warn("Stop payload carries no transcript_path; capturing nothing")
			}
			cwd := in.Cwd
			if cwd == "" {
				if wd, err := os.Getwd(); err == nil {
					cwd = wd
				}
			}
			det, err := ahoy.Detect(cwd)
			if err != nil || det.RootSHA == "" {
				return warn("cannot resolve the repo's root-commit SHA from %q; capturing nothing", cwd)
			}
			raw, err := readTranscript(in.TranscriptPath)
			if err != nil {
				return warn("%v; capturing nothing", err)
			}
			res, err := history.Capture(captureRoot(cwd), det.RootSHA, in.SessionID, raw, "native")
			if err != nil {
				// Includes a hostile session id and a redaction hard-fail: both
				// write nothing, by design in internal/core/history.
				return warn("capture failed (%v)", err)
			}
			if !res.Wrote {
				return warn("session %s already stored (no-op)", res.Record.SessionID)
			}
			fmt.Fprintf(cmd.ErrOrStderr(), "abcd history: stored %s; redacted secrets=%d home=%d\n",
				res.Record.SessionID, res.Record.Secrets, res.Record.HomePaths)
			return nil
		},
	})

	// session-start — SessionStart: warn, visibly, when the transcript store is
	// not bootstrapped for this repo (iss-95).
	//
	// The session-end hook cannot say this at session end — SessionEnd ignores a
	// hook's exit code and stdout, so a not-installed session captures nothing and
	// no one is told. SessionStart is the one session hook with a user-visible
	// channel: it renders a hook's stderr as a notice ONLY on a non-zero exit, and
	// never blocks the session on it. So the "loud" path here deliberately exits 2
	// — the only way the warning reaches the user — while every other path stays
	// silent at exit 0. `ahoy install` and `ahoy doctor` already handle the
	// installed and health-check cases; this covers the plugin-enabled-but-never-
	// installed gap that was otherwise silent.
	hookCmd.AddCommand(&cobra.Command{
		Use:   "session-start",
		Short: "SessionStart: warn about an unbootstrapped store or a stale binary",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			in, err := readHookInput(cmd)
			if err != nil {
				return nil // malformed payload: not a store problem, stay silent
			}
			cwd := in.Cwd
			if cwd == "" {
				if wd, err := os.Getwd(); err == nil {
					cwd = wd
				}
			}
			var notices []string
			// history.transcripts_missing is emitted only when cwd is a git repo
			// (a root SHA resolved) AND this repo's transcripts dir is absent —
			// exactly the state in which session-end would silently capture
			// nothing. A non-repo cwd never carries this gap, so we stay silent
			// there: no `ahoy install` would make a non-repo capturable. A
			// detection that errored tells us nothing: never nag on uncertainty.
			if det, err := ahoy.Detect(cwd); err == nil {
				for _, g := range det.Gaps {
					if g.ID == "history.transcripts_missing" {
						notices = append(notices,
							"abcd: session transcripts will not be captured — the history store is not set up for this repo. Run `/abcd:ahoy install` (or `abcd ahoy install`) to start recording.")
						break
					}
				}
			}
			// The skew notice is a plugin-root fact, not a repo one, so it stands
			// whatever the repo detection above could answer (itd-105).
			if n := binarySkewNotice(); n != "" {
				notices = append(notices, n)
			}
			// itd-111: a dogfood binary behind (or dirty against) its own source
			// checkout tip. os.Executable names the binary; the comparison is
			// git-only and never touches the network (adr-38 tier 1).
			if exe, err := os.Executable(); err == nil {
				if n := stalenessNotice(cwd, exe); n != "" {
					notices = append(notices, n)
				}
			}
			// itd-111 (AC6): a version transition performed since this repo was
			// last set up — the running binary differs from the recorded
			// setup_version. Report only; the fetch that changed it is
			// provisioning's job. Both values come from disk (config + build info).
			if from, to, changed := ahoy.VersionTransition(cwd); changed {
				notices = append(notices, fmt.Sprintf(
					"abcd: the running binary is version %s, but this repo was last set up with %s — run `/abcd:ahoy install` (or `abcd ahoy install`) to reconcile the recorded version.",
					termsafe.Sanitize(to), termsafe.Sanitize(from)))
			}
			if len(notices) == 0 {
				return nil
			}
			for _, n := range notices {
				fmt.Fprintln(cmd.ErrOrStderr(), n)
			}
			return &exitError{Code: 2} // non-zero so SessionStart shows it; SessionStart never blocks
		},
	})

	return hookCmd
}

// maxTranscriptBytes caps the transcript read from disk. Generous for a JSONL
// session log, and bounded so a pathological file cannot stall the Stop hook
// while the scanner walks it.
const maxTranscriptBytes = 64 << 20 // 64 MiB

// readTranscript reads the file named by the Stop payload's transcript_path.
//
// The path is external input, so the read goes through fsutil.ReadGuarded —
// O_NOFOLLOW so a planted symlink is refused rather than followed, O_NONBLOCK
// so a FIFO or device node cannot hang the hook (a hung Stop hook wedges the
// user's session), a regular-file check on the opened descriptor, and a cap+1
// probe so a file that grows past the cap between stat and read is refused
// whole instead of stored silently truncated (iss-347).
func readTranscript(path string) ([]byte, error) {
	raw, err := fsutil.ReadGuarded(path, maxTranscriptBytes)
	if err != nil {
		switch {
		case errors.Is(err, fsutil.ErrNotRegular) || errors.Is(err, syscall.ELOOP):
			return nil, fmt.Errorf("transcript %q is not a readable regular file (a symlink or non-regular transcript is refused)", path)
		case errors.Is(err, fsutil.ErrTooBig):
			return nil, fmt.Errorf("transcript %q is over the %d-byte cap", path, maxTranscriptBytes)
		default:
			return nil, fmt.Errorf("cannot read transcript %q (%v)", path, err)
		}
	}
	if len(raw) == 0 {
		return nil, fmt.Errorf("transcript %q is empty", path)
	}
	return raw, nil
}

// rulesView is the machine-readable envelope for bare `abcd rules`: the kill
// switch plus the active domains.
type rulesView struct {
	Disabled bool                   `json:"disabled"`
	Domains  []rules.ResolvedDomain `json:"domains"`
}

// newRulesCommand builds the `rules` verb — the vendor-neutral front door onto
// internal/core/rules (itd-3). Bare `abcd rules` renders the active rule set;
// a positional DOMAIN scopes to one domain (case-insensitive). Read-only,
// diagnostic — it never mutates and there is no `show` sub-verb (the positional
// argument is the scope, per the bare-command-as-render discipline).
func newRulesCommand(asJSON *bool) *cobra.Command {
	return &cobra.Command{
		Use:   "rules [domain]",
		Short: "Render the active rule set; a positional DOMAIN scopes to one (read-only)",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cwd, err := os.Getwd()
			if err != nil {
				return err
			}
			rs, err := rules.Load(rulesRoot(cwd))
			if err != nil {
				return err
			}
			// Scoped: inspect one domain's configured content regardless of its
			// state OR the kill switch — this diagnostic shows what a domain holds,
			// not what would inject right now (bare `abcd rules` reports disabled).
			if len(args) == 1 {
				name := strings.ToUpper(args[0])
				d, ok := rs.Lookup(name)
				if !ok {
					return &exitError{Code: 2, Msg: fmt.Sprintf("abcd rules: unknown domain %q", name)}
				}
				return render(cmd.OutOrStdout(), *asJSON, d, func(w io.Writer) {
					fmt.Fprint(w, rules.Render([]rules.ResolvedDomain{d}))
				})
			}
			// Bare: render the full active set.
			active := rs.Active()
			return render(cmd.OutOrStdout(), *asJSON, rulesView{Disabled: rs.Disabled, Domains: active}, func(w io.Writer) {
				if rs.Disabled {
					fmt.Fprintln(w, "abcd rules — disabled (kill switch set in .abcd/rules.json)")
					return
				}
				if out := rules.Render(active); out != "" {
					fmt.Fprint(w, out)
					return
				}
				fmt.Fprintln(w, "abcd rules — no active domains")
			})
		},
	}
}

// newIntentCommand builds the `intent` verb — the front door onto
// internal/core/intent (itd-80). Bare `abcd intent` renders the read-only
// lifecycle status board (never mutates); the `plan` and `link` sub-verbs carry
// the mutations. Usage/lookup failures exit 2.
func newIntentCommand(asJSON *bool) *cobra.Command {
	var intentImpact string
	intentCmd := &cobra.Command{
		Use:   "intent [text]",
		Short: "Intent lifecycle; bare invocation is read-only status, quoted text files a draft",
		Args:  cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cwd, err := os.Getwd()
			if err != nil {
				return err
			}
			// Quoted text (no sub-verb) files a new draft — the symmetric create path
			// (itd-46). Bare invocation stays read-only status + help (never mutates).
			if len(args) > 0 {
				// Guard: a mistyped subcommand (e.g. `intent lnk itd-5`) must not be
				// swallowed as draft text and filed. Mirrors the capture guard
				// (unrecognized-input-never-writes, iss-29); genuine prose still files.
				if sug, ok := suspectedTypoedSubcommand(cmd, args); ok {
					return &exitError{Code: 2, Msg: fmt.Sprintf(
						"unknown intent subcommand %q; did you mean %q? (nothing created — reword the text if you meant to file a draft)",
						args[0], sug)}
				}
				return createIntentFromText(cmd, cwd, strings.Join(args, " "), intentImpact, *asJSON)
			}
			v, err := intent.Status(cwd)
			if err != nil {
				return err
			}
			return render(cmd.OutOrStdout(), *asJSON, v, func(w io.Writer) {
				fmt.Fprintf(w, "abcd intent — drafts %d · planned %d · shipped %d · disciplines %d · superseded %d\n",
					v.Buckets[intent.BucketDrafts], v.Buckets[intent.BucketPlanned], v.Buckets[intent.BucketShipped],
					v.Buckets[intent.BucketDisciplines], v.Buckets[intent.BucketSuperseded])
				fmt.Fprintf(w, "  specs: open %d · closed %d\n", v.SpecsOpen, v.SpecsClosed)
				for _, p := range v.Linked {
					// p.Spec is the intent file's spec_id frontmatter, not charset-validated.
					fmt.Fprintf(w, "  link: %s -> %s\n", termsafe.Sanitize(p.Intent), termsafe.Sanitize(p.Spec))
				}
				fmt.Fprint(w, ledgerDecisionRule)
				fmt.Fprint(w, ideateRoutingRule)
			})
		},
	}
	// --impact stamps an optional product judgement onto the seeded draft. It is
	// optional (a draft is "not judged yet"), but when set it is validated and
	// travels unchanged to shipped/, where intent_impact_valid requires it — so the
	// tool's own create->plan->ship path can produce a record that clears the gate.
	intentCmd.Flags().StringVar(&intentImpact, "impact", "", "stamp the draft's product impact: additive|breaking|fix (optional)")

	// new "<text>" — backwards-compatible alias for the sub-verb-free create path
	// (itd-46, lean a): routes to the same create engine and warns on stderr that
	// the `new` sub-verb is deprecated in favour of `abcd intent "<text>"`. The
	// stdout artefact is identical to the quoted-text form.
	intentCmd.AddCommand(&cobra.Command{
		Use:   "new <text>",
		Short: "Deprecated alias for `abcd intent \"<text>\"` (files a draft from the text)",
		Args:  cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cwd, err := os.Getwd()
			if err != nil {
				return err
			}
			if len(args) == 0 {
				return &exitError{Code: 2, Msg: "abcd intent new: text is required — use `abcd intent \"<text>\"`"}
			}
			fmt.Fprintln(cmd.ErrOrStderr(),
				"WARNING: `abcd intent new` is deprecated; use `abcd intent \"<text>\"` (quoted text is the create signal).")
			return createIntentFromText(cmd, cwd, strings.Join(args, " "), "", *asJSON)
		},
	})

	// plan <itd-N> — mint the spec, write both link sides, move drafts -> planned.
	intentCmd.AddCommand(&cobra.Command{
		Use:   "plan <itd-N>",
		Short: "Plan a draft intent: mint its spec, link both sides, move drafts -> planned",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cwd, err := os.Getwd()
			if err != nil {
				return err
			}
			res, err := intent.Plan(cwd, args[0])
			if err != nil {
				return &exitError{Code: 2, Msg: "abcd intent plan: " + err.Error()}
			}
			emitMintWarning(cmd, res.MintWarning)
			return render(cmd.OutOrStdout(), *asJSON, res, func(w io.Writer) {
				fmt.Fprintf(w, "abcd intent plan — %s drafts -> planned, linked %s\n", res.Intent.ID, res.Spec.ID)
				fmt.Fprintf(w, "  intent: %s\n", res.Intent.Path)
				fmt.Fprintf(w, "  spec:   %s\n", res.Spec.Path)
			})
		},
	})

	// ready <itd-N> — the read-only implement-readiness gate. Exit codes are the
	// machine seam an autonomous run gates on: 0 ready, 1 not ready (the rendered
	// report is the output, empty message — the embark-conflicts precedent), 2
	// structural fault.
	intentCmd.AddCommand(&cobra.Command{
		Use:   "ready <itd-N>",
		Short: "Report whether an intent is ready to implement (planned + AC + written spec); exit 1 when not",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cwd, err := os.Getwd()
			if err != nil {
				return err
			}
			res, err := intent.Ready(cwd, args[0])
			if err != nil {
				return &exitError{Code: 2, Msg: "abcd intent ready: " + err.Error()}
			}
			if rerr := render(cmd.OutOrStdout(), *asJSON, res, func(w io.Writer) {
				verdict := "READY"
				if !res.Ready {
					verdict = "NOT READY"
				}
				fmt.Fprintf(w, "abcd intent ready — %s %s (%s)\n", termsafe.Sanitize(res.IntentID), verdict, termsafe.Sanitize(res.Bucket))
				for _, c := range res.Checks {
					mark := "[ ok ]"
					if !c.OK {
						mark = "[fail]"
					}
					// Detail/remedy interpolate frontmatter values, not charset-validated.
					fmt.Fprintf(w, "  %s %s: %s\n", mark, c.Name, termsafe.Sanitize(c.Detail))
					if c.Remedy != "" {
						fmt.Fprintf(w, "         remedy: %s\n", termsafe.Sanitize(c.Remedy))
					}
				}
			}); rerr != nil {
				return rerr
			}
			if !res.Ready {
				return &exitError{Code: 1}
			}
			return nil
		},
	})

	// link <itd-N> <spc-N> — retroactively set spec_id on a planned intent.
	intentCmd.AddCommand(&cobra.Command{
		Use:   "link <itd-N> <spc-N>",
		Short: "Link a planned intent to an existing spec (writes the intent's spec_id)",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cwd, err := os.Getwd()
			if err != nil {
				return err
			}
			res, err := intent.Link(cwd, args[0], args[1])
			if err != nil {
				return &exitError{Code: 2, Msg: "abcd intent link: " + err.Error()}
			}
			return render(cmd.OutOrStdout(), *asJSON, res, func(w io.Writer) {
				fmt.Fprintf(w, "abcd intent link — %s -> %s\n  intent: %s\n", res.Intent.ID, res.Spec.ID, res.Intent.Path)
			})
		},
	})

	intentCmd.AddCommand(newIntentAuditCommand(asJSON))
	return intentCmd
}

// ledgerDecisionRule is the one-line capture-vs-intent decision rule shown in
// both ledgers' bare-form help (itd-46 AC5), so a user knows which ledger to reach
// for. It stays host-agnostic (binary command forms, no plugin/tool names).
const ledgerDecisionRule = "  which ledger? half-formed observation, question, or nitpick -> `abcd capture \"…\"`; a user-facing change you want to ship -> `abcd intent \"…\"`\n"

// ideateRoutingRule sits beside the ledger rule and names the optional third
// route: a big, unproven idea can go through the admission gauntlet first
// (itd-104 AC1).
//
// It is a POINTER, and the wording is load-bearing. Ideate is never a
// precondition for capture or intent, so this line offers a route and never
// implies one is missed — no "should", no "first", no warning when it is skipped.
// The one line of capture friction the ledgers promise stays one line.
const ideateRoutingRule = "  a big, unproven idea? `abcd ideate` runs the optional admission gauntlet and records the verdict either way\n"

// createIntentFromText is the shared quoted-text create path behind both
// `abcd intent "<text>"` and the deprecated `abcd intent new "<text>"` alias: it
// files a new draft via intent.CreateFromText and renders the created record. The
// engine refuses empty/whitespace text and mints the id under the store lock, so
// this surface stays a thin marshaller.
func createIntentFromText(cmd *cobra.Command, cwd, text, impact string, asJSON bool) error {
	it, mintWarning, err := intent.CreateFromText(cwd, text, impact)
	if err != nil {
		return &exitError{Code: 2, Msg: "abcd intent: " + err.Error()}
	}
	emitMintWarning(cmd, mintWarning)
	return render(cmd.OutOrStdout(), asJSON, it, func(w io.Writer) {
		fmt.Fprintf(w, "created %s (%s) — %s\n", it.ID, it.Bucket, it.Path)
	})
}

// emitMintWarning prints a record-id mint degrade note to stderr (loud-staging:
// a stage that degraded to working-tree-only minting must say so, never silently
// fall back). The note is engine-produced and path-free; it is sanitised anyway
// before it touches the terminal. Empty warnings emit nothing.
func emitMintWarning(cmd *cobra.Command, warning string) {
	if warning == "" {
		return
	}
	fmt.Fprintln(cmd.ErrOrStderr(), "warning: "+termsafe.Sanitize(warning))
}

// newIntentAuditCommand builds `abcd intent audit`: `ingest --verdict-json`
// applies a host-produced intent-audit verdict to the shipped intent's Audit
// Notes (fail-closed: ingested | dead_letter | noop); bare `audit <itd-N>`
// re-emits the OWED stub + ephemeral request for a shipped intent.
func newIntentAuditCommand(asJSON *bool) *cobra.Command {
	auditCmd := &cobra.Command{
		Use:   "audit [<itd-N>]",
		Short: "Intent audit (promise vs delivered): re-emit a shipped intent's request, or ingest a verdict",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return cmd.Help()
			}
			cwd, err := os.Getwd()
			if err != nil {
				return err
			}
			res, err := intent.ReEmitAudit(cwd, args[0])
			if err != nil {
				return &exitError{Code: 2, Msg: "abcd intent audit: " + err.Error()}
			}
			return render(cmd.OutOrStdout(), *asJSON, res, func(w io.Writer) {
				fmt.Fprintf(w, "abcd intent audit — %s %s (receipt %s)\n  request: %s\n",
					res.IntentID, res.Status, res.ReceiptID, res.RequestPath)
			})
		},
	}

	var verdictJSON string
	ingestCmd := &cobra.Command{
		Use:   "ingest --verdict-json <path>",
		Short: "Ingest an intent-audit verdict JSON into the shipped intent's Audit Notes",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			cwd, err := os.Getwd()
			if err != nil {
				return err
			}
			if verdictJSON == "" {
				return &exitError{Code: 2, Msg: "abcd intent audit ingest: --verdict-json <path> is required"}
			}
			res, err := intent.IngestVerdict(cwd, verdictJSON)
			if err != nil {
				return &exitError{Code: 2, Msg: "abcd intent audit ingest: " + err.Error()}
			}
			return render(cmd.OutOrStdout(), *asJSON, res, func(w io.Writer) {
				fmt.Fprintf(w, "abcd intent audit ingest — %s (receipt %s, intent %s)\n", res.Status, res.ReceiptID, res.IntentID)
				switch res.Status {
				case "ingested":
					fmt.Fprintf(w, "  criteria %d: MET %d · MET_WITH_CONCERNS %d · NOT_MET %d · INCONCLUSIVE %d\n",
						res.Criteria, res.Met, res.MetWithConcern, res.NotMet, res.Inconclusive)
				case "dead_letter":
					fmt.Fprintf(w, "  DEAD_LETTER: %s\n  raw payload: %s\n", res.Reason, res.DeadLetterPath)
				}
			})
		},
	}
	ingestCmd.Flags().StringVar(&verdictJSON, "verdict-json", "", "path to the intent-audit verdict JSON")
	auditCmd.AddCommand(ingestCmd)
	return auditCmd
}

// specStatusView is the machine-readable envelope for bare `abcd spec`: the
// open/closed counts and every discovered spec record.
type specStatusView struct {
	Open   int         `json:"open"`
	Closed int         `json:"closed"`
	Specs  []spec.Spec `json:"specs"`
}

// newSpecCommand builds the `spec` verb — the front door onto internal/core/spec
// (itd-80). Bare `abcd spec` renders the read-only spec-store status; the `close`
// sub-verb closes a spec AND reconciles its linked intent (planned -> shipped)
// via intent.Reconcile, so one command completes the lifecycle transition.
func newSpecCommand(asJSON *bool) *cobra.Command {
	specCmd := &cobra.Command{
		Use:   "spec",
		Short: "Native spec store; bare invocation is read-only status",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			cwd, err := os.Getwd()
			if err != nil {
				return err
			}
			store, err := spec.Load(cwd)
			if err != nil {
				return err
			}
			view := specStatusView{Specs: store.Specs}
			for _, sp := range store.Specs {
				if sp.Status == spec.StatusClosed {
					view.Closed++
				} else {
					view.Open++
				}
			}
			return render(cmd.OutOrStdout(), *asJSON, view, func(w io.Writer) {
				fmt.Fprintf(w, "abcd spec — open %d · closed %d\n", view.Open, view.Closed)
				for _, sp := range store.Specs {
					fmt.Fprintf(w, "  %s  %s  %s  (%s)\n", termsafe.Sanitize(sp.ID), sp.Status, termsafe.Sanitize(sp.Slug), termsafe.Sanitize(sp.Intent))
				}
			})
		},
	}

	// close <spc-N> — closes the spec AND reconciles the linked intent
	// (planned -> shipped). Fail-closed and idempotent (see intent.Reconcile).
	specCmd.AddCommand(&cobra.Command{
		Use:   "close <spc-N>",
		Short: "Close a spec (open/ -> closed/) and ship its linked intent (planned/ -> shipped/)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cwd, err := os.Getwd()
			if err != nil {
				return err
			}
			res, err := intent.Reconcile(cwd, args[0])
			if err != nil {
				return &exitError{Code: 2, Msg: "abcd spec close: " + err.Error()}
			}
			// The fidelity-review emit is report-only: a failure does NOT fail the
			// close (the intent already shipped), but it is surfaced loudly on stderr.
			if res.AuditEmitError != "" {
				fmt.Fprintf(cmd.ErrOrStderr(), "WARNING: abcd spec close — fidelity-review emit failed for %s (intent shipped anyway): %s\n", res.Intent.ID, res.AuditEmitError)
			}
			return render(cmd.OutOrStdout(), *asJSON, res, func(w io.Writer) {
				fmt.Fprintf(w, "abcd spec close — %s open -> closed\n  %s\n", res.Spec.ID, res.Spec.Path)
				if res.IntentMoved {
					fmt.Fprintf(w, "  reconciled intent %s: %s -> %s\n", res.Intent.ID, res.From, res.To)
				} else {
					fmt.Fprintf(w, "  intent %s already %s (no move)\n", res.Intent.ID, res.To)
				}
				if res.ReceiptID != "" {
					fmt.Fprintf(w, "  fidelity review OWED: receipt %s\n", res.ReceiptID)
				}
			})
		},
	})

	return specCmd
}

// newAhoyCommand builds the `ahoy` sub-tree. Bare `ahoy` runs the read-only
// detection pass (abcd's convention: bare invocation never mutates); the
// install/uninstall/doctor/dry-run sub-verbs are thin consumers of the same
// core engine (detect -> contract -> apply), matching 04-surfaces/01-ahoy.md.
func newAhoyCommand(asJSON *bool) *cobra.Command {
	ahoyCmd := &cobra.Command{
		Use:   "ahoy",
		Short: "Install/update abcd in this repo; bare invocation is read-only status",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			cwd, err := os.Getwd()
			if err != nil {
				return err
			}
			res, err := ahoy.DryRun(cwd)
			if err != nil {
				return err
			}
			// Vintage + staleness from the shared comparator (itd-111): the same
			// source `abcd version` and the session-start notice read. Computed
			// once and carried in both the JSON and the text render.
			vin := ahoy.Vintage(cwd)
			out := ahoyOutput{DetectionResult: res, Vintage: vin.DisplayVintage(), Staleness: vin.Staleness()}
			return render(cmd.OutOrStdout(), *asJSON, out, func(w io.Writer) {
				fmt.Fprintf(w, "abcd ahoy — %s\n", res.FolderKind)
				fmt.Fprintf(w, "  plugin root: %s\n", res.PluginRootStatus)
				fmt.Fprintf(w, "  root sha:    %s\n", res.RootSHA)
				if mode, _ := res.Signals["install_mode"].(string); mode != "" {
					fmt.Fprintf(w, "  install:     %s\n", mode)
				}
				fmt.Fprintf(w, "  vintage:     %s\n", out.Vintage)
				fmt.Fprintf(w, "  staleness:   %s\n", out.Staleness)
				// The citation baseline's coverage and age, present only in a repo
				// that has armed the citation gate. The line embeds counts and a
				// date derived from repo content, so it is sanitised.
				if citations, _ := res.Signals["citations"].(string); citations != "" {
					fmt.Fprintf(w, "  citations:   %s\n", termsafe.Sanitize(citations))
				}
				fmt.Fprintf(w, "  gaps:        %d\n", len(res.Gaps))
				if res.FolderKind != ahoy.UnmanagedFolder {
					fmt.Fprintf(w, "  guard:       %s\n", guardHealthLine(res.Guard))
					for i, line := range banlistHealthLines(*res.Banlist) {
						label := "  banlist:     "
						if i > 0 {
							label = "               "
						}
						fmt.Fprintf(w, "%s%s\n", label, line)
					}
					fmt.Fprintf(w, "               reach: %s\n", res.Banlist.Reach)
				}
				// Classification is read-only; the human report names the
				// next step per folder kind (itd-40 AC2/AC3).
				switch res.FolderKind {
				case ahoy.UnmanagedRepo:
					fmt.Fprintf(w, "  unmanaged git repo — run `/abcd:ahoy install` to adopt it.\n")
				case ahoy.UnmanagedFolder:
					fmt.Fprintf(w, "  not a git repository — nothing to act on.\n")
				}
			})
		},
	}

	// install
	var (
		yes           bool
		adopt         bool
		refuseAdopt   bool
		dev           bool
		allowStale    bool
		binDir        string
		visibility    string
		docsTarget    string
		oracleBackend string
		scanDeep      string
	)
	installCmd := &cobra.Command{
		Use:   "install",
		Short: "Install or update abcd in this repo (idempotent)",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			cwd, err := os.Getwd()
			if err != nil {
				return err
			}
			opts, err := installOptionsFromFlags(cmd, yes, adopt, refuseAdopt, dev, allowStale, binDir, visibility, docsTarget, oracleBackend, scanDeep)
			if err != nil {
				return err
			}
			res, err := ahoy.Install(cwd, opts, newPrompter(cmd))
			if err != nil {
				return err
			}
			return render(cmd.OutOrStdout(), *asJSON, res, func(w io.Writer) {
				fmt.Fprintf(w, "abcd ahoy install — %s\n", res.Status)
				for _, c := range res.Changes {
					fmt.Fprintf(w, "  changed: %s\n", c)
				}
				for _, p := range res.Writes {
					fmt.Fprintf(w, "  wrote: %s\n", p)
				}
				// A refusal is louder than an unexplained missing write: it says
				// what abcd did not do and why (a dangling PATH entry it declined
				// to create, a directory it could not write).
				for _, n := range res.Notes {
					fmt.Fprintf(w, "  note: %s\n", n)
				}
				if len(res.DeclinedCategories) > 0 {
					fmt.Fprintf(w, "  declined: %s\n", strings.Join(res.DeclinedCategories, ", "))
				}
				if len(res.Remaining) > 0 {
					fmt.Fprintf(w, "  remaining gaps: %s\n", strings.Join(res.Remaining, ", "))
				}
				// --yes approves every category but never writes the identity
				// pin, so say which optional work it left and how to apply it.
				if len(res.OptionalSkipped) > 0 {
					fmt.Fprintf(w, "  optional, not covered by --yes: %s\n", strings.Join(res.OptionalSkipped, ", "))
					fmt.Fprint(w, "    the pin records the current git identity, so it is only written against an answered prompt:\n")
					fmt.Fprint(w, "    run `abcd ahoy install` (no --yes) and answer y at each prompt — non-interactively, `yes | abcd ahoy install`\n")
				}
			})
		},
	}
	// No backquotes in a flag's usage string: cobra reads the first backquoted
	// word as the flag's argument placeholder, so a quoted answer would render
	// this boolean as "--yes y" in the help and the generated reference.
	installCmd.Flags().BoolVar(&yes, "yes", false, "approve every resolvable change category without prompting; excludes the optional git-identity pin, which needs an answered prompt (run without --yes, or answer every prompt with: yes | abcd ahoy install)")
	installCmd.Flags().BoolVar(&adopt, "adopt", false, "adopt an unmanaged repo without prompting")
	installCmd.Flags().BoolVar(&refuseAdopt, "refuse-adopt", false, "decline to adopt an unmanaged repo")
	installCmd.Flags().BoolVar(&dev, "dev", false, "track-latest dogfood mode: the PATH entry rebuilds from the source tip on every call instead of pinning the built binary")
	installCmd.Flags().BoolVar(&allowStale, "allow-stale-binary", false, "proceed even when the running binary is stale against its source tip or its vintage cannot be determined; the default is to refuse before any write and name the rebuild fix")
	installCmd.Flags().StringVar(&binDir, "bin-dir", "", "directory for the PATH entry (default ~/.local/bin, or an existing abcd install adopted in place); fails when it is not writable — abcd never escalates privileges")
	installCmd.Flags().StringVar(&visibility, "visibility", "", "repo visibility: private | public")
	installCmd.Flags().StringVar(&docsTarget, "docs-target", "", "marker target: claude_md | agents_md | both | skip")
	installCmd.Flags().StringVar(&oracleBackend, "oracle-backend", "", "oracle backend: host-delegated | native | cli | api | mcp")
	installCmd.Flags().StringVar(&scanDeep, "scan-deep", "", "enable deep scan: true | false")
	ahoyCmd.AddCommand(installCmd)

	// uninstall
	var uninstallBinDir string
	uninstallCmd := &cobra.Command{
		Use:   "uninstall",
		Short: "Remove the marker block, abcd's owned PATH copy, and its provenance record (leaves .abcd/ intact)",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			cwd, err := os.Getwd()
			if err != nil {
				return err
			}
			receipt, err := ahoy.Uninstall(cwd, uninstallBinDir)
			if err != nil {
				return err
			}
			return render(cmd.OutOrStdout(), *asJSON, receipt, func(w io.Writer) {
				fmt.Fprintf(w, "abcd ahoy uninstall\n")
				fmt.Fprintf(w, "  marker removed: %v\n", receipt.Marker.Removed)
				fmt.Fprintf(w, "  symlink: %s\n", symlinkNote(receipt))
			})
		},
	}
	uninstallCmd.Flags().StringVar(&uninstallBinDir, "bin-dir", "", "directory holding the PATH entry to remove; needed only when it was installed with --bin-dir into a directory that is not on PATH")
	ahoyCmd.AddCommand(uninstallCmd)

	// doctor
	ahoyCmd.AddCommand(&cobra.Command{
		Use:   "doctor",
		Short: "Report every gap read-only, including user-scope state (never mutates)",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			cwd, err := os.Getwd()
			if err != nil {
				return err
			}
			report, err := ahoy.Doctor(cwd)
			if err != nil {
				return err
			}
			return render(cmd.OutOrStdout(), *asJSON, report, func(w io.Writer) {
				fmt.Fprintf(w, "abcd ahoy doctor — %s\n", report.Detection.FolderKind)
				fmt.Fprintf(w, "  detection gaps: %d\n", len(report.Detection.Gaps))
				fmt.Fprintf(w, "  audit gaps:     %d\n", len(report.AuditGaps))
			})
		},
	})

	// dry-run
	ahoyCmd.AddCommand(&cobra.Command{
		Use:   "dry-run",
		Short: "Render the detection-result JSON envelope; never mutates",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			cwd, err := os.Getwd()
			if err != nil {
				return err
			}
			res, err := ahoy.DryRun(cwd)
			if err != nil {
				return err
			}
			// dry-run always emits the canonical JSON envelope (spc-16 T1).
			enc := json.NewEncoder(cmd.OutOrStdout())
			enc.SetIndent("", "  ")
			return enc.Encode(res)
		},
	})

	// identity-check — the iss-62 gate's canonical, testable entrypoint. Exits
	// non-zero when the commit identity diverges from the committed pin, so a
	// pre-commit hook (or CI) can fail closed. A match, or an un-pinned repo,
	// exits zero.
	ahoyCmd.AddCommand(&cobra.Command{
		Use:   "identity-check",
		Short: "Exit non-zero if the git commit identity does not match .abcd/config/identity.json",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			cwd, err := os.Getwd()
			if err != nil {
				return err
			}
			res, err := identity.Check(cwd)
			if err != nil {
				return err
			}
			if res.Blocks() {
				return fmt.Errorf("%s\n  fix: git config user.name %q && git config user.email %q",
					res.Reason, res.Pin.Name, res.Pin.Email)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "identity ok (%s)\n", res.Status)
			return nil
		},
	})

	return ahoyCmd
}

// installOptionsFromFlags validates the install flags and builds InstallOptions.
// Only explicitly-set value flags become overrides; unset values fall through to
// the prompter (interactive) or its default (non-interactive).
func installOptionsFromFlags(cmd *cobra.Command, yes, adopt, refuseAdopt, dev, allowStale bool, binDir, visibility, docsTarget, oracleBackend, scanDeep string) (ahoy.InstallOptions, error) {
	opts := ahoy.InstallOptions{Yes: yes, Dev: dev, BinDir: binDir, AllowStaleBinary: allowStale}
	if adopt && refuseAdopt {
		return opts, fmt.Errorf("abcd ahoy install: --adopt and --refuse-adopt are mutually exclusive")
	}
	switch {
	case adopt:
		v := true
		opts.Adopt = &v
	case refuseAdopt:
		v := false
		opts.Adopt = &v
	}
	overrides := map[string]string{}
	set := func(key, val string, allowed []string) error {
		if !cmd.Flags().Changed(flagName(key)) {
			return nil
		}
		if len(allowed) > 0 && !contains(allowed, val) {
			return fmt.Errorf("abcd ahoy install: --%s must be one of %s", flagName(key), strings.Join(allowed, " | "))
		}
		overrides[key] = val
		return nil
	}
	if err := set("visibility", visibility, []string{"private", "public"}); err != nil {
		return opts, err
	}
	if err := set("docs_target", docsTarget, []string{"claude_md", "agents_md", "both", "skip"}); err != nil {
		return opts, err
	}
	if err := set("oracle_backend", oracleBackend, []string{"host-delegated", "native", "cli", "api", "mcp"}); err != nil {
		return opts, err
	}
	if err := set("scan_deep", scanDeep, []string{"true", "false"}); err != nil {
		return opts, err
	}
	if len(overrides) > 0 {
		opts.ValueOverrides = overrides
	}
	return opts, nil
}

// flagName maps an override key to its CLI flag name (underscore -> dash).
func flagName(key string) string { return strings.ReplaceAll(key, "_", "-") }

func contains(set []string, v string) bool {
	for _, s := range set {
		if s == v {
			return true
		}
	}
	return false
}

func symlinkNote(r ahoy.UninstallReceipt) string {
	if r.Symlink.Removed {
		return "removed " + r.Symlink.Target
	}
	return r.Symlink.Note
}

// newPrompter returns the stdin-reading prompter. On a terminal it is the
// interactive path, unchanged. When stdin is NOT a terminal the same prompter
// reads the piped answers (iss-167): `yes | abcd ahoy install` is what a host
// agent reaches for, and a TTY-only prompt turned every such answer into a
// decline, so the interactive path could not be driven at all. One answer
// answers one question, and the questions come in a fixed order
// (ahoy.categoryPromptOrder), so a piped stream lines up with them.
//
// The safe default survives: answers that run out — an empty pipe, a closed
// stdin, /dev/null — read as EOF, and EOF declines every confirm and takes the
// default for every prompt, exactly as the refusing prompter did.
func newPrompter(cmd *cobra.Command) ahoy.Prompter {
	in := cmd.InOrStdin()
	if in == nil {
		return ahoy.RefusingPrompter{}
	}
	p := &stdinPrompter{r: bufio.NewReader(in), w: cmd.ErrOrStderr()}
	if f, ok := in.(*os.File); ok {
		if fi, err := f.Stat(); err == nil && fi.Mode()&os.ModeCharDevice != 0 {
			p.tty = true
		}
	}
	return p
}

// stdinPrompter is the Prompter: it reads answers from stdin, whether a human
// types them or a caller pipes them in.
type stdinPrompter struct {
	r *bufio.Reader
	w io.Writer
	// tty records that a human is answering: the terminal echoes their own
	// typing, so the prompter must not echo it a second time. Off a terminal
	// nothing echoes, so the prompter writes the answer it read — a piped run
	// leaves a transcript of what was asked and what it was answered, instead
	// of a column of unanswered-looking questions.
	tty bool
}

// echo reports the answer read off a non-terminal stdin. The bytes come from
// the caller, so they are sanitised before reaching the terminal.
func (p *stdinPrompter) echo(answer string) {
	if p.tty {
		return
	}
	if answer == "" {
		answer = "<no answer>"
	}
	fmt.Fprintf(p.w, "%s\n", termsafe.Sanitize(answer))
}

func (p *stdinPrompter) Confirm(question string) bool {
	fmt.Fprintf(p.w, "%s [y/N] ", question)
	line, _ := p.r.ReadString('\n')
	line = strings.ToLower(strings.TrimSpace(line))
	p.echo(line)
	return line == "y" || line == "yes"
}

func (p *stdinPrompter) Prompt(key string, choices []string, def string) string {
	fmt.Fprintf(p.w, "%s (%s) [%s]: ", key, strings.Join(choices, "/"), def)
	line, _ := p.r.ReadString('\n')
	line = strings.TrimSpace(line)
	p.echo(line)
	if line == "" {
		return def
	}
	return line
}

// newCaptureCommand builds the `capture` sub-tree — the write side of the issue
// ledger. Bare `capture` renders read-only status; a free-text positional
// appends an issue; list/resolve/wontfix/promote are thin consumers of capture
// core.
func newCaptureCommand(asJSON *bool) *cobra.Command {
	var severity, category, source, slug, foundDuring, foundAt, blockedBy string

	captureCmd := &cobra.Command{
		Use:   "capture [text]",
		Short: "Capture issues to the ledger; bare invocation is read-only status",
		Args:  cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cwd, err := os.Getwd()
			if err != nil {
				return err
			}
			// Bare invocation: read-only status render (never mutates).
			if len(args) == 0 {
				st, err := capture.Status(capture.StatusRequest{RepoRoot: cwd})
				if err != nil {
					return err
				}
				return render(cmd.OutOrStdout(), *asJSON, st, func(w io.Writer) {
					fmt.Fprintf(w, "abcd capture — open %d · resolved %d · wontfix %d\n",
						st.OpenCount, st.ResolvedCount, st.WontfixCount)
					if len(st.RecentOpen) > 0 {
						fmt.Fprintf(w, "recent open:\n")
						for _, iss := range st.RecentOpen {
							fmt.Fprintf(w, "  %s  %s  %s%s\n", iss.ID, iss.Severity, iss.Slug, blockedNote(iss))
						}
					}
					fmt.Fprint(w, ledgerDecisionRule)
					fmt.Fprint(w, ideateRoutingRule)
				})
			}
			// Guard: a mistyped subcommand (e.g. `capture resovle iss-1 …`)
			// must not be swallowed as free text and filed. When args[0] is a
			// near-miss to a real subverb and the shape looks like a subcommand
			// call — a lone token, or a token followed by an issue id — refuse
			// with a did-you-mean and write nothing (unrecognized-input-never-
			// writes, iss-29). Genuine prose still files.
			if sug, ok := suspectedTypoedSubcommand(cmd, args); ok {
				return &exitError{Code: 2, Msg: fmt.Sprintf(
					"unknown capture subcommand %q; did you mean %q? (nothing captured — reword the text if you meant to file it)",
					args[0], sug)}
			}
			// Fast path: append a structured issue from the free-form text.
			text := strings.Join(args, " ")
			sl := slug
			if sl == "" {
				sl = deriveSlug(text)
			}
			blocked, err := parseBlockedBy(blockedBy)
			if err != nil {
				return err
			}
			req := capture.CaptureRequest{
				RepoRoot:    cwd,
				Text:        text,
				Severity:    capture.Severity(orDefault(severity, "minor")),
				Category:    capture.Category(orDefault(category, "observation")),
				Source:      capture.Source(orDefault(source, "user-observation")),
				Slug:        sl,
				FoundDuring: orDefault(foundDuring, "manual-capture"),
				FoundAt:     foundAt,
				BlockedBy:   blocked,
			}
			res, err := capture.Capture(req)
			if err != nil {
				return err
			}
			return render(cmd.OutOrStdout(), *asJSON, res, func(w io.Writer) {
				fmt.Fprintf(w, "captured %s (%s) — %s\n", res.ID, res.Status, res.Path)
			})
		},
	}
	captureCmd.Flags().StringVar(&severity, "severity", "", "severity: nitpick | minor | major | critical (default minor)")
	captureCmd.Flags().StringVar(&category, "category", "", "issue category (default observation)")
	captureCmd.Flags().StringVar(&source, "source", "", "surfacing channel (default user-observation)")
	captureCmd.Flags().StringVar(&slug, "slug", "", "override the slug derived from the text")
	captureCmd.Flags().StringVar(&foundDuring, "found-during", "", "session/command context (default manual-capture)")
	captureCmd.Flags().StringVar(&foundAt, "found-at", "", "optional repo-relative path or conceptual location")
	captureCmd.Flags().StringVar(&blockedBy, "blocked-by", "", "comma-separated iss-ids this issue is blocked by")

	// list — the earned SD001 exception: a filter flag is REQUIRED.
	var lsOpen, lsResolved, lsWontfix, lsAll bool
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List issues by state (one of --open/--resolved/--wontfix/--all required)",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			cwd, err := os.Getwd()
			if err != nil {
				return err
			}
			state, err := listState(lsOpen, lsResolved, lsWontfix, lsAll)
			if err != nil {
				return err
			}
			res, err := capture.List(capture.ListRequest{RepoRoot: cwd, State: state})
			if err != nil {
				return err
			}
			return render(cmd.OutOrStdout(), *asJSON, res, func(w io.Writer) {
				for _, iss := range res.Issues {
					fmt.Fprintf(w, "%s  %s  %s  %s%s\n", iss.ID, iss.Status, iss.Severity, iss.Slug, blockedNote(iss))
				}
				for _, sk := range res.Skipped {
					// Path and Error echo a malformed issue file's own name and content
					// (err.Error() carries offending bytes), so sanitise before the terminal.
					fmt.Fprintf(w, "  skipped %s: %s\n", termsafe.Sanitize(sk.Path), termsafe.Sanitize(sk.Error))
				}
			})
		},
	}
	listCmd.Flags().BoolVar(&lsOpen, "open", false, "issues currently in open/")
	listCmd.Flags().BoolVar(&lsResolved, "resolved", false, "issues currently in resolved/")
	listCmd.Flags().BoolVar(&lsWontfix, "wontfix", false, "issues currently in wontfix/")
	listCmd.Flags().BoolVar(&lsAll, "all", false, "issues across all three states")
	captureCmd.AddCommand(listCmd)

	// resolve — open -> resolved with a note, a required product impact, and
	// optional resolved_by provenance (spc-25): the intent, spec, or commit
	// that fixed it.
	var resolveImpact, resolveByIntent, resolveBySpec, resolveByCommit string
	resolveCmd := &cobra.Command{
		Use:   "resolve <iss-N> <note> --impact <additive|breaking|fix|internal> [--intent itd-N] [--spec spc-N] [--commit sha]",
		Short: "Mark an open issue resolved (open/ -> resolved/), optionally naming what fixed it",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cwd, err := os.Getwd()
			if err != nil {
				return err
			}
			res, err := capture.Resolve(capture.ResolveRequest{
				RepoRoot: cwd, ID: args[0], Resolution: args[1], Impact: resolveImpact,
				ByIntent: resolveByIntent, BySpec: resolveBySpec, ByCommit: resolveByCommit,
			})
			if err != nil {
				return err
			}
			return render(cmd.OutOrStdout(), *asJSON, res, func(w io.Writer) {
				fmt.Fprintf(w, "%s  %s -> %s — %s%s\n", res.ID, res.FromStatus, res.ToStatus, res.Path, resolvedByNote(res.ResolvedBy))
			})
		},
	}
	// resolved/ is gated by issue_impact_valid: the record carries the product
	// judgement the version derivation reads, and there is no default. The flag is
	// mandatory in effect — the core (capture.Resolve -> changelog.ParseImpact)
	// refuses an empty impact — but it is not marked cobra-required, to keep the
	// tree's no-required-flags invariant (TestLiveTreeMarksNoFlagRequired): the
	// requirement is enforced semantically in the core, not by a usage annotation.
	resolveCmd.Flags().StringVar(&resolveImpact, "impact", "", "product impact: additive|breaking|fix|internal (required)")
	resolveCmd.Flags().StringVar(&resolveByIntent, "intent", "", "resolved_by provenance: the itd-N that fixed it (must exist)")
	resolveCmd.Flags().StringVar(&resolveBySpec, "spec", "", "resolved_by provenance: the spc-N that fixed it (must exist)")
	resolveCmd.Flags().StringVar(&resolveByCommit, "commit", "", "resolved_by provenance: the fixing commit sha (7-64 hex chars, shape-checked only)")
	captureCmd.AddCommand(resolveCmd)

	// promote — graduate an issue into an intent draft (spc-24, step 2 of the
	// record walk). Default mode mints a draft and stamps the issue's
	// promoted_to in one invocation; --intent is the stamp-only repair/link
	// mode. The issue keeps its status folder — promotion is not resolution.
	var promoteIntent string
	promoteCmd := &cobra.Command{
		Use:   "promote <iss-N>",
		Short: "Graduate an issue into an intent draft (mints + stamps promoted_to)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cwd, err := os.Getwd()
			if err != nil {
				return err
			}
			res, err := capture.Promote(capture.PromoteRequest{RepoRoot: cwd, ID: args[0], LinkIntent: promoteIntent})
			if err != nil {
				return err
			}
			emitMintWarning(cmd, res.MintWarning)
			return render(cmd.OutOrStdout(), *asJSON, res, func(w io.Writer) {
				verb := "minted"
				if res.Linked {
					verb = "linked"
				}
				// Paths echo ledger/corpus filenames (attacker-shapeable in a
				// hostile clone), so sanitise before the terminal.
				fmt.Fprintf(w, "%s (%s, %s) promoted — %s %s — %s\n",
					res.IssueID, res.IssueStatus, termsafe.Sanitize(res.IssuePath),
					verb, res.IntentID, termsafe.Sanitize(res.IntentPath))
			})
		},
	}
	promoteCmd.Flags().StringVar(&promoteIntent, "intent", "", "stamp-only mode: link this existing itd-N instead of minting a draft")
	captureCmd.AddCommand(promoteCmd)

	// wontfix — open -> wontfix with a reason.
	captureCmd.AddCommand(&cobra.Command{
		Use:   "wontfix <iss-N> <reason>",
		Short: "Record an explicit non-action decision (open/ -> wontfix/)",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cwd, err := os.Getwd()
			if err != nil {
				return err
			}
			res, err := capture.Wontfix(capture.WontfixRequest{RepoRoot: cwd, ID: args[0], Reason: args[1]})
			if err != nil {
				return err
			}
			return render(cmd.OutOrStdout(), *asJSON, res, func(w io.Writer) {
				fmt.Fprintf(w, "%s  %s -> %s — %s\n", res.ID, res.FromStatus, res.ToStatus, res.Path)
			})
		},
	})

	return captureCmd
}

// resolvedByNote renders the stamped provenance members for the resolve text
// surface ("" when none were written). Members are regex-validated ids/shas,
// so no sanitisation is needed.
func resolvedByNote(rb *capture.ResolvedBy) string {
	if rb == nil {
		return ""
	}
	var parts []string
	if rb.Intent != "" {
		parts = append(parts, "intent="+rb.Intent)
	}
	if rb.Spec != "" {
		parts = append(parts, "spec="+rb.Spec)
	}
	if rb.Commit != "" {
		parts = append(parts, "commit="+rb.Commit)
	}
	return " — resolved_by " + strings.Join(parts, " ")
}

// listState maps the mutually-exclusive filter flags to a capture.State, or
// returns the exit-2 "choose a filter" usage error the brief mandates for the
// unfiltered `abcd capture list` form (04-surfaces/06 § 1).
func listState(open, resolved, wontfix, all bool) (capture.State, error) {
	var chosen capture.State
	n := 0
	if open {
		chosen, n = capture.StateOpen, n+1
	}
	if resolved {
		chosen, n = capture.StateResolved, n+1
	}
	if wontfix {
		chosen, n = capture.StateWontfix, n+1
	}
	if all {
		chosen, n = capture.StateAll, n+1
	}
	if n == 0 {
		return "", &exitError{Code: 2, Msg: "capture list: choose a filter: --open / --resolved / --wontfix / --all"}
	}
	if n > 1 {
		return "", &exitError{Code: 2, Msg: "capture list: the filter flags are mutually exclusive"}
	}
	return chosen, nil
}

// deriveSlug ports scripts/abcd/_slug._normalize_core: lowercase, collapse every
// non-[a-z0-9] run to a single hyphen, trim, then truncate to 60 chars.
func deriveSlug(text string) string {
	lowered := strings.ToLower(text)
	collapsed := strings.Trim(slugNonAlnumRe.ReplaceAllString(lowered, "-"), "-")
	if len(collapsed) > 60 {
		collapsed = strings.Trim(collapsed[:60], "-")
	}
	return collapsed
}

var slugNonAlnumRe = regexp.MustCompile(`[^a-z0-9]+`)

// issIDRe validates a --blocked-by token at the CLI boundary (mirrors the core
// ^iss-[0-9]+$ schema constraint).
var issIDRe = regexp.MustCompile(`^iss-[0-9]+$`)

// recordIDRe matches any abcd record id (issue, intent, or spec). It is used
// only by suspectedTypoedSubcommand's shape check — distinct from issIDRe, which
// validates a real iss-only --blocked-by token — so the typo guard recognises a
// subcommand call in either verb family (capture's iss-N ids, intent's itd/spc
// ids) without loosening iss-id validation elsewhere.
var recordIDRe = regexp.MustCompile(`^(iss|itd|spc)-[0-9]+$`)

// retiredSubverbs maps a parent command to sub-verb spellings that were
// renamed in a clean break (no alias): an invocation shaped like a subcommand
// call using a retired spelling is refused with the successor named, so it can
// never be swallowed as free text and silently filed (the same
// unrecognized-input-never-writes contract as the typo guard, iss-29).
var retiredSubverbs = map[string]map[string]string{
	"intent": {"review": "audit"}, // spc-28 (adr-40)
}

// suspectedTypoedSubcommand reports the nearest real subverb when args[0] is a
// near-miss for one (edit distance 1–2) — or a retired spelling of one — and
// the invocation shape resembles a subcommand call rather than free-text
// prose: a lone token, or a token followed by a record id. It is deliberately
// high-precision so it never refuses a legitimate free-text create whose first
// word merely resembles a verb — those carry no trailing record id and are
// multi-word.
func suspectedTypoedSubcommand(parent *cobra.Command, args []string) (string, bool) {
	if len(args) == 0 {
		return "", false
	}
	if successor, ok := retiredSubverbs[parent.Name()][args[0]]; ok {
		// A retired spelling is refused in every shape a subcommand call takes:
		// a lone token, a token before a record id, or the retired verb followed
		// by one of the SUCCESSOR's own registered sub-verbs (`review ingest`
		// must refuse just as `review itd-N` does — the pre-rename two-word
		// invocation may never be swallowed as free text and filed).
		if len(args) == 1 || recordIDRe.MatchString(args[1]) || isSubverbOf(parent, successor, args[1]) {
			return successor, true
		}
	}
	shapedLikeSubcommand := len(args) == 1 || recordIDRe.MatchString(args[1])
	if !shapedLikeSubcommand {
		return "", false
	}
	best, bestDist := "", 3 // accept edit distances 1 and 2
	for _, c := range parent.Commands() {
		name := c.Name()
		if c.Hidden || name == "help" || name == "completion" {
			continue
		}
		if d := levenshtein(args[0], name); d > 0 && d < bestDist {
			best, bestDist = name, d
		}
	}
	return best, best != ""
}

// isSubverbOf reports whether token names a registered sub-command of parent's
// child command named successor (e.g. is "ingest" a sub-verb of "audit").
func isSubverbOf(parent *cobra.Command, successor, token string) bool {
	for _, c := range parent.Commands() {
		if c.Name() != successor {
			continue
		}
		for _, sub := range c.Commands() {
			if sub.Name() == token {
				return true
			}
		}
	}
	return false
}

// levenshtein is the classic edit distance (insert/delete/substitute each cost
// 1). Inputs are subcommand-name sized, so the simple O(n·m) two-row form is
// more than fast enough.
func levenshtein(a, b string) int {
	ra, rb := []rune(a), []rune(b)
	prev := make([]int, len(rb)+1)
	for j := range prev {
		prev[j] = j
	}
	for i := 1; i <= len(ra); i++ {
		cur := make([]int, len(rb)+1)
		cur[0] = i
		for j := 1; j <= len(rb); j++ {
			cost := 1
			if ra[i-1] == rb[j-1] {
				cost = 0
			}
			cur[j] = min(prev[j]+1, cur[j-1]+1, prev[j-1]+cost)
		}
		prev = cur
	}
	return prev[len(rb)]
}

// parseBlockedBy splits the comma-separated --blocked-by value into iss-ids,
// dropping blanks and rejecting any token that is not ^iss-[0-9]+$. An empty
// input yields a nil slice (the field is omitted).
func parseBlockedBy(raw string) ([]string, error) {
	if strings.TrimSpace(raw) == "" {
		return nil, nil
	}
	var ids []string
	for _, tok := range strings.Split(raw, ",") {
		tok = strings.TrimSpace(tok)
		if tok == "" {
			continue
		}
		if !issIDRe.MatchString(tok) {
			return nil, fmt.Errorf("capture: --blocked-by token %q must match iss-N", tok)
		}
		ids = append(ids, tok)
	}
	return ids, nil
}

// blockedNote renders the derived-priority annotation for a row: when the issue
// has blocked_by targets still open, " [blocked-by iss-1,iss-2]"; otherwise "".
func blockedNote(iss capture.Issue) string {
	if len(iss.BlockedByOpen) == 0 {
		return ""
	}
	return " [blocked-by " + strings.Join(iss.BlockedByOpen, ",") + "]"
}

func orDefault(v, def string) string {
	if v == "" {
		return def
	}
	return v
}

// newMemoryCommand builds the `memory` sub-tree over internal/core/memory. Bare
// `memory` renders read-only store status; ingest/ask/lint are the mutating and
// diagnostic verbs (04-surfaces/07). The distiller (ingest) and synthesizer
// (ask) are host-delegated seams: the .5 skill emits validated DistilledPage
// JSON, which this surface feeds through --pages-json / --page-json.
func newMemoryCommand(asJSON *bool) *cobra.Command {
	memoryCmd := &cobra.Command{
		Use:   "memory",
		Short: "Curated knowledge substrate; bare invocation is read-only status",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			cwd, err := os.Getwd()
			if err != nil {
				return err
			}
			st, err := memory.Bare(cwd)
			if err != nil {
				return err
			}
			return render(cmd.OutOrStdout(), *asJSON, st, func(w io.Writer) {
				fmt.Fprintf(w, "abcd memory — %d page(s)", st.Pages)
				if !st.StorePresent {
					fmt.Fprintf(w, " (store not present)")
				}
				fmt.Fprintln(w)
				for _, c := range st.ByClass {
					// Class labels come from page frontmatter; contradiction/headroom lines
					// are read verbatim from repo files — all untrusted terminal output.
					fmt.Fprintf(w, "  %s: %d\n", termsafe.Sanitize(c.Class), c.Count)
				}
				if st.LastIngest != "" {
					fmt.Fprintf(w, "  last ingest: %s\n", termsafe.Sanitize(st.LastIngest))
				}
				for _, line := range st.Contradictions {
					fmt.Fprintf(w, "  contradiction: %s\n", termsafe.Sanitize(line))
				}
				for _, line := range st.Headroom {
					fmt.Fprintf(w, "  %s\n", termsafe.Sanitize(line))
				}
			})
		},
	}

	// ingest <path-or-url> [--keep-original] [--pages-json <file|->]
	var pagesJSON string
	var keepOriginalFlag bool
	ingestCmd := &cobra.Command{
		Use:   "ingest <path-or-url>",
		Short: "Distil an external source into cited memory pages",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cwd, err := os.Getwd()
			if err != nil {
				return err
			}
			res, err := memory.Ingest(memory.IngestRequest{
				RepoRoot:     cwd,
				Source:       args[0],
				KeepOriginal: keepOriginalFlag,
				Distiller:    pagesJSONDistiller(cmd, pagesJSON),
			})
			if err != nil {
				return err
			}
			if err := render(cmd.OutOrStdout(), *asJSON, res, func(w io.Writer) {
				fmt.Fprintf(w, "abcd memory ingest — %s\n", res.Status)
				fmt.Fprintf(w, "  content hash: %s\n", res.ContentHash)
				fmt.Fprintf(w, "  licence:      %s\n", res.Licence)
				if len(res.Pages) > 0 {
					fmt.Fprintf(w, "  pages:        %s\n", strings.Join(res.Pages, ", "))
				}
				if res.KeptOriginal != "" {
					fmt.Fprintf(w, "  kept original: %s\n", res.KeptOriginal)
				}
				if res.KeepOriginalError != "" {
					fmt.Fprintf(w, "  warning: --keep-original failed (the source was still ingested): %s\n", res.KeepOriginalError)
				}
			}); err != nil {
				return err
			}
			// The ingest succeeded but an explicitly-requested --keep-original
			// copy did not: signal it with a non-zero exit while leaving the
			// rendered result (which reports what was durably written) intact.
			if res.KeepOriginalError != "" {
				return &exitError{Code: 1}
			}
			return nil
		},
	}
	ingestCmd.Flags().BoolVar(&keepOriginalFlag, "keep-original", false, "store the original at .abcd/memory/sources/<sha256>.<ext>")
	ingestCmd.Flags().StringVar(&pagesJSON, "pages-json", "", "DistilledPage JSON array (file path, or - for stdin)")
	memoryCmd.AddCommand(ingestCmd)

	// ask <question> [--top-n N] [--file-back] [--page-json <file|->]
	var topN int
	var fileBack bool
	var pageJSON string
	askCmd := &cobra.Command{
		Use:   "ask <question>",
		Short: "Query memory and synthesise a cited answer",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cwd, err := os.Getwd()
			if err != nil {
				return err
			}
			req := memory.AskRequest{RepoRoot: cwd, Question: strings.Join(args, " "), TopN: topN}
			if fileBack {
				page, err := readPageJSON(cmd, pageJSON)
				if err != nil {
					return err
				}
				req.FileBackPage = page
				req.DecideFileBack = func(memory.DistilledPage) bool { return true }
			}
			res, err := memory.Ask(req)
			if err != nil {
				return err
			}
			return render(cmd.OutOrStdout(), *asJSON, res, func(w io.Writer) {
				fmt.Fprintln(w, res.Answer)
				if res.FileBack != nil {
					fmt.Fprintf(w, "\nfiled back (%s): %s\n", res.FileBack.Status, strings.Join(res.FileBack.Pages, ", "))
				}
			})
		},
	}
	askCmd.Flags().IntVar(&topN, "top-n", 0, "retrieval depth (0 uses the pinned default)")
	askCmd.Flags().BoolVar(&fileBack, "file-back", false, "file the synthesised answer back as a new memory page")
	askCmd.Flags().StringVar(&pageJSON, "page-json", "", "the answer page dict as JSON (file path, or - for stdin)")
	memoryCmd.AddCommand(askCmd)

	// lint — full-store curator health-check; blockers exit nonzero.
	memoryCmd.AddCommand(&cobra.Command{
		Use:   "lint",
		Short: "Curator health-check over the whole memory store",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			cwd, err := os.Getwd()
			if err != nil {
				return err
			}
			res, err := memory.Lint(memory.LintRequest{RepoRoot: cwd})
			if err != nil {
				return err
			}
			if err := render(cmd.OutOrStdout(), *asJSON, res, func(w io.Writer) {
				fmt.Fprintf(w, "abcd memory lint — %d blocker(s), %d warning(s), %d info(s)\n",
					res.Summary.Blockers, res.Summary.Warnings, res.Summary.Infos)
				for _, f := range res.Findings {
					fmt.Fprintf(w, "  %s [%s] %s:%d %s\n", f.Code, f.Severity, termsafe.Sanitize(f.File), f.Line, termsafe.Sanitize(f.Message))
				}
				fmt.Fprintf(w, "  report: %s\n", res.ReportDir)
			}); err != nil {
				return err
			}
			// Propagate the curator exit contract: blockers -> nonzero.
			if res.ExitCode != 0 {
				return &exitError{Code: res.ExitCode}
			}
			return nil
		},
	})

	return memoryCmd
}

// pagesJSONDistiller is the ingest transport seam: it lazily reads the
// DistilledPage JSON array from --pages-json (a file, or - for stdin) only when
// distillation is actually needed. A registry-only hit never invokes it, so an
// already-known source re-ingests with no payload.
func pagesJSONDistiller(cmd *cobra.Command, pagesJSON string) memory.Distiller {
	return func(_ string, _ map[string]any) ([]map[string]any, error) {
		if pagesJSON == "" {
			return nil, fmt.Errorf("no distiller output supplied: pass --pages-json <file|-> with the DistilledPage JSON array")
		}
		raw, err := readSource(cmd, pagesJSON)
		if err != nil {
			return nil, fmt.Errorf("cannot read --pages-json: %w", err)
		}
		var arr []map[string]any
		if err := json.Unmarshal(raw, &arr); err != nil {
			return nil, fmt.Errorf("--pages-json must be a JSON array of page dicts: %w", err)
		}
		return arr, nil
	}
}

// readPageJSON reads ONE DistilledPage dict for ask file-back from --page-json.
func readPageJSON(cmd *cobra.Command, pageJSON string) (map[string]any, error) {
	if pageJSON == "" {
		return nil, fmt.Errorf("--file-back requires --page-json <file|-> with the answer page dict")
	}
	raw, err := readSource(cmd, pageJSON)
	if err != nil {
		return nil, fmt.Errorf("cannot read --page-json: %w", err)
	}
	var obj map[string]any
	if err := json.Unmarshal(raw, &obj); err != nil {
		return nil, fmt.Errorf("--page-json must be one JSON object (a DistilledPage dict): %w", err)
	}
	return obj, nil
}

// maxOperandJSONBytes is the house cap (8 MiB, matching the registry/graveyard
// JSON caps) for an untrusted JSON operand read from a file path or stdin.
const maxOperandJSONBytes = 8 << 20

// readSource reads a JSON payload from a file path, or from stdin when spec is
// "-" (the streaming transport the .5 skill uses). The operand is untrusted
// content (host-produced pages, cross-machine artifacts), so both transports are
// bounded and the file path is read behind the trust guards: stdin is refused
// whole when over-cap (a cap+1 probe, never a severed prefix), and a file is
// read with fsutil.ReadGuarded (O_NOFOLLOW so a symlink operand is never
// followed, regular-file on the open fd, and the size cap — all in one call, no
// lstat→read TOCTOU). Refusing the over-cap prefix matters most on the history
// capture path, where a truncated transcript would be stored under a sha256
// idempotency key computed over the prefix (spc-4's refuse-whole invariant).
func readSource(cmd *cobra.Command, spec string) ([]byte, error) {
	if spec == "-" {
		return readCappedStdin(cmd, maxOperandJSONBytes)
	}
	return readGuardedOperand(spec, maxOperandJSONBytes)
}

// newHistoryCommand builds the `history` sub-tree over internal/core/history —
// the native session-transcript store (adr-29). `list`/`show` read; `capture`
// is the redacting write path. The per-repo store is keyed on the root-commit
// SHA resolved from cwd.
func newHistoryCommand(asJSON *bool) *cobra.Command {
	historyCmd := &cobra.Command{
		Use:   "history",
		Short: "Manage the native session-transcript store",
		Args:  cobra.NoArgs,
		RunE:  helpRunE,
	}

	// capture — the redacting write path: read a raw transcript from a file
	// argument (or stdin with "-"/no arg), sanitise it through the scanner
	// (two-stage, fail-closed), and store the record. This is the ONLY path that
	// writes to the store; list/show never mutate.
	var session, kind string
	captureCmd := &cobra.Command{
		Use:   "capture [<transcript-file>|-]",
		Short: "Redact and store a raw session transcript (reads a file or stdin)",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			rootSHA, err := repoRootSHA()
			if err != nil {
				return err
			}
			src := "-"
			if len(args) == 1 {
				src = args[0]
			}
			raw, err := readSource(cmd, src)
			if err != nil {
				return fmt.Errorf("history capture: cannot read transcript: %w", err)
			}
			sess := session
			if sess == "" && src != "-" {
				// Derive a session id from the file basename (sans extension).
				base := filepath.Base(src)
				sess = strings.TrimSuffix(base, filepath.Ext(base))
			}
			if sess == "" {
				return fmt.Errorf("history capture: --session <id> is required when reading from stdin")
			}
			cwd, err := os.Getwd()
			if err != nil {
				return err
			}
			res, err := history.Capture(captureRoot(cwd), rootSHA, sess, raw, orDefault(kind, "native"))
			if err != nil {
				return err
			}
			return render(cmd.OutOrStdout(), *asJSON, res, func(w io.Writer) {
				if !res.Wrote {
					fmt.Fprintf(w, "abcd history capture — %s already stored (no-op); redacted secrets=%d home=%d\n",
						res.Record.SessionID, res.Record.Secrets, res.Record.HomePaths)
					return
				}
				fmt.Fprintf(w, "abcd history capture — stored %s (%s)\n", res.Record.SessionID, res.Record.SourceKind)
				fmt.Fprintf(w, "  path:     %s\n", res.Record.Path)
				fmt.Fprintf(w, "  redacted: secrets=%d home=%d\n", res.Record.Secrets, res.Record.HomePaths)
			})
		},
	}
	captureCmd.Flags().StringVar(&session, "session", "", "session id for the record (default: transcript filename; required for stdin)")
	captureCmd.Flags().StringVar(&kind, "kind", "", "source kind: native | specstory-import (default native)")
	historyCmd.AddCommand(captureCmd)

	// list — records newest-first for this repo.
	historyCmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List stored transcripts for this repo, newest first",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			rootSHA, err := repoRootSHA()
			if err != nil {
				return err
			}
			records, err := history.List(rootSHA)
			if err != nil {
				return err
			}
			// An empty store is an empty LIST in JSON, not bare `null`: the
			// command doc promises "an empty list means no transcripts", and a
			// consumer that iterates the value should get [], as every other
			// --json verb's collection does.
			if records == nil {
				records = []history.Record{}
			}
			return render(cmd.OutOrStdout(), *asJSON, records, func(w io.Writer) {
				if len(records) == 0 {
					fmt.Fprintln(w, "abcd history — no transcripts stored for this repo")
					return
				}
				for _, r := range records {
					fmt.Fprintf(w, "%s  %s  %s  redacted secrets=%d home=%d\n",
						r.CapturedAt.Format("2006-01-02T15:04:05Z"), termsafe.Sanitize(r.SessionID), termsafe.Sanitize(r.SourceKind), r.Secrets, r.HomePaths)
				}
			})
		},
	})

	// show <session-id-or-filename> — metadata + redacted body of one record.
	historyCmd.AddCommand(&cobra.Command{
		Use:   "show <session-id-or-filename>",
		Short: "Show one stored transcript's metadata and redacted body",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			rootSHA, err := repoRootSHA()
			if err != nil {
				return err
			}
			rec, body, err := history.Read(rootSHA, args[0])
			if err != nil {
				return err
			}
			out := struct {
				history.Record
				Body string `json:"body"`
			}{Record: rec, Body: string(body)}
			return render(cmd.OutOrStdout(), *asJSON, out, func(w io.Writer) {
				// The stored transcript body is untrusted (it may have ingested
				// hostile fetched pages or target-repo files); capture redacts
				// only secrets/home paths, so neutralise terminal-control bytes
				// here before they reach the terminal. SanitizeBlock keeps the
				// transcript's line structure. The metadata fields are validated
				// at write time but not re-validated on the read path, so pass
				// them through too.
				fmt.Fprintf(w, "session:    %s\n", termsafe.Sanitize(rec.SessionID))
				fmt.Fprintf(w, "captured:   %s\n", rec.CapturedAt.Format("2006-01-02T15:04:05Z"))
				fmt.Fprintf(w, "source:     %s\n", termsafe.Sanitize(rec.SourceKind))
				fmt.Fprintf(w, "path:       %s\n", termsafe.Sanitize(rec.Path))
				fmt.Fprintf(w, "redacted:   secrets=%d home=%d\n", rec.Secrets, rec.HomePaths)
				fmt.Fprintln(w, "---")
				fmt.Fprint(w, termsafe.SanitizeBlock(string(body)))
			})
		},
	})

	return historyCmd
}

// repoRootSHA resolves the current repo's root-commit SHA (the history store
// key) via the ahoy detection pass. An empty SHA means cwd is not a git repo
// with commits, which the history verbs cannot key on.
// captureRoot resolves the git working-tree root containing cwd so history
// capture honours the per-repo redaction override (the scanner resolves it at
// <root>/.abcd/config/pii.json, without walking up). Without this, a capture run
// from a subdirectory hands the subdirectory to scanner.New, which finds no
// override there and silently redacts with defaults only (B12). It falls back to
// cwd when git cannot answer (not a repo, git absent) — the scanner then behaves
// exactly as before, so the fallback never regresses a non-git use.
func captureRoot(cwd string) string {
	if top, err := gitutil.Run(cwd, "rev-parse", "--show-toplevel"); err == nil && top != "" {
		return top
	}
	return cwd
}

// rulesRoot resolves the repo root the modular-rules loader must read: the
// nearest ancestor of cwd (cwd itself included) that holds a .abcd directory.
// rules.Load/LoadBackstop join ".abcd/rules.json" onto the path they are given,
// so handing them a subdirectory silently ignored the per-repo overrides AND the
// kill switch — a repo that had disabled a domain (or the whole loader) would
// still inject it whenever abcd ran from any nested directory. Falls back to the
// git working-tree root, then cwd, so a repo without a .abcd dir still resolves.
func rulesRoot(cwd string) string {
	dir := cwd
	for {
		if fi, err := os.Stat(filepath.Join(dir, ".abcd")); err == nil && fi.IsDir() {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return captureRoot(cwd)
}

func repoRootSHA() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	det, err := ahoy.Detect(cwd)
	if err != nil {
		return "", err
	}
	if det.RootSHA == "" {
		return "", fmt.Errorf("history: cannot resolve the repo's root-commit SHA (not a git repo with commits)")
	}
	return det.RootSHA, nil
}

// Execute runs the root command; main sets the process exit code on error.
// Run builds the command tree, executes it against args, and renders any error
// as a single diagnostic line — the one place that maps a command error to a
// process exit code, so main stays a thin shell. stdout/stderr are injected so
// the whole front door (including its error surface) is testable.
func Run(args []string, stdout, stderr io.Writer) int {
	root := NewRootCommand()
	root.SetArgs(args)
	root.SetOut(stdout)
	root.SetErr(stderr)

	err := root.Execute()
	if err == nil {
		return 0
	}
	// A command may request a specific exit code (usage errors, the memory-lint
	// curator contract). An empty message means it already rendered its output
	// and only the exit code should propagate.
	code := 1
	var coded interface{ ExitCode() int }
	if errors.As(err, &coded) {
		code = coded.ExitCode()
	}
	if msg := scrubPaths(err); msg != "" {
		// Honour --json for the error surface too: a caller that asked for
		// machine output must get a JSON envelope, never raw Go text (iss-29).
		if asJSON, _ := root.PersistentFlags().GetBool("json"); asJSON {
			enc := json.NewEncoder(stderr)
			enc.SetIndent("", "  ")
			_ = enc.Encode(errorEnvelope{Error: msg})
		} else {
			fmt.Fprintln(stderr, "abcd:", msg)
		}
	}
	return code
}

// errorEnvelope is the --json error shape: a single {"error": "..."} object so
// a machine caller can parse a failure the same way it parses a success.
type errorEnvelope struct {
	Error string `json:"error"`
}

// scrubPaths renders err for machine/stderr output with the DEVELOPER-IDENTITY
// portion of any local path removed. cli.Run routes every command error through
// the --json envelope and the stderr line, and an identity-bearing path reaches
// that surface three ways: an os.PathError/os.LinkError embeds one in Error();
// core fmt-formats one via %s (e.g. capture's ledger-path guards); a custom error
// type renders one (e.g. history's home-rooted StorePathError). All three are
// handled (iss-76 — the identity-scrub generalisation of the one branch iss-29
// fixed):
//
//   - the two roots that carry developer identity — the working directory and the
//     home directory — are redacted to "." and "~" wherever they appear, catching
//     fmt-formatted and custom-error-type paths a typed walk cannot see;
//   - any remaining absolute path embedded by os.PathError/os.LinkError (e.g. a
//     path argument outside both roots) is reduced to its base name.
//
// This is NOT a universal absolute-path scrub: a verb that echoes a user-supplied
// absolute path lying outside both roots (e.g. `memory ingest /tmp/x`) still
// surfaces it — that path carries no developer identity, and sanitising such
// verb-level echoes is tracked separately (iss-81). Scrubbing here rather than by
// regex is deliberate: this error surface also carries URLs (fetch failures) that
// an absolute-path regex would mangle.
//
// The root redaction itself is fsutil.RedactRoot: it moved out of this file when
// the ahoy install receipt had to redact the same two roots (iss-177), so what
// counts as a developer-identity root is stated once and read by both the error
// surface here and the receipt in core.
func scrubPaths(err error) string {
	msg := err.Error()
	if cwd, e := os.Getwd(); e == nil {
		msg = fsutil.RedactRoot(msg, cwd, ".")
	}
	if home, e := os.UserHomeDir(); e == nil {
		msg = fsutil.RedactRoot(msg, home, "~")
	}
	for _, p := range embeddedPaths(err) {
		if filepath.IsAbs(p) {
			msg = strings.ReplaceAll(msg, p, filepath.Base(p))
		}
	}
	return msg
}

// embeddedPaths collects the filesystem paths carried by os.PathError/os.LinkError
// anywhere in err's Unwrap chain, including errors.Join fan-out.
func embeddedPaths(err error) []string {
	var paths []string
	var walk func(error)
	walk = func(e error) {
		for e != nil {
			switch t := e.(type) {
			case *os.PathError:
				paths = append(paths, t.Path)
			case *os.LinkError:
				paths = append(paths, t.Old, t.New)
			}
			switch u := e.(type) {
			case interface{ Unwrap() error }:
				e = u.Unwrap()
			case interface{ Unwrap() []error }:
				for _, sub := range u.Unwrap() {
					walk(sub)
				}
				return
			default:
				return
			}
		}
	}
	walk(err)
	return paths
}

// render writes v as indented JSON when asJSON is set, otherwise delegates to
// the text renderer. Keeping this one helper is what makes every command's
// --json behaviour uniform.
func render(w io.Writer, asJSON bool, v any, text func(io.Writer)) error {
	if asJSON {
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		return enc.Encode(v)
	}
	text(w)
	return nil
}
