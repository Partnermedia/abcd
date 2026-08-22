package site

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestManifestRefusesUnknownAndUnusableKeys pins the manifest's two promises:
// a key the binary does not know is refused, and a key it knows but does not
// act on YET is still validated. The second is the one that rots quietly — a
// path typo under a deferred key would otherwise sit in the file looking
// correct until the slice that consumes it lands, and then fail in a change
// that did not cause it.
func TestManifestRefusesUnknownAndUnusableKeys(t *testing.T) {
	f := newFixture(t)
	manifest := filepath.Join(f.Root(), ".abcd", "site.json")
	original, err := os.ReadFile(manifest)
	if err != nil {
		t.Fatal(err)
	}

	cases := []struct{ name, from, to, says string }{
		{"unknown top-level key", `"schema_version": 1,`, `"schema_version": 1, "colour": "blue",`, "colour"},
		{"unknown nested key", `"letter": "a",`, `"letter": "a", "hue": 3,`, "hue"},
		{"wrong schema version", `"schema_version": 1,`, `"schema_version": 2,`, "schema_version"},
		{"unimplemented icon rule", `"icons": "image-before-lead-in"`, `"icons": "image-after-lead-in"`, "icons"},
		{"unimplemented tab arrangement", `"tabs": "left-h2s, then lead-h3s and remaining-h2s as a labelled group"`, `"tabs": "alphabetical"`, "tabs"},
		{"release read from elsewhere", `"from": "CHANGELOG.md"`, `"from": "RELEASES.md"`, "release.from"},
		// A figure on a layout that lifts none would be read and dropped —
		// including its deferred labels-from-page, which would ask the check
		// slice to compare a diagram nothing renders.
		{"figure on a layout that lifts none", `"layout": "prose",`, `"layout": "cards-from-h2",`, "figure"},
		// Deferred keys: no slice reads these yet, and a typo still fails now.
		{"deferred docs path", `"index": "docs/README.md"`, `"index": "/etc/passwd"`, "docs.index"},
		{"deferred contributors path", `"file": "CONTRIBUTING.md"`, `"file": "../outside.md"`, "policy.file"},
		{"deferred contributors heading", `"heading": "Attribution",`, `"heading": "",`, "policy.heading"},
		{"deferred baseline path", `"unresolved_reference_baseline": ".abcd/site-baseline.json"`, `"unresolved_reference_baseline": "../x.json"`, "baseline"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			body := strings.Replace(string(original), c.from, c.to, 1)
			if body == string(original) {
				t.Fatalf("the fixture manifest does not contain %q", c.from)
			}
			if err := os.WriteFile(manifest, []byte(body), 0o644); err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { os.WriteFile(manifest, original, 0o644) })

			_, err := LoadManifest(f.Root())
			if err == nil {
				t.Fatalf("the manifest accepted %s", c.name)
			}
			if !strings.Contains(err.Error(), c.says) {
				t.Errorf("the refusal does not name %q: %v", c.says, err)
			}
		})
	}
}
