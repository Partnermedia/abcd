package lint_test

import (
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"testing"

	"github.com/intentdriven/abcd/internal/core/issueschema"
)

// preflight's prerequisite list is a DERIVED value, and every surface that
// restates it must name the same set (iss-2608242043243131).
//
// The Makefile declares the list once. The surfaces in the list below restate
// it by hand, and nothing derived them from the recipe, so they drifted every
// time the recipe changed — twice in two releases, each time caught only by a
// host-run semantic reviewer refusing a release:
//
//   - v0.6.4: `lint-issues` made preflight four gates. The commit that added it
//     updated AGENTS.md alone; the install guide, CONTRIBUTING.md and the
//     Makefile's own comment still said three.
//   - v0.6.5: `site-render` made it five. The same three surfaces still said four
//     — including the two just corrected.
//
// The repository already solved this exact shape once, for a different derived
// value: TestInstallGuideDocumentsTheInstallAndUpdatePath reads the module path
// out of go.mod and asserts the documented install command contains it. This is
// that move applied to the gate list.
//
// Deliberately a containment check rather than an equality one. These are
// sentences, not lists — "together with the lint-reviews, lint-issues, … gates"
// — so requiring an exact rendering would fail on a comma. What must hold is
// that every prerequisite the recipe declares is NAMED where the list is
// restated. The reverse direction — a surface naming a gate the recipe no
// longer has — is not checked here; a retired gate's mention is a review
// concern, since a phantom NAME has no recipe line to derive a check from.
func TestPreflightGateListIsNotRestatedWrongly(t *testing.T) {
	root := filepath.Join("..", "..", "..")

	declared := preflightPrereqs(t, root)
	if len(declared) < 2 {
		t.Fatalf("parsed %d preflight prerequisites from the Makefile; the parser or the recipe changed shape", len(declared))
	}

	// Every surface that enumerates the gates. A file joins this list when it
	// starts restating them — which is the moment it becomes able to drift.
	for _, rel := range []string{
		"Makefile",               // the recipe's own comment, above the recipe
		"docs/how-to/install.md", // the build section a contributor copies
		"CONTRIBUTING.md",        // the local-gates paragraph
		"AGENTS.md",              // the definition-of-done list
		"CLAUDE.md",              // AGENTS.md's committed mirror
		".githooks/pre-push",     // the hook that invokes the recipe
	} {
		t.Run(rel, func(t *testing.T) {
			prose := readRepoFile(t, root, rel)
			// The Makefile is checked against the COMMENT BLOCK above the recipe,
			// not the whole file: the recipe line is where `declared` comes from,
			// so a whole-file containment check passes on its own input and can
			// never fail.
			if rel == "Makefile" {
				prose = commentBlockAbovePreflight(prose)
			}
			// No skip for a file that stops enumerating. An earlier draft skipped
			// when "lint-reviews" was absent, which made the gate defeatable by
			// exactly the drift it targets: rewriting a sentence to name two of
			// five gates removes the sentinel along with the gates, and the
			// subtest passes by skipping. The file list is hand-curated — a
			// surface that genuinely stops enumerating leaves this list in the
			// same change.
			for _, gate := range declared {
				if !strings.Contains(prose, gate) {
					t.Errorf("%s enumerates the preflight gates but omits %q.\n\n"+
						"The Makefile declares: %s\n"+
						"A restated list that has fallen behind the recipe understates what "+
						"guards a push, and this drift has reached a release twice.",
						rel, gate, strings.Join(declared, " "))
				}
			}
		})
	}
}

// preflightPrereqs reads the prerequisites off the `preflight:` recipe line.
// Hand-parsed for the same reason the workflow contracts beside it are: this
// repository carries no Makefile parser and adds none for one line.
func preflightPrereqs(t *testing.T, root string) []string {
	t.Helper()
	for _, line := range strings.Split(readRepoFile(t, root, "Makefile"), "\n") {
		rest, ok := strings.CutPrefix(line, "preflight:")
		if !ok {
			continue
		}
		var out []string
		for _, f := range strings.Fields(rest) {
			// Only the lint-style prerequisites are restated in prose; a future
			// non-gate prerequisite should not force itself into a sentence.
			if regexp.MustCompile(`^[a-z][a-z-]*$`).MatchString(f) {
				out = append(out, f)
			}
		}
		return out
	}
	t.Fatal("Makefile declares no `preflight:` recipe")
	return nil
}

// commentBlockAbovePreflight returns the contiguous `#` comment block directly
// above the `preflight:` recipe, which is the part that restates the gate list.
func commentBlockAbovePreflight(makefile string) string {
	lines := strings.Split(makefile, "\n")
	at := -1
	for i, l := range lines {
		if strings.HasPrefix(l, "preflight:") {
			at = i
			break
		}
	}
	if at < 0 {
		return ""
	}
	start := at
	for start > 0 && strings.HasPrefix(strings.TrimSpace(lines[start-1]), "#") {
		start--
	}
	return strings.Join(lines[start:at], "\n")
}

func readRepoFile(t *testing.T, root, rel string) string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(rel)))
	if err != nil {
		t.Fatalf("read %s: %v", rel, err)
	}
	return string(b)
}

// The issue-resolution gate scopes its git pathspecs to the ledger's STATUS
// directories, and the shell cannot import Go — so the script holds the second
// and last spelling of the list issueschema.StatusDirs owns. This reads the
// array the script declares and holds it to that one value.
//
// It matters because the ledger tree gained sibling record families (readings/,
// dispositions/) whose files are not issue records: RS002 and RS003 scan every
// `.md` under the pathspec and would read a `commit:`-shaped line out of one of
// them. Scoping is what keeps them out, and a scope that drifts from the
// canonical list either re-admits them or drops a real status folder — the
// second of which fails open, silently.
func TestIssueResolutionGateScopesToStatusDirs(t *testing.T) {
	root := filepath.Join("..", "..", "..")
	script := readRepoFile(t, root, "scripts/check-issue-resolution.sh")

	m := regexp.MustCompile(`(?m)^STATUS_DIRS=\(([^)]*)\)`).FindStringSubmatch(script)
	if m == nil {
		t.Fatalf("scripts/check-issue-resolution.sh declares no STATUS_DIRS=(...) array;\n" +
			"the gate must scope to the ledger's status directories, and the array is where it says which")
	}
	got := strings.Fields(m[1])
	want := issueschema.StatusDirs
	if !slices.Equal(got, want) {
		t.Fatalf("scripts/check-issue-resolution.sh declares STATUS_DIRS=%v, want %v (issueschema.StatusDirs)",
			got, want)
	}
}

// Every tagged eval lane the Makefile declares is a preflight prerequisite
// (iss-2608311632382737).
//
// The eval files sit behind a build tag, so `go test ./...` — preflight's own
// test step — never compiles them. A lane that preflight does not name
// therefore runs in no gate a push can be stopped by: a defect in it passes
// every local gate and surfaces in CI, where the job that executes it is not a
// required status check. That is not hypothetical. A path-elision defect in the
// amnesia eval's guard was unsatisfiable wherever the process temp directory is
// the Linux one; it landed green locally and was found by an adversarial review
// rather than by any gate.
//
// The roster is DERIVED, not listed here: a target is an eval lane when its
// recipe runs `go test` with `-tags`. So a third lane joins this check by
// existing, and cannot be added to the Makefile while staying invisible to the
// push gate. Both current lanes cost about five seconds locally, which is what
// makes "run them all" the right shape rather than a sampled subset — and the
// two are not redundant despite `smoke` compiling a superset of
// `evals-cold-reading`'s files: they are separate tag sets, so a cold-reading
// file that reaches for a smoke-only helper compiles under one and not the
// other, exactly the split CI's two jobs cover.
func TestPreflightRunsEveryTaggedEvalLane(t *testing.T) {
	root := filepath.Join("..", "..", "..")

	lanes := taggedEvalLanes(t, root)
	if len(lanes) == 0 {
		t.Fatal("parsed no `go test -tags ...` lane from the Makefile; the parser or the recipes changed shape")
	}
	declared := preflightPrereqs(t, root)

	for _, lane := range lanes {
		if !slices.Contains(declared, lane) {
			t.Errorf("the Makefile declares the tagged eval lane %q, but preflight does not run it.\n\n"+
				"preflight declares: %s\n"+
				"`go test ./...` does not compile a tagged file, so a lane preflight omits is "+
				"guarded by nothing a push can fail on.",
				lane, strings.Join(declared, " "))
		}
	}
}

// taggedEvalLanes returns the Makefile targets whose recipe runs `go test` with
// a `-tags` selector — hand-parsed, for the reason preflightPrereqs is.
func taggedEvalLanes(t *testing.T, root string) []string {
	t.Helper()
	target := regexp.MustCompile(`^([a-z][a-z-]*):`)
	var lanes []string
	var current string
	for _, line := range strings.Split(readRepoFile(t, root, "Makefile"), "\n") {
		if m := target.FindStringSubmatch(line); m != nil {
			current = m[1]
			continue
		}
		if !strings.HasPrefix(line, "\t") || current == "" {
			continue
		}
		body := strings.TrimSpace(strings.TrimPrefix(line, "\t"))
		if strings.HasPrefix(body, "go test ") && strings.Contains(body, "-tags ") &&
			!slices.Contains(lanes, current) {
			lanes = append(lanes, current)
		}
	}
	return lanes
}
