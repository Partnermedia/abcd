package changelog

import (
	"fmt"
	"path"
	"regexp"
	"sort"
	"strings"

	"github.com/REPPL/abcd-cli/internal/core/frontmatter"
	"github.com/REPPL/abcd-cli/internal/gitutil"
)

// The two record families a release is cut from. Directory-as-truth is the
// lifecycle authority in this repo — an intent's folder IS its status — so
// "what shipped" is exactly "which record files live in the terminal folders",
// and no separate release ledger is needed.
const (
	intentsShippedDir = ".abcd/development/intents/shipped"
	issuesResolvedDir = ".abcd/work/issues/resolved"
)

// recordPaths are the pathspecs handed to git ls-tree, in the order the sets are
// reported.
var recordPaths = []string{intentsShippedDir, issuesResolvedDir}

// recordFileRe matches a record filename and captures its id. The record
// families are the only files that count: a directory README or a stray note
// living beside them is not a release line. internal/core/lint holds equivalent
// unexported filename matchers for its own scanners (intentFileRe/issueFileRe);
// if a third consumer appears, that pair and this one want consolidating into a
// single exported matcher rather than a third copy.
var recordFileRe = regexp.MustCompile(`^((?:itd|iss)-\d+)[^/]*\.md$`)

// Record is one record in a release cut: where it lives, what it is called, and
// the one product judgement it declares.
type Record struct {
	// Path is the repo-relative, slash-separated path of the record file.
	Path string
	// ID is the record id derived from the FILENAME (itd-73, iss-24), not from
	// frontmatter: the filename is what the id-uniqueness lint already governs,
	// and issue frontmatter quotes its id, so the filename is the cheaper and
	// more consistent source.
	ID string
	// Impact is the parsed judgement, or "" when the record carries none.
	Impact Impact
	// ImpactErr is why the impact could not be parsed, empty when it parsed. A
	// record is reported as unlabelled rather than defaulted, because defaulting
	// a missing judgement silently under-bumps the release.
	ImpactErr string
	// Title names the record: its `# ` heading, else its frontmatter slug, else
	// its id. It is never empty (see summarise).
	Title string
	// Summary is the record's opening paragraph, collapsed and capped. Title and
	// Summary are the SOURCE MATERIAL a changelog composer writes prose from;
	// they are carried on the record because the blob is read once here, and a
	// later reader could disagree with this one about what a record says.
	Summary string
}

// RecordSet is a release cut: the set-difference of record END STATES between
// the anchor tag and HEAD.
//
// It is deliberately NOT a git-log walk of moves. A squash or rebase merge
// collapses the commit that moved a record into its terminal folder, so a walk
// over `<tag>..HEAD -- <paths>` reports a different set depending on how a
// branch happened to land — the same tree, two answers. Comparing end states
// asks the only question that matters ("which records are in the terminal
// folders now, and which were there at the tag?") and is immune to the shape of
// the history in between.
type RecordSet struct {
	// BaseRef is the anchor the cut is measured from (a tag such as v0.3.0).
	BaseRef string
	// Added are records present at HEAD and absent at the anchor — the cut.
	Added []Record
	// Removed are records present at the anchor and absent at HEAD. A record
	// leaves a terminal folder when it is superseded or re-slugged; either way
	// it is a user-visible change that must surface as a Removed/Changed line
	// rather than silently vanish from the release.
	Removed []Record
}

// All returns the whole cut — added then removed — for callers that must account
// for every record, such as the changelog completeness bijection.
func (s RecordSet) All() []Record {
	out := make([]Record, 0, len(s.Added)+len(s.Removed))
	out = append(out, s.Added...)
	out = append(out, s.Removed...)
	return out
}

// ChangelogRequired is the cut minus every internal record: exactly the set of
// records the generated prose must cite, no more and no less. internal records
// are excluded here (not filtered later, per-caller) so the changelog and the
// version agree on one definition of "user-facing".
func (s RecordSet) ChangelogRequired() []Record {
	out := make([]Record, 0, len(s.Added)+len(s.Removed))
	for _, r := range s.All() {
		if r.InChangelog() {
			out = append(out, r)
		}
	}
	return out
}

// InChangelog reports whether this record must be cited in the generated prose.
// It is the one definition of "must be reported", written here rather than at
// each call site so the preview, the composer's view of the cut, and the
// completeness bijection can never disagree about which records are required.
//
// It is NOT simply Impact.InChangelog(). An UNLABELLED record is required too,
// and that is the whole reason this method exists: an unlabelled record's
// judgement is unknown, not internal, and the only unlabelled records that ever
// reach a ready cut are on the Removed side (UnlabelledAdded refuses the other
// direction), read from the anchor tag's IMMUTABLE tree. Folding one into "earns
// no line" would drop a supersession from the permanent release record for a
// defect the operator cannot fix — the silent omission UnlabelledAdded's contract
// promises never happens.
func (r Record) InChangelog() bool { return r.Impact.InChangelog() || r.ImpactErr != "" }

// Unlabelled returns every record in the cut whose impact is absent or invalid,
// for a preview that must show the operator the whole picture. It is NOT the
// refusal set — see UnlabelledAdded for that.
func (s RecordSet) Unlabelled() []Record { return unlabelled(s.All()) }

// UnlabelledAdded returns the unlabelled records on the ADDED side only: the
// ones a caller may refuse the cut over, because an unlabelled record ranks
// below every real impact and deriving over it would quietly under-bump a
// release that may contain a break.
//
// Removed records are deliberately excluded. Their blob is read from the anchor
// tag's immutable tree, so one that predates the impact back-fill can never be
// labelled: at HEAD the file is either gone (supersession) or already carries a
// valid impact under its new name (re-slug). Refusing over it would name a
// remedy the operator cannot perform and would block every release until the
// move was reverted. Such a record still travels in Removed, so the release
// still reports it rather than dropping it silently.
func (s RecordSet) UnlabelledAdded() []Record { return unlabelled(s.Added) }

// unlabelled is the one filter both accessors share, so the definition of
// "unlabelled" is written once.
func unlabelled(records []Record) []Record {
	var out []Record
	for _, r := range records {
		if r.ImpactErr != "" {
			out = append(out, r)
		}
	}
	return out
}

// Impact is the strongest impact in the cut — the one that decides the bump.
// An empty or all-internal cut yields ImpactInternal, which reads at the call
// site as "nothing to release".
func (s RecordSet) Impact() Impact { return maxImpactOf(s.All()) }

// DeclaresBreak reports whether the cut declares that it narrows the public
// surface on purpose — the release guardrail's "was this break declared?" test.
//
// It judges the ADDED side only, and that restriction is the whole point of the
// method existing separately from Impact(). A Removed record's blob is read from
// the anchor tag's immutable tree, so its label is what the LAST release
// declared, not what this one does. Supersession is a supported flow — a shipped
// intent leaves the folder when a later one replaces it — and the historical
// impact back-fill puts `breaking` labels on old intents, so a cut that ships
// nothing breaking can easily carry a superseded `breaking` record on its Removed
// side. Counting that as a declaration would wave an undeclared removal through
// the guardrail, silently, which is the exact fail-open the guardrail exists to
// prevent. The same reasoning excludes Removed from UnlabelledAdded.
//
// A withdrawal that genuinely narrows the surface therefore ships its own
// superseding record labelled `breaking`.
//
// Impact() deliberately keeps counting both sides: the version must account for
// every record in the cut, because a record leaving a terminal folder is itself a
// user-visible change.
func (s RecordSet) DeclaresBreak() bool { return maxImpactOf(s.Added) == ImpactBreaking }

// maxImpactOf is the one place a slice of records is reduced to its strongest
// impact, so the version arithmetic and the guardrail can never disagree about
// what the maximum means.
func maxImpactOf(records []Record) Impact {
	impacts := make([]Impact, 0, len(records))
	for _, r := range records {
		impacts = append(impacts, r.Impact)
	}
	return MaxImpact(impacts)
}

// ShippedSince computes the release cut between baseRef (the anchor tag) and
// HEAD, as the set-difference of the record files in the terminal folders.
//
// Both sides are read out of git, never out of the working tree: a removed
// record exists only in the anchor's tree, and reading the added side from git
// too keeps a dirty or half-staged working tree from changing what a release
// reports. Every returned slice is sorted by path — the set feeds a rendered
// preview and a changelog bijection, both of which must be reproducible.
//
// A baseRef git cannot resolve is an error, never an empty set: reporting
// "nothing shipped" against a tag that does not exist would let a cut silently
// claim there is nothing to release.
func ShippedSince(root string, baseRef string) (RecordSet, error) {
	set := RecordSet{BaseRef: baseRef}

	basePaths, err := recordPathsAt(root, baseRef)
	if err != nil {
		return RecordSet{}, err
	}
	headPaths, err := recordPathsAt(root, "HEAD")
	if err != nil {
		return RecordSet{}, err
	}

	for p := range headPaths {
		if _, atBase := basePaths[p]; !atBase {
			set.Added = append(set.Added, newRecord(root, "HEAD", p))
		}
	}
	for p := range basePaths {
		if _, atHead := headPaths[p]; !atHead {
			set.Removed = append(set.Removed, newRecord(root, baseRef, p))
		}
	}
	sortRecords(set.Added)
	sortRecords(set.Removed)
	return set, nil
}

// recordPathsAt lists the record files present in the terminal folders at ref,
// keyed by repo-relative path. Non-record files (READMEs, notes) are dropped
// here, so every later stage sees records only.
func recordPathsAt(root string, ref string) (map[string]struct{}, error) {
	args := append([]string{"ls-tree", "-r", "-z", "--name-only", ref, "--"}, recordPaths...)
	// -z makes git emit raw NUL-separated paths, so a path containing a quote,
	// a backslash, or a newline cannot be mangled by git's default path quoting
	// or desync the split.
	out, err := gitutil.Run(root, args...)
	if err != nil {
		return nil, fmt.Errorf("listing records at %s: %w", ref, err)
	}
	paths := map[string]struct{}{}
	for _, p := range strings.Split(out, "\x00") {
		if p == "" {
			continue
		}
		if !recordFileRe.MatchString(path.Base(p)) {
			continue
		}
		paths[p] = struct{}{}
	}
	return paths, nil
}

// maxRecordBytes caps the guarded blob read, in the same order as
// maxChangelogBytes and for the same reason: a record is a page of prose, so a
// file that is not one must not stream unbounded input into a read-only preview.
// Only the frontmatter is read, and that is at the top of the blob, so a
// truncated tail cannot change the parsed impact.
const maxRecordBytes = 4 << 20

// newRecord reads one record's blob at ref and parses its impact. A blob that
// cannot be read, or an impact that cannot be parsed, yields an unlabelled
// record carrying the reason rather than an error: one malformed record must not
// hide the rest of the cut from the operator who has to fix it.
func newRecord(root string, ref string, relPath string) Record {
	rec := Record{Path: relPath, ID: recordID(relPath)}
	blob, err := gitutil.RunLimited(root, maxRecordBytes, "cat-file", "blob", ref+":"+relPath)
	if err != nil {
		rec.ImpactErr = fmt.Sprintf("reading %s at %s: %v", relPath, ref, err)
		rec.Title = rec.ID
		return rec
	}
	// The source material is extracted even when the impact is unlabelled: the
	// operator who has to fix that record is helped by seeing what it says.
	rec.Title, rec.Summary = summarise(blob, rec.ID)
	field := frontmatter.Fields(strings.Split(blob, "\n"))["impact"]
	impact, err := ParseImpact(field.Value)
	if err != nil {
		rec.ImpactErr = err.Error()
		return rec
	}
	rec.Impact = impact
	return rec
}

// recordID extracts the id from a record path; the caller has already matched
// the filename, so the empty fallback is unreachable in practice.
func recordID(relPath string) string {
	m := recordFileRe.FindStringSubmatch(path.Base(relPath))
	if m == nil {
		return ""
	}
	return m[1]
}

// sortRecords orders by path, the only field guaranteed unique in a set.
func sortRecords(records []Record) {
	sort.Slice(records, func(i, j int) bool { return records[i].Path < records[j].Path })
}
