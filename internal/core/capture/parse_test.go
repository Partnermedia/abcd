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
