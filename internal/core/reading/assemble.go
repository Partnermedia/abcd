package reading

import (
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
// carried for a caller that wants them in memory; the artefacts are named by
// their basenames beside the output directory the operator asked for, so no
// absolute path enters the result.
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

var (
	// targetRe is the whole grammar of the second operand besides "HEAD".
	targetRe = regexp.MustCompile(`^[0-9a-f]{7,40}$`)
	// itemKeyRe is the bundle key's shape: an ordinal, never a location.
	itemKeyRe = regexp.MustCompile(`^itm-[0-9]{4}$`)
)

// storeNodeType maps an include row's record-store prefix to the node type the
// record graph reports. It is a translation, not a second declaration of which
// stores exist: the graph owns that.
var storeNodeType = map[string]string{
	"itd": "intent",
	"spc": "spec",
	"adr": "adr",
	"iss": "issue",
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

	cands, err := collect(req.RepoRoot, position)
	if err != nil {
		return AssembleResult{}, err
	}
	if err := refuseDirtyIncludedPaths(req.RepoRoot, cands); err != nil {
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
	if err := os.WriteFile(filepath.Join(dir, BundleFileName), bundleRaw, 0o644); err != nil {
		return fmt.Errorf("reading: writing the assembled input: %w", err)
	}
	if err := os.WriteFile(filepath.Join(dir, ManifestFileName), manifestRaw, 0o644); err != nil {
		return fmt.Errorf("reading: writing the manifest: %w", err)
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
func refuseDirtyIncludedPaths(repoRoot string, cands []candidate) error {
	out, err := gitutil.RunCapped(repoRoot, 8<<20, "status", "--porcelain=v1", "-z")
	if err != nil {
		return fmt.Errorf("reading: reading the working-tree status: %w", err)
	}
	included := make(map[string]bool, len(cands))
	for _, c := range cands {
		included[c.path] = true
	}
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
		if included[entry] {
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
// A rename or copy entry carries its source as the following record, which is
// consumed with it.
func dirtyPaths(out string) []string {
	records := strings.Split(out, "\x00")
	var paths []string
	for i := 0; i < len(records); i++ {
		rec := records[i]
		if len(rec) < 4 {
			continue
		}
		status := rec[:2]
		paths = append(paths, strings.TrimPrefix(rec[3:], "\""))
		if status[0] == 'R' || status[0] == 'C' {
			i++ // the source path of a rename or copy
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
			if claimed[rel] {
				continue
			}
			claimed[rel] = true
			raw, err := fsutil.ReadGuarded(filepath.Join(repoRoot, filepath.FromSlash(rel)), MaxFileBytes)
			if err != nil {
				return nil, fmt.Errorf("reading: %s: %w", rel, err)
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
	if err != nil && !os.IsNotExist(err) {
		return lint.RecordGraph{}, fmt.Errorf("reading: loading %s: %w", LintConfigPath, err)
	}
	if closeErr != nil {
		return lint.RecordGraph{}, fmt.Errorf("reading: closing the repository root: %w", closeErr)
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
