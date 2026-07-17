package cli

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/REPPL/abcd-cli/internal/core/audit"
	"github.com/REPPL/abcd-cli/internal/termsafe"
)

// newAuditCommand wires the read-only `abcd audit` verb: it evaluates the
// bundled conformance rules against the working directory and reports them, human
// text by default or machine JSON with --json. It never writes. The exit code is
// Conftest's tri-state — 0 clean, 1 warnings only, 2 any error — so `abcd audit`
// can gate a repo's CI.
func newAuditCommand(asJSON *bool) *cobra.Command {
	var rootDir string
	cmd := &cobra.Command{
		Use:   "audit",
		Short: "Check this repo against the working conventions (read-only)",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			dir := rootDir
			if dir == "" {
				cwd, err := os.Getwd()
				if err != nil {
					return err
				}
				dir = cwd
			}
			dir, err := filepath.Abs(dir)
			if err != nil {
				return err
			}
			// Validate the root before evaluating: the rules read ENOENT as
			// "artifact missing" and would emit confident, fabricated convention
			// violations against a directory that is not there (B41). Fail with a
			// usage error instead, matching the disembark probe guard.
			if info, statErr := os.Stat(dir); statErr != nil || !info.IsDir() {
				shown := rootDir
				if shown == "" {
					shown = dir
				}
				return &exitError{Code: 2, Msg: fmt.Sprintf("audit: %s is not a directory", shown)}
			}

			result, err := audit.Evaluate(audit.DefaultRules(), audit.Context{RepoRoot: dir})
			if err != nil {
				return err
			}

			if *asJSON {
				out, err := audit.JSONSerializer{}.Serialize(result)
				if err != nil {
					return err
				}
				fmt.Fprintln(cmd.OutOrStdout(), string(out))
			} else {
				renderAuditHuman(cmd.OutOrStdout(), result)
			}

			// Tri-state exit: the output is already rendered, so a non-zero code
			// propagates with an empty message (main prints nothing more).
			if result.ExitCode != 0 {
				return &exitError{Code: result.ExitCode}
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&rootDir, "root", "", "repo root to audit (default: current working directory)")
	return cmd
}

// severityGlyph returns the doctor-style marker for a severity.
func severityGlyph(sev audit.Severity) string {
	switch sev {
	case audit.SeverityError:
		return "✗"
	case audit.SeverityWarn:
		return "⚠"
	default:
		return "•"
	}
}

// renderAuditHuman writes the grouped, doctor-style report: a line per finding
// with a severity glyph, the rule id, the citation, the message, and an indented
// fix; skipped (not-applicable) rules and a summary tail. A clean repo gets a
// single green line.
func renderAuditHuman(w io.Writer, res audit.Result) {
	if len(res.Findings) == 0 {
		fmt.Fprintln(w, "abcd audit — ✓ conforms to the working conventions")
	}
	for _, f := range res.Findings {
		// File, Message and Fix are built from the audited repo's own file paths and
		// content (`abcd audit` runs over any repo), so they are untrusted terminal
		// output and pass through the canonical sanitiser; RuleID/Severity are enum
		// constants and need none.
		loc := termsafe.Sanitize(f.File)
		if f.Line > 0 {
			loc = fmt.Sprintf("%s:%d", termsafe.Sanitize(f.File), f.Line)
		}
		fmt.Fprintf(w, "%s [%s] %s — %s\n", severityGlyph(f.Severity), f.RuleID, loc, termsafe.Sanitize(f.Message))
		if f.Fix != "" {
			fmt.Fprintf(w, "    fix: %s\n", termsafe.Sanitize(f.Fix))
		}
	}
	for _, id := range res.Skipped {
		fmt.Fprintf(w, "• [%s] skipped (not applicable)\n", id)
	}
	fmt.Fprintf(w, "abcd audit — %d error(s), %d warning(s)\n", res.Blockers, res.Warnings)
}
