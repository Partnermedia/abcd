package surface

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// writeManifests lays down a throwaway repo root carrying the two plugin
// manifests, so every manifest test states exactly the JSON it is about.
func writeManifests(t *testing.T, plugin, marketplace string) string {
	t.Helper()
	root := t.TempDir()
	dir := filepath.Join(root, ".claude-plugin")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if plugin != "" {
		if err := os.WriteFile(filepath.Join(dir, "plugin.json"), []byte(plugin), 0o644); err != nil {
			t.Fatalf("write plugin.json: %v", err)
		}
	}
	if marketplace != "" {
		if err := os.WriteFile(filepath.Join(dir, "marketplace.json"), []byte(marketplace), 0o644); err != nil {
			t.Fatalf("write marketplace.json: %v", err)
		}
	}
	return root
}

func keysFor(t *testing.T, entries []ManifestEntry, file string) []string {
	t.Helper()
	var out []string
	for _, e := range entries {
		if strings.HasSuffix(e.File, file) {
			out = append(out, e.Key)
		}
	}
	return out
}

// TestManifestEntriesFlattensDeclaredKeys pins what an "entry" is: one leaf key
// path per declared value, objects flattened with dots, arrays of named objects
// keyed by their name, and any other array recorded as a single leaf so that
// reordering or editing its members is invisible to the guardrail.
func TestManifestEntriesFlattensDeclaredKeys(t *testing.T) {
	root := writeManifests(t,
		`{"name":"abcd","author":{"name":"REPPL","url":"https://example.invalid"},"keywords":["a","b"]}`,
		`{"name":"abcd-marketplace","owner":{"name":"REPPL"},"plugins":[{"name":"abcd","source":"./"}]}`)

	entries, err := ManifestEntries(root)
	if err != nil {
		t.Fatalf("ManifestEntries: %v", err)
	}

	wantPlugin := []string{"author.name", "author.url", "keywords", "name"}
	if got := keysFor(t, entries, "plugin.json"); !equalStrings(got, wantPlugin) {
		t.Fatalf("plugin.json keys = %v, want %v", got, wantPlugin)
	}
	wantMarket := []string{"name", "owner.name", "plugins[abcd].name", "plugins[abcd].source"}
	if got := keysFor(t, entries, "marketplace.json"); !equalStrings(got, wantMarket) {
		t.Fatalf("marketplace.json keys = %v, want %v", got, wantMarket)
	}
}

// TestManifestEntriesArrayKeying covers the array rules one at a time: named
// objects are keyed by name (so reordering the array is not a change), duplicate
// or missing names collapse to one leaf (the keying is ambiguous, so recording
// per-element entries would invent surface), and an empty container still
// registers its own presence.
func TestManifestEntriesArrayKeying(t *testing.T) {
	tests := []struct {
		name   string
		plugin string
		want   []string
	}{
		{
			name:   "named objects keyed by name",
			plugin: `{"items":[{"name":"b","v":1},{"name":"a","v":2}]}`,
			want:   []string{"items[a].name", "items[a].v", "items[b].name", "items[b].v"},
		},
		{
			name:   "duplicate names collapse to one leaf",
			plugin: `{"items":[{"name":"a"},{"name":"a"}]}`,
			want:   []string{"items"},
		},
		{
			name:   "unnamed objects collapse to one leaf",
			plugin: `{"items":[{"v":1}]}`,
			want:   []string{"items"},
		},
		{
			name:   "scalar array is one leaf",
			plugin: `{"items":["a","b"]}`,
			want:   []string{"items"},
		},
		{
			name:   "empty containers still register",
			plugin: `{"items":[],"obj":{}}`,
			want:   []string{"items", "obj"},
		},
		{
			name:   "null is a declared key",
			plugin: `{"a":null}`,
			want:   []string{"a"},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			root := writeManifests(t, tc.plugin, `{"name":"m"}`)
			entries, err := ManifestEntries(root)
			if err != nil {
				t.Fatalf("ManifestEntries: %v", err)
			}
			if got := keysFor(t, entries, "plugin.json"); !equalStrings(got, tc.want) {
				t.Fatalf("keys = %v, want %v", got, tc.want)
			}
		})
	}
}

// TestManifestEntriesTreatsVersionAsOrdinary is the guard for the release
// payload: the development tree carries no `version` in plugin.json and the
// rendered release payload does. Absence must not be an error or a special case,
// and presence must read as one ordinary added entry.
func TestManifestEntriesTreatsVersionAsOrdinary(t *testing.T) {
	without := writeManifests(t, `{"name":"abcd"}`, `{"name":"m"}`)
	entries, err := ManifestEntries(without)
	if err != nil {
		t.Fatalf("ManifestEntries without version: %v", err)
	}
	if got := keysFor(t, entries, "plugin.json"); !equalStrings(got, []string{"name"}) {
		t.Fatalf("keys without version = %v, want [name]", got)
	}

	with := writeManifests(t, `{"name":"abcd","version":"0.4.0"}`, `{"name":"m"}`)
	entries, err = ManifestEntries(with)
	if err != nil {
		t.Fatalf("ManifestEntries with version: %v", err)
	}
	if got := keysFor(t, entries, "plugin.json"); !equalStrings(got, []string{"name", "version"}) {
		t.Fatalf("keys with version = %v, want [name version]", got)
	}
}

// TestManifestEntriesRefusesUnreadableManifests keeps the snapshot fail-closed. A
// missing or malformed manifest must be an error: reporting it as "no entries"
// would make every manifest removal look like a surface that was never declared.
func TestManifestEntriesRefusesUnreadableManifests(t *testing.T) {
	tests := []struct {
		name        string
		plugin      string
		marketplace string
		want        string
	}{
		{"plugin missing", "", `{"name":"m"}`, "plugin.json"},
		{"marketplace missing", `{"name":"p"}`, "", "marketplace.json"},
		{"plugin malformed", `{`, `{"name":"m"}`, "plugin.json"},
		{"plugin not an object", `["a"]`, `{"name":"m"}`, "plugin.json"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			root := writeManifests(t, tc.plugin, tc.marketplace)
			if _, err := ManifestEntries(root); err == nil {
				t.Fatalf("ManifestEntries = nil error, want one naming %s", tc.want)
			} else if !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("error = %q, want it to name %s", err, tc.want)
			}
		})
	}
}

// TestManifestEntriesUsesRepoRelativePaths keeps the artefact machine-independent
// and privacy-safe: an absolute path from the machine that generated it would
// both leak a local path into a committed file and make the drift test fail on
// every other machine.
func TestManifestEntriesUsesRepoRelativePaths(t *testing.T) {
	root := writeManifests(t, `{"name":"p"}`, `{"name":"m"}`)
	entries, err := ManifestEntries(root)
	if err != nil {
		t.Fatalf("ManifestEntries: %v", err)
	}
	for _, e := range entries {
		if !strings.HasPrefix(e.File, ".claude-plugin/") {
			t.Fatalf("entry file %q is not repo-relative", e.File)
		}
	}
}
