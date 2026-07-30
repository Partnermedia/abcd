package cli

import (
	"fmt"
	"io"
	"os"

	"github.com/REPPL/abcd-cli/internal/core/banlist"
	"github.com/REPPL/abcd-cli/internal/termsafe"
	"github.com/spf13/cobra"
)

// newBanlistCommand builds the `banlist` verb — the front door onto
// internal/core/banlist (itd-74, spc-20). Bare `abcd banlist` renders both layers
// read-only (abcd's bare-command-as-status convention); `add` and `remove` carry
// the mutations and take the layer explicitly, because writing a private pattern
// into committed config is the exact accident this feature exists to prevent.
func newBanlistCommand(asJSON *bool) *cobra.Command {
	banlistCmd := &cobra.Command{
		Use:   "banlist",
		Short: "Banned-names layers (bare renders both, read-only); add/remove maintain them",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cwd, err := os.Getwd()
			if err != nil {
				return err
			}
			rep, err := banlist.List(cwd)
			if err != nil {
				return err
			}
			return render(cmd.OutOrStdout(), *asJSON, rep, func(w io.Writer) {
				renderPrivateLayer(w, rep.Private)
				fmt.Fprintln(w)
				renderPublicLayer(w, rep.Public)
			})
		},
	}

	banlistCmd.AddCommand(newBanlistListCommand(asJSON))
	banlistCmd.AddCommand(newBanlistAddCommand(asJSON))
	banlistCmd.AddCommand(newBanlistRemoveCommand(asJSON))
	return banlistCmd
}

// newBanlistListCommand is the explicit read verb. Unscoped it renders both layers
// (the same result the bare command renders); a layer flag scopes it to one.
func newBanlistListCommand(asJSON *bool) *cobra.Command {
	var private, public bool
	listCmd := &cobra.Command{
		Use:   "list [--private | --public]",
		Short: "Render the banlist layers; private entries render by key only",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cwd, err := os.Getwd()
			if err != nil {
				return err
			}
			if private && public {
				return &exitError{Code: 2, Msg: "abcd banlist list: --private and --public are alternatives; omit both to render every layer"}
			}
			switch {
			case private:
				rep, err := banlist.ListPrivate(cwd)
				if err != nil {
					return usageError("abcd banlist list", err)
				}
				return render(cmd.OutOrStdout(), *asJSON, rep, func(w io.Writer) { renderPrivateLayer(w, rep) })
			case public:
				rep, err := banlist.ListPublic(cwd)
				if err != nil {
					return usageError("abcd banlist list", err)
				}
				return render(cmd.OutOrStdout(), *asJSON, rep, func(w io.Writer) { renderPublicLayer(w, rep) })
			}
			rep, err := banlist.List(cwd)
			if err != nil {
				return usageError("abcd banlist list", err)
			}
			return render(cmd.OutOrStdout(), *asJSON, rep, func(w io.Writer) {
				renderPrivateLayer(w, rep.Private)
				fmt.Fprintln(w)
				renderPublicLayer(w, rep.Public)
			})
		},
	}
	addLayerFlags(listCmd, &private, &public)
	return listCmd
}

// newBanlistAddCommand adds one entry to the named layer.
func newBanlistAddCommand(asJSON *bool) *cobra.Command {
	var private, public bool
	var severity, successor string
	addCmd := &cobra.Command{
		Use:   "add --private|--public <key> <pattern>",
		Short: "Add one banned-name entry to the named layer",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cwd, err := os.Getwd()
			if err != nil {
				return err
			}
			layer, err := chooseLayer("abcd banlist add", private, public)
			if err != nil {
				return err
			}
			key, pattern := args[0], args[1]
			if layer == layerPrivate {
				res, err := banlist.AddPrivate(banlist.AddPrivateRequest{RepoRoot: cwd, Key: key, Pattern: pattern})
				if err != nil {
					return usageError("abcd banlist add", err)
				}
				return render(cmd.OutOrStdout(), *asJSON, res, func(w io.Writer) {
					// The key, never the pattern: the value the user just typed stays
					// out of the terminal, the scrollback, and any captured log.
					fmt.Fprintf(w, "private banlist — added %s (%d entr%s, %s)\n",
						termsafe.Sanitize(res.Key), res.Entries, plural(res.Entries), res.Path)
					fmt.Fprintln(w, "  the pattern is withheld from output by design; it lives only on this machine")
				})
			}
			res, err := banlist.AddPublic(banlist.AddPublicRequest{
				RepoRoot: cwd, Key: key, Pattern: pattern, Severity: severity, Successor: successor,
			})
			if err != nil {
				return usageError("abcd banlist add", err)
			}
			return render(cmd.OutOrStdout(), *asJSON, res, func(w io.Writer) {
				fmt.Fprintf(w, "public banlist — added %s (%s) in %s\n",
					termsafe.Sanitize(res.Entry.ID), res.Entry.Severity, res.Path)
				fmt.Fprintf(w, "  pattern: %s\n", termsafe.Sanitize(res.Entry.Pattern))
				fmt.Fprintln(w, "  commit the config change: the public layer is enforced in CI, for everyone")
			})
		},
	}
	addLayerFlags(addCmd, &private, &public)
	addCmd.Flags().StringVar(&severity, "severity", "", "public entry severity: blocker (default) | warn")
	addCmd.Flags().StringVar(&successor, "successor", "", "public entry's replacement, cited in the finding (default \"a generic term\")")
	return addCmd
}

// newBanlistRemoveCommand removes one entry from the named layer.
func newBanlistRemoveCommand(asJSON *bool) *cobra.Command {
	var private, public bool
	removeCmd := &cobra.Command{
		Use:   "remove --private|--public <key>",
		Short: "Remove one banned-name entry from the named layer",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cwd, err := os.Getwd()
			if err != nil {
				return err
			}
			layer, err := chooseLayer("abcd banlist remove", private, public)
			if err != nil {
				return err
			}
			if layer == layerPrivate {
				res, err := banlist.RemovePrivate(cwd, args[0])
				if err != nil {
					return usageError("abcd banlist remove", err)
				}
				return render(cmd.OutOrStdout(), *asJSON, res, func(w io.Writer) {
					fmt.Fprintf(w, "private banlist — removed %s (%d entr%s remain, %s)\n",
						termsafe.Sanitize(res.Key), res.Entries, plural(res.Entries), res.Path)
				})
			}
			res, err := banlist.RemovePublic(cwd, args[0])
			if err != nil {
				return usageError("abcd banlist remove", err)
			}
			return render(cmd.OutOrStdout(), *asJSON, res, func(w io.Writer) {
				fmt.Fprintf(w, "public banlist — removed %s from %s\n", termsafe.Sanitize(res.Entry.ID), res.Path)
			})
		},
	}
	addLayerFlags(removeCmd, &private, &public)
	return removeCmd
}

// layer names the two stores.
type layer int

const (
	layerPrivate layer = iota
	layerPublic
)

// addLayerFlags declares the layer selectors on a sub-verb. They are alternatives,
// never defaults: the destination of a banned pattern is the one decision this verb
// must never make on the user's behalf.
func addLayerFlags(cmd *cobra.Command, private, public *bool) {
	cmd.Flags().BoolVar(private, "private", false, "the gitignored per-machine layer ("+banlist.PrivateRelPath+")")
	cmd.Flags().BoolVar(public, "public", false, "the committed, CI-enforced layer ("+banlist.PublicConfigRelPath+")")
}

// chooseLayer resolves the mutually exclusive layer flags. Neither flag and both
// flags are usage errors: a mutation must name its layer, because a private pattern
// written into the committed config is precisely the leak this feature prevents.
// The requirement is enforced here rather than by a cobra required-flag annotation,
// keeping the tree's no-required-flags invariant intact.
func chooseLayer(verb string, private, public bool) (layer, error) {
	switch {
	case private && public:
		return 0, &exitError{Code: 2, Msg: verb + ": --private and --public are alternatives; name exactly one layer"}
	case private:
		return layerPrivate, nil
	case public:
		return layerPublic, nil
	default:
		return 0, &exitError{Code: 2, Msg: verb + ": name the layer — --private (per-machine, untracked) or --public (committed, CI-enforced)"}
	}
}

// usageError renders a core refusal as an exit-2 usage failure: every refusal this
// verb can raise (an unknown key, a duplicate, a hand-curated entry, an invalid
// pattern, an absent store) is the caller's input to fix, not a fault.
func usageError(verb string, err error) error {
	return &exitError{Code: 2, Msg: verb + ": " + err.Error()}
}

func plural(n int) string {
	if n == 1 {
		return "y"
	}
	return "ies"
}

// renderPrivateLayer prints the private layer by key. It states the layer's reach
// plainly: the guard is local by construction, so a reader must never infer CI
// coverage from a populated list.
func renderPrivateLayer(w io.Writer, rep banlist.PrivateReport) {
	fmt.Fprintf(w, "private layer — %s\n", rep.Path)
	if !rep.Present {
		fmt.Fprintln(w, "  INACTIVE on this machine: the store is absent, so nothing is checked against it.")
		fmt.Fprintln(w, "  opt in with: abcd banlist add --private <key> <pattern>")
	} else {
		for _, e := range rep.Entries {
			fmt.Fprintf(w, "  %s (line %d)\n", termsafe.Sanitize(e.Key), e.Line)
		}
		if len(rep.Entries) == 0 {
			fmt.Fprintln(w, "  no entries")
		}
		fmt.Fprintln(w, "  keys only: the pattern values never reach any output, by design")
		for _, line := range rep.Malformed {
			fmt.Fprintf(w, "  line %d is not a usable pattern — the pre-commit guard refuses until it is fixed\n", line)
		}
	}
	fmt.Fprintln(w, "  reach: CI cannot enforce this layer — it protects only machines that have opted in")
}

// renderPublicLayer prints the public family in full: committed config is
// reviewable config, so its patterns belong on screen.
func renderPublicLayer(w io.Writer, rep banlist.PublicReport) {
	fmt.Fprintf(w, "public layer — %s\n", rep.Path)
	if !rep.Present {
		fmt.Fprintln(w, "  absent: this repo has no docs-lint config, so no public token is gated")
		return
	}
	for _, e := range rep.Entries {
		owner := "hand-curated"
		if e.Managed {
			owner = "verb-managed"
		}
		fmt.Fprintf(w, "  %-32s %-8s %-13s %s\n", termsafe.Sanitize(e.ID), e.Severity, owner, termsafe.Sanitize(e.Pattern))
	}
	if len(rep.Entries) == 0 {
		fmt.Fprintln(w, "  no entries")
	}
	fmt.Fprintln(w, "  reach: enforced deterministically by `abcd docs lint` in CI, with the per-line escape")
}
