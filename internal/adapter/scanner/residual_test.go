package scanner

import "testing"

// TestCallerHomeTrimsTheTrailingSlash pins the one normalisation every store
// relies on when it sweeps the literal home: a HOME exported with a trailing
// slash resolves to the same string as one without, so "$HOME/x" matches
// either way.
func TestCallerHomeTrimsTheTrailingSlash(t *testing.T) {
	t.Setenv("HOME", "/base/zzhomeuser42/")
	if got := CallerHome(); got != "/base/zzhomeuser42" {
		t.Errorf("CallerHome() = %q, want the trailing slash trimmed", got)
	}
}

// TestBlockingResidualGatesOnKindNotSeverityAlone pins the stage-two rule the
// stores share: a warn-severity identity or network span refuses the write, a
// warn-severity span of any other kind does not, and a hard_fail span always
// does.
func TestBlockingResidualGatesOnKindNotSeverityAlone(t *testing.T) {
	findings := []Finding{
		{Kind: kindNetLANHost, Severity: SeverityWarn},
		{Kind: "generic:warn_only", Severity: SeverityWarn},
		{Kind: "generic:token", Severity: SeverityHardFail},
	}
	got := BlockingResidual(findings)
	if len(got) != 2 {
		t.Fatalf("BlockingResidual kept %d findings, want 2 (the identity span and the hard_fail one): %+v", len(got), got)
	}
	if got[0].Kind != kindNetLANHost || got[1].Kind != "generic:token" {
		t.Errorf("BlockingResidual kept the wrong spans: %+v", got)
	}
	if len(BlockingResidual(nil)) != 0 {
		t.Error("a clean rescan must not block")
	}
}
