package rules

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestInjectFirstTurnRenders(t *testing.T) {
	rs := Defaults()
	res := Inject(rs, "commit and push", SessionState{}, 0)
	if res.Text == "" || !strings.Contains(res.Text, "COMMITTING") {
		t.Fatalf("first turn did not inject COMMITTING:\n%s", res.Text)
	}
	if len(res.Injected) == 0 {
		t.Fatal("Injected list empty on a match")
	}
	if res.State.Count != 1 {
		t.Fatalf("count = %d, want 1", res.State.Count)
	}
}

func TestInjectDedupsWithinSession(t *testing.T) {
	rs := Defaults()
	first := Inject(rs, "commit and push", SessionState{}, 0)
	second := Inject(rs, "commit and push again", first.State, 0)
	if second.Text != "" {
		t.Fatalf("second identical-domain turn re-injected:\n%s", second.Text)
	}
	if len(second.Injected) != 0 {
		t.Fatalf("dedup failed: %v", second.Injected)
	}
}

func TestInjectReinjectsOnClearedLedger(t *testing.T) {
	rs := Defaults()
	first := Inject(rs, "commit", SessionState{}, 0)
	if first.Text == "" {
		t.Fatal("first turn injected nothing")
	}
	// Same session, threaded state: dedup suppresses.
	second := Inject(rs, "commit", first.State, 0)
	if second.Text != "" {
		t.Fatalf("dedup failed on threaded state:\n%s", second.Text)
	}
	// A cleared ledger (what a reset produces) re-injects, even though the prompt
	// is unchanged — this is the reset/refresh contract at the unit level.
	third := Inject(rs, "commit", SessionState{}, 0)
	if third.Text == "" {
		t.Fatal("cleared ledger did not re-inject")
	}
}

func TestLoadBackstop(t *testing.T) {
	// Absent config -> default.
	if got := LoadBackstop(t.TempDir()); got != DefaultRefreshBackstop {
		t.Fatalf("absent config: backstop = %d, want %d", got, DefaultRefreshBackstop)
	}
	// A valid config value is honoured.
	dir := t.TempDir()
	writeConfig(t, dir, `{"rules":{"force_refresh_every_n":7}}`)
	if got := LoadBackstop(dir); got != 7 {
		t.Fatalf("config value not read: %d", got)
	}
	// A non-positive or malformed value falls back to the default.
	bad := t.TempDir()
	writeConfig(t, bad, `{"rules":{"force_refresh_every_n":0}}`)
	if got := LoadBackstop(bad); got != DefaultRefreshBackstop {
		t.Fatalf("non-positive value not defaulted: %d", got)
	}
}

func TestPruneStateRemovesStaleOnly(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("ABCD_RULES_STATE_DIR", dir)
	if err := SaveState("fresh", SessionState{Count: 1}); err != nil {
		t.Fatal(err)
	}
	if err := SaveState("old", SessionState{Count: 1}); err != nil {
		t.Fatal(err)
	}
	// Age the "old" session's file well past the TTL.
	oldFile := sessionFile("old")
	past := time.Now().Add(-StateTTL - time.Hour)
	if err := os.Chtimes(oldFile, past, past); err != nil {
		t.Fatal(err)
	}
	PruneState(StateTTL)
	if _, err := os.Stat(oldFile); !os.IsNotExist(err) {
		t.Fatal("stale state file survived prune")
	}
	if LoadState("fresh").Count != 1 {
		t.Fatal("fresh state file was pruned")
	}
}

func writeConfig(t *testing.T, dir, body string) {
	t.Helper()
	abcd := filepath.Join(dir, ".abcd")
	if err := os.MkdirAll(abcd, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(abcd, "config.json"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestInjectContentDriftReinjects(t *testing.T) {
	rs := Defaults()
	first := Inject(rs, "commit", SessionState{}, 0)
	// Mutate COMMITTING's rules; the signature changes, so it must re-inject
	// despite the ledger entry.
	d := rs.Domains["COMMITTING"]
	d.Rules = append([]string{"a brand new rule"}, d.Rules...)
	rs.Domains["COMMITTING"] = d
	second := Inject(rs, "commit", first.State, 0)
	if !strings.Contains(second.Text, "a brand new rule") {
		t.Fatalf("content drift did not re-inject:\n%s", second.Text)
	}
}

func TestInjectBackstopForcesRefresh(t *testing.T) {
	rs := Defaults()
	st := SessionState{}
	var lastText string
	// Backstop of 3: turns 1..2 dedup after the first inject; turn 3 (count%3==0)
	// clears the ledger and re-injects.
	for i := 0; i < 3; i++ {
		res := Inject(rs, "commit", st, 3)
		st = res.State
		lastText = res.Text
	}
	if lastText == "" {
		t.Fatal("backstop turn did not force a re-inject")
	}
}

func TestInjectNoMatchZeroBytes(t *testing.T) {
	rs := Defaults()
	res := Inject(rs, "paint a landscape in oils", SessionState{}, 0)
	if res.Text != "" {
		t.Fatalf("no-match must render zero bytes, got %q", res.Text)
	}
}

func TestInjectNeverReflectsPromptBytes(t *testing.T) {
	rs := Defaults()
	marker := "ZZUNIQUEPROMPTMARKERZZ commit"
	res := Inject(rs, marker, SessionState{}, 0)
	if strings.Contains(res.Text, "ZZUNIQUEPROMPTMARKERZZ") {
		t.Fatalf("injected text reflected prompt bytes:\n%s", res.Text)
	}
}

func TestSessionStateRoundTripAndReset(t *testing.T) {
	t.Setenv("ABCD_RULES_STATE_DIR", t.TempDir())
	const sid = "session-abc/../../etc" // traversal attempt; must be neutralised
	if got := LoadState(sid); got.Count != 0 {
		t.Fatal("fresh session should load zero state")
	}
	want := SessionState{Count: 4, Ledger: map[string]string{"COMMITTING": "deadbeef"}}
	if err := SaveState(sid, want); err != nil {
		t.Fatalf("SaveState: %v", err)
	}
	got := LoadState(sid)
	if got.Count != 4 || got.Ledger["COMMITTING"] != "deadbeef" {
		t.Fatalf("round-trip mismatch: %+v", got)
	}
	if err := ResetState(sid); err != nil {
		t.Fatalf("ResetState: %v", err)
	}
	if LoadState(sid).Count != 0 {
		t.Fatal("state survived reset")
	}
	// A second reset on an absent file is not an error.
	if err := ResetState(sid); err != nil {
		t.Fatalf("reset of absent state errored: %v", err)
	}
}

func TestSessionFileStaysInStateDir(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("ABCD_RULES_STATE_DIR", dir)
	// A traversal-flavoured id must still hash to a file directly inside dir.
	f := sessionFile("../../../../etc/passwd")
	if got := strings.TrimSuffix(f[:len(dir)], "/"); got != strings.TrimSuffix(dir, "/") {
		t.Fatalf("session file escaped state dir: %s", f)
	}
	if strings.Contains(f[len(dir):], "..") {
		t.Fatalf("session file path contains traversal: %s", f)
	}
}
