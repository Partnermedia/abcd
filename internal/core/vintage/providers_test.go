package vintage

import (
	"testing"

	"github.com/Partnermedia/abcd/internal/gittest"
)

func TestCheckoutTipProvider(t *testing.T) {
	r := gittest.NewRepo(t)
	r.Commit("first")
	first := r.Git("rev-parse", "HEAD")
	r.Commit("second")
	tip := r.Git("rev-parse", "HEAD")

	t.Run("binary at the tip is fresh", func(t *testing.T) {
		rep := Compare(Current{Revision: tip, Known: true}, CheckoutTip(r.Root(), tip))
		if rep.Outcome != Fresh {
			t.Fatalf("outcome = %q, want fresh (expected=%q)", rep.Outcome, rep.Expected)
		}
	})

	t.Run("binary behind the tip is stale", func(t *testing.T) {
		rep := Compare(Current{Revision: first, Known: true}, CheckoutTip(r.Root(), first))
		if rep.Outcome != Stale {
			t.Fatalf("outcome = %q, want stale", rep.Outcome)
		}
		if rep.Expected != tip {
			t.Fatalf("expected = %q, want tip %q", rep.Expected, tip)
		}
	})

	t.Run("revision foreign to the repo is unknown, never stale", func(t *testing.T) {
		foreign := "0000000000000000000000000000000000000000"
		rep := Compare(Current{Revision: foreign, Known: true}, CheckoutTip(r.Root(), foreign))
		if rep.Outcome != Unknown {
			t.Fatalf("outcome = %q, want unknown", rep.Outcome)
		}
	})

	t.Run("empty repo root yields an unresolved reference", func(t *testing.T) {
		rep := Compare(Current{Revision: tip, Known: true}, CheckoutTip(t.TempDir(), tip))
		if rep.Outcome != Unknown {
			t.Fatalf("outcome = %q, want unknown", rep.Outcome)
		}
	})
}

func TestPinnedVersionProvider(t *testing.T) {
	t.Run("matching pin is fresh", func(t *testing.T) {
		rep := Compare(Current{Revision: "v1.2.3", Known: true}, PinnedVersion("v1.2.3"))
		if rep.Outcome != Fresh {
			t.Fatalf("outcome = %q, want fresh", rep.Outcome)
		}
	})
	t.Run("newer pin than the running version is stale", func(t *testing.T) {
		rep := Compare(Current{Revision: "v1.2.3", Known: true}, PinnedVersion("v1.3.0"))
		if rep.Outcome != Stale {
			t.Fatalf("outcome = %q, want stale", rep.Outcome)
		}
	})
	t.Run("empty pin is an unresolved reference", func(t *testing.T) {
		rep := Compare(Current{Revision: "v1.2.3", Known: true}, PinnedVersion(""))
		if rep.Outcome != Unknown {
			t.Fatalf("outcome = %q, want unknown", rep.Outcome)
		}
	})
}
