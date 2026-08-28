package scanner

// The outbound-artefact scrub: the primitive an autonomous run calls on every
// pull-request body, issue and comment BEFORE it is posted.
//
// abcd opens no pull requests and owns no forge client, and this does not change
// that (spc-45). It is a primitive, not a posting path: the routine holds the
// forge credentials and the routine calls this on the text it is about to send.
// What abcd owns is the definition of a leak, and this is the one door onto it
// for outward-facing text.
//
// Why a scrub rather than a check. Both leak shapes are appended OUTSIDE the
// model's own output, when the artefact is created — so a rule that only refuses
// text the model composed catches nothing, and a forge keeps the pre-edit
// revision of whatever was posted, which makes "post then fix" no remedy at all.
// The routine therefore re-reads what it created and strips, and that is what
// OutboundPolicy makes every autonomous prompt say.

import (
	"fmt"
	"strings"
)

// OutboundPolicy is the policy block an autonomous routine's prompt carries. It
// is a value rather than prose in a document because the prompt assembly, the
// audit rule's fix hint and this package's own contract must say the same thing:
// a policy that exists in three wordings is a policy with three meanings.
const OutboundPolicy = "Never put a live session URL or a tool's own attribution footer " +
	"(\"generated with/by <tool>\") into public text — a pull-request body, an issue, a comment, " +
	"a commit message or a release note. Disclosure goes in the repository's own trailer " +
	"(Assisted-by: <Vendor>:<model-version>) and nowhere else. " +
	"After creating ANY pull request, issue or comment, re-read what was actually created and strip " +
	"any session URL or attribution footer from it: the harness appends them outside your own output, " +
	"so text that left you clean can arrive dirty, and the forge keeps the pre-edit revision of what " +
	"was posted — which is why the strip has to happen immediately, not at review time."

// ScrubOutbound sanitises one outbound artefact and returns the text to post,
// the findings that were removed, and a fail-closed error if anything blocking
// survived.
//
// It is the ScanText → Redact → residual-rescan shape the store-before-commit
// paths already use (history, memory), with one difference: a harness-leak
// finding takes its whole LINE, rather than being masked in place. Masking is
// right for a secret, where the surrounding text is the artefact's own content;
// a footer or a session link IS the whole line the harness added, and leaving a
// masked stub behind would post the shape of the thing the policy bans.
//
// repoRoot supplies the per-repo scanner config, so a repository that raised a
// severity in .abcd/config/pii.json is honoured here exactly as in Stage-1
// redaction.
func ScrubOutbound(repoRoot, text, label string) (string, []Finding, error) {
	sc, err := New(repoRoot)
	if err != nil {
		return "", nil, err
	}

	findings := sc.ScanText(text, label)
	if len(findings) == 0 {
		return text, nil, nil
	}

	// Stage one: drop the lines the harness added whole.
	drop := map[int]bool{}
	for _, f := range findings {
		if IsHarnessLeakKind(f.Kind) {
			drop[f.Line] = true
		}
	}
	stripped := dropLines(text, drop)

	// Stage two: mask whatever else the scan found in what remains. The rescan is
	// over the STRIPPED text so a finding's line number matches the text Redact is
	// about to rewrite — reusing the first scan's numbers would mask the wrong
	// lines once anything has been removed.
	rest := sc.ScanText(stripped, label)
	redacted, _ := Redact(stripped, rest)

	// Stage three, fail closed. Redact is only stage one by its own contract, and
	// a caller that posts what it could not sanitise is worse than one that
	// refuses: the artefact is public the moment it is created.
	for _, f := range sc.ScanText(redacted, label) {
		if f.Severity == SeverityHardFail || IsHarnessLeakKind(f.Kind) {
			return "", findings, fmt.Errorf("outbound artefact %q still carries a %s after redaction; refusing to hand back text to post", label, f.Kind)
		}
	}
	return redacted, findings, nil
}

// dropLines removes the 1-based line numbers in drop from text, preserving the
// trailing-newline shape of the input (a text ending in "\n" still does).
func dropLines(text string, drop map[int]bool) string {
	if len(drop) == 0 {
		return text
	}
	lines := strings.Split(text, "\n")
	kept := make([]string, 0, len(lines))
	for i, line := range lines {
		if drop[i+1] {
			continue
		}
		kept = append(kept, line)
	}
	return strings.Join(kept, "\n")
}
