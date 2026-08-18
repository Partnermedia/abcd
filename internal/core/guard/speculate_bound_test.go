package guard

import (
	"strings"
	"testing"
	"time"
)

// stdinCapBytes mirrors the 1 MiB cap both front doors put on a candidate
// (maxHookStdinBytes in cli.go, maxGuardStdinBytes in guard.go). It is the
// largest command line that can reach Check, so it is the size the bound has to
// hold at — anything smaller measures a case an author would not choose.
const stdinCapBytes = 1 << 20

// worstCaseCandidate is the adversarial shape, not a random one. A repeated
// WRAPPER name is the expensive input: every speculative start pays commandOf,
// which walks the leading wrapper chain, so the cost per start grows with the
// line unless the window bounds it. `foo foo foo …` is ~30x cheaper and would
// have let the regression through.
func worstCaseCandidate(size int) string {
	const unit = "env "
	return "myrunner " + strings.Repeat(unit, size/len(unit))
}

// TestSpeculationIsBoundedAtTheStdinCap is the DoS gate, and it is load-bearing
// rather than an optimisation: the guard runs inside a PreToolUse hook, so a slow
// guard is an outage on every command in the session.
//
// The history is the reason it exists. An earlier attempt at iss-200 made the
// matcher re-tokenize and restart per entry, went quadratic, and was BLOCKED in
// review at ~5 min for this input; expandPayloads' doc comment forbids that shape
// by name. Speculation reintroduced the same danger from a different direction —
// capping starts at 64 still left every start paying commandOf over the whole
// line, which measured 14.2s here. Bounding the per-start window as well brings
// it to ~120ms.
//
// The budget is deliberately loose. It is not a performance target; it is the
// line between "slow" and "the session has hung", and it must not fail on a
// loaded CI runner. Removing either bound blows through it by an order of
// magnitude.
func TestSpeculationIsBoundedAtTheStdinCap(t *testing.T) {
	const budget = 3 * time.Second

	candidate := worstCaseCandidate(stdinCapBytes)
	if len(candidate) < stdinCapBytes {
		t.Fatalf("candidate is %d bytes, want at least the %d-byte stdin cap", len(candidate), stdinCapBytes)
	}

	start := time.Now()
	d, err := Defaults().Check(candidate)
	elapsed := time.Since(start)

	if err != nil {
		t.Fatalf("guard could not evaluate the worst-case candidate: %v", err)
	}
	// It warns rather than allowing: the line is past every bound, so the guard
	// says it stopped looking. A silent allow here would be the defect, not the
	// slowness.
	if d.Verdict != VerdictWarn {
		t.Errorf("verdict = %q, want %q: a candidate too long to inspect was waved through", d.Verdict, VerdictWarn)
	}
	if elapsed > budget {
		t.Fatalf("Check took %s on a %d-byte worst-case candidate, over the %s budget.\n"+
			"The guard gates a PreToolUse hook; this is a denial of service, not a slow test.\n"+
			"Check maxSpeculativeStarts and maxSpeculativeWindow — removing either takes this to ~14s.",
			elapsed.Round(time.Millisecond), len(candidate), budget)
	}
	t.Logf("worst case: %d bytes in %s", len(candidate), elapsed.Round(time.Millisecond))
}

// BenchmarkCheck records what the two tiers cost on the shapes that matter, so a
// future change to matching can be compared rather than guessed at.
func BenchmarkCheck(b *testing.B) {
	cases := []struct{ name, cmd string }{
		{"ordinary", "git status --short"},
		{"tier1-block", "git push --force origin main"},
		{"tier2-warn", "nice git push --force origin main"},
		{"tier2-miss", "myrunner alpha beta gamma delta"},
		{"worst-64k", worstCaseCandidate(64 << 10)},
		{"worst-1mib", worstCaseCandidate(stdinCapBytes)},
	}
	for _, c := range cases {
		r := Defaults()
		b.Run(c.name, func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				if _, err := r.Check(c.cmd); err != nil {
					b.Fatalf("Check: %v", err)
				}
			}
		})
	}
}
