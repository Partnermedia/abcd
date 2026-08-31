package reading

import (
	"encoding/json"
	"strings"
	"testing"
)

// TestTestFilesCarryTheTestKind is itd-198 ac-5. The case rule is part of the
// criterion, not an implementation detail: the Go toolchain builds only a
// lowercase _test.go as a test, so a file matching the suffix in any other case
// stays source.
func TestTestFilesCarryTheTestKind(t *testing.T) {
	root := fixtureRepo(t)
	writeFile(t, root, "widget.go", "package main\n\nfunc Widget() {}\n")
	writeFile(t, root, "widget_test.go", "package main\n\nfunc TestWidget() {}\n")
	writeFile(t, root, "Gadget_TEST.go", "package main\n\nfunc GadgetUpper() {}\n")
	gitCommitAll(t, root)

	res := assembleFixture(t, root, PositionWidening)

	want := map[string]Kind{
		"widget.go":      KindSource,
		"widget_test.go": KindTest,
		"Gadget_TEST.go": KindSource,
	}
	got := map[string]Kind{}
	for _, m := range res.Manifest.Items {
		if _, ok := want[m.Path]; ok {
			got[m.Path] = m.Kind
		}
	}
	for path, wantKind := range want {
		gotKind, ok := got[path]
		if !ok {
			t.Errorf("%s was not admitted at all; the split must not change admission", path)
			continue
		}
		if gotKind != wantKind {
			t.Errorf("%s carries kind %q, want %q", path, gotKind, wantKind)
		}
	}
}

// TestKindSplitDoesNotMoveAdmission is itd-198 ac-4, proved by MUTATION rather
// than by a passing assertion: the test row is removed from the table and the
// admitted path set must be unchanged, because the row labels and never admits.
// A test that only asserted the current path set would pass just as happily if
// the row had narrowed admission at one position.
func TestKindSplitDoesNotMoveAdmission(t *testing.T) {
	root := fixtureRepo(t)
	writeFile(t, root, "widget.go", "package main\n\nfunc Widget() {}\n")
	writeFile(t, root, "widget_test.go", "package main\n\nfunc TestWidget() {}\n")
	gitCommitAll(t, root)

	// Restored via t.Cleanup, not by a trailing assignment: assembleFixture
	// calls t.Fatalf on error, which would skip a trailing restore and leave
	// the package-global Table missing the test row for every later test in
	// the run — a mutating test that corrupts the suite it belongs to.
	restore := Table
	t.Cleanup(func() { Table = restore })

	for _, p := range Positions() {
		Table = restore
		withRow := itemPaths(assembleFixture(t, root, p).Manifest)

		without := make([]Row, 0, len(restore))
		for _, row := range restore {
			if row.Kind == KindTest {
				continue
			}
			without = append(without, row)
		}
		Table = without
		withoutRow := itemPaths(assembleFixture(t, root, p).Manifest)
		Table = restore

		if strings.Join(withRow, "\n") != strings.Join(withoutRow, "\n") {
			t.Errorf("position %s: removing the test row changed the admitted path set; "+
				"the split must relabel and never admit or deny", p)
		}
	}
}

// TestTestKindIsReachableAtEveryPosition guards the mutation above from going
// vacuous: if no test file were admitted anywhere, the path sets would match
// trivially and the criterion would be untested.
func TestTestKindIsReachableAtEveryPosition(t *testing.T) {
	root := fixtureRepo(t)
	writeFile(t, root, "widget_test.go", "package main\n\nfunc TestWidget() {}\n")
	gitCommitAll(t, root)

	for _, p := range Positions() {
		res := assembleFixture(t, root, p)
		found := false
		for _, m := range res.Manifest.Items {
			if m.Kind == KindTest {
				found = true
			}
		}
		if !found {
			t.Errorf("position %s admitted no item of kind %q, so the admission "+
				"mutation proof above is vacuous there", p, KindTest)
		}
	}
}

// TestSizeReportSumsToTotal is itd-198 ac-1. Items and bytes sum exactly; the
// token total is derived from the total bytes rather than from summing the
// per-kind estimates, because integer division per kind would drift from the
// figure a reader compares against a capacity.
func TestSizeReportSumsToTotal(t *testing.T) {
	root := fixtureRepo(t)
	res := assembleFixture(t, root, PositionWidening)
	rep := res.Size

	if rep.Items != res.ItemCount {
		t.Errorf("the report counts %d items, the assembly passed %d", rep.Items, res.ItemCount)
	}
	var items, bytes int
	for _, k := range rep.ByKind {
		items += k.Items
		bytes += k.Bytes
	}
	if items != rep.Items {
		t.Errorf("per-kind items sum to %d, total says %d", items, rep.Items)
	}
	if bytes != rep.Bytes {
		t.Errorf("per-kind bytes sum to %d, total says %d", bytes, rep.Bytes)
	}
	if want := estimateTokens(rep.Bytes); rep.TokensEst != want {
		t.Errorf("total token estimate is %d, want %d from %d bytes", rep.TokensEst, want, rep.Bytes)
	}

	// The bytes reported are the bytes that actually travel.
	var bundleBytes int
	for _, it := range res.Bundle.Items {
		bundleBytes += len(it.Text)
	}
	if bundleBytes != rep.Bytes {
		t.Errorf("the report says %d bytes, the bundle carries %d", rep.Bytes, bundleBytes)
	}
}

// TestSizeReportOmitsKindsThatPassedNothing holds the distinction between an
// absent kind and an empty one. The vocabulary is closed and most of it is
// reachable, so a report listing every kind at zero would say less, not more.
func TestSizeReportOmitsKindsThatPassedNothing(t *testing.T) {
	root := fixtureRepo(t)
	res := assembleFixture(t, root, PositionWidening)

	present := map[Kind]bool{}
	for _, m := range res.Manifest.Items {
		present[m.Kind] = true
	}
	for _, k := range res.Size.ByKind {
		if !present[k.Kind] {
			t.Errorf("the report carries kind %q, which passed no item", k.Kind)
		}
		if k.Items == 0 {
			t.Errorf("the report carries kind %q with zero items", k.Kind)
		}
	}
	for kind := range present {
		found := false
		for _, k := range res.Size.ByKind {
			if k.Kind == kind {
				found = true
			}
		}
		if !found {
			t.Errorf("kind %q passed items and is missing from the report", kind)
		}
	}
}

// TestSizeReportRowsFollowTheVocabularyOrder keeps the rendering stable across
// runs: a map iteration would reorder the report for no reason a reader could
// see, and a diff of two runs would show a change that did not happen.
func TestSizeReportRowsFollowTheVocabularyOrder(t *testing.T) {
	root := fixtureRepo(t)
	rep := assembleFixture(t, root, PositionWidening).Size

	order := map[Kind]int{}
	for i, k := range Kinds() {
		order[k] = i
	}
	for i := 1; i < len(rep.ByKind); i++ {
		prev, cur := rep.ByKind[i-1].Kind, rep.ByKind[i].Kind
		if order[prev] >= order[cur] {
			t.Errorf("report row %d (%s) follows %s, which is not the vocabulary's order", i, cur, prev)
		}
	}
}

// TestDryRunCarriesTheSizeReport is itd-198 ac-2's core half: the report exists
// precisely so a size can be learned WITHOUT producing the artefact whose size
// is in question.
func TestDryRunCarriesTheSizeReport(t *testing.T) {
	root := fixtureRepo(t)
	res, err := Assemble(AssembleRequest{
		RepoRoot: root, Position: PositionDetection, Target: "HEAD", DryRun: true,
	})
	if err != nil {
		t.Fatalf("assemble: %v", err)
	}
	if res.Written {
		t.Fatal("a dry run wrote an artefact")
	}
	if res.Size.Bytes == 0 || res.Size.Items == 0 {
		t.Error("a dry run produced an empty size report")
	}
	if res.Size.TokensEst == 0 {
		t.Error("a dry run produced no token estimate")
	}
}

// TestSizeReportLabelsItsEstimate is itd-198 ac-3. The label travels in the
// artefact rather than only in a rendering, so a report read out of context
// still says it is an estimate rather than looking like a count.
func TestSizeReportLabelsItsEstimate(t *testing.T) {
	root := fixtureRepo(t)
	basis := assembleFixture(t, root, PositionWidening).Size.Basis
	if basis == "" {
		t.Fatal("the report states no basis")
	}
	for _, want := range []string{"estimated", "bytes /"} {
		if !strings.Contains(basis, want) {
			t.Errorf("the basis %q does not contain %q", basis, want)
		}
	}
	if strings.Contains(strings.ToLower(basis), "tokenizer's count") &&
		!strings.Contains(strings.ToLower(basis), "not a tokenizer's count") {
		t.Errorf("the basis %q reads as a tokenizer's count", basis)
	}
}

// TestBundleGainsNoFieldFromTheReport is itd-198 ac-8. The bundle's shape is
// the reading's whole working set, so a field added there is a field a reading
// must be told how to read.
func TestBundleGainsNoFieldFromTheReport(t *testing.T) {
	root := fixtureRepo(t)
	res := assembleFixture(t, root, PositionWidening)
	raw, err := EncodeBundle(res.Bundle)
	if err != nil {
		t.Fatalf("encode: %v", err)
	}

	// Decoded and inspected by SHAPE, not scanned as a substring. A bundle
	// carries verbatim Go source, so a scan for "bytes" matches any fixture
	// file that imports the bytes package — the assertion would have been
	// about the fixture's imports rather than about the bundle's shape.
	var top map[string]json.RawMessage
	if err := json.Unmarshal(raw, &top); err != nil {
		t.Fatalf("decode bundle: %v", err)
	}
	want := map[string]bool{"_type": true, "schema_version": true, "position": true, "items": true}
	for key := range top {
		if !want[key] {
			t.Errorf("the bundle carries the top-level key %q; the report rides on the result alone", key)
		}
	}

	var items []map[string]json.RawMessage
	if err := json.Unmarshal(top["items"], &items); err != nil {
		t.Fatalf("decode items: %v", err)
	}
	wantItem := map[string]bool{"item_key": true, "kind": true, "text": true}
	for i, item := range items {
		for key := range item {
			if !wantItem[key] {
				t.Errorf("bundle item %d carries the key %q", i, key)
			}
		}
	}
}

// TestManifestItemRoundTripsKind is itd-198 ac-7.
func TestManifestItemRoundTripsKind(t *testing.T) {
	root := fixtureRepo(t)
	res := assembleFixture(t, root, PositionWidening)
	raw, err := EncodeManifest(res.Manifest)
	if err != nil {
		t.Fatalf("encode: %v", err)
	}
	back, err := DecodeManifest(raw)
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(back.Items) != len(res.Manifest.Items) {
		t.Fatalf("decoded %d items, encoded %d", len(back.Items), len(res.Manifest.Items))
	}
	for i, m := range back.Items {
		if m.Kind == "" {
			t.Errorf("%s decoded with no kind", m.ItemKey)
		}
		if m.Kind != res.Manifest.Items[i].Kind {
			t.Errorf("%s decoded kind %q, encoded %q", m.ItemKey, m.Kind, res.Manifest.Items[i].Kind)
		}
	}
}

// TestManifestItemKindIsNotOmitted holds the not-omitempty decision. An item
// with no kind must be visible in the bytes, because a shape that can omit the
// field cannot tell a defect from a well-formed item.
func TestManifestItemKindIsNotOmitted(t *testing.T) {
	m := Manifest{
		Type:          ManifestType,
		SchemaVersion: SchemaVersion,
		Items:         []ManifestItem{{ItemKey: "itm-0001", Path: "a.go", SHA256: "x"}},
	}
	raw, err := EncodeManifest(m)
	if err != nil {
		t.Fatalf("encode: %v", err)
	}
	if !strings.Contains(string(raw), "\"kind\"") {
		t.Error("an item with an empty kind encoded without the field; a missing kind " +
			"must be visible rather than indistinguishable from a well-formed item")
	}
}

// TestRenderCoversKindAndSuffix proves by MUTATION that the rendering the
// assembler version digests actually covers the two fields itd-198 adds to it.
// The rendering previously emitted neither, so a kind reassignment moved every
// bundle while the version stood still — the exact vacuity this criterion
// exists to close, which is why it is proved by breaking it.
func TestRenderCoversKindAndSuffix(t *testing.T) {
	restore := Table
	t.Cleanup(func() { Table = restore })

	base := Render()

	kindMutated := make([]Row, len(restore))
	copy(kindMutated, restore)
	kindMutated[0].Kind = KindConfig
	Table = kindMutated
	if Render() == base {
		t.Error("reassigning a row's kind did not move the rendering")
	}

	suffixMutated := make([]Row, len(restore))
	copy(suffixMutated, restore)
	suffixMutated[0].MatchSuffix = append([]string{"_extra.go"}, suffixMutated[0].MatchSuffix...)
	Table = suffixMutated
	if Render() == base {
		t.Error("adding a match suffix to a row did not move the rendering")
	}
}

// TestASuffixOnlyRowIsNotReadAsEveryFile is the trap the two match forms set
// for each other. An empty Match means "every file" on a row that selects by
// nothing else, and means "this form contributes nothing" on a row that selects
// by suffix. Reading the second as the first would admit the whole tree, and
// would render the contract as admitting it too.
func TestASuffixOnlyRowIsNotReadAsEveryFile(t *testing.T) {
	row := Row{
		Positions:   allPositions,
		Source:      ".",
		MatchSuffix: []string{"_test.go"},
		Kind:        KindTest,
		Rule:        "fixture",
	}
	if !row.matches("widget_test.go") {
		t.Error("a suffix-only row does not match the suffix it declares")
	}
	if row.matches("widget.go") {
		t.Error("a suffix-only row matched a file its suffix does not select; an empty " +
			"Match beside a non-empty MatchSuffix must not fall through to every-file")
	}
	if got := matchList(row); got == "every file" {
		t.Error("a suffix-only row renders its Matches column as \"every file\", stating " +
			"the opposite of what it admits in the text the version digests")
	}
}

// TestMatchListPinsEveryBranch pins the three cases of the Matches column
// independently of Row.matches. The two functions decide the same question in
// different files — what a row selects, and what the rendering SAYS it selects
// — and a mutation to one leaves the other's branch unproven, which is how a
// rendering comes to disagree with the table it renders. Each branch is pinned
// by its exact output rather than by "not every file", so a regression that
// merely reworded it is caught too.
func TestMatchListPinsEveryBranch(t *testing.T) {
	for name, tc := range map[string]struct {
		row  Row
		want string
	}{
		"neither form selects anything": {
			row: Row{Source: "."}, want: "every file",
		},
		"suffix only": {
			row:  Row{Source: ".", MatchSuffix: []string{"_test.go"}},
			want: "none",
		},
		"match only": {
			row: Row{Source: ".", Match: []string{".go"}}, want: "`.go`",
		},
		"both forms": {
			row:  Row{Source: ".", Match: []string{".go"}, MatchSuffix: []string{"_test.go"}},
			want: "`.go`",
		},
	} {
		if got := matchList(tc.row); got != tc.want {
			t.Errorf("%s: matchList = %q, want %q", name, got, tc.want)
		}
	}
}

// TestSuffixListPinsEveryBranch does the same for the Suffixes column.
func TestSuffixListPinsEveryBranch(t *testing.T) {
	if got := suffixList(nil); got != "none" {
		t.Errorf("suffixList(nil) = %q, want %q", got, "none")
	}
	if got := suffixList([]string{"_test.go"}); got != "`_test.go`" {
		t.Errorf("suffixList one = %q, want %q", got, "`_test.go`")
	}
	if got := suffixList([]string{"_test.go", "_gen.go"}); got != "`_test.go`, `_gen.go`" {
		t.Errorf("suffixList two = %q, want %q", got, "`_test.go`, `_gen.go`")
	}
}

// TestSizeBasisBringsNoParenthesesOfItsOwn pins the shape the rehearsal caught.
// The CLI wraps the basis in a parenthetical, so a basis carrying its own
// rendered as "((...))" on every run. Every gate was green and every fixture
// passed; only running the thing over a real tree showed it.
func TestSizeBasisBringsNoParenthesesOfItsOwn(t *testing.T) {
	if strings.ContainsAny(sizeBasis, "()") {
		t.Errorf("sizeBasis %q carries a parenthesis; the CLI supplies the pair", sizeBasis)
	}
}

// TestDecodeManifestRefusesAnItemWithoutAKind is the read side of the
// not-omitempty decision. The write side alone was not enough: the type's own
// comment claimed a missing kind is distinguishable from a well-formed item,
// and that was true of what this package WRITES and false of what it READS,
// which is an attestation asserting more than its examination establishes.
func TestDecodeManifestRefusesAnItemWithoutAKind(t *testing.T) {
	root := fixtureRepo(t)
	res := assembleFixture(t, root, PositionWidening)
	raw, err := EncodeManifest(res.Manifest)
	if err != nil {
		t.Fatalf("encode: %v", err)
	}

	// An explicit empty kind, and the key removed altogether: both are the same
	// defect and both must be refused, because a decoder that only caught the
	// spelled-out empty string would pass the commoner shape.
	first := res.Manifest.Items[0]
	for name, bad := range map[string]string{
		"empty kind":   strings.Replace(string(raw), "\"kind\": \""+string(first.Kind)+"\"", "\"kind\": \"\"", 1),
		"absent kind":  strings.Replace(string(raw), "\n      \"kind\": \""+string(first.Kind)+"\",", "", 1),
		"unknown kind": strings.Replace(string(raw), "\"kind\": \""+string(first.Kind)+"\"", "\"kind\": \"warm-ledger\"", 1),
	} {
		if bad == string(raw) {
			t.Fatalf("%s: the substitution did not change the document, so the case tests nothing", name)
		}
		if _, err := DecodeManifest([]byte(bad)); err == nil {
			t.Errorf("a manifest with an %s decoded", name)
		}
	}
}
