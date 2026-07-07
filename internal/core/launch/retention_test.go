package launch

import "testing"

func mustParse(t *testing.T, v string) Semver {
	t.Helper()
	s, err := ParseSemver(v)
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func contains(tags []string, tag string) bool {
	for _, t := range tags {
		if t == tag {
			return true
		}
	}
	return false
}

func TestRetentionPrunesOlderInLine(t *testing.T) {
	plan := ComputeRetention(mustParse(t, "0.1.2"), []Semver{mustParse(t, "0.1.1")})
	if plan.Refused {
		t.Fatalf("unexpected refusal: %+v", plan)
	}
	if !contains(plan.Pruned, "v0.1.1") {
		t.Errorf("expected v0.1.1 pruned: %+v", plan.Pruned)
	}
	if !contains(plan.Kept, "v0.1.2") {
		t.Errorf("published must be kept: %+v", plan.Kept)
	}
	if contains(plan.Pruned, "v0.1.2") {
		t.Errorf("published must never be pruned: %+v", plan.Pruned)
	}
}

func TestRetentionKeepsOtherLines(t *testing.T) {
	plan := ComputeRetention(mustParse(t, "0.2.0"), []Semver{mustParse(t, "0.1.9")})
	if !contains(plan.Kept, "v0.1.9") || !contains(plan.Kept, "v0.2.0") {
		t.Errorf("both lines' newest must be kept: %+v", plan.Kept)
	}
	if len(plan.Pruned) != 0 {
		t.Errorf("nothing to prune across distinct lines: %+v", plan.Pruned)
	}
}

func TestRetentionRefusesOnNewerExisting(t *testing.T) {
	plan := ComputeRetention(mustParse(t, "0.1.2"), []Semver{mustParse(t, "0.1.3")})
	if !plan.Refused || plan.RefusalReason == "" {
		t.Errorf("expected refusal on newer existing release: %+v", plan)
	}
	if len(plan.Pruned) != 0 {
		t.Errorf("a refusal must prune nothing: %+v", plan.Pruned)
	}
}
