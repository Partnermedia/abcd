package ahoy

import (
	"testing"

	"github.com/REPPL/abcd-cli/internal/core/vintage"
	"github.com/REPPL/abcd-cli/internal/gittest"
)

func TestDogfoodStalenessFrom(t *testing.T) {
	r := gittest.NewRepo(t)
	r.Commit("first")
	first := r.Git("rev-parse", "HEAD")
	r.Commit("second")
	tip := r.Git("rev-parse", "HEAD")

	cases := []struct {
		name        string
		cur         vintage.Current
		wantDogfood bool
		wantOutcome vintage.Outcome
	}{
		{
			name:        "binary at the tip is fresh (silent)",
			cur:         vintage.Current{Revision: tip, Known: true},
			wantDogfood: true,
			wantOutcome: vintage.Fresh,
		},
		{
			name:        "binary behind the tip is stale",
			cur:         vintage.Current{Revision: first, Known: true},
			wantDogfood: true,
			wantOutcome: vintage.Stale,
		},
		{
			name:        "dirty rebuild atop a known commit is unknown",
			cur:         vintage.Current{Revision: tip, Known: false},
			wantDogfood: true,
			wantOutcome: vintage.Unknown,
		},
		{
			name:        "revision foreign to the checkout is not a dogfood report",
			cur:         vintage.Current{Revision: "0000000000000000000000000000000000000000", Known: true},
			wantDogfood: false,
		},
		{
			name:        "no embedded revision is not a dogfood report",
			cur:         vintage.Current{Known: false},
			wantDogfood: false,
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := dogfoodStalenessFrom(c.cur, r.Root())
			if got.IsDogfood != c.wantDogfood {
				t.Fatalf("IsDogfood = %v, want %v", got.IsDogfood, c.wantDogfood)
			}
			if c.wantDogfood && got.Outcome != c.wantOutcome {
				t.Fatalf("Outcome = %q, want %q", got.Outcome, c.wantOutcome)
			}
			if c.wantDogfood && c.wantOutcome == vintage.Stale && got.Tip != tip {
				t.Fatalf("Tip = %q, want %q", got.Tip, tip)
			}
		})
	}
}
