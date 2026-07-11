// Package rules is abcd's transport-agnostic modular rules loader (itd-3). It
// owns the whole capability behind two front doors: the `abcd rules [domain]`
// CLI verb and the Claude Code prompt-router hook. Nothing here writes to stdout
// or knows about a harness event — the front doors under internal/surface and
// the hook entrypoint marshal these results for their transport.
//
// The model is a small set of binary-bundled default domains (embedded below)
// merged with an optional per-repo <repoRoot>/.abcd/rules.json override. Each
// domain carries recall keywords + aliases and a list of rules; a prompt is
// recall-matched against the active domains and only the matching rules are
// rendered for injection. A leading *<DOMAIN> star-command activates a domain
// unconditionally (overriding a dormant state, but never the top-level kill
// switch).
//
// Validation is hand-rolled Go (zero new dependencies): a rules.json that fails
// to parse or validate is a fail-closed error the caller surfaces loudly — it
// never silently degrades to zero injection.
package rules

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"hash/fnv"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// RepoRelPath is the per-repo override file, relative to the repo worktree.
const RepoRelPath = ".abcd/rules.json"

// maxRulesFileBytes caps the per-repo rules.json (trust boundary).
const maxRulesFileBytes = 256 * 1024

// Domain state values. An empty string is treated as active.
const (
	StateActive  = "active"
	StateDormant = "dormant"
)

// Domain is one keyed rule domain. The map key in RuleSet.Domains is its name;
// the name is therefore not a serialized field on the value.
type Domain struct {
	State   string   `json:"state,omitempty"`
	Recall  []string `json:"recall,omitempty"`
	Aliases []string `json:"aliases,omitempty"`
	Rules   []string `json:"rules,omitempty"`
}

// RuleSet is the merged, validated rule model: the bundled defaults overlaid
// with the per-repo override.
type RuleSet struct {
	SchemaVersion int               `json:"schema_version"`
	Disabled      bool              `json:"disabled"`
	Domains       map[string]Domain `json:"domains"`
}

// ResolvedDomain pairs a domain with its name for ordered rendering and dedup.
type ResolvedDomain struct {
	Name string `json:"name"`
	Domain
}

// domainNameRe constrains domain keys so a custom domain id can never be used
// to build a filesystem path (path-traversal defence) — uppercase, starting
// with a letter.
var domainNameRe = regexp.MustCompile(`^[A-Z][A-Z0-9_]*$`)

// starCommandRe finds a candidate *<DOMAIN> token. Go's RE2 has no lookahead,
// so the pinned boundary semantics `(?:^|\s)\*([A-Z][A-Z0-9_]*)(?=$|\s)` are
// enforced by checking the surrounding bytes in parseStarCommands.
var starCommandRe = regexp.MustCompile(`\*([A-Z][A-Z0-9_]*)`)

// nonAlnumRe collapses every run of non-alphanumeric characters to one space so
// recall matching is word-bounded (no substring false positives).
var nonAlnumRe = regexp.MustCompile(`[^a-z0-9]+`)

//go:embed defaults/rules.json
var defaultsJSON []byte

// defaultRuleSet is parsed once at init; a malformed embedded asset is a build
// error surfaced as a panic (it can never happen at runtime).
var defaultRuleSet = mustParseDefaults()

func mustParseDefaults() RuleSet {
	var rs RuleSet
	if err := json.Unmarshal(defaultsJSON, &rs); err != nil {
		panic("rules: bundled defaults are malformed: " + err.Error())
	}
	if err := Validate(rs); err != nil {
		panic("rules: bundled defaults fail validation: " + err.Error())
	}
	return rs
}

// Defaults returns a deep copy of the binary-bundled default rule set, safe for
// the caller to mutate.
func Defaults() RuleSet { return cloneRuleSet(defaultRuleSet) }

// Load returns the defaults merged with <repoRoot>/.abcd/rules.json when that
// file exists. An absent file yields the defaults unchanged; a present file
// that cannot be parsed or fails validation is a fail-closed error.
func Load(repoRoot string) (RuleSet, error) {
	path := filepath.Join(repoRoot, ".abcd", "rules.json")
	fi, err := os.Lstat(path)
	if os.IsNotExist(err) {
		return Defaults(), nil
	}
	if err != nil {
		return RuleSet{}, fmt.Errorf("rules: stat %s: %w", RepoRelPath, err)
	}
	// Trust-boundary guards (mirror the ahoy hook-manifest verifier): refuse a
	// symlinked leaf and cap the file size so a hostile rules.json cannot force a
	// symlink-follow or a memory blow-up.
	if fi.Mode()&os.ModeSymlink != 0 {
		return RuleSet{}, fmt.Errorf("rules: %s is a symlink (refusing to follow)", RepoRelPath)
	}
	if !fi.Mode().IsRegular() {
		return RuleSet{}, fmt.Errorf("rules: %s is not a regular file", RepoRelPath)
	}
	if fi.Size() > maxRulesFileBytes {
		return RuleSet{}, fmt.Errorf("rules: %s exceeds the %d-byte cap", RepoRelPath, maxRulesFileBytes)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return RuleSet{}, fmt.Errorf("rules: reading %s: %w", RepoRelPath, err)
	}
	var over RuleSet
	if err := json.Unmarshal(data, &over); err != nil {
		return RuleSet{}, fmt.Errorf("rules: %s is not valid JSON: %w", RepoRelPath, err)
	}
	merged := Merge(Defaults(), over)
	if err := Validate(merged); err != nil {
		return RuleSet{}, fmt.Errorf("rules: %s: %w", RepoRelPath, err)
	}
	return merged, nil
}

// Merge overlays over onto base. Domain fields are per-field: a field set on the
// override wins; an absent field inherits the base (so {"state":"dormant"} on a
// default domain silences it while keeping its recall and rules). New domain
// keys are added. The kill switch is sticky (either side can enable it).
func Merge(base, over RuleSet) RuleSet {
	out := cloneRuleSet(base)
	if over.SchemaVersion != 0 {
		out.SchemaVersion = over.SchemaVersion
	}
	out.Disabled = base.Disabled || over.Disabled
	for name, od := range over.Domains {
		out.Domains[name] = mergeDomain(out.Domains[name], od)
	}
	return out
}

func mergeDomain(base, over Domain) Domain {
	r := base
	if over.State != "" {
		r.State = over.State
	}
	if over.Recall != nil {
		r.Recall = append([]string(nil), over.Recall...)
	}
	if over.Aliases != nil {
		r.Aliases = append([]string(nil), over.Aliases...)
	}
	if over.Rules != nil {
		r.Rules = append([]string(nil), over.Rules...)
	}
	return r
}

// Validate checks structural invariants: schema_version == 1, every domain name
// matches [A-Z][A-Z0-9_]*, and every state is active/dormant (or empty).
func Validate(rs RuleSet) error {
	if rs.SchemaVersion != 1 {
		return fmt.Errorf("schema_version must be 1, got %d", rs.SchemaVersion)
	}
	for name, d := range rs.Domains {
		if !domainNameRe.MatchString(name) {
			return fmt.Errorf("domain name %q must match [A-Z][A-Z0-9_]*", name)
		}
		switch d.State {
		case "", StateActive, StateDormant:
		default:
			return fmt.Errorf("domain %q: unknown state %q", name, d.State)
		}
	}
	return nil
}

// Match returns the domains to inject for prompt, in deterministic (name-sorted)
// order. The top-level kill switch suppresses everything. Otherwise a domain is
// injected if a star-command names it (overriding dormant) or — when active —
// its recall keywords or aliases hit the prompt.
func (rs RuleSet) Match(prompt string) []ResolvedDomain {
	if rs.Disabled {
		return nil
	}
	stars := parseStarCommands(prompt)
	norm := normalize(prompt)

	names := make([]string, 0, len(rs.Domains))
	for name := range rs.Domains {
		names = append(names, name)
	}
	sort.Strings(names)

	var out []ResolvedDomain
	for _, name := range names {
		d := rs.Domains[name]
		if stars[name] {
			out = append(out, ResolvedDomain{Name: name, Domain: d})
			continue
		}
		if d.State == StateDormant {
			continue
		}
		if recallHit(norm, d) {
			out = append(out, ResolvedDomain{Name: name, Domain: d})
		}
	}
	return out
}

// Active returns every injectable domain (state != dormant) in name-sorted
// order — the full set the diagnostic `abcd rules` render shows. The top-level
// kill switch yields nothing.
func (rs RuleSet) Active() []ResolvedDomain {
	if rs.Disabled {
		return nil
	}
	names := make([]string, 0, len(rs.Domains))
	for name := range rs.Domains {
		names = append(names, name)
	}
	sort.Strings(names)
	var out []ResolvedDomain
	for _, name := range names {
		if d := rs.Domains[name]; d.State != StateDormant {
			out = append(out, ResolvedDomain{Name: name, Domain: d})
		}
	}
	return out
}

// Lookup returns one domain by name regardless of its state (a dormant domain is
// still inspectable); ok is false when the name is absent.
func (rs RuleSet) Lookup(name string) (ResolvedDomain, bool) {
	d, ok := rs.Domains[name]
	if !ok {
		return ResolvedDomain{}, false
	}
	return ResolvedDomain{Name: name, Domain: d}, true
}

// parseStarCommands extracts the set of *<DOMAIN> names, enforcing that the star
// is at the start or preceded by whitespace and the name is followed by the end
// or whitespace (the RE2-safe form of the pinned lookahead boundary).
func parseStarCommands(prompt string) map[string]bool {
	out := map[string]bool{}
	for _, loc := range starCommandRe.FindAllStringSubmatchIndex(prompt, -1) {
		starStart, nameEnd := loc[0], loc[3]
		if starStart > 0 && !isSpace(prompt[starStart-1]) {
			continue
		}
		if nameEnd < len(prompt) && !isSpace(prompt[nameEnd]) {
			continue
		}
		out[prompt[loc[2]:loc[3]]] = true
	}
	return out
}

func isSpace(b byte) bool {
	return b == ' ' || b == '\t' || b == '\n' || b == '\r' || b == '\f' || b == '\v'
}

// normalize lowercases prompt, collapses non-alphanumeric runs to single spaces,
// and pads with a leading and trailing space so a term wrapped in spaces matches
// on word boundaries.
func normalize(s string) string {
	lowered := strings.ToLower(s)
	collapsed := nonAlnumRe.ReplaceAllString(lowered, " ")
	return " " + strings.TrimSpace(collapsed) + " "
}

// recallHit reports whether any recall keyword or alias appears in the
// space-normalized prompt on a word boundary. Multi-word terms are supported.
func recallHit(norm string, d Domain) bool {
	for _, term := range d.Recall {
		if termHit(norm, term) {
			return true
		}
	}
	for _, term := range d.Aliases {
		if termHit(norm, term) {
			return true
		}
	}
	return false
}

func termHit(norm, term string) bool {
	t := strings.TrimSpace(nonAlnumRe.ReplaceAllString(strings.ToLower(term), " "))
	if t == "" {
		return false
	}
	return strings.Contains(norm, " "+t+" ")
}

// Render is the single renderer both front doors use. It emits a header plus one
// section per domain. An empty domain list renders to zero bytes (D3: no
// model-facing tokens on a no-match).
func Render(domains []ResolvedDomain) string {
	if len(domains) == 0 {
		return ""
	}
	var b strings.Builder
	fmt.Fprintf(&b, "# abcd rules — %d domain(s) active\n", len(domains))
	for _, d := range domains {
		b.WriteString(renderDomain(d))
	}
	return b.String()
}

// renderDomain renders one domain's block deterministically. Signature hashes
// exactly this, so the format is the dedup unit.
func renderDomain(d ResolvedDomain) string {
	var b strings.Builder
	fmt.Fprintf(&b, "## %s\n", d.Name)
	for _, r := range d.Rules {
		fmt.Fprintf(&b, "- %s\n", r)
	}
	return b.String()
}

// Signature is the per-domain dedup key: an FNV-1a hash of the rendered block,
// so identical rendered content (defaults or override) dedups and any content
// drift invalidates. FNV is sufficient — this is dedup, not security.
func Signature(d ResolvedDomain) string {
	h := fnv.New64a()
	_, _ = h.Write([]byte(renderDomain(d)))
	return fmt.Sprintf("%016x", h.Sum64())
}

func cloneRuleSet(rs RuleSet) RuleSet {
	out := RuleSet{SchemaVersion: rs.SchemaVersion, Disabled: rs.Disabled}
	if rs.Domains != nil {
		out.Domains = make(map[string]Domain, len(rs.Domains))
		for name, d := range rs.Domains {
			out.Domains[name] = Domain{
				State:   d.State,
				Recall:  append([]string(nil), d.Recall...),
				Aliases: append([]string(nil), d.Aliases...),
				Rules:   append([]string(nil), d.Rules...),
			}
		}
	}
	return out
}
