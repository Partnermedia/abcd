package guard

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// The warn-rate gate.
//
// A fail-safe that fires on ordinary work is worse than none: it trains people to
// dismiss the warning, and the next warning it dismisses is the real one. The
// guard's own design record makes this a STOP rather than a metric — "a warn rate
// that trains users to ignore warns" halts the work.
//
// So the two corpora ship as test data rather than as a number in a commit
// message. A rate measured once and written down is a claim; a corpus in the
// repository is a claim the next change has to keep true.

// maxRepoCorpusFires is the ceiling on Tier 2 fires across command lines mined
// from this repository's own prose, Makefile, scripts and workflows. This corpus
// is the closest available stand-in for "commands people actually write here",
// and the fail-safe is only affordable while it stays silent on them.
//
// An absolute count, not a rate: a 1% rate over ~960 lines permits NINE fires
// against a measured zero, which is not "just above the measured rate" however
// the comment describes it, and it would let a change take the guard from silent
// to noticeably noisy without failing anything. Two is headroom for a doc line
// that legitimately demonstrates a hazard.
//
// Raising this is always allowed and never quiet: it is the visible price of a
// noisier guard, and the test logs every fire so the diff shows what started
// firing.
const maxRepoCorpusFires = 2

// TestSpeculationWarnRateOnRepoCorpus measures Tier 2 against real command lines
// from this repository. Every fire is logged, so raising the ceiling is never the
// path of least resistance — the diff shows what started firing.
func TestSpeculationWarnRateOnRepoCorpus(t *testing.T) {
	lines := readCorpus(t, filepath.Join("testdata", "corpus", "repo-mined.txt"))
	if len(lines) < 500 {
		t.Fatalf("the repo corpus holds %d lines; below a few hundred it cannot say anything about a rate", len(lines))
	}

	r := Defaults()
	var speculative, unparsable int
	for _, line := range lines {
		d, err := r.Check(line)
		if err != nil {
			// An unparsable line is not a false positive — the front door reports a
			// fault, which is never read as clearance. Counted so a corpus that
			// silently stopped being checked cannot pass.
			unparsable++
			continue
		}
		if firedTier2(d) {
			speculative++
			t.Logf("Tier 2 fired: %s\n  %s", line, d.Message)
		}
	}

	checked := len(lines) - unparsable
	if checked == 0 {
		t.Fatal("no corpus line was checked: the corpus or the tokenizer is broken")
	}
	t.Logf("repo corpus: %d lines, %d checked, %d unparsable, %d Tier 2 fires (%.3f%%)",
		len(lines), checked, unparsable, speculative, float64(speculative)/float64(checked)*100)

	if speculative > maxRepoCorpusFires {
		t.Fatalf("Tier 2 fires on %d of %d real command lines, over the ceiling of %d.\n"+
			"This is the warn-storm STOP, not a threshold to raise: a fail-safe people learn to ignore is not a fail-safe.",
			speculative, checked, maxRepoCorpusFires)
	}
}

// TestAdversarialCorpus holds both directions at once. A rate alone cannot say
// whether the quiet lines are quiet for the right reason, so every line carries
// the verdict it is supposed to get — including the two accepted false positives,
// which are pinned so they stay two.
func TestAdversarialCorpus(t *testing.T) {
	lines := readCorpus(t, filepath.Join("testdata", "corpus", "adversarial.txt"))
	r := Defaults()

	counts := map[string]int{}
	for _, line := range lines {
		want, command, found := strings.Cut(line, "\t")
		if !found {
			t.Errorf("corpus line is not `<expectation>\\t<command>`: %q", line)
			continue
		}
		want, command = strings.TrimSpace(want), strings.TrimSpace(command)
		counts[want]++

		d, err := r.Check(command)
		if err != nil {
			t.Errorf("%s: guard could not evaluate it: %v", command, err)
			continue
		}
		// A line can fire in BOTH tiers — `git clean -fd ; nice git push --force`
		// is the case the per-segment gate exists for — and the winning message
		// then comes from the registry entry, which is the more specific lesson.
		// So "Tier 2 fired" is membership in Matches, never the winning id.
		speculative := firedTier2(d)

		switch want {
		case "quiet":
			if d.Verdict != VerdictAllow {
				t.Errorf("FALSE POSITIVE — %q\n  want allow, got %s (%s): %s", command, d.Verdict, d.EntryID, d.Message)
			}
		case "warn", "fp":
			if d.Verdict != VerdictWarn {
				t.Errorf("MISSED — %q\n  want a Tier 2 warn, got %s (%s)", command, d.Verdict, d.EntryID)
			} else if !speculative {
				t.Errorf("%q warned, but as %s rather than the fail-safe — the corpus label is wrong or Tier 1 grew", command, d.EntryID)
			}
		case "block":
			if d.Verdict != VerdictBlock {
				t.Errorf("%q: want a Tier 1 block, got %s (%s)", command, d.Verdict, d.EntryID)
			}
			if speculative {
				t.Errorf("%q was decided by the fail-safe, not by the entry that names it", command)
			}
		default:
			t.Errorf("unknown expectation %q on line %q", want, command)
		}
	}

	// The corpus is only evidence while it holds real cases in both directions.
	for _, label := range []string{"quiet", "warn", "block", "fp"} {
		if counts[label] == 0 {
			t.Errorf("the adversarial corpus has no %q cases left: it has stopped being adversarial", label)
		}
	}
	t.Logf("adversarial corpus: %d quiet, %d warn, %d block, %d accepted false positives",
		counts["quiet"], counts["warn"], counts["block"], counts["fp"])
}

// firedTier2 reports whether the fail-safe contributed to this decision.
func firedTier2(d Decision) bool {
	for _, id := range d.Matches {
		if id == speculativeEntryID {
			return true
		}
	}
	return false
}

// readCorpus reads one corpus file, dropping blank lines and `#` comments.
func readCorpus(t *testing.T, path string) []string {
	t.Helper()
	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("cannot read the corpus at %s: %v\n"+
			"adr-42 makes these committed test data: a measurement nobody can repeat is not evidence", path, err)
	}
	defer f.Close()

	var lines []string
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 1<<20), 1<<20)
	for sc.Scan() {
		line := strings.TrimRight(sc.Text(), " \t")
		if strings.TrimSpace(line) == "" || strings.HasPrefix(strings.TrimSpace(line), "#") {
			continue
		}
		lines = append(lines, line)
	}
	if err := sc.Err(); err != nil {
		t.Fatalf("reading %s: %v", path, err)
	}
	return lines
}
