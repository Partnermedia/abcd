package lint_test

import (
	"encoding/json"
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

// TestPreflightRunsEveryTaggedEvalLane holds the position every tagged eval
// lane is supposed to occupy (iss-2608311632382737).
//
// Framework 8.6 makes the read-block eval the only component capable of
// falsifying the assembler's firewall, and framework 8.7 makes the amnesia eval
// the only check on what a re-run is handed. Both live behind a build tag, so
// `go test ./...` — preflight's own test step — compiles neither, and preflight
// ran neither target: a defect in the component that falsifies the firewall
// passed every local gate and surfaced, if at all, in CI, where the job that
// executes it is not a required status check. That is not hypothetical — a
// path-elision defect in the amnesia eval's own guard was unsatisfiable wherever
// the process temp directory is the Linux one, so it landed green locally and
// was found by an adversarial review rather than by any gate.
//
// The roster is DERIVED rather than listed here: a Makefile target is an eval
// lane when its recipe runs `go test` with `-tags`. So a third lane joins the
// push gate by existing, and cannot be added to the Makefile while staying
// invisible to it — which a hand-written pair of names could not promise. The
// restatement gate above then carries whatever this finds into every surface
// that enumerates the gates.
func TestPreflightRunsEveryTaggedEvalLane(t *testing.T) {
	root := filepath.Join("..", "..", "..")

	lanes := taggedEvalLanes(t, root)
	if len(lanes) == 0 {
		t.Fatal("parsed no `go test -tags ...` lane from the Makefile; the parser or the recipes changed shape")
	}
	declared := preflightPrereqs(t, root)

	for _, lane := range lanes {
		if slices.Contains(declared, lane) {
			continue
		}
		t.Errorf("the Makefile declares the tagged eval lane %q, and preflight does not run it "+
			"(it declares: %s).\n\n"+
			"`go test ./...` does not compile a tagged file, so a lane preflight omits is "+
			"guarded by nothing a push can fail on, and the eval that certifies the read-block "+
			"then guards nothing that blocks anything.", lane, strings.Join(declared, " "))
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

// TestColdReadingEvalsIsARequiredStatusCheck holds the other half of that gate
// (iss-2608311051046981): the CI job that runs the read-block eval (framework
// 8.6) blocks a merge rather than merely reporting on one.
//
// Three properties, and each is a way the guarantee can be lost. The workflow
// defines the job, so a rename cannot leave a required context that never
// arrives and wedges the queue. The job stands down on no event — no `if:` and
// no `needs:` — so its context arrives on every event the other required checks
// arrive on, which is what makes requiring it safe. And the committed ruleset
// mirror requires the context, which is the tree's own record of what the live
// ruleset is set to.
func TestColdReadingEvalsIsARequiredStatusCheck(t *testing.T) {
	const job = "cold-reading-evals"
	root := filepath.Join("..", "..", "..")

	workflow := readRepoFile(t, root, ".github/workflows/ci.yml")
	block, ok := workflowJobBlock(workflow, job)
	if !ok {
		t.Fatalf(".github/workflows/ci.yml defines no %q job; a required status check whose "+
			"job does not exist never reports, and a required context that never arrives "+
			"wedges the merge queue", job)
	}
	for _, standDown := range []string{"if:", "needs:"} {
		if !strings.Contains(block, standDown) {
			continue
		}
		t.Errorf("the %q job carries a %q, so it can stand down on some event; a required "+
			"check must report on every event the others do, and a stood-down job that "+
			"still reports its context green is a green for work that did not happen",
			job, standDown)
	}

	const mirror = ".abcd/work/rulesets/main-protection.json"
	var ruleset struct {
		Rules []struct {
			Type       string `json:"type"`
			Parameters struct {
				RequiredStatusChecks []struct {
					Context string `json:"context"`
				} `json:"required_status_checks"`
			} `json:"parameters"`
		} `json:"rules"`
	}
	if err := json.Unmarshal([]byte(readRepoFile(t, root, mirror)), &ruleset); err != nil {
		t.Fatalf("decoding %s: %v", mirror, err)
	}
	var required []string
	for _, rule := range ruleset.Rules {
		for _, c := range rule.Parameters.RequiredStatusChecks {
			required = append(required, c.Context)
		}
	}
	if len(required) == 0 {
		t.Fatalf("%s declares no required status checks at all; the mirror is the tree's "+
			"record of the live ruleset, so an empty list here reads as an unprotected branch", mirror)
	}
	if !slices.Contains(required, job) {
		t.Fatalf("%s does not require the %q context (it requires: %s).\n\n"+
			"The point of the always-run lane is that a record-only pull request cannot reach "+
			"main with warm content in included material, and an unrequired check does not "+
			"stop one.", mirror, job, strings.Join(required, ", "))
	}
	if !slices.IsSorted(required) {
		t.Errorf("%s lists its required contexts out of order (%s); the mirror is refreshed "+
			"by hand from a sorted `jq -S` rendering, and an unsorted list means it was not",
			mirror, strings.Join(required, ", "))
	}
}

// workflowJobBlock returns one job's block from a workflow: the `  <name>:`
// line and everything indented under it, up to the next job. Hand-parsed for
// the reason preflightPrereqs is: this repository carries no YAML parser for
// its own workflows and adds none for one job.
func workflowJobBlock(workflow, job string) (string, bool) {
	lines := strings.Split(workflow, "\n")
	at := -1
	for i, l := range lines {
		if l == "  "+job+":" {
			at = i
			break
		}
	}
	if at < 0 {
		return "", false
	}
	nextJob := regexp.MustCompile(`^  [A-Za-z]`)
	end := len(lines)
	for i := at + 1; i < len(lines); i++ {
		if nextJob.MatchString(lines[i]) {
			end = i
			break
		}
	}
	return strings.Join(lines[at:end], "\n"), true
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
