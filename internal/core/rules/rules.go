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
	"errors"
	"fmt"
	"hash/fnv"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"syscall"
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

var (
	errNotRegular = errors.New("not a regular file")
	errTooBig     = errors.New("exceeds size cap")
)

// readGuarded opens path once, read-only, with O_NOFOLLOW (refuse a symlinked
// leaf) and O_NONBLOCK (a FIFO/device leaf returns immediately instead of
// blocking the open forever), then validates on the SAME file descriptor that it
// is a regular file within limit bytes before reading through a LimitReader — so
// no symlink swap between stat and read, no non-regular leaf, and no size overrun
// can reach the caller. The raw open error is returned so callers can test
// os.IsNotExist / syscall.ELOOP; a non-regular or oversize file returns the
// errNotRegular / errTooBig sentinel.
func readGuarded(path string, limit int64) ([]byte, error) {
	f, err := os.OpenFile(path, os.O_RDONLY|syscall.O_NOFOLLOW|syscall.O_NONBLOCK, 0)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	fi, err := f.Stat()
	if err != nil {
		return nil, err
	}
	if !fi.Mode().IsRegular() {
		return nil, errNotRegular
	}
	if fi.Size() > limit {
		return nil, errTooBig
	}
	data, err := io.ReadAll(io.LimitReader(f, limit+1))
	if err != nil {
		return nil, err
	}
	if int64(len(data)) > limit {
		// The file grew past the cap between fstat and read (a size TOCTOU).
		return nil, errTooBig
	}
	return data, nil
}

// Load returns the defaults merged with <repoRoot>/.abcd/rules.json when that
// file exists. An absent file yields the defaults unchanged; a present file
// that cannot be parsed or fails validation is a fail-closed error.
func Load(repoRoot string) (RuleSet, error) {
	// Refuse a symlinked .abcd directory component before touching the leaf, so a
	// swapped .abcd cannot redirect the read (trust boundary).
	if di, err := os.Lstat(filepath.Join(repoRoot, ".abcd")); err == nil && di.Mode()&os.ModeSymlink != 0 {
		return RuleSet{}, fmt.Errorf("rules: .abcd is a symlink (refusing to follow)")
	}
	path := filepath.Join(repoRoot, ".abcd", "rules.json")
	data, err := readGuarded(path, maxRulesFileBytes)
	if err != nil {
		switch {
		case os.IsNotExist(err):
			return Defaults(), nil
		case errors.Is(err, syscall.ELOOP):
			return RuleSet{}, fmt.Errorf("rules: %s is a symlink (refusing to follow)", RepoRelPath)
		case errors.Is(err, errNotRegular):
			return RuleSet{}, fmt.Errorf("rules: %s is not a regular file", RepoRelPath)
		case errors.Is(err, errTooBig):
			return RuleSet{}, fmt.Errorf("rules: %s exceeds the %d-byte cap", RepoRelPath, maxRulesFileBytes)
		default:
			return RuleSet{}, fmt.Errorf("rules: reading %s: %w", RepoRelPath, err)
		}
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
	idx := indexPrompt(prompt)

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
		if idx.hit(d) {
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

// promptIndex is a prompt prepared for recall matching once per Match call: a
// space-padded normalized form for word-boundary and multi-word matching, a
// second padded form with every token stemmed (so inflected multi-word phrases
// still hit their alias), plus a set of candidate stems for the single tokens so
// inflected forms (commits->commit, pushes->push, committing->commit,
// merging->merge) recall-match their keyword.
type promptIndex struct {
	padded        string          // " tok tok " for boundary/phrase matching
	stemmedPadded string          // " stem stem " for stemmed phrase matching
	stems         map[string]bool // candidate stems of the single tokens
}

// indexPrompt lowercases, collapses non-alphanumeric runs to single spaces, and
// builds both the stemmed-token set (with every candidate root per token) and the
// stemmed padded form used for multi-word phrase matching.
func indexPrompt(s string) promptIndex {
	collapsed := strings.TrimSpace(nonAlnumRe.ReplaceAllString(strings.ToLower(s), " "))
	idx := promptIndex{padded: " " + collapsed + " ", stems: map[string]bool{}}
	tokens := strings.Fields(collapsed)
	stemmed := make([]string, 0, len(tokens))
	for _, tok := range tokens {
		for _, v := range stemVariants(tok) {
			idx.stems[v] = true
		}
		stemmed = append(stemmed, stem(tok))
	}
	idx.stemmedPadded = " " + strings.Join(stemmed, " ") + " "
	return idx
}

// hit reports whether any of a domain's recall keywords or aliases match.
func (idx promptIndex) hit(d Domain) bool {
	for _, term := range d.Recall {
		if idx.termHit(term) {
			return true
		}
	}
	for _, term := range d.Aliases {
		if idx.termHit(term) {
			return true
		}
	}
	return false
}

// termHit matches a single term. A multi-word term is a word-boundary substring
// of the padded prompt OR — with each of its words stemmed — of the stemmed
// padded prompt, so inflected phrases ("pull requests") still hit their alias
// ("pull request"). A single token matches on an exact word boundary OR when its
// stem is among the prompt tokens' candidate stems, so plural/tense variants
// recall their keyword.
func (idx promptIndex) termHit(term string) bool {
	t := strings.TrimSpace(nonAlnumRe.ReplaceAllString(strings.ToLower(term), " "))
	if t == "" {
		return false
	}
	if strings.Contains(t, " ") {
		if strings.Contains(idx.padded, " "+t+" ") {
			return true
		}
		return strings.Contains(idx.stemmedPadded, " "+stemPhrase(t)+" ")
	}
	if strings.Contains(idx.padded, " "+t+" ") {
		return true
	}
	return idx.stems[stem(t)]
}

// stemPhrase stems each whitespace-separated word of a normalized multi-word
// term, so a phrase alias can be compared against the stemmed padded prompt.
func stemPhrase(t string) string {
	words := strings.Fields(t)
	for i, w := range words {
		words[i] = stem(w)
	}
	return strings.Join(words, " ")
}

// stem strips a common English suffix to a root. Short tokens are left untouched
// so acronyms and 2–4 letter keywords (sota, pr, docs) are never over-stemmed —
// the guard that keeps stemming from matching e.g. "test" against "attestation".
func stem(w string) string {
	switch {
	case len(w) > 5 && strings.HasSuffix(w, "ing"):
		return w[:len(w)-3]
	case len(w) > 4 && strings.HasSuffix(w, "ed"):
		return w[:len(w)-2]
	case len(w) > 4 && strings.HasSuffix(w, "es"):
		// "es" plural attaches to sibilant roots (boxes->box, pushes->push);
		// elsewhere it is root+"s" (issues->issue), so only strip "es" after a
		// sibilant, otherwise drop just the trailing "s".
		root := w[:len(w)-2]
		if hasSibilantSuffix(root) {
			return root
		}
		return w[:len(w)-1]
	case len(w) > 3 && strings.HasSuffix(w, "s") && !strings.HasSuffix(w, "ss"):
		return w[:len(w)-1]
	}
	return w
}

// hasSibilantSuffix reports whether a root takes an "-es" plural (s/x/z/ch/sh).
func hasSibilantSuffix(root string) bool {
	for _, suf := range []string{"s", "x", "z", "ch", "sh"} {
		if strings.HasSuffix(root, suf) {
			return true
		}
	}
	return false
}

// stemVariants returns every candidate root a prompt token may share with a base
// keyword. It is the asymmetric counterpart of stem(): a keyword stems to a
// single canonical root, while a prompt token expands to that root plus the two
// forms an "-ing"/"-ed" inflection would otherwise hide — the e-drop restore
// ("merg"->"merge", "rebas"->"rebase") and the undoubled consonant
// ("committ"->"commit"). Only the "-ing"/"-ed" cases branch, so plural handling
// and the short-token guard from stem() are unchanged, and the extra roots stay
// conservative (no bare-vowel or sub-3-char stems) to avoid over-matching.
func stemVariants(w string) []string {
	var root string
	switch {
	case len(w) > 5 && strings.HasSuffix(w, "ing"):
		root = w[:len(w)-3]
	case len(w) > 4 && strings.HasSuffix(w, "ed"):
		root = w[:len(w)-2]
	default:
		return []string{stem(w)}
	}
	out := []string{root, root + "e"}
	if u := undouble(root); u != "" {
		out = append(out, u)
	}
	return out
}

// undouble collapses a trailing doubled consonant to a single one
// ("committ"->"commit", "stopp"->"stop"), or returns "" when the root does not
// end in a doubled consonant. Such doubling is introduced when an "-ing"/"-ed"
// inflection is stripped from a verb whose final consonant was doubled. The
// undoubled root must stay at least 3 characters to avoid tiny, over-matching
// stems.
func undouble(root string) string {
	n := len(root)
	if n < 4 {
		return ""
	}
	if a, b := root[n-2], root[n-1]; a == b && isConsonant(a) {
		return root[:n-1]
	}
	return ""
}

// isConsonant reports whether b is an ASCII lowercase consonant.
func isConsonant(b byte) bool {
	if b < 'a' || b > 'z' {
		return false
	}
	switch b {
	case 'a', 'e', 'i', 'o', 'u':
		return false
	}
	return true
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
