package lifeboat

import (
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/REPPL/abcd-cli/internal/gitutil"
)

// SchemaVersion is the coverage-report schema version. It is stamped into every
// report and checked by the aggregate, so a future breaking change to the shape
// is detectable rather than silently misread.
const SchemaVersion = 2

// maxProbeReadBytes caps any single file the probe reads. A coverage probe reads
// prose (READMEs, ADRs, decision logs), never data blobs, so a file larger than
// this is not section evidence and is skipped. The cap also bounds memory on a
// hostile or accidental giant file.
const maxProbeReadBytes = 4 << 20 // 4 MiB

// maxGitOutputBytes caps how much stdout the probe buffers from any one git
// command. Generous enough for a large legitimate history, bounded so a hostile
// repo cannot exhaust memory through a read-only command.
const maxGitOutputBytes = 16 << 20 // 16 MiB

// maxDirEntries caps how many entries ListDir returns from one directory, so a
// directory with millions of files cannot exhaust memory when the probe indexes
// it.
const maxDirEntries = 50000

// Confidence qualifies a non-blank status: how sure the adapter is that the
// evidence it cites actually grounds the section. It is meaningless for a blank.
type Confidence string

const (
	ConfidenceHigh   Confidence = "high"
	ConfidenceMedium Confidence = "medium"
	ConfidenceLow    Confidence = "low"
)

// Evidence is what one Source reports for its section against one repository.
// The orchestrator stamps the Tier and Section from the Source itself, so an
// adapter cannot misreport which tier or section it speaks for.
//
// Contract by status:
//   - grounded / partial: Sources must be non-empty — every claim cites a file
//     or a git ref. Confidence should be set.
//   - blank: Searched should say what was looked for and Question should name
//     the thing a human must answer. A blank is a first-class result.
type Evidence struct {
	Status     Status
	Confidence Confidence
	Sources    []string // evidence cited (repo-relative paths, git refs)
	Searched   []string // what was looked for (esp. on a blank)
	Question   string   // the human question (esp. on a blank)
}

// blank is the conventional empty result for a Source that found nothing.
func blank(searched []string, question string) Evidence {
	return Evidence{Status: StatusBlank, Searched: searched, Question: question}
}

// Source is one tiered adapter: it reads a single brief section at a single
// tier and reports what it found. Probe must be side-effect-free and must never
// write to the source repository — a probe is read-only by construction.
//
// (M3 adds a Plan method to this interface so that pack is "plan plus a write"
// over the same adapters; probe needs only Probe.)
type Source interface {
	Section() Section
	Tier() Tier
	Probe(*SourceContext) Evidence
}

// allSources is the registry: every tier's adapters, concatenated. The three
// tier constructors live in sources_git.go, sources_conventions.go, and
// sources_native.go so each can be developed independently.
func allSources() []Source {
	var s []Source
	s = append(s, gitSources()...)
	s = append(s, conventionSources()...)
	s = append(s, nativeSources()...)
	return s
}

// SourceContext is the read-only material every Source probes. It is built once
// per repository and shared across all adapters, so git history is queried and
// files are read through a single contained, cached surface. Every read is
// contained to the repository root via os.Root (no symlinked component can
// redirect a read outside the repo) and bounded in size, so probing a hostile
// or archived tree cannot escape it, hang on a FIFO, or exhaust memory.
type SourceContext struct {
	RepoRoot string

	root    *os.Root // containment scope for every file read; nil if unopenable
	isGit   bool
	rootSHA string

	mu       sync.Mutex
	gitCache map[string]gitResult
}

type gitResult struct {
	out string
	err error
}

// newSourceContext opens repoRoot for contained reads and records whether it is
// a git repository. It never writes.
func newSourceContext(repoRoot string) (*SourceContext, error) {
	abs, err := filepath.Abs(repoRoot)
	if err != nil {
		return nil, err
	}
	c := &SourceContext{RepoRoot: abs, gitCache: map[string]gitResult{}}
	// os.OpenRoot refuses any later path component that escapes the root,
	// symlinked intermediates included — the same containment the privacy audit
	// adopted. A root that cannot be opened leaves reads returning "absent".
	if root, err := os.OpenRoot(abs); err == nil {
		c.root = root
	}
	if gitutil.InRepo(abs) {
		c.isGit = true
		c.rootSHA = firstRootSHA(abs)
	}
	return c, nil
}

// Close releases the containment handle.
func (c *SourceContext) Close() {
	if c.root != nil {
		_ = c.root.Close()
	}
}

// IsGit reports whether the source is a git working tree.
func (c *SourceContext) IsGit() bool { return c.isGit }

// RootSHA is the canonical root-commit SHA, or "" outside a git repo.
func (c *SourceContext) RootSHA() string { return c.rootSHA }

// Git runs a read-only git subcommand under the repo, isolated from the
// developer's git config, and caches the result so repeated identical queries
// across adapters cost one exec. Outside a git repo it returns an error.
func (c *SourceContext) Git(args ...string) (string, error) {
	key := strings.Join(args, "\x00")
	c.mu.Lock()
	if r, ok := c.gitCache[key]; ok {
		c.mu.Unlock()
		return r.out, r.err
	}
	c.mu.Unlock()

	// Cap git stdout: an untrusted repo can make a read-only command emit
	// arbitrarily much, and the probe must not let that grow memory without
	// bound.
	out, err := gitutil.RunLimited(c.RepoRoot, maxGitOutputBytes, args...)

	c.mu.Lock()
	c.gitCache[key] = gitResult{out: out, err: err}
	c.mu.Unlock()
	return out, err
}

// GitLines runs a git subcommand and splits stdout into non-empty lines.
func (c *SourceContext) GitLines(args ...string) []string {
	out, err := c.Git(args...)
	if err != nil || out == "" {
		return nil
	}
	return splitLines(out)
}

// CommitCount is the number of commits reachable from HEAD, or 0 outside a repo.
func (c *SourceContext) CommitCount() int {
	out, err := c.Git("rev-list", "--count", "HEAD")
	if err != nil {
		return 0
	}
	n, _ := strconv.Atoi(strings.TrimSpace(out))
	return n
}

// ReadFile reads a repo-relative file through the containment root, bounded and
// non-blocking. It returns (data, true) only for a regular file within the cap;
// a missing file, a directory, a FIFO/device, an escaping path, or an oversized
// file returns (nil, false) — never an error and never a blocked read.
func (c *SourceContext) ReadFile(rel string) ([]byte, bool) {
	if c.root == nil {
		return nil, false
	}
	f, err := c.root.OpenFile(filepath.FromSlash(rel), os.O_RDONLY|nonBlock, 0)
	if err != nil {
		return nil, false
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil || !info.Mode().IsRegular() {
		return nil, false
	}
	if info.Size() > maxProbeReadBytes {
		return nil, false
	}
	data, err := io.ReadAll(io.LimitReader(f, maxProbeReadBytes))
	if err != nil {
		return nil, false
	}
	return data, true
}

// Exists reports whether a repo-relative path exists (of any kind) within the
// containment root.
func (c *SourceContext) Exists(rel string) bool {
	if c.root == nil {
		return false
	}
	_, err := c.root.Stat(filepath.FromSlash(rel))
	return err == nil
}

// IsDir reports whether a repo-relative path is a directory within the root.
func (c *SourceContext) IsDir(rel string) bool {
	if c.root == nil {
		return false
	}
	info, err := c.root.Stat(filepath.FromSlash(rel))
	return err == nil && info.IsDir()
}

// FindFirst returns the first candidate that exists (case-sensitive, as given),
// or "" if none do. Adapters pass the conventional spellings they care about
// (e.g. "README.md", "README", "readme.md").
func (c *SourceContext) FindFirst(candidates ...string) string {
	for _, cand := range candidates {
		if c.Exists(cand) {
			return cand
		}
	}
	return ""
}

// ListDir returns the names (not paths) of entries directly under a repo-relative
// directory, sorted. It never recurses and never escapes the root.
func (c *SourceContext) ListDir(rel string) []string {
	if c.root == nil {
		return nil
	}
	f, err := c.root.Open(filepath.FromSlash(rel))
	if err != nil {
		return nil
	}
	defer f.Close()
	// Bounded: a directory with millions of entries cannot balloon memory here.
	entries, err := f.ReadDir(maxDirEntries)
	if err != nil && len(entries) == 0 {
		return nil
	}
	names := make([]string, 0, len(entries))
	for _, e := range entries {
		names = append(names, e.Name())
	}
	sort.Strings(names)
	return names
}

// firstRootSHA returns the canonical (first) root-commit SHA, or "".
func firstRootSHA(repoRoot string) string {
	out, err := gitutil.Run(repoRoot, "rev-list", "--max-parents=0", "HEAD")
	if err != nil {
		return ""
	}
	fields := strings.Fields(out)
	if len(fields) == 0 {
		return ""
	}
	return fields[0]
}

// Probe runs every registered adapter over one repository and reduces the
// results to a Coverage report. Adapters run concurrently; the reduction is
// deterministic. It never writes to the source repository.
func Probe(repoRoot string) (Coverage, error) {
	ctx, err := newSourceContext(repoRoot)
	if err != nil {
		return Coverage{}, err
	}
	defer ctx.Close()

	present := tiersPresent(ctx)
	presentSet := map[Tier]bool{}
	for _, t := range present {
		presentSet[t] = true
	}

	// Run adapters concurrently. An adapter whose tier is not present in this
	// repo is skipped rather than run-and-blanked, so its tier-specific
	// "searched"/"question" never colours a repo that tier is absent from.
	type result struct {
		section Section
		tier    Tier
		ev      Evidence
	}
	var (
		wg      sync.WaitGroup
		mu      sync.Mutex
		results []result
	)
	for _, s := range allSources() {
		if !presentSet[s.Tier()] {
			continue
		}
		wg.Add(1)
		go func(s Source) {
			defer wg.Done()
			ev := s.Probe(ctx)
			mu.Lock()
			results = append(results, result{section: s.Section(), tier: s.Tier(), ev: ev})
			mu.Unlock()
		}(s)
	}
	wg.Wait()

	// Index the best evidence per section: highest status wins; on a tie the
	// richer tier wins (a grounded-at-conventions beats grounded-at-git).
	best := map[Section]result{}
	for _, r := range results {
		if r.ev.Status == StatusBlank {
			continue // a blank never displaces a real result or another blank
		}
		cur, ok := best[r.section]
		if !ok || beats(r, cur) {
			best[r.section] = r
		}
	}
	// For a blank fallback, keep the richest tier's blank so its searched/
	// question is the most informative available.
	blankFallback := map[Section]result{}
	for _, r := range results {
		if r.ev.Status != StatusBlank {
			continue
		}
		cur, ok := blankFallback[r.section]
		if !ok || tierRank(r.tier) > tierRank(cur.tier) {
			blankFallback[r.section] = r
		}
	}

	// Assemble one row per brief section, in the mapping's canonical order, so
	// the report is stable and every section always appears.
	sections := make([]SectionCoverage, 0, len(Table))
	var sum Summary
	for _, m := range Table {
		sc := SectionCoverage{Name: m.Section, Status: StatusBlank}
		if r, ok := best[m.Section]; ok {
			sc = SectionCoverage{
				Name:       m.Section,
				Status:     r.ev.Status,
				Confidence: r.ev.Confidence,
				Tier:       r.tier,
				Evidence:   dedupeSorted(r.ev.Sources),
				Searched:   dedupeSorted(r.ev.Searched),
				Question:   r.ev.Question,
			}
		} else if r, ok := blankFallback[m.Section]; ok {
			sc.Searched = dedupeSorted(r.ev.Searched)
			sc.Question = r.ev.Question
		} else {
			// No adapter spoke for this section at all: derive an honest blank
			// from the hypothesis row so the report still names what a lifeboat
			// would look for and the question a human must answer.
			sc.Searched = splitReads(m.Reads)
			sc.Question = "Nothing probed grounds " + string(m.Section) + "; a human must supply it."
		}
		// Every section carries its kind (adr-36); a blank starts life open, so
		// the round-trip can track whether a human later answers or defers it.
		sc.Kind = m.Section.Kind()
		switch sc.Status {
		case StatusGrounded:
			sum.Grounded++
		case StatusPartial:
			sum.Partial++
		default:
			sc.Resolution = ResolutionOpen
			sum.Blank++
		}
		sections = append(sections, sc)
	}

	return Coverage{
		SchemaVersion: SchemaVersion,
		Repo: RepoInfo{
			Name:    filepath.Base(ctx.RepoRoot),
			RootSHA: ctx.RootSHA(),
			Commits: ctx.CommitCount(),
		},
		TiersPresent: present,
		Sections:     sections,
		Summary:      sum,
	}, nil
}

// beats reports whether candidate a is a better result than incumbent b:
// higher status, or equal status at a richer tier.
func beats(a, b struct {
	section Section
	tier    Tier
	ev      Evidence
}) bool {
	if a.ev.Status.rank() != b.ev.Status.rank() {
		return a.ev.Status.rank() > b.ev.Status.rank()
	}
	return tierRank(a.tier) > tierRank(b.tier)
}

// tiersPresent reports which tiers a repository actually has, poorest first.
func tiersPresent(c *SourceContext) []Tier {
	var present []Tier
	if c.IsGit() {
		present = append(present, TierGit)
	}
	if hasConventions(c) {
		present = append(present, TierConventions)
	}
	if hasNative(c) {
		present = append(present, TierNative)
	}
	return present
}

// hasConventions is true when any file or directory the Tier-1 convention
// adapters actually read exists — the union of their evidence sets, not just the
// headline docs. The tier gate skips every adapter of an absent tier
// (probe.go:314), so narrowing this below what the adapters consult produces
// false blanks: a repo carrying only build manifests, CI workflows, ADR dirs,
// a glossary, or an issues file would have its whole Tier-1 set skipped even
// though those adapters would find grounding evidence. Composed from the
// adapters' own name lists in sources_conventions.go so the two cannot drift.
func hasConventions(c *SourceContext) bool {
	candidates := []string{
		"docs", "LICENSE", "LICENSE.md", "CONTRIBUTING.md", "CONTRIBUTING",
		"ISSUES.md", "ISSUES",
		// Directory evidence the adapters treat as grounding.
		".github/workflows", "issues", ".github/ISSUE_TEMPLATE",
	}
	candidates = append(candidates, convReadmeNames...)      // convReadme
	candidates = append(candidates, convChangelogNames...)   // convWhatWorkedSource
	candidates = append(candidates, convGlossaryDocNames...) // convGlossarySource
	candidates = append(candidates, convPlatformFiles...)    // convPlatformSource (Dockerfile, Makefile, go.mod, package.json)
	candidates = append(candidates, convADRDirs...)          // convADRsSource
	for _, ml := range convManifestLocks {                   // convDependenciesSource
		candidates = append(candidates, ml.manifest)
	}
	return c.FindFirst(candidates...) != ""
}

// hasNative is true when the repo carries an abcd record.
func hasNative(c *SourceContext) bool {
	return c.IsDir(".abcd/development") || c.IsDir(".abcd/work")
}

// tierRank orders tiers by richness for tie-breaking.
func tierRank(t Tier) int {
	switch t {
	case TierGit:
		return 0
	case TierConventions:
		return 1
	case TierNative:
		return 2
	}
	return -1
}

// splitLines returns the non-empty, whitespace-trimmed lines of s.
func splitLines(s string) []string {
	raw := strings.Split(s, "\n")
	out := make([]string, 0, len(raw))
	for _, l := range raw {
		if t := strings.TrimSpace(l); t != "" {
			out = append(out, t)
		}
	}
	return out
}

// splitReads turns a mapping row's free-text "Reads" into discrete searched
// entries, so a blank derived from the hypothesis still lists concrete targets.
func splitReads(reads string) []string {
	parts := strings.FieldsFunc(reads, func(r rune) bool { return r == ',' || r == ';' })
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if t := strings.TrimSpace(p); t != "" {
			out = append(out, t)
		}
	}
	return out
}

// dedupeSorted returns the unique, sorted, non-empty members of in, or nil.
func dedupeSorted(in []string) []string {
	if len(in) == 0 {
		return nil
	}
	seen := map[string]bool{}
	out := make([]string, 0, len(in))
	for _, s := range in {
		s = strings.TrimSpace(s)
		if s == "" || seen[s] {
			continue
		}
		seen[s] = true
		out = append(out, s)
	}
	if len(out) == 0 {
		return nil
	}
	sort.Strings(out)
	return out
}
