package cli

// reading.go is the front door onto internal/core/reading — the cold-reading
// input assembler (itd-183, spc-61).
//
// The verb's whole interface is three closed operands. A position selects the
// reading's object from the include table; a target names the commit the
// assembly describes; a scope names what the reading is ABOUT, so an assembly
// passes the intersection of the two rather than a position's whole corpus
// (itd-199, admitted by adr-58). There is no free-text argument anywhere,
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
	"strconv"
	"strings"

	"github.com/intentdriven/abcd/internal/core/reading"
	"github.com/intentdriven/abcd/internal/termsafe"
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

	var position, target, outDir, scope string
	var dryRun bool
	assembleCmd := &cobra.Command{
		Use:   "assemble --position <position> --target <HEAD|sha> --scope <itd-N|spc-N|kind|preset>",
		Short: "Assemble one reading's input and its manifest",
		Long: "Walk the repository under the include table at one reading position and write two\n" +
			"artefacts: the assembled input, which carries no repository path, and the manifest,\n" +
			"which maps every passed item back to its path, its field and its hash.\n\n" +
			"The invocation carries no free text. --position takes one of four closed tokens;\n" +
			"--target takes HEAD or a hexadecimal commit sha of 7 to 40 digits, because a branch\n" +
			"or a tag moves and the manifest's re-runnability rests on a reference that cannot;\n" +
			"--scope names what the reading is about, as a record id (itd-N or spc-N), a material\n" +
			"kind, or a preset named in .abcd/config/reading-presets.json. All three are required.",
		Example: "  abcd reading assemble --position widening --target HEAD --scope cold --dry-run\n" +
			"  abcd reading assemble --position entailment --target HEAD --scope cold \\\n" +
			"    --out .abcd/.work.local/scratch/reading-runs/manual --json",
		Args: func(_ *cobra.Command, args []string) error {
			if len(args) > 0 {
				return &exitError{Code: 2, Msg: "reading assemble: this verb takes no positional argument; " +
					"a reading's object and question come from its definition, and the invocation " +
					"carries a position, a target state and a scope, and nothing else"}
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, _ []string) error {
			// All three operands are required rather than defaulted. A defaulted
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
			// Required for the same reason the other two are: a reading is
			// commissioned ABOUT something, and a defaulted scope would pick
			// that on the operator's behalf. It is a closed form, never prose
			// and never a path (adr-58).
			if scope == "" {
				return &exitError{Code: 2, Msg: "reading assemble: --scope is required: a record id " +
					"(itd-N, spc-N), a material kind, or a committed preset named in " +
					reading.PresetConfigPath}
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
				Scope:       scope,
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
		"the reading position: "+positionTokens()+"\n"+
			"(comparative does not assemble: its object is the widening reading's\n"+
			"pre-admission output, which no channel supplies, so it refuses)")
	assembleCmd.Flags().StringVar(&target, "target", "",
		"the commit the assembly describes: HEAD, or a hexadecimal sha of 7 to 40 digits")
	assembleCmd.Flags().StringVar(&scope, "scope", "",
		"what the reading is about: a record id (itd-N, spc-N), a material kind,\n"+
			"or a committed preset. No repository path is accepted here; a preset is where one may be named")
	assembleCmd.Flags().StringVar(&outDir, "out", "",
		"an empty or absent directory the assembled input and the manifest are written to\n"+
			"(default: the local-tier run directory)")
	assembleCmd.Flags().BoolVar(&dryRun, "dry-run", false,
		"write nothing; with --out the two artefacts still land in that directory")

	var readingJSON string
	ingestCmd := &cobra.Command{
		Use:   "ingest --reading-json <path>",
		Short: "Validate one reading's returned output and write its records",
		Long: "Validate the JSON a cold reading returned and write its reading records.\n\n" +
			"The verb checks what the reading was LICENSED to produce, not only what it saw: the\n" +
			"supply regime is read from the position's definition and compared with the output's own\n" +
			"claim, and the reserved names a regime declares are refused with the licence stated (the\n" +
			"generative position declares none). A registry of named signatures watches for prose that\n" +
			"ranks, settles or proposes without the field; those signatures are observed rather than\n" +
			"enforcing, at every position, so a hit raises a review flag on the run record and the item\n" +
			"lands.\n\n" +
			"Item identifiers are minted here. The payload carries none, so a supplied one is refused\n" +
			"as an unknown field. A refusal becomes DURABLE once the run's identity is proven — the run\n" +
			"id resolving to a parked manifest whose content hash matches — and from there a list-level\n" +
			"refusal writes refusal.json under the run's directory; before that point nothing is written\n" +
			"anywhere. Nothing durable is written or DELETED until the whole payload validates: a refusal\n" +
			"after the run is proven leaves its refusal record and nothing else. The reading records land\n" +
			"as one batch and the run metadata is written last as the commit marker: a run without one\n" +
			"never happened.\n\n" +
			"An ingest interrupted before that marker leaves an orphaned stage, and every invocation names\n" +
			"it. Only the next one whose payload validates sweeps it: it ROLLS THAT RUN'S READING RECORDS\n" +
			"OUT OF THE COMMITTED LEDGER because the run never happened, and clears the stage. A refused\n" +
			"run reports the orphans it left in place, and the ids a sweep removed are reported as\n" +
			"rolled_back_records on every exit, including a failing one.",
		Example: "  abcd reading ingest --reading-json ./reading-output.json --json",
		Args: func(_ *cobra.Command, args []string) error {
			if len(args) > 0 {
				return &exitError{Code: 2, Msg: "reading ingest: this verb takes no positional argument; " +
					"the output names its own run, position and regime, and there is no operand that " +
					"could set one"}
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, _ []string) error {
			if readingJSON == "" {
				return &exitError{Code: 2, Msg: "reading ingest: --reading-json <path> is required: " +
					"the JSON the reading returned"}
			}
			// The operator's path means what the shell means by it; the core is
			// handed the resolved one. The working directory is a transport fact
			// the core does not hold.
			cwd := mustCwd()
			resolved := readingJSON
			if !filepath.IsAbs(resolved) {
				resolved = filepath.Join(cwd, filepath.FromSlash(resolved))
			}
			res, err := reading.Ingest(reading.IngestRequest{
				RepoRoot:   captureRoot(cwd),
				OutputPath: resolved,
			})
			if err != nil {
				// A refusal that produced a durable record renders it before it
				// exits. The record path is the operator's handle on the event,
				// and the plugin page tells a host to report `refusal_record` —
				// which it could never find if the render only ran on success.
				//
				// The same goes for everything else the result discloses about
				// the committed tier: an orphaned stage seen and left in place,
				// a stage cleared, a record rolled back. The core owns the
				// predicate, so a refusal reached before the run's identity is
				// proven — which writes no record — still renders when it has
				// something to say. Keying the render on the refusal record
				// alone once left a delete in the ledger reported as a bare
				// type error (iss-2608311517509690). A refusal with nothing to
				// disclose renders nothing, because there is no run to name.
				if res.HasDisclosure() {
					_ = render(cmd.OutOrStdout(), *asJSON, res, func(w io.Writer) {
						renderIngestResult(w, res)
					})
				}
				return &exitError{Code: 2, Msg: "reading ingest: " + trimCorePrefix(scrubPaths(err))}
			}
			return render(cmd.OutOrStdout(), *asJSON, res, func(w io.Writer) {
				renderIngestResult(w, res)
			})
		},
	}
	ingestCmd.Flags().StringVar(&readingJSON, "reading-json", "",
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
	renderPiles(w, s.Piles)
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
	// Printed only when there is one. An orphan is an abnormal state — an ingest
	// that reached the ledger and never reached its commit marker, whose reading
	// records are sitting in the committed ledger for a run that never happened
	// — so it earns a line, and a "none" every other run would train a reader
	// past it.
	if len(s.OrphanedIngests) > 0 {
		fmt.Fprintf(w, "  interrupted:    %s; their reading records are in the ledger for a run with no "+
			"commit marker, and the next ingest that validates sweeps them\n",
			strings.Join(s.OrphanedIngests, ", "))
	}
	// A leftover stage is the other state a stage survives in, and it is NOT an
	// orphan: the run committed and only its stage failed to clear. It gets a
	// line of its own so it is never listed beside a run whose records are
	// about to be rolled back (iss-2609012043437282).
	if len(s.LeftoverStages) > 0 {
		fmt.Fprintf(w, "  leftover stage: %s; the run committed and its records stay, and the next ingest "+
			"that validates clears the stage alone\n",
			strings.Join(s.LeftoverStages, ", "))
	}
}

// renderPiles writes which pile each position assembles from.
//
// The shared case is stated rather than left implied. Positions sharing one
// pile is the DEFAULT and the thing the closing-run comparison has to be able
// to read off a run: a render that named only the exceptions would leave a
// reader inferring the rule from a silence (ruled 2026-09-01).
func renderPiles(w io.Writer, piles []reading.PositionPileStatus) {
	own := make([]string, 0, len(piles))
	for _, p := range piles {
		if p.Pile == reading.PileOwn {
			own = append(own, fmt.Sprintf("%s (%d row(s))", p.Position, p.Rows))
		}
	}
	if len(own) == 0 {
		fmt.Fprintf(w, "  piles:          one shared assembly at every position\n")
		return
	}
	fmt.Fprintf(w, "  piles:          own pile at %s; every other position shares one assembly\n",
		strings.Join(own, ", "))
	for _, p := range piles {
		if p.Pile != reading.PileOwn {
			continue
		}
		// Sanitised: a pile's rule is author-written prose that reaches a
		// rendered line, which invariant 13 holds to being termsafe.
		fmt.Fprintf(w, "    %-12s %s\n", p.Position+":", termsafe.Sanitize(p.Rule))
	}
}

// renderAssembleResult writes one assembly's text render.
func renderAssembleResult(w io.Writer, res reading.AssembleResult) {
	fmt.Fprintf(w, "%s: %d item(s) assembled at the %s position of %s\n",
		res.RunID, res.ItemCount, res.Position, shortSha(res.TargetCommit))
	fmt.Fprintf(w, "  manifest hash: %s\n", res.ManifestHash)
	// Which pile the run drew from, on the line above the scope, because the
	// two answer different questions: the pile is what the POSITION may see and
	// the scope is what this run was about.
	fmt.Fprintf(w, "  pile:          %s, %s\n", res.Pile.Source, res.Pile.Hash)
	renderScope(w, res.Scope)
	renderSizeReport(w, res.Size)
	if !res.Written {
		fmt.Fprintln(w, "  written:       nothing (dry run; name --out to write the two artefacts)")
		return
	}
	fmt.Fprintf(w, "  written:       %s and %s in %s\n",
		reading.BundleFileName, reading.ManifestFileName, res.OutDir)
}

// renderScope writes what the reading was commissioned about, and says plainly
// when the run departed from the committed presets — a run nobody can tell was
// an override is a run whose drift from the reviewed configuration is
// invisible.
func renderScope(w io.Writer, s reading.Scope) {
	clauses := make([]string, 0, len(s.Selectors))
	for _, sel := range s.Selectors {
		switch {
		case sel.Kind != "":
			clauses = append(clauses, string(sel.Kind))
		case sel.Record != "":
			clauses = append(clauses, sel.Record)
		case sel.Path != "":
			clauses = append(clauses, sel.Path+"/")
		}
	}
	note := ""
	if s.Overridden {
		note = " (overridden at invocation, not a committed preset)"
	}
	// Sanitised because these are runtime-read strings: the source token comes
	// from the operator and the path clauses come from a file on disk, and
	// invariant 13 holds every such string to being termsafe before it joins a
	// rendered line. This is the first render site in this file that emits file
	// content, so it is the first that needs it.
	fmt.Fprintf(w, "  scope:         %s%s\n", termsafe.Sanitize(s.Source), note)
	fmt.Fprintf(w, "    selects:     %s\n", termsafe.Sanitize(strings.Join(clauses, ", ")))
}

// renderSizeReport writes what an assembly would cost, per material kind and in
// total. It is written before the dry-run branch above returns, because the
// whole point of the report is that it is available WITHOUT writing an artefact
// (itd-198 ac-2).
func renderSizeReport(w io.Writer, s reading.SizeReport) {
	// "item text" is in the label deliberately. These bytes are what a reader
	// receives; bundle.json on disk is larger, because JSON escaping and the
	// per-item envelope ride on top. An operator who read a bare "size" here
	// and then stat'd the artefact two lines below would have been told a
	// number that does not describe the file they are looking at.
	fmt.Fprintf(w, "  size (item text): %s, ~%s tokens (%s)\n",
		humanBytes(s.Bytes), thousands(s.TokensEst), s.Basis)
	for _, k := range s.ByKind {
		fmt.Fprintf(w, "    %-18s %6d item(s)  %9s  ~%s tokens\n",
			k.Kind, k.Items, humanBytes(k.Bytes), thousands(k.TokensEst))
	}
}

// humanBytes renders a byte count at a readable scale. The unit is decimal, not
// binary: the figure exists for a human deciding whether an artefact is
// plausible, and 9.8 MB answers that where 9,371 KiB does not.
func humanBytes(n int) string {
	switch {
	case n >= 1_000_000_000:
		return fmt.Sprintf("%.1f GB", float64(n)/1e9)
	case n >= 1_000_000:
		return fmt.Sprintf("%.1f MB", float64(n)/1e6)
	case n >= 1_000:
		return fmt.Sprintf("%.1f kB", float64(n)/1e3)
	default:
		return fmt.Sprintf("%d B", n)
	}
}

// thousands groups an integer for reading. A token estimate is the number a
// human compares against a capacity they hold in their head, and an ungrouped
// seven-digit run is the shape that gets misread by an order of magnitude.
func thousands(n int) string {
	s := strconv.Itoa(n)
	neg := ""
	if strings.HasPrefix(s, "-") {
		neg, s = "-", s[1:]
	}
	var b strings.Builder
	for i, c := range s {
		if i > 0 && (len(s)-i)%3 == 0 {
			b.WriteByte(',')
		}
		b.WriteRune(c)
	}
	return neg + b.String()
}

// trimCorePrefix drops the core package's own tag from a message this front door
// is about to prefix with the verb's name.
//
// Printing both reads as a stutter — `abcd: reading ingest: reading: ...` — and
// the refusal message is load-bearing for this verb: six of itd-185's thirteen
// criteria require the offending field, the item's ordinal or the signature id
// to be NAMED, so a message a reader skims past is a criterion half-met rather
// than a cosmetic complaint. The `assemble` path above carries the same stutter
// and is deliberately left alone: it is a change to a shipped verb's output
// rather than to what this delivery adds, and it is captured separately.
func trimCorePrefix(msg string) string {
	return strings.TrimPrefix(msg, "reading: ")
}

// renderIngestResult writes one ingest's text render.
//
// The refused items and the review flags are rendered by ORDINAL, rule and
// signature id, never by body text: the bodies belong in the ledger records,
// which the record writer redacts on the way in, and a refusal quoting a
// reading's prose back to a terminal would leave that redaction behind.
func renderIngestResult(w io.Writer, res reading.IngestResult) {
	// Two headings this render cannot write. A refusal before the run's
	// identity is proven has no run to head the render with; what follows is
	// then the disclosure alone. And a run whose DEFINITION did not resolve is
	// proven but carries no regime — the record leaves that field empty on
	// purpose, because the regime is the definition's — so the render says so
	// rather than interpolating the emptiness, which read as "under the  regime":
	// a doubled space asserting a regime that is not there.
	switch {
	case res.RunID == "":
		fmt.Fprintln(w, "no run: the output was refused before the run it names was proven")
	case res.Regime == "":
		fmt.Fprintf(w, "%s: %d record(s) at the %s position; the regime did not resolve\n",
			res.RunID, len(res.Records), res.Position)
	default:
		fmt.Fprintf(w, "%s: %d record(s) at the %s position under the %s regime\n",
			res.RunID, len(res.Records), res.Position, res.Regime)
	}
	if res.RunRecordPath != "" {
		fmt.Fprintf(w, "  run metadata:  %s\n", res.RunRecordPath)
	}
	if res.RefusalPath != "" {
		fmt.Fprintf(w, "  refused:       the run; recorded at %s\n", res.RefusalPath)
	}
	// Neither marker means the run did not finish: the records above may have
	// reached the ledger, and without this the header's record count reads as an
	// outcome for a run that never committed.
	if res.RunRecordPath == "" && res.RefusalPath == "" {
		fmt.Fprintln(w, "  committed:     no; the run has no commit marker, so it never happened")
	}
	if res.RefusedCount > 0 {
		fmt.Fprintf(w, "  refused items: %d\n", res.RefusedCount)
	}
	for _, r := range res.RefusedItems {
		// The elision entry names no item, so it is not rendered as one: there
		// is no item 0, and printing one would send a reader looking for it.
		if r.Ordinal == 0 {
			fmt.Fprintf(w, "                 (%s) %s\n", r.Rule, r.Detail)
			continue
		}
		fmt.Fprintf(w, "                 item %d (%s): %s\n", r.Ordinal, r.Rule, r.Detail)
	}
	for _, f := range res.ReviewFlags {
		fmt.Fprintf(w, "  review flag:   item %d matches %s\n", f.Ordinal, f.SignatureID)
	}
	if len(res.ClearedStages) > 0 {
		fmt.Fprintf(w, "  cleared:       orphaned stage(s) of %s\n", strings.Join(res.ClearedStages, ", "))
	}
	if len(res.RolledBack) > 0 {
		fmt.Fprintf(w, "  rolled back:   %s removed from the ledger (their run never committed)\n",
			strings.Join(res.RolledBack, ", "))
	}
	if len(res.PendingStages) > 0 {
		fmt.Fprintf(w, "  left in place: orphaned stage(s) of %s; the sweep runs only under a payload that "+
			"validates, and the next ingest that does rolls them back and clears them\n",
			strings.Join(res.PendingStages, ", "))
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
