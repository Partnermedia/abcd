package capture

import (
	"reflect"
	"testing"
)

// TestInlineListRoundTripQuotedCommas (B24) pins the quote-aware inline-list
// tokenizer: yamlScalar/buildIssueText legally emit quoted items containing
// commas, quotes, and backslashes, and parseFrontmatterAndBody must read them
// back verbatim instead of splitting mid-item on every bare comma.
func TestInlineListRoundTripQuotedCommas(t *testing.T) {
	items := []string{"design review, session 3", `a","b`, `back\slash`, "gamma"}
	text, err := buildIssueText(
		[]kv{{"id", "iss-1"}, {"synthesis_clusters", items}},
		"body\n",
	)
	if err != nil {
		t.Fatalf("buildIssueText: %v", err)
	}
	fm, _, err := parseFrontmatterAndBody(text)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	got, ok := fm["synthesis_clusters"].([]string)
	if !ok {
		t.Fatalf("synthesis_clusters is %T, want []string", fm["synthesis_clusters"])
	}
	if !reflect.DeepEqual(got, items) {
		t.Fatalf("round-trip corrupted the inline list:\n got: %#v\nwant: %#v", got, items)
	}
}

// TestParseScalarOrListSkipsQuotedComma isolates the tokenizer: a comma inside a
// quoted item is not a separator, so a two-item list is not blown apart.
func TestParseScalarOrListSkipsQuotedComma(t *testing.T) {
	v, err := parseScalarOrList(`["alpha, beta", "gamma"]`)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"alpha, beta", "gamma"}
	if !reflect.DeepEqual(v, want) {
		t.Fatalf("got %#v, want %#v", v, want)
	}
}

// TestParseRejectsDuplicateKey pins the duplicate-key guard: a repeated top-level
// key is rejected rather than silently kept last-wins, which would diverge from
// setScalarField's first-occurrence rewrite.
func TestParseRejectsDuplicateKey(t *testing.T) {
	text := "---\nid: iss-1\nseverity: minor\nid: iss-2\n---\nbody\n"
	if _, _, err := parseFrontmatterAndBody(text); err == nil {
		t.Fatal("duplicate top-level key was accepted")
	}
}

// TestParseRejectsDuplicateNestedKey guards the nested-object variant.
func TestParseRejectsDuplicateNestedKey(t *testing.T) {
	text := "---\nid: iss-1\nresolved_by:\n  intent: itd-1\n  intent: itd-2\n---\nbody\n"
	if _, _, err := parseFrontmatterAndBody(text); err == nil {
		t.Fatal("duplicate nested key was accepted")
	}
}

// TestValidateStrictTypeChecksResolvedBy proves a non-string resolved_by
// sub-value is rejected rather than validating clean and then silently dropping
// to "" on read (a lossy, undetected round-trip).
func TestValidateStrictTypeChecksResolvedBy(t *testing.T) {
	fm := map[string]any{
		"schema_version": 1,
		"id":             "iss-1",
		"slug":           "x",
		"severity":       "minor",
		"category":       "bug",
		"source":         "agent-finding",
		"found_during":   "review",
		"resolved_by":    map[string]any{"intent": 42}, // non-string
	}
	if err := validateStrict(fm); err == nil {
		t.Fatal("non-string resolved_by sub-value was accepted")
	}
	// A well-formed string value still validates.
	fm["resolved_by"] = map[string]any{"intent": "itd-1"}
	if err := validateStrict(fm); err != nil {
		t.Fatalf("valid resolved_by rejected: %v", err)
	}
}

// TestValidateStrictImpact pins the issue schema against the impact field
// (spc-10): the ledger's own records carry it, so a reader that rejects it as an
// unknown property silently drops every resolved issue out of `abcd capture`.
// The value is checked against the one shared enum, exactly as severity,
// category, and source are, so a mislabel is caught on read rather than at the
// release cut.
func TestValidateStrictImpact(t *testing.T) {
	base := func() map[string]any {
		return map[string]any{
			"schema_version": 1,
			"id":             "iss-1",
			"slug":           "x",
			"severity":       "minor",
			"category":       "bug",
			"source":         "agent-finding",
			"found_during":   "review",
		}
	}
	for _, v := range []string{"additive", "breaking", "fix", "internal"} {
		fm := base()
		fm["impact"] = v
		if err := validateStrict(fm); err != nil {
			t.Errorf("impact %q rejected: %v", v, err)
		}
	}
	for _, v := range []any{"braking", "Additive", 42} {
		fm := base()
		fm["impact"] = v
		if err := validateStrict(fm); err == nil {
			t.Errorf("invalid impact %#v was accepted", v)
		}
	}
	// The YAML nulls read as ABSENT, not as a malformed value — the same verdict
	// the record-lint blocker issue_impact_valid reaches through
	// frontmatter.IsNull. One value must not be legal to one gate and malformed
	// to the other: `impact: null` on an open issue would otherwise pass the lint
	// and then fail `abcd capture resolve` on the very record it judged fine.
	for _, v := range []string{"", "null", "~"} {
		fm := base()
		fm["impact"] = v
		if err := validateStrict(fm); err != nil {
			t.Errorf("null-ish impact %q rejected: %v", v, err)
		}
	}
	// Absent stays legal: an open issue has not been judged yet, and the
	// record-lint blocker is what gates the move into resolved/.
	if err := validateStrict(base()); err != nil {
		t.Fatalf("absent impact rejected: %v", err)
	}
}
