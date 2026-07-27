package cli

// The `abcd docs cite` sub-tree: the two verbs that maintain the citation
// baseline.
//
// `refresh` is the only place abcd dials out on behalf of documentation, and it
// runs when a maintainer asks. `confirm` closes the manual queue that refresh
// prints. Both live under `docs` because the baseline exists to serve
// `docs lint` — the gate is the customer, these verbs are how it gets fed.

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/REPPL/abcd-cli/internal/core/cite"
	"github.com/REPPL/abcd-cli/internal/core/launch"
	"github.com/REPPL/abcd-cli/internal/core/lint"
	"github.com/REPPL/abcd-cli/internal/termsafe"
)

// newCiteCommand builds the `docs cite` sub-tree. configPath and rootDir are the
// docs command's own flag targets, so `cite` resolves its config exactly the way
// `lint` does and the two verbs can never read different roots.
func newCiteCommand(asJSON *bool) *cobra.Command {
	citeCmd := &cobra.Command{
		Use:   "cite",
		Short: "Maintain the citation baseline the docs lint enforces offline",
		Args:  cobra.NoArgs,
	}
	citeCmd.AddCommand(newCiteRefreshCommand(asJSON))
	citeCmd.AddCommand(newCiteConfirmCommand(asJSON))
	return citeCmd
}

func newCiteRefreshCommand(asJSON *bool) *cobra.Command {
	var configPath, rootDir string
	cmd := &cobra.Command{
		Use:   "refresh",
		Short: "Fetch every cited URL once and rewrite the committed citation baseline",
		Long: "Fetch every cited URL once and rewrite the committed citation baseline.\n\n" +
			"This is the only abcd verb that reaches the network on behalf of documentation. " +
			"Each URL gets exactly one bounded attempt; a failure is recorded as an outcome, " +
			"never retried. Sources that refuse automated fetchers are printed as a manual " +
			"checklist for `abcd docs cite confirm` rather than recorded as broken.",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			root, cfg, err := loadCiteConfig("docs cite refresh", rootDir, configPath)
			if err != nil {
				return err
			}
			res, err := cite.Refresh(cite.RefreshRequest{RepoRoot: root, Config: cfg})
			if err != nil {
				return &exitError{Code: 2, Msg: "docs cite refresh: " + err.Error()}
			}
			return render(cmd.OutOrStdout(), *asJSON, res, func(w io.Writer) {
				renderRefresh(w, res)
			})
		},
	}
	addCiteFlags(cmd, &configPath, &rootDir)
	return cmd
}

// renderRefresh prints the run as a transcript an operator can act on: the
// counts, then every broken link, then the manual queue as a checklist. URLs and
// details come from repo content and third-party servers, so every one is
// sanitised before it reaches a terminal.
func renderRefresh(w io.Writer, res cite.RefreshResult) {
	ok, broken, blocked, preserved := res.Counts()
	fmt.Fprintf(w, "abcd docs cite refresh — %d cited, %d checked, %d preserved\n",
		res.Cited, res.Fetched, preserved)
	fmt.Fprintf(w, "  alive:   %d\n", ok)
	fmt.Fprintf(w, "  broken:  %d\n", broken)
	fmt.Fprintf(w, "  blocked: %d\n", blocked)

	for _, o := range res.Outcomes {
		if o.Status != cite.StatusBroken {
			continue
		}
		fmt.Fprintf(w, "  BROKEN %s — %s\n", termsafe.Sanitize(o.URL), termsafe.Sanitize(o.Detail))
	}
	for _, u := range res.Dropped {
		fmt.Fprintf(w, "  dropped (no longer cited): %s\n", termsafe.Sanitize(u))
	}

	if len(res.Queue) > 0 {
		fmt.Fprintf(w, "\nmanual queue — %d source(s) refuse automated fetching.\n", len(res.Queue))
		fmt.Fprintf(w, "Open each, confirm the document is there, then record it:\n\n")
		for _, q := range res.Queue {
			fmt.Fprintf(w, "  [ ] %s\n", termsafe.Sanitize(q.URL))
			if q.Detail != "" {
				fmt.Fprintf(w, "        %s\n", termsafe.Sanitize(q.Detail))
			}
			for _, s := range q.Sites {
				fmt.Fprintf(w, "        cited at %s:%d\n", termsafe.Sanitize(s.File), s.Line)
			}
		}
		fmt.Fprintf(w, "\n  abcd docs cite confirm <url> [<url>...]\n")
	}
	fmt.Fprintf(w, "\nbaseline written: %s\n", res.BaselinePath)
}

func newCiteConfirmCommand(asJSON *bool) *cobra.Command {
	var configPath, rootDir, receiptPath string
	cmd := &cobra.Command{
		Use:   "confirm [url...]",
		Short: "Record that a human verified a cited URL the fetcher could not read",
		Long: "Record that a human verified a cited URL the fetcher could not read.\n\n" +
			"Name the URLs directly, or pass --receipt with a receipt file. Both write the " +
			"same dated manual entry: the baseline records THAT a human confirmed the " +
			"citation and WHEN, never how. Only URLs the documentation actually cites can " +
			"be confirmed.",
		Args: cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			root, cfg, err := loadCiteConfig("docs cite confirm", rootDir, configPath)
			if err != nil {
				return err
			}

			// The two producers of one schema: positional arguments assemble the
			// simple case in memory, a file carries the full form. Neither is a
			// second pathway — both become the same Receipt before anything is
			// validated or written.
			var receipt cite.Receipt
			switch {
			case receiptPath != "" && len(args) > 0:
				return &exitError{Code: 2, Msg: "docs cite confirm: pass URLs or --receipt, not both"}
			case receiptPath != "":
				receipt, err = cite.LoadReceipt(receiptPath)
				if err != nil {
					return &exitError{Code: 2, Msg: "docs cite confirm: " + citeErrorDetail(err, receiptPath)}
				}
			case len(args) > 0:
				receipt.SchemaVersion = cite.ReceiptSchemaVersion
				for _, u := range args {
					receipt.Confirmed = append(receipt.Confirmed, cite.ConfirmedCitation{URL: u})
				}
			default:
				return &exitError{Code: 2, Msg: "docs cite confirm: name at least one URL, or pass --receipt <file>"}
			}

			res, err := cite.Confirm(cite.ConfirmRequest{RepoRoot: root, Config: cfg, Receipt: receipt})
			if err != nil {
				return &exitError{Code: 2, Msg: "docs cite confirm: " + err.Error()}
			}
			return render(cmd.OutOrStdout(), *asJSON, res, func(w io.Writer) {
				for _, u := range res.Recorded {
					fmt.Fprintf(w, "  confirmed %s\n", termsafe.Sanitize(u))
				}
				fmt.Fprintf(w, "abcd docs cite confirm — %d receipt(s) recorded in %s\n",
					len(res.Recorded), res.BaselinePath)
			})
		},
	}
	addCiteFlags(cmd, &configPath, &rootDir)
	cmd.Flags().StringVar(&receiptPath, "receipt", "",
		"path to a receipt file listing the confirmed citations (the format the generated checklist page emits)")
	return cmd
}

func addCiteFlags(cmd *cobra.Command, configPath, rootDir *string) {
	cmd.Flags().StringVar(configPath, "config", "", "path to docs-lint.json (default: <root>/.abcd/docs-lint.json)")
	cmd.Flags().StringVar(rootDir, "root", "", "repo root (default: current working directory)")
}

// loadCiteConfig resolves the root and the docs-lint config the same way
// `docs lint` does, so the refresh fetches exactly the set the gate enforces.
func loadCiteConfig(verb, rootDir, configPath string) (string, lint.Config, error) {
	root := rootDir
	if root == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return "", lint.Config{}, err
		}
		root = cwd
	}
	root, err := filepath.Abs(root)
	if err != nil {
		return "", lint.Config{}, err
	}
	cfgPath := configPath
	if cfgPath == "" {
		cfgPath = filepath.Join(root, ".abcd", "docs-lint.json")
	}
	cfg, err := lint.LoadConfig(cfgPath)
	if err != nil {
		ref := filepath.Join(".abcd", "docs-lint.json")
		if configPath != "" {
			ref = configPath
		}
		if os.IsNotExist(err) {
			return "", lint.Config{}, &exitError{Code: 2, Msg: fmt.Sprintf(
				"%s: config not found at %s — run in a prepared repo or pass --config", verb, ref)}
		}
		return "", lint.Config{}, &exitError{Code: 2, Msg: fmt.Sprintf(
			"%s: cannot read config %s: %s", verb, ref, citeErrorDetail(err, ref))}
	}
	return root, cfg, nil
}

// citationPreflight measures the citation baseline for the release preflight.
//
// It lives on this side of the boundary because internal/core/lint imports
// internal/core/launch for its semver, so the measurement cannot be taken inside
// the gate it feeds. Everything is soft: a repo with no docs-lint config, or one
// that has not armed the citation rule, gets a nil preflight and the gate
// reports itself as not run rather than inventing a pass.
func citationPreflight(repoRoot string) *launch.CitationPreflight {
	cfg, err := lint.LoadConfig(filepath.Join(repoRoot, ".abcd", "docs-lint.json"))
	if err != nil {
		return nil
	}
	if rc, ok := cfg.Rules["citation_baseline"]; !ok || !rc.Enabled {
		return nil
	}
	sum, err := lint.CitationAgeSummary(cfg, repoRoot)
	if err != nil {
		// An unreadable baseline is not "no gate" — it is a record a release must
		// not proceed on, so it is reported as a present-but-empty one whose
		// citations all lack receipts, which refuses.
		cited, cerr := lint.CollectCitedURLs(cfg, repoRoot)
		if cerr != nil {
			return nil
		}
		return &launch.CitationPreflight{Present: true, Cited: len(cited), Missing: len(cited)}
	}
	return &launch.CitationPreflight{
		Present:         sum.Present,
		Cited:           sum.Cited,
		Recorded:        sum.Recorded,
		Missing:         sum.Missing,
		Broken:          sum.Broken,
		Approaching:     sum.Approaching,
		Overdue:         sum.Overdue,
		ApproachingURLs: sum.ApproachingURLs,
		OverdueURLs:     sum.OverdueURLs,
	}
}

// citeErrorDetail strips the path-bearing wrapper off a filesystem error. A
// *PathError embeds the ABSOLUTE path, which must never reach machine output
// (iss-29); its inner Err is the bare cause.
func citeErrorDetail(err error, ref string) string {
	var pe *os.PathError
	if errors.As(err, &pe) {
		return ref + ": " + pe.Err.Error()
	}
	return strings.TrimSpace(err.Error())
}
