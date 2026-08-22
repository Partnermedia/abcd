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

	export, err := BuildRecordExport(repoRoot, manifest.Checks.UnresolvedReferenceBaseline, graph, hist, stamp, manifest.Record)
	if err != nil {
		return Result{}, err
	}
	recordJSON, err := json.MarshalIndent(export, "", "  ")
	if err != nil {
		return Result{}, err
	}
	recordJSON = append(recordJSON, '\n')

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

	if err := write("index.html", []byte(html)); err != nil {
		return Result{}, err
	}
	if err := write("record.json", recordJSON); err != nil {
		return Result{}, err
	}
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
	b, ok, err := LoadBaseline(repoRoot, baselineRel)
	if err != nil {
		return Status{}, err
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
