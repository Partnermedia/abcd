package capture

import (
	"fmt"
	"strings"
	"testing"

	"github.com/intentdriven/abcd/internal/core/frontmatter"
)

// buildRecord assembles a ledger-shaped issue record with caller-chosen
// delimiter lines, so a test can vary ONLY the delimiter and hold the block
// bytes fixed.
func buildRecord(openDelim, closeDelim string) string {
	return openDelim + "\n" +
		"id: 7\n" +
		"kind: issue\n" +
		"status: open\n" +
		"impact: minor\n" +
		closeDelim + "\n" +
		"\n" +
		"# Sample defect\n" +
		"\n" +
		"Body prose.\n"
}

// TestDelimiterToleranceMatchesCanonicalScanner (GitHub #338) pins ONE delimiter
// contract across the two readers of a ledger record's bytes: capture's parser
// and the canonical whitespace-tolerant scanner in internal/core/frontmatter,
// which record-lint's ledger gate and the lifeboat graveyard both read through.
//
// A `--- ` (trailing-space) delimiter used to yield a lint-green issue file that
// every capture verb refused with "frontmatter not terminated", and that
// `abcd iss-N` reported as "not found in the issue ledger" while the file sat in
// open/ — CWE-436, two parsers disagreeing about one format with the permissive
// one on the gate side. No abcd writer emits the shape; hand-editing and
// `abcd embark`'s byte-for-byte foreign records do.
func TestDelimiterToleranceMatchesCanonicalScanner(t *testing.T) {
	cases := []struct {
		name  string
		open  string
		close string
	}{
		{"canonical", "---", "---"},
		{"trailing space on close", "---", "--- "},
		{"trailing space on open", "--- ", "---"},
		{"trailing space on both", "--- ", "--- "},
		{"trailing tab on close", "---", "---\t"},
		{"trailing spaces and tabs", "---  \t", "---\t "},
		// A BOM sits invisibly ahead of the opening delimiter. The canonical
		// scanner trims it (TrimBOM) so every frontmatter-keyed gate still reads
		// the record; capture refused the same bytes outright. Same class, same
		// rule, so it comes along with the shared predicate.
		{"BOM before open", "\ufeff---", "---"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			text := buildRecord(tc.open, tc.close)

			fm, body, err := parseFrontmatterAndBody(text)
			if err != nil {
				t.Fatalf("capture refused a record the canonical scanner reads: %v", err)
			}

			// Both readers must agree on the block's keys.
			canonical := frontmatter.Fields(strings.Split(text, "\n"))
			for key, field := range canonical {
				got, ok := fm[key]
				if !ok {
					t.Fatalf("canonical scanner read %q=%q; capture read no such key", key, field.Value)
				}
				if gotStr := scalarString(got); gotStr != field.Value {
					t.Fatalf("key %q: canonical=%q capture=%q", key, field.Value, gotStr)
				}
			}
			if len(fm) != len(canonical) {
				t.Fatalf("key sets diverge: capture=%v canonical=%v", fm, canonical)
			}

			// The tolerant close must also be the body boundary: block lines must
			// not leak into the body, and body prose must not leak in as fields.
			if strings.Contains(body, "impact:") {
				t.Fatalf("frontmatter leaked into body: %q", body)
			}
			if want := "# Sample defect\n\nBody prose.\n"; body != want {
				t.Fatalf("body = %q, want %q", body, want)
			}
		})
	}
}

// TestDelimiterStrictnessPreserved pins what the tolerance must NOT widen:
// tolerance is about trailing whitespace on the delimiter line, not about what
// counts as a delimiter. A fat-fingered `----`, a delimiter with trailing
// content, an indented `---`, and a missing close each stay refused — the
// canonical scanner rejects them too, so the two readers still agree.
func TestDelimiterStrictnessPreserved(t *testing.T) {
	cases := []struct {
		name string
		text string
	}{
		{"four dashes open", buildRecord("----", "---")},
		{"four dashes close", buildRecord("---", "----")},
		{"open carries content", buildRecord("--- yaml", "---")},
		{"close carries content", buildRecord("---", "--- yaml")},
		{"indented close is not a close", "---\nid: 7\n  ---\nbody\n"},
		{"no close at all", "---\nid: 7\n"},
		{"leading blank line before open", "\n---\nid: 7\n---\n"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if _, _, err := parseFrontmatterAndBody(tc.text); err == nil {
				t.Fatalf("parseFrontmatterAndBody accepted %q; want ErrMalformedFrontmatter", tc.text)
			}
			// The canonical scanner must reach the same verdict: no fields.
			if got := frontmatter.Fields(strings.Split(tc.text, "\n")); len(got) != 0 {
				t.Fatalf("canonical scanner read %v from %q; capture refused it", got, tc.text)
			}
		})
	}
}

// TestDelimiterToleranceSurvivesRewrite pins the writer half: a record whose
// delimiters carry trailing whitespace must be mutable by the status-transition
// rewrites, not just readable. A reader that accepts bytes the writer refuses is
// the same split verdict in the other direction.
func TestDelimiterToleranceSurvivesRewrite(t *testing.T) {
	text := buildRecord("---", "--- ")

	out, err := setScalarField(text, "status", "resolved")
	if err != nil {
		t.Fatalf("setScalarField refused a tolerant-delimiter record: %v", err)
	}
	fm, _, err := parseFrontmatterAndBody(out)
	if err != nil {
		t.Fatalf("rewritten record no longer parses: %v", err)
	}
	if got := scalarString(fm["status"]); got != "resolved" {
		t.Fatalf("status = %q, want %q", got, "resolved")
	}

	out2, err := setMapField(out, "resolved_by", []kv{{key: "commit", val: "deadbeef"}})
	if err != nil {
		t.Fatalf("setMapField refused a tolerant-delimiter record: %v", err)
	}
	fm2, _, err := parseFrontmatterAndBody(out2)
	if err != nil {
		t.Fatalf("record with inserted nested block no longer parses: %v", err)
	}
	sub, ok := fm2["resolved_by"].(map[string]any)
	if !ok {
		t.Fatalf("resolved_by = %#v, want a nested map", fm2["resolved_by"])
	}
	if got := scalarString(sub["commit"]); got != "deadbeef" {
		t.Fatalf("resolved_by.commit = %q, want %q", got, "deadbeef")
	}
}

// scalarString renders a parsed frontmatter value the way the canonical scanner
// reports it (raw scalar text), for cross-reader comparison.
func scalarString(v any) string {
	switch t := v.(type) {
	case string:
		return t
	case nil:
		return ""
	default:
		return fmt.Sprint(t)
	}
}
