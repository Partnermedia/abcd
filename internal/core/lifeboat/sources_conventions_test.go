package lifeboat

import (
	"os"
	"path/filepath"
	"testing"
)

// convTierFixture builds a plain (non-git) directory that exercises the Tier-1
// conventions signals: a README with real prose and a features section, a
// go.mod/go.sum manifest pair, an ADR under docs/adr, and a CHANGELOG. It
// deliberately carries no glossary, so that section is a genuine absent-material
// blank. Returns the directory.
func convTierFixture(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	write := func(rel, content string) {
		full := filepath.Join(dir, rel)
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	write("README.md", "# Widget\n\n"+
		"Widget is a small tool that turns sprockets into cogs. It is designed "+
		"for people who need cogs quickly and cannot afford a full foundry, and "+
		"it has been used in production by several teams.\n\n"+
		"## Features\n\n- Fast sprocket conversion\n- Deterministic output\n\n"+
		"## Usage\n\n```\nwidget convert in.spr\n```\n\n"+
		"## Non-goals\n\nWidget will never smelt raw ore.\n")
	write("go.mod", "module example.com/widget\n\ngo 1.22\n")
	write("go.sum", "example.com/dep v1.0.0 h1:abc=\n")
	write("docs/adr/0001-record-architecture-decisions.md", "# 1. Record ADRs\n\nContext.\n")
	write("CHANGELOG.md", "# Changelog\n\n## 1.0.0\n\n- Initial release\n")

	return dir
}

// convSourceForSection returns the Tier-1 adapter that speaks for section,
// failing the test if none does.
func convSourceForSection(t *testing.T, section Section) Source {
	t.Helper()
	for _, s := range conventionSources() {
		if s.Section() == section {
			return s
		}
	}
	t.Fatalf("no conventions source for section %s", section)
	return nil
}

// TestHasConventionsMatchesAdapterEvidence guards that the tier gate is not
// narrower than what the Tier-1 adapters read: a repo carrying build manifests
// and CI workflows but none of the headline docs (README/LICENSE/CHANGELOG/
// CONTRIBUTING) must still count as having conventions, so its platform and
// dependency adapters run instead of being skipped and reported as false blanks.
func TestHasConventionsMatchesAdapterEvidence(t *testing.T) {
	dir := t.TempDir()
	write := func(rel, content string) {
		full := filepath.Join(dir, rel)
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	// Only manifest + lockfile + CI workflows — none of the headline doc sentinels.
	write("go.mod", "module example.com/tool\n\ngo 1.22\n")
	write("go.sum", "example.com/dep v1.0.0 h1:abc=\n")
	write(".github/workflows/ci.yml", "name: ci\n")

	ctx, err := newSourceContext(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer ctx.Close()

	if !hasConventions(ctx) {
		t.Fatal("hasConventions=false for a go.mod/go.sum/.github/workflows repo; Tier-1 adapters would be skipped and blanked")
	}
	present := tiersPresent(ctx)
	found := false
	for _, tr := range present {
		if tr == TierConventions {
			found = true
		}
	}
	if !found {
		t.Fatalf("tiersPresent=%v omits TierConventions for a manifest+CI repo", present)
	}

	// The gated adapter must actually ground, proving the tier was not skipped.
	if ev := convSourceForSection(t, "constraints/platform").Probe(ctx); ev.Status != StatusGrounded {
		t.Fatalf("constraints/platform status = %s, want grounded", ev.Status)
	}
}

// TestConvContextGroundsFromREADME is the flagship Tier-1 assertion: a README
// with real prose grounds "product/context" and cites the README file.
func TestConvContextGroundsFromREADME(t *testing.T) {
	ctx, err := newSourceContext(convTierFixture(t))
	if err != nil {
		t.Fatal(err)
	}
	defer ctx.Close()

	src := convSourceForSection(t, "product/context")
	if src.Tier() != TierConventions {
		t.Fatalf("product/context tier = %s, want %s", src.Tier(), TierConventions)
	}
	ev := src.Probe(ctx)
	if ev.Status != StatusGrounded {
		t.Fatalf("product/context status = %s, want grounded", ev.Status)
	}
	if len(ev.Sources) == 0 {
		t.Fatal("grounded product/context cites no evidence")
	}
	if ev.Sources[0] != "README.md" {
		t.Errorf("product/context evidence = %v, want README.md cited", ev.Sources)
	}
	if ev.Confidence == "" {
		t.Error("grounded product/context has no confidence")
	}
}

// TestConvPlatformGroundsFromManifests confirms build manifests ground
// "constraints/platform" with the manifest cited.
func TestConvPlatformGroundsFromManifests(t *testing.T) {
	ctx, err := newSourceContext(convTierFixture(t))
	if err != nil {
		t.Fatal(err)
	}
	defer ctx.Close()

	ev := convSourceForSection(t, "constraints/platform").Probe(ctx)
	if ev.Status != StatusGrounded {
		t.Fatalf("constraints/platform status = %s, want grounded", ev.Status)
	}
	if !containsSource(ev.Sources, "go.mod") {
		t.Errorf("constraints/platform evidence = %v, want go.mod cited", ev.Sources)
	}
}

// TestConvDependenciesGroundsFromPair confirms a manifest+lockfile pair grounds
// "constraints/dependencies" and cites both files.
func TestConvDependenciesGroundsFromPair(t *testing.T) {
	ctx, err := newSourceContext(convTierFixture(t))
	if err != nil {
		t.Fatal(err)
	}
	defer ctx.Close()

	ev := convSourceForSection(t, "constraints/dependencies").Probe(ctx)
	if ev.Status != StatusGrounded {
		t.Fatalf("constraints/dependencies status = %s, want grounded", ev.Status)
	}
	if !containsSource(ev.Sources, "go.mod") || !containsSource(ev.Sources, "go.sum") {
		t.Errorf("constraints/dependencies evidence = %v, want go.mod and go.sum", ev.Sources)
	}
}

// TestConvADRsGroundFromDocsAdr confirms ADRs under docs/adr ground "docs/adrs"
// and the citation lists the ADR file.
func TestConvADRsGroundFromDocsAdr(t *testing.T) {
	ctx, err := newSourceContext(convTierFixture(t))
	if err != nil {
		t.Fatal(err)
	}
	defer ctx.Close()

	ev := convSourceForSection(t, "docs/adrs").Probe(ctx)
	if ev.Status != StatusGrounded {
		t.Fatalf("docs/adrs status = %s, want grounded", ev.Status)
	}
	if !containsSource(ev.Sources, "docs/adr/0001-record-architecture-decisions.md") {
		t.Errorf("docs/adrs evidence = %v, want the ADR file listed", ev.Sources)
	}
}

// TestConvWhatWorkedPartialFromChangelog confirms a CHANGELOG partially grounds
// "evidence/what-worked" with the file cited.
func TestConvWhatWorkedPartialFromChangelog(t *testing.T) {
	ctx, err := newSourceContext(convTierFixture(t))
	if err != nil {
		t.Fatal(err)
	}
	defer ctx.Close()

	ev := convSourceForSection(t, "evidence/what-worked").Probe(ctx)
	if ev.Status != StatusPartial {
		t.Fatalf("evidence/what-worked status = %s, want partial", ev.Status)
	}
	if !containsSource(ev.Sources, "CHANGELOG.md") {
		t.Errorf("evidence/what-worked evidence = %v, want CHANGELOG.md", ev.Sources)
	}
}

// TestConvGlossaryBlankWhenAbsent holds the "a blank is a result" contract for
// Tier-1: with no glossary in the fixture, the section returns a blank that names
// what it searched and the question a human must answer.
func TestConvGlossaryBlankWhenAbsent(t *testing.T) {
	ctx, err := newSourceContext(convTierFixture(t))
	if err != nil {
		t.Fatal(err)
	}
	defer ctx.Close()

	ev := convSourceForSection(t, "glossary").Probe(ctx)
	if ev.Status != StatusBlank {
		t.Fatalf("glossary status = %s, want blank (no glossary in fixture)", ev.Status)
	}
	if ev.Question == "" {
		t.Error("blank glossary carries no question for a human")
	}
	if len(ev.Searched) == 0 {
		t.Error("blank glossary names nothing it searched")
	}
}

// TestConvOutOfScopePartialFromNonGoals confirms a README non-goals section
// partially grounds "delivery/out-of-scope".
func TestConvOutOfScopePartialFromNonGoals(t *testing.T) {
	ctx, err := newSourceContext(convTierFixture(t))
	if err != nil {
		t.Fatal(err)
	}
	defer ctx.Close()

	ev := convSourceForSection(t, "delivery/out-of-scope").Probe(ctx)
	if ev.Status != StatusPartial {
		t.Fatalf("delivery/out-of-scope status = %s, want partial", ev.Status)
	}
	if len(ev.Sources) == 0 {
		t.Fatal("partial delivery/out-of-scope cites no evidence")
	}
}

// TestConvAdaptersAllReportTierConventions guards the tier contract: every Tier-1
// adapter reports TierConventions.
func TestConvAdaptersAllReportTierConventions(t *testing.T) {
	for _, s := range conventionSources() {
		if s.Tier() != TierConventions {
			t.Errorf("source for %s reports tier %s, want %s", s.Section(), s.Tier(), TierConventions)
		}
	}
}

// TestConvProbeIsDeterministic asserts the Tier-1 adapters are byte-stable across
// runs against the same directory.
func TestConvProbeIsDeterministic(t *testing.T) {
	dir := convTierFixture(t)
	ctxA, err := newSourceContext(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer ctxA.Close()
	ctxB, err := newSourceContext(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer ctxB.Close()

	for _, s := range conventionSources() {
		a := s.Probe(ctxA)
		b := s.Probe(ctxB)
		if len(a.Sources) != len(b.Sources) {
			t.Fatalf("%s: source count differs across runs", s.Section())
		}
		for i := range a.Sources {
			if a.Sources[i] != b.Sources[i] {
				t.Errorf("%s: source[%d] differs: %q vs %q", s.Section(), i, a.Sources[i], b.Sources[i])
			}
		}
	}
}

// containsSource reports whether want appears in sources.
func containsSource(sources []string, want string) bool {
	for _, s := range sources {
		if s == want {
			return true
		}
	}
	return false
}
