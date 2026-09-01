package ahoy

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// conflictedConfig is the advisory's recipe: a config.json that carried the
// disclosure fence and the native-scanning opt-out until a merge left a
// conflict marker in it. It does not parse, and it is the user's data.
const conflictedConfig = `{"repo":{"visibility":"public"},"scan":{"native_secret_scanning":false},"meta":{"schema_version":1}
<<<<<<<
`

// conflictedRepo is an adoptable repo whose config.json is conflictedConfig.
func conflictedRepo(t *testing.T) string {
	t.Helper()
	repo := t.TempDir()
	if err := os.Mkdir(filepath.Join(repo, ".git"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(repo, ".abcd"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(configPath(repo), []byte(conflictedConfig), 0o644); err != nil {
		t.Fatal(err)
	}
	return repo
}

// TestInstallNeverRebuildsMalformedConfig is GHSA-mchq-gm34-3j34: a config
// that cannot be parsed is not rewritten by any install step — stepConfigValues
// already refused, stepVersionStamp must not rebuild it as a meta-only file
// behind that refusal — and the partial install says why.
func TestInstallNeverRebuildsMalformedConfig(t *testing.T) {
	setupHermetic(t)
	repo := conflictedRepo(t)

	res, err := Install(repo, installOpts(), RefusingPrompter{})
	if err != nil {
		t.Fatal(err)
	}
	after, err := os.ReadFile(configPath(repo))
	if err != nil {
		t.Fatal(err)
	}
	if string(after) != conflictedConfig {
		t.Fatalf("install rewrote a config.json it could not parse:\n%s", after)
	}
	for _, w := range res.Writes {
		if strings.HasSuffix(w, filepath.Join(".abcd", "config.json")) || strings.HasSuffix(w, "config.json") {
			t.Fatalf("install reported writing config.json: %v", res.Writes)
		}
	}
	said := false
	for _, n := range res.Notes {
		if strings.Contains(n, "config.json") && strings.Contains(n, "could not be parsed") {
			said = true
		}
	}
	if !said {
		t.Fatalf("no note names the unparseable config.json; notes: %v", res.Notes)
	}
	if res.Status != "partial" {
		t.Fatalf("status = %q, want partial (the config is left for the operator to repair)", res.Status)
	}
}

// TestDetectMalformedConfigIsOneDiagnosticGap: detection on an unparseable
// config raises one non-resolvable config.malformed gap and nothing that would
// arm a rewrite (install_meta.missing, config.*_missing).
func TestDetectMalformedConfigIsOneDiagnosticGap(t *testing.T) {
	setupHermetic(t)
	repo := conflictedRepo(t)

	det, err := Detect(repo)
	if err != nil {
		t.Fatal(err)
	}
	var malformed []Gap
	for _, g := range det.Gaps {
		switch g.ID {
		case "config.malformed":
			malformed = append(malformed, g)
		case "install_meta.missing",
			"config.visibility_missing", "config.docs_target_missing",
			"config.oracle_backend_missing", "config.scan_deep_missing":
			t.Errorf("gap %s raised on an unparseable config; it would arm a rewrite of user data", g.ID)
		}
	}
	if len(malformed) != 1 {
		t.Fatalf("config.malformed gaps = %d, want exactly one: %+v", len(malformed), det.Gaps)
	}
	if malformed[0].Resolvable || !malformed[0].Required {
		t.Fatalf("config.malformed must be a required, non-resolvable diagnostic: %+v", malformed[0])
	}
	if !strings.Contains(malformed[0].Detail, "could not be parsed") {
		t.Fatalf("config.malformed detail = %q, want it to say the file could not be parsed", malformed[0].Detail)
	}
}
