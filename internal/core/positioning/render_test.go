package positioning

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/intentdriven/abcd/internal/gittest"
)

// snapshot records every file's content and modification time under root.
func snapshot(t *testing.T, root string) map[string]string {
	t.Helper()
	out := map[string]string{}
	err := filepath.Walk(root, func(p string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if fi.IsDir() {
			return nil
		}
		data, rerr := os.ReadFile(p)
		if rerr != nil {
			return rerr
		}
		rel, _ := filepath.Rel(root, p)
		out[rel] = fi.ModTime().String() + "\x00" + string(data)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return out
}

func TestProposeWritesNothing(t *testing.T) {
	root := writeRepo(t, map[string]string{
		"brief.md": canonBrief, "README.md": drift143README,
		"plugin.json": drift143Manifest, "AGENTS.md": drift143Agents,
	})
	before := snapshot(t, root)

	p, err := Propose(root, abcdLikeConfig())
	if err != nil {
		t.Fatalf("Propose: %v", err)
	}
	if len(p.Diffs) == 0 {
		t.Fatal("Propose produced no diffs on a drifted repo — nothing to assert about")
	}

	after := snapshot(t, root)
	if len(before) != len(after) {
		t.Fatalf("Propose changed the file set: %d files before, %d after", len(before), len(after))
	}
	for rel, was := range before {
		if now, ok := after[rel]; !ok || now != was {
			t.Errorf("Propose modified %s — it must write nothing", rel)
		}
	}
}

func TestProposeEmitsAUnifiedDiffPerDriftedSurface(t *testing.T) {
	root := writeRepo(t, map[string]string{
		"brief.md": canonBrief, "README.md": drift143README,
		"plugin.json": cleanManifest, "AGENTS.md": cleanAgents,
	})
	p, err := Propose(root, abcdLikeConfig())
	if err != nil {
		t.Fatalf("Propose: %v", err)
	}
	if len(p.Diffs) != 1 {
		t.Fatalf("got %d diffs, want 1 (only the drifted surface)", len(p.Diffs))
	}
	d := p.Diffs[0]
	if d.SurfaceID != "readme-strapline" || d.File != "README.md" {
		t.Fatalf("diff targets %s/%s", d.SurfaceID, d.File)
	}
	u := d.Unified
	for _, want := range []string{
		"--- a/README.md",
		"+++ b/README.md",
		"@@ ",
		"-  <p>An opinionated, intent-driven development framework for product thinkers.</p>",
		"+  <p>A host-agnostic configuration layer for intent-driven development.</p>",
	} {
		if !strings.Contains(u, want) {
			t.Errorf("unified diff missing %q:\n%s", want, u)
		}
	}
	// The surrounding lines are preserved as context, not rewritten.
	if !strings.Contains(u, "   <h1>Agent-Based Configuration for Development</h1>") {
		t.Errorf("diff carries no context lines:\n%s", u)
	}
}

func TestProposeRewritesAJSONManifestValueEscaped(t *testing.T) {
	root := writeRepo(t, map[string]string{
		"brief.md": canonBrief, "README.md": cleanREADME,
		"plugin.json": drift143Manifest, "AGENTS.md": cleanAgents,
	})
	p, err := Propose(root, abcdLikeConfig())
	if err != nil {
		t.Fatalf("Propose: %v", err)
	}
	if len(p.Diffs) != 1 {
		t.Fatalf("got %d diffs, want 1", len(p.Diffs))
	}
	u := p.Diffs[0].Unified
	if !strings.Contains(u, `+  "description": "A host-agnostic configuration layer for intent-driven development. A single Go binary`) {
		t.Errorf("manifest proposal is not a valid JSON member line:\n%s", u)
	}
	// The proposed manifest must still parse as JSON.
	if !strings.Contains(u, `-  "description": "A host-agnostic configuration layer for development`) {
		t.Errorf("manifest proposal does not remove the drifted value:\n%s", u)
	}
}

func TestProposeIsEmptyOnACleanRepo(t *testing.T) {
	p, err := Propose(cleanRepo(t), abcdLikeConfig())
	if err != nil {
		t.Fatalf("Propose: %v", err)
	}
	if len(p.Diffs) != 0 {
		t.Fatalf("clean repo produced %d diffs:\n%s", len(p.Diffs), p.Unified())
	}
	if p.Unified() != "" {
		t.Errorf("clean repo produced diff text: %q", p.Unified())
	}
}

// An unlocatable surface has no region to replace; the proposal must say so
// rather than inventing a hunk against a line it never found.
func TestProposeSkipsUnlocatableSurfaces(t *testing.T) {
	root := writeRepo(t, map[string]string{
		"brief.md": canonBrief, "README.md": "# Title\n\nnothing locatable\n",
		"plugin.json": cleanManifest, "AGENTS.md": cleanAgents,
	})
	p, err := Propose(root, abcdLikeConfig())
	if err != nil {
		t.Fatalf("Propose: %v", err)
	}
	if len(p.Diffs) != 0 {
		t.Fatalf("unlocatable surface produced a diff:\n%s", p.Unified())
	}
	if p.Report.Drifted() != 1 {
		t.Fatalf("Drifted() = %d, want 1 (the unlocatable surface still reports)", p.Report.Drifted())
	}
}

// The proposal is only useful if it is a real patch. `git apply` is the
// arbiter: a phantom trailing line, a wrong hunk count, or a missing
// no-newline marker all make it refuse.
func TestProposedDiffAppliesWithGitApply(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not on PATH")
	}
	cases := map[string]map[string]string{
		// The drifted line sits near end-of-file, so the trailing context
		// reaches the last line.
		"drift at end of file": {
			"brief.md":  canonBrief,
			"README.md": "# Widget\n\n<p>An opinionated widget framework for busy people.</p>\n",
		},
		"no trailing newline": {
			"brief.md":  canonBrief,
			"README.md": "# Widget\n\n<p>An opinionated widget framework for busy people.</p>",
		},
		"single-line manifest": {
			"brief.md":    canonBrief,
			"plugin.json": `{"name":"w","description":"The widget, but faster."}` + "\n",
		},
		"crlf line endings": {
			"brief.md":  canonBrief,
			"README.md": "# Widget\r\n\r\n<p>An opinionated widget framework.</p>\r\n\r\n## More\r\n",
		},
		"drift mid-file": {
			"brief.md":  canonBrief,
			"README.md": "# Widget\n\nintro\n\n<p>An opinionated widget framework.</p>\n\nmore\n\ntail\n",
		},
		"trailing blank line": {
			"brief.md":  canonBrief,
			"README.md": "# Widget\n\n<p>An opinionated widget framework.</p>\n\n",
		},
	}
	for name, files := range cases {
		t.Run(name, func(t *testing.T) {
			root := writeRepo(t, files)
			gitInit(t, root)
			p, err := Propose(root, abcdLikeConfig())
			if err != nil {
				t.Fatalf("Propose: %v", err)
			}
			if len(p.Diffs) == 0 {
				t.Fatal("no diff produced")
			}
			patch := filepath.Join(t.TempDir(), "p.patch")
			if err := os.WriteFile(patch, []byte(p.Unified()), 0o644); err != nil {
				t.Fatal(err)
			}
			cmd := exec.Command("git", "-C", root, "apply", patch)
			cmd.Env = gittest.Env(t)
			if out, err := cmd.CombinedOutput(); err != nil {
				t.Fatalf("git apply refused the proposal: %v: %s\n--- patch ---\n%s", err, out, p.Unified())
			}
			// The patch and the Proposed content are two renderings of the same
			// change; if they disagree, one of them is lying to the maintainer.
			for _, d := range p.Diffs {
				got, rerr := os.ReadFile(filepath.Join(root, d.File))
				if rerr != nil {
					t.Fatal(rerr)
				}
				if string(got) != d.Proposed {
					t.Errorf("%s after applying the patch = %q, but Proposed = %q", d.File, got, d.Proposed)
				}
			}
		})
	}
}

func gitInit(t *testing.T, root string) {
	t.Helper()
	cmd := exec.Command("git", "-C", root, "init", "-q")
	cmd.Env = gittest.Env(t)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git init: %v: %s", err, out)
	}
}

func TestProposedApplicationRemovesTheDrift(t *testing.T) {
	root := writeRepo(t, map[string]string{
		"brief.md": canonBrief, "README.md": drift143README,
		"plugin.json": drift143Manifest, "AGENTS.md": drift143Agents,
	})
	p, err := Propose(root, abcdLikeConfig())
	if err != nil {
		t.Fatalf("Propose: %v", err)
	}
	// Apply each proposal to a scratch copy of the repo and re-check: adopting
	// every proposed diff must leave the repo clean, or the proposal is not a
	// fix a human can adopt.
	applied := map[string]string{"brief.md": canonBrief}
	for _, name := range []string{"README.md", "plugin.json", "AGENTS.md"} {
		data, err := os.ReadFile(filepath.Join(root, name))
		if err != nil {
			t.Fatal(err)
		}
		applied[name] = string(data)
	}
	for _, d := range p.Diffs {
		applied[d.File] = d.Proposed
	}
	rep, err := Check(writeRepo(t, applied), abcdLikeConfig())
	if err != nil {
		t.Fatalf("Check after applying: %v", err)
	}
	if rep.Drifted() != 0 {
		for _, s := range rep.Surfaces {
			t.Logf("%s: %s found=%q", s.ID, s.Status, s.Found)
		}
		t.Fatalf("applying every proposal left %d surfaces drifted", rep.Drifted())
	}
}
