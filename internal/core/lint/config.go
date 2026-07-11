package lint

import (
	"encoding/json"
	"os"
)

// Config is the on-disk record-lint configuration (.abcd/record-lint.json).
type Config struct {
	// Roots are repo-relative directories the lint walks (markdown record).
	Roots []string `json:"roots"`
	// BannedTokens are line-level substring/regex bans (check family A).
	BannedTokens []BannedToken `json:"banned_tokens"`
	// Rules holds the per-check configuration for the remaining families,
	// keyed by rule id (no_git_metadata, links_resolve, ...).
	Rules map[string]RuleConfig `json:"rules"`
	// ExemptPaths are repo-relative path prefixes whose files skip the
	// content-drift checks (banned_tokens, intent_lifecycle) — the historical,
	// non-forward-looking part of the record. Structural checks stay universal.
	ExemptPaths []string `json:"exempt_paths"`
	// ExemptIfStatus lists leading-frontmatter status: values that likewise
	// exempt a file from the content-drift checks (e.g. superseded records).
	ExemptIfStatus []string `json:"exempt_if_status"`
}

// BannedToken is one entry in the banned_tokens family (check A).
type BannedToken struct {
	ID       string `json:"id"`
	Pattern  string `json:"pattern"`
	Message  string `json:"message"`
	Severity string `json:"severity"`
	// AllowContext lists regexps that, if any matches the same line, suppress
	// the finding (the token is legitimate in that context).
	AllowContext []string `json:"allow_context"`
	// SkipCodeFences omits fenced-code lines from scanning. A nil pointer means
	// the default (true); set false to also scan inside fences.
	SkipCodeFences *bool `json:"skip_code_fences"`
}

// skipFences resolves the SkipCodeFences pointer to its effective value.
func (t BannedToken) skipFences() bool {
	if t.SkipCodeFences == nil {
		return true
	}
	return *t.SkipCodeFences
}

// RuleConfig is the shared shape for the non-token check families. Only the
// fields relevant to a given rule are populated.
type RuleConfig struct {
	Enabled  bool   `json:"enabled"`
	Severity string `json:"severity"`
	// Fields is the no_git_metadata banned frontmatter key list.
	Fields []string `json:"fields"`
	// Exempt is the directory_coverage glob allowlist.
	Exempt []string `json:"exempt"`
	// IntentsDir is the intent_lifecycle intents subdirectory (relative to a root).
	IntentsDir string `json:"intents_dir"`
	// Allowlist is the stray_root_docs permitted basename-stem list (upper-cased,
	// extension-stripped) for top-level markdown files.
	Allowlist []string `json:"allowlist"`
	// Registry is a rule's registry file, repo-relative. For persona_registry it
	// is the persona roster (.abcd/development/personas.json); for
	// surface_coverage it is the brief surface table
	// (.abcd/development/brief/04-surfaces/README.md).
	Registry string `json:"registry"`
	// CommandsDir is the surface_coverage plugin-command directory (commands/abcd);
	// each *.md file (README excepted) is a shipped command surface. It lies
	// outside Roots — the rule reads the surface tree and cross-checks the brief.
	CommandsDir string `json:"commands_dir"`
	// SkillsDir is the surface_coverage skills directory (skills); each immediate
	// subdirectory is a shipped skill surface. Also outside Roots.
	SkillsDir string `json:"skills_dir"`
	// Target is the context_status_free single-file target, repo-relative
	// (.abcd/work/CONTEXT.md). The rule runs even though the target lies outside
	// Roots; a missing target is not an error.
	Target string `json:"target"`
	// Patterns is the context_status_free line-match regexp list; when empty the
	// rule falls back to contextStatusDefaultPatterns.
	Patterns []string `json:"patterns"`
	// ReceiptsDir is the receipt_gate directory of sha-keyed semantic-pass
	// receipts (VSA-shaped JSON), repo-relative (default .abcd/work/reviews).
	// Outside Roots.
	ReceiptsDir string `json:"receipts_dir"`
	// RequiredGates lists the semantic gates that must each have a PROMOTE receipt
	// for the target commit before a release (e.g. docs-currency-reviewer,
	// iss35-brief-surface-crosscheck).
	RequiredGates []string `json:"required_gates"`
	// Commit is the receipt_gate target commit sha whose receipts are verified.
	// Release-time input (release.yml supplies the tagged commit); empty while the
	// rule is disabled for ordinary development.
	Commit string `json:"commit"`
	// Runbook is the gate_lockstep runbook path (its numbered "Deterministic
	// gates" list), repo-relative.
	Runbook string `json:"runbook"`
	// Workflow is the gate_lockstep CI workflow path — the source of truth for the
	// deterministic gate list, repo-relative.
	Workflow string `json:"workflow"`
	// Job is the gate_lockstep workflow job whose step names are the gate list.
	Job string `json:"job"`
	// IgnoreSteps are workflow step names that are setup, not gates, and so are
	// excluded from the lockstep comparison.
	IgnoreSteps []string `json:"ignore_steps"`
	// MinGates is the gate_lockstep non-empty floor: each side must parse at least
	// this many gates or the rule fails closed (an under-count means the parser or
	// a heading/job rename silently dropped gates). It is the safety net that makes
	// the hand-parse fail-closed. Enforced as at least 1 when the rule is enabled.
	MinGates int `json:"min_gates"`
}

// ArmReceiptGate returns cfg with the receipt_gate rule armed for a release: it
// is enabled and pointed at the target commit, and — when a non-empty list is
// supplied — its required gates are overridden. This is how a release runs the
// gate: the CALLER (a CI workflow) supplies the arming, so the decision to gate,
// the target commit, and the required-gates list are trust-rooted to the workflow
// rather than the in-tree, committer-editable config (phase-2 review Finding 2).
// The input cfg is not mutated (the Rules map is copied). Other rules are
// unchanged; the deterministic gates still run alongside.
func ArmReceiptGate(cfg Config, commit string, requiredGates []string) Config {
	rules := make(map[string]RuleConfig, len(cfg.Rules)+1)
	for k, v := range cfg.Rules {
		rules[k] = v
	}
	rc := rules["receipt_gate"]
	rc.Enabled = true
	rc.Commit = commit
	// An armed release gate is blocking by definition — force the severity so the
	// gate's teeth are trust-rooted to the caller (a CI workflow) like Enabled and
	// Commit, never the committer-editable config. A downgraded severity landed in
	// the in-tree file must not defang the gate at release time.
	rc.Severity = severityBlocker
	if len(requiredGates) > 0 {
		rc.RequiredGates = requiredGates
	}
	rules["receipt_gate"] = rc
	cfg.Rules = rules
	return cfg
}

// LoadConfig reads and decodes a record-lint config file.
func LoadConfig(path string) (Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, err
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return Config{}, err
	}
	return cfg, nil
}
