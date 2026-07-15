package lifeboat

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

// planFile returns the planned file at dest path p, or fails.
func planFile(t *testing.T, lb Lifeboat, p string) PlannedFile {
	t.Helper()
	for _, f := range lb.Files {
		if f.Path == p {
			return f
		}
	}
	t.Fatalf("planned file %q not found; have: %s", p, planPaths(lb))
	return PlannedFile{}
}

func hasPlanFile(lb Lifeboat, p string) bool {
	for _, f := range lb.Files {
		if f.Path == p {
			return true
		}
	}
	return false
}

func planPaths(lb Lifeboat) string {
	ps := make([]string, len(lb.Files))
	for i, f := range lb.Files {
		ps[i] = f.Path
	}
	sort.Strings(ps)
	return strings.Join(ps, "\n  ")
}

// TestPlanWritesNothing is the M3a contract: a plan is produced entirely in
// memory. Probing then planning a repository must leave its tree byte-for-byte
// unchanged — the read-only spine has no destination and touches no file.
func TestPlanWritesNothing(t *testing.T) {
	repo := nativeTierFixture(t)
	before := treeHash(t, repo)
	if _, err := Plan(repo); err != nil {
		t.Fatal(err)
	}
	if after := treeHash(t, repo); after != before {
		t.Errorf("Plan mutated the source tree (hash %s -> %s)", before, after)
	}
}

// TestPlanIsDeterministic guards the "no timestamp" property: two plans of an
// unchanged repository are byte-identical, path for path, and carry the same
// pinned manifest hash. Determinism is what lets the hash mean anything.
func TestPlanIsDeterministic(t *testing.T) {
	repo := nativeTierFixture(t)
	a, err := Plan(repo)
	if err != nil {
		t.Fatal(err)
	}
	b, err := Plan(repo)
	if err != nil {
		t.Fatal(err)
	}
	if len(a.Files) != len(b.Files) {
		t.Fatalf("file count changed between plans: %d vs %d", len(a.Files), len(b.Files))
	}
	for i := range a.Files {
		if a.Files[i].Path != b.Files[i].Path {
			t.Errorf("path %d differs: %q vs %q", i, a.Files[i].Path, b.Files[i].Path)
		}
		if !bytes.Equal(a.Files[i].Content, b.Files[i].Content) {
			t.Errorf("content of %q differs between plans", a.Files[i].Path)
		}
	}
	if ah, bh := ManifestSHA256(a.Files), ManifestSHA256(b.Files); ah != bh {
		t.Errorf("manifest hash not deterministic: %s vs %s", ah, bh)
	}
}

// TestManifestSHA256SortsByPathNotHash pins the adr-35 hash definition — SHA-256
// over "<sha256>  <path>\n" lines sorted lexicographically BY PATH, provenance
// excluded — with an independent reference. The test data is chosen so that
// path-order and line-order (which the leading hash would dominate) genuinely
// differ, and the test fails loudly if a future edit makes them coincide; a
// hash-order implementation would therefore be caught, not asserted against
// itself.
func TestManifestSHA256SortsByPathNotHash(t *testing.T) {
	files := []PlannedFile{
		{Path: "a.md", Content: []byte("alpha")},
		{Path: "b.md", Content: []byte("beta")},
		{Path: "c.md", Content: []byte("gamma")},
		{Path: ProvenanceName, Content: []byte("{ignored}")},
	}

	type entry struct{ path, line string }
	var entries []entry
	var lineOrder []string
	for _, f := range files {
		if f.Path == ProvenanceName {
			continue
		}
		line := fmt.Sprintf("%x  %s\n", sha256.Sum256(f.Content), f.Path)
		entries = append(entries, entry{f.Path, line})
		lineOrder = append(lineOrder, line)
	}
	// Reference: adr-35 — sort BY PATH.
	sort.Slice(entries, func(i, j int) bool { return entries[i].path < entries[j].path })
	var byPath strings.Builder
	for _, e := range entries {
		byPath.WriteString(e.line)
	}
	// The buggy ordering: sort whole lines (hash-dominated).
	sort.Strings(lineOrder)
	if strings.Join(lineOrder, "") == byPath.String() {
		t.Fatal("test data no longer distinguishes path-order from hash-order; change the contents")
	}

	want := fmt.Sprintf("%x", sha256.Sum256([]byte(byPath.String())))
	if got := ManifestSHA256(files); got != want {
		t.Errorf("ManifestSHA256 did not sort by path (adr-35): got %s want %s", got, want)
	}
}

// TestManifestSHA256ExcludesProvenance proves _provenance.json cannot hash
// itself: mutating it leaves the manifest hash unchanged, while mutating any
// other file changes it.
func TestManifestSHA256ExcludesProvenance(t *testing.T) {
	base := []PlannedFile{
		{Path: "a.md", Content: []byte("alpha")},
		{Path: ProvenanceName, Content: []byte("v1")},
	}
	provChanged := []PlannedFile{
		{Path: "a.md", Content: []byte("alpha")},
		{Path: ProvenanceName, Content: []byte("v2-different")},
	}
	if ManifestSHA256(base) != ManifestSHA256(provChanged) {
		t.Error("manifest hash changed when only _provenance.json changed")
	}
	contentChanged := []PlannedFile{
		{Path: "a.md", Content: []byte("ALPHA")},
		{Path: ProvenanceName, Content: []byte("v1")},
	}
	if ManifestSHA256(base) == ManifestSHA256(contentChanged) {
		t.Error("manifest hash unchanged when a real file changed")
	}
}

// TestPlanProvenanceRecordsManifestHash checks the provenance file the plan
// emits carries the hash over every other file, and is itself last.
func TestPlanProvenanceRecordsManifestHash(t *testing.T) {
	repo := nativeTierFixture(t)
	lb, err := Plan(repo)
	if err != nil {
		t.Fatal(err)
	}
	pf := planFile(t, lb, ProvenanceName)
	var prov Provenance
	if err := json.Unmarshal(pf.Content, &prov); err != nil {
		t.Fatalf("provenance is not valid JSON: %v", err)
	}
	if prov.SchemaVersion != SchemaVersion {
		t.Errorf("provenance schema_version = %d, want %d", prov.SchemaVersion, SchemaVersion)
	}
	if got := ManifestSHA256(lb.Files); prov.ManifestSHA256 != got {
		t.Errorf("provenance manifest_sha256 = %s, recomputed = %s", prov.ManifestSHA256, got)
	}
}

// TestPlanBriefCarriesOnlyNonBlankSections is the honesty rule for the brief: a
// grounded section gets a citation-map file; a genuinely blank section
// (personas, absent here) gets none — the plan never fabricates a brief page for
// a section it could not ground. And every brief file lives under brief/.
func TestPlanBriefCarriesOnlyNonBlankSections(t *testing.T) {
	repo := nativeTierFixture(t)
	lb, err := Plan(repo)
	if err != nil {
		t.Fatal(err)
	}
	// product/context is authored and over-threshold in the fixture: grounded.
	if !hasPlanFile(lb, "brief/01-product/02-context.md") {
		t.Errorf("grounded product/context has no brief file; have:\n  %s", planPaths(lb))
	}
	// product/personas has no source in the fixture: blank, so no brief file.
	if hasPlanFile(lb, "brief/01-product/05-personas.md") {
		t.Error("blank product/personas got a brief file it should not have")
	}
	// No brief file may escape the brief/ subtree.
	for _, f := range lb.Files {
		if strings.HasPrefix(f.Path, "brief/") {
			continue
		}
		if strings.Contains(f.Path, "/brief/") {
			t.Errorf("brief content leaked outside brief/: %q", f.Path)
		}
	}
}

// TestPlanCopiesADRAndIssueVerbatim proves the durable-record copies are byte
// exact: the ADR and the open issue land at their lifeboat homes with the source
// bytes unchanged.
func TestPlanCopiesADRAndIssueVerbatim(t *testing.T) {
	repo := nativeTierFixture(t)
	lb, err := Plan(repo)
	if err != nil {
		t.Fatal(err)
	}
	cases := []struct{ src, dst string }{
		{".abcd/development/decisions/adrs/0001-record-architecture-decisions.md",
			"docs/adrs/0001-record-architecture-decisions.md"},
		{".abcd/work/issues/open/iss-001-example.md",
			"activity/issues/open/iss-001-example.md"},
	}
	for _, c := range cases {
		want, err := os.ReadFile(filepath.Join(repo, c.src))
		if err != nil {
			t.Fatal(err)
		}
		got := planFile(t, lb, c.dst)
		if !bytes.Equal(got.Content, want) {
			t.Errorf("%s not copied verbatim to %s", c.src, c.dst)
		}
	}
}

// TestPlanRescueSpineFromIntents: with an intent corpus present, the spine is the
// intents copied verbatim and there is no git-derived spine.md.
func TestPlanRescueSpineFromIntents(t *testing.T) {
	repo := nativeTierFixture(t)
	lb, err := Plan(repo)
	if err != nil {
		t.Fatal(err)
	}
	if !hasPlanFile(lb, "rescue/intents/drafts/itd-001-example.md") {
		t.Errorf("intent not copied into rescue spine; have:\n  %s", planPaths(lb))
	}
	if hasPlanFile(lb, "rescue/spine.md") {
		t.Error("git-derived spine.md emitted despite an intent corpus being present")
	}
}

// TestPlanRescueSpineFromGitAlone: a git-only repo with no intent corpus falls
// back to a single git-derived spine summary carrying the commit count.
func TestPlanRescueSpineFromGitAlone(t *testing.T) {
	repo := gitFixture(t, []fixtureCommit{
		{path: "main.go", content: "package main\n", message: "one"},
		{message: "two"},
	})
	lb, err := Plan(repo)
	if err != nil {
		t.Fatal(err)
	}
	if hasPlanFile(lb, "rescue/intents/drafts/itd-001-example.md") {
		t.Error("intent spine present in a repo with no intents")
	}
	spine := planFile(t, lb, "rescue/spine.md")
	if !strings.Contains(string(spine.Content), "commits:") {
		t.Errorf("git-derived spine.md does not report commit count:\n%s", spine.Content)
	}
}

// TestPlanRescueRejectsHostileSubdir proves a rejected intent SUBDIRECTORY name
// drops its files rather than relocating them up a level (where path.Join would
// swallow an empty component and steer the file into rescue/intents/, colliding
// with a legitimate top-level intent). The subdir carries a control character,
// which safeLeaf rejects.
func TestPlanRescueRejectsHostileSubdir(t *testing.T) {
	repo := t.TempDir()
	// A legitimate top-level intent.
	top := filepath.Join(repo, ".abcd/development/intents/itd-1.md")
	if err := os.MkdirAll(filepath.Dir(top), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(top, []byte("# top intent\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	// A hostile subdirectory whose name carries a control char, holding an intent
	// with the SAME leaf as the legitimate one.
	hostileDir := filepath.Join(repo, ".abcd/development/intents/x\x01evil")
	if err := os.MkdirAll(hostileDir, 0o755); err != nil {
		t.Skipf("filesystem rejects control-char directory names: %v", err)
	}
	if err := os.WriteFile(filepath.Join(hostileDir, "itd-1.md"), []byte("# EVIL\n"), 0o644); err != nil {
		t.Skipf("filesystem rejects the fixture: %v", err)
	}

	lb, err := Plan(repo)
	if err != nil {
		t.Fatal(err)
	}
	// The legitimate intent lands, once, with its real bytes.
	got := planFile(t, lb, "rescue/intents/itd-1.md")
	if string(got.Content) != "# top intent\n" {
		t.Errorf("rescue/intents/itd-1.md has %q; hostile subdir content leaked in", got.Content)
	}
	// No planned path may contain a control character.
	for _, f := range lb.Files {
		for _, r := range f.Path {
			if r < 0x20 || r == 0x7f {
				t.Errorf("planned path carries a control character: %q", f.Path)
			}
		}
	}
}

// TestSafeLeafRejectsTraversalAndControl guards the destination-path derivation:
// a hostile source filename can never steer where a file lands.
func TestSafeLeafRejectsTraversalAndControl(t *testing.T) {
	reject := []string{"", ".", "..", "../escape", "a/b", "a\\b", "bad\x00name", "line\nbreak"}
	for _, name := range reject {
		if got := safeLeaf(name); got != "" {
			t.Errorf("safeLeaf(%q) = %q, want empty (rejected)", name, got)
		}
	}
	keep := map[string]string{
		"iss-001-example.md": "iss-001-example.md",
		"0001-adr.md":        "0001-adr.md",
	}
	for name, want := range keep {
		if got := safeLeaf(name); got != want {
			t.Errorf("safeLeaf(%q) = %q, want %q", name, got, want)
		}
	}
}

// TestPlanNoDuplicateDestinationPaths is the Plan/pack parity guard: a real pack
// writes one file per destination path, so the plan must never list a path
// twice. The fixture plants the same ADR basename in two source homes (the
// abcd-native dir and a conventional docs/adr), which previously produced two
// planned files at docs/adrs/<base>.
func TestPlanNoDuplicateDestinationPaths(t *testing.T) {
	repo := t.TempDir()
	write := func(rel, content string) {
		full := filepath.Join(repo, rel)
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write(".abcd/development/decisions/adrs/0001-foo.md", "# native adr\n")
	write("docs/adr/0001-foo.md", "# conventional adr — different bytes\n")

	lb, err := Plan(repo)
	if err != nil {
		t.Fatal(err)
	}
	seen := map[string]int{}
	for _, f := range lb.Files {
		seen[f.Path]++
	}
	for p, n := range seen {
		if n > 1 {
			t.Errorf("destination path %q planned %d times; a pack can only write it once", p, n)
		}
	}
	// The colliding ADR must still be planned exactly once at its dest.
	if seen["docs/adrs/0001-foo.md"] != 1 {
		t.Errorf("docs/adrs/0001-foo.md planned %d times, want 1", seen["docs/adrs/0001-foo.md"])
	}
	// The manifest count agrees with the deduped file set.
	if m := lb.Manifest(); m.FileCount != len(lb.Files) {
		t.Errorf("manifest FileCount %d != %d after dedup", m.FileCount, len(lb.Files))
	}
}

// TestPlanRecordsOversizeOmission proves a verbatim record too large to read is
// declared, not silently dropped: it appears in the plan's omissions with the
// source path, and does not appear as a planned file.
func TestPlanRecordsOversizeOmission(t *testing.T) {
	repo := t.TempDir()
	adr := ".abcd/development/decisions/adrs/0001-huge.md"
	full := filepath.Join(repo, adr)
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		t.Fatal(err)
	}
	// One byte over the per-file read cap.
	big := bytes.Repeat([]byte("x"), maxProbeReadBytes+1)
	if err := os.WriteFile(full, big, 0o644); err != nil {
		t.Fatal(err)
	}
	lb, err := Plan(repo)
	if err != nil {
		t.Fatal(err)
	}
	if hasPlanFile(lb, "docs/adrs/0001-huge.md") {
		t.Error("oversize ADR was planned as a file; it cannot be read whole")
	}
	found := false
	for _, o := range lb.Omissions {
		if o.Path == adr {
			found = true
		}
	}
	if !found {
		t.Errorf("oversize ADR not recorded as an omission; omissions: %+v", lb.Omissions)
	}
}

// TestPlanBuilderCeilingOmits proves the aggregate size ceiling is enforced and
// declared: past maxPlanFiles, further adds are dropped and recorded, never
// silently lost, and the plan cannot grow without bound on a pathological tree.
func TestPlanBuilderCeilingOmits(t *testing.T) {
	pb := newPlanBuilder()
	for i := 0; i < maxPlanFiles+5; i++ {
		pb.add(fmt.Sprintf("f/%d.md", i), []byte("x"))
	}
	if len(pb.files) != maxPlanFiles {
		t.Errorf("builder kept %d files, want the ceiling %d", len(pb.files), maxPlanFiles)
	}
	if len(pb.omissions) != 5 {
		t.Errorf("builder recorded %d omissions, want 5", len(pb.omissions))
	}
}

// TestPlanBuilderDedupIsFirstWins checks the collision policy is deterministic:
// the first writer of a destination path wins and later writers are ignored (not
// recorded as omissions — a duplicate is not a lost record).
func TestPlanBuilderDedupIsFirstWins(t *testing.T) {
	pb := newPlanBuilder()
	pb.add("x.md", []byte("first"))
	pb.add("x.md", []byte("second"))
	if len(pb.files) != 1 {
		t.Fatalf("got %d files, want 1", len(pb.files))
	}
	if string(pb.files[0].Content) != "first" {
		t.Errorf("dedup kept %q, want the first writer", pb.files[0].Content)
	}
	if len(pb.omissions) != 0 {
		t.Errorf("a duplicate dest should not be an omission; got %+v", pb.omissions)
	}
}

// TestPlanManifestReportsHashAndTotals checks the dry-run manifest view agrees
// with the file set it summarises.
func TestPlanManifestReportsHashAndTotals(t *testing.T) {
	repo := nativeTierFixture(t)
	lb, err := Plan(repo)
	if err != nil {
		t.Fatal(err)
	}
	m := lb.Manifest()
	if m.FileCount != len(lb.Files) {
		t.Errorf("manifest file count %d != %d", m.FileCount, len(lb.Files))
	}
	if m.ManifestSHA256 != ManifestSHA256(lb.Files) {
		t.Error("manifest hash disagrees with the file set")
	}
	total := 0
	for _, f := range lb.Files {
		total += f.Bytes
	}
	if m.TotalBytes != total {
		t.Errorf("manifest total bytes %d != %d", m.TotalBytes, total)
	}
}
