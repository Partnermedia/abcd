package cli

// reading.go is the front door onto internal/core/reading — the cold-reading
// input assembler (itd-183, spc-61).
//
// The verb's whole interface is two closed operands. A position selects the
// reading's object from the include table; a target names the commit the
// assembly describes. There is no free-text argument anywhere, at any position,
// because a prose operand is a channel ledger content could travel down in the
// framing of a request, and the point of the assembler is that no such channel
// exists (ruling (5) of 2026-08-28).
//
// Nothing here runs a reading. The verb produces the input a reading would be
// given and the manifest an auditor checks it by; dispatching it to a reader is
// host work, and the host obligation to grant that reader no repository access
// is stated in the plugin surface, never claimed as an enforcement this binary
// performs.
//
// Exit codes follow the ideate shape: 0 when the assembly landed, 2 for every
// structural refusal (an unknown position, a target that is not a commit, a
// dirty included path, a free-text operand).

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/intentdriven/abcd/internal/core/reading"
	"github.com/spf13/cobra"
)

// newReadingCommand builds the `reading` sub-tree.
func newReadingCommand(asJSON *bool) *cobra.Command {
	readingCmd := &cobra.Command{
		Use:   "reading",
		Short: "Cold-reading input assembler: what a reading sees, and the manifest proving it",
		Long: "Assemble the input a cold reading is handed.\n\n" +
			"Blindness is a property of the input, not a promise the reader makes: a positive include\n" +
			"table names what may travel, fields are projected out of records rather than files copied\n" +
			"whole, and a hashed manifest records what was passed so a reader can judge contamination\n" +
			"rather than accept a disclosure on trust.\n\n" +
			"Bare `abcd reading` renders the assembler's state and writes nothing.",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			status, err := reading.Describe(captureRoot(mustCwd()))
			if err != nil {
				return &exitError{Code: 2, Msg: "reading: " + scrubPaths(err)}
			}
			return render(cmd.OutOrStdout(), *asJSON, status, func(w io.Writer) {
				renderReadingStatus(w, status)
			})
		},
	}

	var position, target, outDir string
	var dryRun bool
	assembleCmd := &cobra.Command{
		Use:   "assemble --position <position> --target <HEAD|sha>",
		Short: "Assemble one reading's input and its manifest",
		Long: "Walk the repository under the include table at one reading position and write two\n" +
			"artefacts: the assembled input, which carries no repository path, and the manifest,\n" +
			"which maps every passed item back to its path, its field and its hash.\n\n" +
			"The invocation carries no free text. --position takes one of four closed tokens;\n" +
			"--target takes HEAD or a hexadecimal commit sha of 7 to 40 digits, because a branch\n" +
			"or a tag moves and the manifest's re-runnability rests on a reference that cannot.",
		Example: "  abcd reading assemble --position widening --target HEAD --dry-run\n" +
			"  abcd reading assemble --position entailment --target HEAD \\\n" +
			"    --out .abcd/.work.local/scratch/reading-runs/manual --json",
		Args: func(_ *cobra.Command, args []string) error {
			if len(args) > 0 {
				return &exitError{Code: 2, Msg: "reading assemble: this verb takes no positional argument; " +
					"a reading's object and question come from its definition, and the invocation " +
					"carries a position and a target state and nothing else"}
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, _ []string) error {
			// The two operands are required rather than defaulted. A defaulted
			// position would pick a reading's object on the operator's behalf, and
			// a defaulted target would let the manifest name a commit nobody chose.
			if position == "" {
				return &exitError{Code: 2, Msg: "reading assemble: --position is required, one of " +
					positionTokens()}
			}
			if target == "" {
				return &exitError{Code: 2, Msg: "reading assemble: --target is required: HEAD, " +
					"or a hexadecimal commit sha of 7 to 40 digits"}
			}
			pos, err := reading.ParsePosition(position)
			if err != nil {
				return &exitError{Code: 2, Msg: "reading assemble: " + scrubPaths(err)}
			}
			// The core takes a relative output directory against the REPOSITORY
			// root, which is right for the default it computes itself and wrong
			// for a path an operator typed: `--out ./run` from a subdirectory
			// means what the shell means by it. The front door is where that
			// difference belongs, because the working directory is a transport
			// fact the core does not hold.
			cwd := mustCwd()
			resolvedOut := outDir
			if resolvedOut != "" && !filepath.IsAbs(resolvedOut) {
				resolvedOut = filepath.Join(cwd, filepath.FromSlash(resolvedOut))
			}
			res, err := reading.Assemble(reading.AssembleRequest{
				RepoRoot:    captureRoot(cwd),
				Position:    pos,
				Target:      target,
				OutDir:      resolvedOut,
				OutDirLabel: outDir,
				DryRun:      dryRun,
			})
			if err != nil {
				return &exitError{Code: 2, Msg: "reading assemble: " + scrubPaths(err)}
			}
			// The core was handed the resolved path so it could write there; the
			// operator is shown the string they typed. A resolved absolute path on
			// the success surface is a local path leaving the machine the moment
			// the plugin page's "report out_dir" instruction is followed.
			if outDir != "" {
				res.OutDir = outDir
			}
			return render(cmd.OutOrStdout(), *asJSON, res, func(w io.Writer) {
				renderAssembleResult(w, res)
			})
		},
	}
	assembleCmd.Flags().StringVar(&position, "position", "",
		"the reading position: "+positionTokens())
	assembleCmd.Flags().StringVar(&target, "target", "",
		"the commit the assembly describes: HEAD, or a hexadecimal sha of 7 to 40 digits")
	assembleCmd.Flags().StringVar(&outDir, "out", "",
		"an empty or absent directory the assembled input and the manifest are written to\n"+
			"(default: the local-tier run directory)")
	assembleCmd.Flags().BoolVar(&dryRun, "dry-run", false,
		"write nothing; with --out the two artefacts still land in that directory")

	var outputJSON string
	ingestCmd := &cobra.Command{
		Use:   "ingest --output-json <path>",
		Short: "Validate one reading's returned output and write its records",
		Long: "Validate the JSON a cold reading returned and write its reading records.\n\n" +
			"The verb checks what the reading was LICENSED to produce, not only what it saw: the\n" +
			"supply regime is read from the position's definition and compared with the output's own\n" +
			"claim, each regime's reserved names are refused with the licence stated, and a registry\n" +
			"of named signatures catches prose that ranks, settles or proposes without the field.\n\n" +
			"Item identifiers are minted here. The payload carries none, so a supplied one is refused\n" +
			"as an unknown field. Nothing durable is written until the whole payload validates, and\n" +
			"the run metadata is written last as the commit marker: a run without one never happened,\n" +
			"and an orphaned stage is named and cleared by the next invocation.",
		Example: "  abcd reading ingest --output-json ./reading-output.json --json",
		Args: func(_ *cobra.Command, args []string) error {
			if len(args) > 0 {
				return &exitError{Code: 2, Msg: "reading ingest: this verb takes no positional argument; " +
					"the output names its own run, position and regime, and there is no operand that " +
					"could set one"}
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, _ []string) error {
			if outputJSON == "" {
				return &exitError{Code: 2, Msg: "reading ingest: --output-json <path> is required: " +
					"the JSON the reading returned"}
			}
			// The operator's path means what the shell means by it; the core is
			// handed the resolved one. The working directory is a transport fact
			// the core does not hold.
			cwd := mustCwd()
			resolved := outputJSON
			if !filepath.IsAbs(resolved) {
				resolved = filepath.Join(cwd, filepath.FromSlash(resolved))
			}
			res, err := reading.Ingest(reading.IngestRequest{
				RepoRoot:   captureRoot(cwd),
				OutputPath: resolved,
			})
			if err != nil {
				return &exitError{Code: 2, Msg: "reading ingest: " + scrubPaths(err)}
			}
			return render(cmd.OutOrStdout(), *asJSON, res, func(w io.Writer) {
				renderIngestResult(w, res)
			})
		},
	}
	ingestCmd.Flags().StringVar(&outputJSON, "output-json", "",
		"path to the JSON the cold reading returned")

	readingCmd.AddCommand(assembleCmd)
	readingCmd.AddCommand(ingestCmd)
	return readingCmd
}

// positionTokens renders the closed position set for a flag description and a
// refusal message, composed from the core rather than spelled twice.
func positionTokens() string {
	names := make([]string, 0, len(reading.Positions()))
	for _, p := range reading.Positions() {
		names = append(names, string(p))
	}
	return strings.Join(names, ", ")
}

// mustCwd reads the working directory, falling back to "." so a render never
// fails on a directory question it does not need answered.
func mustCwd() string {
	cwd, err := os.Getwd()
	if err != nil {
		return "."
	}
	return cwd
}

// renderReadingStatus writes the bare verb's text render.
func renderReadingStatus(w io.Writer, s reading.Status) {
	fmt.Fprintf(w, "reading assembler %s (schema %d)\n", s.AssemblerVersion, s.SchemaVersion)
	fmt.Fprintf(w, "  include rows:   %d across %d positions (%s)\n",
		s.IncludeRows, len(s.Positions), positionTokens())
	fmt.Fprintf(w, "  exclusion rows: %d, asserted into every manifest\n", s.ExclusionRows)
	fmt.Fprintf(w, "  charter:        %s\n", s.CharterPath)
	if len(s.Definitions) == 0 {
		fmt.Fprintf(w, "  definitions:    none found under %s/\n", reading.DefinitionsDir)
	} else {
		fmt.Fprintf(w, "  definitions:    %s\n", strings.Join(s.Definitions, ", "))
	}
	if len(s.StagedRuns) == 0 {
		fmt.Fprintln(w, "  staged runs:    none")
	} else {
		fmt.Fprintf(w, "  staged runs:    %s\n", strings.Join(s.StagedRuns, ", "))
	}
}

// renderAssembleResult writes one assembly's text render.
func renderAssembleResult(w io.Writer, res reading.AssembleResult) {
	fmt.Fprintf(w, "%s: %d item(s) assembled at the %s position of %s\n",
		res.RunID, res.ItemCount, res.Position, shortSha(res.TargetCommit))
	fmt.Fprintf(w, "  manifest hash: %s\n", res.ManifestHash)
	if !res.Written {
		fmt.Fprintln(w, "  written:       nothing (dry run; name --out to write the two artefacts)")
		return
	}
	fmt.Fprintf(w, "  written:       %s and %s in %s\n",
		reading.BundleFileName, reading.ManifestFileName, res.OutDir)
}

// renderIngestResult writes one ingest's text render.
//
// The refused items and the review flags are rendered by ORDINAL, rule and
// signature id, never by body text: the bodies belong in the ledger records,
// which the record writer redacts on the way in, and a refusal quoting a
// reading's prose back to a terminal would leave that redaction behind.
func renderIngestResult(w io.Writer, res reading.IngestResult) {
	fmt.Fprintf(w, "%s: %d record(s) at the %s position under the %s regime\n",
		res.RunID, len(res.Records), res.Position, res.Regime)
	if res.RunRecordPath != "" {
		fmt.Fprintf(w, "  run metadata:  %s\n", res.RunRecordPath)
	}
	for _, r := range res.RefusedItems {
		fmt.Fprintf(w, "  refused:       item %d (%s): %s\n", r.Ordinal, r.Rule, r.Detail)
	}
	for _, f := range res.ReviewFlags {
		fmt.Fprintf(w, "  review flag:   item %d matches %s\n", f.Ordinal, f.SignatureID)
	}
	if len(res.ClearedStages) > 0 {
		fmt.Fprintf(w, "  cleared:       orphaned stage(s) of %s\n", strings.Join(res.ClearedStages, ", "))
	}
	if res.Degraded != "" {
		fmt.Fprintf(w, "  redaction:     %s\n", res.Degraded)
	}
}

// shortSha abbreviates a commit for the text render.
func shortSha(sha string) string {
	if len(sha) > 12 {
		return sha[:12]
	}
	return sha
}
