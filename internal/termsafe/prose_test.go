package termsafe

import (
	"regexp"
	"strings"
	"testing"
)

// TestCleanProseNeutralisesStructure proves the shared prose cleaner does the
// four things every untrusted-prose boundary needs: line breaks cannot forge
// structure, HTML comment markers cannot open or close a comment, terminal
// display attacks are masked, and the result is capped.
func TestCleanProseNeutralisesStructure(t *testing.T) {
	cases := []struct {
		name string
		in   string
		cap  int
		want string
	}{
		{"newline-to-space", "one\ntwo", 100, "one two"},
		{"carriage-return-to-space", "one\rtwo", 100, "one two"},
		{"comment-open-broken", "text <!-- hidden", 100, "text < !-- hidden"},
		{"comment-close-broken", "hidden --> text", 100, "hidden -- > text"},
		{"escape-masked", "red" + string(rune(0x1b)) + "[31m", 100, "red?[31m"},
		{"trimmed", "   padded   ", 100, "padded"},
		{"capped", strings.Repeat("a", 20), 8, "aaaaaaaa"},
		{"cap-trims-trailing-space", "abcdefg hijk", 8, "abcdefg"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := CleanProse(c.in, c.cap); got != c.want {
				t.Errorf("CleanProse(%q, %d) = %q, want %q", c.in, c.cap, got, c.want)
			}
		})
	}
}

// TestCleanProseNeutralisesRawHTML proves no prose field can open raw HTML in the
// record it lands in. In CommonMark a `<` followed by a letter, `/`, `!`, or `?`
// begins an HTML block, and several of those block types run to the end of the
// document — so one unclosed tag can make every later section of a durable record
// render as inert text while a forged section above it renders normally.
func TestCleanProseNeutralisesRawHTML(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
	}{
		{"script-tag", "<script>alert(1)</script>", "< script>alert(1)< /script>"},
		{"table-tag", "<table><tr><td>forged", "< table>< tr>< td>forged"},
		{"declaration", "<!DOCTYPE html>", "< !DOCTYPE html>"},
		{"comment-open", "text <!-- hidden", "text < !-- hidden"},
		{"processing-instruction", "<?php echo", "< ?php echo"},
		{"closing-tag", "</div>", "< /div>"},
		// A masked control byte becomes '?', so a `<` before one would otherwise
		// FORM a processing instruction after sanitisation. Neutralising after
		// Sanitize is what closes this.
		{"escape-forms-instruction", "<" + string(rune(0x1b)) + "php", "< ?php"},
		// Ordinary prose keeps its angle brackets: these open nothing.
		{"comparison-kept", "a < b and c > d", "a < b and c > d"},
		{"arrow-kept", "x <- y", "x <- y"},
		{"trailing-kept", "ends with <", "ends with <"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := CleanProse(c.in, 200); got != c.want {
				t.Errorf("CleanProse(%q) = %q, want %q", c.in, got, c.want)
			}
			if got := CleanProseLine(c.in, 200); strings.Contains(got, "<script") ||
				strings.Contains(got, "<table") || strings.Contains(got, "<!") || strings.Contains(got, "<?") {
				t.Errorf("CleanProseLine(%q) = %q still opens raw HTML", c.in, got)
			}
		})
	}
}

// TestCleanProseKeepsInteriorRuns proves CleanProse leaves interior whitespace
// alone — it trims, it does not collapse. That is the difference from
// CleanProseLine, and a caller picking the wrong one would silently reflow prose.
func TestCleanProseKeepsInteriorRuns(t *testing.T) {
	if got, want := CleanProse("a  b\tc", 100), "a  b c"; got != want {
		t.Errorf("CleanProse kept-runs = %q, want %q", got, want)
	}
}

// TestCleanProseLineCollapsesWhitespace proves the line form collapses every
// whitespace run to one space, so prose landing in a line-structured file (a
// changelog bullet, a markdown table cell) occupies exactly one line.
func TestCleanProseLineCollapsesWhitespace(t *testing.T) {
	cases := []struct {
		name string
		in   string
		cap  int
		want string
	}{
		{"runs-collapsed", "a  b\tc", 100, "a b c"},
		{"newline-collapsed", "one\n\ntwo", 100, "one two"},
		{"comment-open-broken", "text <!-- hidden", 100, "text < !-- hidden"},
		{"capped-after-collapse", "aaaa  bbbb  cccc", 10, "aaaa bbbb"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := CleanProseLine(c.in, c.cap); got != c.want {
				t.Errorf("CleanProseLine(%q, %d) = %q, want %q", c.in, c.cap, got, c.want)
			}
		})
	}
}

// TestCleanProseTruncationStaysValidUTF8 proves a cap landing mid-rune does not
// leave a broken code point behind — the truncated tail is dropped, never emitted
// as replacement bytes.
func TestCleanProseTruncationStaysValidUTF8(t *testing.T) {
	// "é" is two bytes; a 3-byte cap over "aaé" lands mid-rune.
	got := CleanProse("aaé", 3)
	if got != "aa" {
		t.Errorf("CleanProse mid-rune cap = %q, want %q", got, "aa")
	}
	for _, r := range got {
		if r == '�' {
			t.Fatalf("CleanProse emitted a replacement rune: %q", got)
		}
	}
}

// opensRawHTML reports whether s still carries a `<` that CommonMark would read
// as the start of raw HTML outside a code span — the shape every neutralisation
// case below must be free of.
func opensRawHTML(s string) bool {
	return htmlOpenerRe.MatchString(s) || strings.Contains(s, "-->")
}

// TestCleanProseKeepsCodeSpans is iss-2609011217083577: CommonMark never parses
// HTML inside a code span, so the neutralisation buys nothing there and corrupts
// a documented placeholder into a shell redirect (`<path>` → `< path>`) instead.
// With KeepCodeSpans the content of a CLOSED span is left alone; everything
// outside a span keeps every existing neutralisation, and any shape a renderer
// could read differently from the cleaner fails closed.
func TestCleanProseKeepsCodeSpans(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
	}{
		// The record's case: the invocation is a code span, the placeholder survives.
		{"placeholder-in-span", "run `abcd reading ingest --reading-json <path>` first",
			"run `abcd reading ingest --reading-json <path>` first"},
		{"comment-close-in-span", "the `-->` token", "the `-->` token"},
		{"whole-field-is-a-span", "`<script>`", "`<script>`"},
		// CommonMark 6.1: a run of N backticks is closed by the next run of exactly N.
		{"double-backtick-span", "``a ` <b>`` end", "``a ` <b>`` end"},
		{"triple-backtick-span", "``` <script> ```", "``` <script> ```"},
		// An unterminated opener is literal backticks, so what follows is prose.
		{"unterminated-single", "` <script>", "` < script>"},
		{"unterminated-double-then-single", "``a` <b>", "``a` < b>"},
		{"mismatched-lengths", "``<b>`", "``< b>`"},
		// CommonMark example: "`foo``bar``" — the single is literal, the doubles pair.
		{"single-literal-doubles-pair", "`<a>``<b>``", "`< a>``<b>``"},
		// An opener after a closed span is prose again.
		{"opener-after-closed-span", "`x` <script>", "`x` < script>"},
		{"prose-both-sides", "<a> `<b>` <c>", "< a> `<b>` < c>"},
		{"two-spans", "`<a>` and `<b>`", "`<a>` and `<b>`"},
		{"comment-open-after-span", "`x` <!-- hidden", "`x` < !-- hidden"},
		// Fail-closed shapes: a backslash before a backtick is an escape outside a
		// span (so the span the cleaner would see is not the one a renderer sees),
		// and a pipe lets a table row re-split the field around the span. Either
		// disables the exemption for the whole field.
		{"escaped-backtick-disables", "\\`<script>`", "\\`< script>`"},
		{"escaped-backtick-anywhere-disables", "`<a>` \\` `<b>`", "`< a>` \\` `< b>`"},
		{"pipe-disables", "`<a>` | x", "`< a>` | x"},
		{"pipe-inside-span-disables", "`a|<b>`", "`a|< b>`"},
		// Ordinary prose is untouched either way.
		{"comparison-kept", "a < b and `c > d`", "a < b and `c > d`"},
		{"no-backticks", "<table><tr>", "< table>< tr>"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := CleanProse(c.in, 200, KeepCodeSpans)
			if got != c.want {
				t.Errorf("CleanProse(%q, KeepCodeSpans) = %q, want %q", c.in, got, c.want)
			}
			if again := CleanProse(got, 200, KeepCodeSpans); again != got {
				t.Errorf("CleanProse is not idempotent: %q -> %q -> %q", c.in, got, again)
			}
			// The line form honours the option the same way; whitespace collapse is
			// the only thing it adds.
			if got := CleanProseLine(c.in, 200, KeepCodeSpans); got != strings.Join(strings.Fields(c.want), " ") {
				t.Errorf("CleanProseLine(%q, KeepCodeSpans) = %q, want %q", c.in, got, c.want)
			}
		})
	}
}

// TestCleanProseDefaultStillNeutralisesInsideSpans pins that the exemption is
// opt-in: a caller that wraps a cleaned field in its own backticks would re-pair
// the field's spans, so the default has to stay the neutralise-everything form.
func TestCleanProseDefaultStillNeutralisesInsideSpans(t *testing.T) {
	for _, in := range []string{"`<script>`", "run `abcd x <path>`", "``<!-- hidden``"} {
		if got := CleanProse(in, 200); opensRawHTML(got) {
			t.Errorf("CleanProse(%q) = %q, want the span content neutralised by default", in, got)
		}
		if got := CleanProseLine(in, 200); opensRawHTML(got) {
			t.Errorf("CleanProseLine(%q) = %q, want the span content neutralised by default", in, got)
		}
	}
}

// TestCleanProseCapCannotReopenASpan: the cap runs after the neutralisation, so a
// cut landing inside an exempt span leaves an unterminated opener behind — and the
// raw HTML the span was protecting would then be prose. The result must be both
// within the cap and free of any opener.
var closedSingleSpanRe = regexp.MustCompile("`[^`]*`")

func TestCleanProseCapCannotReopenASpan(t *testing.T) {
	in := "`aa <script> b` and `<table>` tail"
	for capBytes := 1; capBytes <= len(in)+2; capBytes++ {
		got := CleanProse(in, capBytes, KeepCodeSpans)
		if len(got) > capBytes {
			t.Errorf("cap %d: len(%q) = %d exceeds the cap", capBytes, got, len(got))
		}
		// Only a still-closed span may carry the raw form: strip every closed
		// single-backtick span (the fixture uses no other run length) and what is
		// left must open nothing.
		if outside := closedSingleSpanRe.ReplaceAllString(got, ""); opensRawHTML(outside) {
			t.Errorf("cap %d: %q left a raw opener outside a closed span", capBytes, got)
		}
		if again := CleanProse(got, capBytes, KeepCodeSpans); again != got {
			t.Errorf("cap %d: not idempotent: %q -> %q", capBytes, got, again)
		}
	}
}
