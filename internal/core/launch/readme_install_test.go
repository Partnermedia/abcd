package launch

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestREADMEDocumentsTheInstallAndUpdatePath is the itd-67 acceptance detector
// for the documented distribution path, run against THIS repository rather than
// a synthetic fixture.
//
// It is pinned to the real tree because every name in the documented commands is
// owned by a different file: the marketplace slug by the module path (and so by
// the git remote), the marketplace and plugin names by
// .claude-plugin/marketplace.json. Prose repeating those names drifts silently —
// a renamed marketplace, or an install line naming a repository that is not this
// one, still reads perfectly while sending users somewhere that does not resolve.
// Deriving each expected string from its owning file makes the README fail with
// the rename instead of after it.
func TestREADMEDocumentsTheInstallAndUpdatePath(t *testing.T) {
	root := repoRootForTest(t)

	readme, err := os.ReadFile(filepath.Join(root, "README.md"))
	if err != nil {
		t.Fatalf("read the committed README: %v", err)
	}
	prose := string(readme)

	slug := moduleRepoSlug(t, root)
	marketplace, plugin := marketplaceNames(t, root)

	cases := []struct {
		name string
		want string
	}{
		{"marketplace add names this repository", "/plugin marketplace add " + slug},
		{"install names the plugin in its marketplace", "/plugin install " + plugin + "@" + marketplace},
		{"update names the plugin", "/plugin update " + plugin},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if !strings.Contains(prose, tc.want) {
				t.Errorf("README does not document %q", tc.want)
			}
		})
	}
}

// moduleRepoSlug derives the "<owner>/<repo>" a marketplace add must name from
// go.mod's module path. go.mod is the one place the repository identity is
// already load-bearing for the build, so it cannot drift unnoticed.
func moduleRepoSlug(t *testing.T, root string) string {
	t.Helper()
	gomod, err := os.ReadFile(filepath.Join(root, "go.mod"))
	if err != nil {
		t.Fatalf("read go.mod: %v", err)
	}
	for _, line := range strings.Split(string(gomod), "\n") {
		path, ok := strings.CutPrefix(strings.TrimSpace(line), "module ")
		if !ok {
			continue
		}
		parts := strings.Split(strings.TrimSpace(path), "/")
		if len(parts) < 3 {
			t.Fatalf("module path %q is not host/owner/repo", path)
		}
		return strings.Join(parts[len(parts)-2:], "/")
	}
	t.Fatal("go.mod declares no module path")
	return ""
}

// marketplaceNames reads the marketplace name and its single plugin name from
// the committed manifest, so the documented install line is checked against the
// names the harness actually resolves.
func marketplaceNames(t *testing.T, root string) (marketplace, plugin string) {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join(root, ".claude-plugin", "marketplace.json"))
	if err != nil {
		t.Fatalf("read the committed marketplace manifest: %v", err)
	}
	var manifest struct {
		Name    string `json:"name"`
		Plugins []struct {
			Name string `json:"name"`
		} `json:"plugins"`
	}
	if err := json.Unmarshal(raw, &manifest); err != nil {
		t.Fatalf("parse the committed marketplace manifest: %v", err)
	}
	if manifest.Name == "" || len(manifest.Plugins) == 0 || manifest.Plugins[0].Name == "" {
		t.Fatalf("marketplace manifest names no marketplace or plugin: %+v", manifest)
	}
	return manifest.Name, manifest.Plugins[0].Name
}
