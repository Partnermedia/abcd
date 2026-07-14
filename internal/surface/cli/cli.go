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
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"syscall"

	"github.com/REPPL/abcd-cli/internal/core"
	"github.com/REPPL/abcd-cli/internal/core/ahoy"
	"github.com/REPPL/abcd-cli/internal/core/capture"
	"github.com/REPPL/abcd-cli/internal/core/history"
	"github.com/REPPL/abcd-cli/internal/core/identity"
	"github.com/REPPL/abcd-cli/internal/core/intent"
	"github.com/REPPL/abcd-cli/internal/core/launch"
	"github.com/REPPL/abcd-cli/internal/core/lint"
	"github.com/REPPL/abcd-cli/internal/core/memory"
	"github.com/REPPL/abcd-cli/internal/core/rules"
	"github.com/REPPL/abcd-cli/internal/core/spec"
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

// NewRootCommand builds the abcd command tree. Bare `abcd` renders a read-only
// status board (abcd's convention: bare invocation never mutates); subcommands
// carry the actions.
func NewRootCommand() *cobra.Command {
	var asJSON bool

	root := &cobra.Command{
		Use:           "abcd",
		Short:         "Agent-based configuration for development",
		SilenceUsage:  true,
		SilenceErrors: true,
		Args:          cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			cwd, err := os.Getwd()
			if err != nil {
				return err
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

	root.AddCommand(&cobra.Command{
		Use:   "version",
		Short: "Print abcd's version",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			v := core.NewVersion()
			return render(cmd.OutOrStdout(), asJSON, v, func(w io.Writer) {
				fmt.Fprintf(w, "%s %s\n", v.Name, v.Version)
			})
		},
	})

	root.AddCommand(newAhoyCommand(&asJSON))
	root.AddCommand(newAuditCommand(&asJSON))

	var launchDryRun bool
	launchCmd := &cobra.Command{
		Use:   "launch",
		Short: "Preview the public launch bundle and release gates (read-only)",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			cwd, err := os.Getwd()
			if err != nil {
				return err
			}
			if !launchDryRun {
				return fmt.Errorf("abcd launch: pass --dry-run to preview the bundle (publishing is not wired at this stage)")
			}
			rep, err := launch.DryRun(launch.DryRunRequest{RepoRoot: cwd})
			if err != nil {
				return err
			}
			return render(cmd.OutOrStdout(), asJSON, rep, func(w io.Writer) {
				fmt.Fprintf(w, "abcd launch (dry-run) — version %s\n", rep.Version)
				fmt.Fprintf(w, "  files bundled:  %d\n", len(rep.Bundle.Included))
				fmt.Fprintf(w, "  scan hardfails: %d\n", rep.Scan.HardFails)
				fmt.Fprintf(w, "  would publish:  %v\n", rep.WouldPublish)
				if len(rep.WouldRefuseOn) > 0 {
					fmt.Fprintf(w, "  would refuse on: %v\n", rep.WouldRefuseOn)
				}
			})
		},
	}
	launchCmd.Flags().BoolVar(&launchDryRun, "dry-run", false, "preview the launch bundle and gates without publishing")
	root.AddCommand(launchCmd)

	root.AddCommand(newCaptureCommand(&asJSON))
	root.AddCommand(newMemoryCommand(&asJSON))
	root.AddCommand(newRulesCommand(&asJSON))
	root.AddCommand(newHookCommand())
	root.AddCommand(newHistoryCommand(&asJSON))
	root.AddCommand(newDocsCommand(&asJSON))
	root.AddCommand(newIntentCommand(&asJSON))
	root.AddCommand(newSpecCommand(&asJSON))

	return root
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
	}

	var configPath string
	var rootDir string
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
					fmt.Fprintf(w, "%s:%d: [%s %s] %s\n",
						f.File, f.Line, strings.ToUpper(f.Severity), f.RuleID, f.Message)
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
	docsCmd.AddCommand(lintCmd)

	return docsCmd
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

// readHookInput reads and size-caps the hook stdin payload.
func readHookInput(cmd *cobra.Command) (hookInput, error) {
	raw, err := io.ReadAll(io.LimitReader(cmd.InOrStdin(), maxHookStdinBytes))
	if err != nil {
		return hookInput{}, err
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
		Args:   cobra.NoArgs,
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
			rs, err := rules.Load(cwd)
			if err != nil {
				fmt.Fprintf(cmd.ErrOrStderr(), "abcd rules: %v; injecting nothing\n", err)
				return nil
			}
			session := hookSession(in)
			// The fixed-N backstop comes from the repo's config (default 15 when
			// unset); event-driven reset is the primary refresh (D1).
			res := rules.Inject(rs, in.Prompt, rules.LoadState(session), rules.LoadBackstop(cwd))
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
		Short: "Stop: redact and store the session transcript",
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
			res, err := history.Capture(cwd, det.RootSHA, in.SessionID, raw, "native")
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

	return hookCmd
}

// maxTranscriptBytes caps the transcript read from disk. Generous for a JSONL
// session log, and bounded so a pathological file cannot stall the Stop hook
// while the scanner walks it.
const maxTranscriptBytes = 64 << 20 // 64 MiB

// readTranscript reads the file named by the Stop payload's transcript_path.
//
// The path is external input, so the open is defensive on the two failure modes
// that would actually hurt: O_NONBLOCK so a FIFO or device node cannot hang the
// hook (a hung Stop hook wedges the user's session), and a regular-file check so
// only a real file is ever read. The size cap bounds the redaction pass.
func readTranscript(path string) ([]byte, error) {
	f, err := os.OpenFile(path, os.O_RDONLY|syscall.O_NONBLOCK, 0)
	if err != nil {
		return nil, fmt.Errorf("cannot open transcript %q (%v)", path, err)
	}
	defer f.Close()

	st, err := f.Stat()
	if err != nil {
		return nil, fmt.Errorf("cannot stat transcript %q (%v)", path, err)
	}
	if !st.Mode().IsRegular() {
		return nil, fmt.Errorf("transcript %q is not a regular file", path)
	}
	if st.Size() > maxTranscriptBytes {
		return nil, fmt.Errorf("transcript %q is %d bytes, over the %d-byte cap", path, st.Size(), maxTranscriptBytes)
	}

	raw, err := io.ReadAll(io.LimitReader(f, maxTranscriptBytes))
	if err != nil {
		return nil, fmt.Errorf("cannot read transcript %q (%v)", path, err)
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
			rs, err := rules.Load(cwd)
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
	intentCmd := &cobra.Command{
		Use:   "intent",
		Short: "Intent lifecycle; bare invocation is read-only status",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			cwd, err := os.Getwd()
			if err != nil {
				return err
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
					fmt.Fprintf(w, "  link: %s -> %s\n", p.Intent, p.Spec)
				}
			})
		},
	}

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
			return render(cmd.OutOrStdout(), *asJSON, res, func(w io.Writer) {
				fmt.Fprintf(w, "abcd intent plan — %s drafts -> planned, linked %s\n", res.Intent.ID, res.Spec.ID)
				fmt.Fprintf(w, "  intent: %s\n", res.Intent.Path)
				fmt.Fprintf(w, "  spec:   %s\n", res.Spec.Path)
			})
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

	intentCmd.AddCommand(newIntentReviewCommand(asJSON))
	return intentCmd
}

// newIntentReviewCommand builds `abcd intent review`: `ingest --verdict-json`
// applies a host-produced intent-fidelity verdict to the shipped intent's Audit
// Notes (fail-closed: ingested | dead_letter | noop); bare `review <itd-N>`
// re-emits the OWED stub + ephemeral request for a shipped intent.
func newIntentReviewCommand(asJSON *bool) *cobra.Command {
	reviewCmd := &cobra.Command{
		Use:   "review [<itd-N>]",
		Short: "Fidelity review: re-emit a shipped intent's request, or ingest a verdict",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return cmd.Help()
			}
			cwd, err := os.Getwd()
			if err != nil {
				return err
			}
			res, err := intent.ReEmitReview(cwd, args[0])
			if err != nil {
				return &exitError{Code: 2, Msg: "abcd intent review: " + err.Error()}
			}
			return render(cmd.OutOrStdout(), *asJSON, res, func(w io.Writer) {
				fmt.Fprintf(w, "abcd intent review — %s %s (receipt %s)\n  request: %s\n",
					res.IntentID, res.Status, res.ReceiptID, res.RequestPath)
			})
		},
	}

	var verdictJSON string
	ingestCmd := &cobra.Command{
		Use:   "ingest --verdict-json <path>",
		Short: "Ingest an intent-fidelity verdict JSON into the shipped intent's Audit Notes",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			cwd, err := os.Getwd()
			if err != nil {
				return err
			}
			if verdictJSON == "" {
				return &exitError{Code: 2, Msg: "abcd intent review ingest: --verdict-json <path> is required"}
			}
			res, err := intent.IngestVerdict(cwd, verdictJSON)
			if err != nil {
				return &exitError{Code: 2, Msg: "abcd intent review ingest: " + err.Error()}
			}
			return render(cmd.OutOrStdout(), *asJSON, res, func(w io.Writer) {
				fmt.Fprintf(w, "abcd intent review ingest — %s (receipt %s, intent %s)\n", res.Status, res.ReceiptID, res.IntentID)
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
	ingestCmd.Flags().StringVar(&verdictJSON, "verdict-json", "", "path to the intent-fidelity verdict JSON")
	reviewCmd.AddCommand(ingestCmd)
	return reviewCmd
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
					fmt.Fprintf(w, "  %s  %s  %s  (%s)\n", sp.ID, sp.Status, sp.Slug, sp.Intent)
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
			if res.ReviewEmitError != "" {
				fmt.Fprintf(cmd.ErrOrStderr(), "WARNING: abcd spec close — fidelity-review emit failed for %s (intent shipped anyway): %s\n", res.Intent.ID, res.ReviewEmitError)
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
			return render(cmd.OutOrStdout(), *asJSON, res, func(w io.Writer) {
				fmt.Fprintf(w, "abcd ahoy — %s\n", res.FolderKind)
				fmt.Fprintf(w, "  plugin root: %s\n", res.PluginRootStatus)
				fmt.Fprintf(w, "  root sha:    %s\n", res.RootSHA)
				fmt.Fprintf(w, "  gaps:        %d\n", len(res.Gaps))
			})
		},
	}

	// install
	var (
		yes           bool
		adopt         bool
		refuseAdopt   bool
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
			opts, err := installOptionsFromFlags(cmd, yes, adopt, refuseAdopt, visibility, docsTarget, oracleBackend, scanDeep)
			if err != nil {
				return err
			}
			res, err := ahoy.Install(cwd, opts, newPrompter(cmd))
			if err != nil {
				return err
			}
			return render(cmd.OutOrStdout(), *asJSON, res, func(w io.Writer) {
				fmt.Fprintf(w, "abcd ahoy install — %s\n", res.Status)
				for _, p := range res.Writes {
					fmt.Fprintf(w, "  wrote: %s\n", p)
				}
				if len(res.DeclinedCategories) > 0 {
					fmt.Fprintf(w, "  declined: %s\n", strings.Join(res.DeclinedCategories, ", "))
				}
				if len(res.Remaining) > 0 {
					fmt.Fprintf(w, "  remaining gaps: %s\n", strings.Join(res.Remaining, ", "))
				}
			})
		},
	}
	installCmd.Flags().BoolVar(&yes, "yes", false, "approve every resolvable change category without prompting")
	installCmd.Flags().BoolVar(&adopt, "adopt", false, "adopt an unmanaged repo without prompting")
	installCmd.Flags().BoolVar(&refuseAdopt, "refuse-adopt", false, "decline to adopt an unmanaged repo")
	installCmd.Flags().StringVar(&visibility, "visibility", "", "repo visibility: private | public")
	installCmd.Flags().StringVar(&docsTarget, "docs-target", "", "marker target: claude_md | agents_md | both | skip")
	installCmd.Flags().StringVar(&oracleBackend, "oracle-backend", "", "oracle backend: host-delegated | native | cli | api | mcp")
	installCmd.Flags().StringVar(&scanDeep, "scan-deep", "", "enable deep scan: true | false")
	ahoyCmd.AddCommand(installCmd)

	// uninstall
	ahoyCmd.AddCommand(&cobra.Command{
		Use:   "uninstall",
		Short: "Remove the marker block and owned PATH symlink (leaves .abcd/ intact)",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			cwd, err := os.Getwd()
			if err != nil {
				return err
			}
			receipt, err := ahoy.Uninstall(cwd)
			if err != nil {
				return err
			}
			return render(cmd.OutOrStdout(), *asJSON, receipt, func(w io.Writer) {
				fmt.Fprintf(w, "abcd ahoy uninstall\n")
				fmt.Fprintf(w, "  marker removed: %v\n", receipt.Marker.Removed)
				fmt.Fprintf(w, "  symlink: %s\n", symlinkNote(receipt))
			})
		},
	})

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
func installOptionsFromFlags(cmd *cobra.Command, yes, adopt, refuseAdopt bool, visibility, docsTarget, oracleBackend, scanDeep string) (ahoy.InstallOptions, error) {
	opts := ahoy.InstallOptions{Yes: yes}
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

// newPrompter returns an interactive stdin prompter when stdin is a terminal,
// and a refusing prompter otherwise so non-interactive runs never block on input.
func newPrompter(cmd *cobra.Command) ahoy.Prompter {
	if f, ok := cmd.InOrStdin().(*os.File); ok {
		if fi, err := f.Stat(); err == nil && fi.Mode()&os.ModeCharDevice != 0 {
			return &stdinPrompter{r: bufio.NewReader(f), w: cmd.ErrOrStderr()}
		}
	}
	return ahoy.RefusingPrompter{}
}

// stdinPrompter is the interactive Prompter: it reads answers from stdin.
type stdinPrompter struct {
	r *bufio.Reader
	w io.Writer
}

func (p *stdinPrompter) Confirm(question string) bool {
	fmt.Fprintf(p.w, "%s [y/N] ", question)
	line, _ := p.r.ReadString('\n')
	line = strings.ToLower(strings.TrimSpace(line))
	return line == "y" || line == "yes"
}

func (p *stdinPrompter) Prompt(key string, choices []string, def string) string {
	fmt.Fprintf(p.w, "%s (%s) [%s]: ", key, strings.Join(choices, "/"), def)
	line, _ := p.r.ReadString('\n')
	line = strings.TrimSpace(line)
	if line == "" {
		return def
	}
	return line
}

// newCaptureCommand builds the `capture` sub-tree — the write side of the issue
// ledger. Bare `capture` renders read-only status; a free-text positional
// appends an issue; list/resolve/wontfix are thin consumers of capture core.
// (promote is skill-orchestrated, never a CLI sub-verb — brief 04-surfaces/06.)
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
					fmt.Fprintf(w, "  skipped %s: %s\n", sk.Path, sk.Error)
				}
			})
		},
	}
	listCmd.Flags().BoolVar(&lsOpen, "open", false, "issues currently in open/")
	listCmd.Flags().BoolVar(&lsResolved, "resolved", false, "issues currently in resolved/")
	listCmd.Flags().BoolVar(&lsWontfix, "wontfix", false, "issues currently in wontfix/")
	listCmd.Flags().BoolVar(&lsAll, "all", false, "issues across all three states")
	captureCmd.AddCommand(listCmd)

	// resolve — open -> resolved with a note.
	captureCmd.AddCommand(&cobra.Command{
		Use:   "resolve <iss-N> <note>",
		Short: "Mark an open issue resolved (open/ -> resolved/)",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cwd, err := os.Getwd()
			if err != nil {
				return err
			}
			res, err := capture.Resolve(capture.ResolveRequest{RepoRoot: cwd, ID: args[0], Resolution: args[1]})
			if err != nil {
				return err
			}
			return render(cmd.OutOrStdout(), *asJSON, res, func(w io.Writer) {
				fmt.Fprintf(w, "%s  %s -> %s — %s\n", res.ID, res.FromStatus, res.ToStatus, res.Path)
			})
		},
	})

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

// suspectedTypoedSubcommand reports the nearest real subverb when args[0] is a
// near-miss for one (edit distance 1–2) and the invocation shape resembles a
// subcommand call rather than free-text prose: a lone token, or a token
// followed by an issue id. It is deliberately high-precision so it never
// refuses a legitimate free-text capture whose first word merely resembles a
// verb — those carry no trailing iss-id and are multi-word.
func suspectedTypoedSubcommand(parent *cobra.Command, args []string) (string, bool) {
	if len(args) == 0 {
		return "", false
	}
	shapedLikeSubcommand := len(args) == 1 || issIDRe.MatchString(args[1])
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
					fmt.Fprintf(w, "  %s: %d\n", c.Class, c.Count)
				}
				if st.LastIngest != "" {
					fmt.Fprintf(w, "  last ingest: %s\n", st.LastIngest)
				}
				for _, line := range st.Contradictions {
					fmt.Fprintf(w, "  contradiction: %s\n", line)
				}
				for _, line := range st.Headroom {
					fmt.Fprintf(w, "  %s\n", line)
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
					fmt.Fprintf(w, "  %s [%s] %s:%d %s\n", f.Code, f.Severity, f.File, f.Line, f.Message)
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

// readSource reads a JSON payload from a file path, or from stdin when spec is
// "-" (the streaming transport the .5 skill uses).
func readSource(cmd *cobra.Command, spec string) ([]byte, error) {
	if spec == "-" {
		return io.ReadAll(cmd.InOrStdin())
	}
	return os.ReadFile(spec)
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
		RunE: func(cmd *cobra.Command, _ []string) error {
			return cmd.Help()
		},
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
			res, err := history.Capture(cwd, rootSHA, sess, raw, orDefault(kind, "native"))
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
			return render(cmd.OutOrStdout(), *asJSON, records, func(w io.Writer) {
				if len(records) == 0 {
					fmt.Fprintln(w, "abcd history — no transcripts stored for this repo")
					return
				}
				for _, r := range records {
					fmt.Fprintf(w, "%s  %s  %s  redacted secrets=%d home=%d\n",
						r.CapturedAt.Format("2006-01-02T15:04:05Z"), r.SessionID, r.SourceKind, r.Secrets, r.HomePaths)
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
				fmt.Fprintf(w, "session:    %s\n", rec.SessionID)
				fmt.Fprintf(w, "captured:   %s\n", rec.CapturedAt.Format("2006-01-02T15:04:05Z"))
				fmt.Fprintf(w, "source:     %s\n", rec.SourceKind)
				fmt.Fprintf(w, "path:       %s\n", rec.Path)
				fmt.Fprintf(w, "redacted:   secrets=%d home=%d\n", rec.Secrets, rec.HomePaths)
				fmt.Fprintln(w, "---")
				fmt.Fprint(w, string(body))
			})
		},
	})

	return historyCmd
}

// repoRootSHA resolves the current repo's root-commit SHA (the history store
// key) via the ahoy detection pass. An empty SHA means cwd is not a git repo
// with commits, which the history verbs cannot key on.
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
func scrubPaths(err error) string {
	msg := err.Error()
	if cwd, e := os.Getwd(); e == nil {
		msg = redactRoot(msg, cwd, ".")
	}
	if home, e := os.UserHomeDir(); e == nil {
		msg = redactRoot(msg, home, "~")
	}
	for _, p := range embeddedPaths(err) {
		if filepath.IsAbs(p) {
			msg = strings.ReplaceAll(msg, p, filepath.Base(p))
		}
	}
	return msg
}

// redactRoot replaces every occurrence of the absolute directory root (followed
// by a path separator) in s with repl. The filesystem root ("/") and empty or
// relative roots are skipped so a message is never mangled into meaninglessness.
func redactRoot(s, root, repl string) string {
	if len(root) <= 1 || !filepath.IsAbs(root) {
		return s
	}
	sep := string(os.PathSeparator)
	return strings.ReplaceAll(s, root+sep, repl+sep)
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
