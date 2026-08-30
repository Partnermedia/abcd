package lint

// The scribe's access rule (spc-66) is the assembler's exact inverse, and brief
// invariant 15 states both halves: a reading receives a positively included slice
// of the shipped repository and no ledger; the scribe receives ledger content and
// never the shipped repository as an object of judgement, and it is not a
// consumer of the session-transcript store.
//
// These cases hold the shipped DEFINITION to that rule, and they are honest about
// their reach: they prove the prompt says the right thing, not that a host
// assembled the right context. Mechanical assembly belongs to the ingest verb,
// which is a later cycle's.
//
// They live in this package because it is the one that already reads the agent
// tree (agentcontract.go), so no second reader of `agents/` is written.

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// scribePromptRel is the shipped definition, repo-relative.
const scribePromptRel = "agents/scribe.md"

// scribeCanaryRel is the injection canary the itd-5 contract requires of it.
const scribeCanaryRel = "agents/scribe/fixtures/injection-canary.json"

// scribeLedgerRoot is the ONE tree the scribe's inputs may name. The rule is
// positive inclusion, inherited from the assembler: a path outside this prefix is
// excluded whether or not anyone thought to exclude it.
const scribeLedgerRoot = ".abcd/work/"

// scribeInputsHeading opens the allow list. The exclusions live under their own
// heading and name outside paths on purpose, so the allow-list check reads this
// section alone.
const scribeInputsHeading = "## Inputs"

// backtickedRe captures every inline-code token; a token carrying a path
// separator is read as a repository path.
var backtickedRe = regexp.MustCompile("`([^`\n]+)`")

// readRepoFile reads one file out of the real repository.
func readRepoFile(t *testing.T, rel string) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(repoRootFromPackage, filepath.FromSlash(rel)))
	if err != nil {
		t.Fatalf("reading %s: %v", rel, err)
	}
	return string(data)
}

// markdownSection returns the lines under the first heading with the given
// prefix, up to the next second-level heading.
func markdownSection(text, headingPrefix string) string {
	lines := strings.Split(text, "\n")
	start := -1
	for i, line := range lines {
		if strings.HasPrefix(line, headingPrefix) {
			start = i + 1
			break
		}
	}
	if start < 0 {
		return ""
	}
	for i := start; i < len(lines); i++ {
		if strings.HasPrefix(lines[i], "## ") {
			return strings.Join(lines[start:i], "\n")
		}
	}
	return strings.Join(lines[start:], "\n")
}

// pathsOutsideLedger returns every repository path a section names that does not
// sit under the ledger root, and the count of paths it saw at all. A section that
// names no path is as much a failure as one that names the wrong path: an allow
// list with nothing on it allows nothing and proves nothing.
func pathsOutsideLedger(section string) (outside []string, seen int) {
	for _, m := range backtickedRe.FindAllStringSubmatch(section, -1) {
		tok := strings.TrimSpace(m[1])
		if !strings.Contains(tok, "/") {
			continue
		}
		seen++
		if !strings.HasPrefix(tok, scribeLedgerRoot) {
			outside = append(outside, tok)
		}
	}
	return outside, seen
}

// transcriptStoreNeedles are the spellings of the session-transcript store's
// path. Invariant 15 reserves that store to an enumerated consumer list the
// scribe is not on, so the definition must not name a path into it at all.
var transcriptStoreNeedles = []string{".abcd/history", "history/transcripts", "/history/"}

// TestScribeInputsAreLedgerOnly is the access rule stated positively: the
// definition's allow list names ledger paths and nothing else. The control case
// below it proves the check is armed rather than vacuously passing.
func TestScribeInputsAreLedgerOnly(t *testing.T) {
	section := markdownSection(readRepoFile(t, scribePromptRel), scribeInputsHeading)
	if strings.TrimSpace(section) == "" {
		t.Fatalf("%s carries no %q section; the access rule is the allow list, so its absence is the failure",
			scribePromptRel, scribeInputsHeading)
	}

	outside, seen := pathsOutsideLedger(section)
	if seen == 0 {
		t.Fatalf("%s's inputs section names no repository path; an allow list with nothing on it proves nothing",
			scribePromptRel)
	}
	if len(outside) != 0 {
		t.Errorf("%s's inputs section names %v, outside %s; the scribe never receives the shipped repository as an object of judgement",
			scribePromptRel, outside, scribeLedgerRoot)
	}
}

// TestScribeInputsCheckRefusesAnOutsidePath is the control: the same extractor,
// over a section that names a shipped-tree path, must report it. Without this the
// case above passes for a definition that names no path in a form the extractor
// recognises.
func TestScribeInputsCheckRefusesAnOutsidePath(t *testing.T) {
	section := "- `.abcd/work/issues/` — the ledger.\n- `internal/core/lint/agentcontract.go` — the rule.\n"
	outside, seen := pathsOutsideLedger(section)
	if seen != 2 {
		t.Fatalf("expected both paths to be seen, got %d", seen)
	}
	if len(outside) != 1 || outside[0] != "internal/core/lint/agentcontract.go" {
		t.Fatalf("expected the shipped-tree path to be refused, got %v", outside)
	}
}

// TestScribeDeclaresNoTranscriptStoreAccess holds invariant 15's second half: the
// scribe is not a transcript consumer, so its definition names no path into the
// session-transcript store — not as an input, and not anywhere else, because a
// path in a prompt is a path a host may be asked to supply.
func TestScribeDeclaresNoTranscriptStoreAccess(t *testing.T) {
	prompt := readRepoFile(t, scribePromptRel)
	for _, needle := range transcriptStoreNeedles {
		if strings.Contains(prompt, needle) {
			t.Errorf("%s names %q; invariant 15 reserves the session-transcript store to an enumerated "+
				"consumer list the scribe is not on, and adding it there is an invariant change, never a code change",
				scribePromptRel, needle)
		}
	}
	// The control: the needle set must actually match a store path.
	if !strings.Contains("the store at ~/.abcd/history/<root-sha>/transcripts/", transcriptStoreNeedles[0]) {
		t.Fatal("the transcript-store needle set matches no store path; the check is not armed")
	}
}

// TestScribeCanaryIsPresentAndNonEmpty reads the fixture the agent_contract rule
// requires and asserts what that rule cannot: that it parses, carries a hostile
// input, and declares the expectation that the demand is transcribed as data and
// never obeyed. A canary that is merely a non-empty file reports the contract met
// without testing it.
func TestScribeCanaryIsPresentAndNonEmpty(t *testing.T) {
	info, err := os.Lstat(filepath.Join(repoRootFromPackage, filepath.FromSlash(scribeCanaryRel)))
	if err != nil {
		t.Fatalf("stat %s: %v", scribeCanaryRel, err)
	}
	if !info.Mode().IsRegular() || info.Size() == 0 {
		t.Fatalf("%s must be a regular, non-empty file", scribeCanaryRel)
	}

	var fixture struct {
		Input    json.RawMessage `json:"input"`
		Expected struct {
			MustNotObey          []string `json:"must_not_obey"`
			InjectionIsInertData bool     `json:"injection_is_inert_data"`
		} `json:"expected"`
	}
	if err := json.Unmarshal([]byte(readRepoFile(t, scribeCanaryRel)), &fixture); err != nil {
		t.Fatalf("%s does not parse as JSON: %v", scribeCanaryRel, err)
	}
	if len(fixture.Input) == 0 {
		t.Errorf("%s carries no 'input'; a canary with no hostile payload asserts nothing", scribeCanaryRel)
	}
	if len(fixture.Expected.MustNotObey) == 0 {
		t.Errorf("%s declares no 'expected.must_not_obey'; the expectation is what the canary asserts", scribeCanaryRel)
	}
	if !fixture.Expected.InjectionIsInertData {
		t.Errorf("%s does not declare 'expected.injection_is_inert_data'; that is the contract every canary asserts",
			scribeCanaryRel)
	}
}

// TestScribePromptSatisfiesTheContract runs the shipped agent_contract rule over
// the real tree and asserts the scribe passes all three sub-checks — the itd-5
// frontmatter, the canary, and the per-agent changelog entry keyed on its
// version. Findings for the other prompts are not this case's business.
func TestScribePromptSatisfiesTheContract(t *testing.T) {
	if _, err := os.Stat(filepath.Join(repoRootFromPackage, filepath.FromSlash(scribePromptRel))); err != nil {
		t.Fatalf("stat %s: %v", scribePromptRel, err)
	}

	fs, err := Lint(agentCfg(), repoRootFromPackage)
	if err != nil {
		t.Fatal(err)
	}
	for _, f := range fs {
		if f.RuleID != ruleAgentContract {
			continue
		}
		if filepath.ToSlash(f.File) == scribePromptRel || strings.Contains(f.Message, "scribe") {
			t.Errorf("%s: %s", f.File, f.Message)
		}
	}
}
