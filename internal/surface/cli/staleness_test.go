package cli

import (
	"strings"
	"testing"

	"github.com/Partnermedia/abcd/internal/core/ahoy"
	"github.com/Partnermedia/abcd/internal/core/vintage"
)

func TestFormatStalenessNotice(t *testing.T) {
	rev := "1111111111111111111111111111111111111111"
	tip := "2222222222222222222222222222222222222222"
	exe := "/home/alice/.local/bin/abcd"

	t.Run("stale fires and names binary, revision, tip, and rebuild fix", func(t *testing.T) {
		n := formatStalenessNotice(ahoy.DogfoodReport{IsDogfood: true, Outcome: vintage.Stale, Revision: rev, Tip: tip}, exe)
		for _, want := range []string{exe, rev[:12], tip[:12], "make build"} {
			if !strings.Contains(n, want) {
				t.Fatalf("notice %q missing %q", n, want)
			}
		}
	})

	t.Run("unknown (dirty rebuild) fires", func(t *testing.T) {
		n := formatStalenessNotice(ahoy.DogfoodReport{IsDogfood: true, Outcome: vintage.Unknown, Revision: rev, Tip: tip}, exe)
		if n == "" {
			t.Fatal("expected a notice for an unknown (dirty) dogfood binary")
		}
		if !strings.Contains(n, "make build") {
			t.Fatalf("notice %q missing rebuild fix", n)
		}
	})

	t.Run("fresh is silent", func(t *testing.T) {
		n := formatStalenessNotice(ahoy.DogfoodReport{IsDogfood: true, Outcome: vintage.Fresh, Revision: tip, Tip: tip}, exe)
		if n != "" {
			t.Fatalf("fresh binary should be silent, got %q", n)
		}
	})

	t.Run("non-dogfood is silent", func(t *testing.T) {
		n := formatStalenessNotice(ahoy.DogfoodReport{IsDogfood: false, Outcome: vintage.Stale, Revision: rev, Tip: tip}, exe)
		if n != "" {
			t.Fatalf("non-dogfood should be silent, got %q", n)
		}
	})
}
