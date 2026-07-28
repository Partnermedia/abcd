
package positioning

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestZZProbe(t *testing.T) {
	cases := map[string]string{
		"trailing_nl":    "# Widget\n\n<p>An opinionated widget framework for busy people.</p>\n",
		"no_trailing_nl": "# Widget\n\n<p>An opinionated widget framework for busy people.</p>",
		"midfile":        "# Widget\n\nintro\n\n<p>An opinionated widget framework.</p>\n\nmore\n\ntail\n",
		"crlf":           "# Widget\r\n\r\n<p>An opinionated widget framework.</p>\r\n\r\n## More\r\n",
		"blank_after":    "# Widget\n\n<p>An opinionated widget framework.</p>\n\n",
	}
	for name, readme := range cases {
		t.Run(name, func(t *testing.T) {
			root := writeRepo(t, map[string]string{"brief.md": canonBrief, "README.md": readme})
			gitInit(t, root)
			p, err := Propose(root, abcdLikeConfig())
			if err != nil {
				t.Fatalf("Propose: %v", err)
			}
			if len(p.Diffs) == 0 {
				t.Logf("NO DIFF")
				return
			}
			t.Logf("ORIG=%q", readme)
			t.Logf("PATCH:\n%s---END---", p.Unified())
			patch := filepath.Join(t.TempDir(), "p.patch")
			os.WriteFile(patch, []byte(p.Unified()), 0o644)
			out, err := exec.Command("git", "-C", root, "apply", patch).CombinedOutput()
			if err != nil {
				t.Errorf("git apply FAILED: %v: %s", err, out)
				return
			}
			got, _ := os.ReadFile(filepath.Join(root, "README.md"))
			t.Logf("APPLIED=%q", string(got))
			t.Logf("PROPOSED=%q", p.Diffs[0].Proposed)
			if string(got) != p.Diffs[0].Proposed {
				t.Errorf("MISMATCH applied vs Proposed")
			}
		})
	}
}
