package cli

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

// TestVersionJSON proves the CLI -> core -> JSON round-trip the Phase 0 exit
// criterion requires.
func TestVersionJSON(t *testing.T) {
	out := runCLI(t, "version", "--json")

	var got map[string]any
	if err := json.Unmarshal(out, &got); err != nil {
		t.Fatalf("output is not JSON: %v\n%s", err, out)
	}
	if got["name"] != "abcd" {
		t.Fatalf("name = %v, want abcd", got["name"])
	}
	if got["version"] == "" || got["version"] == nil {
		t.Fatalf("version missing: %v", got)
	}
}

func TestVersionText(t *testing.T) {
	out := runCLI(t, "version")
	if !strings.HasPrefix(string(out), "abcd ") {
		t.Fatalf("text output = %q, want it to start with \"abcd \"", out)
	}
}

func TestBareStatusJSON(t *testing.T) {
	out := runCLI(t, "--json")
	var got map[string]any
	if err := json.Unmarshal(out, &got); err != nil {
		t.Fatalf("bare status output is not JSON: %v\n%s", err, out)
	}
	if _, ok := got["dir"]; !ok {
		t.Fatalf("status JSON missing dir: %v", got)
	}
}

func runCLI(t *testing.T, args ...string) []byte {
	t.Helper()
	cmd := NewRootCommand()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs(args)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute %v: %v\n%s", args, err, out.String())
	}
	return out.Bytes()
}
