package reading

// scope.go is what a reading is ABOUT.
//
// The assembler shipped able to name a position and a commit and nothing else,
// which meant every reading received its position's whole corpus: about 9.2 MB
// and 2.4 million estimated tokens over this repository, with three of the four
// positions receiving a byte-identical item set. An instrument that cannot be
// pointed at anything cannot be given to a reader.
//
// A scope NARROWS and never widens. It intersects what the include table
// already admits at the position, so no scope reaches a row the table denies
// there, and the structural deny, the exclusion floor and the dirty gate all
// run over the unfiltered walk before this filter is applied (itd-199, spc-69).
//
// The invocation carries no prose. adr-58 supersedes the 2026-08-28 M8 ruling
// to admit this third operand and no more, on the property M8 was protecting:
// every token is a shape-validated closed form, and a repository path is never
// accepted at the invocation. A path may be named only inside a committed
// preset, where it is reviewed, shape-validated and inside the dirty gate.

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// PresetConfigPath is the committed preset configuration. It is the ONE place a
// repository path may be named, and it joins the record-lint configuration in
// the dirty set by the same argument: an uncommitted edit to it reshapes an
// assembly as surely as an edit to a record does.
const PresetConfigPath = ".abcd/config/reading-presets.json"

// PresetSchemaVersion is the preset file's own shape version, separate from the
// artefacts' SchemaVersion: the configuration and the output are different
// shapes with different reasons to move.
const PresetSchemaVersion = 1

// recordIDRe is the record-id token form. It is deliberately narrow: a token
// that is not one of these shapes is not a record id, and is never guessed at.
var recordIDRe = regexp.MustCompile(`^(itd|spc|adr|iss)-[0-9]+$`)

// presetNameRe bounds a preset name to a shape that cannot be confused with a
// path, a record id or prose.
var presetNameRe = regexp.MustCompile(`^[a-z][a-z0-9-]{0,31}$`)

// Selector is one clause of a scope. A candidate is selected if ANY clause
// matches it, so a scope is a union of its clauses and an empty scope selects
// nothing.
type Selector struct {
	// Kind selects every item of one material class.
	Kind Kind `json:"kind,omitempty"`
	// Record selects the items whose path names one record.
	Record string `json:"record,omitempty"`
	// Path selects the items at or beneath one repo-relative directory. It is
	// reachable ONLY from a committed preset, never from the invocation.
	Path string `json:"path,omitempty"`
}

// Scope is what one run was about: the resolved selectors, and how they were
// named. The Source and Overridden fields are provenance, not selection.
type Scope struct {
	// Source is the token the operator gave, echoed so an artefact says what it
	// was asked for and not only what that resolved to.
	Source string `json:"source"`
	// Selectors are the resolved clauses, in a canonical order so one scope
	// hashes to one value.
	Selectors []Selector `json:"selectors"`
	// Overridden records that the run departed from the committed presets.
	// Naming a preset is running as reviewed; naming a record or a kind
	// directly is the departure worth counting.
	Overridden bool `json:"overridden"`
}

// selects reports whether one collected candidate is in scope.
func (s Scope) selects(c candidate) bool {
	for _, sel := range s.Selectors {
		switch {
		case sel.Kind != "" && c.kind == sel.Kind:
			return true
		case sel.Record != "" && pathNamesRecord(c.path, sel.Record):
			return true
		case sel.Path != "" && underPath(c.path, sel.Path):
			return true
		}
	}
	return false
}

// pathNamesRecord reports whether a repo-relative path is the record's own
// file. A record's file is named for its id, so the basename either IS the id
// with an extension or begins with the id and a hyphen — the slug form every
// record store writes. A prefix test alone would make itd-19 select itd-198.
func pathNamesRecord(rel, id string) bool {
	base := path.Base(rel)
	if strings.HasPrefix(base, id+"-") {
		return true
	}
	return base == id+".md" || base == id+".json"
}

// underPath reports whether a repo-relative path lies at or beneath dir.
func underPath(rel, dir string) bool {
	dir = path.Clean(dir)
	if dir == "." {
		return true
	}
	return rel == dir || strings.HasPrefix(rel, dir+"/")
}

// canonicalise sorts a scope's selectors and drops duplicates, so two scopes
// naming the same thing in a different order are one scope and hash alike.
func canonicalise(sels []Selector) []Selector {
	seen := make(map[Selector]bool, len(sels))
	out := make([]Selector, 0, len(sels))
	for _, s := range sels {
		if s == (Selector{}) || seen[s] {
			continue
		}
		seen[s] = true
		out = append(out, s)
	}
	sort.Slice(out, func(i, j int) bool {
		a, b := out[i], out[j]
		if a.Kind != b.Kind {
			return a.Kind < b.Kind
		}
		if a.Record != b.Record {
			return a.Record < b.Record
		}
		return a.Path < b.Path
	})
	return out
}

// Hash is the scope's content hash, so a manifest can name the scope a run
// actually had and a reader can tell two runs apart by it.
func (s Scope) Hash() (string, error) {
	data, err := json.Marshal(struct {
		Selectors []Selector `json:"selectors"`
	}{Selectors: s.Selectors})
	if err != nil {
		return "", fmt.Errorf("hashing reading scope: %w", err)
	}
	return sha256Hex(data), nil
}

// PositionScope is one position's scope inside a preset.
type PositionScope struct {
	Kinds   []Kind   `json:"kinds"`
	Records []string `json:"records"`
	Paths   []string `json:"paths"`
}

// Preset is one committed, named scope per position.
type Preset struct {
	// Extends names the preset this one adds to. It is a UNION and never a
	// replacement, which is what makes "warm is cold plus a delta" a property
	// rather than a review note: a scope added to cold appears in warm without
	// anyone remembering to add it twice, and warm can never be narrower.
	Extends string `json:"extends,omitempty"`
	// Positions maps a position token to its scope. A preset carries a scope
	// PER POSITION rather than one scope every position shares, because the
	// finding this exists to fix is that three of the four positions received
	// a byte-identical item set, and one scope over four near-identical
	// admissions reproduces it exactly.
	Positions map[string]PositionScope `json:"positions"`
}

// PresetFile is the committed configuration.
type PresetFile struct {
	SchemaVersion int               `json:"schema_version"`
	Presets       map[string]Preset `json:"presets"`
}

// LoadPresets reads and validates the committed preset configuration.
//
// It is strict on unknown fields, like every other artefact this package reads:
// a key nobody reads is a scope nobody applies, and a preset that silently does
// less than it says is the failure this whole intent exists to close.
func LoadPresets(repoRoot string) (PresetFile, error) {
	abs := filepath.Join(repoRoot, filepath.FromSlash(PresetConfigPath))

	// The preset's whole safety argument is that it is COMMITTED and reviewed:
	// adr-58 admits a preset name at the invocation on that ground alone, and
	// brief invariant 15 now says so. Neither half was checked, and the dirty
	// gate cannot supply it — git reports nothing for a file it ignores, so a
	// repository that gitignores `.abcd/` (which brief invariant 5 does for
	// public visibility) ran happily against an untracked preset and stamped
	// the run `overridden: false`, asserting "ran as reviewed" on an
	// examination that established only "git reported no modification". That is
	// an attestation exceeding its evidence, which brief invariant 16 forbids.
	tracked, err := trackedSet(repoRoot)
	if err != nil {
		return PresetFile{}, err
	}
	if !tracked[PresetConfigPath] {
		return PresetFile{}, fmt.Errorf("%s is not tracked: a preset is admitted at the "+
			"invocation because it is committed and reviewed, so an untracked or ignored one "+
			"has no such standing and is refused rather than read", PresetConfigPath)
	}

	// And it must be a regular file. A committed SYMLINK is tracked, and git
	// reports nothing when its target changes, so the effective configuration
	// would be permanently unreviewed and freely mutable with every gate green.
	fi, err := os.Lstat(abs)
	if err != nil {
		return PresetFile{}, fmt.Errorf("reading %s: %w", PresetConfigPath, err)
	}
	if fi.Mode()&os.ModeSymlink != 0 {
		return PresetFile{}, fmt.Errorf("%s is a symlink: its target is outside what the commit "+
			"records, so what it resolves to is not what was reviewed", PresetConfigPath)
	}
	if !fi.Mode().IsRegular() {
		return PresetFile{}, fmt.Errorf("%s is not a regular file", PresetConfigPath)
	}

	raw, err := os.ReadFile(abs)
	if err != nil {
		return PresetFile{}, fmt.Errorf("reading %s: %w", PresetConfigPath, err)
	}
	if err := refuseDuplicatePresetKeys(raw); err != nil {
		return PresetFile{}, err
	}
	dec := json.NewDecoder(strings.NewReader(string(raw)))
	dec.DisallowUnknownFields()
	var pf PresetFile
	if err := dec.Decode(&pf); err != nil {
		return PresetFile{}, fmt.Errorf("decoding %s: %w", PresetConfigPath, err)
	}
	if pf.SchemaVersion != PresetSchemaVersion {
		return PresetFile{}, fmt.Errorf("decoding %s: schema_version is %d, want %d",
			PresetConfigPath, pf.SchemaVersion, PresetSchemaVersion)
	}
	if len(pf.Presets) == 0 {
		return PresetFile{}, fmt.Errorf("%s names no preset", PresetConfigPath)
	}
	if err := validatePresets(pf); err != nil {
		return PresetFile{}, err
	}
	return pf, nil
}

// refuseDuplicatePresetKeys refuses a preset object naming one key twice.
//
// Go's JSON decoder takes the LAST duplicate silently, and DisallowUnknownFields
// says nothing about duplicates. Against the one file whose entire safety
// argument is that a human reviewed it, silent last-wins is a review-evasion
// vector: a second `"cold"` block low in the file replaces the reviewed one,
// and a reviewer reading top-down sees the first.
func refuseDuplicatePresetKeys(raw []byte) error {
	dec := json.NewDecoder(strings.NewReader(string(raw)))
	var probe struct {
		Presets map[string]json.RawMessage `json:"presets"`
	}
	if err := json.Unmarshal(raw, &probe); err != nil {
		return nil // the strict decode below reports the real parse error
	}
	seen := map[string]int{}
	dec = json.NewDecoder(strings.NewReader(string(raw)))
	depth := 0
	inPresets := false
	for {
		tok, err := dec.Token()
		if err != nil {
			break
		}
		switch t := tok.(type) {
		case json.Delim:
			switch t {
			case '{':
				depth++
			case '}':
				depth--
				if depth <= 1 {
					inPresets = false
				}
			}
		case string:
			if depth == 1 && t == "presets" {
				inPresets = true
				continue
			}
			if inPresets && depth == 2 {
				seen[t]++
				if seen[t] > 1 {
					return fmt.Errorf("%s names the preset %q more than once; the last would win "+
						"silently, so a reviewed block could be replaced by one further down the file",
						PresetConfigPath, t)
				}
			}
		}
	}
	return nil
}

// validatePresets refuses a configuration that could not mean one thing.
func validatePresets(pf PresetFile) error {
	kinds := make(map[Kind]bool, len(Kinds()))
	for _, k := range Kinds() {
		kinds[k] = true
	}
	for name, p := range pf.Presets {
		// A collision is refused rather than resolved by precedence. A preset
		// named for a kind or shaped like a record id would make one token mean
		// two things, and a resolution ORDER would decide which silently.
		if !presetNameRe.MatchString(name) {
			return fmt.Errorf("%s: preset %q is not a valid name", PresetConfigPath, name)
		}
		if kinds[Kind(name)] {
			return fmt.Errorf("%s: preset %q collides with the material kind of the same name; "+
				"one token may not mean two things", PresetConfigPath, name)
		}
		if recordIDRe.MatchString(name) {
			return fmt.Errorf("%s: preset %q is shaped like a record id", PresetConfigPath, name)
		}
		if p.Extends != "" {
			parent, ok := pf.Presets[p.Extends]
			if !ok {
				return fmt.Errorf("%s: preset %q extends %q, which does not exist",
					PresetConfigPath, name, p.Extends)
			}
			// One level only. The containment guarantee is checkable at one
			// level and an argument at two, and an argument is what this
			// mechanism exists to replace.
			if parent.Extends != "" {
				return fmt.Errorf("%s: preset %q extends %q, which itself extends %q; "+
					"one level of extension only", PresetConfigPath, name, p.Extends, parent.Extends)
			}
		}
		for pos, ps := range p.Positions {
			position, err := ParsePosition(pos)
			if err != nil {
				return fmt.Errorf("%s: preset %q: %w", PresetConfigPath, name, err)
			}
			// The comparative position refuses before a scope is resolved, so
			// a scope for it would describe a run that cannot happen.
			if position == PositionComparative {
				return fmt.Errorf("%s: preset %q names a scope for the comparative position, "+
					"which does not assemble: its object is the widening run's pre-admission "+
					"output and no channel supplies it", PresetConfigPath, name)
			}
			for _, k := range ps.Kinds {
				if !kinds[k] {
					return fmt.Errorf("%s: preset %q at %s names the unknown kind %q; "+
						"the vocabulary is closed", PresetConfigPath, name, pos, k)
				}
			}
			for _, r := range ps.Records {
				if !recordIDRe.MatchString(r) {
					return fmt.Errorf("%s: preset %q at %s names %q, which is not a record id",
						PresetConfigPath, name, pos, r)
				}
			}
			for _, raw := range ps.Paths {
				if err := validPresetPath(raw); err != nil {
					return fmt.Errorf("%s: preset %q at %s: %w", PresetConfigPath, name, pos, err)
				}
			}
		}
	}
	return nil
}

// validPresetPath bounds the one place a repository path may be named. A preset
// cannot reach what the table denies — the filter runs after collection, so a
// denied path was never a candidate — but a path that LOOKS like it reaches
// there is refused at the door rather than silently selecting nothing.
func validPresetPath(raw string) error {
	if raw == "" {
		return fmt.Errorf("names an empty path")
	}
	if filepath.IsAbs(raw) || strings.HasPrefix(raw, "/") || strings.HasPrefix(raw, "~") {
		return fmt.Errorf("names the absolute path %q; preset paths are repo-relative", raw)
	}
	// A path carrying a control character renders into a terminal line, and
	// invariant 13 holds every runtime-read string to being sanitised before it
	// gets there. Refusing it at the door is better than sanitising it at every
	// render site that will ever exist.
	for _, r := range raw {
		if r < 0x20 || r == 0x7f {
			return fmt.Errorf("names a path carrying a control character")
		}
	}
	if strings.TrimSpace(raw) != raw {
		return fmt.Errorf("names %q, which is surrounded by whitespace", raw)
	}
	if strings.Contains(raw, "\\") {
		return fmt.Errorf("names %q, which uses backslash separators; paths are POSIX and "+
			"repo-relative", raw)
	}
	clean := path.Clean(raw)
	if clean == ".." || strings.HasPrefix(clean, "../") {
		return fmt.Errorf("names %q, which escapes the repository", raw)
	}
	// "." selects the whole repository, so it is a scope that narrows nothing.
	// A scope exists to narrow; an all-selecting path clause is either a
	// mistake or a way to spell "everything" that review would not read as one.
	if clean == "." {
		return fmt.Errorf("names %q, which selects the whole repository; a scope narrows, so "+
			"name the kinds you want rather than a path that excludes nothing", raw)
	}
	if pathContainsDeniedSegment(clean) || prefixDenied(clean) {
		return fmt.Errorf("names %q, which the structural deny covers; a preset cannot "+
			"reach what the table denies", raw)
	}
	return nil
}

// ResolveScope turns one invocation token into the scope a run had.
//
// The three token forms are tried in a fixed order, but the order is belt and
// braces rather than the rule: validatePresets refuses a preset whose name
// collides with a kind or a record-id shape, so at most one form can match.
func ResolveScope(pf PresetFile, position Position, token string) (Scope, error) {
	switch {
	case token == "":
		return Scope{}, fmt.Errorf("no scope named: a reading is commissioned ABOUT something. " +
			"Name a record id (itd-N, spc-N, adr-N, iss-N), a material kind, or a committed preset")

	case recordIDRe.MatchString(token):
		return Scope{
			Source:     token,
			Selectors:  canonicalise([]Selector{{Record: token}}),
			Overridden: true,
		}, nil

	case isKnownKind(token):
		return Scope{
			Source:     token,
			Selectors:  canonicalise([]Selector{{Kind: Kind(token)}}),
			Overridden: true,
		}, nil
	}

	preset, ok := pf.Presets[token]
	if !ok {
		return Scope{}, fmt.Errorf("unknown scope %q: name a record id (itd-N, spc-N, adr-N, "+
			"iss-N), a material kind (%s), or a committed preset (%s)",
			token, joinKinds(), joinPresets(pf))
	}

	sels := presetSelectors(pf, preset, position)
	if len(sels) == 0 {
		return Scope{}, fmt.Errorf("preset %q names no scope at the %s position, so it would "+
			"assemble nothing; a scope that selects nothing is a refusal rather than an empty bundle",
			token, position)
	}
	// Naming a committed preset is running as reviewed, so it is not an
	// override. The stamp counts departures from what was reviewed.
	return Scope{Source: token, Selectors: sels, Overridden: false}, nil
}

// presetSelectors resolves one preset at one position, unioning the parent it
// extends. The union is the whole of the warm-contains-cold guarantee.
func presetSelectors(pf PresetFile, p Preset, position Position) []Selector {
	var sels []Selector
	if p.Extends != "" {
		if parent, ok := pf.Presets[p.Extends]; ok {
			sels = append(sels, positionSelectors(parent, position)...)
		}
	}
	sels = append(sels, positionSelectors(p, position)...)
	return canonicalise(sels)
}

// positionSelectors flattens one preset's own entry at one position.
func positionSelectors(p Preset, position Position) []Selector {
	ps, ok := p.Positions[string(position)]
	if !ok {
		return nil
	}
	out := make([]Selector, 0, len(ps.Kinds)+len(ps.Records)+len(ps.Paths))
	for _, k := range ps.Kinds {
		out = append(out, Selector{Kind: k})
	}
	for _, r := range ps.Records {
		out = append(out, Selector{Record: r})
	}
	for _, raw := range ps.Paths {
		out = append(out, Selector{Path: path.Clean(raw)})
	}
	return out
}

// isKnownKind reports whether a token is a material kind.
func isKnownKind(token string) bool {
	for _, k := range Kinds() {
		if string(k) == token {
			return true
		}
	}
	return false
}

// joinKinds and joinPresets render the closed sets a refusal names, so an
// operator is told what IS accepted rather than only what was not.
func joinKinds() string {
	out := make([]string, 0, len(Kinds()))
	for _, k := range Kinds() {
		out = append(out, string(k))
	}
	return strings.Join(out, ", ")
}

func joinPresets(pf PresetFile) string {
	out := make([]string, 0, len(pf.Presets))
	for name := range pf.Presets {
		out = append(out, name)
	}
	sort.Strings(out)
	return strings.Join(out, ", ")
}
