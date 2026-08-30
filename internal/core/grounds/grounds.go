// Package grounds is the recorded-grounds vocabulary as DATA: the closed set of
// three values, the `<token>: <text>` grammar, the parser, the renderer, and the
// substance floor that refuses a degenerate text.
//
// It is a leaf for the same reason core/issueschema and core/changelog are: three
// writers record grounds — the intent record writer (core/intent), the issue
// ledger writer (core/capture), and the committed-record gate that reads them
// back (core/lint) — and a vocabulary spelled three times is a vocabulary the
// three can disagree about, which is how a value one surface writes becomes a
// value another refuses. It imports only the standard library: no filesystem, no
// transport, no record store.
//
// The grounds name the CONJECTURE being acted on, not the route taken. "Planned
// it because it is next" restates the decision; "planned it because we expect a
// stamped identity to survive rewording, which nothing else does" is a
// conjecture somebody can later find wrong. That distinction is a review
// property and no machine reads a sentence and knows which it is, so what lives
// here is a FLOOR — it refuses the degenerate cases and claims nothing more. The
// substantive requirement is carried by the interview prompts on the plugin
// surface.
package grounds

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"
)

// Token is one of the three recorded dispositions a ground may carry. The set is
// closed: a fourth value would be a fourth meaning nothing downstream reads.
type Token string

// The vocabulary (the 2026-08-28 decision-log entry): grounds extend the ADR
// family's decision-granularity reasoning rather than duplicating it, at the
// finer conjecture granularity, in these three values.
const (
	// Pursued records the reasoning behind work that went FORWARD — the half of
	// the record that had no home before and evaporated at the gate.
	Pursued Token = "pursued"
	// Deferred records a conjecture considered and left for later.
	Deferred Token = "deferred"
	// Declined records deliberate non-action: the wontfix's own type.
	Declined Token = "declined"
)

// Vocabulary is the closed set, in the order a surface should offer it. It is
// the ONE copy every gate and every flag description reads.
var Vocabulary = []Token{Pursued, Deferred, Declined}

// Grounds is one recorded ground: the disposition plus the free-text conjecture.
// It is comparable, so a round trip through String and Parse can be asserted
// whole.
type Grounds struct {
	Token Token  `json:"token"`
	Text  string `json:"text"`
}

// MinTextLetters is the substance floor in LETTERS. It refuses the degenerate
// texts — a single word, a shrug, a restated flag name — without pretending to
// judge whether what clears it names a conjecture. It is deliberately low: a
// floor that refused real but terse reasoning would push the caller into padding,
// which records less than a short honest sentence does.
//
// Letters rather than runes, because a rune count is answered by whatever
// occupies a rune and the padding is exactly what says nothing: three one-letter
// words and seventeen dots carry twenty runes and three letters of reasoning,
// and a text of letters that render as nothing carries no reasoning at all
// (iss-2608301455387735). isTextLetter is which runes count.
const MinTextLetters = 20

// MinTextWords is the same floor measured in LEXICAL UNITS, and it is the half
// that does the work. A rune count alone is answerable with nothing at all:
// twenty zero-width spaces, twenty dots or twenty digits each occupy twenty
// runes and carry no letters, and a refusal a caller can satisfy with characters
// that render as emptiness is not a floor (iss-2608301206034359). Three is the lowest count that can carry a subject, a
// verb and an object — the shape of a claim somebody could later find wrong —
// and the corpus's shortest real entry runs to thirty-two words, so nothing
// honest is anywhere near it.
//
// What a unit IS depends on the script, and that is the whole of the fix for
// iss-2608301301044588: a letter-run is a word only where the script separates
// words, and counting runs made a Chinese or Japanese text of any length exactly
// one word. See textUnits for the counting and scriptioContinua for the scripts.
const MinTextWords = 3

// spaceRunRe folds a multi-line operand into the single line both carriers hold:
// a YAML scalar in the ledger, a Markdown bullet on the intent record.
var spaceRunRe = regexp.MustCompile(`\s+`)

// scriptioContinua names the scripts written WITHOUT inter-word spaces, whose
// LETTERS each count as one unit. A letter-run in such a script is the whole
// text however much it says, so counting runs put every one of them at one word
// and — the argument being mandatory — refused an operator writing in them from
// promoting or resolving anything (iss-2608301301044588).
//
// These are Unicode SCRIPT tables, which are not letter tables: six of the nine
// carry their script's own digits, punctuation or combining marks alongside its
// letters. Only isScriptioContinuaLetter's members are counted, and the letter
// test there is load-bearing rather than tidy.
//
// The list is deliberately PARTIAL, and stated rather than implied, because a
// complete-looking list that is wrong is worse than a short one that is honest.
// Han, Hiragana and Katakana are the measured cases; Thai, Lao, Khmer, Myanmar,
// Tibetan and Javanese are here because they share the property, not because any
// record exercised them. Two limits follow. A scriptio-continua script NOT named
// here still reads as one word and is still refused — this list is where that is
// fixed, one script at a time. And a script written WITH spaces stays out even
// where it looks adjacent: Hangul is spaced, so Korean is judged by its words
// like any other spaced script, and counting its syllables would let two Korean
// words clear a three-word floor.
//
// One character is not one word in any of these scripts. The floor does not
// claim it is: MinTextLetters demands twenty letters of any text whatever its
// script, and what this list adds is that those letters are counted one to a
// unit rather than one to a run.
var scriptioContinua = []*unicode.RangeTable{
	unicode.Han, unicode.Hiragana, unicode.Katakana,
	unicode.Thai, unicode.Lao, unicode.Khmer,
	unicode.Myanmar, unicode.Tibetan, unicode.Javanese,
}

// degenerateWords are the words a text may not consist SOLELY of: the vocabulary
// itself, and the names of the verbs that ask for it. A text made only of these
// has restated the route taken and recorded no reasoning at all — the exact
// failure the argument exists to close.
var degenerateWords = map[string]bool{
	"pursued": true, "deferred": true, "declined": true,
	"ready": true, "promote": true, "resolve": true, "wontfix": true,
	"grounds": true,
}

// Parse reads the `<token>: <text>` grammar — the one spelling written to a
// `grounds:` frontmatter scalar and to a `- <token>: <text>` bullet alike. It
// splits on the FIRST colon, because the text is prose and may carry colons of
// its own.
//
// It checks the GRAMMAR and the VOCABULARY, and deliberately not the substance
// floor. Parse is what the readers and the committed-record gate use, and those
// judge values already written: a wontfix stamps its grounds from a reason whose
// own contract is merely non-empty, so a floor here would make a reader skip
// records the ledger has always accepted. The floor is an input gate, applied by
// New at the writing verb over what a caller supplies.
func Parse(s string) (Grounds, error) {
	tok, text, ok := strings.Cut(s, ":")
	if !ok {
		return Grounds{}, fmt.Errorf(
			"grounds %q is not `<token>: <text>` (want one of %s followed by the conjecture)", s, vocabularyList())
	}
	t, err := ParseToken(tok)
	if err != nil {
		return Grounds{}, err
	}
	folded := Fold(text)
	if folded == "" {
		return Grounds{}, fmt.Errorf("grounds %q carries no text after its token", s)
	}
	return Grounds{Token: t, Text: folded}, nil
}

// ParseToken validates one vocabulary value. Surrounding whitespace and case are
// forgiven — the value is stored canonically either way, and refusing `Pursued:`
// would spend a refusal on nothing.
func ParseToken(s string) (Token, error) {
	t := Token(strings.ToLower(strings.TrimSpace(s)))
	for _, v := range Vocabulary {
		if t == v {
			return v, nil
		}
	}
	return "", fmt.Errorf("grounds token %q is not one of %s", strings.TrimSpace(s), vocabularyList())
}

// New builds a validated Grounds from a token and a free text, PUTTING THE TEXT
// TO THE SUBSTANCE FLOOR. It is the writing verb's gate over what a caller
// supplies; Parse is the reader's gate over what is already written. The text is
// folded to one line first — both carriers hold a single line — so a text that
// is only line breaks is refused as the empty text it is.
func New(tok Token, text string) (Grounds, error) {
	t, err := ParseToken(string(tok))
	if err != nil {
		return Grounds{}, err
	}
	folded := Fold(text)
	if err := ValidateText(folded); err != nil {
		return Grounds{}, err
	}
	return Grounds{Token: t, Text: folded}, nil
}

// Fold collapses every run of whitespace to a single space and trims the ends —
// the normalisation both writers need, held here so neither invents its own.
func Fold(text string) string {
	return strings.TrimSpace(spaceRunRe.ReplaceAllString(text, " "))
}

// ValidateText is the substance floor over an already-folded text. It refuses
// the empty text, the text carrying a control character no record field can
// hold, the text below either half of the floor — the UNIT count and the LETTER
// count, neither of which a text of zero-width spaces can satisfy — and the text
// made only of the vocabulary's own words or the asking verb's name. Everything
// else passes: what this cannot do is tell a conjecture from a restatement of
// the decision, and it does not claim to.
//
// A unit is a word where the script separates words and a letter where it does
// not (textUnits, scriptioContinua), so the floor is a property of the TEXT
// rather than of the writing system. The UNIT count is asked FIRST. Every unit
// of a text the scriptio-continua refusal speaks for is a single letter, so such
// a text has as many letters as units: asking the letter count first would refuse
// each of them for its letters and that refusal would never be reached.
//
// Both halves of the feature inherit the floor from here — the writing verbs
// through New, and the record reader through intent.ParseGrounds — which is why
// it belongs at this one place: a floor the writer and the reader disagree about
// is a gate reporting no recorded grounds about a record that visibly carries
// one.
func ValidateText(text string) error {
	if text == "" {
		return fmt.Errorf("grounds text is empty; name the conjecture being acted on, not the route taken")
	}
	// The control-character refusal belongs HERE, at the argument boundary every
	// route crosses, and not at the serialiser that ultimately refuses it. Fold
	// normalises the whitespace Go's `\s` knows, which excludes the vertical tab,
	// so U+000B reached capture's pre-mint gate clean and was refused by
	// yamlScalar afterwards — under the ledger lock, with the draft already
	// minted and an orphan left behind. Refusing what yamlScalar refuses, before
	// anything is written, closes the class for all three writers
	// (iss-2608301206032013).
	for _, r := range text {
		if r < 0x20 {
			return fmt.Errorf(
				"grounds text carries the control character U+%04X, which no record field can hold; "+
					"remove it and restate the conjecture being acted on", r)
		}
	}
	units := textUnits(text)
	// The floor is measured in lexical units, not in runes: a rune count is
	// satisfied by anything that occupies a rune, and the texts worth refusing
	// are precisely the ones that occupy runes and say nothing.
	if len(units) < MinTextWords {
		// The refusal must be true of the text that was refused. Telling somebody
		// writing a script with no word breaks to add words names a remedy their
		// language cannot supply, and this message is the only instruction they
		// get.
		if scriptioContinuaOnly(units) {
			return fmt.Errorf(
				"grounds text %q carries %d letter(s), and the floor asks for %d where the script "+
					"has no word breaks; name the conjecture being acted on, not the route taken",
				text, len(units), MinTextWords)
		}
		return fmt.Errorf(
			"grounds text %q carries %d word(s), below the %d-word floor; name the conjecture being acted on, not the route taken",
			text, len(units), MinTextWords)
	}
	// The letter count is the other half: it refuses the terse-but-lettered
	// shrug ("no time now") that has the three units. It counts the runes
	// textUnits made units out of, so what answers one half answers the other
	// and padding answers neither.
	letters := 0
	for _, r := range text {
		if isTextLetter(r) {
			letters++
		}
	}
	if letters < MinTextLetters {
		return fmt.Errorf(
			"grounds text %q carries %d letter(s), below the %d-letter floor; name the conjecture being acted on, not the route taken",
			text, letters, MinTextLetters)
	}
	onlyDegenerate := true
	for _, w := range units {
		if !degenerateWords[w] {
			onlyDegenerate = false
			break
		}
	}
	if onlyDegenerate {
		return fmt.Errorf(
			"grounds text %q only repeats the vocabulary or the verb's own name; name the conjecture being acted on, not the route taken", text)
	}
	return nil
}

// textUnits splits a text into the units MinTextWords counts: one unit per
// maximal run of letters in a script that separates words, and one unit per
// LETTER in a script that does not. Anything that is not a letter separates
// units and is never a unit itself — in EITHER branch, a script's own digits and
// punctuation included, and the letters that render as nothing with them
// (isTextLetter) — which is what keeps the floor unanswerable by punctuation, by
// digits, or by characters that render as emptiness: a zero-width run beside a
// single ideograph still counts one, an ideographic comma counts nothing, and
// neither does a Thai digit, a Tibetan tsheg or a Hangul filler.
//
// Units are lower-cased, because the degeneracy check reads them back against
// the vocabulary's own words.
func textUnits(text string) []string {
	var out []string
	var run []rune
	flush := func() {
		if len(run) > 0 {
			out = append(out, strings.ToLower(string(run)))
			run = run[:0]
		}
	}
	for _, r := range text {
		switch {
		case isScriptioContinuaLetter(r):
			flush()
			out = append(out, string(r))
		case isTextLetter(r):
			run = append(run, r)
		default:
			flush()
		}
	}
	flush()
	return out
}

// isScriptioContinuaLetter reports whether the rune is a LETTER of a script
// written without word breaks. The letter test is not redundant: a Unicode
// SCRIPT table is not a letter table, and Thai, Lao, Khmer, Myanmar, Tibetan and
// Javanese each carry their own digits, punctuation and combining marks in
// theirs. Counting a script table's members would have made twenty Thai digits,
// or twenty Tibetan tsheg — which are twenty dots — a twenty-unit text, reopening
// the padding class of iss-2608301206034359 once per script.
func isScriptioContinuaLetter(r rune) bool {
	return isTextLetter(r) && unicode.IsOneOf(scriptioContinua, r)
}

// isTextLetter reports whether the rune is a letter that renders as something.
// Four letters are default-ignorable — the Hangul fillers U+115F, U+1160, U+3164
// and U+FFA0 — and a conforming renderer draws them as nothing, so a text built
// from them alone is invisible end to end while answering a letter count
// (iss-2608301455387735). Go's Other_Default_Ignorable_Code_Point is the whole
// test needed: the rest of the default-ignorable set is format characters and
// variation selectors, and none of those is a letter.
//
// This is the ONE place the floor decides what a letter is. Both halves read it
// — textUnits for the split, ValidateText for the count — so a rune that is not
// a unit is not counted a letter either, and no padding answers one half while
// failing the other.
func isTextLetter(r rune) bool {
	return unicode.IsLetter(r) && !unicode.Is(unicode.Other_Default_Ignorable_Code_Point, r)
}

// scriptioContinuaOnly reports whether every unit the floor counted is a
// character of a script written without word breaks. It is what selects the
// refusal message, and it asks for ALL of them rather than any: a text carrying
// a word alongside an ideograph has words, and can be told to add one.
func scriptioContinuaOnly(units []string) bool {
	if len(units) == 0 {
		return false
	}
	for _, u := range units {
		rs := []rune(u)
		if len(rs) != 1 || !isScriptioContinuaLetter(rs[0]) {
			return false
		}
	}
	return true
}

// String renders the canonical `<token>: <text>` value — what a frontmatter
// scalar carries and what Parse reads back.
func (g Grounds) String() string { return string(g.Token) + ": " + g.Text }

// Bullet renders the record form: one top-level Markdown bullet under the
// record's `## Grounds` heading, so the entry reads as prose to a human and as
// data to the gate with no second parser.
func (g Grounds) Bullet() string { return "- " + g.String() }

// vocabularyList renders the closed set for a refusal message.
func vocabularyList() string {
	parts := make([]string, len(Vocabulary))
	for i, v := range Vocabulary {
		parts[i] = string(v)
	}
	return strings.Join(parts, "/")
}
