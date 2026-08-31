package reading

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/intentdriven/abcd/internal/gittest"
)

// narrowPresets is a preset file that genuinely narrows, so a test can tell a
// scope that works from one that selects everything by accident.
//
// It carries the fixture's all-selecting preset too, because the narrowing
// tests need an UNSCOPED baseline to compare against and that baseline is
// itself expressed as a scope. Writing this file over the fixture's own would
// otherwise remove the very preset the baseline is measured with.
func narrowPresets() string {
	return `{
  "schema_version": 1,
  "presets": {
    "` + fixtureScopeName + `": {
      "positions": {
        "widening": {"kinds": ["brief-section", "glossary-term", "intent-projection", "discipline", "spec", "source", "test", "doc", "config"], "records": [], "paths": []},
        "entailment": {"kinds": ["brief-section", "glossary-term", "intent-projection", "discipline", "spec", "source", "test", "doc", "config"], "records": [], "paths": []},
        "detection": {"kinds": ["brief-section", "glossary-term", "intent-projection", "discipline", "spec", "source", "test", "doc", "config"], "records": [], "paths": []}
      }
    },
    "cold": {
      "positions": {
        "widening": {"kinds": ["brief-section"], "records": [], "paths": []},
        "entailment": {"kinds": ["intent-projection"], "records": [], "paths": []},
        "detection": {"kinds": ["spec"], "records": [], "paths": []}
      }
    },
    "warm": {
      "extends": "cold",
      "positions": {
        "widening": {"kinds": ["doc"], "records": [], "paths": []},
        "entailment": {"kinds": ["spec"], "records": [], "paths": []},
        "detection": {"kinds": ["doc"], "records": [], "paths": []}
      }
    }
  }
}
`
}

// TestScopeIsRequired is itd-199 ac-1.
func TestScopeIsRequired(t *testing.T) {
	root := fixtureRepo(t)
	_, err := Assemble(AssembleRequest{
		RepoRoot: root, Position: PositionWidening, Target: "HEAD", DryRun: true,
	})
	if err == nil {
		t.Fatal("an assembly with no scope was accepted")
	}
	for _, want := range []string{"record id", "kind", "preset"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("the refusal does not name %q: %v", want, err)
		}
	}
}

// TestScopeNarrowsNeverWidens is the property the whole mechanism rests on: a
// scope intersects what the position already admits and can only narrow it.
//
// Proved against the unscoped set rather than against an expected list, so a
// scope that somehow reached a row the table denies fails here rather than
// quietly redefining what the test expects.
func TestScopeNarrowsNeverWidens(t *testing.T) {
	root := fixtureRepo(t)
	writeFile(t, root, ".abcd/config/reading-presets.json", narrowPresets())
	gitCommitAll(t, root)

	for _, p := range AssemblingPositions() {
		wide := map[string]bool{}
		for _, m := range assembleFixture(t, root, p).Manifest.Items {
			wide[m.Path] = true
		}
		for _, preset := range []string{"cold", "warm"} {
			res, err := Assemble(AssembleRequest{
				RepoRoot: root, Position: p, Target: "HEAD", Scope: preset, DryRun: true,
			})
			if err != nil {
				continue // a preset selecting nothing at this position refuses, which is its own test
			}
			for _, m := range res.Manifest.Items {
				if !wide[m.Path] {
					t.Errorf("position %s preset %s admitted %s, which the unscoped assembly did not: "+
						"a scope must intersect the table's admission and only narrow it", p, preset, m.Path)
				}
			}
			if len(res.Manifest.Items) >= len(wide) {
				t.Errorf("position %s preset %s did not narrow at all (%d of %d items); the fixture "+
					"is not exercising the filter", p, preset, len(res.Manifest.Items), len(wide))
			}
		}
	}
}

// TestWarmContainsCold is itd-199 ac-11, and it is the reason `extends` is a
// union rather than a replacement: warm can never be narrower than cold, and a
// scope added to cold appears in warm without anyone remembering to add it
// twice.
func TestWarmContainsCold(t *testing.T) {
	root := fixtureRepo(t)
	writeFile(t, root, ".abcd/config/reading-presets.json", narrowPresets())
	pf, err := LoadPresets(root)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	for _, p := range AssemblingPositions() {
		cold := presetSelectors(pf, pf.Presets["cold"], p)
		warm := presetSelectors(pf, pf.Presets["warm"], p)
		have := make(map[Selector]bool, len(warm))
		for _, s := range warm {
			have[s] = true
		}
		for _, s := range cold {
			if !have[s] {
				t.Errorf("position %s: warm does not contain cold's selector %+v", p, s)
			}
		}
		if len(warm) <= len(cold) {
			t.Errorf("position %s: warm (%d) is not wider than cold (%d); the fixture does not "+
				"exercise the delta", p, len(warm), len(cold))
		}
	}
}

// TestExtendsIsAUnionNotAReplacement proves the guard above by MUTATION: a
// warm entry that REPLACED cold's at a position would still pass a test that
// only asserted warm is non-empty.
func TestExtendsIsAUnionNotAReplacement(t *testing.T) {
	root := fixtureRepo(t)
	writeFile(t, root, ".abcd/config/reading-presets.json", narrowPresets())
	pf, err := LoadPresets(root)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	warm := presetSelectors(pf, pf.Presets["warm"], PositionWidening)

	// warm names doc at widening and cold names brief-section. A replacement
	// would yield doc alone.
	var sawCold, sawWarm bool
	for _, s := range warm {
		if s.Kind == KindBriefSection {
			sawCold = true
		}
		if s.Kind == KindDoc {
			sawWarm = true
		}
	}
	if !sawCold {
		t.Error("warm dropped cold's selector at widening; extends must union, never replace")
	}
	if !sawWarm {
		t.Error("warm dropped its own selector at widening")
	}
}

// TestPresetNameIsNotAnOverrideAndADirectScopeIs is itd-199 ac-7.
func TestPresetNameIsNotAnOverrideAndADirectScopeIs(t *testing.T) {
	root := fixtureRepo(t)
	writeFile(t, root, ".abcd/config/reading-presets.json", narrowPresets())
	gitCommitAll(t, root)

	byPreset, err := Assemble(AssembleRequest{
		RepoRoot: root, Position: PositionDetection, Target: "HEAD", Scope: "cold", DryRun: true,
	})
	if err != nil {
		t.Fatalf("preset scope: %v", err)
	}
	if byPreset.Manifest.ScopeOverridden {
		t.Error("naming a committed preset was stamped as an override; running as reviewed is not a departure")
	}

	direct, err := Assemble(AssembleRequest{
		RepoRoot: root, Position: PositionDetection, Target: "HEAD", Scope: "spec", DryRun: true,
	})
	if err != nil {
		t.Fatalf("direct scope: %v", err)
	}
	if !direct.Manifest.ScopeOverridden {
		t.Error("naming a kind directly was not stamped as an override; the stamp exists to count " +
			"departures from the committed presets")
	}
}

// TestComparativeRefuses is itd-199 ac-10.
func TestComparativeRefuses(t *testing.T) {
	root := fixtureRepo(t)
	// A KIND token, deliberately, not the fixture preset. A preset carries no
	// comparative entry — validatePresets forbids one — so a preset-scoped
	// invocation would be refused by scope resolution even with the position
	// check gone, and this test would pass while proving nothing about the
	// refusal it names. A kind resolves without touching the presets, so the
	// position check is the only thing that can refuse here.
	_, err := Assemble(AssembleRequest{
		RepoRoot: root, Position: PositionComparative, Target: "HEAD",
		Scope: string(KindSpec), DryRun: true,
	})
	if err == nil {
		t.Fatal("the comparative position assembled; its object has no channel, so a bundle it " +
			"returns is about something other than what it was asked to read")
	}
	for _, want := range []string{"comparative", "channel"} {
		if !strings.Contains(strings.ToLower(err.Error()), want) {
			t.Errorf("the refusal does not name %q: %v", want, err)
		}
	}
}

// TestThreePositionsCarryDistinctItemSets is itd-199 ac-9, and it is the
// measured finding this intent exists to fix: three of the four positions used
// to receive a byte-identical item set.
func TestThreePositionsCarryDistinctItemSets(t *testing.T) {
	root := fixtureRepo(t)
	writeFile(t, root, ".abcd/config/reading-presets.json", narrowPresets())
	gitCommitAll(t, root)

	seen := map[string]Position{}
	for _, p := range AssemblingPositions() {
		res, err := Assemble(AssembleRequest{
			RepoRoot: root, Position: p, Target: "HEAD", Scope: "cold", DryRun: true,
		})
		if err != nil {
			t.Fatalf("position %s: %v", p, err)
		}
		key := strings.Join(itemPaths(res.Manifest), "\n")
		if other, dup := seen[key]; dup {
			t.Errorf("positions %s and %s carry the same item set; a table that cannot say what a "+
				"reading is about cannot distinguish readings that are about different things", other, p)
		}
		seen[key] = p
	}
}

// TestScopeCannotReachTheLedgerTier is brief invariant 14, held against an
// operator-written scope rather than only against the table.
func TestScopeCannotReachTheLedgerTier(t *testing.T) {
	root := fixtureRepo(t)
	for _, scope := range []string{fixtureScopeName, "doc", "config", "source"} {
		res, err := Assemble(AssembleRequest{
			RepoRoot: root, Position: PositionWidening, Target: "HEAD", Scope: scope, DryRun: true,
		})
		if err != nil {
			continue
		}
		for _, m := range res.Manifest.Items {
			if strings.Contains(m.Path, ".work.local") || strings.HasPrefix(m.Path, ".abcd/work/") {
				t.Errorf("scope %q admitted %s from the ledger tier", scope, m.Path)
			}
		}
	}
}

// TestDraftsStayDeniedAtWidening is itd-199 ac-4: a scope intersects what the
// table admits at the position and never widens it, so the deliberate drafts
// asymmetry survives a scope that names the intent kind.
func TestDraftsStayDeniedAtWidening(t *testing.T) {
	root := fixtureRepo(t)
	res, err := Assemble(AssembleRequest{
		RepoRoot: root, Position: PositionWidening, Target: "HEAD",
		Scope: string(KindIntentProjection), DryRun: true,
	})
	if err != nil {
		t.Fatalf("assemble: %v", err)
	}
	for _, m := range res.Manifest.Items {
		if strings.Contains(m.Path, "/drafts/") || strings.Contains(m.Path, "/planned/") {
			t.Errorf("a scope naming the intent kind at widening admitted %s; a scope narrows the "+
				"table's admission and never widens it", m.Path)
		}
	}
	if text := bundleText(res.Bundle); strings.Contains(text, sentinelDraftBody) {
		t.Error("the draft body reached a scoped widening assembly")
	}
}

// TestBundleCarriesTheScopeAndManifestCarriesItsHash is ac-5 and ac-6.
func TestBundleCarriesTheScopeAndManifestCarriesItsHash(t *testing.T) {
	root := fixtureRepo(t)
	res := assembleFixture(t, root, PositionWidening)

	if len(res.Bundle.Scope.Kinds) == 0 && len(res.Bundle.Scope.Records) == 0 &&
		res.Bundle.Scope.LocationNarrowings == 0 {
		t.Error("the bundle states no scope; a reader told its object is the shipped tree and " +
			"handed a subset reports the rest as a finding")
	}
	if res.Manifest.ScopeHash == "" {
		t.Error("the manifest carries no scope hash")
	}
	want, err := res.Manifest.Scope.Hash()
	if err != nil {
		t.Fatalf("hash: %v", err)
	}
	if res.Manifest.ScopeHash != want {
		t.Errorf("the manifest's scope hash is %q, want %q", res.Manifest.ScopeHash, want)
	}
	// The bundle carries what the reading was given and NOT who asked for it.
	if strings.Contains(string(mustEncodeBundle(t, res.Bundle)), "overridden") {
		t.Error("the bundle carries the override stamp; that is the auditor's fact, not the reading's")
	}
}

func mustEncodeBundle(t *testing.T, b Bundle) []byte {
	t.Helper()
	raw, err := EncodeBundle(b)
	if err != nil {
		t.Fatalf("encode: %v", err)
	}
	return raw
}

// TestPresetCollisionsAreRefused holds the decision that a colliding name is a
// configuration error rather than something a resolution ORDER decides
// silently.
func TestPresetCollisionsAreRefused(t *testing.T) {
	root := fixtureRepo(t)
	for name, body := range map[string]string{
		"a preset named for a kind": `{"schema_version": 1, "presets": {"source": {"positions":
			{"widening": {"kinds": ["doc"], "records": [], "paths": []}}}}}`,
		"a preset naming an unknown kind": `{"schema_version": 1, "presets": {"p": {"positions":
			{"widening": {"kinds": ["warm-ledger"], "records": [], "paths": []}}}}}`,
		"a preset naming a denied path": `{"schema_version": 1, "presets": {"p": {"positions":
			{"widening": {"kinds": [], "records": [], "paths": [".abcd/.work.local"]}}}}}`,
		"a preset naming an absolute path": `{"schema_version": 1, "presets": {"p": {"positions":
			{"widening": {"kinds": [], "records": [], "paths": ["/etc"]}}}}}`,
		"a preset escaping the repository": `{"schema_version": 1, "presets": {"p": {"positions":
			{"widening": {"kinds": [], "records": [], "paths": ["../elsewhere"]}}}}}`,
		"a preset scoping the comparative position": `{"schema_version": 1, "presets": {"p": {"positions":
			{"comparative": {"kinds": ["doc"], "records": [], "paths": []}}}}}`,
		"a two-level extends chain": `{"schema_version": 1, "presets": {
			"a": {"positions": {"widening": {"kinds": ["doc"], "records": [], "paths": []}}},
			"b": {"extends": "a", "positions": {}},
			"c": {"extends": "b", "positions": {}}}}`,
		"a wrong schema version": `{"schema_version": 99, "presets": {"p": {"positions": {}}}}`,
		"an unknown field":       `{"schema_version": 1, "presets": {}, "extra": true}`,
	} {
		writeFile(t, root, ".abcd/config/reading-presets.json", body)
		if _, err := LoadPresets(root); err == nil {
			t.Errorf("%s was accepted", name)
		}
	}
}

// TestUncommittedPresetRefuses is itd-199 ac-8: the preset configuration
// reshapes an assembly, so an uncommitted edit to it refuses exactly as an
// uncommitted edit to the record configuration does.
func TestUncommittedPresetRefuses(t *testing.T) {
	root := fixtureRepo(t)
	writeFile(t, root, ".abcd/config/reading-presets.json", narrowPresets())
	// deliberately NOT committed
	_, err := Assemble(AssembleRequest{
		RepoRoot: root, Position: PositionWidening, Target: "HEAD", Scope: "cold",
	})
	if err == nil {
		t.Fatal("an assembly ran against an uncommitted preset configuration")
	}
	if !strings.Contains(err.Error(), PresetConfigPath) {
		t.Errorf("the refusal does not name the preset configuration: %v", err)
	}
}

// TestAScopeSelectingNothingRefuses holds the choice of a refusal over an empty
// bundle. A reader handed an empty assembly has no way to tell "nothing matched"
// from "this object is empty", and would report the second.
func TestAScopeSelectingNothingRefuses(t *testing.T) {
	root := fixtureRepo(t)
	_, err := Assemble(AssembleRequest{
		RepoRoot: root, Position: PositionWidening, Target: "HEAD",
		Scope: "adr-9999", DryRun: true,
	})
	if err == nil {
		t.Fatal("a scope selecting no item assembled an empty bundle")
	}
	if !strings.Contains(err.Error(), "nothing to assemble") {
		t.Errorf("the refusal does not say the scope selected nothing: %v", err)
	}
}

// TestRecordSelectorDoesNotMatchByPrefix guards the obvious way to write the
// record matcher wrong: itd-19 must not select itd-198.
func TestRecordSelectorDoesNotMatchByPrefix(t *testing.T) {
	if pathNamesRecord(".abcd/development/intents/planned/itd-198-a-thing.md", "itd-19") {
		t.Error("itd-19 selected itd-198's record; a record selector matches the id, not a prefix of it")
	}
	if !pathNamesRecord(".abcd/development/intents/planned/itd-198-a-thing.md", "itd-198") {
		t.Error("itd-198 did not select its own record")
	}
	if !pathNamesRecord(".abcd/development/decisions/adrs/adr-58.md", "adr-58") {
		t.Error("an id-only basename was not selected")
	}
}

// TestScopeHashIsOrderIndependent holds the canonical ordering: two scopes
// naming the same thing in a different order are one scope, so a manifest's
// scope hash identifies WHAT was selected rather than how it was typed.
func TestScopeHashIsOrderIndependent(t *testing.T) {
	a := Scope{Selectors: canonicalise([]Selector{{Kind: KindDoc}, {Kind: KindSpec}})}
	b := Scope{Selectors: canonicalise([]Selector{{Kind: KindSpec}, {Kind: KindDoc}, {Kind: KindDoc}})}
	ha, err := a.Hash()
	if err != nil {
		t.Fatal(err)
	}
	hb, err := b.Hash()
	if err != nil {
		t.Fatal(err)
	}
	if ha != hb {
		t.Errorf("two scopes naming the same selectors hashed differently: %s vs %s", ha, hb)
	}
}

// TestBundleScopeCarriesNoRepositoryPath is brief invariant 15 held against the
// scope, and it exists because the obvious implementation broke it: writing one
// Scope type into both artefacts put a preset's path selectors — which are
// repository paths — into the reading's own working set
// (iss-2608312058244357).
//
// The manifest MAY carry paths; that is its job, and it is why the bundle can
// be pathless and still checkable. The bundle may not, under any scope.
func TestBundleScopeCarriesNoRepositoryPath(t *testing.T) {
	root := fixtureRepo(t)
	const secret = "internal/core/lint"
	writeFile(t, root, ".abcd/config/reading-presets.json", `{
  "schema_version": 1,
  "presets": {"pathy": {"positions": {"widening":
    {"kinds": ["doc"], "records": [], "paths": ["`+secret+`"]}}}}
}`)
	gitCommitAll(t, root)

	res, err := Assemble(AssembleRequest{
		RepoRoot: root, Position: PositionWidening, Target: "HEAD", Scope: "pathy", DryRun: true,
	})
	if err != nil {
		t.Fatalf("assemble: %v", err)
	}
	raw, err := EncodeBundle(res.Bundle)
	if err != nil {
		t.Fatalf("encode: %v", err)
	}
	// Checked against the SCOPE BLOCK and the structural keys, not against the
	// whole document. Item text legitimately contains path-like strings — a
	// record's prose mentions packages, and Go source carries import paths —
	// and that is pre-existing and inherent to carrying content at all. What
	// invariant 15 forbids is the bundle's own STRUCTURE mapping items to
	// locations: item keys are ordinals and the path mapping lives in the
	// manifest alone. A whole-document scan would pass here only for as long as
	// no fixture happened to mention the path, which is not a property.
	var doc map[string]json.RawMessage
	if err := json.Unmarshal(raw, &doc); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if strings.Contains(string(doc["scope"]), secret) {
		t.Errorf("the bundle's scope carries the repository path %q; the assembled input is the "+
			"reading's whole working set and no repository path may enter its structure", secret)
	}
	for key, val := range doc {
		if key == "items" {
			continue
		}
		if strings.Contains(string(val), secret) {
			t.Errorf("the bundle's %q carries the repository path %q", key, secret)
		}
	}
	var items []map[string]json.RawMessage
	if err := json.Unmarshal(doc["items"], &items); err != nil {
		t.Fatalf("decode items: %v", err)
	}
	for i, item := range items {
		if _, has := item["path"]; has {
			t.Errorf("bundle item %d carries a path key; the mapping is the manifest's alone", i)
		}
	}
	// The reading must still be TOLD it was narrowed, or it reports the absent
	// material as a finding against the record.
	if res.Bundle.Scope.LocationNarrowings == 0 {
		t.Error("the bundle does not record that a narrowing by location applied, so the reading " +
			"cannot tell it was handed a subset")
	}
	// And the manifest, which is the auditor's, must still carry the path.
	if !strings.Contains(mustEncodeManifest(t, res.Manifest), secret) {
		t.Errorf("the manifest does not carry %q; the auditor's account must stay complete", secret)
	}
}

func mustEncodeManifest(t *testing.T, m Manifest) string {
	t.Helper()
	raw, err := EncodeManifest(m)
	if err != nil {
		t.Fatalf("encode manifest: %v", err)
	}
	return string(raw)
}

// TestNoBundleFieldIsAScopeSelector guards the shape rather than one path: a
// future field added to Selector would flow into the bundle again unless the
// projection is explicit, so the bundle's scope is pinned to the three keys it
// is allowed to have.
func TestNoBundleFieldIsAScopeSelector(t *testing.T) {
	root := fixtureRepo(t)
	res := assembleFixture(t, root, PositionWidening)
	raw, err := EncodeBundle(res.Bundle)
	if err != nil {
		t.Fatalf("encode: %v", err)
	}
	var doc struct {
		Scope map[string]any `json:"scope"`
	}
	if err := json.Unmarshal(raw, &doc); err != nil {
		t.Fatalf("decode: %v", err)
	}
	allowed := map[string]bool{"kinds": true, "records": true, "location_narrowings": true}
	for key := range doc.Scope {
		if !allowed[key] {
			t.Errorf("the bundle's scope carries the key %q; only a pathless projection may "+
				"reach a reading", key)
		}
	}
	if _, leaked := doc.Scope["selectors"]; leaked {
		t.Error("the bundle carries raw selectors, which is the shape that leaked a path")
	}
}

// TestGatesSeeTheUnfilteredWalk pins the pipeline's ORDER, which was a claim in
// a comment and nothing more until this test existed.
//
// The scope filter must run AFTER the dirty gate and the exclusion assertion,
// so a narrow scope cannot shrink the set those gates examine. Moving the
// filter earlier passed the entire package before this test: the dirty gate's
// predicate is a pure function of the position and so refuses under either
// order, and the exclusion floor's paths are structurally denied so no
// candidate can breach one. The property was real, load-bearing and
// unfalsifiable — which is exactly the shape itd-195 says to make executable.
func TestGatesSeeTheUnfilteredWalk(t *testing.T) {
	root := fixtureRepo(t)
	writeFile(t, root, ".abcd/config/reading-presets.json", narrowPresets())
	gitCommitAll(t, root)

	var sawScoped, sawWide int
	restore := assertExclusionsHook
	t.Cleanup(func() { assertExclusionsHook = restore })
	assertExclusionsHook = func(cands []candidate, ex []Exclusion) error {
		sawScoped = len(cands)
		return restore(cands, ex)
	}
	if _, err := Assemble(AssembleRequest{
		RepoRoot: root, Position: PositionDetection, Target: "HEAD", Scope: "cold", DryRun: true,
	}); err != nil {
		t.Fatalf("narrow assembly: %v", err)
	}

	assertExclusionsHook = func(cands []candidate, ex []Exclusion) error {
		sawWide = len(cands)
		return restore(cands, ex)
	}
	wide, err := Assemble(AssembleRequest{
		RepoRoot: root, Position: PositionDetection, Target: "HEAD",
		Scope: fixtureScopeName, DryRun: true,
	})
	if err != nil {
		t.Fatalf("wide assembly: %v", err)
	}

	if sawScoped != sawWide {
		t.Errorf("the exclusion assertion saw %d candidates under a narrow scope and %d under a "+
			"wide one; the gates must run over the unfiltered walk, or a narrow scope can quiet "+
			"a breach a wide one would have caught", sawScoped, sawWide)
	}
	// And the guard must not be vacuous: the scope has to actually narrow, or
	// the two counts would match for an uninteresting reason.
	narrow, err := Assemble(AssembleRequest{
		RepoRoot: root, Position: PositionDetection, Target: "HEAD", Scope: "cold", DryRun: true,
	})
	if err != nil {
		t.Fatalf("narrow assembly: %v", err)
	}
	if len(narrow.Manifest.Items) >= len(wide.Manifest.Items) {
		t.Fatalf("the narrow scope emitted %d items and the wide one %d; this test proves nothing "+
			"unless the scope narrows", len(narrow.Manifest.Items), len(wide.Manifest.Items))
	}
	if sawScoped <= len(narrow.Manifest.Items) {
		t.Errorf("the exclusion assertion saw %d candidates but the run emitted %d; the gate was "+
			"handed the filtered set, not the walk", sawScoped, len(narrow.Manifest.Items))
	}
}

// TestAnUntrackedPresetRefuses is the committed-and-reviewed half of adr-58,
// which brief invariant 15 now asserts in prose and nothing checked.
//
// The dirty gate cannot supply it: git reports nothing for a file it ignores,
// so a repository gitignoring `.abcd/` — which brief invariant 5 does for
// public visibility — ran against an untracked preset and stamped the run
// `overridden: false`, asserting "ran as reviewed" on an examination that
// established only "git reported no modification".
func TestAnUntrackedPresetRefuses(t *testing.T) {
	root := fixtureRepo(t)
	// The preset must never have been committed, or the dirty gate catches it
	// as a MODIFIED tracked file and this test passes for the wrong reason.
	// So: ignore the directory, remove the fixture's committed preset, commit
	// that removal, and only then write the forged one.
	writeFile(t, root, ".gitignore", ".abcd/config/\n")
	if err := os.Remove(filepath.Join(root, ".abcd", "config", "reading-presets.json")); err != nil {
		t.Fatal(err)
	}
	gitCommitAll(t, root)
	writeFile(t, root, ".abcd/config/reading-presets.json", narrowPresets())

	// Precondition: git must see a clean tree, or the dirty gate is what
	// refuses below and this test proves nothing about trackedness.
	if out := gitStatusPorcelain(t, root); out != "" {
		t.Fatalf("the forged preset is visible to git, so this test cannot isolate the "+
			"tracked check:\n%s", out)
	}

	_, err := Assemble(AssembleRequest{
		RepoRoot: root, Position: PositionDetection, Target: "HEAD", Scope: "cold", DryRun: true,
	})
	if err == nil {
		t.Fatal("an assembly ran against an untracked, ignored preset; a preset is admitted at " +
			"the invocation because it is committed and reviewed, and this one was neither")
	}
	if !strings.Contains(err.Error(), "not tracked") {
		t.Errorf("the refusal does not say the preset is untracked: %v", err)
	}
}

// TestASymlinkedPresetRefuses closes the other half. A committed symlink IS
// tracked, and git reports nothing when its target changes, so the effective
// configuration would be permanently unreviewed and freely mutable with every
// gate green.
func TestASymlinkedPresetRefuses(t *testing.T) {
	root := fixtureRepo(t)
	outside := filepath.Join(t.TempDir(), "elsewhere.json")
	if err := os.WriteFile(outside, []byte(narrowPresets()), 0o644); err != nil {
		t.Fatal(err)
	}
	target := filepath.Join(root, ".abcd", "config", "reading-presets.json")
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(target); err != nil && !os.IsNotExist(err) {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, target); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}
	gitCommitAll(t, root)

	_, err := Assemble(AssembleRequest{
		RepoRoot: root, Position: PositionDetection, Target: "HEAD", Scope: "cold", DryRun: true,
	})
	if err == nil {
		t.Fatal("an assembly ran from a preset symlinked out of the tree; what it resolves to " +
			"is not what the commit recorded")
	}
	if !strings.Contains(err.Error(), "symlink") {
		t.Errorf("the refusal does not name the symlink: %v", err)
	}
}

// TestDuplicatePresetKeysAreRefused holds F2: Go's decoder takes the last
// duplicate silently, so a second block low in the file would replace the
// reviewed one while a reviewer reading top-down sees the first.
func TestDuplicatePresetKeysAreRefused(t *testing.T) {
	root := fixtureRepo(t)
	writeFile(t, root, ".abcd/config/reading-presets.json", `{
  "schema_version": 1,
  "presets": {
    "cold": {"positions": {"widening": {"kinds": ["brief-section"], "records": [], "paths": []}}},
    "cold": {"positions": {"widening": {"kinds": ["source", "test", "doc"], "records": [], "paths": []}}}
  }
}`)
	if _, err := LoadPresets(root); err == nil {
		t.Error("a preset file naming one preset twice was accepted; the last would win silently")
	} else if !strings.Contains(err.Error(), "more than once") {
		t.Errorf("the refusal does not name the duplication: %v", err)
	}
}

// TestPresetPathsRefuseTerminalAndWholeRepoForms holds F4 and the "." case: a
// control character in a path reaches a rendered line, and "." is a scope that
// narrows nothing.
func TestPresetPathsRefuseTerminalAndWholeRepoForms(t *testing.T) {
	for name, bad := range map[string]string{
		"an escape sequence":   "internal/core/\x1b]0;x\alint",
		"a NUL":                "internal/\x00core",
		"a newline":            "internal/core\n",
		"surrounding space":    " internal/core ",
		"backslash separator":  `internal\core\reading`,
		"the whole repo":       ".",
		"the whole repo slash": "./",
		"a case-folded deny":   "Internal/Core/Reading",
	} {
		if err := validPresetPath(bad); err == nil {
			t.Errorf("%s (%q) was accepted as a preset path", name, bad)
		}
	}
	if err := validPresetPath("internal/core/lint"); err != nil {
		t.Errorf("an ordinary repo-relative path was refused: %v", err)
	}
}

// gitStatusPorcelain is the tree's own account of itself, used to prove a
// precondition rather than to assert a result.
func gitStatusPorcelain(t *testing.T, root string) string {
	t.Helper()
	cmd := exec.Command("git", "-C", root, "status", "--porcelain", "-uall")
	// Hermetic, like every other git call in the tests: the ambient environment
	// must not decide what this probe sees (iss-28).
	cmd.Env = gittest.Env(t)
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("git status: %v", err)
	}
	return strings.TrimSpace(string(out))
}
