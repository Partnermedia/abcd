// Package capture is abcd's transport-agnostic issue-ledger engine: the write
// side of a per-repo issue ledger that replaces the free-form .work/issues.md.
// Every capability is a function taking a structured request and returning a
// structured result; nothing here writes to stdout or knows about a CLI, MCP,
// or prompt surface. The front doors under internal/surface/* marshal these
// results for their transport.
//
// The ledger lives at <repoRoot>/.abcd/work/issues with three
// status directories (open/, resolved/, wontfix/) whose folder membership IS
// the status signal — there is no status: frontmatter field. Each issue is a
// YAML-frontmatter + Markdown-body file named iss-<N>-<slug>.md with an
// unpadded, per-repo id namespace.
//
// This package ports scripts/abcd/_issue_lib.py + issue_workflow.py to Go.
package capture

import (
	"errors"
	"regexp"
)

// LedgerRelPath is the ledger root relative to the repo worktree.
const LedgerRelPath = ".abcd/work/issues"

// Enumerated field types (validated at the boundary; values mirror
// scripts/abcd/schemas/issue.schema.json).
type (
	// Severity is the capture-time severity guess.
	Severity string
	// Category is the loose issue taxonomy.
	Category string
	// Source is the surfacing channel the issue was discovered through.
	Source string
	// State is a ledger status directory (or "all" for a cross-status scan).
	State string
)

// Severity enum values.
const (
	SeverityNitpick  Severity = "nitpick"
	SeverityMinor    Severity = "minor"
	SeverityMajor    Severity = "major"
	SeverityCritical Severity = "critical"
)

// State enum values.
const (
	StateAll      State = "all"
	StateOpen     State = "open"
	StateResolved State = "resolved"
	StateWontfix  State = "wontfix"
)

var validSeverities = map[Severity]bool{
	SeverityNitpick: true, SeverityMinor: true,
	SeverityMajor: true, SeverityCritical: true,
}

var validCategories = map[Category]bool{
	"bug": true, "documentation": true, "drift": true, "inconsistency": true,
	"tech-debt": true, "security": true, "ux": true, "process": true,
	"architectural-insight": true, "future-work-seed": true, "observation": true,
}

var validSources = map[Source]bool{
	"plan-review": true, "impl-review": true, "manual-test": true,
	"review-followup": true, "agent-finding": true, "user-observation": true,
	"drift-detection": true, "memory-curation": true,
}

// ResolvedBy is an optional structured pointer to what resolved an issue.
type ResolvedBy struct {
	Intent string `json:"intent,omitempty"`
	Spec   string `json:"spec,omitempty"`
	Commit string `json:"commit,omitempty"`
}

// Issue is a fully-read ledger entry (frontmatter + provenance + body).
type Issue struct {
	SchemaVersion  int         `json:"schema_version"`
	ID             string      `json:"id"`
	Slug           string      `json:"slug"`
	Severity       Severity    `json:"severity"`
	Category       Category    `json:"category"`
	Source         Source      `json:"source"`
	FoundDuring    string      `json:"found_during"`
	FoundAt        string      `json:"found_at,omitempty"`
	RelatedIntents []string    `json:"related_intents,omitempty"`
	RelatedSpecs   []string    `json:"related_specs,omitempty"`
	RelatedIssues  []string    `json:"related_issues,omitempty"`
	BlockedBy      []string    `json:"blocked_by,omitempty"` // iss-N dependency edges
	PromotedTo     string      `json:"promoted_to,omitempty"`
	Resolution     string      `json:"resolution,omitempty"`
	WontfixReason  string      `json:"wontfix_reason,omitempty"`
	ResolvedBy     *ResolvedBy `json:"resolved_by,omitempty"`
	Status         State       `json:"status"` // derived from folder
	Path           string      `json:"path"`   // repo-relative locator (iss-81)
	Body           string      `json:"body"`
	// BlockedByOpen is the derived subset of BlockedBy whose targets are still in
	// open/ (the priority projection populated by List/Status). Not a stored
	// field: an empty slice means the issue is unblocked.
	BlockedByOpen []string `json:"blocked_by_open,omitempty"`
}

// CaptureRequest is the input to Capture (append a new issue).
type CaptureRequest struct {
	RepoRoot       string
	IssuesRoot     string
	Text           string // markdown body
	Severity       Severity
	Category       Category
	Source         Source
	Slug           string // caller-supplied; normalised to kebab-case
	FoundDuring    string // required, non-empty
	FoundAt        string // optional; "" omits the field
	RelatedIntents []string
	RelatedSpecs   []string
	BlockedBy      []string // iss-N dependency edges; each must match ^iss-[0-9]+$
	ForceID        string   // migrator-only; "" = auto-allocate
}

// CaptureResult is the outcome of a successful Capture.
type CaptureResult struct {
	ID     string `json:"id"`
	Slug   string `json:"slug"`
	Path   string `json:"path"`
	Status State  `json:"status"` // always "open"
}

// ResolveRequest moves an open issue to resolved/.
type ResolveRequest struct {
	RepoRoot   string
	IssuesRoot string
	ID         string
	Resolution string
}

// WontfixRequest moves an open issue to wontfix/.
type WontfixRequest struct {
	RepoRoot   string
	IssuesRoot string
	ID         string
	Reason     string
}

// TransitionResult is the outcome of a Resolve or Wontfix.
type TransitionResult struct {
	ID         string `json:"id"`
	Path       string `json:"path"`
	FromStatus State  `json:"from_status"`
	ToStatus   State  `json:"to_status"`
}

// ListRequest queries one state (or "all").
type ListRequest struct {
	RepoRoot   string
	IssuesRoot string
	State      State // "" is treated as "all"
}

// SkipRecord surfaces a corrupt/invalid ledger file without failing the scan.
type SkipRecord struct {
	Path  string `json:"path"`
	Error string `json:"error"`
}

// ListResult is Issues sorted ascending by numeric N plus a corrupt roster.
type ListResult struct {
	Issues  []Issue      `json:"issues"`
	Skipped []SkipRecord `json:"skipped"`
}

// StatusRequest is the input to the read-only status render.
type StatusRequest struct {
	RepoRoot   string
	IssuesRoot string
}

// StatusResult is the bare-invocation status snapshot (guaranteed no mutation).
type StatusResult struct {
	OpenCount     int          `json:"open_count"`
	ResolvedCount int          `json:"resolved_count"`
	WontfixCount  int          `json:"wontfix_count"`
	RecentOpen    []Issue      `json:"recent_open"` // up to 10, newest first
	Skipped       []SkipRecord `json:"skipped"`
}

// Sentinel errors the surface maps to exit codes and messages. Core never
// prints them.
var (
	// ErrUnknownIssueID means the id was absent from all three dirs.
	ErrUnknownIssueID = errors.New("unknown issue id")
	// ErrTransitionConflict means the id was found but not in open/ (already
	// resolved/wontfixed), or a concurrent move consumed it.
	ErrTransitionConflict = errors.New("transition conflict")
	// ErrDuplicateIssueID means a ForceID (or on-disk state) collided.
	ErrDuplicateIssueID = errors.New("duplicate issue id")
	// ErrAllocatorContention means the lock timed out or the O_EXCL retry
	// budget was exhausted.
	ErrAllocatorContention = errors.New("allocator contention")
	// ErrChecksumMismatch means a concurrent edit occurred during a transition.
	ErrChecksumMismatch = errors.New("checksum mismatch")
	// ErrInvariantViolation means frontmatter passed the schema but violates a
	// folder-status cross-field invariant.
	ErrInvariantViolation = errors.New("invariant violation")
	// ErrMalformedFrontmatter means frontmatter could not be parsed or failed
	// schema validation.
	ErrMalformedFrontmatter = errors.New("malformed frontmatter")
	// ErrMissingRequiredField means a schema-required field was absent.
	ErrMissingRequiredField = errors.New("missing required field")
	// ErrPathUnsafe means the ledger root or a status dir is a symlink.
	ErrPathUnsafe = errors.New("path unsafe")
)

// Field regexes mirroring issue.schema.json.
var (
	reIssID       = regexp.MustCompile(`^iss-[0-9]+$`)
	reItdID       = regexp.MustCompile(`^itd-[0-9]+$`)
	reFnID        = regexp.MustCompile(`^fn-[0-9]+$`)
	reSlug        = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)
	reFilenameID  = regexp.MustCompile(`^(iss-[0-9]+)(?:-[a-z0-9]+(?:-[a-z0-9]+)*)?\.md$`)
	reAbcdListID  = regexp.MustCompile(`^(itd|fn|iss)-[0-9]+$`)
	reSortIssID   = regexp.MustCompile(`^iss-([0-9]+)(-|$|\.)`)
	reScalarKey   = regexp.MustCompile(`^[a-z][a-z0-9_]*$`)
	statusDirs    = [3]State{StateOpen, StateResolved, StateWontfix}
	statusDirName = map[State]string{StateOpen: "open", StateResolved: "resolved", StateWontfix: "wontfix"}
)
