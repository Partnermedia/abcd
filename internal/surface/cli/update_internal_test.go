package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Partnermedia/abcd/internal/core/ahoy"
	"github.com/Partnermedia/abcd/internal/core/update"
)

// TestRefusalReportRedactsHomeInTargetPath pins that the dispatch-refusal
// receipt redacts a home-rooted target path at construction. The refusal is a
// success-shaped envelope the CLI error-surface scrub never touches, and its
// detail is already ~-redacted in core — an unredacted target_path beside it
// would hand back the very home root the detail hides (iss-2608220142158516).
func TestRefusalReportRedactsHomeInTargetPath(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		t.Skip("no home dir resolvable")
	}
	abs := filepath.Join(home, ".local", "bin", "abcd")
	tgt := ahoy.UpdateTarget{Path: abs, Kind: ahoy.UpdateTargetPluginRoot}
	r := update.Plan(tgt)
	if r == nil {
		t.Fatal("expected a plan refusal for a plugin-root target")
	}
	rep := refusalReport(tgt, r)
	if strings.Contains(rep.TargetPath, home) {
		t.Errorf("refusal receipt leaked the absolute home root: %q", rep.TargetPath)
	}
	if !strings.Contains(rep.TargetPath, "~") {
		t.Errorf("TargetPath = %q, want the ~-redacted path", rep.TargetPath)
	}
	if rep.Action != update.ActionRefused || rep.Refusal == nil {
		t.Errorf("receipt shape = %+v, want a refused action carrying the refusal", rep)
	}
}
