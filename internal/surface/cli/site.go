package cli

import (
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"

	"github.com/Partnermedia/abcd/internal/core/site"
	"github.com/Partnermedia/abcd/internal/termsafe"
)

// newSiteCommand builds the `site` verb family: the website as a rendered
// surface of this repository (adr-47).
//
// Bare `abcd site` is a STATUS BOARD and writes nothing — what the repo has
// declared (the composition manifest, the interface-string allowlist, the
// reference baseline) and what the last build left in the output directory.
// `site build` renders: the landing page and record.json from repository text
// and committed assets, into a directory the repository does not track.
func newSiteCommand(asJSON *bool) *cobra.Command {
	siteCmd := &cobra.Command{
		Use:   "site",
		Short: "The website rendered from this repository: what is declared, and what was built (read-only)",
		Args:  cobra.NoArgs,
	}

	var statusOut string
	siteCmd.RunE = func(cmd *cobra.Command, _ []string) error {
		cwd, err := os.Getwd()
		if err != nil {
			return err
		}
		st, err := site.Describe(cwd, statusOut)
		if err != nil {
			return &exitError{Code: 2, Msg: "abcd site: " + scrubPaths(err)}
		}
		return render(cmd.OutOrStdout(), *asJSON, st, func(w io.Writer) {
			renderSiteStatus(w, st)
		})
	}
	siteCmd.Flags().StringVar(&statusOut, "out", site.DefaultOutDir, "output directory to report on")

	var buildOut string
	var version, commit, stampDate string
	buildCmd := &cobra.Command{
		Use:   "build",
		Short: "Render the site into the output directory (writes nothing outside it)",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			cwd, err := os.Getwd()
			if err != nil {
				return err
			}
			res, err := site.Build(site.Request{
				RepoRoot: cwd,
				OutDir:   buildOut,
				Stamp:    site.BuildStamp{Version: version, Commit: commit, GeneratedAt: stampDate},
			})
			if err != nil {
				return &exitError{Code: 2, Msg: "abcd site build: " + scrubPaths(err)}
			}
			return render(cmd.OutOrStdout(), *asJSON, res, func(w io.Writer) {
				renderSiteBuild(w, res)
			})
		},
	}
	buildCmd.Flags().StringVar(&buildOut, "out", site.DefaultOutDir, "directory to render into")
	buildCmd.Flags().StringVar(&version, "version", "", "version for the footer and the build stamp (default: the newest dated CHANGELOG heading)")
	buildCmd.Flags().StringVar(&commit, "commit", "", "commit for the footer and the build stamp (default: git HEAD)")
	buildCmd.Flags().StringVar(&stampDate, "date", "", "date for the build stamp (default: the newest release's date)")
	siteCmd.AddCommand(buildCmd)

	return siteCmd
}

// renderSiteStatus prints the read-only board.
func renderSiteStatus(w io.Writer, st site.Status) {
	fmt.Fprintf(w, "abcd site — %s\n", mark(st.Manifest, site.ManifestRelPath))
	if !st.Manifest {
		fmt.Fprintf(w, "  this repo declares no site composition; nothing to build\n")
		return
	}
	fmt.Fprintf(w, "  chapters:     %d\n", st.Chapters)
	fmt.Fprintf(w, "  issue ledger: %s\n", publishedWord(st.IssueLedge))
	fmt.Fprintf(w, "  ui strings:   %s\n", mark(st.UIStrings, termsafe.Sanitize(st.UIPath)))
	if st.Baseline {
		fmt.Fprintf(w, "  baseline:     %s (%d unresolved references admitted)\n",
			termsafe.Sanitize(st.BaselinePath), st.BaselineN)
	} else {
		fmt.Fprintf(w, "  baseline:     absent (%s)\n", termsafe.Sanitize(st.BaselinePath))
	}
	if st.Version != "" {
		fmt.Fprintf(w, "  release:      v%s\n", termsafe.Sanitize(st.Version))
	}
	if st.Commit != "" {
		fmt.Fprintf(w, "  commit:       %s\n", termsafe.Sanitize(st.Commit))
	}
	if st.OutExists {
		fmt.Fprintf(w, "  output:       %s (%d entries)\n", termsafe.Sanitize(st.OutDir), st.OutFiles)
	} else {
		fmt.Fprintf(w, "  output:       %s (not built)\n", termsafe.Sanitize(st.OutDir))
	}
	fmt.Fprintf(w, "run `abcd site build` to render\n")
}

// renderSiteBuild prints what a build wrote and what it measured.
func renderSiteBuild(w io.Writer, res site.Result) {
	fmt.Fprintf(w, "abcd site build — %s\n", termsafe.Sanitize(res.OutDir))
	// The explorer writes a page per record, so a full listing is hundreds of
	// lines of scroll. The head of it still says what shape the tree took, and
	// the count says how much of it is not shown.
	const listed = 12
	for i, f := range res.Files {
		if i == listed && len(res.Files) > listed+1 {
			fmt.Fprintf(w, "  … %d more\n", len(res.Files)-listed)
			break
		}
		fmt.Fprintf(w, "  %s\n", termsafe.Sanitize(f))
	}
	fmt.Fprintf(w, "  pages:   %d rendered from the record\n", res.Pages)
	fmt.Fprintf(w, "  record:  %d records · %d links · %d mentions\n", res.Records, res.Links, res.Mentions)
	fmt.Fprintf(w, "  refs:    %d unresolved (baseline %d)\n", res.Unresolved, res.Baseline)
	fmt.Fprintf(w, "  layout:  %d overlapping bubbles\n", res.Overlaps)
	stamp := res.Version
	if stamp != "" {
		stamp = "v" + stamp
	}
	if res.Commit != "" {
		if stamp != "" {
			stamp += " · "
		}
		stamp += res.Commit
	}
	if stamp != "" {
		fmt.Fprintf(w, "  built:   %s\n", termsafe.Sanitize(stamp))
	}
	fmt.Fprintf(w, "wrote %d files (%d bytes)\n", len(res.Files), res.Bytes)
}

// mark renders a present/absent line for one declared input.
func mark(present bool, label string) string {
	if present {
		return label
	}
	return label + " (absent)"
}

// publishedWord says whether the working-tier issue ledger is opted in.
func publishedWord(on bool) string {
	if on {
		return "published"
	}
	return "not published"
}
