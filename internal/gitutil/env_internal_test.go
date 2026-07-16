package gitutil

import (
	"strings"
	"testing"
)

// lastEnvValue returns the value git would resolve for key from an env slice:
// the LAST occurrence wins, matching how a child process reads its environment.
func lastEnvValue(env []string, key string) (string, bool) {
	val := ""
	found := false
	for _, kv := range env {
		if i := strings.IndexByte(kv, '='); i >= 0 && kv[:i] == key {
			val = kv[i+1:]
			found = true
		}
	}
	return val, found
}

// TestGitEnvPinsLocale asserts the isolated git environment pins LC_ALL=C and
// LANG=C so a translated git cannot alter the English chrome (e.g. the
// "N files changed" shortstat) the graveyard's rewrite signal parses — the
// cross-host determinism invariant. The pin must WIN over an ambient value.
func TestGitEnvPinsLocale(t *testing.T) {
	t.Setenv("LC_ALL", "fr_FR.UTF-8")
	t.Setenv("LANG", "fr_FR.UTF-8")

	env := gitEnv()

	if v, ok := lastEnvValue(env, "LC_ALL"); !ok || v != "C" {
		t.Errorf("resolved LC_ALL = %q (found=%v), want C", v, ok)
	}
	if v, ok := lastEnvValue(env, "LANG"); !ok || v != "C" {
		t.Errorf("resolved LANG = %q (found=%v), want C", v, ok)
	}
}
