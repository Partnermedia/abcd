package rules

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func names(ds []ResolvedDomain) []string {
	out := make([]string, len(ds))
	for i, d := range ds {
		out[i] = d.Name
	}
	return out
}

func has(ds []ResolvedDomain, name string) bool {
	for _, d := range ds {
		if d.Name == name {
			return true
		}
	}
	return false
}

func TestDefaultsParseAndValidate(t *testing.T) {
	rs := Defaults()
	if rs.SchemaVersion != 1 {
		t.Fatalf("schema_version = %d, want 1", rs.SchemaVersion)
	}
	if err := Validate(rs); err != nil {
		t.Fatalf("bundled defaults fail validation: %v", err)
	}
	for _, want := range []string{"COMMITTING", "DOCUMENTATION", "ROADMAP", "ISSUES", "INTENTS", "LIFEBOAT", "PII", "OPINIONS"} {
		if _, ok := rs.Domains[want]; !ok {
			t.Errorf("default domain %q missing", want)
		}
	}
}

func TestOpinionsDomainPointsAtPrinciplesNotCopies(t *testing.T) {
	rs := Defaults()
	// Recall on an opinion/convention/SOTA prompt.
	if !has(rs.Match("what's the SOTA approach and our convention here"), "OPINIONS") {
		t.Fatal("OPINIONS did not recall-match a conventions prompt")
	}
	op := rs.Domains["OPINIONS"]
	// Every rule points at a principle file (one-canonical-primitive): it names a
	// path under .abcd/development/principles/, it does not inline the principle.
	pointers := 0
	for _, r := range op.Rules {
		if strings.Contains(r, ".abcd/development/principles/") {
			pointers++
		}
	}
	if pointers < len(op.Rules)-1 { // allow the one index rule to name the dir
		t.Fatalf("OPINIONS rules should point at principle files, got %d/%d", pointers, len(op.Rules))
	}
}

func TestMatchRecallKeyword(t *testing.T) {
	rs := Defaults()
	got := rs.Match("let's commit and push this")
	if !has(got, "COMMITTING") {
		t.Fatalf("expected COMMITTING, got %v", names(got))
	}
}

func TestMatchMultiWordAlias(t *testing.T) {
	rs := Defaults()
	got := rs.Match("please open the pull request now")
	if !has(got, "COMMITTING") {
		t.Fatalf("expected COMMITTING via multi-word alias, got %v", names(got))
	}
}

func TestMatchNoHitInjectsNothing(t *testing.T) {
	rs := Defaults()
	got := rs.Match("render a react component with a gradient background")
	if len(got) != 0 {
		t.Fatalf("expected zero domains on no-match, got %v", names(got))
	}
}

func TestMatchWordBoundaryNoSubstringFalsePositive(t *testing.T) {
	rs := Defaults()
	// "scommitted" must not trigger COMMITTING's "commit" keyword.
	got := rs.Match("the discommitted witticism")
	if has(got, "COMMITTING") {
		t.Fatalf("substring false positive: %v", names(got))
	}
}

func TestStarCommandActivatesRegardlessOfKeyword(t *testing.T) {
	rs := Defaults()
	got := rs.Match("*ROADMAP draft the next milestone")
	if !has(got, "ROADMAP") {
		t.Fatalf("star-command did not activate ROADMAP: %v", names(got))
	}
}

func TestStarCommandBoundaries(t *testing.T) {
	// Synthetic domain whose name is NOT one of its recall keywords, so a hit can
	// only come from star-command parsing (not incidental recall on the name).
	rs := RuleSet{SchemaVersion: 1, Domains: map[string]Domain{
		"ZONK": {State: StateActive, Recall: []string{"zzznomatch"}, Rules: []string{"r"}},
	}}
	// Positive control: a well-formed star-command activates.
	if !has(rs.Match("*ZONK do the thing"), "ZONK") {
		t.Fatal("well-formed *ZONK not activated")
	}
	// Boundary rejections: star preceded by a non-space, glued to a longer token,
	// or not an uppercase name are NOT star-commands.
	for _, p := range []string{"path/*ZONK", "e*ZONK", "*ZONKING now", "list *.py files", "* ZONK bullet"} {
		if got := rs.Match(p); has(got, "ZONK") {
			t.Errorf("%q wrongly parsed as a star-command: %v", p, names(got))
		}
	}
}

func TestStarCommandActivatesDormant(t *testing.T) {
	rs := Defaults()
	d := rs.Domains["ROADMAP"]
	d.State = StateDormant
	rs.Domains["ROADMAP"] = d
	// dormant: no recall activation...
	if has(rs.Match("update the roadmap"), "ROADMAP") {
		t.Fatal("dormant domain activated by recall")
	}
	// ...but star-command overrides dormant.
	if !has(rs.Match("*ROADMAP go"), "ROADMAP") {
		t.Fatal("star-command did not override dormant")
	}
}

func TestKillSwitchSuppressesEverything(t *testing.T) {
	rs := Defaults()
	rs.Disabled = true
	if got := rs.Match("commit and push"); len(got) != 0 {
		t.Fatalf("kill switch did not suppress recall: %v", names(got))
	}
	// Star-command must NOT bypass the kill switch.
	if got := rs.Match("*ROADMAP go"); len(got) != 0 {
		t.Fatalf("star-command bypassed the kill switch: %v", names(got))
	}
}

func TestMergePerFieldOverride(t *testing.T) {
	base := Defaults()
	over := RuleSet{
		SchemaVersion: 1,
		Domains: map[string]Domain{
			"ROADMAP": {State: StateDormant}, // silence, keep recall/rules
			"CUSTOM":  {State: StateActive, Recall: []string{"widget"}, Rules: []string{"do the thing"}},
		},
	}
	merged := Merge(base, over)
	// ROADMAP keeps its default recall but is now dormant.
	if got := merged.Domains["ROADMAP"]; got.State != StateDormant || len(got.Recall) == 0 {
		t.Fatalf("per-field override wrong: state=%q recall=%v", got.State, got.Recall)
	}
	// CUSTOM added.
	if _, ok := merged.Domains["CUSTOM"]; !ok {
		t.Fatal("custom domain not merged in")
	}
	if !has(merged.Match("build a widget"), "CUSTOM") {
		t.Fatal("merged custom domain does not recall-match")
	}
}

func TestMergeKillSwitchIsSticky(t *testing.T) {
	base := Defaults()
	if got := Merge(base, RuleSet{Disabled: true}); !got.Disabled {
		t.Fatal("repo override could not enable the kill switch")
	}
}

func TestValidateRejectsBadDomainName(t *testing.T) {
	rs := RuleSet{SchemaVersion: 1, Domains: map[string]Domain{"bad-name": {}}}
	if err := Validate(rs); err == nil {
		t.Fatal("expected validation error for lowercase/hyphen domain name")
	}
}

func TestValidateRejectsBadSchemaVersion(t *testing.T) {
	if err := Validate(RuleSet{SchemaVersion: 2}); err == nil {
		t.Fatal("expected validation error for schema_version != 1")
	}
}

func TestValidateRejectsBadState(t *testing.T) {
	rs := RuleSet{SchemaVersion: 1, Domains: map[string]Domain{"X": {State: "paused"}}}
	if err := Validate(rs); err == nil {
		t.Fatal("expected validation error for unknown state")
	}
}

func TestLoadAbsentFileReturnsDefaults(t *testing.T) {
	dir := t.TempDir()
	rs, err := Load(dir)
	if err != nil {
		t.Fatalf("Load with no rules.json: %v", err)
	}
	if len(rs.Domains) != len(Defaults().Domains) {
		t.Fatalf("absent rules.json should yield defaults, got %d domains", len(rs.Domains))
	}
}

func TestLoadMergesRepoFile(t *testing.T) {
	dir := t.TempDir()
	writeRepoRules(t, dir, `{"schema_version":1,"disabled":false,"domains":{"ROADMAP":{"state":"dormant"}}}`)
	rs, err := Load(dir)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if rs.Domains["ROADMAP"].State != StateDormant {
		t.Fatal("repo override not applied")
	}
	if len(rs.Domains["ROADMAP"].Rules) == 0 {
		t.Fatal("per-field merge dropped default rules")
	}
}

func TestLoadMalformedFileFailsClosed(t *testing.T) {
	dir := t.TempDir()
	writeRepoRules(t, dir, `{ this is not json `)
	if _, err := Load(dir); err == nil {
		t.Fatal("malformed rules.json must fail closed, not silently fall back")
	}
}

func TestRenderContainsDomainsAndHeader(t *testing.T) {
	rs := Defaults()
	out := Render(rs.Match("commit and push"))
	if out == "" {
		t.Fatal("render produced nothing for a match")
	}
	if !contains(out, "COMMITTING") {
		t.Fatalf("render missing domain name:\n%s", out)
	}
}

func TestRenderEmptyIsZeroBytes(t *testing.T) {
	if out := Render(nil); out != "" {
		t.Fatalf("no-match render must be zero bytes (D3), got %q", out)
	}
}

func TestSignatureStableAndDistinct(t *testing.T) {
	rs := Defaults()
	commit := pick(rs, "COMMITTING")
	docs := pick(rs, "DOCUMENTATION")
	if Signature(commit) != Signature(commit) {
		t.Fatal("signature not stable")
	}
	if Signature(commit) == Signature(docs) {
		t.Fatal("distinct domains share a signature")
	}
	// Content drift changes the signature.
	drift := commit
	drift.Rules = append([]string{"a new rule"}, drift.Rules...)
	if Signature(drift) == Signature(commit) {
		t.Fatal("signature did not change on content drift")
	}
}

func TestActiveExcludesDormantAndKillSwitch(t *testing.T) {
	rs := Defaults()
	full := len(rs.Active())
	if full != len(rs.Domains) {
		t.Fatalf("Active() = %d, want all %d default domains", full, len(rs.Domains))
	}
	d := rs.Domains["PII"]
	d.State = StateDormant
	rs.Domains["PII"] = d
	if got := len(rs.Active()); got != full-1 {
		t.Fatalf("dormant domain still active: %d", got)
	}
	if has(rs.Active(), "PII") {
		t.Fatal("Active() returned a dormant domain")
	}
	rs.Disabled = true
	if got := rs.Active(); got != nil {
		t.Fatalf("kill switch: Active() = %v, want nil", names(got))
	}
}

func TestLookup(t *testing.T) {
	rs := Defaults()
	if _, ok := rs.Lookup("NOSUCH"); ok {
		t.Fatal("Lookup returned ok for an absent domain")
	}
	rd, ok := rs.Lookup("PII")
	if !ok || rd.Name != "PII" || len(rd.Rules) == 0 {
		t.Fatalf("Lookup(PII) = %+v ok=%v", rd, ok)
	}
}

func pick(rs RuleSet, name string) ResolvedDomain {
	return ResolvedDomain{Name: name, Domain: rs.Domains[name]}
}

func writeRepoRules(t *testing.T, dir, body string) {
	t.Helper()
	abcd := filepath.Join(dir, ".abcd")
	if err := os.MkdirAll(abcd, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(abcd, "rules.json"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

func contains(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
