package cli

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/REPPL/abcd-cli/internal/core/history"
)

// sessionEndRepo builds an isolated git repo with one commit (so it has a
// root-commit SHA, the history store's key) and a hermetic ~/.abcd history store
// keyed on it. Returns the repo dir and its root SHA.
func sessionEndRepo(t *testing.T) (repo, rootSHA string) {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not on PATH")
	}
	repo = t.TempDir()
	env := append(os.Environ(),
		"GIT_CONFIG_GLOBAL=/dev/null", "GIT_CONFIG_NOSYSTEM=1",
		"GIT_AUTHOR_NAME=t", "GIT_AUTHOR_EMAIL=t@e",
		"GIT_COMMITTER_NAME=t", "GIT_COMMITTER_EMAIL=t@e",
	)
	for _, args := range [][]string{
		{"init", "-q"},
		{"commit", "-q", "--allow-empty", "-m", "root"},
	} {
		cmd := exec.Command("git", args...)
		cmd.Dir = repo
		cmd.Env = env
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v: %s", args, err, out)
		}
	}
	cmd := exec.Command("git", "rev-list", "--max-parents=0", "HEAD")
	cmd.Dir = repo
	cmd.Env = env
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("git rev-list: %v", err)
	}
	rootSHA = strings.TrimSpace(string(out))

	// Hermetic store: HOME drives ~/.abcd/history. Capture requires the
	// transcripts dir to exist already (abcd install creates it).
	home := t.TempDir()
	t.Setenv("HOME", home)
	if err := os.MkdirAll(filepath.Join(home, ".abcd", "history", rootSHA, "transcripts"), 0o755); err != nil {
		t.Fatal(err)
	}
	return repo, rootSHA
}

// endPayload is the SessionEnd-hook JSON the harness writes to the verb's stdin.
func endPayload(t *testing.T, session, cwd, transcript string) string {
	t.Helper()
	b, err := json.Marshal(map[string]string{
		"session_id":      session,
		"cwd":             cwd,
		"transcript_path": transcript,
		"hook_event_name": "SessionEnd",
	})
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

// TestHookSessionEndCapturesTranscript is the milestone: finishing a session
// leaves a record in the store. Without this hook the transcript corpus never
// accrues, and no later code can recover a session that was not captured while
// it ran.
func TestHookSessionEndCapturesTranscript(t *testing.T) {
	repo, rootSHA := sessionEndRepo(t)
	tp := filepath.Join(t.TempDir(), "sess.jsonl")
	if err := os.WriteFile(tp, []byte(`{"role":"user","text":"hello"}`+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	_, errlog := runHook(t, endPayload(t, "sess-1", repo, tp), "hook", "session-end")

	recs, err := history.List(rootSHA)
	if err != nil {
		t.Fatalf("history.List: %v (stderr: %s)", err, errlog)
	}
	if len(recs) != 1 {
		t.Fatalf("want 1 stored transcript, got %d (stderr: %s)", len(recs), errlog)
	}
	if recs[0].SessionID != "sess-1" {
		t.Errorf("session id = %q, want sess-1", recs[0].SessionID)
	}
}

// TestHookSessionEndIsIdempotent holds the re-capture property: the same
// transcript captured twice stores one record. A SessionEnd hook can fire more than
// once for a session, and the corpus must not grow a duplicate each time.
func TestHookSessionEndIsIdempotent(t *testing.T) {
	repo, rootSHA := sessionEndRepo(t)
	tp := filepath.Join(t.TempDir(), "sess.jsonl")
	if err := os.WriteFile(tp, []byte(`{"role":"user","text":"hello"}`+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	in := endPayload(t, "sess-2", repo, tp)

	runHook(t, in, "hook", "session-end")
	runHook(t, in, "hook", "session-end")

	recs, err := history.List(rootSHA)
	if err != nil {
		t.Fatal(err)
	}
	if len(recs) != 1 {
		t.Fatalf("re-capture must be a no-op: want 1 record, got %d", len(recs))
	}
}

// TestHookSessionEndOneRecordPerSession is why this is wired to SessionEnd, not
// Stop. Stop fires once per assistant *turn*, and a live transcript grows
// between turns, so wiring Stop would store a fresh, larger superset every turn
// — Capture's sha256 dedup only collapses byte-identical re-captures. SessionEnd
// fires once at session termination. This test simulates a session that grew
// over several turns and asserts the store holds exactly one record for it.
func TestHookSessionEndOneRecordPerSession(t *testing.T) {
	repo, rootSHA := sessionEndRepo(t)
	tp := filepath.Join(t.TempDir(), "session.jsonl")

	// SessionEnd fires once, against the final, fully-grown transcript.
	body := ""
	for turn := 1; turn <= 5; turn++ {
		body += `{"role":"user","text":"turn ` + string(rune('0'+turn)) + `"}` + "\n"
	}
	if err := os.WriteFile(tp, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	runHook(t, endPayload(t, "grown-session", repo, tp), "hook", "session-end")

	recs, err := history.List(rootSHA)
	if err != nil {
		t.Fatal(err)
	}
	if len(recs) != 1 {
		t.Fatalf("one session must leave one record, got %d — is this wired to Stop (per-turn) instead of SessionEnd?", len(recs))
	}
}

// TestHookSessionEndNeverBlocksTheHost is the fail-closed contract. Every one of
// these is a payload the harness could plausibly hand us, and none may exit
// non-zero or write a record — a SessionEnd hook that errors or hangs wedges the
// user's session, which is a far worse outcome than a missed transcript.
func TestHookSessionEndNeverBlocksTheHost(t *testing.T) {
	cases := []struct {
		name  string
		stdin func(t *testing.T, repo string) string
	}{
		{"malformed json", func(*testing.T, string) string { return "{not json" }},
		{"empty payload", func(*testing.T, string) string { return "" }},
		{"no transcript_path", func(t *testing.T, repo string) string {
			return endPayload(t, "s", repo, "")
		}},
		{"transcript does not exist", func(t *testing.T, repo string) string {
			return endPayload(t, "s", repo, filepath.Join(t.TempDir(), "absent.jsonl"))
		}},
		{"transcript is a directory", func(t *testing.T, repo string) string {
			return endPayload(t, "s", repo, t.TempDir())
		}},
		{"hostile session id", func(t *testing.T, repo string) string {
			tp := filepath.Join(t.TempDir(), "s.jsonl")
			if err := os.WriteFile(tp, []byte("x\n"), 0o644); err != nil {
				t.Fatal(err)
			}
			return endPayload(t, "../../escape", repo, tp)
		}},
		{"cwd is not a git repo", func(t *testing.T, _ string) string {
			tp := filepath.Join(t.TempDir(), "s.jsonl")
			if err := os.WriteFile(tp, []byte("x\n"), 0o644); err != nil {
				t.Fatal(err)
			}
			return endPayload(t, "s", t.TempDir(), tp)
		}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			repo, rootSHA := sessionEndRepo(t)
			// runHook fails the test if the command exits non-zero.
			runHook(t, tc.stdin(t, repo), "hook", "session-end")

			recs, err := history.List(rootSHA)
			if err != nil {
				t.Fatalf("history.List: %v", err)
			}
			if len(recs) != 0 {
				t.Errorf("a rejected payload must write nothing, got %d record(s)", len(recs))
			}
		})
	}
}

// TestHookSessionEndRedactsOnThisPath proves the redaction pass runs on the hook
// path itself, not just when `history capture` is called directly. Automatic
// capture makes this load-bearing: a secret or an absolute home path in a
// transcript must be masked before it lands in the store, and this asserts the
// hook — not some other caller — is what triggers that.
func TestHookSessionEndRedactsOnThisPath(t *testing.T) {
	repo, rootSHA := sessionEndRepo(t)
	home, err := os.UserHomeDir() // the hermetic HOME set by sessionEndRepo
	if err != nil {
		t.Fatal(err)
	}
	tp := filepath.Join(t.TempDir(), "sess.jsonl")
	// A ghp_ PAT-shaped token the bundled scanner matches (\bghp_[A-Za-z0-9]{36,}\b),
	// plus an absolute home path. The token is ASSEMBLED at runtime, never written
	// as a contiguous literal, so no committed source line carries a secret-shaped
	// string for the full-history secret scan to flag — the same discipline the
	// scanner's own fixtures use.
	token := "ghp_" + strings.Repeat("A", 40)
	body := `{"text":"token ` + token + ` and path ` + home + `/secret.env"}` + "\n"
	if err := os.WriteFile(tp, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}

	runHook(t, endPayload(t, "redacts", repo, tp), "hook", "session-end")

	recs, err := history.List(rootSHA)
	if err != nil || len(recs) != 1 {
		t.Fatalf("want 1 record, got %d (err %v)", len(recs), err)
	}
	if recs[0].Secrets == 0 {
		t.Error("the GitHub token was not counted as redacted on the hook path")
	}
	if recs[0].HomePaths == 0 {
		t.Error("the absolute home path was not counted as redacted on the hook path")
	}
	_, stored, err := history.Read(rootSHA, "redacts")
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(stored), token) {
		t.Error("the raw GitHub token survived into the stored record")
	}
	if strings.Contains(string(stored), home+"/secret.env") {
		t.Error("the raw absolute home path survived into the stored record")
	}
}

// TestHistoryCaptureFromSubdirHonoursRepoPiiConfig proves capture resolves the
// git working-tree root, not the process cwd, before loading the per-repo
// redaction override. The scanner looks for .abcd/config/pii.json at the repo
// root only (no upward walk), so a capture run from a subdirectory that passed
// the subdirectory would silently redact with default patterns and let a
// custom-pattern secret land in the store in cleartext (B12).
func TestHistoryCaptureFromSubdirHonoursRepoPiiConfig(t *testing.T) {
	repo, rootSHA := sessionEndRepo(t)

	// A per-repo override adding a pattern the bundled defaults do NOT match.
	cfgDir := filepath.Join(repo, ".abcd", "config")
	if err := os.MkdirAll(cfgDir, 0o755); err != nil {
		t.Fatal(err)
	}
	cfg := `{"patterns":{"acme_secret":{"regex":"ACME-WIDGET-[0-9]{6}","kind":"token","label":"acme secret","severity":"hard_fail"}}}`
	if err := os.WriteFile(filepath.Join(cfgDir, "pii.json"), []byte(cfg), 0o644); err != nil {
		t.Fatal(err)
	}

	// Run capture from a subdirectory of the repo.
	sub := filepath.Join(repo, "internal", "deep")
	if err := os.MkdirAll(sub, 0o755); err != nil {
		t.Fatal(err)
	}
	token := "ACME-WIDGET-123456"
	tp := filepath.Join(t.TempDir(), "sess.jsonl")
	if err := os.WriteFile(tp, []byte(`{"text":"secret `+token+`"}`+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Chdir(sub)

	runCLI(t, "history", "capture", tp, "--session", "subdir-cfg")

	recs, err := history.List(rootSHA)
	if err != nil || len(recs) != 1 {
		t.Fatalf("want 1 record, got %d (err %v)", len(recs), err)
	}
	if recs[0].Secrets == 0 {
		t.Error("the custom-pattern secret was not redacted — capture used the subdirectory, not the repo root, for pii.json (B12)")
	}
	_, stored, err := history.Read(rootSHA, "subdir-cfg")
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(stored), token) {
		t.Error("the raw custom-pattern secret survived into the stored record (B12)")
	}
}

// TestHookSessionEndWritesNothingToStdout keeps the SessionEnd hook silent on the
// model-facing stream. Diagnostics belong on stderr, out of band — the same rule
// the prompt-router follows.
func TestHookSessionEndWritesNothingToStdout(t *testing.T) {
	repo, _ := sessionEndRepo(t)
	tp := filepath.Join(t.TempDir(), "sess.jsonl")
	if err := os.WriteFile(tp, []byte("hello\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	stdout, _ := runHook(t, endPayload(t, "sess-3", repo, tp), "hook", "session-end")
	if stdout != "" {
		t.Errorf("session-end must produce zero stdout, got %q", stdout)
	}
}
