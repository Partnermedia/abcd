package launch

import (
	"errors"
	"testing"
)

func TestLoadIncludesValid(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, ".abcd/config/launch-payload.json", `{"includes": ["./commands/", "docs", "docs", "README.md"]}`)
	inc, err := LoadIncludes(root)
	if err != nil {
		t.Fatal(err)
	}
	// Normalised (./ stripped, trailing / dropped) and de-duplicated.
	want := []string{"commands", "docs", "README.md"}
	if len(inc) != len(want) {
		t.Fatalf("got %v want %v", inc, want)
	}
	for i := range want {
		if inc[i] != want[i] {
			t.Errorf("include[%d]=%q want %q", i, inc[i], want[i])
		}
	}
}

func TestLoadIncludesPreflightErrors(t *testing.T) {
	cases := []struct {
		name    string
		content string
	}{
		{"denied_rooted", `{"includes": [".abcd/x"]}`},
		{"denied_glob", `{"includes": [".a*/x"]}`},
		{"absolute", `{"includes": ["/etc/passwd"]}`},
		{"windows_abs", `{"includes": ["C:\\secret"]}`},
		{"dotdot", `{"includes": ["../outside"]}`},
		{"backslash", `{"includes": ["a\\b"]}`},
		{"empty_array", `{"includes": []}`},
		{"empty_string", `{"includes": [""]}`},
		{"not_object", `["commands"]`},
		{"missing_key", `{"other": 1}`},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			root := t.TempDir()
			writeFile(t, root, ".abcd/config/launch-payload.json", c.content)
			_, err := LoadIncludes(root)
			if err == nil {
				t.Fatalf("expected preflight error for %s", c.name)
			}
			var pe *PreflightError
			if !errors.As(err, &pe) {
				t.Errorf("expected *PreflightError, got %T: %v", err, err)
			}
		})
	}
}

func TestLoadIncludesMissingConfig(t *testing.T) {
	root := t.TempDir()
	_, err := LoadIncludes(root)
	if err == nil {
		t.Fatal("expected preflight error for missing config")
	}
}

// TestAbcdCannotBeReincluded proves there is no allowlist path into the bundle
// for .abcd/** — the config line is refused at preflight AND, even if a raw
// include slice somehow carried it, the resolver still excludes it.
func TestAbcdCannotBeReincluded(t *testing.T) {
	root := t.TempDir()
	// Preflight refusal of the config line.
	writeFile(t, root, ".abcd/config/launch-payload.json", `{"includes": [".abcd/x"]}`)
	if _, err := LoadIncludes(root); err == nil {
		t.Error("LoadIncludes must refuse a .abcd-rooted include")
	}
	// Even a hand-forced include slice cannot promote a .abcd path.
	writeFile(t, root, ".abcd/development/brief.md", "SECRET")
	b, err := ResolveBundle(root, []string{".", ".abcd"})
	if err != nil {
		t.Fatal(err)
	}
	for _, f := range b.Included {
		if firstSegment(f.LogicalPath) == ".abcd" {
			t.Errorf(".abcd path was promoted into the bundle: %s", f.LogicalPath)
		}
	}
}
