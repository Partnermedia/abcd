package site

import (
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"
)

// release.yml's `site` job hands the deploy to site.yml as a reusable workflow, and
// that call is resolved when the workflow file is PARSED — before any job starts. A
// `with:` block naming an input site.yml does not declare, or omitting one it
// requires, fails the whole release run up front.
//
// Every other coupling between these two files is a runtime one, and `needs: release`
// contains those: the Release is already published and a red site job cannot touch it.
// This coupling is different in kind, because a parse failure means no job runs at
// all — there is nothing for a job dependency to isolate. So it is pinned here
// instead, and a drifted contract fails CI on the pull request rather than on release
// day, which is the only day anyone would otherwise find out.
//
// Hand-parsed deliberately. This repository carries no YAML dependency and adds none
// (internal/core/lint's gate_lockstep hand-parses the same workflows for the same
// reason), and both blocks read are fixed-indentation keys and comments.
const (
	siteWorkflowRel    = ".github/workflows/site.yml"
	releaseWorkflowRel = ".github/workflows/release.yml"

	// The job in release.yml that calls site.yml, and the path it must call.
	callerJob      = "site"
	calledWorkflow = "./.github/workflows/site.yml"
)

// yamlKeyLine matches a mapping key and captures its indentation and name. Values are
// ignored: this test reads the SHAPE of two blocks, never their contents.
var yamlKeyLine = regexp.MustCompile(`^( *)([A-Za-z0-9_-]+):`)

func TestSiteWorkflowInputContract(t *testing.T) {
	siteLines := workflowLines(t, siteWorkflowRel)
	releaseLines := workflowLines(t, releaseWorkflowRel)

	declared := declaredCallInputs(t, siteLines)
	supplied := suppliedCallArgs(t, releaseLines)

	if len(declared) == 0 {
		t.Fatalf("%s: parsed no workflow_call inputs; the parser or the file changed shape", siteWorkflowRel)
	}
	if len(supplied) == 0 {
		t.Fatalf("%s: parsed no `with:` keys on the %q job; the parser or the file changed shape",
			releaseWorkflowRel, callerJob)
	}

	// Every key the caller supplies must be declared. An undeclared key is an
	// immediate parse error on the real run.
	for _, name := range supplied {
		if _, ok := declared[name]; !ok {
			t.Errorf("%s's %q job passes `%s:`, which %s does not declare as a workflow_call input",
				releaseWorkflowRel, callerJob, name, siteWorkflowRel)
		}
	}

	// Every REQUIRED input must be supplied. An omitted required input is the same
	// parse error from the other direction.
	suppliedSet := make(map[string]bool, len(supplied))
	for _, name := range supplied {
		suppliedSet[name] = true
	}
	for name, required := range declared {
		if required && !suppliedSet[name] {
			t.Errorf("%s declares `%s` as a REQUIRED workflow_call input, but %s's %q job does not pass it",
				siteWorkflowRel, name, releaseWorkflowRel, callerJob)
		}
	}

	// The sets must be equal, not merely compatible. An optional input the caller
	// never passes is a contract nobody exercises: either the caller should pass it
	// or site.yml should stop declaring it, and silence here would hide the choice.
	declaredNames := make([]string, 0, len(declared))
	for name := range declared {
		declaredNames = append(declaredNames, name)
	}
	sort.Strings(declaredNames)
	sortedSupplied := append([]string(nil), supplied...)
	sort.Strings(sortedSupplied)
	if strings.Join(declaredNames, ",") != strings.Join(sortedSupplied, ",") {
		t.Errorf("workflow_call contract drifted:\n  %s declares: %s\n  %s passes:   %s",
			siteWorkflowRel, strings.Join(declaredNames, ", "),
			releaseWorkflowRel, strings.Join(sortedSupplied, ", "))
	}
}

// TestSiteWorkflowCallerTargetsSiteWorkflow anchors the test above: it only means
// anything while release.yml's `site` job actually calls site.yml.
func TestSiteWorkflowCallerTargetsSiteWorkflow(t *testing.T) {
	lines := workflowLines(t, releaseWorkflowRel)
	job := keyLineIndex(t, lines, callerJob, 2, 0, len(lines))
	end := blockEnd(lines, job, 2)
	for _, line := range lines[job:end] {
		if strings.TrimSpace(line) == "uses: "+calledWorkflow {
			return
		}
	}
	t.Fatalf("%s's %q job does not `uses: %s`; the input contract test is checking nothing",
		releaseWorkflowRel, callerJob, calledWorkflow)
}

// workflowLines reads a workflow from the repository root.
func workflowLines(t *testing.T, rel string) []string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(repoRoot(), filepath.FromSlash(rel)))
	if err != nil {
		t.Fatalf("read %s: %v", rel, err)
	}
	return strings.Split(string(b), "\n")
}

// declaredCallInputs returns site.yml's `on.workflow_call.inputs` as name -> required.
func declaredCallInputs(t *testing.T, lines []string) map[string]bool {
	t.Helper()
	call := keyLineIndex(t, lines, "workflow_call", 2, 0, len(lines))
	callEnd := blockEnd(lines, call, 2)
	inputs := keyLineIndex(t, lines, "inputs", 4, call, callEnd)

	declared := make(map[string]bool)
	for _, idx := range childKeyLines(lines, inputs, 4) {
		name := yamlKeyLine.FindStringSubmatch(lines[idx])[2]
		required := false
		for _, field := range childKeyLines(lines, idx, 6) {
			if strings.TrimSpace(lines[field]) == "required: true" {
				required = true
			}
		}
		declared[name] = required
	}
	return declared
}

// suppliedCallArgs returns the keys of the `with:` block on release.yml's caller job.
func suppliedCallArgs(t *testing.T, lines []string) []string {
	t.Helper()
	job := keyLineIndex(t, lines, callerJob, 2, 0, len(lines))
	jobEnd := blockEnd(lines, job, 2)
	with := keyLineIndex(t, lines, "with", 4, job, jobEnd)

	var names []string
	for _, idx := range childKeyLines(lines, with, 4) {
		names = append(names, yamlKeyLine.FindStringSubmatch(lines[idx])[2])
	}
	return names
}

// keyLineIndex finds the line index of `name:` at exactly `indent` spaces, searching
// [from, to). It fails the test rather than returning a sentinel: every caller here
// depends on the key existing, and a silent miss would turn this test green against a
// file it never read.
func keyLineIndex(t *testing.T, lines []string, name string, indent, from, to int) int {
	t.Helper()
	want := strings.Repeat(" ", indent) + name + ":"
	for i := from; i < to && i < len(lines); i++ {
		if lines[i] == want || strings.HasPrefix(lines[i], want+" ") {
			return i
		}
	}
	t.Fatalf("no `%s` key at indent %d in the searched range; the workflow changed shape", name, indent)
	return -1
}

// blockEnd returns the index one past the last line belonging to the block opened at
// lines[open], whose key sits at `indent`. The block ends at the next key indented at
// or above it.
func blockEnd(lines []string, open, indent int) int {
	for i := open + 1; i < len(lines); i++ {
		m := yamlKeyLine.FindStringSubmatch(lines[i])
		if m != nil && len(m[1]) <= indent {
			return i
		}
	}
	return len(lines)
}

// childKeyLines returns the indices of the DIRECT child keys of the block opened at
// lines[open] — those at exactly indent+2. Blank lines and comments are skipped;
// deeper lines belong to a child and are stepped over.
func childKeyLines(lines []string, open, indent int) []int {
	var out []int
	for i := open + 1; i < len(lines); i++ {
		trimmed := strings.TrimSpace(lines[i])
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		m := yamlKeyLine.FindStringSubmatch(lines[i])
		if m == nil {
			// A list item or a scalar continuation. It belongs to a child while it
			// stays deeper than this block's key; otherwise the block is over.
			if len(lines[i])-len(strings.TrimLeft(lines[i], " ")) > indent {
				continue
			}
			break
		}
		switch got := len(m[1]); {
		case got <= indent:
			return out
		case got == indent+2:
			out = append(out, i)
		}
	}
	return out
}
