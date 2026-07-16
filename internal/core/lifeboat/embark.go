package lifeboat

// embark.go is the read-and-write half of the M5 record round-trip (itd-88,
// adr-35): it takes a PACKED lifeboat (an UNTRUSTED directory) and reports, or
// performs, the write of its record families back into a target repo. The
// lifeboat is untrusted input — every read is bounded, every path is validated,
// and no symlink is ever followed — so a hostile or corrupt lifeboat cannot
// exhaust memory, smuggle a file past manifest verification, or steer a write
// outside the target family. The core is transport-agnostic: it returns
// structured results and never prints (the surface renders them).
//
// The shape vocabulary (EmbarkPlan, EmbarkResult, Conflict, PlannedEmbark,
// IgnoredFile, MarkerResult, CoverageHandoff, the embarkFamilies inverse table
// and the closure/exclusion sets) lives in embark_types.go, shared with the
// packer so the two sides cannot drift.

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"github.com/REPPL/abcd-cli/internal/core/ahoy"
	"github.com/REPPL/abcd-cli/internal/fsutil"
)

// maxProvenanceBytes caps the _provenance.json read (its own manifest header).
const maxProvenanceBytes = 1 << 20

// EmbarkProbe inspects the lifeboat at lifeboatDir against targetDir WITHOUT
// writing. Both are resolved to absolute dirs (the CLI defaults targetDir to
// cwd). It gates the lifeboat, verifies the manifest, computes the plan (mapped
// writes, conflicts, ignored files), predicts the marker action, and reads the
// coverage handoff. It returns a structural error for a non-lifeboat /
// schema-too-new / failed-verification / bad-target input; a plan WITH conflicts
// is a SUCCESS (probe is a report, never a refusal).
func EmbarkProbe(lifeboatDir, targetDir string) (EmbarkPlan, error) {
	pr, err := runPlanner(lifeboatDir, targetDir)
	if err != nil {
		return EmbarkPlan{}, err
	}
	// Predict the marker action without writing (one code path with the write
	// side, so a probe cannot mispredict what a real embark would do).
	marker := embarkMarker(pr.targetAbs, true)
	return EmbarkPlan{
		SchemaVersion:        EmbarkSchemaVersion,
		LifeboatDir:          pr.lifeboatAbs,
		TargetDir:            pr.targetAbs,
		SourceName:           pr.prov.SourceName,
		ManifestVerified:     true,
		ManifestSHA256:       pr.prov.ManifestSHA256,
		RecordManifestSHA256: pr.prov.RecordManifestSHA256,
		Coverage:             pr.coverage,
		Planned:              pr.planned,
		Conflicts:            pr.conflicts,
		Ignored:              pr.ignored,
		Marker:               marker,
	}, nil
}

// EmbarkFrom performs the write. It runs the same planner as EmbarkProbe; if the
// plan carries ANY conflict it returns (result-with-Conflicts, ErrEmbarkConflicts)
// having written NOTHING. Otherwise it writes each ActionCreate file through
// os.Root containment + independent lexical validation + fsutil.WriteFileAtomic,
// skips ActionUnchanged files, ensures the marker last, and returns the summary.
func EmbarkFrom(lifeboatDir, targetDir string) (EmbarkResult, error) {
	pr, err := runPlanner(lifeboatDir, targetDir)
	if err != nil {
		return EmbarkResult{}, err
	}
	res := EmbarkResult{
		SchemaVersion: EmbarkSchemaVersion,
		LifeboatDir:   pr.lifeboatAbs,
		TargetDir:     pr.targetAbs,
		SourceName:    pr.prov.SourceName,
		Coverage:      pr.coverage,
		Ignored:       pr.ignored,
	}
	if len(pr.conflicts) > 0 {
		// Refuse: write nothing, hand the surface the full conflict slice for one
		// bulk report. The marker is only PREDICTED (dry run — no write), so the
		// "nothing was written" promise holds.
		res.Conflicts = pr.conflicts
		res.Marker = embarkMarker(pr.targetAbs, true)
		return res, ErrEmbarkConflicts
	}

	written, unchanged, bytesW, families, err := writeEmbark(pr.targetAbs, pr.planned)
	if err != nil {
		return EmbarkResult{}, fmt.Errorf("embark: %w", err)
	}
	res.Written = written
	res.Unchanged = unchanged
	res.BytesWritten = bytesW
	res.Families = families
	// Marker last: the records land first, then the current block is re-injected
	// (never foreign prose copied) into the target CLAUDE.md.
	res.Marker = embarkMarker(pr.targetAbs, false)
	return res, nil
}

// VerifyManifest re-hashes every non-excluded file in the lifeboat and compares
// the result to _provenance.json's manifest_sha256. It enforces the trust
// boundary during the walk: it refuses a symlink anywhere in the tree, a path
// that fails validRelPath, a file over maxEmbarkFileBytes, a tree over
// maxEmbarkFiles / maxEmbarkTotalBytes. A flipped, missing, or extra
// manifest-relevant file changes the reproduced hash and is fatal. It returns
// nil iff the lifeboat is intact. The excluded set (_provenance.json, the
// post-pack layer-3 graveyard/lessons.json and graveyard/low-confidence/**) is
// the same set the packer left out of manifest_sha256.
func VerifyManifest(dir string) error {
	abs, err := filepath.Abs(dir)
	if err != nil {
		return err
	}
	if !fsutil.IsRealDir(abs) {
		return fmt.Errorf("lifeboat %s is not a directory", filepath.Base(abs))
	}
	prov, err := readProvenance(abs)
	if err != nil {
		return err
	}
	root, err := os.OpenRoot(abs)
	if err != nil {
		return err
	}
	defer root.Close()

	rels, err := walkLifeboatFiles(root)
	if err != nil {
		return err
	}
	files := make([]PlannedFile, 0, len(rels))
	total := 0
	for _, rel := range rels {
		if isManifestExcluded(rel) {
			continue // the layer-3 interpretation and the header are not sealed
		}
		data, err := readLifeboatFile(root, abs, rel)
		if err != nil {
			return err
		}
		total += len(data)
		if total > maxEmbarkTotalBytes {
			return fmt.Errorf("lifeboat exceeds the %d-byte total cap", maxEmbarkTotalBytes)
		}
		files = append(files, PlannedFile{Path: rel, Content: data})
	}
	got := ManifestSHA256(files)
	if got != prov.ManifestSHA256 {
		return errors.New("lifeboat manifest verification failed: the on-disk tree does not match _provenance.json")
	}
	return nil
}

// plannerResult is the shared output of runPlanner, consumed by both EmbarkProbe
// and EmbarkFrom so the read path and the write path plan identically.
type plannerResult struct {
	lifeboatAbs string
	targetAbs   string
	prov        Provenance
	planned     []PlannedEmbark
	conflicts   []Conflict
	ignored     []IgnoredFile
	coverage    *CoverageHandoff
}

// runPlanner gates the lifeboat and the target, verifies the manifest, then
// walks the lifeboat mapping only the embarked families into planned writes (or
// conflicts), reporting every other file. It writes nothing. Ordering is
// deterministic: the walk is sorted, so planned/conflicts/ignored are emitted in
// lifeboat-path order.
func runPlanner(lifeboatDir, targetDir string) (plannerResult, error) {
	lifeboatAbs, err := filepath.Abs(lifeboatDir)
	if err != nil {
		return plannerResult{}, err
	}
	targetAbs, err := filepath.Abs(targetDir)
	if err != nil {
		return plannerResult{}, err
	}

	// Gate the lifeboat: a real directory carrying a parseable _provenance.json.
	if !fsutil.IsRealDir(lifeboatAbs) {
		return plannerResult{}, fmt.Errorf("lifeboat %s is not a directory", filepath.Base(lifeboatAbs))
	}
	if !isAbcdLifeboat(lifeboatAbs) {
		return plannerResult{}, fmt.Errorf("%s is not an abcd lifeboat (no parseable %s)", filepath.Base(lifeboatAbs), ProvenanceName)
	}
	prov, err := readProvenance(lifeboatAbs)
	if err != nil {
		return plannerResult{}, err
	}
	// Schema gate: refuse a lifeboat newer than this abcd understands, message
	// mirroring the graveyard lessons ingest.
	if prov.SchemaVersion > SchemaVersion {
		return plannerResult{}, fmt.Errorf("lifeboat schema v%d; this abcd knows up to v%d — upgrade abcd",
			prov.SchemaVersion, SchemaVersion)
	}
	if err := VerifyManifest(lifeboatAbs); err != nil {
		return plannerResult{}, err
	}
	// Gate the target: a real directory (a symlinked or absent target is
	// structural). Non-git is allowed — embark has no git requirement.
	if !fsutil.IsRealDir(targetAbs) {
		return plannerResult{}, fmt.Errorf("target %s is not a directory", filepath.Base(targetAbs))
	}

	root, err := os.OpenRoot(lifeboatAbs)
	if err != nil {
		return plannerResult{}, err
	}
	defer root.Close()

	rels, err := walkLifeboatFiles(root)
	if err != nil {
		return plannerResult{}, err
	}

	var (
		planned   []PlannedEmbark
		conflicts []Conflict
		ignored   []IgnoredFile
	)
	// One lifeboat file per embark target: a second claimant (a bucket-less
	// intent alongside its bucketed twin) would be a silent last-writer-wins
	// overwrite, so it surfaces as a conflict instead.
	claimed := map[string]string{}
	for _, rel := range rels {
		family, targetRel, disp, detail := resolveTarget(rel)
		switch disp {
		case dispReportOnly:
			ignored = append(ignored, IgnoredFile{LifeboatPath: rel, Reason: IgnoredReportOnly})
		case dispUnmapped:
			ignored = append(ignored, IgnoredFile{LifeboatPath: rel, Reason: IgnoredUnmapped, Detail: detail})
		case dispUnknown:
			ignored = append(ignored, IgnoredFile{LifeboatPath: rel, Reason: IgnoredUnknown})
		case dispPlanned:
			data, err := readLifeboatFile(root, lifeboatAbs, rel)
			if err != nil {
				return plannerResult{}, err
			}
			// Defence in depth: the table already guarantees a safe targetRel, but
			// re-validate before it becomes a write path.
			if !validRelPath(targetRel) {
				return plannerResult{}, fmt.Errorf("embark: refusing unsafe target path %q", targetRel)
			}
			if first, dup := claimed[targetRel]; dup {
				conflicts = append(conflicts, Conflict{
					Path:         targetRel,
					LifeboatPath: rel,
					Kind:         ConflictDuplicateTarget,
					Detail:       "also mapped from " + first,
				})
				continue
			}
			claimed[targetRel] = rel
			pe, cf := classifyEmbark(targetAbs, rel, targetRel, family, data)
			if cf != nil {
				conflicts = append(conflicts, *cf)
			} else {
				planned = append(planned, pe)
			}
		}
	}

	return plannerResult{
		lifeboatAbs: lifeboatAbs,
		targetAbs:   targetAbs,
		prov:        prov,
		planned:     planned,
		conflicts:   conflicts,
		ignored:     ignored,
		coverage:    readCoverageHandoff(lifeboatAbs),
	}, nil
}

// disposition is where resolveTarget routes one lifeboat file.
type disposition int

const (
	dispPlanned disposition = iota
	dispReportOnly
	dispUnmapped
	dispUnknown
)

// resolveTarget maps a lifeboat-relative path to a target write via the
// embarkFamilies inverse table (the single source of truth). A file under a
// known family root that cannot resolve (unknown bucket, unsafe leaf, or
// bucket-less where no default exists) is Unmapped; a report-only path is
// ReportOnly; anything else is Unknown. A family is checked before the
// report-only prefixes so an unresolvable in-family file is Unmapped, not
// Unknown. detail is a short human note for the Unmapped case.
func resolveTarget(rel string) (family, targetRel string, disp disposition, detail string) {
	for _, f := range embarkFamilies {
		if !strings.HasPrefix(rel, f.LifeboatPrefix) {
			continue
		}
		rest := rel[len(f.LifeboatPrefix):]
		if f.Buckets == nil {
			// Flat family (adrs): <prefix><leaf>.
			leaf := rest
			if strings.Contains(leaf, "/") || safeLeaf(leaf) == "" {
				return f.Name, "", dispUnmapped, "unresolvable path under " + f.Name
			}
			return f.Name, f.TargetPrefix + leaf, dispPlanned, ""
		}
		// Bucketed family.
		seg := strings.SplitN(rest, "/", 2)
		if len(seg) == 1 {
			// Bucket-less file directly under the family root.
			if f.DefaultBucket == "" {
				return f.Name, "", dispUnmapped, "bucket-less " + f.Name + " has no default bucket"
			}
			leaf := rest
			if safeLeaf(leaf) == "" {
				return f.Name, "", dispUnmapped, "unsafe leaf under " + f.Name
			}
			return f.Name, f.TargetPrefix + f.DefaultBucket + "/" + leaf, dispPlanned, ""
		}
		bucket, leaf := seg[0], seg[1]
		if !containsStr(f.Buckets, bucket) {
			return f.Name, "", dispUnmapped, "unknown " + f.Name + " bucket " + sanitize(bucket)
		}
		if strings.Contains(leaf, "/") || safeLeaf(leaf) == "" {
			return f.Name, "", dispUnmapped, "unsafe leaf under " + f.Name
		}
		return f.Name, f.TargetPrefix + bucket + "/" + leaf, dispPlanned, ""
	}
	for _, p := range reportOnlyPrefixes {
		if strings.HasPrefix(rel, p) {
			return "", "", dispReportOnly, ""
		}
	}
	return "", "", dispUnknown, ""
}

// classifyEmbark decides one planned write against the current target, lexical +
// lstat with NO symlink follow. A parent component that is a file or symlink is a
// parent-not-dir conflict; a target that exists and is not a regular file is a
// target-not-regular conflict; a regular file with equal bytes is an idempotent
// unchanged skip; differing bytes is an exists-differs conflict; an absent target
// is a create.
func classifyEmbark(targetAbs, lifeboatRel, targetRel, family string, content []byte) (PlannedEmbark, *Conflict) {
	if cf := checkParents(targetAbs, targetRel, lifeboatRel); cf != nil {
		return PlannedEmbark{}, cf
	}
	full := filepath.Join(targetAbs, filepath.FromSlash(targetRel))
	fi, err := os.Lstat(full)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return PlannedEmbark{
				LifeboatPath: lifeboatRel, TargetPath: targetRel, Family: family,
				Bytes: len(content), Action: ActionCreate, Content: content,
			}, nil
		}
		return PlannedEmbark{}, &Conflict{
			Path: targetRel, LifeboatPath: lifeboatRel, Kind: ConflictTargetNotRegular,
			Detail: "cannot inspect the target path",
		}
	}
	if fi.Mode()&os.ModeSymlink != 0 || !fi.Mode().IsRegular() {
		return PlannedEmbark{}, &Conflict{
			Path: targetRel, LifeboatPath: lifeboatRel, Kind: ConflictTargetNotRegular,
			Detail: "target exists and is not a regular file",
		}
	}
	existing, err := os.ReadFile(full)
	if err != nil {
		return PlannedEmbark{}, &Conflict{
			Path: targetRel, LifeboatPath: lifeboatRel, Kind: ConflictTargetNotRegular,
			Detail: "cannot read the target file",
		}
	}
	if bytes.Equal(existing, content) {
		return PlannedEmbark{
			LifeboatPath: lifeboatRel, TargetPath: targetRel, Family: family,
			Bytes: len(content), Action: ActionUnchanged, Content: content,
		}, nil
	}
	return PlannedEmbark{}, &Conflict{
		Path: targetRel, LifeboatPath: lifeboatRel, Kind: ConflictExistsDiffers,
		Detail: "target file differs from the lifeboat copy",
	}
}

// checkParents walks the existing parent components of targetRel under targetAbs
// and returns a parent-not-dir conflict for the first that is a symlink or a
// non-directory. A component that does not yet exist ends the walk — it (and
// every deeper one) is created by the contained MkdirAll on the write path.
func checkParents(targetAbs, targetRel, lifeboatRel string) *Conflict {
	dir := path.Dir(targetRel)
	if dir == "." {
		return nil
	}
	cur := targetAbs
	for _, seg := range strings.Split(dir, "/") {
		cur = filepath.Join(cur, seg)
		fi, err := os.Lstat(cur)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				return nil
			}
			return &Conflict{Path: targetRel, LifeboatPath: lifeboatRel, Kind: ConflictParentNotDir, Detail: "cannot inspect a parent directory"}
		}
		if fi.Mode()&os.ModeSymlink != 0 {
			return &Conflict{Path: targetRel, LifeboatPath: lifeboatRel, Kind: ConflictParentNotDir, Detail: "a parent path is a symlink"}
		}
		if !fi.IsDir() {
			return &Conflict{Path: targetRel, LifeboatPath: lifeboatRel, Kind: ConflictParentNotDir, Detail: "a parent path is not a directory"}
		}
	}
	return nil
}

// writeEmbark performs the no-conflict write set through the two-layer idiom
// (os.Root containment + independent lexical validation + the canonical
// fsutil.WriteFileAtomic, reusing writeIntoLifeboat). Per-file atomic; the SET is
// not transactional, which is acceptable — the conflict gate ran first, unchanged
// files are skipped, and a re-run is idempotent, so a partial write from an I/O
// fault re-completes on the next embark.
func writeEmbark(targetAbs string, planned []PlannedEmbark) (written, unchanged, bytesWritten int, families map[string]int, err error) {
	families = map[string]int{}
	root, err := os.OpenRoot(targetAbs)
	if err != nil {
		return 0, 0, 0, nil, err
	}
	defer root.Close()
	for _, p := range planned {
		if p.Action == ActionUnchanged {
			unchanged++
			continue
		}
		if !validRelPath(p.TargetPath) {
			return 0, 0, 0, nil, fmt.Errorf("refusing unsafe target path %q", p.TargetPath)
		}
		if err := writeIntoLifeboat(root, targetAbs, p.TargetPath, p.Content); err != nil {
			return 0, 0, 0, nil, err
		}
		written++
		bytesWritten += len(p.Content)
		families[p.Family]++
	}
	return written, unchanged, bytesWritten, families, nil
}

// embarkMarker predicts (dryRun) or performs the CURRENT abcd marker block in the
// target CLAUDE.md via ahoy.EnsureMarker, then maps the outcome to a MarkerAction.
// The install/refresh distinction is derived from whether the file already
// carried a block (via the exported ahoy.StripMarkerBlock), since EnsureMarker's
// signature reports only whether it changed. A symlinked/unwritable CLAUDE.md is
// MarkerActionSkip (non-fatal — the records still land).
func embarkMarker(targetAbs string, dryRun bool) MarkerResult {
	p := filepath.Join(targetAbs, "CLAUDE.md")
	hadBlock := false
	if data, err := os.ReadFile(p); err == nil {
		if _, had := ahoy.StripMarkerBlock(data); had {
			hadBlock = true
		}
	}
	changed, err := ahoy.EnsureMarker(p, dryRun)
	res := MarkerResult{Target: "CLAUDE.md", Changed: changed}
	switch {
	case err != nil:
		res.Action = MarkerActionSkip
		res.Changed = false
		res.Note = sanitize(err.Error())
	case !changed:
		res.Action = MarkerActionCurrent
	case hadBlock:
		res.Action = MarkerActionRefresh
	default:
		res.Action = MarkerActionInstall
	}
	return res
}

// readCoverageHandoff reads the lifeboat's coverage.json (bounded) and pulls the
// blanks-first payload. Absent → Present:false; present-but-unparseable →
// Degraded (a note, never fatal). Only blank sections become BlankPrompts, each
// sanitised before it can reach a terminal.
func readCoverageHandoff(abs string) *CoverageHandoff {
	p := filepath.Join(abs, "coverage.json")
	fi, err := os.Lstat(p)
	if err != nil {
		return &CoverageHandoff{Present: false}
	}
	if fi.Mode()&os.ModeSymlink != 0 || !fi.Mode().IsRegular() || fi.Size() > maxEmbarkFileBytes {
		return &CoverageHandoff{Present: true, Degraded: true, Note: "coverage.json is not a readable regular file"}
	}
	data, err := os.ReadFile(p)
	if err != nil {
		return &CoverageHandoff{Present: true, Degraded: true, Note: "coverage.json could not be read"}
	}
	var cov Coverage
	if err := json.Unmarshal(data, &cov); err != nil {
		return &CoverageHandoff{Present: true, Degraded: true, Note: "coverage.json is present but could not be parsed"}
	}
	h := &CoverageHandoff{Present: true, Summary: cov.Summary}
	for _, s := range cov.Sections {
		if s.Status != StatusBlank {
			continue
		}
		h.Blanks = append(h.Blanks, BlankPrompt{
			Section:  s.Name,
			Kind:     s.Kind,
			Question: sanitize(s.Question),
			Searched: sanitizeAll(s.Searched),
		})
	}
	return h
}

// walkLifeboatFiles returns every regular file's lifeboat-relative POSIX path,
// sorted, after enforcing the trust boundary through the containment root: no
// symlink anywhere (a symlinked entry is fatal, not followed), every path
// validRelPath, and the file count under maxEmbarkFiles. It reads no content.
func walkLifeboatFiles(root *os.Root) ([]string, error) {
	var rels []string
	count := 0
	err := fs.WalkDir(root.FS(), ".", func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if p == "." {
			return nil
		}
		if d.Type()&fs.ModeSymlink != 0 {
			return fmt.Errorf("lifeboat contains a symlink at %q (refusing to follow)", p)
		}
		if d.IsDir() {
			return nil
		}
		if !d.Type().IsRegular() {
			return fmt.Errorf("lifeboat contains a non-regular file at %q", p)
		}
		if !validRelPath(p) {
			return fmt.Errorf("lifeboat contains an unsafe path %q", p)
		}
		count++
		if count > maxEmbarkFiles {
			return fmt.Errorf("lifeboat exceeds the %d-file cap", maxEmbarkFiles)
		}
		rels = append(rels, p)
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Strings(rels)
	return rels, nil
}

// readLifeboatFile reads one lifeboat file through the containment root behind
// the trust guards: no symlink, regular file, size under maxEmbarkFileBytes.
func readLifeboatFile(root *os.Root, abs, rel string) ([]byte, error) {
	fi, err := root.Lstat(rel)
	if err != nil {
		return nil, err
	}
	if fi.Mode()&os.ModeSymlink != 0 {
		return nil, fmt.Errorf("lifeboat file %q is a symlink (refusing to follow)", rel)
	}
	if !fi.Mode().IsRegular() {
		return nil, fmt.Errorf("lifeboat file %q is not a regular file", rel)
	}
	if fi.Size() > maxEmbarkFileBytes {
		return nil, fmt.Errorf("lifeboat file %q exceeds the %d-byte cap", rel, maxEmbarkFileBytes)
	}
	f, err := root.Open(rel)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	data, err := io.ReadAll(io.LimitReader(f, maxEmbarkFileBytes+1))
	if err != nil {
		return nil, err
	}
	if len(data) > maxEmbarkFileBytes {
		return nil, fmt.Errorf("lifeboat file %q exceeds the %d-byte cap", rel, maxEmbarkFileBytes)
	}
	return data, nil
}

// readProvenance reads and parses the lifeboat's _provenance.json behind the
// trust guards (no symlink, regular file, bounded). It is the manifest header and
// the schema/name/hash source.
func readProvenance(abs string) (Provenance, error) {
	var prov Provenance
	p := filepath.Join(abs, ProvenanceName)
	fi, err := os.Lstat(p)
	if err != nil {
		return prov, fmt.Errorf("lifeboat: reading %s: %w", ProvenanceName, err)
	}
	if fi.Mode()&os.ModeSymlink != 0 {
		return prov, fmt.Errorf("lifeboat: %s is a symlink (refusing to follow)", ProvenanceName)
	}
	if !fi.Mode().IsRegular() {
		return prov, fmt.Errorf("lifeboat: %s is not a regular file", ProvenanceName)
	}
	if fi.Size() > maxProvenanceBytes {
		return prov, fmt.Errorf("lifeboat: %s exceeds the %d-byte cap", ProvenanceName, maxProvenanceBytes)
	}
	data, err := os.ReadFile(p)
	if err != nil {
		return prov, fmt.Errorf("lifeboat: reading %s: %w", ProvenanceName, err)
	}
	if err := json.Unmarshal(data, &prov); err != nil {
		return prov, fmt.Errorf("lifeboat: %s is not valid JSON: %w", ProvenanceName, err)
	}
	return prov, nil
}

// isManifestExcluded reports whether a lifeboat path was left out of
// manifest_sha256 (the header and the post-pack layer-3 interpretation), so
// VerifyManifest reproduces the pinned hash exactly.
func isManifestExcluded(rel string) bool {
	for _, e := range manifestExcludedExact {
		if rel == e {
			return true
		}
	}
	for _, p := range manifestExcludedPrefixes {
		if strings.HasPrefix(rel, p) {
			return true
		}
	}
	return false
}

// containsStr reports set membership for a small string slice.
func containsStr(set []string, v string) bool {
	for _, s := range set {
		if s == v {
			return true
		}
	}
	return false
}
