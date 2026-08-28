package update

import (
	"strings"
	"testing"

	"github.com/intentdriven/abcd/internal/core/ahoy"
)

// TestPlanBrewCellarFoldsCaseOnFoldingFS proves the Homebrew Cellar refusal
// matches case-insensitively on a case-folding filesystem: a case-variant Cellar
// path still wins the brew remedy, instead of falling through to the generic
// foreign-binary refusal that names the wrong fix. With the filesystem
// case-SENSITIVE the prefix match stays byte-exact (iss-2608270908349399).
func TestPlanBrewCellarFoldsCaseOnFoldingFS(t *testing.T) {
	restore := caseFoldingFS
	t.Cleanup(func() { caseFoldingFS = restore })

	// A case-variant spelling of /opt/homebrew/Cellar/ as macOS/APFS may echo it.
	variant := ahoy.UpdateTarget{
		Path:         "/opt/homebrew/bin/abcd",
		ResolvedPath: "/opt/homebrew/CELLAR/abcd/0.6.1/bin/abcd",
		Kind:         ahoy.UpdateTargetForeign,
	}

	t.Run("folding FS routes a case-variant Cellar path to the brew remedy", func(t *testing.T) {
		caseFoldingFS = func() bool { return true }
		r := Plan(variant)
		if r == nil || !strings.Contains(r.Remedy, "brew upgrade abcd") {
			t.Fatalf("a case-variant Cellar path must refuse naming brew upgrade on a folding FS: %+v", r)
		}
	})

	t.Run("case-sensitive FS keeps the exact-case match", func(t *testing.T) {
		caseFoldingFS = func() bool { return false }
		// The variant falls through to the generic foreign refusal (wrong-for-brew
		// but correct for a case-sensitive host, where CELLAR is a different dir).
		if r := Plan(variant); r == nil || strings.Contains(r.Remedy, "brew upgrade") {
			t.Fatalf("with fold off a case-variant Cellar path must NOT match brew: %+v", r)
		}
		// An exact-case Cellar path still wins the brew remedy with fold off.
		exact := ahoy.UpdateTarget{
			Path:         "/opt/homebrew/bin/abcd",
			ResolvedPath: "/opt/homebrew/Cellar/abcd/0.6.1/bin/abcd",
			Kind:         ahoy.UpdateTargetForeign,
		}
		if r := Plan(exact); r == nil || !strings.Contains(r.Remedy, "brew upgrade abcd") {
			t.Fatalf("an exact-case Cellar path must refuse naming brew upgrade with fold off: %+v", r)
		}
	})
}
