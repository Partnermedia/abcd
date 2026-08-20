package cli

import (
	"strings"
	"testing"
)

// An over-cap hook payload must be reported as over the cap, not as malformed
// JSON: readHookInput used to truncate at the cap, hand json.Unmarshal a
// severed prefix, and blame the host ("unexpected end of JSON input") for a
// size limit abcd imposed (iss-201's class at the shared helper).
func TestHookPromptRouterReportsAnOverCapPayload(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	repo := t.TempDir()
	pad := strings.Repeat("a", maxHookStdinBytes) // payload = pad + JSON framing > cap
	_, errlog := runHook(t, hookInputJSON(t, "s", repo, pad), "hook", "prompt-router")
	if !strings.Contains(errlog, "over the") || !strings.Contains(errlog, "cap") {
		t.Errorf("an over-cap payload must name the cap on stderr, got: %s", errlog)
	}
	if strings.Contains(errlog, "unexpected end of JSON") {
		t.Errorf("an over-cap payload must not be misreported as malformed JSON: %s", errlog)
	}
}

// The boundary case: a payload at exactly the cap still parses and routes.
func TestHookPromptRouterAcceptsAPayloadAtTheCap(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	repo := t.TempDir()
	probe := hookInputJSON(t, "s", repo, "")
	pad := strings.Repeat("a", maxHookStdinBytes-len(probe))
	in := hookInputJSON(t, "s", repo, pad)
	if len(in) > maxHookStdinBytes {
		t.Fatalf("fixture overshot the cap by %d bytes", len(in)-maxHookStdinBytes)
	}
	_, errlog := runHook(t, in, "hook", "prompt-router")
	if strings.Contains(errlog, "unreadable hook payload") {
		t.Errorf("an at-cap payload must parse, got: %s", errlog)
	}
}

// The guard hook shares the defect (iss-201's original site): over-cap must
// fail open naming the cap, not blaming the host's JSON.
func TestGuardHookReportsAnOverCapPayload(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	pad := strings.Repeat("a", maxHookStdinBytes)
	_, stderr, code := runGuard(`{"tool_input":{"command":"`+pad+`"}}`, "guard", "hook")
	if code != 1 {
		t.Fatalf("guard hook over-cap must fail open with exit 1, got %d", code)
	}
	if !strings.Contains(stderr, "over the") || !strings.Contains(stderr, "cap") {
		t.Errorf("over-cap guard payload must name the cap, got: %s", stderr)
	}
	if strings.Contains(stderr, "not readable JSON") {
		t.Errorf("over-cap guard payload misreported as malformed JSON: %s", stderr)
	}
}
