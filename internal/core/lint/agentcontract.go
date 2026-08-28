package lint

// The agent trust-contract family (agent_contract): the itd-5 prompt-quality
// contract, enforced.
//
// `agents/` holds host-delegated PROMPTS — the one part of the shipped surface a
// model reads as instruction — and it sits in neither lint root (record-lint
// walks .abcd/development, docs-lint walks docs/ and README.md). agents/README.md
// has documented the contract since M6 and named the linter that would enforce it
// as not yet built, so the five prompts that read the most attacker-influenceable
// input in the repository acquired the contract by hand and nothing checked that
// the sixth would (iss-278).
//
// This is that detector. It is a dedicated rule rather than an extra entry in
// cfg.Roots: the record stores' schema rules judge a RECORD's shape, and an agent
// prompt is not a record — it has a different frontmatter, a different lifecycle,
// and its own changelog. So the rule enumerates the tree itself, in the
// once-outside-the-loop, repo-root-scoped style of checkStrayRootDocs and
// checkDeliveryState.
//
// Three sub-checks, per agents/README.md § The itd-5 contract:
//
//  1. the trust-contract frontmatter (prompt_version, reads_untrusted_input,
//     capability_scope.task_classes, capability_scope.designed_for);
//  2. the injection-canary fixture every untrusted-input agent must ship;
//  3. a per-agent CHANGELOG entry for an agent added or changed in a diff.

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/intentdriven/abcd/internal/fsutil"
	"github.com/intentdriven/abcd/internal/gitutil"
)

const ruleAgentContract = "agent_contract"

// defaultAgentsDir is the prompt tree, repo-relative. It is a directory of
// PROMPTS, not of records, which is why it is named here rather than in
// cfg.Roots.
const defaultAgentsDir = "agents"

// agentCanaryFixture is the per-agent injection-canary contract: an agent that
// reads attacker-influenceable input carries at least one, under its own
// fixtures directory (agents/README.md § Injection canaries).
const agentCanaryFixture = "injection-canary.json"

var (
	// A prompt_version is a semver, because the changelog entry is keyed on it:
	// a version that cannot be spelled cannot be reconciled with an entry, and
	// the bump grammar (MAJOR schema break / MINOR behaviour / PATCH edit) is
	// what the entry is FOR.
	promptVersionRe = regexp.MustCompile(`^\d+\.\d+\.\d+(?:[-+][0-9A-Za-z.-]+)?$`)
	// A nested frontmatter key: the `capability_scope` block's members sit one
	// level in, and the shared top-level scanner (internal/core/frontmatter)
	// deliberately ignores them. This is the ONLY nesting the contract has, so it
	// is read here rather than by promoting the whole package to a YAML parser.
	agentNestedKeyRe = regexp.MustCompile(`^[ \t]+([A-Za-z0-9_]+)[ \t]*:(.*)$`)
	// A caller-supplied diff range reaches a git subprocess, so it is validated
	// as a revision (or revision range) before use: a leading '-' would be read
	// as an OPTION by git, which is argument injection into a gate that a CI
	// workflow arms. The class is deliberately narrow — the spellings a range
	// actually takes (sha, ref, HEAD~2, a..b, a...b) and nothing else.
	agentDiffRangeRe = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._/@^~-]*(?:\.{2,3}[A-Za-z0-9][A-Za-z0-9._/@^~-]*)?$`)
)

// maxAgentPromptBytes caps a prompt read. A prompt is a few kilobytes of
// markdown; the cap bounds a hostile tree without ever constraining a real one,
// exactly as maxLintConfigBytes does for the config.
const maxAgentPromptBytes = 1 << 20

// agentPrompt is one agent prompt as this rule sees it.
type agentPrompt struct {
	// name is the file stem, which is the agent's identity: the fixtures
	// directory and the changelog entry are both keyed on it.
	name string
	// rel is the prompt file, repo-relative.
	rel string
	// fields are the top-level frontmatter keys.
	fields map[string]fmField
	// scope are the members of the nested capability_scope block.
	scope map[string]string
}

// checkAgentContract enforces the itd-5 trust contract over the agent-prompt
// tree. A tree that is not there contributes nothing and is not an error — a
// repository with no agents is a state, not a fault.
func checkAgentContract(repoRoot string, cfg RuleConfig) ([]Finding, error) {
	dir := cfg.AgentsDir
	if dir == "" {
		dir = defaultAgentsDir
	}
	if err := containedRepoPath(dir); err != nil {
		return nil, &configError{ruleAgentContract + ": agents_dir " + quote(dir) + " " + err.Error() +
			"; the lint reads only inside the repository"}
	}
	dirAbs := filepath.Join(repoRoot, filepath.FromSlash(dir))
	entries, err := os.ReadDir(dirAbs)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var prompts []agentPrompt
	var out []Finding
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() || !hasMarkdownExt(name) {
			continue
		}
		// README.md and CHANGELOG.md are the tree's own prose, not prompts.
		stem := strings.TrimSuffix(name, filepath.Ext(name))
		if strings.EqualFold(stem, "README") || strings.EqualFold(stem, "CHANGELOG") {
			continue
		}
		rel := filepath.Join(dir, name)
		realPath, err := containedRealPath(repoRoot, filepath.Join(dirAbs, name))
		if err != nil {
			return nil, &configError{ruleAgentContract + ": prompt " + quote(rel) + " " + err.Error() +
				"; the lint reads only inside the repository"}
		}
		content, err := fsutil.ReadGuarded(realPath, maxAgentPromptBytes)
		if err != nil {
			return nil, err
		}
		lines := strings.Split(string(content), "\n")
		prompts = append(prompts, agentPrompt{
			name:   stem,
			rel:    rel,
			fields: frontmatterFields(lines),
			scope:  agentCapabilityScope(lines),
		})
	}

	for _, p := range prompts {
		out = append(out, checkAgentTrustContract(repoRoot, dir, p, cfg.Severity)...)
	}

	changelogFindings, err := checkAgentChangelog(repoRoot, dir, prompts, cfg)
	if err != nil {
		return nil, err
	}
	return append(out, changelogFindings...), nil
}

// checkAgentTrustContract is sub-checks 1 and 2 for one prompt: the itd-5
// frontmatter, and the injection canary an untrusted-input agent must ship.
//
// The DECLARATION is required of every prompt, not only of the ones that admit
// they read untrusted input. A rule that fires only on `reads_untrusted_input:
// true` is a rule any new agent opts out of by deleting one line — and the class
// this exists to catch is precisely a prompt that reads attacker-influenceable
// input while saying nothing about it. Silence is not `false`; it is undeclared,
// and it is reported.
func checkAgentTrustContract(repoRoot, dir string, p agentPrompt, severity string) []Finding {
	add := func(msg string) Finding {
		return Finding{File: p.rel, Line: 1, RuleID: ruleAgentContract, Severity: severity, Message: msg}
	}
	var out []Finding

	declared, ok := p.fields["reads_untrusted_input"]
	if !ok {
		return append(out, add("agent prompt declares no 'reads_untrusted_input': the itd-5 trust contract "+
			"(agents/README.md) is declared, never inferred — an undeclared prompt reads as safe to every reader "+
			"and to this gate, which is how a prompt that reads attacker-influenceable input ships without a canary. "+
			"Declare 'reads_untrusted_input: true' (and carry the contract fields) or 'false'"))
	}

	// prompt_version is the itd-5 contract for EVERY prompt: it is what the
	// per-agent changelog entry is keyed on, whatever the prompt reads.
	if v, ok := p.fields["prompt_version"]; !ok || strings.TrimSpace(v.value) == "" {
		out = append(out, add("agent prompt is missing 'prompt_version': the itd-5 contract versions every prompt, "+
			"and the per-agent CHANGELOG entry is keyed on it"))
	} else if !promptVersionRe.MatchString(strings.Trim(strings.TrimSpace(v.value), `"'`)) {
		out = append(out, add("agent prompt's 'prompt_version' is not a semver ("+quote(v.value)+
			"): the bump grammar (MAJOR schema break, MINOR behaviour, PATCH edit) and the CHANGELOG entry both key on it"))
	}

	if !isTrueValue(declared.value) {
		return out
	}

	// The untrusted-input contract: the capability scope, then the canary.
	if len(p.scope) == 0 {
		out = append(out, add("agent prompt reads untrusted input but declares no 'capability_scope': the itd-5 "+
			"contract requires 'capability_scope.task_classes' and 'capability_scope.designed_for'"))
	} else {
		if strings.TrimSpace(p.scope["task_classes"]) == "" {
			out = append(out, add("agent prompt reads untrusted input but is missing 'capability_scope.task_classes': "+
				"an agent whose scope is undeclared has no bound on what it may be dispatched for"))
		}
		if strings.TrimSpace(p.scope["designed_for"]) == "" {
			out = append(out, add("agent prompt reads untrusted input but is missing 'capability_scope.designed_for': "+
				"the one-line statement of what the prompt was built to do"))
		}
	}

	canaryRel := filepath.Join(dir, p.name, "fixtures", agentCanaryFixture)
	if _, err := os.Stat(filepath.Join(repoRoot, canaryRel)); err != nil {
		out = append(out, add("agent prompt reads untrusted input but ships no "+agentCanaryFixture+
			" fixture (expected at "+filepath.ToSlash(canaryRel)+"): the canary is the only thing that asserts the "+
			"prompt quotes hostile input as data rather than obeying it"))
	}
	return out
}

// checkAgentChangelog is sub-check 3: every agent ADDED OR CHANGED in the diff
// under lint must have gained its per-agent CHANGELOG entry in the same change.
//
// It is diff-scoped by construction, and that is the whole design: the invariant
// is "a prompt change is announced", which is a statement about a CHANGE, not
// about a tree. A full-tree lint has no diff to make it about, so the sub-check
// is a no-op there rather than a guess — the frontmatter and canary checks still
// run, because those ARE statements about a tree.
//
// The range is supplied by the CALLER (ArmAgentDiff, from a CI invocation), never
// read out of the in-tree config, for the reason ArmReceiptGate states: a gate a
// committer can point at an empty range is a gate a committer can disarm.
func checkAgentChangelog(repoRoot, dir string, prompts []agentPrompt, cfg RuleConfig) ([]Finding, error) {
	if cfg.DiffRange == "" {
		return nil, nil
	}
	if !agentDiffRangeRe.MatchString(cfg.DiffRange) {
		return nil, &configError{ruleAgentContract + ": diff range " + quote(cfg.DiffRange) +
			" is not a git revision or revision range; a value git would read as an option is refused"}
	}
	if !gitutil.InRepo(repoRoot) {
		return nil, nil
	}

	changelogRel := cfg.Changelog
	if changelogRel == "" {
		changelogRel = filepath.ToSlash(filepath.Join(dir, "CHANGELOG.md"))
	}
	changed, err := changedPaths(repoRoot, cfg.DiffRange)
	if err != nil {
		return nil, &configError{ruleAgentContract + ": reading the diff for " + quote(cfg.DiffRange) +
			": " + err.Error()}
	}

	data, err := os.ReadFile(filepath.Join(repoRoot, filepath.FromSlash(changelogRel)))
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}
	entries := agentChangelogEntries(string(data))

	var out []Finding
	for _, p := range prompts {
		if !changed[filepath.ToSlash(p.rel)] {
			continue
		}
		version := strings.Trim(strings.TrimSpace(p.fields["prompt_version"].value), `"'`)
		if version == "" {
			continue // already reported by the frontmatter sub-check
		}
		if entries[p.name+" "+version] {
			continue
		}
		out = append(out, Finding{
			File: changelogRel, Line: 0, RuleID: ruleAgentContract, Severity: cfg.Severity,
			Message: "agent '" + p.name + "' changed in this diff but " + changelogRel +
				" carries no entry for " + p.name + " " + version +
				"; the itd-5 contract announces every prompt change in the same change (add a '### " +
				p.name + " " + version + "' entry, bumping prompt_version if the change is behavioural)",
		})
	}
	return out, nil
}

// agentChangelogHeadingRe matches a per-agent entry heading — "### <name>
// <semver>", with anything after (the changelog annotates a rename in the same
// heading). The heading level is not pinned: an entry is an entry at any depth.
var agentChangelogHeadingRe = regexp.MustCompile(`^#{1,6}[ \t]+([a-z0-9][a-z0-9-]*)[ \t]+v?(\d+\.\d+\.\d+)\b`)

// agentChangelogEntries indexes a changelog by "<agent> <version>".
func agentChangelogEntries(text string) map[string]bool {
	entries := map[string]bool{}
	for _, line := range strings.Split(text, "\n") {
		if m := agentChangelogHeadingRe.FindStringSubmatch(line); m != nil {
			entries[m[1]+" "+m[2]] = true
		}
	}
	return entries
}

// changedPaths returns the slash-separated repo-relative paths the range touched.
// NUL-delimited (-z) so a path git would otherwise quote is read verbatim, and
// `--` terminates the revision list so a range that somehow survived validation
// still cannot be read as a pathspec.
func changedPaths(repoRoot, rangeSpec string) (map[string]bool, error) {
	out, err := gitutil.Run(repoRoot, "diff", "--name-only", "-z", rangeSpec, "--")
	if err != nil {
		return nil, err
	}
	changed := map[string]bool{}
	for _, p := range strings.Split(out, "\x00") {
		if p != "" {
			changed[p] = true
		}
	}
	return changed, nil
}

// agentCapabilityScope reads the members of the nested `capability_scope` block
// out of a prompt's leading frontmatter. The block ends at the first line that is
// not indented (the next top-level key, or the closing delimiter).
func agentCapabilityScope(lines []string) map[string]string {
	if start := frontmatterOpen(lines); start > 0 {
		lines = lines[start:]
	}
	if len(lines) == 0 || strings.TrimSpace(lines[0]) != "---" {
		return nil
	}
	scope := map[string]string{}
	inScope := false
	for i := 1; i < len(lines); i++ {
		line := strings.TrimRight(lines[i], "\r")
		if strings.TrimSpace(line) == "---" {
			break
		}
		if m := agentNestedKeyRe.FindStringSubmatch(line); m != nil {
			if inScope {
				if _, seen := scope[m[1]]; !seen {
					scope[m[1]] = strings.TrimSpace(m[2])
				}
			}
			continue
		}
		inScope = strings.HasPrefix(line, "capability_scope:")
	}
	if !inScope && len(scope) == 0 {
		return nil
	}
	return scope
}

// isTrueValue reads a frontmatter boolean the way the prompts spell it. Only the
// affirmative spellings count: anything else (including an empty value) is not a
// declaration that the prompt reads untrusted input.
func isTrueValue(v string) bool {
	switch strings.ToLower(strings.Trim(strings.TrimSpace(v), `"'`)) {
	case "true", "yes":
		return true
	}
	return false
}
