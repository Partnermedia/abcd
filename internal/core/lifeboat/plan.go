package lifeboat

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"path"
	"sort"
	"strings"

	"github.com/REPPL/abcd-cli/internal/core/ahoy"
)

// PlannedFile is one file the packer would write into a lifeboat, produced
// without touching disk. Path is destination-relative, POSIX-separated, cleaned,
// with no leading slash and no ".." — it is safe to join under a destination
// root. Content is the exact bytes that would be written.
type PlannedFile struct {
	Path    string `json:"path"`
	Content []byte `json:"-"`
	// Bytes is the content length, carried so a dry-run render (and the JSON
	// form of a plan) can report size without echoing the content.
	Bytes int `json:"bytes"`
}

// Omission records a file that belonged in the lifeboat but was left out, and
// why. A lifeboat that silently drops a record is worse than one that declares
// the gap: the coverage experiment is about honesty, so every skipped record
// (too large to read, unreadable, or dropped because the plan hit its size cap)
// is named here rather than vanishing.
type Omission struct {
	Path   string `json:"path"`
	Reason string `json:"reason"`
}

// Plan-wide safety ceilings. The per-file (maxProbeReadBytes) and per-directory
// (maxDirEntries) caps bound one read; these bound the whole plan, so probing a
// hostile or pathological tree cannot exhaust the operator's memory by
// multiplying many bounded reads. A file that would breach a ceiling is omitted
// and recorded, never read.
const (
	maxPlanFiles      = 20000
	maxPlanTotalBytes = 512 << 20 // 512 MiB
)

// Lifeboat is the complete plan for one disembark: every file that would be
// written, the coverage it was derived from, and any records deliberately left
// out. Pack (M3b) is Plan plus a write over exactly this file set — one code
// path, so a dry-run cannot lie about what a real pack will do. Nothing here
// touches the destination.
type Lifeboat struct {
	Coverage  Coverage      `json:"coverage"`
	Files     []PlannedFile `json:"files"`
	Omissions []Omission    `json:"omissions,omitempty"`
}

// ProvenanceName is the lifeboat's commit marker and the key to embark's
// destination-safety gate. It is written last and hashes every other file.
const ProvenanceName = "_provenance.json"

// Provenance is the lifeboat's manifest header: what produced it, from what
// source, the pinned hash over every other file, and any records left out. It
// deliberately carries no timestamp, so a re-plan of an unchanged source is
// byte-identical and the hash is stable.
type Provenance struct {
	SchemaVersion  int        `json:"schema_version"`
	Generator      string     `json:"generator"`
	SourceName     string     `json:"source_name"`
	SourceRootSHA  string     `json:"source_root_sha,omitempty"`
	TiersPresent   []Tier     `json:"tiers_present"`
	ManifestSHA256 string     `json:"manifest_sha256"`
	Omissions      []Omission `json:"omissions,omitempty"`
}

// planBuilder assembles a lifeboat's file set with three invariants the review
// pass demanded: no two files share a destination path (a real pack writes one
// file per path, so a plan that lists two would over-describe the pack), the
// whole plan is bounded (maxPlanFiles / maxPlanTotalBytes), and any record left
// out is recorded rather than silently dropped.
type planBuilder struct {
	files     []PlannedFile
	seen      map[string]bool
	bytes     int
	omissions []Omission
	full      bool
}

func newPlanBuilder() *planBuilder {
	return &planBuilder{seen: map[string]bool{}}
}

// add places content at dest, unless dest is already taken (first writer wins,
// deterministically — callers add in a fixed order) or the plan has hit a size
// ceiling. A rejected add for a ceiling is recorded as an omission.
func (pb *planBuilder) add(dest string, content []byte) {
	if pb.seen[dest] {
		return
	}
	if pb.full || len(pb.files) >= maxPlanFiles || pb.bytes+len(content) > maxPlanTotalBytes {
		pb.full = true
		pb.omissions = append(pb.omissions, Omission{Path: dest, Reason: "plan size ceiling reached"})
		return
	}
	pb.seen[dest] = true
	pb.bytes += len(content)
	pb.files = append(pb.files, PlannedFile{Path: dest, Content: content, Bytes: len(content)})
}

// copyRecord reads src and adds it verbatim at dest. If src matched a record
// filter but cannot be read — it exceeds the per-file cap, or is otherwise
// unreadable — the omission is recorded, so a large or hostile record is
// declared, not silently lost.
func (pb *planBuilder) copyRecord(ctx *SourceContext, src, dest string) {
	if data, ok := ctx.ReadFile(src); ok {
		// Neutralise any abcd marker block before it travels: a verbatim record
		// carrying a live BEGIN…END fence would plant a stale rules-loader in
		// whatever repo later embarks the lifeboat. Stripping here (inside Plan)
		// keeps the manifest hash over the bytes a pack actually writes.
		if stripped, changed := ahoy.StripMarkerBlock(data); changed {
			data = stripped
		}
		pb.add(dest, data)
		return
	}
	pb.omissions = append(pb.omissions, Omission{Path: src, Reason: "unreadable or exceeds per-file size cap"})
}

// Plan produces the complete lifeboat plan for a repository, read-only. It runs
// the probe, then assembles: the grounded and partial brief section files (each
// citing its source), the coverage report (JSON and Markdown), verbatim copies
// of the ADRs and the issue ledger, the rescue spine, and _provenance.json with
// the pinned manifest hash. It writes nothing — Plan has no destination.
func Plan(repoRoot string) (Lifeboat, error) {
	cov, err := Probe(repoRoot)
	if err != nil {
		return Lifeboat{}, err
	}
	ctx, err := newSourceContext(repoRoot)
	if err != nil {
		return Lifeboat{}, err
	}
	defer ctx.Close()

	pb := newPlanBuilder()

	// 1. Brief: only grounded and partial sections, each a citation map back to
	//    its evidence. Composing prose from those sources is a later step (M6);
	//    the brief here is the honest map, not synthesised text. Sections whose
	//    mapping home is outside brief/ — graveyard, docs/adrs, activity/issues,
	//    rescue/spine — are materialised at top level by the dedicated copy steps
	//    below, not as brief stubs.
	for _, s := range cov.Sections {
		if s.Status == StatusBlank {
			continue
		}
		leaf, ok := briefLeaf(s.Name)
		if !ok {
			continue
		}
		pb.add(path.Join("brief", leaf), briefSectionDoc(s))
	}

	// 2. Coverage — first-class, both machine and human forms.
	if j, err := json.MarshalIndent(cov, "", "  "); err == nil {
		pb.add("coverage.json", append(j, '\n'))
	}
	pb.add("coverage.md", []byte(cov.Render()))

	// 3. ADRs, verbatim, wherever they were found. Two source homes may hold the
	//    same basename (a migrated ADR left in both docs/adr and docs/adrs); the
	//    first in sorted-source order wins the dest path, the rest are dropped by
	//    the builder — the plan never lists one dest twice.
	for _, src := range planADRSources(ctx) {
		leaf := safeLeaf(path.Base(src))
		if leaf == "" {
			continue
		}
		pb.copyRecord(ctx, src, path.Join("docs", "adrs", leaf))
	}

	// 4. The issue ledger, verbatim, under its state subdirectories.
	for _, state := range nativeIssueStates {
		dir := path.Join(nativeIssuesDir, state)
		for _, name := range ctx.ListDir(dir) {
			if !strings.HasPrefix(name, "iss-") || !strings.HasSuffix(name, ".md") {
				continue
			}
			leaf := safeLeaf(name)
			if leaf == "" {
				continue
			}
			pb.copyRecord(ctx, path.Join(dir, name), path.Join("activity", "issues", state, leaf))
		}
	}

	// 5. The rescue spine: the intent corpus where one exists, a git-derived
	//    summary where it does not.
	planRescueSpine(ctx, pb)

	files := pb.files

	// 6. Provenance, assembled last, hashing every other file. It is appended
	//    outside the builder's ceiling so it is always present.
	prov := Provenance{
		SchemaVersion:  SchemaVersion,
		Generator:      "abcd disembark",
		SourceName:     cov.Repo.Name,
		SourceRootSHA:  cov.Repo.RootSHA,
		TiersPresent:   cov.TiersPresent,
		ManifestSHA256: ManifestSHA256(files),
		Omissions:      pb.omissions,
	}
	pj, err := json.MarshalIndent(prov, "", "  ")
	if err != nil {
		return Lifeboat{}, err
	}
	files = append(files, PlannedFile{Path: ProvenanceName, Content: append(pj, '\n'), Bytes: len(pj) + 1})

	sort.Slice(files, func(i, j int) bool { return files[i].Path < files[j].Path })
	return Lifeboat{Coverage: cov, Files: files, Omissions: pb.omissions}, nil
}

// ManifestSHA256 is the pinned lifeboat hash (adr-35): SHA-256 over the
// concatenation of "<sha256>  <path>\n" for every file EXCEPT _provenance.json
// (which cannot hash itself), sorted lexicographically BY PATH — not by the
// assembled line, whose leading hash would otherwise dominate the ordering. It
// is deterministic for a given file set.
func ManifestSHA256(files []PlannedFile) string {
	type entry struct {
		path string
		line string
	}
	entries := make([]entry, 0, len(files))
	for _, f := range files {
		if f.Path == ProvenanceName {
			continue
		}
		sum := sha256.Sum256(f.Content)
		entries = append(entries, entry{f.Path, fmt.Sprintf("%x  %s\n", sum, f.Path)})
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].path < entries[j].path })
	var buf strings.Builder
	for _, e := range entries {
		buf.WriteString(e.line)
	}
	h := sha256.Sum256([]byte(buf.String()))
	return fmt.Sprintf("%x", h)
}

// PlanManifest is the dry-run view of a Lifeboat: what `disembark plan` would
// write, by path and size, plus the pinned manifest hash and any omitted
// records — never the content. It is what a human (or a machine) reads to see
// the shape of a pack before any pack happens.
type PlanManifest struct {
	SchemaVersion  int           `json:"schema_version"`
	SourceName     string        `json:"source_name"`
	ManifestSHA256 string        `json:"manifest_sha256"`
	FileCount      int           `json:"file_count"`
	TotalBytes     int           `json:"total_bytes"`
	Files          []PlannedFile `json:"files"`
	Omissions      []Omission    `json:"omissions,omitempty"`
}

// Manifest reduces a Lifeboat to its dry-run manifest: paths and sizes, the
// pinned hash, the totals, and any omissions — the content is deliberately
// dropped (PlannedFile marshals path and byte count only).
func (lb Lifeboat) Manifest() PlanManifest {
	total := 0
	for _, f := range lb.Files {
		total += f.Bytes
	}
	return PlanManifest{
		SchemaVersion:  SchemaVersion,
		SourceName:     lb.Coverage.Repo.Name,
		ManifestSHA256: ManifestSHA256(lb.Files),
		FileCount:      len(lb.Files),
		TotalBytes:     total,
		Files:          lb.Files,
		Omissions:      lb.Omissions,
	}
}

// RenderManifest returns the human-readable dry-run: the file list with sizes,
// then the totals, any omissions, and the pinned hash. It writes nothing — it
// only describes what a pack would write.
func (lb Lifeboat) RenderManifest() string {
	m := lb.Manifest()
	var b strings.Builder
	fmt.Fprintf(&b, "lifeboat plan for %s (dry run — nothing written)\n\n", sanitize(m.SourceName))
	for _, f := range m.Files {
		fmt.Fprintf(&b, "  %8d  %s\n", f.Bytes, f.Path)
	}
	fmt.Fprintf(&b, "\n%d files · %d bytes\n", m.FileCount, m.TotalBytes)
	if len(m.Omissions) > 0 {
		fmt.Fprintf(&b, "\n%d record(s) omitted:\n", len(m.Omissions))
		for _, o := range m.Omissions {
			fmt.Fprintf(&b, "  - %s (%s)\n", sanitize(o.Path), sanitize(o.Reason))
		}
	}
	fmt.Fprintf(&b, "\nmanifest sha256: %s\n", m.ManifestSHA256)
	return b.String()
}

// briefLeaf resolves a section's mapping LifeboatPath to the brief file the plan
// writes, relative to the brief/ root. It reports false for a section whose home
// is outside brief/ (graveyard, docs/adrs, activity/issues, rescue) — those are
// materialised at top level, not as brief stubs. A directory path (trailing "/")
// becomes its README.md.
func briefLeaf(section Section) (string, bool) {
	lp := lifeboatPathFor(section)
	if !strings.HasPrefix(lp, "brief/") {
		return "", false
	}
	lp = strings.TrimPrefix(lp, "brief/")
	if strings.HasSuffix(lp, "/") {
		return lp + "README.md", true
	}
	return lp, true
}

// lifeboatPathFor returns the mapping LifeboatPath for a section, or the section
// name as a fallback (every section in a coverage report has a Table row).
func lifeboatPathFor(section Section) string {
	for _, m := range Table {
		if m.Section == section {
			return m.LifeboatPath
		}
	}
	return string(section)
}

// briefSectionDoc renders one grounded/partial brief section as a citation map:
// its status, the tier and confidence, and the evidence it was grounded from.
// It is deterministic and cites every source — never synthesised prose.
func briefSectionDoc(s SectionCoverage) []byte {
	var b strings.Builder
	fmt.Fprintf(&b, "# %s\n\n", s.Name)
	fmt.Fprintf(&b, "Status: %s", s.Status)
	if s.Confidence != "" {
		fmt.Fprintf(&b, " (%s, %s confidence)", s.Tier, s.Confidence)
	}
	b.WriteString("\n\n")
	if len(s.Evidence) > 0 {
		b.WriteString("Grounded from:\n\n")
		ev := append([]string(nil), s.Evidence...)
		sort.Strings(ev)
		for _, e := range ev {
			fmt.Fprintf(&b, "- %s\n", sanitize(e))
		}
		b.WriteString("\n")
	}
	b.WriteString("> Citation map produced by `abcd disembark`. The cited sources are the\n")
	b.WriteString("> evidence; composing them into prose is a later step.\n")
	return []byte(b.String())
}

// planADRSources returns the repo-relative ADR files to copy verbatim, from the
// abcd-native decisions directory and every conventional ADR home, unique and
// sorted. It reuses the ADR-location knowledge the probe's adapters already hold.
func planADRSources(ctx *SourceContext) []string {
	var out []string
	dirs := append([]string{nativeADRDir}, convADRDirs...)
	for _, dir := range dirs {
		if !ctx.IsDir(dir) {
			continue
		}
		for _, name := range ctx.ListDir(dir) {
			low := strings.ToLower(name)
			if strings.HasSuffix(low, ".md") || strings.HasSuffix(low, ".markdown") {
				out = append(out, path.Join(dir, name))
			}
		}
	}
	return dedupeSorted(out)
}

// planRescueSpine adds the spine files to the builder: the intent corpus copied
// verbatim where one exists, otherwise a single git-derived spine summary. A
// repo with neither yields no spine file.
func planRescueSpine(ctx *SourceContext, pb *planBuilder) {
	intents := 0
	for _, sub := range append([]string{""}, ctx.ListDir(nativeIntentsDir)...) {
		dir := nativeIntentsDir
		safeSub := ""
		if sub != "" {
			// A rejected subdirectory name must DROP its files, not relocate them
			// up a level — path.Join would silently swallow an empty segment, and
			// that is exactly the destination steering safeLeaf exists to prevent.
			safeSub = safeLeaf(sub)
			if safeSub == "" {
				continue
			}
			dir = path.Join(nativeIntentsDir, sub)
		}
		if !ctx.IsDir(dir) {
			continue
		}
		for _, name := range ctx.ListDir(dir) {
			if !strings.HasPrefix(name, "itd-") || !strings.HasSuffix(name, ".md") {
				continue
			}
			leaf := safeLeaf(name)
			if leaf == "" {
				continue
			}
			dest := path.Join("rescue", "intents", leaf)
			if safeSub != "" {
				dest = path.Join("rescue", "intents", safeSub, leaf)
			}
			pb.copyRecord(ctx, path.Join(dir, name), dest)
			intents++
		}
	}
	if intents > 0 {
		return
	}
	// No intent corpus: derive a spine from git history alone.
	if !ctx.IsGit() {
		return
	}
	n := ctx.CommitCount()
	if n == 0 {
		return
	}
	authors := map[string]bool{}
	for _, a := range ctx.GitLines("log", "--format=%an") {
		authors[a] = true
	}
	var b strings.Builder
	b.WriteString("# Rescue spine (git-derived)\n\n")
	b.WriteString("No written intent corpus was found; this spine is reconstructed from\n")
	b.WriteString("git history alone.\n\n")
	fmt.Fprintf(&b, "- commits: %d\n", n)
	fmt.Fprintf(&b, "- contributors: %d\n", len(authors))
	pb.add(path.Join("rescue", "spine.md"), []byte(b.String()))
}

// safeLeaf returns name unchanged if it is already a safe single path
// component, or "" if it is not. It REJECTS rather than normalises: a name that
// is empty, ".", "..", contains a path separator, carries a control character,
// or is not equal to its own base is refused outright — the plan drops that file
// rather than silently relocating a "../escape" to "escape". Destination paths
// must never be steerable by a hostile source filename.
func safeLeaf(name string) string {
	if name == "" || name == "." || name == ".." {
		return ""
	}
	if strings.ContainsAny(name, "/\\") {
		return ""
	}
	for _, r := range name {
		if r < 0x20 || r == 0x7f {
			return ""
		}
	}
	if path.Base(name) != name {
		return ""
	}
	return name
}
