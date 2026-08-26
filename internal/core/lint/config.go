package lint

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"

	"github.com/intentdriven/abcd/internal/fsutil"
)

// maxLintConfigBytes caps the docs-lint/record-lint config read (trust
// boundary). The shipped configs are a few kilobytes; the cap is many times
// larger, so it bounds a hostile /dev/zero-symlinked config without ever
// constraining a real one.
const maxLintConfigBytes = 256 * 1024

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
	// content-AUTHORING checks (banned_tokens, persona_registry) — the historical,
	// non-forward-looking part of the record, which is excused from how it is
	// written but never from being well-formed. Both intent-tree checks
	// (intent_lifecycle, intent_impact_valid — they share one scan) stay universal
	// (iss-39); the spec-store checks (spec_lifecycle, spec_id_unique) still skip an
	// exempt file. record_schema is cross-store and never consults this at all.
	ExemptPaths []string `json:"exempt_paths"`
	// ExemptIfStatus lists leading-frontmatter status: values that likewise
	// exempt a file from the content-authoring checks (e.g. superseded records).
	ExemptIfStatus []string `json:"exempt_if_status"`
}

// BannedToken is one entry in the banned_tokens family (check A).
type BannedToken struct {
	ID       string `json:"id"`
	Pattern  string `json:"pattern"`
	Message  string `json:"message"`
	Severity string `json:"severity"`
	// Successor is the machine-readable old->new mapping: what to use instead of
	// the banned token. It is REQUIRED (a ban with no successor left its
	// replacement in prose only, iss-51) and is auto-cited in the finding message.
	Successor string `json:"successor"`
	// AllowContext lists regexps that, if any matches the same line, suppress
	// the finding (the token is legitimate in that context). It is REQUIRED to be
	// non-empty: every ban must declare where its token is legitimately allowed.
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
	// IntentsDir is the intents subdirectory (relative to a root) read by the
	// intent-tree rules, intent_lifecycle and intent_impact_valid. Rules that name
	// the same directory share one scan of it. spec_lifecycle also reads it to
	// resolve the intent corpus its specs link to.
	IntentsDir string `json:"intents_dir"`
	// SpecsDir is the spec_lifecycle specs subdirectory (relative to a root),
	// mirroring IntentsDir. Default "specs".
	SpecsDir string `json:"specs_dir"`
	// IssuesDir is the issue-ledger root (repo-relative) read by the ledger rules,
	// issue_id_unique and issue_impact_valid; it holds the open/, resolved/, and
	// wontfix/ status directories. Default .abcd/work/issues. It lies outside Roots
	// — the rules read the ledger and run once, sharing one scan of it.
	IssuesDir string `json:"issues_dir"`
	// Allowlist is the stray_root_docs permitted basename-stem list (upper-cased,
	// extension-stripped) for top-level markdown files.
	Allowlist []string `json:"allowlist"`
	// Registry is a rule's registry file, repo-relative. For persona_registry it
	// is the persona roster (.abcd/development/personas.json); for
	// surface_coverage it is the brief surface table
	// (.abcd/development/brief/04-surfaces/README.md).
	Registry string `json:"registry"`
	// CommandsDir is the surface_coverage plugin-command directory (commands);
	// each *.md file (README and BareCommand excepted) is a shipped command
	// surface. It lies outside Roots — the rule reads the surface tree and
	// cross-checks the brief. The directory is flat: a harness maps a
	// subdirectory of it to an extra namespace segment, so a file one level down
	// registers as /<plugin>:<dir>:<verb> rather than the documented
	// /<plugin>:<verb>.
	CommandsDir string `json:"commands_dir"`
	// BareCommand is the CommandsDir file stem that backs the plugin's BARE
	// top-level command rather than a /<plugin>:<verb> surface — for abcd, the
	// `abcd.md` whose registry row is the bare `/abcd`. It is excluded from the
	// command-surface set the way README is, because the registry names it
	// without a sub-verb and matching it by name would demand a row that cannot
	// exist. Empty means the plugin has no such file.
	BareCommand string `json:"bare_command"`
	// SkillsDir is the surface_coverage skills directory (skills); each immediate
	// subdirectory is a shipped skill surface. Also outside Roots.
	SkillsDir string `json:"skills_dir"`
	// Snapshot is the committed command-tree snapshot surface_coverage's
	// sub-verb pass checks table rows against (spc-27), repo-relative — for
	// abcd, .abcd/development/release/surface.json, whose bytes the surface
	// drift test keeps equal to the live cobra tree. Empty leaves the sub-verb
	// grain unarmed (the pre-spc-27 surface-grain check only).
	Snapshot string `json:"snapshot"`
	// HostDelegated lists surfaces whose workflow runs in the host agent with
	// no Go verb (consult, ingest, prepare-this-repo): their sub-verb tables
	// are format-checked only, never compared to the cobra tree. Explicit
	// config, never a hard-coded skip — an unlisted surface gets the full
	// comparison.
	HostDelegated []string `json:"host_delegated"`
	// OperatorInternal lists top-level verbs absent from the surface registry
	// by design (spec, rules, hook, completion — operator plumbing, not
	// product surface): the sub-verb reverse sweep does not demand a surface
	// file for them. Explicit config, same rationale as HostDelegated.
	OperatorInternal []string `json:"operator_internal"`
	// Target is the context_status_free single-file target, repo-relative
	// (.abcd/work/CONTEXT.md). The rule runs even though the target lies outside
	// Roots; a missing target is not an error.
	Target string `json:"target"`
	// Patterns is the context_status_free line-match regexp list; when empty the
	// rule falls back to contextStatusDefaultPatterns.
	Patterns []string `json:"patterns"`
	// Section is the context_citation_currency heading matcher: the rule reads only
	// the section of Target whose heading matches (the sharp-edges list), because a
	// citation to a terminal record is legitimate everywhere else. Empty falls back
	// to defaultContextSection.
	Section string `json:"section"`
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
	// GlossaryDir is the forbidden_synonyms (GL002) glossary directory, repo-relative
	// (default .abcd/development/brief/glossary). The rule walks it for term files and
	// reads each term's forbidden_synonyms frontmatter list — the glossary is the
	// single source of truth for what a forbidden synonym is.
	GlossaryDir string `json:"glossary_dir"`
	// Enforce is the forbidden_synonyms subset that GL002 mechanically gates. Each
	// entry MUST be declared as a forbidden_synonym by some glossary term (the rule
	// errors otherwise, so the config can never gate a word the glossary does not
	// forbid). Enforcement is a deliberate subset because most forbidden synonyms
	// ("user", "release", "project", "feature", ...) are common English words whose
	// live-prose false-positive rate blows the detector's budget; "epic" is the
	// mechanically-clean member (itd-43). Promotion path: add a synonym here once the
	// corpus is swept clean of its non-substituting uses.
	Enforce []string `json:"enforce"`
	// ExemptPrefixes are repo-relative path prefixes whose files GL002 skips — the
	// historical, git-tracked records the rename intent (itd-43 AC1) exempts:
	// research/, decisions/ (dated ADRs), plans/ (dated), shipped/superseded intents,
	// the issue ledger, and review records. The glossary directory itself is always
	// exempt (a term file names its own forbidden synonyms legitimately).
	ExemptPrefixes []string `json:"exempt_prefixes"`
	// AllowContext lists regexps that, if any matches a line, suppress every GL002
	// finding on that line — the legitimate-mention escape (naming the old token in an
	// external reference like `epic-review`, or the rename itself `epic->spec`).
	AllowContext []string `json:"allow_context"`
	// CrosswalkHeading is the citation_crosswalk_rows heading matcher: a table is
	// judged a crosswalk only when its nearest preceding heading matches this
	// regexp. Default defaultCrosswalkHeading. Keying on the heading rather than on
	// table shape is what keeps the rule off every ordinary table in the corpus.
	CrosswalkHeading string `json:"crosswalk_heading"`
	// RefusedDomains is the citation_source_policy list of aggregator domains a
	// citation may not point at. Matching is host-normalised (case-folded, www.
	// dropped) and covers subdomains. It ships EMPTY: naming a domain is a
	// project's editorial policy, never something the gate invents for it.
	RefusedDomains []string `json:"refused_domains"`
	// Baseline is the citation_baseline record's repo-relative path. Default
	// DefaultBaselinePath.
	Baseline string `json:"baseline"`
	// WarnAfterDays is the citation_baseline staleness warn threshold in days
	// (spc-17: 180). Zero means the default.
	WarnAfterDays int `json:"warn_after_days"`
	// BlockAfterDays is the age at which an entry becomes OVERDUE and is reported
	// under the citation_baseline_overdue rule id (spc-17: 365). Zero means the
	// default.
	BlockAfterDays int `json:"block_after_days"`
	// OverdueSeverity is the severity of a citation_baseline_overdue finding. It
	// defaults to warn because the COMMIT gate never calendar-blocks (spc-17);
	// the release gate is what promotes it to a blocker.
	OverdueSeverity string `json:"overdue_severity"`
	// RecordStores are the record_schema stores: the repo-relative directory of
	// each identified record store, keyed by id prefix (adr, itd, spc, iss). The
	// rule reasons ACROSS the stores, which straddle Roots (the issue ledger is
	// working-tier, the rest is the design record), so each is named repo-relative
	// here. An unnamed or absent store contributes nothing. Only the LOCATIONS are
	// configurable — which lifecycle buckets each store declares is the record's
	// schema and lives in code, so a config can never quietly add or hide one.
	RecordStores map[string]string `json:"record_stores"`
	// Indexes are the index_drift pairs: each names a marked region in a document
	// and the directory that region enumerates. The rule reads documents outside
	// Roots (a package or command README lives beside its code), so the pairs are
	// declared here rather than discovered by the walk.
	Indexes []IndexSpec `json:"indexes"`
	// Changelog is the delivery_state changelog, repo-relative. Default
	// "CHANGELOG.md". It sits at the repo root, outside Roots.
	Changelog string `json:"changelog"`
	// IntentsRoot is the delivery_state intents store, repo-relative — the whole
	// store, not a subdirectory of a root, because the rule resolves a cited id to
	// the lifecycle bucket holding it and the changelog it reads is repo-scoped.
	// Default .abcd/development/intents.
	IntentsRoot string `json:"intents_root"`
	// DeliverySections are additional changelog change-type headings delivery_state
	// reads as delivery claims. They are UNIONED with deliveryStateSections, never
	// substituted for them: a repo names its own headings here to widen the gate,
	// and a config that could narrow it could silently disarm it.
	DeliverySections []string `json:"delivery_sections"`
}

// IndexSpec is one index_drift pair — a hand-written enumeration of a
// directory's contents, and the directory it must agree with.
type IndexSpec struct {
	// ID names the region: the document fences it with `<!-- index: <id> -->`
	// and `<!-- /index -->`.
	ID string `json:"id"`
	// Doc is the enumerating document, repo-relative.
	Doc string `json:"doc"`
	// Dir is the enumerated directory, repo-relative. In absent mode it is the
	// base each listed path resolves against.
	Dir string `json:"dir"`
	// Entry is the regexp a backticked token in the region must match to count as
	// an entry. It is what keeps the rule off the prose around the list: a token
	// the pattern does not describe (a flag, a tool name) is not an entry.
	Entry string `json:"entry"`
	// DirEntry is the mirror of Entry on the directory side: the regexp a file's
	// stem must match to be an entry, reduced to submatch 1 when the pattern has a
	// group. It is what lets a document enumerate records by IDENTITY rather than
	// by filename — `itd-8` for `itd-8-with-code-bundling.md` — without a second
	// rule: without it the two sides can only agree when the document transcribes
	// whole slugs, which is a listing nobody writes by hand. Empty compares whole
	// stems.
	DirEntry string `json:"dir_entry"`
	// Suffix is the file extension an exact index enumerates (".md" lists file
	// stems); empty enumerates the directory's immediate subdirectories.
	Suffix string `json:"suffix"`
	// Mode is "exact" (default — the listing and the directory must agree) or
	// "absent" (every listed path must NOT exist, the planned-seams shape).
	Mode string `json:"mode"`
}

// ArmReceiptGate returns cfg with the receipt_gate rule armed for a release: it
// is enabled, pointed at the target commit, and its required-gates list is set to
// the caller's list verbatim. This is how a release runs the gate: the CALLER (a
// CI workflow) supplies the arming, so the decision to gate, the target commit,
// and the required-gates list are trust-rooted to the workflow rather than the
// in-tree, committer-editable config (phase-2 review Finding 2). The caller's list
// is authoritative even when empty: an empty list clears the gates so
// checkReceiptGate fails closed, rather than inheriting a config a committer could
// have shrunk (an argless arming must not silently pick up in-tree gates). The
// input cfg is not mutated (the Rules map is copied). Other rules are unchanged;
// the deterministic gates still run alongside.
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
	// Verbatim, including empty: the caller is the trust root, so an empty list
	// clears the gates (fail-closed at check time) rather than inheriting the
	// committer-editable config.
	rc.RequiredGates = requiredGates
	rules["receipt_gate"] = rc
	cfg.Rules = rules
	return cfg
}

// LoadConfig reads and decodes a record-lint config file. The config is a trust
// boundary: it is a committed, cross-repo-clonable file (a hostile clone can
// commit .abcd/docs-lint.json as a git mode-120000 symlink), and the read is
// reachable automatically through the session hooks (ahoy.Detect), so it is
// guarded exactly like its .abcd/*.json siblings guard.Load and rules.Load —
// a symlinked config directory or leaf is refused, a FIFO/device leaf returns
// immediately instead of hanging the open, and an over-cap file is rejected
// rather than read into an OOM. Error strings stay path-free (iss-29); the
// os.IsNotExist branch is preserved so callers that default an absent config
// still work.
func LoadConfig(path string) (Config, error) {
	// Refuse a symlinked config directory before touching the leaf, so a swapped
	// .abcd cannot redirect the read at a config the repository does not own.
	if di, err := os.Lstat(filepath.Dir(path)); err == nil && di.Mode()&os.ModeSymlink != 0 {
		return Config{}, fmt.Errorf("lint: config directory is a symlink (refusing to follow)")
	}
	data, err := fsutil.ReadGuarded(path, maxLintConfigBytes)
	if err != nil {
		switch {
		case os.IsNotExist(err):
			return Config{}, err
		case errors.Is(err, syscall.ELOOP):
			return Config{}, fmt.Errorf("lint: config is a symlink (refusing to follow)")
		case errors.Is(err, fsutil.ErrNotRegular):
			return Config{}, fmt.Errorf("lint: config is not a regular file")
		case errors.Is(err, fsutil.ErrTooBig):
			return Config{}, fmt.Errorf("lint: config exceeds the %d-byte cap", maxLintConfigBytes)
		default:
			return Config{}, err
		}
	}
	// Strict decode: a misspelt key would silently zero-value the field it
	// missed ("enabld" disarms a rule, "severty" strips its exit-code weight),
	// and both misreads survive review because the file still looks armed.
	var cfg Config
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&cfg); err != nil {
		return Config{}, err
	}
	if err := cfg.validateBannedTokens(); err != nil {
		return Config{}, err
	}
	if err := cfg.validateSeverities(); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

// validateSeverities refuses a severity outside the engine's enum on any
// enabled rule or banned token. The exit paths count Severity == "blocker"
// verbatim, so an off-enum value would emit findings that serialize yet count
// toward no exit code — a clean exit beside a non-empty findings list, which
// the sibling engines (repolint.Evaluate, guard.Validate, banlist.AddPublic)
// name a rule bug and fail closed on. A disabled rule is inert and its
// severity is not consulted, so it is not checked.
func (c Config) validateSeverities() error {
	for id, rc := range c.Rules {
		if !rc.Enabled {
			continue
		}
		if rc.Severity != severityBlocker && rc.Severity != severityWarn {
			return &configError{"rule " + id + " has severity " + strconv.Quote(rc.Severity) + "; an enabled rule must declare \"blocker\" or \"warn\", or its findings count toward no exit code"}
		}
	}
	for i, t := range c.BannedTokens {
		if t.Severity == severityBlocker || t.Severity == severityWarn {
			continue
		}
		who := t.ID
		if who == "" {
			who = "index " + strconv.Itoa(i)
		}
		return &configError{"banned_tokens entry " + who + " has severity " + strconv.Quote(t.Severity) + "; want \"blocker\" or \"warn\""}
	}
	return nil
}

// validateBannedTokens enforces the strict banned_tokens schema (iss-51): every
// entry must declare a non-empty successor (the machine-readable replacement,
// not prose alone) and a non-empty allow_context (where the token is legitimately
// allowed). A violation is a load-time rejection, so a defective ban can never
// reach the linter. Errors identify the offending entry by id (or index when the
// id is itself absent).
func (c Config) validateBannedTokens() error {
	for i, t := range c.BannedTokens {
		who := t.ID
		if who == "" {
			who = "index " + strconv.Itoa(i)
		}
		if strings.TrimSpace(t.Successor) == "" {
			return &configError{"banned_tokens entry " + who + " has no successor; a ban must declare the machine-readable replacement (iss-51)"}
		}
		if len(t.AllowContext) == 0 {
			return &configError{"banned_tokens entry " + who + " has an empty allow_context; a ban must declare where its token is legitimately allowed (iss-51)"}
		}
	}
	return nil
}
