package capture

import (
	"reflect"
	"testing"
)

func TestBuildIssueTextRoundTrip(t *testing.T) {
	tests := []struct {
		name   string
		fields []kv
		body   string
		want   string
	}{
		{
			name: "scalars only",
			fields: []kv{
				{"schema_version", 1},
				{"id", "iss-7"},
				{"slug", "broken-thing"},
				{"found_during", "manual smoke"},
			},
			body: "The body.\n",
			want: "---\nschema_version: 1\nid: \"iss-7\"\nslug: \"broken-thing\"\nfound_during: \"manual smoke\"\n---\n\nThe body.\n",
		},
		{
			name:   "empty list emits bracket-pair",
			fields: []kv{{"id", "iss-1"}, {"related_intents", []string{}}},
			body:   "",
			want:   "---\nid: \"iss-1\"\nrelated_intents: []\n---\n\n",
		},
		{
			name:   "abcd id list is unquoted inline",
			fields: []kv{{"related_intents", []string{"itd-4", "fn-12", "iss-3"}}},
			body:   "b",
			want:   "---\nrelated_intents: [itd-4, fn-12, iss-3]\n---\n\nb",
		},
		{
			name:   "non-id list is per-item quoted",
			fields: []kv{{"synthesis_clusters", []string{"cluster a", "cluster b"}}},
			body:   "b",
			want:   "---\nsynthesis_clusters: [\"cluster a\", \"cluster b\"]\n---\n\nb",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := buildIssueText(tc.fields, tc.body)
			if err != nil {
				t.Fatalf("buildIssueText: %v", err)
			}
			if got != tc.want {
				t.Fatalf("got:\n%q\nwant:\n%q", got, tc.want)
			}
		})
	}
}

func TestParseFrontmatterAndBody(t *testing.T) {
	text := "---\nschema_version: 1\nid: \"iss-7\"\nrelated_intents: [itd-4, itd-9]\nsynthesis_clusters: [\"a b\"]\n---\n\nHello body\nsecond line\n"
	fm, body, err := parseFrontmatterAndBody(text)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if fm["schema_version"] != 1 {
		t.Errorf("schema_version = %v (%T), want int 1", fm["schema_version"], fm["schema_version"])
	}
	if fm["id"] != "iss-7" {
		t.Errorf("id = %v, want iss-7", fm["id"])
	}
	if !reflect.DeepEqual(fm["related_intents"], []string{"itd-4", "itd-9"}) {
		t.Errorf("related_intents = %#v", fm["related_intents"])
	}
	if !reflect.DeepEqual(fm["synthesis_clusters"], []string{"a b"}) {
		t.Errorf("synthesis_clusters = %#v", fm["synthesis_clusters"])
	}
	if body != "Hello body\nsecond line\n" {
		t.Errorf("body = %q", body)
	}
}

func TestParseRejectsMissingOpener(t *testing.T) {
	if _, _, err := parseFrontmatterAndBody("no frontmatter here"); err == nil {
		t.Fatal("expected error for missing opening ---")
	}
}

func TestYamlScalarRejectsControlChar(t *testing.T) {
	if _, err := yamlScalar("bad\nvalue"); err == nil {
		t.Fatal("expected control-char rejection")
	}
}

func TestYamlScalarEscaping(t *testing.T) {
	got, err := yamlScalar(`he said "hi" \ end`)
	if err != nil {
		t.Fatal(err)
	}
	want := `"he said \"hi\" \\ end"`
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
	// Round-trips through unquote.
	if back := unquote(got[1 : len(got)-1]); back != `he said "hi" \ end` {
		t.Fatalf("unquote = %q", back)
	}
}

func TestSetScalarFieldReplaceAndInsert(t *testing.T) {
	content := "---\nid: \"iss-1\"\ncreated: \"2026-01-01\"\n---\n\nbody\n"
	// Insert a new field before the closing ---.
	got, err := setScalarField(content, "updated", "2026-02-02")
	if err != nil {
		t.Fatal(err)
	}
	want := "---\nid: \"iss-1\"\ncreated: \"2026-01-01\"\nupdated: \"2026-02-02\"\n---\n\nbody\n"
	if got != want {
		t.Fatalf("insert got:\n%q\nwant:\n%q", got, want)
	}
	// Replace an existing field in place.
	got2, err := setScalarField(got, "created", "2026-03-03")
	if err != nil {
		t.Fatal(err)
	}
	if want2 := "---\nid: \"iss-1\"\ncreated: \"2026-03-03\"\nupdated: \"2026-02-02\"\n---\n\nbody\n"; got2 != want2 {
		t.Fatalf("replace got:\n%q\nwant:\n%q", got2, want2)
	}
}
