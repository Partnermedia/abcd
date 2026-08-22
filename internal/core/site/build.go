package site

// `abcd site build` — the whole render, from repository to output tree.
//
// The order is: read the committed inputs (manifest, allowlist, baseline,
// package metadata), read the record ONCE through the lint engine's own scan,
// walk git history ONCE, then compose. Nothing here reaches the network, and
// nothing writes outside the output directory: every write goes through an
// os.Root opened at the destination, so a path arriving from a manifest cannot
// escape it even by construction.
//
// The build is transport-agnostic. It returns a Result describing what it wrote;
// the front door prints it.

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"github.com/Partnermedia/abcd/internal/core/changelog"
	"github.com/Partnermedia/abcd/internal/core/lint"
	"github.com/Partnermedia/abcd/internal/fsutil"
)

// DefaultOutDir is where the build writes when the caller names no directory.
const DefaultOutDir = "site"

// The committed sources the build copies into the output tree verbatim, and the
// names they take there. Cloudflare reads the destinations; the sources are
// reviewable files with comments, which the destinations keep.
var copiedSources = []struct{ src, dst string }{
	{"site-src/redirects", "_redirects"},
	{"site-src/headers", "_headers"},
	{"site-src/site.css", "site.css"},
	{"site-src/site.js", "site.js"},
	// The chart's runtime, loaded by `/record/graph/` alone. It is a separate
	// file so the landing page never pays for it.
	{"site-src/record.js", "record.js"},
}

// The marker that makes the output directory identifiable as this build's own.
//
// The build REWRITES its output tree rather than adding to it, so it has to
// empty the directory first — and emptying a directory that a person named on a
// command line is not something to do on a guess. The marker is the proof: the
// build clears a directory only when it finds its own marker there, and refuses
// a non-empty directory without one. That makes deleting somebody else's files
// structurally impossible rather than merely unlikely.
//
// The name begins with a dot so it reads as tooling metadata rather than
// content. Static-asset hosts conventionally skip dotfiles, but this build does
// not depend on that and does not assert it: the marker names no path, carries
// no secret and describes only itself, so a host that did serve it would leak
// nothing. It is excluded from the header-coverage walk as build metadata, on
// the same footing as `_headers` and `_redirects`.
const (
	siteMarkerName = ".abcd-site-build"
	siteMarkerBody = "Output of `abcd site build`. This directory is rewritten on every build;" +
		" the presence of this file is what permits that instead of a refusal.\n"
)

// outDirState is what the build found where it is about to write.
type outDirState int

const (
	outDirAbsent  outDirState = iota // nothing there yet
	outDirEmpty                      // there, holding nothing
	outDirOurs                       // there, holding a tree this build wrote
	outDirForeign                    // there, holding something else entirely
)

// inspectOutDir reports what is in the output directory, without changing it.
func inspectOutDir(outDir string) (outDirState, error) {
	entries, err := os.ReadDir(outDir)
	switch {
	case os.IsNotExist(err):
		return outDirAbsent, nil
	case err != nil:
		return 0, err
	case len(entries) == 0:
		return outDirEmpty, nil
	}
	switch _, err := os.Stat(filepath.Join(outDir, siteMarkerName)); {
	case err == nil:
		return outDirOurs, nil
	case os.IsNotExist(err):
		return outDirForeign, nil
	default:
		return 0, err
	}
}

// errForeignOutDir is the refusal. It names the directory, says why it stopped,
// and says what would make it proceed — because the reader is looking at a build
// that did nothing, and the useful question is which case they are in.
//
// There are three, and the third is the one worth spelling out: a tree left by a
// build from before the marker existed is genuinely ours, and neither "use an
// empty directory" nor "use one an earlier build wrote" tells that reader
// anything they can act on. Emptying it themselves does.
func errForeignOutDir(outDir string) error {
	return fmt.Errorf("site: %s is not empty and holds no %s, so this build cannot tell it apart from a directory it did not write; refusing to remove it — point --out at an empty directory, or empty this one yourself if it is an old build's output",
		outDir, siteMarkerName)
}

// purgeOutDir clears a directory a previous build wrote, keeping the marker.
//
// The ENTRIES go, not the directory: a caller may be sitting in it, it may be a
// symlink or a mount point, and re-creating it would change what it is.
//
// The marker STAYS, and that is the whole subtlety here. `os.ReadDir` returns
// names in order and `.abcd-site-build` sorts ahead of everything, so removing
// it like any other entry would remove it first — and a purge interrupted after
// that point (killed process, full disk, one unremovable file) leaves a
// non-empty directory carrying no marker, which every later build then refuses.
// The tool would jam on its own wreckage and need a human with `rm` to get it
// going again. Keeping the marker costs nothing: the build rewrites it with
// identical bytes, so it is present at every instant and correct at the end.
//
// A symlink among the entries is removed as a LINK; `os.RemoveAll` does not
// follow it, so nothing outside this directory is reachable from here.
func purgeOutDir(outDir string) error {
	entries, err := os.ReadDir(outDir)
	if err != nil {
		return err
	}
	for _, e := range entries {
		if e.Name() == siteMarkerName {
			continue
		}
		if err := os.RemoveAll(filepath.Join(outDir, e.Name())); err != nil {
			return err
		}
	}
	return nil
}

// The installer: one committed template, served at one route. It is NOT in
// copiedSources because it is the one copied file the build adds a byte to —
// the stamp comment — and a reader is owed the difference being deliberate.
const (
	installTemplateRelPath = "site-src/install.sh.tmpl"
	installScriptName      = "install.sh"
)

// renderInstallScript renders the committed installer to the bytes served at
// /install.sh: the template, plus one comment naming the build that published
// it.
//
// The template is complete shell and the render is a copy — nothing is
// substituted, because a reader who verifies the script against the file in the
// repository must find them the same, and a template with holes in it makes that
// comparison a judgement call. The stamp is a COMMENT, and it goes after the
// shebang, which must stay on the first line for the kernel and for `sh` to read
// it.
//
// A build with no version and no commit adds nothing. "built from " naming
// nothing is worse than silence: it looks like a stamp that failed rather than a
// build that had nothing to stamp.
func renderInstallScript(tmpl []byte, stamp BuildStamp) []byte {
	var parts []string
	if v := stampWord(strings.TrimPrefix(stamp.Version, "v")); v != "" {
		parts = append(parts, "v"+v)
	}
	if c := stampWord(stamp.Commit); c != "" {
		parts = append(parts, c)
	}
	if len(parts) == 0 {
		return tmpl
	}
	comment := "# built from " + strings.Join(parts, " ") + "\n"

	body := string(tmpl)
	if !strings.HasPrefix(body, "#!") {
		return []byte(comment + body)
	}
	// After the shebang LINE, so a script that is one line long still gets a
	// stamp on a line of its own.
	cut := strings.Index(body, "\n")
	if cut < 0 {
		return []byte(body + "\n" + comment)
	}
	return []byte(body[:cut+1] + comment + body[cut+1:])
}

// stampWord reduces one stamp field to what may appear inside a shell comment.
//
// The fields are repository facts — a version parsed out of the changelog, a
// git object name — so this is not defence against an attacker. It is defence
// against the OUTPUT FORMAT: a comment is terminated by a newline, and this file
// is a script people pipe into a shell, so a stamp carrying one would not
// corrupt a document, it would add a COMMAND. Nothing downstream would notice,
// because the result is still a valid script.
//
// The rule is a whitelist rather than an escape, because there is no escaping
// inside a `#` comment: the only safe answer to a character that could end the
// line is to not write it. What survives is what a version or an object name is
// made of; a field left empty by it is treated as a field that said nothing.
func stampWord(s string) string {
	var b strings.Builder
	for _, r := range s {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9',
			r == '.', r == '-', r == '_', r == '+':
			b.WriteRune(r)
		}
	}
	return b.String()
}

// pluginManifestRelPath carries the package's own metadata: its name, its forge
// URL, its licence and its author. The site reads them rather than restating
// them, so there is one place a rename has to happen.
const pluginManifestRelPath = ".claude-plugin/plugin.json"

const maxPluginManifestBytes = 256 * 1024

// RepoMeta is the package metadata the site links and credits from.
type RepoMeta struct {
	Name       string `json:"name"`
	Repository string `json:"repository"`
	License    string `json:"license"`
	Author     struct {
		Name string `json:"name"`
	} `json:"author"`
	// AuthorName flattens Author.Name for the footer.
	AuthorName string `json:"-"`
}

// LoadRepoMeta reads the package manifest. An absent one is a state: the pages
// that would carry a forge link render without it.
func LoadRepoMeta(repoRoot string) (RepoMeta, error) {
	data, err := fsutil.ReadGuarded(joinRepo(repoRoot, pluginManifestRelPath), maxPluginManifestBytes)
	if err != nil {
		if os.IsNotExist(err) {
			return RepoMeta{}, nil
		}
		return RepoMeta{}, err
	}
	var m RepoMeta
	if err := json.Unmarshal(data, &m); err != nil {
		return RepoMeta{}, fmt.Errorf("site: %s: %w", pluginManifestRelPath, err)
	}
	m.AuthorName = m.Author.Name
	return m, nil
}

// Request is one build.
type Request struct {
	// RepoRoot is the repository to render.
	RepoRoot string
	// OutDir is where to write, absolute or relative to the caller's cwd.
	OutDir string
	// Stamp is the build metadata the footer and record.json carry. Every field
	// is injected so a test can pin the output byte for byte; empty fields are
	// filled from the repository (the version from the changelog, the commit
	// from git HEAD).
	Stamp BuildStamp
}

// Result describes what a build wrote.
type Result struct {
	OutDir string `json:"out_dir"`
	// Files are the output-relative paths written, sorted.
	Files []string `json:"files"`
	// Bytes is the total size written.
	Bytes int64 `json:"bytes"`
	// Pages is how many HTML pages the explorer rendered, beyond the landing
	// page: one per route and one per record.
	Pages int `json:"pages"`
	// Records, Links and Mentions summarise the exported graph.
	Records  int `json:"records"`
	Links    int `json:"links"`
	Mentions int `json:"mentions"`
	// Unresolved is the number of typed references the record cannot account
	// for; Baseline the committed ratchet's size.
	Unresolved int `json:"unresolved"`
	Baseline   int `json:"baseline"`
	// Overlaps is the coil packing's sanity count. It is zero or the chart is
	// wrong, and it is reported rather than assumed.
	Overlaps int `json:"overlaps"`
	// Version and Commit are what the footer says the site was built from.
	Version string `json:"version"`
	Commit  string `json:"commit"`
}

// ErrNoManifest is returned when the repository declares no composition.
var ErrNoManifest = errors.New("site: this repo declares no site composition (" + ManifestRelPath + " is absent)")

// Build renders the site into req.OutDir.
func Build(req Request) (Result, error) {
	repoRoot := req.RepoRoot
	outDir := req.OutDir
	if outDir == "" {
		outDir = DefaultOutDir
	}
	if !filepath.IsAbs(outDir) {
		abs, err := filepath.Abs(outDir)
		if err != nil {
			return Result{}, err
		}
		outDir = abs
	}

	// Refuse EARLY. The render below is minutes of work, and a build that waits
	// until the writes to discover it may not touch the directory has spent them
	// for nothing. The PURGE waits, though: it happens at the first write, so a
	// failure anywhere in between leaves the previous output standing rather than
	// clearing a good tree in exchange for one that never arrived.
	outState, err := inspectOutDir(outDir)
	if err != nil {
		return Result{}, err
	}
	if outState == outDirForeign {
		return Result{}, errForeignOutDir(outDir)
	}

	manifest, err := LoadManifest(repoRoot)
	if err != nil {
		if os.IsNotExist(err) {
			return Result{}, ErrNoManifest
		}
		return Result{}, err
	}
	ui, err := LoadUI(repoRoot, manifest.UIStrings)
	if err != nil {
		return Result{}, err
	}
	repo, err := LoadRepoMeta(repoRoot)
	if err != nil {
		return Result{}, err
	}

	lintCfg, err := lint.LoadConfig(joinRepo(repoRoot, ".abcd/record-lint.json"))
	if err != nil && !os.IsNotExist(err) {
		return Result{}, err
	}
	graph, err := lint.LoadRecordGraph(lintCfg, repoRoot)
	if err != nil {
		return Result{}, err
	}
	// The principle store carries no frontmatter, so the lint scan cannot see
	// it. It joins the graph here, and a repository that keeps none simply has
	// none — the pages that would list them are omitted.
	principles, err := LoadPrinciples(repoRoot, PrinciplesDir(lintCfg))
	if err != nil {
		return Result{}, err
	}
	hist, err := LoadHistory(repoRoot)
	if err != nil {
		return Result{}, err
	}

	stamp := req.Stamp
	if stamp.Commit == "" {
		stamp.Commit = HeadCommit(repoRoot)
	}
	releases, _, err := changelog.DatedReleases(repoRoot)
	if err != nil {
		return Result{}, err
	}
	version, releaseDate := "", ""
	if len(releases) > 0 {
		version, releaseDate = releases[0].Version, releases[0].Date
	}
	if stamp.Version == "" {
		stamp.Version = version
	}
	if stamp.GeneratedAt == "" && len(releases) > 0 {
		// With no injected stamp the build still refuses to read the clock: the
		// newest release's own date is a fact about the tree, and a build of one
		// tree must produce one set of bytes.
		stamp.GeneratedAt = releases[0].Date
	}

	c := &composer{
		repoRoot: repoRoot, manifest: manifest, ui: ui, repo: repo,
		assets: newAssetPipe(repoRoot), graph: graph, history: hist, stamp: stamp,
		beta: isPreOne(version), version: version, releaseDate: releaseDate,
	}
	html, err := c.ComposeLanding()
	if err != nil {
		return Result{}, err
	}

	export, err := BuildRecordExport(repoRoot, manifest.Checks.UnresolvedReferenceBaseline, graph, principles, hist, stamp, manifest.Record)
	if err != nil {
		return Result{}, err
	}
	recordJSON, err := json.MarshalIndent(export, "", "  ")
	if err != nil {
		return Result{}, err
	}
	recordJSON = append(recordJSON, '\n')

	// The record explorer. Its optional inputs — a bibliography, a principle
	// store, a changelog — each omit the page they feed rather than failing.
	bib, err := LoadBibliography(repoRoot, CSLRelPath, AcknowledgementsRelPath)
	if err != nil {
		return Result{}, err
	}
	recordRoot := ""
	if len(lintCfg.Roots) > 0 {
		recordRoot = lintCfg.Roots[0]
	}
	pages, err := newExplorer(c, export, bib, recordRoot).Pages()
	if err != nil {
		return Result{}, err
	}

	res := Result{
		OutDir:     outDir,
		Records:    len(export.Nodes),
		Links:      len(export.Edges),
		Mentions:   len(export.Mentions),
		Unresolved: len(export.Health.Unresolved),
		Baseline:   export.Health.BaselineCount,
		Overlaps:   export.Layout.Overlaps,
		Version:    stamp.Version,
		Commit:     stamp.Commit,
	}

	// The output tree is a RENDER of this commit, not an accumulation of every
	// build that ever ran here. A file left by an earlier one is stale the moment
	// this build does not rewrite it, and it is stale INVISIBLY — it is served,
	// and it looks exactly like a file that built.
	if outState == outDirOurs {
		if err := purgeOutDir(outDir); err != nil {
			return Result{}, err
		}
	}
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return Result{}, err
	}
	root, err := os.OpenRoot(outDir)
	if err != nil {
		return Result{}, err
	}
	defer root.Close()

	write := func(rel string, data []byte) error {
		if !fsutil.ValidRelPath(rel) {
			return fmt.Errorf("site: refusing to write %q: not a relative path inside the output directory", rel)
		}
		if err := fsutil.WriteFileAtomicInRoot(root, rel, data, 0o644); err != nil {
			return err
		}
		res.Files = append(res.Files, rel)
		res.Bytes += int64(len(data))
		return nil
	}

	// FIRST, before anything else in the tree: a build killed halfway must still
	// leave a directory the next one recognises as its own. With the marker
	// written last, the wreckage of a failed build would carry none, and every
	// later build would refuse it — the tool jammed on its own debris, freed only
	// by hand.
	if err := write(siteMarkerName, []byte(siteMarkerBody)); err != nil {
		return Result{}, err
	}
	if err := write("index.html", []byte(html)); err != nil {
		return Result{}, err
	}
	if err := write("record.json", recordJSON); err != nil {
		return Result{}, err
	}
	// Written in a fixed order, so two builds of one tree write the same files
	// in the same sequence and the reported list is a function of the record.
	routes := make([]string, 0, len(pages))
	for route := range pages {
		routes = append(routes, route)
	}
	sort.Strings(routes)
	for _, route := range routes {
		if err := write(route, []byte(pages[route])); err != nil {
			return Result{}, err
		}
	}
	res.Pages = len(pages)
	for _, cp := range copiedSources {
		data, err := fsutil.ReadGuarded(joinRepo(repoRoot, cp.src), maxAssetBytes)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return Result{}, err
		}
		if err := write(cp.dst, data); err != nil {
			return Result{}, err
		}
	}
	// The installer. A repository that commits no template serves no /install.sh
	// — the same graceful absence every other copied source has.
	switch tmpl, err := fsutil.ReadGuarded(joinRepo(repoRoot, installTemplateRelPath), maxAssetBytes); {
	case err == nil:
		if err := write(installScriptName, renderInstallScript(tmpl, stamp)); err != nil {
			return Result{}, err
		}
	case os.IsNotExist(err):
	default:
		return Result{}, err
	}
	for _, pair := range c.assets.Copies() {
		data, err := fsutil.ReadGuarded(joinRepo(repoRoot, pair[0]), maxAssetBytes)
		if err != nil {
			return Result{}, err
		}
		if err := write(path.Clean(pair[1]), data); err != nil {
			return Result{}, err
		}
	}
	sort.Strings(res.Files)
	return res, nil
}

// isPreOne reports whether a version's major component is 0 — the rule the Beta
// badge renders on, so no copy changes at v1.
func isPreOne(version string) bool {
	if version == "" {
		return false
	}
	major, _, _ := strings.Cut(strings.TrimPrefix(version, "v"), ".")
	return major == "0"
}

// Status is the read-only board the bare verb renders: what the repository has
// declared, and what the last build left behind.
type Status struct {
	Manifest  bool   `json:"manifest"`
	UIStrings bool   `json:"ui_strings"`
	UIPath    string `json:"ui_strings_path,omitempty"`
	Baseline  bool   `json:"baseline"`
	BaselineN int    `json:"baseline_entries"`
	// BaselinePath is the ratchet the board actually read: the manifest's
	// declaration where it makes one, the default otherwise.
	BaselinePath string `json:"baseline_path"`
	OutDir       string `json:"out_dir"`
	OutExists    bool   `json:"out_exists"`
	OutFiles     int    `json:"out_files"`
	Chapters     int    `json:"chapters"`
	IssueLedge   bool   `json:"issue_ledger"`
	Version      string `json:"version"`
	Commit       string `json:"commit"`
}

// Describe reads the site's declared state without writing anything.
func Describe(repoRoot, outDir string) (Status, error) {
	if outDir == "" {
		outDir = DefaultOutDir
	}
	st := Status{OutDir: outDir, Commit: HeadCommit(repoRoot), BaselinePath: BaselineRelPath}
	baselineRel := ""
	m, err := LoadManifest(repoRoot)
	switch {
	case err == nil:
		st.Manifest = true
		st.Chapters = len(m.Home.Chapters)
		st.IssueLedge = m.Record.IssueLedger
		st.UIPath = m.UIStrings
		if ok, _ := fsutil.Exists(joinRepo(repoRoot, m.UIStrings)); ok {
			st.UIStrings = true
		}
		if baselineRel = m.Checks.UnresolvedReferenceBaseline; baselineRel != "" {
			st.BaselinePath = baselineRel
		}
	case os.IsNotExist(err):
	default:
		return Status{}, err
	}
	// The board reports the path it ACTUALLY read, so a reader can tell the
	// declared ratchet from the default one without opening the manifest.
	//
	// A baseline that is declared but missing, or present but unreadable, is
	// REPORTED here rather than raised. This verb is the read-only board, and a
	// broken baseline is precisely the kind of thing somebody runs it to
	// discover — a board that exits non-zero instead of saying "absent" answers
	// the question by refusing to answer it. `build` still refuses, because a
	// health count measured against nothing would be published as if it meant
	// something.
	b, ok, err := LoadBaseline(repoRoot, baselineRel)
	if err != nil {
		ok, b = false, Baseline{}
	}
	st.Baseline, st.BaselineN = ok, len(b.UnresolvedReferences)

	releases, _, err := changelog.DatedReleases(repoRoot)
	if err != nil {
		return Status{}, err
	}
	if len(releases) > 0 {
		st.Version = releases[0].Version
	}

	abs := outDir
	if !filepath.IsAbs(abs) {
		abs = joinRepo(repoRoot, outDir)
	}
	entries, err := os.ReadDir(abs)
	if err == nil {
		st.OutExists = true
		st.OutFiles = len(entries)
	}
	return st, nil
}
