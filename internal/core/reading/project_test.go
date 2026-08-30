package reading

import (
	"strings"
	"testing"
	"time"
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
	got := strings.Join(renderedTexts("See <https://example.invalid/x>"), " ")
	if !strings.Contains(got, "https://example.invalid/x") {
		t.Errorf("the autolink was stripped as a tag: %q", got)
	}
}

// TestAttributeMaskStaysOnItsOwnLine: an attribute value ends on the line it
// opens on. Ending it at the next matching quote found anywhere in the document
// let one unbalanced quote blank every angle bracket up to some unrelated quote
// thousands of bytes later, erasing a raw HTML heading from the masked reading.
//
// The refusal itself no longer turns on this: the heading scan reads the
// unmasked document too and refuses on either reading, so a runaway mask can no
// longer hide a heading end to end. What it can still do is destroy the masked
// reading, which is the reading that sees a `>` written inside an attribute
// value — so the bound is asserted here, on the mask's own contract.
func TestAttributeMaskStaysOnItsOwnLine(t *testing.T) {
	const runaway = "<div id=\"\n\n<h2>Audit Notes</h2>\n\nprose\n\n\">\n"
	got := maskMarkupData(runaway, true)
	if len(got) != len(runaway) {
		t.Fatalf("the mask changed the document length from %d to %d", len(runaway), len(got))
	}
	if !strings.Contains(got, "<h2>Audit Notes</h2>") {
		t.Errorf("the mask crossed the tag its value opened in and erased a heading: %q", got)
	}

	// And it still masks the value it was built for, which closes on its own line.
	const inline = "<h2 title=\"a>b\">Audit Notes</h2>"
	if m := maskMarkupData(inline, true); !strings.Contains(m, "a b") {
		t.Errorf("a greater-than inside a same-line attribute value was left as structure: %q", m)
	}
}

// TestRawHeadingScanStaysLinearInTheOpenerCount: the bound scan materialised
// every candidate bound in the whole remainder of the document for every heading
// opener, though it breaks at the first hard one. That is quadratic with a large
// constant, and a committed markdown file up to the size cap the assembler sets
// did not finish — a silent hang, which is the one staging a fail-closed floor
// cannot afford. The bound is generous: the walk is milliseconds and the
// materialising scan was minutes.
func TestRawHeadingScanStaysLinearInTheOpenerCount(t *testing.T) {
	var b strings.Builder
	b.WriteString("# A spec\n\n")
	for range 8000 {
		b.WriteString("<h2>Ordinary heading</h2>\n")
	}
	doc := b.String()
	headings := map[string]bool{"Audit Notes": true}
	start := time.Now()
	if err := verifyRedaction("spc-x.md", doc, doc, nil, headings); err != nil {
		t.Fatalf("the scan refused an ordinary document: %v", err)
	}
	if elapsed := time.Since(start); elapsed > 10*time.Second {
		t.Errorf("the raw heading scan took %s over %d openers; it materialises the whole "+
			"bound list per opener", elapsed, 8000)
	}
}

// TestMaskStaysLinearInTheAssignmentCount: the line bound searched the whole
// remainder of the document for a newline once per attribute assignment, which
// is quadratic — a 4 MiB file of assignments on one line took 21 seconds where
// the walk takes milliseconds, and the size cap bounds one file rather than how
// many of them a repository holds. The cursor onto the next newline only ever
// advances. Measured: 7 ms with the cursor, 12.3 s without it.
func TestMaskStaysLinearInTheAssignmentCount(t *testing.T) {
	var b strings.Builder
	b.WriteString("# S\n\n<a ")
	for range 800000 {
		b.WriteString("=\"x\"")
	}
	doc := b.String()
	start := time.Now()
	if got := maskMarkupData(doc, true); len(got) != len(doc) {
		t.Fatalf("the mask changed the document length from %d to %d", len(doc), len(got))
	}
	if elapsed := time.Since(start); elapsed > 5*time.Second {
		t.Errorf("the mask took %s over 800000 assignments on one line; it searches the whole "+
			"remainder for a newline per assignment", elapsed)
	}
}
