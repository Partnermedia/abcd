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

// TestGroundsRefusesControlCharacters is iss-2608301206032013: Go's RE2 `\s`
// excludes the vertical tab, so Fold carried U+000B through and the pre-mint
// gate passed a value the frontmatter serialiser then rejected under the ledger
// lock — after the draft had been minted. The argument boundary refuses exactly
// what yamlScalar refuses, so no route reaches the serialiser carrying one.
func TestGroundsRefusesControlCharacters(t *testing.T) {
	for name, text := range map[string]string{
		"vertical tab": "we expect the gate\vto refuse a control character",
		"NUL":          "we expect the gate\x00to refuse a control character",
		"bell":         "we expect the gate\ato refuse a control character",
		"escape":       "we expect the gate\x1bto refuse a control character",
		"tab-adjacent": "we expect the gate to\v refuse a control character",
	} {
		if _, err := New(Pursued, text); err == nil {
			t.Fatalf("%s: New(%q) = nil error, want a refusal", name, text)
		}
		if err := ValidateText(Fold(text)); err == nil {
			t.Fatalf("%s: ValidateText = nil error, want a refusal", name)
		}
	}
	// The whitespace Fold already normalises is not a control-character refusal:
	// a multi-line operand is folded, not rejected, and a control character at
	// either END is trimmed away by Fold before the text is judged — it is not in
	// the value, so there is nothing to refuse.
	for name, text := range map[string]string{
		"folded multi-line":  "we expect a stamped identity\n\tto survive rewording",
		"trimmed at the end": "\vwe expect a stamped identity to survive rewording\v",
	} {
		if _, err := New(Pursued, text); err != nil {
			t.Fatalf("%s: New = %v, want it accepted", name, err)
		}
	}
}

// TestGroundsFloorCountsScriptioContinuaCharacters is iss-2608301301044588: the
// floor was measured in letter-RUNS, and a script written without inter-word
// spaces has exactly one of them however much it says. A substantive Chinese or
// Japanese conjecture was refused as "1 word", and because the reader applies
// the same floor the bullet was then silently dropped — so the readiness gate
// reported no recorded grounds about a record that visibly carried one. The
// argument is mandatory, so that refused an operator writing in those scripts
// from promoting or resolving anything at all.
func TestGroundsFloorCountsScriptioContinuaCharacters(t *testing.T) {
	for name, text := range map[string]string{
		"chinese":  "我们预期带有戳记的身份能够在改写之后继续存活下来",
		"japanese": "スタンプされた識別子は改名されても生き残ると予期している",
		"mixed":    "識別子 ABCD は改名されても生き残ると予期している",
		// Natural prose in most of these scripts is inert as a pin: its combining
		// marks are Mn rather than L, so they broke the OLD letter-runs too and
		// the text was never refused (the issue record calls this "Thai survives
		// by accident"). Measured, prose leaves seven of the nine entries free to
		// be deleted with the suite still green. A MARK-FREE run of a script's
		// letters is one run under the old count and one unit per letter under
		// this one, so each of these fails if its entry leaves the set — which is
		// the only thing that pins the set. They are alphabet runs, not prose:
		// what they assert is the counting, and readability is not the floor's
		// claim.
		"hiragana letter run": "あいうえおかきくけこさしすせそたちつてと",
		"katakana letter run": "アイウエオカキクケコサシスセソタチツテト",
		"thai letter run":     "กขคงจฉชซฌญฎฏฐฑฒณดตถทธน",
		"lao letter run":      "ກຂຄງຈສຊຍດຕຖທນບປຜຝພຟມ",
		"khmer letter run":    "កខគឃងចឆជឈញដឋឌឍណតថទធន",
		"myanmar letter run":  "ကခဂဃငစဆဇဈဉညဋဌဍဎဏတထဒဓ",
		"tibetan letter run":  "ཀཁགངཅཆཇཉཏཐདནཔཕབམཙཚཛཝ",
		"javanese letter run": "ꦲꦤꦕꦫꦏꦢꦠꦱꦮꦭꦥꦝꦗꦪꦚꦩꦒꦧꦛꦔ",
		// The prose cases stay as the shapes an operator would actually type.
		"thai prose":    "เราคาดว่าตัวระบุที่ประทับตราไว้จะอยู่รอดหลังการเขียนใหม่",
		"khmer prose":   "យើងរំពឹងថាអត្តសញ្ញាណដែលមានត្រាបោះនឹងនៅរស់រានក្រោយការសរសេរឡើងវិញ",
		"lao prose":     "ພວກເຮົາຄາດວ່າຕົວລະບຸທີ່ມີຕາປະທັບຈະຢູ່ລອດຫຼັງການຂຽນຄືນໃໝ່",
		"burmese prose": "တံဆိပ်ခတ်ထားသောအမှတ်အသားသည်ပြန်လည်ရေးသားပြီးနောက်ရှင်သန်နေမည်ဟုမျှော်လင့်သည်",
		"tibetan prose": "ཐེལ་ཙེ་བརྒྱབ་པའི་ངོས་འཛིན་དེ་ཡང་བསྐྱར་འབྲི་རྗེས་གནས་ཐུབ་པར་རེ་བ་བྱེད",
	} {
		if _, err := New(Pursued, text); err != nil {
			t.Fatalf("%s: New(%q) = %v, want the conjecture accepted", name, text, err)
		}
		// The reader inherits the same floor, so the bullet is read back rather
		// than silently dropped under the record's `## Grounds` heading.
		if err := ValidateText(Fold(text)); err != nil {
			t.Fatalf("%s: ValidateText(%q) = %v, want it accepted", name, text, err)
		}
	}
}

// TestGroundsScriptioContinuaFloorRefusesPadding pins the half that must NOT
// move: counting one character of a script without word breaks as one unit is
// an admission, and an admission is where a closed hole reopens. A text padded
// with characters that render as emptiness, or with punctuation, carries no
// units however long it is; and appending one ideograph to a single long word
// does not buy the two units it lacks.
func TestGroundsScriptioContinuaFloorRefusesPadding(t *testing.T) {
	for name, text := range map[string]string{
		"zero-width padded ideograph":  strings.Repeat("​", 19) + "中",
		"ideographic punctuation":      strings.Repeat("、。", 10),
		"one long word plus ideograph": "supercalifragilistic 中",
		"fullwidth digits":             "１２３４５６７８９０１２３４５６７８９０",
		// A Unicode SCRIPT table is not a letter table: it carries the script's
		// own digits, punctuation and combining marks too. Counting a script's
		// characters therefore has to count its LETTERS, or the digit-and-dot
		// padding iss-2608301206034359 closed reopens once per script — and the
		// Tibetan tsheg case is twenty dots exactly.
		"thai digits":            strings.Repeat("๐", 20),
		"lao digits":             strings.Repeat("໐", 20),
		"tibetan tsheg":          strings.Repeat("་", 20),
		"khmer khan":             strings.Repeat("។", 20),
		"myanmar section mark":   strings.Repeat("၊", 20),
		"javanese pada":          strings.Repeat("꧈", 20),
		"thai vowel signs":       strings.Repeat("่", 20),
		"thai digits + one word": "supercalifragilistic ๐๐",
		// Korean is written WITH inter-word spaces, so it is judged by the word
		// count like any other spaced script: two words is two words.
		"korean two words": "안녕하십니까여러분들 안녕하십니까여러분들",
	} {
		if _, err := New(Pursued, text); err == nil {
			t.Fatalf("%s: New(%q) = nil error, want a refusal", name, text)
		}
		if err := ValidateText(Fold(text)); err == nil {
			t.Fatalf("%s: ValidateText(%q) = nil error, want a refusal", name, text)
		}
	}
}

// TestGroundsFloorRefusalNamesTheUnitThatRefused: the message a caller is shown
// must be true of the text that was refused. Telling an operator writing a
// script with no word breaks to add words names a remedy their language cannot
// supply, and the message the gate prints is the only instruction they get.
func TestGroundsFloorRefusalNamesTheUnitThatRefused(t *testing.T) {
	_, err := New(Pursued, strings.Repeat("​", 19)+"中")
	if err == nil {
		t.Fatal("New = nil error, want a refusal")
	}
	if strings.Contains(err.Error(), "word floor") || !strings.Contains(err.Error(), "letter(s)") {
		t.Fatalf("refusal of a text without word breaks = %q, want it to ask for letters and not for words", err)
	}
	for _, text := range []string{
		"supercalifragilistic ... 123",
		// A text that carries a word ALONGSIDE an ideograph has words, and can
		// honestly be told to add one — so the character message is wrong here,
		// and would also report a twenty-letter word as one character.
		"supercalifragilistic 中",
	} {
		_, err = New(Pursued, text)
		if err == nil {
			t.Fatalf("New(%q) = nil error, want a refusal", text)
		}
		if !strings.Contains(err.Error(), "word floor") {
			t.Fatalf("refusal of %q = %q, want the word floor named", text, err)
		}
	}
}
