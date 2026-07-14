package lifeboat

import "testing"

// covWith builds a Coverage whose named sections carry the given statuses; every
// other mapping section is blank. It lets a test state a repo's coverage as a
// small map rather than a full 23-row literal.
func covWith(name string, commits int, tiers []Tier, statuses map[Section]Status) Coverage {
	sections := make([]SectionCoverage, 0, len(Table))
	var sum Summary
	for _, m := range Table {
		st := StatusBlank
		if s, ok := statuses[m.Section]; ok {
			st = s
		}
		sections = append(sections, SectionCoverage{Name: m.Section, Status: st})
		switch st {
		case StatusGrounded:
			sum.Grounded++
		case StatusPartial:
			sum.Partial++
		default:
			sum.Blank++
		}
	}
	return Coverage{
		SchemaVersion: SchemaVersion,
		Repo:          RepoInfo{Name: name, Commits: commits},
		TiersPresent:  tiers,
		Sections:      sections,
		Summary:       sum,
	}
}

// TestAggregateBuildsSectionByRepoTable is the M2 readout: the aggregate holds
// one row per brief section and one cell per repo, each cell that section's
// status in that repo. This table is what the gate reads.
func TestAggregateBuildsSectionByRepoTable(t *testing.T) {
	rich := covWith("rich", 200, []Tier{TierGit, TierConventions, TierNative}, map[Section]Status{
		"graveyard":       StatusGrounded,
		"product/context": StatusGrounded,
		"docs/adrs":       StatusGrounded,
	})
	poor := covWith("poor", 12, []Tier{TierGit}, map[Section]Status{
		"graveyard": StatusGrounded,
	})

	agg := Aggregate([]Coverage{rich, poor})

	if len(agg.Sections) != len(Table) {
		t.Fatalf("aggregate has %d rows, mapping has %d", len(agg.Sections), len(Table))
	}
	row := findRow(t, agg, "graveyard")
	if row.Cells["rich"] != StatusGrounded || row.Cells["poor"] != StatusGrounded {
		t.Errorf("graveyard cells = %v, want both grounded", row.Cells)
	}
	ctx := findRow(t, agg, "product/context")
	if ctx.Cells["rich"] != StatusGrounded || ctx.Cells["poor"] != StatusBlank {
		t.Errorf("product/context cells = %v, want rich grounded, poor blank", ctx.Cells)
	}
}

// TestAggregateAlwaysBlankNeedsEveryRepoBlank is the "what the record is worth"
// signal: a section counts as always-blank only when NO probed repo grounded or
// even partially grounded it. One partial rescues it from the verdict.
func TestAggregateAlwaysBlankNeedsEveryRepoBlank(t *testing.T) {
	a := covWith("a", 100, []Tier{TierGit}, map[Section]Status{
		"product/personas": StatusBlank,
	})
	b := covWith("b", 100, []Tier{TierGit, TierConventions}, map[Section]Status{
		"product/personas": StatusBlank,
	})
	agg := Aggregate([]Coverage{a, b})
	if !containsSection(agg.AlwaysBlankSections(), "product/personas") {
		t.Error("product/personas blank in every repo but not reported always-blank")
	}

	// One partial anywhere must remove it from the always-blank set.
	b2 := covWith("b", 100, []Tier{TierGit, TierConventions, TierNative}, map[Section]Status{
		"product/personas": StatusPartial,
	})
	agg2 := Aggregate([]Coverage{a, b2})
	if containsSection(agg2.AlwaysBlankSections(), "product/personas") {
		t.Error("product/personas is partial in one repo but still reported always-blank")
	}
}

// TestAggregateDisambiguatesDuplicateRepoNames guards the cross-repo table
// against two repos sharing a basename (e.g. two clones both named "cli"): their
// columns must not collapse into one.
func TestAggregateDisambiguatesDuplicateRepoNames(t *testing.T) {
	a := covWith("cli", 10, []Tier{TierGit}, map[Section]Status{"graveyard": StatusGrounded})
	b := covWith("cli", 20, []Tier{TierGit}, map[Section]Status{"graveyard": StatusBlank})
	agg := Aggregate([]Coverage{a, b})
	if len(agg.Repos) != 2 {
		t.Fatalf("want 2 repo columns, got %d", len(agg.Repos))
	}
	if agg.Repos[0].Name == agg.Repos[1].Name {
		t.Errorf("duplicate repo names not disambiguated: both %q", agg.Repos[0].Name)
	}
	row := findRow(t, agg, "graveyard")
	if len(row.Cells) != 2 {
		t.Errorf("graveyard row has %d cells, want 2 distinct repo columns", len(row.Cells))
	}
}

func findRow(t *testing.T, agg AggregateReport, section Section) AggregateRow {
	t.Helper()
	for _, r := range agg.Sections {
		if r.Section == section {
			return r
		}
	}
	t.Fatalf("section %s not in aggregate", section)
	return AggregateRow{}
}

func containsSection(in []Section, s Section) bool {
	for _, x := range in {
		if x == s {
			return true
		}
	}
	return false
}
