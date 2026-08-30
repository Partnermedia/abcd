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

// MinTextLen is the substance floor in characters. It refuses the degenerate
// texts — a single word, a shrug, a restated flag name — without pretending to
// judge whether what clears it names a conjecture. It is deliberately low: a
// floor that refused real but terse reasoning would push the caller into padding,
// which records less than a short honest sentence does.
const MinTextLen = 20

var (
	// wordRe picks the letter-runs out of a text for the degeneracy check, so
	// punctuation and digits cannot pad a text past it.
	wordRe = regexp.MustCompile(`[\p{L}]+`)
	// spaceRunRe folds a multi-line operand into the single line both carriers
	// hold: a YAML scalar in the ledger, a Markdown bullet on the intent record.
	spaceRunRe = regexp.MustCompile(`\s+`)
)

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
// its own, then validates both halves through New.
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
	return New(t, text)
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

// New builds a validated Grounds from a token and a free text. The text is
// folded to one line first — both carriers hold a single line — and then put to
// the substance floor, so a text that is only line breaks is refused as the empty
// text it is.
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
// the empty text, the text below the floor, and the text made only of the
// vocabulary's own words or the asking verb's name. Everything else passes: what
// this cannot do is tell a conjecture from a restatement of the decision, and it
// does not claim to.
func ValidateText(text string) error {
	if text == "" {
		return fmt.Errorf("grounds text is empty; name the conjecture being acted on, not the route taken")
	}
	if len([]rune(text)) < MinTextLen {
		return fmt.Errorf(
			"grounds text %q is shorter than the %d-character floor; name the conjecture being acted on, not the route taken",
			text, MinTextLen)
	}
	words := wordRe.FindAllString(strings.ToLower(text), -1)
	onlyDegenerate := len(words) > 0
	for _, w := range words {
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
