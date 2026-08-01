package memory

import (
	"reflect"
	"strings"
	"testing"
)

// yaml_boundary_test.go — iss-30 instance 6: the two frontmatter-parser surfaces
// the detector named and nothing exercised, block scalars and double-quoted
// escapes. Both are driven directly with literal documents; no fixture files.

// ---------------------------------------------------------------------------
// Block scalars — only the literal "|" indicator is supported
// ---------------------------------------------------------------------------

// TestParseBlockScalar covers the block-scalar shape the parser actually
// supports: a literal "|" whose body is taken verbatim, de-indented by the first
// content line's indent, with interior blank lines kept and trailing ones
// dropped. The block ends at the first line indented less than the block.
func TestParseBlockScalar(t *testing.T) {
	cases := []struct {
		name  string
		lines []string
		want  map[string]any
	}{
		{
			name:  "multi_line_block_then_next_key",
			lines: []string{"body: |", "  line one", "  line two", "after: tail"},
			want:  map[string]any{"body": "line one\nline two", "after": "tail"},
		},
		{
			name:  "interior_blank_line_is_kept",
			lines: []string{"body: |", "  para one", "", "  para two", "after: tail"},
			want:  map[string]any{"body": "para one\n\npara two", "after": "tail"},
		},
		{
			name:  "leading_blank_line_is_kept",
			lines: []string{"body: |", "", "  para one", "after: tail"},
			want:  map[string]any{"body": "\npara one", "after": "tail"},
		},
		{
			name:  "trailing_blank_lines_are_dropped",
			lines: []string{"body: |", "  only", "", "   ", "after: tail"},
			want:  map[string]any{"body": "only", "after": "tail"},
		},
		{
			name:  "indent_relative_to_the_block_is_preserved",
			lines: []string{"body: |", "  - item", "      deeper", "after: tail"},
			want:  map[string]any{"body": "- item\n    deeper", "after": "tail"},
		},
		{
			// The whole point of a block scalar: its content is opaque text, not
			// YAML. Keys, dashes and hashes inside it must survive verbatim.
			name:  "content_is_not_reparsed_as_yaml",
			lines: []string{"body: |", "  key: value", "  - item", "  # not a comment", "  \"quoted\": [a, b]", "after: tail"},
			want:  map[string]any{"body": "key: value\n- item\n# not a comment\n\"quoted\": [a, b]", "after": "tail"},
		},
		{
			name:  "block_is_the_last_key_in_the_document",
			lines: []string{"title: Rotation", "body: |", "  final line", "  and another"},
			want:  map[string]any{"title": "Rotation", "body": "final line\nand another"},
		},
		{
			name:  "empty_block_at_end_of_document",
			lines: []string{"body: |"},
			want:  map[string]any{"body": ""},
		},
		{
			// A "|" with no INDENTED body opens an EMPTY block. The unindented line
			// that follows is a top-level key of its own, not block content: were it
			// read as content the block would go on consuming every later line, and
			// the rest of the document would vanish with no parse error.
			name:  "empty_block_does_not_swallow_the_next_key",
			lines: []string{"body: |", "after: tail"},
			want:  map[string]any{"body": "", "after": "tail"},
		},
		{
			name:  "empty_block_then_blank_line_does_not_swallow_the_next_key",
			lines: []string{"body: |", "", "after: tail"},
			want:  map[string]any{"body": "", "after": "tail"},
		},
		{
			name:  "block_followed_only_by_blank_lines",
			lines: []string{"body: |", "  text", "", ""},
			want:  map[string]any{"body": "text"},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := parseYAMLLines(tc.lines)
			if err != nil {
				t.Fatalf("parse: %v\nlines: %q", err, tc.lines)
			}
			if !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("parsed = %#v, want %#v", got, tc.want)
			}
		})
	}
}

// TestParseBlockScalarThroughFrontmatter proves the block-scalar path is reached
// through the real entry point, delimiters and all. The terminator check is an
// EXACT column-0 "---" match, so what this pins is that an indented occurrence
// inside a block body is ordinary content and only the column-0 line ends the
// region — the indented one was never a candidate terminator to begin with.
func TestParseBlockScalarThroughFrontmatter(t *testing.T) {
	doc := "---\n" +
		"title: Rotation\n" +
		"body: |\n" +
		"  first line\n" +
		"\n" +
		"  --- not a delimiter\n" +
		"licence: MIT\n" +
		"---\n" +
		"page body\n"
	fm, err := parseFrontmatter(doc)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if fm["body"] != "first line\n\n--- not a delimiter" {
		t.Fatalf("body = %q", fm["body"])
	}
	if fm["licence"] != "MIT" {
		t.Fatalf("the key after the block was not parsed: %#v", fm)
	}
	// The same document must split identically at file level.
	region, body, err := splitFileFrontmatter(doc)
	if err != nil {
		t.Fatalf("split: %v", err)
	}
	if body != "page body\n" {
		t.Fatalf("body = %q, want the text after the closing delimiter", body)
	}
	if !strings.Contains(region, "  --- not a delimiter\n") {
		t.Fatalf("region lost the indented block line:\n%s", region)
	}
}

// TestParseBlockScalarEmptyBlockKeepsLaterKeys is the document-level form of the
// same property: a "|" whose body was left empty (a hand-authored page, or a
// distiller that emitted the indicator and nothing under it) must not swallow the
// keys beneath it. The load-bearing key here is `source` — a page whose source
// block silently disappeared reads back with no source hashes at all, which is
// what the reconcile/repair path uses to decide a page is orphaned.
func TestParseBlockScalarEmptyBlockKeepsLaterKeys(t *testing.T) {
	doc := "---\n" +
		"type: topic\n" +
		"notes: |\n" +
		"domain: auth\n" +
		"slug: tokens\n" +
		"source: { class: external_article, source_hash: abc123 }\n" +
		"---\n" +
		"\nbody\n"
	fm, err := parseFrontmatter(doc)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if fm["notes"] != "" {
		t.Fatalf("notes = %#v, want an EMPTY block — the unindented lines below it are not its body", fm["notes"])
	}
	for _, k := range []string{"domain", "slug", "source"} {
		if _, ok := fm[k]; !ok {
			t.Fatalf("key %q was swallowed by the empty block scalar: %#v", k, fm)
		}
	}
	if fm["domain"] != "auth" || fm["slug"] != "tokens" {
		t.Fatalf("keys after the empty block parsed wrongly: %#v", fm)
	}
	src, ok := fm["source"].(map[string]any)
	if !ok || src["source_hash"] != "abc123" {
		t.Fatalf("source block = %#v, want the flow map with its source_hash intact", fm["source"])
	}
}

// TestParseBlockScalarUnsupportedIndicators pins the parser's DELIBERATE limit:
// only the literal "|" opens a block. A folded ">" or a chomping "|-"/"|+" is
// read as an ordinary scalar, so the indented lines that follow it are seen as
// top-level content and the document is REFUSED. That is the property worth
// locking — an unsupported indicator fails loudly rather than silently swallowing
// (or silently dropping) the lines beneath it.
func TestParseBlockScalarUnsupportedIndicators(t *testing.T) {
	for _, indicator := range []string{">", ">-", ">+", "|-", "|+"} {
		t.Run(indicator, func(t *testing.T) {
			_, err := parseYAMLLines([]string{"body: " + indicator, "  line one", "  line two"})
			if err == nil {
				t.Fatalf("indicator %q is not supported; a block written with it must be refused, not silently mis-parsed", indicator)
			}
			if !strings.Contains(err.Error(), "zero indent") {
				t.Fatalf("error = %v, want the top-level-indent refusal", err)
			}
			// With no indented body there is nothing to mis-read: the indicator is
			// simply the scalar's text.
			got, err := parseYAMLLines([]string{"body: " + indicator, "after: tail"})
			if err != nil {
				t.Fatalf("parse: %v", err)
			}
			if got["body"] != indicator {
				t.Fatalf("body = %#v, want the indicator read as a plain scalar %q", got["body"], indicator)
			}
		})
	}
}

// TestParseBlockScalarNotSupportedWhenNested proves block scalars are a
// TOP-LEVEL shape only: inside a nested dict the "|" is read as a scalar and the
// indented body is refused as unsupported nesting, rather than silently
// truncating the value.
func TestParseBlockScalarNotSupportedWhenNested(t *testing.T) {
	_, err := parseYAMLLines([]string{"meta:", "  body: |", "    line one"})
	if err == nil {
		t.Fatal("a nested block scalar must be refused, not silently accepted")
	}
	if !strings.Contains(err.Error(), "deeper nesting not supported") {
		t.Fatalf("error = %v, want the deeper-nesting refusal", err)
	}
}

// ---------------------------------------------------------------------------
// Double-quoted escapes — unescapeDoubleQuoted and its two call sites
// ---------------------------------------------------------------------------

// TestUnescapeDoubleQuoted enumerates the escape set the parser supports — the
// exact set the dumper's doubleQuote emits — and the two ways an escape can be
// malformed. Anything outside the set is refused rather than passed through
// half-decoded.
func TestUnescapeDoubleQuoted(t *testing.T) {
	ok := []struct{ in, want string }{
		{`plain text`, "plain text"},
		{``, ""},
		{`a\nb`, "a\nb"},
		{`a\tb`, "a\tb"},
		{`a\rb`, "a\rb"},
		{`a\"b`, `a"b`},
		{`a\\b`, `a\b`},
		{`\n\t\r\"\\`, "\n\t\r\"\\"},
		{`\\\\`, `\\`},
		{`ends with escaped backslash\\`, `ends with escaped backslash\`},
		{`\\n`, `\n`}, // an escaped backslash followed by a literal n, NOT a newline
	}
	for _, tc := range ok {
		got, err := unescapeDoubleQuoted(tc.in)
		if err != nil {
			t.Fatalf("unescape %q: %v", tc.in, err)
		}
		if got != tc.want {
			t.Fatalf("unescape %q = %q, want %q", tc.in, got, tc.want)
		}
	}

	bad := []struct{ in, wantErr string }{
		{`a\`, "dangling backslash"},
		{`\`, "dangling backslash"},
		{`\\\`, "dangling backslash"},
		{`a\qb`, `unsupported escape sequence '\q'`},
		{`\u0041`, `unsupported escape sequence '\u'`},
		{`\/`, `unsupported escape sequence '\/'`},
		{`\0`, `unsupported escape sequence '\0'`},
		{`it\'s`, `unsupported escape sequence '\''`},
	}
	for _, tc := range bad {
		got, err := unescapeDoubleQuoted(tc.in)
		if err == nil {
			t.Fatalf("unescape %q must fail, got %q", tc.in, got)
		}
		if !strings.Contains(err.Error(), tc.wantErr) {
			t.Fatalf("unescape %q error = %v, want it to contain %q", tc.in, err, tc.wantErr)
		}
	}
}

// TestParseScalarQuoting drives the top-level scalar call site: a double-quoted
// value is unescaped, its errors surface with the offending text, and quoting
// suppresses the type coercion a bare token would get.
func TestParseScalarQuoting(t *testing.T) {
	ok := []struct {
		line string
		want any
	}{
		{`k: "a\nb"`, "a\nb"},
		{`k: "tab\tseparated"`, "tab\tseparated"},
		{`k: "quote \" inside"`, `quote " inside`},
		{`k: "back\\slash"`, `back\slash`},
		{`k: ""`, ""},
		// Quoting suppresses coercion: these stay strings, unlike their bare forms.
		{`k: "42"`, "42"},
		{`k: "true"`, "true"},
		{`k: "null"`, "null"},
		// A BARE scalar is never unescaped — the backslash-n is two characters.
		{`k: a\nb`, `a\nb`},
		{`k: 42`, 42},
		{`k: true`, true},
		{`k: 'it''s'`, "it's"},
	}
	for _, tc := range ok {
		got, err := parseYAMLLines([]string{tc.line})
		if err != nil {
			t.Fatalf("parse %q: %v", tc.line, err)
		}
		if !reflect.DeepEqual(got["k"], tc.want) {
			t.Fatalf("parse %q = %#v, want %#v", tc.line, got["k"], tc.want)
		}
	}

	bad := []struct{ line, wantErr string }{
		{`k: "unterminated`, "unterminated double-quoted string"},
		{`k: "`, "unterminated double-quoted string"},
		// The closing quote is ESCAPED, so the string is not terminated either —
		// the trailing-quote shape must not be mistaken for a closed string.
		{`k: "trailing escape\"`, "dangling backslash"},
		{`k: "a\qb"`, "unsupported escape sequence"},
		{`k: 'unterminated`, "unterminated single-quoted string"},
	}
	for _, tc := range bad {
		got, err := parseYAMLLines([]string{tc.line})
		if err == nil {
			t.Fatalf("parse %q must fail, got %#v", tc.line, got)
		}
		if !strings.Contains(err.Error(), tc.wantErr) {
			t.Fatalf("parse %q error = %v, want it to contain %q", tc.line, err, tc.wantErr)
		}
	}
}

// TestParseFlowEscapes drives the other call site: keys and values inside an
// inline flow map or list. The load-bearing case is a key holding an ESCAPED
// quote followed by a colon — the flow splitter and the key/value colon scan
// must both honour the escape, or the pair is split in the wrong place.
func TestParseFlowEscapes(t *testing.T) {
	cases := []struct {
		name string
		line string
		want any
	}{
		{
			name: "escaped_newline_in_key",
			line: `k: { "a\nb": v }`,
			want: map[string]any{"a\nb": "v"},
		},
		{
			name: "escaped_quote_then_colon_in_key",
			line: `k: { "a\":b": v }`,
			want: map[string]any{`a":b`: "v"},
		},
		{
			name: "escaped_comma_in_quoted_key",
			line: `k: { "a, b": v, other: w }`,
			want: map[string]any{"a, b": "v", "other": "w"},
		},
		{
			name: "single_quoted_key_with_doubled_quote",
			line: `k: { 'it''s': v }`,
			want: map[string]any{"it's": "v"},
		},
		{
			name: "escapes_in_flow_map_value",
			line: `k: { a: "x\\y", b: "p\nq" }`,
			want: map[string]any{"a": `x\y`, "b": "p\nq"},
		},
		{
			name: "escapes_in_flow_list",
			line: `k: ["a\tb", "c\"d", plain]`,
			want: []any{"a\tb", `c"d`, "plain"},
		},
		{
			name: "brace_inside_a_quoted_value",
			line: `k: { a: "}", b: "{" }`,
			want: map[string]any{"a": "}", "b": "{"},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := parseYAMLLines([]string{tc.line})
			if err != nil {
				t.Fatalf("parse %q: %v", tc.line, err)
			}
			if !reflect.DeepEqual(got["k"], tc.want) {
				t.Fatalf("parse %q = %#v, want %#v", tc.line, got["k"], tc.want)
			}
		})
	}

	bad := []struct{ name, line, wantErr string }{
		{"unsupported_escape_in_key", `k: { "a\qb": v }`, "unsupported escape sequence"},
		{"unsupported_escape_in_value", `k: { a: "b\qc" }`, "unsupported escape sequence"},
		{"quote_not_at_end_of_key", `k: { "a"b: v }`, "unterminated double-quoted key"},
		{"unclosed_quote_in_flow_map", `k: { "abc: v }`, "unterminated double-quoted string in flow sequence"},
		{"unclosed_single_quote_in_flow_map", `k: { 'abc: v }`, "unterminated single-quoted string in flow sequence"},
		// An escaped closing quote leaves the flow token unterminated, so the
		// SPLITTER refuses it before the key parser ever sees a dangling
		// backslash — the two escape scans agree, which is the point.
		{"escaped_closing_quote_in_key", `k: { "ab\": v }`, "unterminated double-quoted string in flow sequence"},
	}
	for _, tc := range bad {
		t.Run(tc.name, func(t *testing.T) {
			got, err := parseYAMLLines([]string{tc.line})
			if err == nil {
				t.Fatalf("parse %q must fail, got %#v", tc.line, got)
			}
			if !strings.Contains(err.Error(), tc.wantErr) {
				t.Fatalf("parse %q error = %v, want it to contain %q", tc.line, err, tc.wantErr)
			}
		})
	}
}

// TestDoubleQuoteEscapeRoundTrip is the closure property between the writer and
// the reader: every string the dumper quotes must come back byte-identical
// through BOTH unescape call sites — as a top-level scalar and as a flow-map key
// and value. A writer that emitted an escape the reader does not support (or a
// reader that dropped one) would break here.
func TestDoubleQuoteEscapeRoundTrip(t *testing.T) {
	subjects := []string{
		"", " leading", "trailing ", "plain text",
		"\n", "\t", "\r", "\\", "\"",
		"line one\nline two", "tab\there", "carriage\rreturn",
		"crlf\r\ninside", "harmless\r---\rleaked", "trailing---\rmore",
		`back\slash`, `a\nb`, `c:\windows\path`,
		`quote"d`, `a":b`, `both "\" and \n`,
		"42", "true", "null", "-dash", "?query", ",", "{", "}", "[", "]", "#hash", "@at",
	}
	for _, s := range subjects {
		dumped := dumpString(s)

		// Call site 1: the top-level scalar path.
		got, err := parseScalar(dumped)
		if err != nil {
			t.Fatalf("scalar round trip of %q (dumped %s): %v", s, dumped, err)
		}
		if got != any(s) {
			t.Fatalf("scalar round trip of %q gave %#v (dumped %s)", s, got, dumped)
		}

		// Call site 2: as a flow-map key AND value, which additionally exercises
		// the flow splitter and the unquoted-colon scan.
		flow, err := dumpFlowMap(map[string]any{s: s})
		if err != nil {
			t.Fatalf("dumpFlowMap %q: %v", s, err)
		}
		parsed, err := parseFlowMap(flow)
		if err != nil {
			t.Fatalf("flow round trip of %q (dumped %s): %v", s, flow, err)
		}
		if len(parsed) != 1 {
			t.Fatalf("flow round trip of %q produced %d pairs (dumped %s)", s, len(parsed), flow)
		}
		v, ok := parsed[s]
		if !ok {
			t.Fatalf("flow round trip lost the key %q: %#v (dumped %s)", s, parsed, flow)
		}
		if v != any(s) {
			t.Fatalf("flow round trip of value %q gave %#v (dumped %s)", s, v, flow)
		}

		// Call site 3: the REAL file boundary — dumpFrontmatter → joinFileFrontmatter
		// → parseFrontmatter, the invariant stated at the top of yaml.go. This is
		// the only call site where an unquoted control character can change the
		// document's LINE STRUCTURE, so the two above cannot stand in for it:
		// parseFrontmatter normalises \r to \n before splitting, so a bare \r the
		// dumper emitted becomes a real line break on read — and a "---" after it
		// terminates the frontmatter early, dropping every later key into the page
		// body. The sentinel key sorts after the subject precisely to catch that.
		region, err := dumpFrontmatter(map[string]any{"subject": s, "zz_sentinel": "kept"})
		if err != nil {
			t.Fatalf("dumpFrontmatter %q: %v", s, err)
		}
		reparsed, err := parseFrontmatter(joinFileFrontmatter(region, "page body\n"))
		if err != nil {
			t.Fatalf("frontmatter round trip of %q failed to parse (region %q): %v", s, region, err)
		}
		if reparsed["subject"] != any(s) {
			t.Fatalf("frontmatter round trip of %q gave %#v (region %q)", s, reparsed["subject"], region)
		}
		if reparsed["zz_sentinel"] != "kept" {
			t.Fatalf("frontmatter round trip of %q lost the key after it: %#v (region %q)", s, reparsed, region)
		}
	}
}
