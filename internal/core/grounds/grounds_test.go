package grounds

import (
	"strings"
	"testing"
)

// conjecture is a text that clears the substance floor, used wherever a test is
// about something other than the text.
const conjecture = "we expect a stamped identity to survive rewording"

// TestGroundsParseVocabulary holds the closed set: the three values parse, and
// everything else is refused with nothing recorded.
func TestGroundsParseVocabulary(t *testing.T) {
	for _, tok := range []Token{Pursued, Deferred, Declined} {
		g, err := Parse(string(tok) + ": " + conjecture)
		if err != nil {
			t.Fatalf("Parse(%q) = %v, want the token accepted", tok, err)
		}
		if g.Token != tok {
			t.Fatalf("Parse(%q) token = %q, want %q", tok, g.Token, tok)
		}
		if g.Text != conjecture {
			t.Fatalf("Parse(%q) text = %q, want %q", tok, g.Text, conjecture)
		}
	}
	for _, bad := range []string{
		"planned: " + conjecture,
		"PURSUED!: " + conjecture,
		"",
		conjecture,
		": " + conjecture,
	} {
		if _, err := Parse(bad); err == nil {
			t.Fatalf("Parse(%q) = nil error, want a refusal", bad)
		}
	}
}

// TestGroundsParseSplitsOnFirstColon: the text is prose and may itself carry a
// colon; only the FIRST colon separates the token from the text.
func TestGroundsParseSplitsOnFirstColon(t *testing.T) {
	const text = "the gate refuses it: nothing else reads the record"
	g, err := Parse("pursued: " + text)
	if err != nil {
		t.Fatal(err)
	}
	if g.Token != Pursued {
		t.Fatalf("token = %q, want pursued", g.Token)
	}
	if g.Text != text {
		t.Fatalf("text = %q, want %q", g.Text, text)
	}
}

// TestGroundsParseChecksGrammarNotTheFloor: Parse is the reader's gate over a
// value already written, so it judges the grammar and the vocabulary and leaves
// the substance floor to New — a reader that applied the floor would skip
// records the ledger has always accepted.
func TestGroundsParseChecksGrammarNotTheFloor(t *testing.T) {
	if _, err := Parse("declined: out of scope"); err != nil {
		t.Fatalf("Parse of a short but well-formed value = %v, want it accepted", err)
	}
	if _, err := Parse("declined:   "); err == nil {
		t.Fatal("Parse of a value with no text = nil error, want a refusal")
	}
}

// TestGroundsRefusesDegenerateText pins the substance floor: the shape check
// that refuses the cases carrying no reasoning at all. It is a floor, not a
// judgement — whether a text really names a conjecture is review's question.
func TestGroundsRefusesDegenerateText(t *testing.T) {
	for name, text := range map[string]string{
		"empty":       "",
		"whitespace":  "   \t  ",
		"below floor": "because",
		"token only":  "pursued",
		"token twice": "pursued pursued pursued pursued",
		"verb only":   "promote promote promote promote",
	} {
		if _, err := New(Pursued, text); err == nil {
			t.Fatalf("%s: New(%q) = nil error, want a refusal", name, text)
		}
	}
	if _, err := New(Pursued, conjecture); err != nil {
		t.Fatalf("New(%q) = %v, want the conjecture accepted", conjecture, err)
	}
}

// TestGroundsRenderRoundTrip: what the renderer writes is what the parser
// reads, so the intent writer, the ledger writer and the lint share one copy of
// the grammar rather than three spellings of it.
func TestGroundsRenderRoundTrip(t *testing.T) {
	g, err := New(Declined, "declined because the ledger already carries the reason")
	if err != nil {
		t.Fatal(err)
	}
	back, err := Parse(g.String())
	if err != nil {
		t.Fatalf("Parse(%q) = %v", g.String(), err)
	}
	if back != g {
		t.Fatalf("round trip = %+v, want %+v", back, g)
	}
	if got, want := g.Bullet(), "- "+g.String(); got != want {
		t.Fatalf("Bullet() = %q, want %q", got, want)
	}
}

// TestGroundsCollapsesWhitespace: the value is written as one YAML scalar and
// one Markdown bullet, so a multi-line operand is folded rather than carried
// into a shape neither surface can hold.
func TestGroundsCollapsesWhitespace(t *testing.T) {
	g, err := Parse("pursued:\n  we expect a stamped identity\n  to survive rewording\n")
	if err != nil {
		t.Fatal(err)
	}
	if g.Text != "we expect a stamped identity to survive rewording" {
		t.Fatalf("text = %q, want the folded single line", g.Text)
	}
}

// TestGroundsFloorCountsLettersNotRunes is iss-2608301206034359: the floor
// measured rune LENGTH, so a text with no letters in it at all cleared it —
// twenty zero-width spaces, twenty dots, twenty digits. The refusal this
// vocabulary was promoted to make is not a floor if it is answerable with
// nothing.
func TestGroundsFloorCountsLettersNotRunes(t *testing.T) {
	for name, text := range map[string]string{
		"zero-width spaces": strings.Repeat("​", 20),
		"dots":              strings.Repeat(".", 20),
		"digits":            "12345678901234567890",
		"one long word":     "internationalisation",
		"two words":         "unrecoverable regressions",
		"padded one word":   "supercalifragilistic ... 123",
	} {
		if _, err := New(Pursued, text); err == nil {
			t.Fatalf("%s: New(%q) = nil error, want a refusal", name, text)
		}
		// The reader holds a hand-typed bullet to the same floor.
		if err := ValidateText(Fold(text)); err == nil {
			t.Fatalf("%s: ValidateText(%q) = nil error, want a refusal", name, text)
		}
	}
	if _, err := New(Pursued, conjecture); err != nil {
		t.Fatalf("New(%q) = %v, want the conjecture accepted", conjecture, err)
	}
}
