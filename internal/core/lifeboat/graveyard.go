package lifeboat

import "regexp"

// The graveyard is the lifeboat's record of what a project tried and left behind
// (adr-35, M4). It has three layers, strictly ordered:
//
//   - Layer 1 — Archaeology (graveyard/archaeology.json): Tier-0 git evidence,
//     deterministic, no interpretation.
//   - Layer 2 — Recorded abandonment (graveyard/abandoned.json): what the project
//     itself declared dead (superseded intents/ADRs, wontfix issues, rejected
//     options).
//   - Layer 3 — Interpretation (graveyard/lessons.json): host-delegated LLM
//     output, validated by a Go cite-or-be-dropped gate (IngestLessons).
//
// This file is the shared vocabulary every layer builds on: the finding/lesson
// types, the schema constants, the id grammar, and the safety caps. Layers 1 and
// 2 are pure functions over a *SourceContext, emitted inside Plan (pack-time,
// part of manifest_sha256). Layer 3 is a separate CLI verb over an already-packed
// lifeboat and is deliberately NOT part of the manifest hash.
const (
	// GraveyardSchemaVersion is the schema of archaeology.json / abandoned.json.
	GraveyardSchemaVersion = 1
	// LessonsSchemaVersion is the schema of lessons.json / low-confidence/*.json.
	LessonsSchemaVersion = 1
)

// Signal names the kind of abandonment a Finding records. Layer-1 (git) signals
// and layer-2 (record) signals share one namespace but never collide, because
// each finding id is prefixed by its signal family (see the id grammar).
type Signal string

const (
	// Layer 1 — git archaeology.
	SignalRevert            Signal = "revert"
	SignalUnmergedBranch    Signal = "unmerged-branch"
	SignalDeletedPath       Signal = "deleted-path"
	SignalRemovedDependency Signal = "removed-dependency"
	SignalWholesaleRewrite  Signal = "wholesale-rewrite"

	// Layer 2 — recorded abandonment.
	SignalSupersededIntent Signal = "superseded-intent"
	SignalSupersededADR    Signal = "superseded-adr"
	SignalAlternatives     Signal = "alternatives-considered"
	SignalWontfixIssue     Signal = "wontfix-issue"
	SignalRejectedOption   Signal = "rejected-option"
)

// signalRank is the fixed grouping order findings are emitted in — never a global
// re-sort, so each signal's own deterministic within-group ordering survives.
var signalRank = map[Signal]int{
	SignalRevert:            0,
	SignalUnmergedBranch:    1,
	SignalDeletedPath:       2,
	SignalRemovedDependency: 3,
	SignalWholesaleRewrite:  4,
	SignalSupersededIntent:  5,
	SignalSupersededADR:     6,
	SignalAlternatives:      7,
	SignalWontfixIssue:      8,
	SignalRejectedOption:    9,
}

// Finding is one piece of graveyard evidence: a stable, signal-namespaced id, the
// signal that produced it, a one-line summary, and the verbatim-but-sanitised
// evidence lines that justify it. Layer-3 lessons cite a Finding by its id.
type Finding struct {
	ID       string   `json:"id"`
	Signal   Signal   `json:"signal"`
	Summary  string   `json:"summary"`
	Evidence []string `json:"evidence"`
}

// Archaeology is the layer-1 dig: deterministic git evidence, grouped in
// signalRank order. Findings is always non-nil in a marshalled file ("[]", never
// null), so an empty dig is an honest, stable datum.
type Archaeology struct {
	SchemaVersion int       `json:"schema_version"`
	Findings      []Finding `json:"findings"`
}

// Abandoned is the layer-2 record: what the project itself declared dead.
type Abandoned struct {
	SchemaVersion int       `json:"schema_version"`
	Findings      []Finding `json:"findings"`
}

// Lesson is one interpreted lesson (layer 3). Its Evidence cites layer-1/2
// finding ids; the cite-or-be-dropped gate keeps only the refs that resolve.
type Lesson struct {
	ID         string     `json:"id"`
	Lesson     string     `json:"lesson"`
	Confidence Confidence `json:"confidence"`
	Evidence   []string   `json:"evidence"`
}

// LessonsFile is the on-disk shape of lessons.json and every
// low-confidence/<id>.json — the input the interpreter emits AND the sanitised
// output the verb writes.
type LessonsFile struct {
	SchemaVersion int      `json:"schema_version"`
	Lessons       []Lesson `json:"lessons"`
}

// LessonDrop records one lesson the gate refused, and why — a drop is reported,
// never fatal, so a batch survives its worst entry.
type LessonDrop struct {
	ID     string `json:"id"`
	Reason string `json:"reason"`
}

// LessonsResult is the transport-agnostic outcome of IngestLessons: what was
// written, what was routed to low-confidence, and what was dropped. It never
// prints; a front door renders it.
type LessonsResult struct {
	Lifeboat      string       `json:"lifeboat"`
	Written       int          `json:"written"`
	LowConfidence int          `json:"low_confidence"`
	Dropped       int          `json:"dropped"`
	Drops         []LessonDrop `json:"drops,omitempty"`
}

// collectFindingIDs is the live id set a layer-3 evidence ref must hit to
// survive: the union of every layer-1 and layer-2 finding id. Empty ids are
// skipped (a malformed finding cannot be cited).
func collectFindingIDs(sets ...[]Finding) map[string]bool {
	ids := map[string]bool{}
	for _, s := range sets {
		for _, f := range s {
			if f.ID != "" {
				ids[f.ID] = true
			}
		}
	}
	return ids
}

// lessonIDRe is the whole path-traversal defence for the low-confidence write:
// a lesson id must be a kebab-case token under the "les-" family, so
// low-confidence/<id>.json can never carry a separator, "..", or a control
// character and thus can never escape the low-confidence directory.
var lessonIDRe = regexp.MustCompile(`^les-[a-z0-9]+(?:-[a-z0-9]+)*$`)

// Layer-3 safety caps. The lesson payload is untrusted model output a hostile or
// archived repo can influence, so every dimension is bounded: the whole payload
// (maxLessonsBytes), the number of entries (maxLessons), each id (maxLessonIDLen),
// each prose body (maxLessonProseBytes), and the evidence refs read per entry
// (maxLessonEvidenceRefs). maxGraveyardFileBytes bounds a packed layer-1/2 file.
const (
	maxLessonsBytes       = 1 << 20 // 1 MiB — the whole lessons payload
	maxLessons            = 500     // entries per batch
	maxLessonIDLen        = 128
	maxLessonProseBytes   = 4000
	maxLessonEvidenceRefs = 64
	maxGraveyardFileBytes = 8 << 20 // 8 MiB — one packed archaeology/abandoned file
)

// MaxLessonsBytes is the exported cap a front door uses to bound its read of the
// untrusted lessons payload (file or stdin) before handing it to IngestLessons.
// Exporting the single constant — rather than a whole reader helper — is the
// smaller surface: the CLI already owns the trust-guard reading idiom it shares
// with --verdict-json, and needs only the ceiling value.
const MaxLessonsBytes = maxLessonsBytes
