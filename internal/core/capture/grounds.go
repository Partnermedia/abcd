package capture

// grounds.go is the ledger's half of the grounds argument (spc-57): the triage
// routes record the conjecture being acted on, in the vocabulary core/grounds
// holds once for every writer.
//
// Promote and resolve REFUSE without it. They mint the value in the same call
// and have no corpus to fix, so there is nothing to stage: a route recorded
// without its reasoning is exactly the evaporation the argument closes. Wontfix
// needs no new required flag — transition already refuses an empty
// wontfix_reason, so a wontfix can never be recorded without grounds — what it
// lacked was the TYPE, which it now stamps as `declined:`.

import (
	"fmt"
	"strings"

	"github.com/intentdriven/abcd/internal/core/grounds"
)

// requireGrounds validates a caller's grounds operand, redacts its text, and
// re-validates what will actually be written.
//
// The order is deliberate and matches the note's: redact BEFORE validating,
// never after, so no rewritten span reaches a value the validator has already
// passed. It REDACTS AND REPORTS rather than refusing on a degraded scanner,
// exactly as redactLedgerText does for the note — a ledger that rejects writes
// is a ledger that stops being written to.
func requireGrounds(repoRoot, verb, raw string) (g grounds.Grounds, redacted int, degraded string, err error) {
	if strings.TrimSpace(raw) == "" {
		return grounds.Grounds{}, 0, "", fmt.Errorf(
			"%s: grounds are required (nothing written) — say why this is being pursued as "+
				"`<pursued|deferred|declined>: <the conjecture being acted on>`, not the route taken", verb)
	}
	parsed, err := grounds.Parse(raw)
	if err != nil {
		return grounds.Grounds{}, 0, "", fmt.Errorf("%s: %w; nothing written", verb, err)
	}
	redText, n, deg := redactLedgerText(repoRoot, parsed.Text)
	validated, err := grounds.New(parsed.Token, redText)
	if err != nil {
		return grounds.Grounds{}, 0, "", fmt.Errorf("%s: %w; nothing written", verb, err)
	}
	return validated, n, deg, nil
}

// wontfixGrounds resolves the grounds a wontfix stamps: `declined: <reason>`
// from the reason it already takes, or the caller's own text when the conjecture
// is worth stating separately from the user-facing reason.
//
// The reason-derived form deliberately SKIPS the substance floor. A wontfix
// reason is already a required, non-empty value with its own contract, and
// putting a new length rule on it here would refuse records the ledger has
// always accepted — a refusal this change was never asked for. The floor governs
// what a caller supplies to the argument, which is where it belongs.
//
// The token is fixed at declined. A non-action is what that value names, so
// `pursued:` on a wontfix would record a route the record's own folder
// contradicts.
func wontfixGrounds(repoRoot, raw, reason string) (g grounds.Grounds, redacted int, degraded string, err error) {
	if strings.TrimSpace(raw) == "" {
		redText, n, deg := redactLedgerText(repoRoot, reason)
		folded := grounds.Fold(redText)
		if folded == "" {
			// transition refuses an empty reason on its own; this only guards the
			// case where redaction consumed the whole of it.
			return grounds.Grounds{}, 0, "", fmt.Errorf("wontfix: the reason is empty after redaction; nothing written")
		}
		return grounds.Grounds{Token: grounds.Declined, Text: folded}, n, deg, nil
	}
	g, redacted, degraded, err = requireGrounds(repoRoot, "wontfix", raw)
	if err != nil {
		return grounds.Grounds{}, 0, "", err
	}
	if g.Token != grounds.Declined {
		return grounds.Grounds{}, 0, "", fmt.Errorf(
			"wontfix: grounds must be `declined: <text>` (a wontfix IS the non-action that value names), got %q; nothing written",
			g.Token)
	}
	return g, redacted, degraded, nil
}

// groundsField renders the frontmatter entry a transition stamps. The value goes
// through yamlScalar (a quoted string), never rawScalar: it is free prose whose
// colons and spaces a bare scalar could not carry.
func groundsField(g grounds.Grounds) kv { return kv{"grounds", g.String()} }
