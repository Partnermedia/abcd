// Package ahoy is abcd's install/update engine for `abcd ahoy`. It follows a
// detect -> contract -> apply architecture: a single detection pass builds an
// in-memory DetectionResult (the ahoy-state shape) and every sub-verb (install,
// dry-run, doctor, bare status) is a thin consumer of that one pass. Idempotency
// is a property of detection, never a version stamp: every check compares actual
// on-disk/registry state.
//
// The package performs I/O only under a caller-supplied cwd and the user-scope
// ~/.abcd/ store (and, on install, an owned PATH symlink). It never writes to
// stdout, never calls os.Exit, and never imports a transport (cobra/MCP), so it
// is fully testable and reusable across surfaces. Interactive decisions are
// routed through the injected Prompter seam.
package ahoy

// FolderKind is the classification of the folder ahoy runs in. There is no
// workspace layer: abcd manages exactly one kind of folder, a repository.
type FolderKind string

const (
	// ManagedRepo is a git repo abcd already manages (a strong marker fired).
	ManagedRepo FolderKind = "managed-repo"
	// UnmanagedRepo is a git repo with no abcd markers yet (adoptable).
	UnmanagedRepo FolderKind = "unmanaged-repo"
	// UnmanagedFolder is not a git repo and has no abcd markers.
	UnmanagedFolder FolderKind = "unmanaged-folder"
)

// GapCategory groups gaps for the one-approval-per-category apply protocol.
type GapCategory string

const (
	// SafeAutocreate covers artefacts written without a per-item prompt once
	// the category is approved (the .abcd/ skeleton, history-store dirs).
	SafeAutocreate GapCategory = "safe-autocreate"
	// ConfigChange covers transparent-confirm changes (visibility, symlink).
	ConfigChange GapCategory = "config-change"
	// PluginOwned covers the marker block and the (verify-only) hook manifest.
	PluginOwned GapCategory = "plugin-owned"
	// Dependency covers opt-in scanners on PATH (surfaced, never auto-run).
	Dependency GapCategory = "dependency"
	// UserState covers ~/.abcd/history registry state (guided, never auto-edited).
	UserState GapCategory = "user-state"
)

// Gap is one detected discrepancy between desired and actual state.
type Gap struct {
	ID         string      `json:"id"`
	Category   GapCategory `json:"category"`
	Scope      string      `json:"scope"` // "repo" | "machine"
	Title      string      `json:"title"`
	Detail     string      `json:"detail"`
	FixHint    string      `json:"fix_hint"`
	Required   bool        `json:"required"`   // advisory gaps set false
	Resolvable bool        `json:"resolvable"` // false => diagnostic only
}

// RepoIdentity is the deterministic identity of the repo under cwd.
type RepoIdentity struct {
	Name    string `json:"name"`
	Github  string `json:"github"`
	RootSHA string `json:"root_sha"`
}

// DetectionResult is the canonical envelope. dry-run marshals exactly this
// (with Adopted=nil).
type DetectionResult struct {
	FolderKind       FolderKind     `json:"folder_kind"`
	Adopted          *bool          `json:"adopted"` // nil on detect/dry-run
	RootSHA          string         `json:"root_sha"`
	PluginRootStatus string         `json:"plugin_root_status"` // "resolved" | "missing"
	RepoIdentity     RepoIdentity   `json:"repo_identity"`
	Signals          map[string]any `json:"signals"`
	Gaps             []Gap          `json:"gaps"`

	// pluginRoot is the resolved plugin root; not serialized.
	pluginRoot string
}

// InstallConfig is the four configuration values collected/loaded during install.
type InstallConfig struct {
	Visibility    string // private | public
	DocsTarget    string // claude_md | agents_md | both | skip
	OracleBackend string // host-delegated | native | cli | api | mcp
	ScanDeep      *bool  // nil = unset (gap not emitted)
}

// InstallOptions encodes the non-interactive prompt-protocol flags.
type InstallOptions struct {
	Adopt              *bool                // --adopt / --refuse-adopt / nil
	Yes                bool                 // approve every resolvable category
	ApprovedCategories map[GapCategory]bool // nil => interactive; explicit => partial subset
	ValueOverrides     map[string]string    // visibility/docs_target/oracle_backend/scan_deep
}

// InstallResult is the outcome of Install.
type InstallResult struct {
	Status             string   `json:"status"` // already_up_to_date | clean | partial | aborted
	Writes             []string `json:"writes"`
	Remaining          []string `json:"remaining"`           // required+resolvable gap ids left
	DeclinedCategories []string `json:"declined_categories"` // sorted category wire values
}

// ApplyResult is the outcome of one apply step.
type ApplyResult struct {
	Category     GapCategory
	GapsResolved []string
	GapsSkipped  []string
	Notes        []string
}

// MarkerReceipt records the per-target outcome of the uninstall marker removal.
type MarkerReceipt struct {
	Removed []string `json:"removed"` // relative filenames a block was stripped from
	Skipped []string `json:"skipped"` // relative filenames left untouched
}

// SymlinkReceipt records the outcome of the uninstall symlink removal.
type SymlinkReceipt struct {
	Target  string `json:"target"`
	Removed bool   `json:"removed"`
	Note    string `json:"note"`
}

// UninstallReceipt is the outcome of Uninstall.
type UninstallReceipt struct {
	Marker  MarkerReceipt  `json:"marker"`
	Symlink SymlinkReceipt `json:"symlink"`
}

// DoctorReport is the outcome of Doctor: the detection envelope plus read-only
// cross-machine reconciliation gaps.
type DoctorReport struct {
	Detection DetectionResult `json:"detection"`
	AuditGaps []Gap           `json:"audit_gaps"`
}

// Prompter is the interactive seam. The CLI supplies interactive impls;
// non-interactive modes supply refusing impls that auto-decline rather than
// block on stdin.
type Prompter interface {
	// Confirm asks a yes/no question and returns the answer.
	Confirm(question string) bool
	// Prompt asks the user to pick one of choices, defaulting to def.
	Prompt(key string, choices []string, def string) string
}

// RefusingPrompter auto-declines every confirm and returns the default for
// every prompt. It is the safe default for non-interactive callers.
type RefusingPrompter struct{}

// Confirm always declines.
func (RefusingPrompter) Confirm(string) bool { return false }

// Prompt always returns the supplied default.
func (RefusingPrompter) Prompt(_ string, _ []string, def string) string { return def }
