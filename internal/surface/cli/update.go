package cli

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/Partnermedia/abcd/internal/core/ahoy"
	"github.com/Partnermedia/abcd/internal/core/update"
	"github.com/Partnermedia/abcd/internal/termsafe"
)

// newUpdater is the seam through which `abcd update` reaches the network — a
// package var like newReleaseFetcher, so a test can prove the dispatch
// refuses BEFORE anything network-capable exists and that no other verb ever
// constructs it (the adr-38 seam extended to the update verb, spc-32 AC8).
var newUpdater = update.NewGitHubUpdater

func newUpdateCommand(asJSON *bool) *cobra.Command {
	var yes bool
	cmd := &cobra.Command{
		Use:   "update [tag]",
		Short: "Complete a chosen update: fetch, verify, and swap the PATH-installed binary",
		Long: "Fetches the named release (or resolves the latest, naming it before acting),\n" +
			"verifies the platform binary against the same release's checksums.txt, and\n" +
			"swaps the PATH-installed copy atomically. The verb is the only ask: abcd\n" +
			"never checks for or applies updates on its own (adr-38). A plugin-root\n" +
			"binary, the dev shim, and package-manager installs are refused with the\n" +
			"command that owns them.",
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			requested := ""
			if len(args) == 1 {
				requested = args[0]
			}
			// Dispatch first, network never: every refusal shape exits before
			// the updater exists.
			tgt := ahoy.ResolveUpdateTarget()
			if r := update.Plan(tgt); r != nil {
				rep := update.Report{Origin: "", Action: update.ActionRefused, TargetPath: tgt.Path, Refusal: r}
				renderUpdateReport(cmd.OutOrStdout(), *asJSON, rep)
				return fmt.Errorf("update refused (%s)", r.Shape)
			}

			u := newUpdater()
			tag, err := u.ResolveTag(requested)
			if err != nil {
				return err
			}
			// Name the tag before acting. On a TTY the bare form asks; an
			// explicit tag was already the user's own naming.
			if requested == "" && !yes && isTTY(os.Stdin) && isTTY(os.Stderr) {
				fmt.Fprintf(cmd.ErrOrStderr(), "resolved latest release: %s — proceed? [y/N] ", termsafe.Sanitize(tag))
				line, _ := bufio.NewReader(cmd.InOrStdin()).ReadString('\n')
				if a := strings.ToLower(strings.TrimSpace(line)); a != "y" && a != "yes" {
					fmt.Fprintln(cmd.ErrOrStderr(), "declined — nothing was fetched.")
					return nil
				}
			}
			var progress io.Writer
			if isTTY(os.Stderr) {
				progress = cmd.ErrOrStderr()
			}
			rep, err := u.Apply(tgt.Path, tag, progress)
			if err != nil {
				return err
			}
			renderUpdateReport(cmd.OutOrStdout(), *asJSON, rep)
			if rep.Refusal != nil {
				return fmt.Errorf("update refused (%s)", rep.Refusal.Shape)
			}
			return nil
		},
	}
	cmd.Flags().BoolVar(&yes, "yes", false, "skip the TTY confirmation of a freshly resolved tag")
	return cmd
}

// isTTY reports whether f is a character device — the progress/confirmation
// gate. Piped and hooked invocations are silent except for the receipt.
func isTTY(f *os.File) bool {
	fi, err := f.Stat()
	return err == nil && fi.Mode()&os.ModeCharDevice != 0
}

// renderUpdateReport prints the receipt in both modes. Tags and paths pass
// through termsafe on the text render: the tag is read off HTTP responses.
func renderUpdateReport(w io.Writer, asJSON bool, rep update.Report) {
	_ = render(w, asJSON, rep, func(w io.Writer) {
		switch rep.Action {
		case update.ActionSwapped:
			fmt.Fprintf(w, "updated %s: %s -> %s\n", termsafe.Sanitize(rep.TargetPath), termsafe.Sanitize(rep.OldVersion), termsafe.Sanitize(rep.NewVersion))
			fmt.Fprintf(w, "  origin:   %s\n", rep.Origin)
			fmt.Fprintf(w, "  verified: sha256 %s (release checksums.txt)\n", termsafe.Sanitize(rep.Digest))
		case update.ActionCurrent:
			fmt.Fprintf(w, "already current: %s is %s (verified against the release checksums)\n", termsafe.Sanitize(rep.TargetPath), termsafe.Sanitize(rep.Tag))
		case update.ActionRefused:
			if rep.Refusal != nil {
				fmt.Fprintf(w, "refused (%s): %s\n", rep.Refusal.Shape, termsafe.Sanitize(rep.Refusal.Detail))
				fmt.Fprintf(w, "  remedy: %s\n", rep.Refusal.Remedy)
			}
		}
		if len(rep.EnvIgnored) > 0 {
			fmt.Fprintf(w, "  ignored from the environment: %s\n", strings.Join(rep.EnvIgnored, ", "))
		}
	})
}
