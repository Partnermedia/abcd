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
