package cli

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/REPPL/abcd-cli/internal/core/changelog"
	"github.com/REPPL/abcd-cli/internal/core/release"
	"github.com/REPPL/abcd-cli/internal/termsafe"
	"github.com/spf13/cobra"
)

// emitCut is the front door's half of the release cut: it walks the LIVE cobra
// tree and the live manifests for the current surface, then hands them to the
// core, which owns every judgement. It is the same split as GuardSurface, for
// the same reason — internal/core may not know cobra.
func emitCut(repoRoot string) (release.Cut, error) {
	current, err := SurfaceSnapshot(repoRoot)
	if err != nil {
		return release.Cut{}, err
	}
	return release.Emit(repoRoot, current)
}

// ingestCut is emitCut's write-half twin: the same live-surface walk, plus the
// two things a core must never reach for itself — the untrusted payload, and the
// clock. The date lands in a durable release heading, so it is an argument a test
// can pin rather than a wall-clock read buried in a writer.
func ingestCut(repoRoot string, raw []byte) (release.IngestResult, error) {
	current, err := SurfaceSnapshot(repoRoot)
	if err != nil {
		return release.IngestResult{}, err
	}
	return release.Ingest(repoRoot, current, raw, time.Now())
}

// newLaunchShipCommand builds `abcd launch ship`, the release-cut write verb.
//
// It has the dual-mode shape of the disembark synthesis verbs, and for the same
// reason: a deterministic step and a host-delegated one are two halves of ONE
// verb, not two commands that could drift apart.
//
//   - WITHOUT --changelog-json it runs the EMIT step: derive the version from
//     the records that shipped, run the surface guardrail, and render the cut a
//     changelog composer works from. This step writes nothing.
//   - WITH --changelog-json it runs the INGEST step: validate the host's composed
//     prose against the cut's record set (the completeness bijection) and write
//     the dated CHANGELOG heading. The payload is untrusted host output and is
//     read behind the shared guarded-operand path.
//
// Exit codes are the machine seam an autonomous run gates on, following `abcd
// intent ready`:
//
//	0  the cut is ready — it may go to the composer; with a payload, the dated
//	   section is written.
//	1  the cut REFUSES. The whole report is rendered first and every refusal
//	   names the specific record, version, or surface that blocks it; the code
//	   is the only extra signal. A refusal is a result, not a crash.
//	2  a structural fault — the repository could not be read, an operand was
//	   unusable, or the composed changelog does not describe the cut. The
//	   diagnostic is path-scrubbed. Nothing is written on this path.
func newLaunchShipCommand(asJSON *bool) *cobra.Command {
	var changelogJSON string
	cmd := &cobra.Command{
		Use:   "ship [--changelog-json <file|->]",
		Short: "Cut a release: derive the version and the record set from what shipped (exit 1 when the cut refuses)",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			cwd, err := os.Getwd()
			if err != nil {
				return err
			}
			// Read the payload before anything else: it is untrusted host input,
			// and reading it through the shared guarded-operand path keeps the
			// ingest seam behind exactly the trust boundary the other delegated
			// verbs use. An empty flag is the sentinel for deterministic mode.
			raw, err := readSynthesisPayload(cmd, changelogJSON)
			if err != nil {
				return &exitError{Code: 2, Msg: "abcd launch ship: " + scrubPaths(err)}
			}
			if raw != nil {
				res, err := ingestCut(cwd, raw)
				if err != nil {
					return &exitError{Code: 2, Msg: "abcd launch ship: " + scrubPaths(err)}
				}
				if rerr := render(cmd.OutOrStdout(), *asJSON, res, func(w io.Writer) {
					renderIngest(w, res)
				}); rerr != nil {
					return rerr
				}
				if !res.Cut.Ready {
					return &exitError{Code: 1}
				}
				return nil
			}

			cut, err := emitCut(cwd)
			if err != nil {
				return &exitError{Code: 2, Msg: "abcd launch ship: " + scrubPaths(err)}
			}
			if rerr := render(cmd.OutOrStdout(), *asJSON, cut, func(w io.Writer) {
				renderCut(w, "abcd launch ship", cut)
			}); rerr != nil {
				return rerr
			}
			if !cut.Ready {
				return &exitError{Code: 1}
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&changelogJSON, "changelog-json", "",
		"path to the host-composed changelog JSON (or - for stdin); absent runs the deterministic emit step")
	return cmd
}

// newChangelogCommand builds `abcd changelog`, the DETERMINISTIC-ONLY preview of
// the next release cut.
//
// It renders exactly what `abcd launch ship` would emit — the derived version,
// the deciding impact, the record list, and the guardrail status — and no prose:
// the changelog text is composed once, at the reviewed ship, never twice with a
// chance of disagreeing. It performs ZERO writes.
//
// It always exits 0 (2 only on a structural fault), which is the deliberate
// difference from `launch ship`. This is a status render in the shape of `abcd
// capture` bare and `abcd launch --dry-run`: a refused cut is information a
// reader asked for, not a gate they tripped. The gate is the ship verb.
func newChangelogCommand(asJSON *bool) *cobra.Command {
	return &cobra.Command{
		Use:   "changelog",
		Short: "Preview the next release cut — derived version, records, guardrail (read-only, no prose)",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			cwd, err := os.Getwd()
			if err != nil {
				return err
			}
			cut, err := emitCut(cwd)
			if err != nil {
				return &exitError{Code: 2, Msg: "abcd changelog: " + scrubPaths(err)}
			}
			return render(cmd.OutOrStdout(), *asJSON, cut, func(w io.Writer) {
				renderCut(w, "abcd changelog", cut)
			})
		},
	}
}

// renderCut writes the human rendering of a cut, shared by the preview and the
// ship verb so the two can never describe the same repository differently.
//
// Every value that came out of a record — a title, a path, a refusal quoting
// either — is sanitised: record frontmatter and prose are author-supplied text
// that reaches a terminal here.
func renderCut(w io.Writer, verb string, cut release.Cut) {
	base := cut.BaseTag
	if base == "" {
		base = "(no release tag)"
	}
	if cut.Ready {
		fmt.Fprintf(w, "%s — %s -> %s (%s)\n", verb, base, cut.NextTag, cut.Impact)
	} else {
		fmt.Fprintf(w, "%s — REFUSED (base %s)\n", verb, base)
	}
	if len(cut.DecidedBy) > 0 {
		fmt.Fprintf(w, "  decided by: %s\n", termsafe.Sanitize(strings.Join(cut.DecidedBy, ", ")))
	}
	fmt.Fprintf(w, "  guard:      %s\n", guardLine(cut.Guard))
	renderEntries(w, "added", cut.Added)
	renderEntries(w, "removed", cut.Removed)
	for _, refusal := range cut.Refusals {
		fmt.Fprintf(w, "  refused (%s):\n", refusal.Kind)
		for _, line := range strings.Split(termsafe.Sanitize(refusal.Reason), "\n") {
			fmt.Fprintf(w, "    %s\n", strings.TrimRight(line, " "))
		}
	}
}

// renderIngest writes the human rendering of a completed ship: the same cut
// report the emit step renders, then what landed in the release record.
//
// The cut is rendered FIRST and unchanged, so the two halves of one verb read
// identically — an operator comparing a dry preview with the ship that followed
// is comparing the same lines, not two dialects of the same report.
func renderIngest(w io.Writer, res release.IngestResult) {
	renderCut(w, "abcd launch ship", res.Cut)
	if !res.Written {
		return
	}
	fmt.Fprintf(w, "  wrote:      %s\n", res.Path)
	fmt.Fprintf(w, "    %s\n", res.Heading)
	fmt.Fprintf(w, "    %d line(s), citing %s\n", res.Lines, termsafe.Sanitize(strings.Join(res.Cited, ", ")))
}

// guardLine renders the guardrail verdict, keeping a REFUSED guard visually
// distinct from a passing one: conflating "nothing broke" with "I could not
// tell" is the whole risk the guardrail exists to close.
func guardLine(g changelog.SurfaceGuard) string {
	line := string(g.Status)
	if line == "" {
		line = "(not run)"
	}
	if len(g.Breaks) > 0 {
		line += fmt.Sprintf(" (%d surface change(s))", len(g.Breaks))
	}
	return line
}

// renderEntries lists one side of the cut, marking the records the changelog
// prose must NOT cite so a reader can see why a record is present but silent.
func renderEntries(w io.Writer, label string, entries []release.Entry) {
	fmt.Fprintf(w, "  %-11s %d record(s)\n", label+":", len(entries))
	for _, e := range entries {
		note := ""
		if !e.InChangelog {
			note = "  (excluded from the changelog)"
		}
		fmt.Fprintf(w, "    [%-9s] %-8s %s%s\n", e.Impact, termsafe.Sanitize(e.ID), termsafe.Sanitize(e.Title), note)
	}
}
