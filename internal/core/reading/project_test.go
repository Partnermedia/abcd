package reading

import (
	"strings"
	"testing"
)

// TestRenderEquivalenceIsDeterministic: a determinism instrument cannot have a
// coin-flip refusal. Decoding entities by ranging a Go map made the verdict for
// `Audit&amp;nbsp;Notes` depend on whether `&amp;` happened to be applied before
// `&nbsp;` — the same input, the same repository state, two different answers.
func TestRenderEquivalenceIsDeterministic(t *testing.T) {
	const probe = "Audit&amp;nbsp;Notes"
	first := sameRendering(probe, "Audit Notes")
	for i := range 50 {
		if got := sameRendering(probe, "Audit Notes"); got != first {
			t.Fatalf("call %d gave %v where the first gave %v; the verdict is not a function of the input",
				i, got, first)
		}
	}
}

// TestNumericCharacterReferenceIsRecognised: a numeric or hex reference renders
// as the letter it names, so a title carrying one is the excluded title.
func TestNumericCharacterReferenceIsRecognised(t *testing.T) {
	for _, probe := range []string{"Audit N&#111;tes", "Audit N&#x6f;tes", "&#65;udit Notes"} {
		if !sameRendering(probe, "Audit Notes") {
			t.Errorf("%q does not read as the excluded heading", probe)
		}
	}
}

// TestRenderedTextLeavesAnAutolinkAlone: stripping tags must not eat an autolink,
// which is a URL a heading may legitimately carry.
func TestRenderedTextLeavesAnAutolinkAlone(t *testing.T) {
	got := renderedText("See <https://example.invalid/x>")
	if !strings.Contains(got, "https://example.invalid/x") {
		t.Errorf("the autolink was stripped as a tag: %q", got)
	}
}
