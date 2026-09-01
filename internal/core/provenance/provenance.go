// Package provenance is the record's disclosure vocabulary as DATA: where an
// item came from (`origin`) and how its text was produced (`production_mode`),
// plus the one parser that reads and renders them.
//
// It is a leaf for the same reason core/issueschema and core/changelog are: the
// WRITERS that stamp the pair (core/intent, core/spec, core/capture) and the
// GATE that judges it (core/lint) must agree about what a legal value is, and two
// hand-kept copies of a closed set drift the moment one side gains a member. It
// imports no filesystem and no transport — only core/issueschema, for the reading
// families' own spelling of their id prefixes, which is not restated here.
//
// Neither key touches authorship. They are disclosure at field granularity, on
// the same footing as the `Assisted-by:` trailer at commit granularity (itd-178).
package provenance

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/intentdriven/abcd/internal/core/issueschema"
)

// The two frontmatter keys, spelled once. Both are single-line SCALARS, which is
// a constraint on the encoding rather than an accident: a nested mapping is
// readable by capture's YAML subset but invisible to frontmatter.Fields, the
// same-line scanner every intent and spec reader uses, so a nested value would
// need a second record parser (spc-56).
const (
	KeyOrigin         = "origin"
	KeyProductionMode = "production_mode"
)

// Kind is the arrival path an `origin` value names.
type Kind string

// The three members of the closed set. The string values are the wire format —
// what the frontmatter carries — so renaming a constant is safe and changing a
// value is a record migration.
const (
	// KindResearcherAuthored is the default for a verb a person invoked.
	KindResearcherAuthored Kind = "researcher-authored"
	// KindExtractedFromRecord is stamped by capture.Promote, the one shipped path
	// that derives a record from another record.
	KindExtractedFromRecord Kind = "extracted-from-record"
	// KindContributedByReading carries the run and item identifiers that resolve
	// to a reading record. It is minted only by the reading-ingest verb, which
	// carries both ids already and never asks an operator for them.
	KindContributedByReading Kind = "contributed-by-reading"
)

// The reading pointer's two id shapes, built from the reading families' own
// prefixes rather than from a second spelling of them.
var (
	readingRunRe  = regexp.MustCompile(`^` + issueschema.ReadingRunFamily + `-[0-9]+$`)
	readingItemRe = regexp.MustCompile(`^` + issueschema.ReadingItemFamily + `-[0-9]+$`)
)

// Origin is a parsed `origin` value: the arrival path, plus the reading pointer
// when the path is a reading contribution. Run and Item are empty for every
// other kind, so a zero pointer beside a reading kind is unrepresentable after a
// successful parse.
type Origin struct {
	Kind Kind
	Run  string
	Item string
}

// String renders the value exactly as a record carries it, so a parse and a
// render round-trip byte-for-byte. The rendered value contains no ": " sequence,
// which is what keeps it a plain YAML scalar to every reader in the corpus.
func (o Origin) String() string {
	if o.Kind == KindContributedByReading {
		return string(o.Kind) + " " + o.Run + "/" + o.Item
	}
	return string(o.Kind)
}

// ParseOrigin is the ONE predicate that reads an `origin` value, and the one the
// lint calls.
//
// Matching is exact: no case folding and no trimming, matching how the shared
// impact parser reads its enum. The frontmatter scanner already trims the value
// it hands over, so whitespace still present here came from inside the value and
// is an authoring defect worth naming rather than absorbing.
func ParseOrigin(v string) (Origin, error) {
	switch Kind(v) {
	case KindResearcherAuthored, KindExtractedFromRecord:
		return Origin{Kind: Kind(v)}, nil
	}
	rest, isReading := strings.CutPrefix(v, string(KindContributedByReading)+" ")
	if !isReading {
		return Origin{}, fmt.Errorf("invalid origin %q: want %s, %s, or %s <%s-N>/<%s-N>",
			v, KindResearcherAuthored, KindExtractedFromRecord, KindContributedByReading,
			issueschema.ReadingRunFamily, issueschema.ReadingItemFamily)
	}
	run, item, hasSlash := strings.Cut(rest, "/")
	if !hasSlash || !readingRunRe.MatchString(run) || !readingItemRe.MatchString(item) {
		return Origin{}, fmt.Errorf(
			"invalid origin %q: %s must carry a reading pointer spelled <%s-N>/<%s-N>, so the run and the item both resolve to a reading record",
			v, KindContributedByReading, issueschema.ReadingRunFamily, issueschema.ReadingItemFamily)
	}
	return Origin{Kind: KindContributedByReading, Run: run, Item: item}, nil
}

// Mode is how a record's text was produced.
type Mode string

// The three members of the closed set (the ruled authorship account's
// vocabulary; spc-56 ships the mechanism that stamps it, not the vocabulary
// decision).
const (
	ModeHandWritten          Mode = "hand-written"
	ModeDictatedAndFormatted Mode = "dictated-and-formatted"
	ModeScribeTranscribed    Mode = "scribe-transcribed"
)

// DefaultMode is what an unset production mode means: a repository that declares
// none, and a verb invoked with no `--production-mode`, both stamp this.
const DefaultMode = ModeHandWritten

// modeValues lists the set in the order the error messages name it, so the legal
// values are written once and every message stays in step with the constants.
var modeValues = []Mode{ModeHandWritten, ModeDictatedAndFormatted, ModeScribeTranscribed}

// ParseMode is the boundary validator for a PRESENT production mode. An empty
// value is an error here, never a default: the defaulting rule has exactly one
// door (ModeOrDefault), so a caller cannot absorb an unset value by accident and
// a config member that is present but blank is a fault rather than a shrug.
func ParseMode(v string) (Mode, error) {
	for _, known := range modeValues {
		if Mode(v) == known {
			return known, nil
		}
	}
	return "", fmt.Errorf("invalid production mode %q: want exactly one of %s (lower-case, no surrounding whitespace)", v, ModeList())
}

// ModeOrDefault is the ONE defaulting door: an absent value means DefaultMode,
// and anything else goes through ParseMode. Every write path normalises here, so
// the "absent means hand-written" rule is stated once rather than at each
// writer.
func ModeOrDefault(v string) (Mode, error) {
	if v == "" {
		return DefaultMode, nil
	}
	return ParseMode(v)
}

// ModeList renders the legal set for an error message or a flag's help text.
func ModeList() string {
	parts := make([]string, 0, len(modeValues))
	for _, m := range modeValues {
		parts = append(parts, string(m))
	}
	return strings.Join(parts, "|")
}

// Stamp is the pair a write path stamps: both keys together, never one of them.
// Every mint writes both, which is what makes a lone key a state no command
// could have produced.
type Stamp struct {
	Origin Origin
	Mode   Mode
}

// NewStamp validates a kind and a possibly-empty production mode, and is the
// only constructor a write path uses.
//
// It refuses KindContributedByReading outright. That value carries a pointer to
// a reading record, and the reading-ingest verb that holds those identifiers does
// not exist in this repository yet — so nothing here can supply them, and a
// constructor that accepted the kind would let a caller mint a dangling pointer.
// The lint's resolution check for the value is exercised by fixture until the
// verb ships (spc-56, stated rather than discovered later).
func NewStamp(kind Kind, mode string) (Stamp, error) {
	switch kind {
	case KindResearcherAuthored, KindExtractedFromRecord:
	case KindContributedByReading:
		return Stamp{}, fmt.Errorf(
			"origin %s is minted only by the reading-ingest verb, which carries the run and item identifiers; no write path in this repository can supply them", kind)
	default:
		return Stamp{}, fmt.Errorf("unknown origin kind %q: want %s or %s",
			kind, KindResearcherAuthored, KindExtractedFromRecord)
	}
	m, err := ModeOrDefault(mode)
	if err != nil {
		return Stamp{}, err
	}
	return Stamp{Origin: Origin{Kind: kind}, Mode: m}, nil
}

// OriginValue and ModeValue render the two frontmatter values. Writers ask the
// stamp for them rather than reaching into it, so the rendering of the pair has
// one home.
func (s Stamp) OriginValue() string { return s.Origin.String() }

// ModeValue renders the production-mode value.
func (s Stamp) ModeValue() string { return string(s.Mode) }
