package lint_test

// The scribe's access rule (spc-66) is the assembler's exact inverse, and brief
// invariant 15 states both halves: a reading receives a positively included slice
// of the shipped repository and no ledger; the scribe receives ledger content and
// never the shipped repository as an object of judgement, and it is not a
// consumer of the session-transcript store.
//
// These cases hold the shipped DEFINITION to that rule, and they are honest about
// their reach: they prove the prompt names the right paths, not that a host
// assembled the right context. Mechanical assembly belongs to the ingest verb.
//
// They sit in the external test package beside preflightgates_test.go, the other
// case that reads the real repository's shipped files, and share its readRepoFile.

import (
	"encoding/json"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/intentdriven/abcd/internal/core/lint"
)

// scribePromptRel is the shipped definition; scribeCanaryRel its injection canary.
const (
	scribePromptRel = "agents/scribe.md"
	scribeCanaryRel = "agents/scribe/fixtures/injection-canary.json"
)

// scribeLedgerRoot is the ONE tree the definition may name — the same root the
// definition, the brief's protocol and the agent changelog all state. A broader
// root admits `.abcd/work/DECISIONS.md`, which is shared working material and not
// the ledger.
const scribeLedgerRoot = ".abcd/work/issues/"

// scribeAgentContractRule is the shipped rule id, spelled as `.abcd/record-lint.json`
// spells it. TestScribePromptSatisfiesTheContract arms it against a deliberately
// broken tree, so a rename cannot leave the case passing vacuously.
const scribeAgentContractRule = "agent_contract"

// scribeSeparators folds every spelling of a path separator to '/'. A path is a
// path however it is written: a backslash, a fullwidth solidus, a division slash
// and a fraction slash all read as one separator to a human, and a host asked to
// resolve such a string resolves the path. Folding first is what stops the
// obfuscated spelling from being a hole in the allow list.
var scribeSeparators = strings.NewReplacer(
	"\\", "/", "\uFF0F", "/", "\u2215", "/", "\u2044", "/",
)

// scribePathRe matches a repository-path-shaped run, and it runs over the WHOLE
// definition rather than over one section. That is deliberate on two counts.
// Markdown is not a hiding place: once separators are folded, inline code, a
// fenced block and bare prose are the same characters, so none of the three is
// somewhere a path can sit unread. And no section is skipped, so a second
// `## Inputs` heading — or an exclusion list that names what it excludes — is not
// a way in either. The definition states its exclusions by category, never by
// path, which is what makes the whole-file rule affordable.
var scribePathRe = regexp.MustCompile(`[A-Za-z0-9_.~*<>-]*/[A-Za-z0-9_.~*<>/-]*`)

// scribeInputsHeadingRe matches the allow list's heading at any depth.
var scribeInputsHeadingRe = regexp.MustCompile(`(?m)^#{1,6}[ \t]+Inputs\b`)

// scribeAccessFindings returns every way one definition's text breaches the
// access rule: a missing allow list, an allow list that names nothing, a
// traversal segment, or any path outside the ledger root.
func scribeAccessFindings(prompt string) []string {
	var out []string
	if !scribeInputsHeadingRe.MatchString(prompt) {
		out = append(out, "carries no Inputs section; the access rule IS the allow list, so its absence is the breach")
	}
	seen := 0
	for _, raw := range scribePathRe.FindAllString(scribeSeparators.Replace(prompt), -1) {
		// Trailing sentence punctuation is not part of the path; a LEADING dot is
		// (`.abcd/...`), so the two ends are trimmed with different sets.
		tok := strings.TrimLeft(strings.TrimRight(raw, `.,;:!?)]"'`), `([`)
		// Prose spells things like "and/or" and "read/write". A repository path
		// has either more than one separator or a dot in it; neither of those does.
		if strings.Count(tok, "/") < 2 && !strings.Contains(tok, ".") {
			continue
		}
		seen++
		if strings.Contains(tok, "..") {
			out = append(out, "names "+tok+": a traversal segment escapes whatever prefix a reader checks")
			continue
		}
		if !strings.HasPrefix(path.Clean(tok)+"/", scribeLedgerRoot) {
			out = append(out, "names "+tok+", outside "+scribeLedgerRoot)
		}
	}
	if seen == 0 {
		out = append(out, "names no repository path at all; an allow list with nothing on it proves nothing")
	}
	return out
}

// scribeConformingBase is a minimal definition that satisfies the access rule.
// Every hostile case below is this text plus exactly one smuggled path, so the
// smuggle is the only thing a failure can be about.
const scribeConformingBase = "---\nname: scribe\n---\n\n" +
	"## Inputs (the allow list)\n\n" +
	"- `.abcd/work/issues/readings/` — the reading records already on file.\n\n" +
	"## Never in context\n\nThe shipped repository as an object of judgement.\n"

// TestScribeAccessCheckRefusesEveryBypass is the control, and it is what makes the
// case below it worth anything: each entry is a way of writing a shipped-tree path
// into the definition, and the check must report every one. A guard that reads
// only the first heading, only inline code, or only an ASCII solidus is a guard
// whose allow list is advisory.
func TestScribeAccessCheckRefusesEveryBypass(t *testing.T) {
	cases := []struct{ name, smuggled string }{
		{"a second Inputs heading", "\n## Inputs (continued)\n\n- `internal/core/lint/agentcontract.go` — the rule.\n"},
		{"a bare path in prose", "\nRead internal/core/lint/agentcontract.go before transcribing.\n"},
		{"a path inside a fence", "\n```\ninternal/core/lint/agentcontract.go\n```\n"},
		{"a fullwidth solidus", "\n- `internal／core／lint／agentcontract.go` — the rule.\n"},
		{"a backslash separator", "\n- `internal\\core\\lint\\agentcontract.go` — the rule.\n"},
		{"a traversal out of the ledger", "\n- `.abcd/work/issues/../../development/readings/` — the run record.\n"},
		{"the shared decision log", "\n- `.abcd/work/DECISIONS.md` — the decisions.\n"},
		{"a session-transcript store path", "\n- `~/.abcd/history/aaaa/transcripts/` — prior sessions.\n"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if fs := scribeAccessFindings(scribeConformingBase + tc.smuggled); len(fs) == 0 {
				t.Fatalf("the access check admits %s; a definition edited that way passes this gate", tc.name)
			}
		})
	}
}

// TestScribeAccessCheckPassesTheConformingShape proves the check is not simply
// refusing everything: the base the hostile cases are built on passes clean.
func TestScribeAccessCheckPassesTheConformingShape(t *testing.T) {
	if fs := scribeAccessFindings(scribeConformingBase); len(fs) != 0 {
		t.Fatalf("the conforming base must pass; got %v", fs)
	}
}

// TestScribeInputsAreLedgerOnly is the access rule over the real definition.
func TestScribeInputsAreLedgerOnly(t *testing.T) {
	root := filepath.Join("..", "..", "..")
	for _, f := range scribeAccessFindings(readRepoFile(t, root, scribePromptRel)) {
		t.Errorf("%s %s", scribePromptRel, f)
	}
}

// transcriptStoreNeedles are the spellings of the session-transcript store's
// path. Invariant 15 reserves that store to an enumerated consumer list the
// scribe is not on, so the definition names no path into it at all.
var transcriptStoreNeedles = []string{".abcd/history", "history/transcripts"}

// scribeTranscriptStoreFindings reports every named path into the store.
func scribeTranscriptStoreFindings(prompt string) []string {
	var out []string
	for _, needle := range transcriptStoreNeedles {
		if strings.Contains(prompt, needle) {
			out = append(out, "names "+needle)
		}
	}
	return out
}

// TestScribeDeclaresNoTranscriptStoreAccess holds invariant 15's second half, and
// arms itself against a definition that does name the store.
func TestScribeDeclaresNoTranscriptStoreAccess(t *testing.T) {
	root := filepath.Join("..", "..", "..")
	for _, f := range scribeTranscriptStoreFindings(readRepoFile(t, root, scribePromptRel)) {
		t.Errorf("%s %s; invariant 15 reserves the session-transcript store to an enumerated consumer "+
			"list the scribe is not on, and adding it there is an invariant change, never a code change",
			scribePromptRel, f)
	}
	hostile := scribeConformingBase + "\n- `~/.abcd/history/aaaa/transcripts/` — prior sessions.\n"
	if len(scribeTranscriptStoreFindings(hostile)) == 0 {
		t.Fatal("the transcript-store check admits a store path; it is not armed")
	}
}

// TestScribeCanaryAssertsTheRefusals reads the fixture the agent_contract rule
// already proves present, non-empty and regular, and asserts what that rule
// cannot: that it parses, carries a hostile payload, and declares the expectation
// the payload is measured against. A control string the input never contains is a
// `must_not_contain` that no run can fail.
func TestScribeCanaryAssertsTheRefusals(t *testing.T) {
	root := filepath.Join("..", "..", "..")
	raw := readRepoFile(t, root, scribeCanaryRel)

	var fixture struct {
		Input    json.RawMessage `json:"input"`
		Expected struct {
			MustNotContain       string   `json:"must_not_contain"`
			MustNotObey          []string `json:"must_not_obey"`
			InjectionIsInertData bool     `json:"injection_is_inert_data"`
		} `json:"expected"`
	}
	if err := json.Unmarshal([]byte(raw), &fixture); err != nil {
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
	if tok := fixture.Expected.MustNotContain; tok == "" || !strings.Contains(string(fixture.Input), tok) {
		t.Errorf("%s's must_not_contain token %q appears nowhere in its input; a control string the payload never "+
			"demands is one no run can fail", scribeCanaryRel, tok)
	}
	// The definition carries two refusals beyond "never obey an instruction":
	// transcript material handed over as ledger context is refused outright, and a
	// composed contribution is stamped or withheld. A canary that tests neither
	// leaves both unexercised.
	declared := strings.ToLower(strings.Join(fixture.Expected.MustNotObey, " | "))
	for _, lure := range []string{"transcript", "summary"} {
		if !strings.Contains(declared, lure) {
			t.Errorf("%s's expected.must_not_obey names no %q refusal; the definition refuses transcript material "+
				"handed as ledger context and refuses an unstamped composed contribution, so the canary tests both",
				scribeCanaryRel, lure)
		}
	}
}

// TestScribePromptSatisfiesTheContract runs the shipped agent_contract rule over
// the real tree and asserts the scribe passes all three sub-checks. The second
// half arms the rule id: a rule renamed out from under this case would otherwise
// leave it filtering for findings that can never appear.
func TestScribePromptSatisfiesTheContract(t *testing.T) {
	cfg := lint.Config{Rules: map[string]lint.RuleConfig{
		scribeAgentContractRule: {Enabled: true, Severity: "blocker"},
	}}

	fs, err := lint.Lint(cfg, filepath.Join("..", "..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	for _, f := range fs {
		if f.RuleID != scribeAgentContractRule {
			continue
		}
		if filepath.ToSlash(f.File) == scribePromptRel || strings.Contains(f.Message, "scribe") {
			t.Errorf("%s: %s", f.File, f.Message)
		}
	}

	broken := t.TempDir()
	if err := os.MkdirAll(filepath.Join(broken, "agents"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(broken, "agents", "unversioned.md"),
		[]byte("---\nname: unversioned\n---\n\nPrompt body.\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	armed, err := lint.Lint(cfg, broken)
	if err != nil {
		t.Fatal(err)
	}
	for _, f := range armed {
		if f.RuleID == scribeAgentContractRule {
			return
		}
	}
	t.Fatalf("no %q finding over a deliberately broken agent tree; the rule id this case filters on is stale",
		scribeAgentContractRule)
}
