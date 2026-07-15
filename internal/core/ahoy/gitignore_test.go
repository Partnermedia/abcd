package ahoy

import (
	"strings"
	"testing"
)

// TestRemoveGitignoreBlocksBalanced proves a matched BEGIN..END span is removed
// in full while surrounding user content is preserved.
func TestRemoveGitignoreBlocksBalanced(t *testing.T) {
	in := "node_modules/\n# BEGIN ABCD\n.work/\n# END ABCD\nbuild/\n"
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
	in := "node_modules/\n# BEGIN ABCD\n.work/\n# comment\nsecrets.env\nbuild/\n"
	got := removeGitignoreBlocks(in, "\n")
	for _, must := range []string{"node_modules/", "secrets.env", "build/", ".work/", "# comment"} {
		if !strings.Contains(got, must) {
			t.Errorf("orphan-BEGIN rewrite deleted user content %q; got %q", must, got)
		}
	}
	if strings.Contains(got, "# BEGIN ABCD") {
		t.Errorf("orphan BEGIN marker should be dropped; got %q", got)
	}
}
