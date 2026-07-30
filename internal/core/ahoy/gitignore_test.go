package ahoy

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestVisibilityEntriesMatchBrief pins the abcd-managed entry set per visibility
// to the brief's visibility table (§1) and to the three-tier layout it describes:
// a private repo commits the whole `.abcd/` namespace and gitignores only the
// local-ephemeral tier; a public repo gitignores the entire `.abcd/` namespace
// (which subsumes the local tier) plus the legacy root-level `memory/` snapshot.
// No entry may name a path the layout does not have — a phantom entry ignores
// nothing while leaving the local tier tracked, which `abcd audit` then flags.
func TestVisibilityEntriesMatchBrief(t *testing.T) {
	want := map[string][]string{
		"private": {".abcd/.work.local/"},
		"public":  {".abcd/", "memory/"},
	}
	if len(visibilityEntries) != len(want) {
		t.Fatalf("visibilityEntries has %d visibilities; want %d", len(visibilityEntries), len(want))
	}
	for vis, wantEntries := range want {
		gotEntries, ok := visibilityEntries[vis]
		if !ok {
			t.Errorf("visibility %q missing from visibilityEntries", vis)
			continue
		}
		if len(gotEntries) != len(wantEntries) {
			t.Errorf("visibility %q entries = %q; want %q", vis, gotEntries, wantEntries)
			continue
		}
		for i := range wantEntries {
			if gotEntries[i] != wantEntries[i] {
				t.Errorf("visibility %q entries = %q; want %q", vis, gotEntries, wantEntries)
				break
			}
		}
	}
	for vis, entries := range visibilityEntries {
		for _, e := range entries {
			if e == ".work/" || e == ".work" {
				t.Errorf("visibility %q ignores phantom root path %q — the layout has no root-level .work/", vis, e)
			}
		}
	}
}

// TestCanonicalGitignoreBlockContent proves the fence body derived from the
// entry set is exactly the canonical block, header included, in a stable order.
func TestCanonicalGitignoreBlockContent(t *testing.T) {
	cases := map[string][]string{
		"private": {gitignoreBegin, gitignoreHeader, ".abcd/.work.local/", gitignoreEnd},
		"public":  {gitignoreBegin, gitignoreHeader, ".abcd/", "memory/", gitignoreEnd},
	}
	for vis, want := range cases {
		got := canonicalGitignoreBlock(vis)
		if strings.Join(got, "\n") != strings.Join(want, "\n") {
			t.Errorf("canonicalGitignoreBlock(%q) =\n%s\nwant\n%s", vis, strings.Join(got, "\n"), strings.Join(want, "\n"))
		}
	}
}

// TestApplyVisibilityBlockHealsPhantomBlock is the iss-169 upgrade shape: a repo
// carrying the old fence (phantom root-level `.work/`, local tier left tracked)
// reads as drift, and one apply replaces it with the canonical block while
// preserving the user's own ignore rules.
func TestApplyVisibilityBlockHealsPhantomBlock(t *testing.T) {
	cases := []struct {
		visibility string
		legacy     []string
		want       map[string]bool
	}{
		{"private", []string{".work/"}, map[string]bool{".abcd/.work.local/": true}},
		{"public", []string{".abcd/", "memory/", ".work/"}, map[string]bool{".abcd/": true, "memory/": true}},
	}
	for _, tc := range cases {
		dir := t.TempDir()
		path := filepath.Join(dir, ".gitignore")
		legacy := "node_modules/\n" + gitignoreBegin + "\n" + gitignoreHeader + "\n" +
			strings.Join(tc.legacy, "\n") + "\n" + gitignoreEnd + "\n"
		if err := os.WriteFile(path, []byte(legacy), 0o644); err != nil {
			t.Fatal(err)
		}
		if !gitignoreBlockDrifts(dir, tc.visibility) {
			t.Errorf("%s: legacy phantom block reads as no drift; want drift", tc.visibility)
		}
		wrote, err := applyVisibilityBlock(dir, tc.visibility)
		if err != nil {
			t.Fatalf("%s: applyVisibilityBlock: %v", tc.visibility, err)
		}
		if !wrote {
			t.Errorf("%s: applyVisibilityBlock wrote = false; want true (block must heal)", tc.visibility)
		}
		raw, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		inner, count := extractGitignoreBlock(raw)
		if count != 1 {
			t.Fatalf("%s: %d abcd blocks after apply; want exactly 1", tc.visibility, count)
		}
		got := parseGitignoreEntries(inner)
		if len(got) != len(tc.want) {
			t.Errorf("%s: healed entries = %v; want %v", tc.visibility, got, tc.want)
		}
		for e := range tc.want {
			if !got[e] {
				t.Errorf("%s: healed block missing %q; got %v", tc.visibility, e, got)
			}
		}
		if got[".work/"] {
			t.Errorf("%s: healed block still carries phantom .work/", tc.visibility)
		}
		if !strings.Contains(string(raw), "node_modules/") {
			t.Errorf("%s: apply dropped the user's own ignore rule; got %q", tc.visibility, raw)
		}
		if gitignoreBlockDrifts(dir, tc.visibility) {
			t.Errorf("%s: drift still reported after apply", tc.visibility)
		}
	}
}

// TestRemoveGitignoreBlocksBalanced proves a matched BEGIN..END span is removed
// in full while surrounding user content is preserved.
func TestRemoveGitignoreBlocksBalanced(t *testing.T) {
	in := "node_modules/\n# BEGIN ABCD\n.abcd/.work.local/\n# END ABCD\nbuild/\n"
	got := removeGitignoreBlocks(in, "\n")
	want := "node_modules/\nbuild/\n"
	if got != want {
		t.Errorf("balanced block not cleanly removed:\n got %q\nwant %q", got, want)
	}
}

// TestRemoveGitignoreBlocksUnbalancedPreservesUserContent is the data-loss
// regression: an orphan BEGIN with no matching END must not delete every line to
// EOF. Only the orphan BEGIN marker is dropped; the user's own ignore rules after
// it survive.
func TestRemoveGitignoreBlocksUnbalancedPreservesUserContent(t *testing.T) {
	in := "node_modules/\n# BEGIN ABCD\n.abcd/.work.local/\n# comment\nsecrets.env\nbuild/\n"
	got := removeGitignoreBlocks(in, "\n")
	for _, must := range []string{"node_modules/", "secrets.env", "build/", ".abcd/.work.local/", "# comment"} {
		if !strings.Contains(got, must) {
			t.Errorf("orphan-BEGIN rewrite deleted user content %q; got %q", must, got)
		}
	}
	if strings.Contains(got, "# BEGIN ABCD") {
		t.Errorf("orphan BEGIN marker should be dropped; got %q", got)
	}
}
