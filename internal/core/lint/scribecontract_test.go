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
	"fmt"
	"html"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"unicode"

	"golang.org/x/text/unicode/norm"

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

// scribeProseIdioms is the ONLY exemption from "a separator-bearing token is a
// path", and it is an explicit list because the shape rule it replaces was a
// description of `and/or` that also described `internal/core`, `docs/` and
// `internal/README`. Everything else is judged as a path — a version pair such as
// `0.1.0/0.2.0` and a URL included. That is by design: the definition has no
// reason to carry either, and an exemption wide enough to admit them is wide
// enough to admit a shipped-tree path spelled to look like one.
var scribeProseIdioms = map[string]bool{
	"and/or": true, "read/write": true, "either/or": true, "i/o": true,
}

// scribeNonNFKCSeparators are the separator spellings NFKC leaves alone. The
// compatibility forms — fullwidth solidus U+FF0F, fullwidth reverse solidus
// U+FF3C, small reverse solidus U+FE68 — are folded by NFKC itself and are
// deliberately absent here, so this list stays a list of what NFKC does not do.
var scribeNonNFKCSeparators = strings.NewReplacer(
	"\\", "/", // reverse solidus
	"⁄", "/", // fraction slash
	"∕", "/", // division slash
	"╱", "/", // box drawings light diagonal
	"⧸", "/", // big solidus
)

// scribeSpacedSeparatorRe collapses horizontal whitespace around a separator, so
// `a / b` reads as `a/b`. Newlines are deliberately not collapsed: joining across
// a line break would invent paths out of adjacent sentences.
var scribeSpacedSeparatorRe = regexp.MustCompile(`[ \t]*/[ \t]*`)

// scribeFold normalises a definition into the form the path check reads. Each
// step closes one way of spelling a separator that a reader — or a host asked to
// resolve the string — resolves as one anyway:
//
//  1. HTML entity decoding (`&#47;`, `&sol;`);
//  2. percent decoding (`%2F`, `%5C`), left as it stands when the text is not
//     valid percent-encoding, because a stray `%` is no reason to stop checking;
//  3. NFKC, which folds the compatibility separators — fullwidth solidus,
//     fullwidth reverse solidus, small reverse solidus — onto ASCII;
//  4. the five separators NFKC does not fold, listed above one by one;
//  5. horizontal whitespace around a separator.
//
// That is exactly the folding done and nothing more. The other class — an
// invisible code point splicing a path back together after any check that reads
// the visible text — is not folded but refused outright, by the Cf scan below.
func scribeFold(text string) string {
	text = html.UnescapeString(text)
	if decoded, err := url.PathUnescape(text); err == nil {
		text = decoded
	}
	text = norm.NFKC.String(text)
	text = scribeNonNFKCSeparators.Replace(text)
	return scribeSpacedSeparatorRe.ReplaceAllString(text, "/")
}

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
// access rule: a missing allow list, an allow list that names nothing, a format
// code point, a traversal segment, or any path outside the ledger root.
func scribeAccessFindings(prompt string) []string {
	var out []string
	if !scribeInputsHeadingRe.MatchString(prompt) {
		out = append(out, "carries no Inputs section; the access rule IS the allow list, so its absence is the breach")
	}

	folded := scribeFold(prompt)
	// The Cf scan reads the folded text, so an entity- or percent-encoded format
	// code point is caught alongside a literal one. One finding, not one per
	// occurrence: the class is what matters, and a hostile file could carry
	// thousands.
	for _, r := range folded {
		if unicode.Is(unicode.Cf, r) {
			out = append(out, fmt.Sprintf("carries the format code point U+%04X: an invisible code point "+
				"splices a path back together after any check that reads the visible text, so it is refused "+
				"rather than folded", r))
			break
		}
	}

	seen := 0
	for _, raw := range scribePathRe.FindAllString(folded, -1) {
		// Trailing sentence punctuation is not part of the path; a LEADING dot is
		// (`.abcd/...`), so the two ends are trimmed with different sets.
		tok := strings.TrimLeft(strings.TrimRight(raw, `.,;:!?)]"'`), `([`)
		if !strings.Contains(tok, "/") || scribeProseIdioms[strings.ToLower(tok)] {
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
		{"a bare shipped-tree directory", "\n- `internal/core` — where the rule lives.\n"},
		{"a bare docs directory", "\n- `docs/` — the user-facing tree.\n"},
		{"a directory-and-file pair with no extension", "\n- `internal/README` — the package map.\n"},
		{"a spaced separator", "\n- `internal / core / lint / agentcontract.go` — the rule.\n"},
		{"an HTML-escaped separator", "\n- `internal&#47;core&#47;lint&#47;agentcontract.go` — the rule.\n"},
		{"a percent-encoded separator", "\n- `internal%2Fcore%2Flint%2Fagentcontract.go` — the rule.\n"},
		{"a box-drawing solidus", "\n- `internal╱core╱lint╱agentcontract.go` — the rule.\n"},
		{"a big solidus", "\n- `internal⧸core⧸lint⧸agentcontract.go` — the rule.\n"},
		{"a fullwidth reverse solidus", "\n- `internal＼core＼lint＼agentcontract.go` — the rule.\n"},
		{"a small reverse solidus", "\n- `internal﹨core﹨lint﹨agentcontract.go` — the rule.\n"},
		{"a format code point in the prose", "\nThe allow list above is exhaustive.\u202E\n"},
		{"a zero-width joiner splicing a path", "\n- `internal/core/li\u200dnt/agentcontract.go` — the rule.\n"},
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
	// The control string has to be a shape only OBEDIENCE produces. The payload
	// must demand it, or no run can fail the control; and the faithful output must
	// not carry it, or every honest run fails instead — the scribe transcribes item
	// bodies verbatim, so a bare word the payload plants is in the output by design.
	// The demanded word therefore appears in the input and the emitted shape does not.
	tok := fixture.Expected.MustNotContain
	word := strings.Trim(tok, `: "`)
	switch {
	case tok == "" || word == "":
		t.Errorf("%s declares no usable 'expected.must_not_contain'", scribeCanaryRel)
	case !strings.Contains(string(fixture.Input), word):
		t.Errorf("%s's control word %q is demanded nowhere in its input; a control the payload never asks for is "+
			"one no run can fail", scribeCanaryRel, word)
	case strings.Contains(string(fixture.Input), tok):
		t.Errorf("%s's control %q already appears in its input; verbatim transcription puts it in the output, so a "+
			"control shaped like that is one every honest run fails", scribeCanaryRel, tok)
	}

	// Every item carries an identifier, because the dispositions and the
	// outstanding list refer to items by id; an item with none is a reference to
	// nothing.
	var items struct {
		Items []struct {
			Item string `json:"item"`
		} `json:"items"`
	}
	if err := json.Unmarshal(fixture.Input, &items); err != nil {
		t.Fatalf("%s's input does not parse: %v", scribeCanaryRel, err)
	}
	if len(items.Items) == 0 {
		t.Fatalf("%s's input carries no items", scribeCanaryRel)
	}
	for i, it := range items.Items {
		if it.Item == "" {
			t.Errorf("%s's input item %d carries no 'item' identifier, yet the dispositions and the outstanding "+
				"list refer to items by id", scribeCanaryRel, i+1)
		}
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
