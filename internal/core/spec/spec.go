// Package spec is abcd's transport-agnostic native spec store (intent
// lifecycle, itd-64). It owns the in-memory model of spec records and the
// disk operations that mint, load, and transition them. Every function takes a
// structured request and returns a structured result; nothing here writes to
// stdout or knows about a CLI, MCP, or hook surface — the front doors under
// internal/surface/* marshal these results for their transport.
//
// A spec record is a markdown file under
// <repoRoot>/.abcd/development/specs/{open,closed}/spc-N-<slug>.md. The bucket
// directory IS the status (directory-as-truth: open/ vs closed/, with no
// status: frontmatter field), mirroring how intents encode their lifecycle by
// bucket. The load-bearing field is intent: itd-N, the link a spec declares to
// the intent it realises.
//
// Frontmatter is read by the shared internal/core/frontmatter line scanner, not
// a YAML parser — the package pulls in zero new dependencies. Ids are validated
// against strict regexes before any path is built, so a hostile id can never
// traverse out of the spec store.
package spec

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// Bucket directory names. The directory a spec file lives in is its status.
const (
	StatusOpen   = "open"
	StatusClosed = "closed"
)

// Repo-relative store locations.
const (
	// SpecsRelDir is the spec-store root, relative to the repo worktree.
	SpecsRelDir = ".abcd/development/specs"
	// IntentsRelDir is where intents live; NextID scans it for reserved spec ids.
	IntentsRelDir = ".abcd/development/intents"
)

// maxSpecFileBytes caps any spec/intent markdown file read (trust boundary).
const maxSpecFileBytes = 256 * 1024

// intentBuckets are the intent lifecycle directories NextID scans for reserved
// spec_id values.
var intentBuckets = []string{"drafts", "planned", "shipped", "disciplines", "superseded"}

var (
	// specIDRe constrains a spec id so it can never be used to build a path that
	// escapes the store (path-traversal defence).
	specIDRe = regexp.MustCompile(`^spc-[0-9]+$`)
	// intentIDRe constrains the load-bearing intent link the same way.
	intentIDRe = regexp.MustCompile(`^itd-[0-9]+$`)
	// slugRe constrains a slug to kebab-case, since a slug becomes a filename.
	slugRe = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)
	// specNumRe extracts the numeric N from a spec id or a spec_id value that may
	// carry a trailing slug (spc-1, spc-2-thing).
	specNumRe = regexp.MustCompile(`^spc-([0-9]+)`)
	// specFileRe matches a spec-store filename and captures its id.
	specFileRe = regexp.MustCompile(`^(spc-[0-9]+)-.*\.md$`)
	// intentFileRe matches an intent filename.
	intentFileRe = regexp.MustCompile(`^itd-[0-9]+.*\.md$`)
)

// Spec is one spec record. Status is the bucket it was found in; Path is
// repo-relative (never an absolute local path).
type Spec struct {
	ID     string `json:"id"`     // spc-N
	Slug   string `json:"slug"`   // kebab-case
	Intent string `json:"intent"` // itd-N, the load-bearing link
	Status string `json:"status"` // open | closed (directory-as-truth)
	Path   string `json:"path"`   // repo-relative markdown path
}

// Store is the in-memory set of spec records discovered under both buckets.
type Store struct {
	Specs []Spec `json:"specs"`
}

// Lookup returns the spec with the given id; ok is false when absent.
func (s Store) Lookup(specID string) (Spec, bool) {
	for _, sp := range s.Specs {
		if sp.ID == specID {
			return sp, true
		}
	}
	return Spec{}, false
}

// ByIntent returns the spec linked to the given intent id; ok is false when no
// spec realises that intent.
func (s Store) ByIntent(intentID string) (Spec, bool) {
	for _, sp := range s.Specs {
		if sp.Intent == intentID {
			return sp, true
		}
	}
	return Spec{}, false
}

// Validate enforces the id regexes and that intent is a well-formed itd-N. It
// is the fail-closed guard both Load and the minting path run before trusting a
// record's id in a filesystem path.
func Validate(s Spec) error {
	if !specIDRe.MatchString(s.ID) {
		return fmt.Errorf("spec: id %q must match ^spc-[0-9]+$", s.ID)
	}
	if !intentIDRe.MatchString(s.Intent) {
		return fmt.Errorf("spec %s: intent %q must match ^itd-[0-9]+$", s.ID, s.Intent)
	}
	return nil
}

// specNum extracts the numeric N from a spec id or spec_id value, or 0 if none.
func specNum(id string) int {
	m := specNumRe.FindStringSubmatch(id)
	if m == nil {
		return 0
	}
	n, err := strconv.Atoi(m[1])
	if err != nil {
		// An over-int64 (or otherwise unparseable) number is not a real
		// reservation: Atoi returns the clamped MaxInt64 alongside the error, and
		// keeping it would make NextID compute max+1 and wrap to a NEGATIVE id
		// (spc--9223…). Treat it as no number so the id space stays sane.
		return 0
	}
	return n
}

// renderSpec is the minimal spec-file body: frontmatter carrying the id, slug,
// and the load-bearing intent link, plus a title and a Summary placeholder.
func renderSpec(id, slug, intentID string) string {
	var b strings.Builder
	b.WriteString("---\n")
	fmt.Fprintf(&b, "id: %s\n", id)
	fmt.Fprintf(&b, "slug: %s\n", slug)
	fmt.Fprintf(&b, "intent: %s\n", intentID)
	b.WriteString("---\n")
	fmt.Fprintf(&b, "# %s\n\n", slug)
	b.WriteString("## Summary\n\n")
	// A clear author-guidance placeholder, not a bare "TODO" that reads as drift:
	// the spec body is the design record the intent's fidelity review audits against.
	fmt.Fprintf(&b, "_Draft: describe what %s delivers for %s — scope, approach, and how "+
		"it satisfies the intent's Acceptance Criteria. This spec is the design record "+
		"the fidelity review audits against._\n", id, intentID)
	return b.String()
}
