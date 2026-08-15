package cli

import (
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"

	"github.com/REPPL/abcd-cli/internal/core"
	"github.com/REPPL/abcd-cli/internal/core/ahoy"
)

// versionOutput is `abcd version`'s structured result: identity, install mode,
// vintage, and staleness relative to the on-disk reference. The vintage fields
// come from one comparator (ahoy.Vintage), shared with `abcd ahoy` and the
// session-start notice.
type versionOutput struct {
	core.VersionInfo
	InstallMode string `json:"install_mode,omitempty"`
	Vintage     string `json:"vintage"`
	Staleness   string `json:"staleness"`
}

func newVersionCommand(asJSON *bool) *cobra.Command {
	return &cobra.Command{
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
			return render(cmd.OutOrStdout(), *asJSON, out, func(w io.Writer) {
				fmt.Fprintf(w, "%s %s\n", v.Name, v.Version)
				if out.InstallMode != "" {
					fmt.Fprintf(w, "  install:   %s\n", out.InstallMode)
				}
				fmt.Fprintf(w, "  vintage:   %s\n", out.Vintage)
				fmt.Fprintf(w, "  staleness: %s\n", out.Staleness)
			})
		},
	}
}
