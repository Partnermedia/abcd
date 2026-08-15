package cli

import (
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"

	"github.com/REPPL/abcd-cli/internal/core"
	"github.com/REPPL/abcd-cli/internal/core/ahoy"
	"github.com/REPPL/abcd-cli/internal/core/vintage"
)

// newReleaseFetcher is the seam through which `version --check` reaches the
// network. It is a package var so a test can inject a recording stub and prove
// two things: the check path fetches exactly once, and NO other path (version
// bare, ahoy, session-start, install) ever constructs it — the zero-network
// invariant of adr-38 tier 1.
var newReleaseFetcher = vintage.NewGitHubReleaseFetcher

// checkSource names the network source `version --check` consulted, so the
// report says where "latest" came from (AC5).
const checkSource = "github.com/REPPL/abcd-cli releases"

// versionOutput is `abcd version`'s structured result: identity, install mode,
// vintage, and staleness relative to the on-disk reference. The vintage fields
// come from one comparator (ahoy.Vintage), shared with `abcd ahoy` and the
// session-start notice. Check is populated only by --check.
type versionOutput struct {
	core.VersionInfo
	InstallMode string       `json:"install_mode,omitempty"`
	Vintage     string       `json:"vintage"`
	Staleness   string       `json:"staleness"`
	Check       *checkResult `json:"check,omitempty"`
}

// ahoyOutput is `abcd ahoy`'s bare render: the detection envelope plus the
// shared vintage/staleness, so the JSON surface relays them too (AC3).
type ahoyOutput struct {
	ahoy.DetectionResult
	Vintage   string `json:"vintage"`
	Staleness string `json:"staleness"`
}

// checkResult is the explicit network check's outcome, named source and all.
type checkResult struct {
	Latest  string `json:"latest,omitempty"`
	Source  string `json:"source"`
	Verdict string `json:"verdict"`
}

func newVersionCommand(asJSON *bool) *cobra.Command {
	var check bool
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Print abcd's version, install mode, and vintage",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			cwd, err := os.Getwd()
			if err != nil {
				return err
			}
			v := core.NewVersion()
			vin := ahoy.Vintage(cwd)
			out := versionOutput{
				VersionInfo: v,
				InstallMode: vin.Mode,
				Vintage:     vin.DisplayVintage(),
				Staleness:   vin.Staleness(),
			}
			// The ONLY network path in itd-111 (adr-38 tier 2): an explicit
			// --check fetches the latest release once and compares through the
			// same comparator as every disk path.
			if check {
				out.Check = runReleaseCheck(v.Version)
			}
			return render(cmd.OutOrStdout(), *asJSON, out, func(w io.Writer) {
				fmt.Fprintf(w, "%s %s\n", v.Name, v.Version)
				if out.InstallMode != "" {
					fmt.Fprintf(w, "  install:   %s\n", out.InstallMode)
				}
				fmt.Fprintf(w, "  vintage:   %s\n", out.Vintage)
				fmt.Fprintf(w, "  staleness: %s\n", out.Staleness)
				if out.Check != nil {
					if out.Check.Latest != "" {
						fmt.Fprintf(w, "  latest:    %s (source: %s)\n", out.Check.Latest, out.Check.Source)
					}
					fmt.Fprintf(w, "  check:     %s\n", out.Check.Verdict)
				}
			})
		},
	}
	cmd.Flags().BoolVar(&check, "check", false, "fetch the latest release once and compare (the only command that touches the network); names its source")
	return cmd
}

// runReleaseCheck fetches the latest release exactly once and compares it
// against the running version through the shared comparator, naming the source.
func runReleaseCheck(currentVersion string) *checkResult {
	res := &checkResult{Source: checkSource}
	// The single network touch: Expected() performs the one fetch.
	exp := vintage.ReleaseProvider(newReleaseFetcher()).Expected()
	if !exp.Known {
		res.Verdict = "the latest release could not be determined (no network, or the source is unreachable)"
		return res
	}
	res.Latest = exp.Revision
	cur := vintage.Current{Revision: currentVersion, Known: currentVersion != "" && currentVersion != "dev"}
	if !cur.Known {
		res.Verdict = "latest is " + exp.Revision + "; this is an unversioned (dev) build, so there is nothing to compare"
		return res
	}
	// Compare through the shared comparator over the tag already fetched, so the
	// fetch is not repeated.
	if vintage.Compare(cur, vintage.PinnedVersion(exp.Revision)).Outcome == vintage.Fresh {
		res.Verdict = "up to date"
	} else {
		res.Verdict = "update available: " + currentVersion + " -> " + exp.Revision
	}
	return res
}
