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
			"  abcd reading assemble --position entailment --target HEAD --out ./run --json",
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
			res, err := reading.Assemble(reading.AssembleRequest{
				RepoRoot: captureRoot(mustCwd()),
				Position: pos,
				Target:   target,
				OutDir:   outDir,
				DryRun:   dryRun,
			})
			if err != nil {
				return &exitError{Code: 2, Msg: "reading assemble: " + scrubPaths(err)}
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
		"the directory the assembled input and the manifest are written to (default: the local-tier run directory)")
	assembleCmd.Flags().BoolVar(&dryRun, "dry-run", false,
		"write nothing into the repository; with --out the two artefacts still land there")

	readingCmd.AddCommand(assembleCmd)
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

// shortSha abbreviates a commit for the text render.
func shortSha(sha string) string {
	if len(sha) > 12 {
		return sha[:12]
	}
	return sha
}
