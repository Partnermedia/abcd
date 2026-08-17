package repolint_test

import (
	"encoding/json"
	"testing"

	"github.com/REPPL/abcd-cli/internal/core/repolint"
)

// fakeRule is a test double: it declares meta and returns canned findings, so
// the evaluator can be exercised without any real filesystem rule.
type fakeRule struct {
	meta     repolint.RuleMeta
	where    func(repolint.Context) bool
	findings []repolint.Finding
	err      error
}

func (r fakeRule) Meta() repolint.RuleMeta { return r.meta }
func (r fakeRule) Where(ctx repolint.Context) bool {
	if r.where == nil {
		return true
	}
	return r.where(ctx)
}
func (r fakeRule) Eval(repolint.Context) ([]repolint.Finding, error) { return r.findings, r.err }

func meta(id string, sev repolint.Severity) repolint.RuleMeta {
	return repolint.RuleMeta{ID: id, Severity: sev, Fix: "do the thing", PolicyInfo: "because"}
}

// A clean repo (every rule returns no findings) exits 0.
func TestEvaluateCleanExitsZero(t *testing.T) {
	res, err := repolint.Evaluate([]repolint.Rule{
		fakeRule{meta: meta("a", repolint.SeverityError)},
		fakeRule{meta: meta("b", repolint.SeverityWarn)},
	}, repolint.Context{})
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Findings) != 0 {
		t.Errorf("findings = %d, want 0", len(res.Findings))
	}
	if res.ExitCode != 0 {
		t.Errorf("exit = %d, want 0", res.ExitCode)
	}
}

// Warnings but no errors exits 1 (Conftest tri-state).
func TestEvaluateWarningsExitsOne(t *testing.T) {
	res, err := repolint.Evaluate([]repolint.Rule{
		fakeRule{meta: meta("w", repolint.SeverityWarn), findings: []repolint.Finding{
			{RuleID: "w", Severity: repolint.SeverityWarn, Message: "m"},
		}},
	}, repolint.Context{})
	if err != nil {
		t.Fatal(err)
	}
	if res.Warnings != 1 || res.Blockers != 0 {
		t.Errorf("blockers=%d warnings=%d, want 0/1", res.Blockers, res.Warnings)
	}
	if res.ExitCode != 1 {
		t.Errorf("exit = %d, want 1", res.ExitCode)
	}
}

// Any error exits 2, even when warnings are also present.
func TestEvaluateErrorExitsTwo(t *testing.T) {
	res, err := repolint.Evaluate([]repolint.Rule{
		fakeRule{meta: meta("e", repolint.SeverityError), findings: []repolint.Finding{
			{RuleID: "e", Severity: repolint.SeverityError, Message: "boom"},
		}},
		fakeRule{meta: meta("w", repolint.SeverityWarn), findings: []repolint.Finding{
			{RuleID: "w", Severity: repolint.SeverityWarn, Message: "meh"},
		}},
	}, repolint.Context{})
	if err != nil {
		t.Fatal(err)
	}
	if res.Blockers != 1 || res.Warnings != 1 {
		t.Errorf("blockers=%d warnings=%d, want 1/1", res.Blockers, res.Warnings)
	}
	if res.ExitCode != 2 {
		t.Errorf("exit = %d, want 2", res.ExitCode)
	}
}

// A rule whose severity is Off is skipped entirely — never evaluated.
func TestEvaluateOffRuleSkipped(t *testing.T) {
	evaluated := false
	res, err := repolint.Evaluate([]repolint.Rule{
		fakeRule{meta: meta("x", repolint.SeverityOff), findings: []repolint.Finding{
			{RuleID: "x", Severity: repolint.SeverityError, Message: "should never appear"},
		}, where: func(repolint.Context) bool { evaluated = true; return true }},
	}, repolint.Context{})
	if err != nil {
		t.Fatal(err)
	}
	if evaluated {
		t.Error("an Off rule must not be evaluated (Where was called)")
	}
	if len(res.Findings) != 0 || res.ExitCode != 0 {
		t.Errorf("Off rule leaked findings=%d exit=%d", len(res.Findings), res.ExitCode)
	}
}

// A rule whose Where predicate is false is skipped, not failed — the "docs/
// absent => docs-currency skipped" acceptance shape.
func TestEvaluateWhereFalseSkips(t *testing.T) {
	res, err := repolint.Evaluate([]repolint.Rule{
		fakeRule{
			meta:     meta("cond", repolint.SeverityError),
			where:    func(repolint.Context) bool { return false },
			findings: []repolint.Finding{{RuleID: "cond", Severity: repolint.SeverityError, Message: "unreached"}},
		},
	}, repolint.Context{})
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Findings) != 0 || res.ExitCode != 0 {
		t.Errorf("a Where=false rule was not skipped: findings=%d exit=%d", len(res.Findings), res.ExitCode)
	}
	if len(res.Skipped) != 1 || res.Skipped[0] != "cond" {
		t.Errorf("skipped = %v, want [cond]", res.Skipped)
	}
}

// Findings are sorted deterministically (rule id, then file, then line).
func TestEvaluateSortsFindings(t *testing.T) {
	res, err := repolint.Evaluate([]repolint.Rule{
		fakeRule{meta: meta("z", repolint.SeverityWarn), findings: []repolint.Finding{
			{RuleID: "z", Severity: repolint.SeverityWarn, File: "b.md", Line: 2},
			{RuleID: "z", Severity: repolint.SeverityWarn, File: "a.md", Line: 9},
		}},
		fakeRule{meta: meta("a", repolint.SeverityWarn), findings: []repolint.Finding{
			{RuleID: "a", Severity: repolint.SeverityWarn, File: "z.md", Line: 1},
		}},
	}, repolint.Context{})
	if err != nil {
		t.Fatal(err)
	}
	got := []string{res.Findings[0].RuleID, res.Findings[1].RuleID, res.Findings[2].RuleID}
	want := []string{"a", "z", "z"}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("finding order = %v, want a,z,z", got)
		}
	}
	// within rule z, a.md before b.md
	if res.Findings[1].File != "a.md" {
		t.Errorf("intra-rule sort: got %s first, want a.md", res.Findings[1].File)
	}
}

// A finding carrying a severity that is neither error nor warn is a rule bug: it
// would serialize into findings yet count as neither blocker nor warning,
// yielding a clean exit alongside a non-empty findings list. Evaluate fails
// closed on it rather than emit that contradiction.
func TestEvaluateRejectsUnknownFindingSeverity(t *testing.T) {
	_, err := repolint.Evaluate([]repolint.Rule{
		fakeRule{meta: meta("x", repolint.SeverityWarn), findings: []repolint.Finding{
			{RuleID: "x", Severity: repolint.Severity("nonsense"), Message: "m"},
		}},
	}, repolint.Context{})
	if err == nil {
		t.Fatal("expected Evaluate to reject a finding with an unknown severity, got nil")
	}
}

// A rule that errors aborts the audit with that error — a broken check must not
// silently read as "clean".
func TestEvaluateRuleErrorPropagates(t *testing.T) {
	_, err := repolint.Evaluate([]repolint.Rule{
		fakeRule{meta: meta("bad", repolint.SeverityError), err: errSentinel},
	}, repolint.Context{})
	if err == nil {
		t.Fatal("expected the rule error to propagate, got nil")
	}
}

// The JSON serializer emits {"findings": []} for a clean repo (acceptance
// criterion) and stable rule ids for a dirty one.
func TestJSONSerializerCleanRepo(t *testing.T) {
	res := repolint.Result{} // no findings
	out, err := repolint.JSONSerializer{}.Serialize(res)
	if err != nil {
		t.Fatal(err)
	}
	var decoded struct {
		Findings []any `json:"findings"`
	}
	if err := json.Unmarshal(out, &decoded); err != nil {
		t.Fatalf("output is not valid JSON: %v\n%s", err, out)
	}
	if decoded.Findings == nil {
		t.Error(`clean repo must emit "findings": [] (a present empty array), not null`)
	}
	if len(decoded.Findings) != 0 {
		t.Errorf("clean repo findings = %d, want 0", len(decoded.Findings))
	}
}

// The JSON always carries a present "skipped" array, even when nothing was
// skipped — the command doc promises { findings, skipped }, so neither key may
// vanish.
func TestJSONSerializerSkippedAlwaysPresent(t *testing.T) {
	out, err := repolint.JSONSerializer{}.Serialize(repolint.Result{})
	if err != nil {
		t.Fatal(err)
	}
	var decoded map[string]json.RawMessage
	if err := json.Unmarshal(out, &decoded); err != nil {
		t.Fatal(err)
	}
	raw, ok := decoded["skipped"]
	if !ok {
		t.Fatalf(`clean result must still emit a "skipped" key: %s`, out)
	}
	if string(raw) != "[]" {
		t.Errorf(`empty "skipped" must be [], got %s`, raw)
	}
}

func TestJSONSerializerCarriesRuleID(t *testing.T) {
	res := repolint.Result{Findings: []repolint.Finding{
		{RuleID: "three-tier-layout", Severity: repolint.SeverityError, File: ".abcd", Line: 0, Message: "missing work/"},
	}}
	out, err := repolint.JSONSerializer{}.Serialize(res)
	if err != nil {
		t.Fatal(err)
	}
	var decoded struct {
		Findings []struct {
			RuleID   string `json:"ruleId"`
			Severity string `json:"severity"`
		} `json:"findings"`
	}
	if err := json.Unmarshal(out, &decoded); err != nil {
		t.Fatal(err)
	}
	if len(decoded.Findings) != 1 || decoded.Findings[0].RuleID != "three-tier-layout" {
		t.Fatalf("ruleId not carried: %s", out)
	}
	if decoded.Findings[0].Severity != "error" {
		t.Errorf("severity serialized as %q, want error", decoded.Findings[0].Severity)
	}
}

var errSentinel = &sentinelErr{}

type sentinelErr struct{}

func (*sentinelErr) Error() string { return "sentinel" }
