package intent

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/intentdriven/abcd/internal/gittest"
)

// identityRepo is a temporary repository the privacy scanner can build an
// identity from. The repo-local user name is where the hard-fail real_name
// pattern comes from; without it the detector has nothing to detect and a
// redaction test would pass on silence. The calls go through gittest.Env so they
// read no ambient configuration (iss-28).
func identityRepo(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	for _, args := range [][]string{
		{"init"},
		{"config", "user.name", "Jonathan Kensington-Pryce"},
		{"config", "user.email", "jkp@example.com"},
	} {
		cmd := exec.Command("git", args...)
		cmd.Dir = root
		cmd.Env = gittest.Env(t)
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("%v: %v\n%s", args, err, out)
		}
	}
	return root
}

// leakyVerdict is a schema-valid verdict whose every free-text field carries the
// three shapes the privacy scanner is armed for: a third party's absolute home
// path, a LAN hostname and a person's name.
func leakyVerdict(t *testing.T, rcp, conditionID string) string {
	t.Helper()
	const (
		path = "/Users/zzotherperson/checkouts/abcd/internal/core/intent/audit.go"
		host = "buildbox.local"
		who  = "Jonathan Kensington-Pryce"
	)
	v := map[string]any{
		"_type":      "abcd/intent-fidelity-verdict/v1",
		"receipt_id": rcp,
		"verifier":   map[string]any{"id": "intent-fidelity-reviewer", "version": "claude-opus-4-8"},
		"policy":     map[string]any{"rubric_hash": "sha256:aa", "prompt_hash": "sha256:bb"},
		"criteria": []any{
			map[string]any{
				"criterion_id": "ac-1",
				"verdict":      "MET",
				"rationale":    "checked against " + path + " on " + host + " with " + who,
				"evidence": []any{
					map[string]any{
						"ref":   path + ":41",
						"quote": "reviewed by " + who + " on " + host,
					},
				},
			},
		},
		"acceptance_rollup": map[string]any{"MET": 1, "MET_WITH_CONCERNS": 0, "NOT_MET": 0, "INCONCLUSIVE": 0},
		"gap_audit": map[string]any{
			"honoured": []any{
				map[string]any{
					"claim": "the marker is parked, per " + who + " reading " + path,
					"evidence": []any{
						map[string]any{"ref": path + ":2", "quote": "run on " + host},
					},
				},
			},
			"diverged": []any{},
			"missing":  []any{},
		},
		"scope_conditions": []any{
			map[string]any{
				"condition_id": conditionID,
				"disposition":  "narrowed",
				"rationale":    "narrowed after " + who + " ran it on " + host,
				"narrowing":    "holds only under " + path,
				"evidence": []any{
					map[string]any{"ref": path + ":9", "quote": "checked on " + host},
				},
			},
		},
	}
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

// assertNoLeak fails on any of the three identity shapes surviving into a
// committed record, and insists the redaction actually happened rather than the
// field having been dropped.
func assertNoLeak(t *testing.T, body string) {
	t.Helper()
	for _, leak := range []string{"zzotherperson", "buildbox.local", "Kensington-Pryce"} {
		if strings.Contains(body, leak) {
			t.Errorf("agent prose carrying %q reached the committed record:\n%s", leak, body)
		}
	}
	for _, mask := range []string{"[redacted-path]", "[redacted-hostname]", "[redacted-name]"} {
		if !strings.Contains(body, mask) {
			t.Errorf("the record carries no %s, so the field was dropped rather than redacted:\n%s", mask, body)
		}
	}
}

// The intent-audit ingest writes AGENT-PRODUCED prose into a committed record,
// and applied no privacy scanner to it: a rationale, a narrowing or an evidence
// reference carrying an absolute home path, a hostname or a person's name landed
// verbatim in the shipped intent, with only the committed-file privacy lint
// downstream to catch it (iss-2608300924205748).
//
// framework 7.1 lists "Audit Notes (verdict from `intent audit ingest`)" as a
// field of the intent record — "a prior verdict is revision history" — so what
// the ingest writes is durable committed material, exactly what AGENTS.md's
// privacy rule governs. Every free-text field of the verdict goes through the
// canonical primitive before the write; the structural fields (ids, enum
// verdicts, hashes) are validated and carry nothing to redact.
func TestIngestVerdictRedactsAgentProse(t *testing.T) {
	root := identityRepo(t)
	const rcp = "rcp-0123456789ab"
	writeFile(t, root, shippedDir+"/itd-10-alpha.md",
		strings.Replace(shippedWithMarker("itd-10", "alpha", "spc-1", "OWED", rcp),
			"## Scope Conditions\n\n"+NullityToken,
			"## Scope Conditions\n\n- the deployment is single-tenant <!-- cond: cond-2609021016272867 -->", 1))

	res, err := IngestVerdict(root, writeVerdict(t, root, leakyVerdict(t, rcp, "cond-2609021016272867")))
	if err != nil {
		t.Fatalf("IngestVerdict: %v", err)
	}
	if res.Status != "ingested" {
		t.Fatalf("status = %q, want ingested", res.Status)
	}

	body, err := os.ReadFile(filepath.Join(root, shippedDir, "itd-10-alpha.md"))
	if err != nil {
		t.Fatal(err)
	}
	assertNoLeak(t, string(body))

	// The verdict still MEANS what it meant: only what reaches disk changed.
	for _, want := range []string{"ac-1 — MET", "cond-2609021016272867 — narrowed", "narrowing:", "honoured", "evidence:"} {
		if !strings.Contains(string(body), want) {
			t.Errorf("the ingested block no longer carries %q:\n%s", want, body)
		}
	}
}

// The dead-letter path writes the same kind of prose from the same untrusted
// payload — its `reason` is derived from the verdict's own content — so it is
// redacted on the same terms. A quarantine that leaks is still a leak.
func TestDeadLetterRedactsAgentProse(t *testing.T) {
	root := identityRepo(t)
	const rcp = "rcp-0123456789ab"
	writeFile(t, root, shippedDir+"/itd-10-alpha.md", shippedWithMarker("itd-10", "alpha", "spc-1", "OWED", rcp))

	// An out-of-enum verdict token quarantines the payload and quotes the token
	// back in the reason, which is where the untrusted prose gets in.
	var m map[string]any
	if err := json.Unmarshal([]byte(leakyVerdict(t, rcp, "cond-2609021016272867")), &m); err != nil {
		t.Fatal(err)
	}
	crit := m["criteria"].([]any)[0].(map[string]any)
	crit["verdict"] = "MET on buildbox.local per Jonathan Kensington-Pryce at /Users/zzotherperson/x"
	raw, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		t.Fatal(err)
	}

	res, err := IngestVerdict(root, writeVerdict(t, root, string(raw)))
	if err != nil {
		t.Fatalf("IngestVerdict: %v", err)
	}
	if res.Status != "dead_letter" {
		t.Fatalf("status = %q, want dead_letter", res.Status)
	}
	body, err := os.ReadFile(filepath.Join(root, shippedDir, "itd-10-alpha.md"))
	if err != nil {
		t.Fatal(err)
	}
	assertNoLeak(t, string(body))
}
