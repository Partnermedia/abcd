package reading

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/intentdriven/abcd/internal/core/lint"
	"github.com/intentdriven/abcd/internal/fsutil"
	"github.com/intentdriven/abcd/internal/gitutil"
)

// AssembleRequest is one assembly: a position, a target commit, and where the
// two artefacts go. It carries no free-text operand of any kind, because there
// is no channel through which ledger content may travel in the framing of a
// request (ruling (5)).
type AssembleRequest struct {
	// RepoRoot is the repository the assembly reads.
	RepoRoot string
	// Position is the reading position. The set is closed.
	Position Position
	// Target is "HEAD" or a hexadecimal commit sha of 7 to 40 digits. Branch
	// names and tags are refused as mutable: the manifest's re-runnability rests
	// on a reference that cannot move.
	Target string
	// OutDir is the operator-named directory the two artefacts are written to.
	// Empty means the default local-tier run directory.
	OutDir string
	// DryRun writes nothing into the repository's own tiers. With OutDir set the
	// artefacts still land there; with OutDir empty nothing is written at all
	// and the result is rendered only.
	DryRun bool
}

// AssembleResult is what one assembly produced. The bundle and the manifest are
// carried for a caller that wants them in memory, and the artefacts are named by
// their basenames rather than by a full path.
//
// OutDir is the one field that can hold an absolute path, and only because the
// operator supplied one to --out; it is echoed back verbatim so a caller can
// find what it asked for. Neither ARTEFACT carries it: no path the operator did
// not type reaches this result, and nothing committed reaches it at all.
type AssembleResult struct {
	RunID            string   `json:"run_id"`
	Position         Position `json:"position"`
	TargetCommit     string   `json:"target_commit"`
	AssemblerVersion string   `json:"assembler_version"`
	ItemCount        int      `json:"item_count"`
	ManifestHash     string   `json:"manifest_hash"`
	OutDir           string   `json:"out_dir,omitempty"`
	Artefacts        []string `json:"artefacts"`
	Written          bool     `json:"written"`

	Bundle   Bundle   `json:"-"`
	Manifest Manifest `json:"-"`
}

// BundleFileName and ManifestFileName are the two artefacts an assembly writes:
// separate files, so the assembled input can go to a reader while the manifest
// stays with the auditor.
const (
	BundleFileName   = "bundle.json"
	ManifestFileName = "manifest.json"
)

// DefaultRunDir is the local-tier parent an unnamed run is parked under.
const DefaultRunDir = ".abcd/.work.local/scratch/reading-runs"

// MaxFileBytes bounds one admitted file. A file past the cap is a refusal, not
// a truncation: a silently shortened item would be an assembled input no re-run
// could reproduce from the manifest's hash.
const MaxFileBytes = 4 << 20

// LintConfigPath is the record-lint configuration the record scan reads its
// stores from. Enumeration comes from that scan and nowhere else: there is one
// parser of the record's shape in this binary.
const LintConfigPath = ".abcd/record-lint.json"

// lintRecordSchemaRule is the rule whose configuration names the record stores.
const lintRecordSchemaRule = "record_schema"

var (
	// targetRe is the whole grammar of the second operand besides "HEAD".
	targetRe = regexp.MustCompile(`^[0-9a-f]{7,40}$`)
	// itemKeyRe is the bundle key's shape: an ordinal, never a location.
	itemKeyRe = regexp.MustCompile(`^itm-[0-9]{4}$`)
)

// storeNodeType maps an include row's record-store prefix to the node type the
// record graph reports. It is a translation, not a second declaration of which
// stores exist: the graph owns that, and an entry here for a store no row names
// would be a claim this package reads a family it structurally cannot.
var storeNodeType = map[string]string{
	"itd": "intent",
	"spc": "spec",
}

// candidate is one admitted (file, field) pair before it is keyed.
type candidate struct {
	path     string
	field    string
	fieldIdx int
	kind     Kind
	text     string
}

// Assemble walks the repository under the include table at the given position
// and produces the assembled input and its manifest.
func Assemble(req AssembleRequest) (AssembleResult, error) {
	position, err := ParsePosition(string(req.Position))
	if err != nil {
		return AssembleResult{}, err
	}
	if req.RepoRoot == "" {
		return AssembleResult{}, fmt.Errorf("reading: no repository root given")
	}
	target, err := resolveTarget(req.RepoRoot, req.Target)
	if err != nil {
		return AssembleResult{}, err
	}
	if err := refuseSelfAdmittingOutDir(req.RepoRoot, req.OutDir); err != nil {
		return AssembleResult{}, err
	}

	cands, err := collect(req.RepoRoot, position)
	if err != nil {
		return AssembleResult{}, err
	}
	if err := refuseDirtyIncludedPaths(req.RepoRoot, position, cands); err != nil {
		return AssembleResult{}, err
	}

	exclusions := ExclusionsFor(position)
	if err := assertExclusions(cands, exclusions); err != nil {
		return AssembleResult{}, err
	}

	runID, err := mintRunID()
	if err != nil {
		return AssembleResult{}, fmt.Errorf("reading: minting a run id: %w", err)
	}

	bundle := Bundle{
		Type:          BundleType,
		SchemaVersion: SchemaVersion,
		Position:      position,
		Items:         make([]BundleItem, 0, len(cands)),
	}
	manifest := Manifest{
		Type:             ManifestType,
		SchemaVersion:    SchemaVersion,
		RunID:            runID,
		Position:         position,
		TargetCommit:     target,
		AssemblerVersion: AssemblerVersion,
		Items:            make([]ManifestItem, 0, len(cands)),
		Exclusions:       exclusions,
	}
	for i, c := range cands {
		key := fmt.Sprintf("itm-%04d", i+1)
		bundle.Items = append(bundle.Items, BundleItem{ItemKey: key, Kind: c.kind, Text: c.text})
		manifest.Items = append(manifest.Items, ManifestItem{
			ItemKey: key, Path: c.path, Field: c.field, SHA256: sha256Hex([]byte(c.text)),
		})
	}

	hash, err := ManifestHash(manifest)
	if err != nil {
		return AssembleResult{}, err
	}
	res := AssembleResult{
		RunID:            runID,
		Position:         position,
		TargetCommit:     target,
		AssemblerVersion: AssemblerVersion,
		ItemCount:        len(bundle.Items),
		ManifestHash:     hash,
		Artefacts:        []string{},
		Bundle:           bundle,
		Manifest:         manifest,
	}

	outDir, writeIt := resolveOutDir(req, runID)
	res.OutDir = outDir
	if !writeIt {
		return res, nil
	}
	if err := writeArtefacts(req.RepoRoot, outDir, bundle, manifest); err != nil {
		return AssembleResult{}, err
	}
	res.Written = true
	res.Artefacts = []string{BundleFileName, ManifestFileName}
	return res, nil
}

// resolveOutDir decides where the two artefacts go and whether they are written
// at all. A dry run with no operator-named directory writes nothing, on the
// render-without-writing idiom `disembark plan` carries.
func resolveOutDir(req AssembleRequest, runID string) (string, bool) {
	if req.OutDir != "" {
		return req.OutDir, true
	}
	if req.DryRun {
		return "", false
	}
	return DefaultRunDir + "/" + runID, true
}

// writeArtefacts writes the assembled input and the manifest as two separate
// files. A relative output directory is taken against the repository root; an
// absolute one is used as given.
func writeArtefacts(repoRoot, outDir string, b Bundle, m Manifest) error {
	dir := outDir
	if !filepath.IsAbs(dir) {
		dir = filepath.Join(repoRoot, filepath.FromSlash(outDir))
	}
	if err := requireEmptyDir(outDir, dir); err != nil {
		return err
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("reading: creating the run directory: %w", err)
	}
	bundleRaw, err := EncodeBundle(b)
	if err != nil {
		return err
	}
	manifestRaw, err := EncodeManifest(m)
	if err != nil {
		return err
	}
	// Both artefacts go through the repository's one atomic write: temp file,
	// then rename. A reader never opens a half-written bundle, and a run that
	// dies mid-write leaves no artefact rather than a plausible short one whose
	// hash nothing matches.
	if err := fsutil.WriteFileAtomic(filepath.Join(dir, BundleFileName), bundleRaw, 0o644); err != nil {
		return fmt.Errorf("reading: writing the assembled input: %w", err)
	}
	if err := fsutil.WriteFileAtomic(filepath.Join(dir, ManifestFileName), manifestRaw, 0o644); err != nil {
		return fmt.Errorf("reading: writing the manifest: %w", err)
	}
	return nil
}

// requireEmptyDir refuses an output directory that already holds something. The
// two artefacts are one run's evidence, and dropping them beside another run's
// leaves a directory whose manifest describes only half of what is in it.
func requireEmptyDir(named, dir string) error {
	entries, err := os.ReadDir(dir)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("reading: reading the output directory: %w", err)
	}
	if len(entries) > 0 {
		return fmt.Errorf("reading: the output directory %s is not empty (%d entr(y|ies)); one run's "+
			"artefacts are one run's evidence, so name an empty or absent directory", named, len(entries))
	}
	return nil
}

// refuseSelfAdmittingOutDir refuses an output directory the include table can
// reach. Writing a run where the table admits it is committing the NEXT run's
// contamination: the artefacts land as ordinary files, a later commit puts them
// in the tree, and ruling (18) is breached by a path the operator chose one run
// earlier. The refusal belongs at the moment the directory is named.
//
// Only a directory inside the repository can be reached, so an output path that
// resolves outside it is always fine.
func refuseSelfAdmittingOutDir(repoRoot, outDir string) error {
	if outDir == "" {
		return nil
	}
	abs := outDir
	if !filepath.IsAbs(abs) {
		abs = filepath.Join(repoRoot, filepath.FromSlash(outDir))
	}
	rel, err := filepath.Rel(repoRoot, abs)
	if err != nil {
		return nil // not expressible against the repository, so not inside it
	}
	rel = filepath.ToSlash(rel)
	if rel == ".." || strings.HasPrefix(rel, "../") {
		return nil
	}
	for _, name := range []string{BundleFileName, ManifestFileName} {
		candidate := path.Join(rel, name)
		for _, p := range Positions() {
			if Admits(p, candidate) {
				return fmt.Errorf("reading: the output directory %s is inside the include table's reach "+
					"(%s would be admitted at the %s position), so a committed run would become a later "+
					"run's input; write outside the repository, or under %s", outDir, candidate, p, DefaultRunDir)
			}
		}
	}
	return nil
}

// resolveTarget validates the second operand and resolves it against HEAD.
//
// Assembly reads the WORKING TREE, so a target that is not the commit in front
// of the assembler is a refusal rather than a checkout: the manifest would
// otherwise name a commit whose content it never read.
func resolveTarget(repoRoot, target string) (string, error) {
	if target != "HEAD" && !targetRe.MatchString(target) {
		return "", fmt.Errorf("reading: target %q is neither HEAD nor a hexadecimal commit sha of 7 to 40 digits; "+
			"branch names and tags move, and the manifest's re-runnability rests on a reference that cannot", target)
	}
	head, err := gitutil.Run(repoRoot, "rev-parse", "HEAD")
	if err != nil {
		return "", fmt.Errorf("reading: resolving HEAD: %w", err)
	}
	head = strings.TrimSpace(head)
	if head == "" {
		return "", fmt.Errorf("reading: HEAD does not resolve to a commit")
	}
	if target != "HEAD" && !strings.HasPrefix(head, target) {
		return "", fmt.Errorf("reading: HEAD is %s, which is not the target %s; assembly reads the working tree",
			head, target)
	}
	return head, nil
}

// refuseDirtyIncludedPaths refuses an assembly whose own input is uncommitted.
// Dirtiness elsewhere in the tree is not the assembler's business: a reading
// cannot see it, so it cannot contaminate the run.
//
// A path counts as included on either of two grounds, and it needs both. The
// assembly's own item paths catch an edit or an untracked addition. The include
// table's admissibility catches the case the item paths cannot: a file DELETED
// in the working tree is in neither the assembly nor the walk, yet it is part of
// the commit the manifest names, so a run over it would describe a target it
// never read.
func refuseDirtyIncludedPaths(repoRoot string, position Position, cands []candidate) error {
	// -uall, not the default -unormal: git collapses an untracked DIRECTORY to a
	// single entry, and an admitted file inside a newly created directory would
	// then never be named — the prefix check below can only test paths the
	// assembly already holds, and an untracked file is not one of them.
	out, err := gitutil.RunCapped(repoRoot, 8<<20, "status", "--porcelain=v1", "-z", "-uall")
	if err != nil {
		return fmt.Errorf("reading: reading the working-tree status: %w", err)
	}
	included := make(map[string]bool, len(cands))
	for _, c := range cands {
		included[c.path] = true
	}
	// The record configuration decides which stores the record scan reads, so an
	// uncommitted edit to it reshapes the assembly as surely as an edit to a
	// record does. It sits under the deny, so no include row ever puts it in this
	// set; it is named here instead.
	included[LintConfigPath] = true
	var dirty []string
	for _, entry := range dirtyPaths(out) {
		if strings.HasSuffix(entry, "/") {
			for p := range included {
				if strings.HasPrefix(p, entry) {
					dirty = append(dirty, p)
				}
			}
			continue
		}
		if included[entry] || Admits(position, entry) {
			dirty = append(dirty, entry)
		}
	}
	if len(dirty) == 0 {
		return nil
	}
	sort.Strings(dirty)
	return fmt.Errorf("reading: %d included path(s) are uncommitted, starting with %s; "+
		"a dirty tree cannot be described by a commit reference, so the manifest would promise "+
		"a re-run it could not deliver", len(dirty), dirty[0])
}

// dirtyPaths parses `git status --porcelain=v1 -z` into the paths it reports.
//
// The -z form is what makes this parseable: it emits each entry NUL-terminated
// and never quotes or escapes a path, so a filename holding a space, a quote or
// a newline arrives verbatim and core.quotepath cannot change the format under
// the parser. A rename or copy entry carries its source as the following
// record, which is consumed with it.
func dirtyPaths(out string) []string {
	records := strings.Split(out, "\x00")
	var paths []string
	for i := 0; i < len(records); i++ {
		rec := records[i]
		if len(rec) < 4 {
			continue
		}
		status := rec[:2]
		paths = append(paths, rec[3:])
		// A rename or copy carries its SOURCE as the following record, and either
		// status column can declare one: `R ` is a staged rename, ` R` a worktree
		// one. The source is the path that was in the target commit, so dropping
		// it loses exactly the file whose disappearance from the include set this
		// gate exists to catch.
		if status[0] == 'R' || status[0] == 'C' || status[1] == 'R' || status[1] == 'C' {
			i++
			if i < len(records) && records[i] != "" {
				paths = append(paths, records[i])
			}
		}
	}
	return paths
}

// assertExclusions is the fail-closed half of the exclusion floor: the manifest
// DECLARES what was refused, and this refuses to emit an item that contradicts
// the declaration. A floor a run can quietly violate is a disclosure, not a
// gate.
func assertExclusions(cands []candidate, exclusions []Exclusion) error {
	for _, e := range exclusions {
		if e.Signal != "directory" && e.Signal != "file" {
			continue
		}
		for _, c := range cands {
			if c.path == e.Detail || strings.HasPrefix(c.path, e.Detail+"/") {
				return fmt.Errorf("reading: item %s lies under the excluded %s %s, which the manifest asserts was refused",
					c.path, e.Signal, e.Detail)
			}
		}
	}
	return nil
}

// collect gathers every admitted (file, field) pair at the position, in
// lexicographic path order with each file's fields in the order its row names
// them. The first row that reaches a path owns the projection applied to it.
func collect(repoRoot string, position Position) ([]candidate, error) {
	graph, err := loadGraph(repoRoot)
	if err != nil {
		return nil, err
	}
	tracked, err := trackedSet(repoRoot)
	if err != nil {
		return nil, err
	}
	claimed := map[string]bool{}
	var out []candidate

	exclusions := ExclusionsFor(position)
	for _, row := range Table {
		if !row.AdmittedAt(position) {
			continue
		}
		paths, err := rowPaths(repoRoot, row, graph)
		if err != nil {
			return nil, err
		}
		for _, rel := range paths {
			if claimed[rel] || !tracked[rel] {
				continue
			}
			claimed[rel] = true
			raw, err := fsutil.ReadGuarded(filepath.Join(repoRoot, filepath.FromSlash(rel)), MaxFileBytes)
			if err != nil {
				return nil, fmt.Errorf("reading: %s: %w", rel, err)
			}
			if err := refuseOwnArtefact(rel, raw); err != nil {
				return nil, err
			}
			doc, err := redactExcluded(rel, string(raw), exclusions)
			if err != nil {
				return nil, err
			}
			if len(row.Fields) == 0 {
				out = append(out, candidate{path: rel, kind: row.Kind, text: doc})
				continue
			}
			for i, field := range row.Fields {
				text, ok, err := projectField(rel, doc, field)
				if err != nil {
					return nil, err
				}
				if !ok {
					continue
				}
				out = append(out, candidate{path: rel, field: field, fieldIdx: i, kind: row.Kind, text: text})
			}
		}
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].path != out[j].path {
			return out[i].path < out[j].path
		}
		return out[i].fieldIdx < out[j].fieldIdx
	})
	return out, nil
}

// refuseOwnArtefact refuses an admitted file that IS one of this assembler's own
// artefacts. Ruling (18) says the instrument's own output never becomes its
// input, and the path deny alone cannot keep that promise: a run written to a
// directory the table reaches, then committed, arrives as an ordinary config
// item, its manifest's repository paths riding into the bundle text while the
// current run's manifest still asserts the exclusion.
//
// The artefacts self-identify, which is what makes this checkable rather than
// guessed at: both carry a top-level `_type`. A file that is not JSON, or whose
// JSON carries neither tag, is not one of ours and passes untouched.
func refuseOwnArtefact(rel string, raw []byte) error {
	if !strings.EqualFold(path.Ext(rel), ".json") {
		return nil
	}
	var probe struct {
		Type string `json:"_type"`
	}
	if err := json.Unmarshal(raw, &probe); err != nil {
		return nil // not a JSON object this assembler wrote
	}
	if probe.Type != BundleType && probe.Type != ManifestType {
		return nil
	}
	return fmt.Errorf("reading: %s is this assembler's own output (_type %s), and the instrument's own "+
		"output never becomes its input; move it outside the include table's reach", rel, probe.Type)
}

// requireConfiguredStores refuses a configuration that is silent about a store
// the include table names. It is the same refusal as an absent configuration,
// arriving one level in: an unnamed store contributes nothing to the record
// scan, and a row enumerating nothing is a hole the run would not report.
func requireConfiguredStores(repoRoot string, cfg lint.Config) error {
	configured := cfg.Rules[lintRecordSchemaRule].RecordStores
	for _, row := range Table {
		if row.Store == "" {
			continue
		}
		dir, ok := configured[row.Store]
		if !ok {
			return fmt.Errorf("reading: %s names no %q record store, so the include row %q would "+
				"enumerate nothing", LintConfigPath, row.Store, row.Source)
		}
		// A key present is not a store present. A retarget — a typo, a rename the
		// configuration did not follow — leaves the key in place and points it at
		// nothing, and the scan then reports an empty store exactly as it reports
		// a store with no records.
		info, err := os.Stat(filepath.Join(repoRoot, filepath.FromSlash(dir)))
		if err != nil || !info.IsDir() {
			return fmt.Errorf("reading: %s points the %q record store at %s, which is not a directory, "+
				"so the include row %q would enumerate nothing", LintConfigPath, row.Store, dir, row.Source)
		}
	}
	return nil
}

// trackedSet is the file set the target commit actually carries, read from git
// rather than from the filesystem.
//
// The walk and the dirty gate ask two different sources, and the gap between
// them is exactly the class of file the manifest's re-runnability rests on: a
// GITIGNORED file matching an include row is on disk, so a filesystem walk
// passes it, and `git status` says nothing about it, so the dirty gate cannot
// refuse it. Build output, a virtual environment and a vendored tree all land
// there, and an auditor re-running the assembly in a clean clone would get a
// different bundle under a different hash.
//
// Intersecting the walk with the tracked set closes that, and closes a
// submodule's inner content with it: git reports a gitlink, never the files
// beneath it, so they are absent from the tracked set by construction. An
// untracked file that is NOT ignored stays a refusal rather than a silent
// omission — the dirty gate sees it, and a genuine divergence from the target
// commit must be said out loud.
func trackedSet(repoRoot string) (map[string]bool, error) {
	files, err := gitutil.TrackedFiles(repoRoot)
	if err != nil {
		return nil, fmt.Errorf("reading: listing the tracked files: %w", err)
	}
	set := make(map[string]bool, len(files))
	for _, f := range files {
		set[f] = true
	}
	return set, nil
}

// loadGraph reads the record corpus once, through the record_schema rule's own
// scan. A second parser of the record's shape would drift the moment a store
// gained a bucket, so there is not one.
func loadGraph(repoRoot string) (lint.RecordGraph, error) {
	lintRoot, err := os.OpenRoot(repoRoot)
	if err != nil {
		return lint.RecordGraph{}, fmt.Errorf("reading: opening the repository root: %w", err)
	}
	cfg, err := lint.LoadConfigInRoot(lintRoot, LintConfigPath)
	closeErr := lintRoot.Close()
	if err != nil {
		// A missing configuration is a REFUSAL, not an empty result. Without it
		// every record row contributes nothing and the run reports a clean
		// assembly of a reading that saw none of the record it exists to read
		// against — a silence indistinguishable from a repository with no record
		// at all.
		if os.IsNotExist(err) {
			return lint.RecordGraph{}, fmt.Errorf("reading: %s is absent, so the record scan enumerates "+
				"nothing and every record the include table names would be silently missing", LintConfigPath)
		}
		return lint.RecordGraph{}, fmt.Errorf("reading: loading %s: %w", LintConfigPath, err)
	}
	if closeErr != nil {
		return lint.RecordGraph{}, fmt.Errorf("reading: closing the repository root: %w", closeErr)
	}
	if err := requireConfiguredStores(repoRoot, cfg); err != nil {
		return lint.RecordGraph{}, err
	}
	graph, err := lint.LoadRecordGraph(cfg, repoRoot)
	if err != nil {
		return lint.RecordGraph{}, fmt.Errorf("reading: scanning the record: %w", err)
	}
	return graph, nil
}

// rowPaths resolves one row to the repo-relative files it admits, sorted.
func rowPaths(repoRoot string, row Row, graph lint.RecordGraph) ([]string, error) {
	var paths []string
	if row.Store != "" {
		nodeType, ok := storeNodeType[row.Store]
		if !ok {
			return nil, fmt.Errorf("reading: include row %q names an unknown record store %q", row.Source, row.Store)
		}
		for _, n := range graph.Nodes {
			if n.Type != nodeType {
				continue
			}
			if row.Bucket != "" && n.Lifecycle != row.Bucket {
				continue
			}
			if row.Reaches(n.Path) {
				paths = append(paths, n.Path)
			}
		}
		sort.Strings(paths)
		return paths, nil
	}

	base := repoRoot
	if row.Source != "." {
		base = filepath.Join(repoRoot, filepath.FromSlash(row.Source))
	}
	if _, err := os.Stat(base); os.IsNotExist(err) {
		return nil, nil
	}
	err := filepath.WalkDir(base, func(abs string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, rerr := filepath.Rel(repoRoot, abs)
		if rerr != nil {
			return rerr
		}
		rel = filepath.ToSlash(rel)
		if rel == "." {
			return nil
		}
		if d.IsDir() {
			if !reachableDir(row, rel) {
				return fs.SkipDir
			}
			return nil
		}
		if d.Type()&fs.ModeSymlink != 0 {
			return nil // a symlinked leaf is never followed
		}
		if row.Reaches(rel) {
			paths = append(paths, rel)
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("reading: walking %s: %w", row.Source, err)
	}
	sort.Strings(paths)
	return paths, nil
}

// reachableDir reports whether the walk may descend into a directory. It is the
// deny applied one level early, so a denied namespace is pruned rather than
// walked and discarded file by file.
func reachableDir(row Row, rel string) bool {
	sub := rel
	if row.Source != "." {
		src := path.Clean(row.Source)
		if rel == src {
			return true
		}
		if !strings.HasPrefix(rel, src+"/") {
			return false
		}
		sub = rel[len(src)+1:]
	}
	return !pathContainsDeniedSegment(sub) && !prefixDenied(rel)
}
