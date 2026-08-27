package intent

import (
	"fmt"

	"github.com/intentdriven/abcd/internal/adapter/scanner"
)

// redactIntentText sanitises caller-supplied free text bound for a committed
// draft intent through the ONE canonical detector — the same scanner the
// transcript store, the capture ledger and the launch bundler use. It is the
// intent store's equivalent of capture.redactLedgerText and the write-time
// sanitiser stage of history.Capture.
//
// A draft intent is committed, durable material minted from free text: the
// quoted-text create path (CreateFromText) hands the caller's prose straight to
// title and body, and nothing upstream constrains what it holds. Before this
// ran, `abcd intent "<text>"` wrote a secret token or an absolute home path into
// .abcd/development/intents/ verbatim, with every lint gate green — the exact
// leak capture already closed for the issue ledger (gh-486; sibling of
// iss-2608231025198888).
//
// Discipline (mirrors the two committed siblings):
//   - FAIL CLOSED on a degraded scanner. A scanner that cannot be constructed,
//     or whose per-repo pii.json weakened the pattern set, returns an error and
//     the caller writes nothing — history.Capture's stance, not capture's
//     redact-and-report one, because a draft intent is durable committed prose
//     and a broken detector must never let an unredacted secret reach it under a
//     false "clean" signal (the fail-open this exists to close).
//   - REDACT the surviving spans and return the count, so the caller can say
//     what happened rather than rewriting in silence.
//
// It reuses scanner.Redact — the single masking primitive — so no second
// scanner or masking rule is introduced.
func redactIntentText(repoRoot, text string) (redacted string, count int, err error) {
	sc, err := scanner.New(repoRoot)
	if err != nil {
		return "", 0, fmt.Errorf("intent: refusing to persist text with an unavailable scanner: %w", err)
	}
	if unavail, reason := sc.Unavailable(); unavail {
		return "", 0, fmt.Errorf("intent: refusing to persist text with a degraded scanner: %s", reason)
	}
	findings := sc.ScanText(text, "intent")
	if len(findings) == 0 {
		return text, 0, nil
	}
	out, n := scanner.Redact(text, findings)
	return out, n, nil
}
