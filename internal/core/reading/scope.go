package reading

// scope.go is what a reading is ABOUT.
//
// The assembler shipped able to name a position and a commit and nothing else,
// which meant every reading received its position's whole corpus: about 9.2 MB
// and 2.4 million estimated tokens over this repository, with three of the four
// positions receiving a byte-identical item set. An instrument that cannot be
// pointed at anything cannot be given to a reader.
//
// A preset entry NARROWS and never widens. It intersects what the include table
// already admits at the position, so no entry reaches a row the table denies
// there, and the structural deny, the exclusion floor and the dirty gate all
// run over the unfiltered walk before this filter is applied (itd-199, spc-69).
//
// WHICH entry applies is not an operand. The design fixes the invocation at a
// position and a target state (framework v4 section 8.2 and ruling M8;
// companion v4 section 4.1), and adr-2609021016286571 supersedes adr-58
// accordingly: itd-199's scope operand is withdrawn, and the assembler applies
// the committed entry for the position it was invoked at. What the operand did
// is not lost — the presets already map each position to what it reads, and a
// preset is a record fact the repository supplies rather than something typed.
// Changing what a position reads is a commit to the preset file, reviewed and
// inside the dirty gate. There is no override at invocation and nothing to
// stamp, and no repository path is accepted at the invocation: a path may be
// named only inside a committed preset.

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
//
// It names ONLY the families the include table can admit. `adr-N` and `iss-N`
// were in it, and were dead: no row names a decisions store or an issue store,
// so those tokens validated, resolved, selected nothing, and ended in the
// selects-nothing refusal at every position — while the CLI help, the plugin
// page and the generated reference all advertised them. An affordance that can
// never work is worse than an absent one, because the operator's first move is
// to doubt their own invocation. If the table ever admits those families, this
// regex is where they come back.
var recordIDRe = regexp.MustCompile(`^(itd|spc)-[0-9]+$`)

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

// AppliedPreset is what one run was about: the committed entry for the invoked
// position, resolved to selectors.
//
// It carries no source token and no override stamp, and that absence is the
// point. Both were provenance about a CHOICE the operator made at the
// invocation, and there is no such choice: the entry is settled by the position
// and by what is committed (adr-2609021016286571). A run is reproducible from
// the commit the manifest names and the entry it records.
type AppliedPreset struct {
	// Selectors are the resolved clauses, in a canonical order so one entry
	// hashes to one value.
	Selectors []Selector `json:"selectors"`
}

// selects reports whether one collected candidate is in the applied entry.
func (s AppliedPreset) selects(c candidate) bool {
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

// canonicalise sorts an entry's selectors and drops duplicates, so two entries
// naming the same thing in a different order are one entry and hash alike.
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

// Hash is the applied entry's content hash, so a manifest can name the entry a
// run actually had and a reader can tell two runs apart by it. It also means a
// preset edited later can never make a past run unreadable: the run names the
// entry it applied, not the file as it stands now.
func (s AppliedPreset) Hash() (string, error) {
	data, err := json.Marshal(struct {
		Selectors []Selector `json:"selectors"`
	}{Selectors: s.Selectors})
	if err != nil {
		return "", fmt.Errorf("hashing the applied reading preset: %w", err)
	}
	return sha256Hex(data), nil
}

// PositionScope is one position's scope inside a preset.
type PositionScope struct {
	Kinds   []Kind   `json:"kinds"`
	Records []string `json:"records"`
	Paths   []string `json:"paths"`
}

// Preset holds the committed entries, one per position.
//
// `extends` retired with the second preset name. It existed to make "warm is
// cold plus a delta" a property rather than a review note, and there is no
// second name for it to relate: one entry per position stands, and a repository
// that wants a wider reading commits a wider entry (adr-2609021016286571).
type Preset struct {
	// Positions maps a position token to its entry. A preset carries an entry
	// PER POSITION rather than one every position shares, because the finding
	// this exists to fix is that three of the four positions received a
	// byte-identical item set, and one entry over four near-identical
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
	// trackedSet reads the git INDEX, not HEAD, so on its own this establishes
	// "known to git" rather than "committed". The committed property is
	// delivered by this check together with the dirty gate downstream, which
	// refuses an added-but-uncommitted preset as an `A ` entry. Both are needed:
	// the dirty gate alone missed an ignored file, and this check alone would
	// admit a staged one.
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
	// One preset file, one entry per position (adr-2609021016286571). Nothing
	// at the invocation names a preset any more, so a file holding two has
	// nothing to choose between them: whichever the loader picked would be a
	// resolution ORDER deciding silently, which is the failure this package
	// refuses everywhere else it can arise. A repository with two calibrations
	// commits one and records the other in its history
	// (cond-2609021004074586).
	if len(pf.Presets) > 1 {
		return PresetFile{}, fmt.Errorf("%s names %d presets (%s); the invocation names none, so "+
			"there is nothing to choose between them. Commit one entry per position and keep the "+
			"other in the file's history", PresetConfigPath, len(pf.Presets), joinPresets(pf))
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
		// The name is no longer a token anything resolves — nothing at the
		// invocation names a preset — so these three refusals are now about the
		// FILE rather than about resolution: a key named for a material kind or
		// shaped like a record id reads, to a reviewer, as the thing it is
		// named after rather than as the entry set, and the one file whose
		// whole safety argument is that a human reviewed it is the last place
		// to leave a name that means two things.
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
		for pos, ps := range p.Positions {
			position, err := ParsePosition(pos)
			if err != nil {
				return fmt.Errorf("%s: preset %q: %w", PresetConfigPath, name, err)
			}
			// The comparative position refuses before the presets are loaded at
			// all, so an entry for it would describe a run that cannot happen.
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

// PresetFor returns the committed entry for one position.
//
// It takes no token, because there is none to take: the invocation is a
// position and a target state, and which entry applies follows from the
// position alone (adr-2609021016286571). A file naming no entry for the
// position refuses rather than defaulting to everything — a position served the
// whole corpus because its entry was forgotten is exactly the silent widening
// the presets exist to close.
func PresetFor(pf PresetFile, position Position) (AppliedPreset, error) {
	preset, err := solePreset(pf)
	if err != nil {
		return AppliedPreset{}, err
	}
	sels := canonicalise(positionSelectors(preset, position))
	if len(sels) == 0 {
		return AppliedPreset{}, fmt.Errorf("%s names no entry for the %s position, so a run there "+
			"would assemble nothing; an entry that selects nothing is a refusal rather than an "+
			"empty bundle. Commit an entry for %s", PresetConfigPath, position, position)
	}
	return AppliedPreset{Selectors: sels}, nil
}

// solePreset returns the file's one preset. LoadPresets already refuses a file
// holding none or more than one, so this is the reader's half of that rule and
// never the place the count is decided.
func solePreset(pf PresetFile) (Preset, error) {
	if len(pf.Presets) != 1 {
		return Preset{}, fmt.Errorf("%s holds %d presets; one entry per position means one preset",
			PresetConfigPath, len(pf.Presets))
	}
	for _, p := range pf.Presets {
		return p, nil
	}
	return Preset{}, fmt.Errorf("%s holds no preset", PresetConfigPath)
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

// joinPresets renders the preset names a refusal reports, so a file holding
// more than one says which ones rather than only how many.
func joinPresets(pf PresetFile) string {
	out := make([]string, 0, len(pf.Presets))
	for name := range pf.Presets {
		out = append(out, name)
	}
	sort.Strings(out)
	return strings.Join(out, ", ")
}
