package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/REPPL/abcd-cli/internal/core/lifeboat"
)

// ---------------------------------------------------------------------------
// Fixtures — the CLI surface builds its own packed lifeboats by hand (the core
// package's sealLifeboat/oracleFixture helpers are unexported). sealSynthLifeboat
// reproduces the manifest hash the way VerifyManifest reproduces it (hash every
// file except _provenance.json, which is written LAST), so a synthesis lifeboat
// actually verifies and the oracle can reach a SHIP verdict.
// ---------------------------------------------------------------------------

type synthOpts struct {
	sourceName string
	briefPR    bool
	coverage   *lifeboat.Summary
}

// buildSynthLifeboat hand-builds a packed lifeboat carrying one citable ADR (a
// live adr-12 record id + a packed path) and, per opts, the brief press-release
// section and a coverage.json — then seals it. The ADR gives principles a
// deterministic entry and gives a delegated payload a real id to cite.
func buildSynthLifeboat(t *testing.T, opts synthOpts) string {
	t.Helper()
	if opts.sourceName == "" {
		opts.sourceName = "fix"
	}
	dir := filepath.Join(t.TempDir(), "lifeboat")
	adrDir := filepath.Join(dir, "docs", "adrs")
	if err := os.MkdirAll(adrDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(adrDir, "0012-example.md"),
		[]byte("# 12. Example\n\n## Decision\n\n- keep the cascade fixed\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if opts.briefPR {
		pr := filepath.Join(dir, "brief", "01-product")
		if err := os.MkdirAll(pr, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(pr, "01-press-release.md"),
			[]byte("# abcd carries a project's theory across a boundary\n\nA host-agnostic configuration layer for development.\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if opts.coverage != nil {
		cov := lifeboat.Coverage{
			SchemaVersion: lifeboat.SchemaVersion,
			Repo:          lifeboat.RepoInfo{Name: opts.sourceName},
			Summary:       *opts.coverage,
		}
		writeJSON(t, filepath.Join(dir, "coverage.json"), cov)
	}
	sealSynthLifeboat(t, dir, opts.sourceName)
	return dir
}

// sealSynthLifeboat writes _provenance.json whose manifest_sha256 is reproduced
// exactly as VerifyManifest reproduces it (every file except the header, sorted
// and hashed by ManifestSHA256). The header is written LAST so it is never in its
// own hash. No excluded artifacts (synthesis/lessons/audit) exist yet, so hashing
// every non-header file matches the verifier.
func sealSynthLifeboat(t *testing.T, dir, sourceName string) {
	t.Helper()
	var files []lifeboat.PlannedFile
	err := filepath.WalkDir(dir, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(dir, p)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)
		if rel == lifeboat.ProvenanceName {
			return nil
		}
		data, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		files = append(files, lifeboat.PlannedFile{Path: rel, Content: data})
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	h := lifeboat.ManifestSHA256(files)
	prov := fmt.Sprintf(`{"schema_version":%d,"generator":"test","source_name":%q,"manifest_sha256":%q}`,
		lifeboat.SchemaVersion, sourceName, h)
	if err := os.WriteFile(filepath.Join(dir, lifeboat.ProvenanceName), []byte(prov), 0o644); err != nil {
		t.Fatal(err)
	}
}

// synthPayloadFile writes an untrusted synthesis payload to a fresh temp dir and
// returns its path.
func synthPayloadFile(t *testing.T, content string) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), "payload.json")
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return p
}

// realSrcDir returns a real directory usable as the oracle's <source-repo> arg,
// its base name set so it can match (or diverge from) the provenance source_name.
func realSrcDir(t *testing.T, base string) string {
	t.Helper()
	d := filepath.Join(t.TempDir(), base)
	if err := os.MkdirAll(d, 0o755); err != nil {
		t.Fatal(err)
	}
	return d
}

// ---------------------------------------------------------------------------
// principles
// ---------------------------------------------------------------------------

// TestDisembarkPrinciplesDeterministic: no flag → deterministic evidence-only
// principles.json written from the packed ADR, exit 0.
func TestDisembarkPrinciplesDeterministic(t *testing.T) {
	dir := buildSynthLifeboat(t, synthOpts{})
	out := runCLI(t, "disembark", "principles", dir, "--json")
	var res lifeboat.PrinciplesResult
	if err := json.Unmarshal(out, &res); err != nil {
		t.Fatalf("principles --json is not a result: %v\n%s", err, out)
	}
	if res.Mode != lifeboat.ModeDeterministic {
		t.Errorf("mode = %q, want deterministic", res.Mode)
	}
	if res.Written != 1 {
		t.Errorf("Written = %d, want 1", res.Written)
	}
	data, err := os.ReadFile(filepath.Join(dir, "principles.json"))
	if err != nil {
		t.Fatalf("principles.json not written: %v", err)
	}
	if !strings.Contains(string(data), "prn-adr-12") {
		t.Errorf("principles.json missing prn-adr-12:\n%s", data)
	}
}

// TestDisembarkPrinciplesDelegated: a valid delegated payload citing a real
// adr-12 is written, exit 0, mode delegated.
func TestDisembarkPrinciplesDelegated(t *testing.T) {
	dir := buildSynthLifeboat(t, synthOpts{})
	payload := synthPayloadFile(t, `{"schema_version":1,"mode":"delegated","prompt_version":"0.1.0",`+
		`"principles":[{"id":"prn-cascade","principle":"the cascade is fixed","confidence":"high","evidence":["adr-12"]}]}`)
	out := runCLI(t, "disembark", "principles", dir, "--principles-json", payload, "--json")
	var res lifeboat.PrinciplesResult
	if err := json.Unmarshal(out, &res); err != nil {
		t.Fatalf("not a result: %v\n%s", err, out)
	}
	if res.Mode != lifeboat.ModeDelegated || res.Written != 1 {
		t.Errorf("res = %+v; want delegated, 1 written", res)
	}
}

// TestDisembarkPrinciplesStdin: the "-" stdin transport works (one verb suffices).
func TestDisembarkPrinciplesStdin(t *testing.T) {
	dir := buildSynthLifeboat(t, synthOpts{})
	payload := `{"schema_version":1,"mode":"delegated","prompt_version":"0.1.0",` +
		`"principles":[{"id":"prn-x","principle":"y","confidence":"high","evidence":["adr-12"]}]}`
	out := runCLIStdin(t, payload, "disembark", "principles", dir, "--principles-json", "-", "--json")
	var res lifeboat.PrinciplesResult
	if err := json.Unmarshal(out, &res); err != nil {
		t.Fatalf("not a result: %v\n%s", err, out)
	}
	if res.Written != 1 {
		t.Errorf("Written = %d, want 1", res.Written)
	}
}

// TestDisembarkPrinciplesUnknownFieldExit2: a payload with a smuggled unknown
// field is a structural refusal — exit 2, scrubbed (no absolute path leak).
func TestDisembarkPrinciplesUnknownFieldExit2(t *testing.T) {
	dir := buildSynthLifeboat(t, synthOpts{})
	payload := synthPayloadFile(t, `{"schema_version":1,"mode":"delegated","prompt_version":"0.1.0",`+
		`"principles":[],"smuggled":true}`)
	var stdout, stderr bytes.Buffer
	code := Run([]string{"disembark", "principles", dir, "--principles-json", payload}, &stdout, &stderr)
	if code != 2 {
		t.Fatalf("exit = %d, want 2\nstdout:%s\nstderr:%s", code, stdout.String(), stderr.String())
	}
	if strings.Contains(stderr.String(), filepath.Dir(dir)) {
		t.Errorf("error leaked an absolute path: %q", stderr.String())
	}
}

// ---------------------------------------------------------------------------
// press-release
// ---------------------------------------------------------------------------

// TestDisembarkPressReleaseDeterministic: no flag → deterministic press-release.json
// composed from the packed brief, exit 0.
func TestDisembarkPressReleaseDeterministic(t *testing.T) {
	dir := buildSynthLifeboat(t, synthOpts{briefPR: true})
	out := runCLI(t, "disembark", "press-release", dir, "--json")
	var res lifeboat.PressReleaseResult
	if err := json.Unmarshal(out, &res); err != nil {
		t.Fatalf("press-release --json is not a result: %v\n%s", err, out)
	}
	if res.Mode != lifeboat.ModeDeterministic {
		t.Errorf("mode = %q, want deterministic", res.Mode)
	}
	if _, err := os.Stat(filepath.Join(dir, "press-release.json")); err != nil {
		t.Errorf("press-release.json not written: %v", err)
	}
}

// TestDisembarkPressReleaseDelegated: a valid delegated document citing a packed
// brief path replaces the derived one, exit 0, mode delegated.
func TestDisembarkPressReleaseDelegated(t *testing.T) {
	dir := buildSynthLifeboat(t, synthOpts{briefPR: true})
	payload := synthPayloadFile(t, `{"schema_version":1,"mode":"delegated","prompt_version":"0.1.0",`+
		`"headline":"H","body":"B","evidence":["brief/01-product/01-press-release.md"]}`)
	out := runCLI(t, "disembark", "press-release", dir, "--press-release-json", payload, "--json")
	var res lifeboat.PressReleaseResult
	if err := json.Unmarshal(out, &res); err != nil {
		t.Fatalf("not a result: %v\n%s", err, out)
	}
	if res.Mode != lifeboat.ModeDelegated {
		t.Errorf("mode = %q, want delegated", res.Mode)
	}
}

// TestDisembarkPressReleaseUncitedExit2: a delegated document whose evidence
// resolves to nothing is the whole-document refusal — exit 2, and the previously
// derived press-release.json is left byte-for-byte untouched.
func TestDisembarkPressReleaseUncitedExit2(t *testing.T) {
	dir := buildSynthLifeboat(t, synthOpts{briefPR: true})
	// Establish the deterministic derived file first.
	runCLI(t, "disembark", "press-release", dir)
	before, err := os.ReadFile(filepath.Join(dir, "press-release.json"))
	if err != nil {
		t.Fatalf("derived press-release.json missing: %v", err)
	}
	payload := synthPayloadFile(t, `{"schema_version":1,"mode":"delegated","prompt_version":"0.1.0",`+
		`"headline":"H","body":"B","evidence":["no/such/path.md"]}`)
	var stdout, stderr bytes.Buffer
	code := Run([]string{"disembark", "press-release", dir, "--press-release-json", payload}, &stdout, &stderr)
	if code != 2 {
		t.Fatalf("exit = %d, want 2\nstderr:%s", code, stderr.String())
	}
	after, err := os.ReadFile(filepath.Join(dir, "press-release.json"))
	if err != nil {
		t.Fatalf("derived press-release.json vanished after refusal: %v", err)
	}
	if !bytes.Equal(before, after) {
		t.Errorf("refused delegated run mutated the derived file:\nbefore:%s\nafter:%s", before, after)
	}
}

// ---------------------------------------------------------------------------
// oracle
// ---------------------------------------------------------------------------

// TestDisembarkOracleRequiresSourceArg: the source-repo arg is required — a
// single-arg invocation is a usage error, exit 2.
func TestDisembarkOracleRequiresSourceArg(t *testing.T) {
	dir := buildSynthLifeboat(t, synthOpts{coverage: &lifeboat.Summary{Grounded: 7, Blank: 3}})
	var stdout, stderr bytes.Buffer
	code := Run([]string{"disembark", "oracle", dir}, &stdout, &stderr)
	if code != 2 {
		t.Fatalf("exit = %d, want 2 for a missing <source-repo>\nstderr:%s", code, stderr.String())
	}
}

// TestDisembarkOracleDeterministicShip: a healthy sealed lifeboat (verified
// manifest, blank<=grounded coverage) yields SHIP in the rendered text, exit 0.
func TestDisembarkOracleDeterministicShip(t *testing.T) {
	dir := buildSynthLifeboat(t, synthOpts{sourceName: "src", coverage: &lifeboat.Summary{Grounded: 7, Partial: 4, Blank: 3}})
	src := realSrcDir(t, "src")
	out := string(runCLI(t, "disembark", "oracle", dir, src))
	if !strings.Contains(out, "SHIP") {
		t.Errorf("rendered oracle audit missing SHIP:\n%s", out)
	}
	if _, err := os.Stat(filepath.Join(dir, "audit")); err != nil {
		t.Errorf("audit/ directory not written: %v", err)
	}
}

// TestDisembarkOracleJSON: --json emits a parseable OracleResult (one verb suffices).
func TestDisembarkOracleJSON(t *testing.T) {
	dir := buildSynthLifeboat(t, synthOpts{sourceName: "src", coverage: &lifeboat.Summary{Grounded: 7, Blank: 3}})
	src := realSrcDir(t, "src")
	out := runCLI(t, "disembark", "oracle", dir, src, "--json")
	var res lifeboat.OracleResult
	if err := json.Unmarshal(out, &res); err != nil {
		t.Fatalf("oracle --json is not a result: %v\n%s", err, out)
	}
	if res.Verdict != lifeboat.VerdictShip {
		t.Errorf("verdict = %q, want SHIP", res.Verdict)
	}
	if res.AuditPath == "" {
		t.Errorf("result missing audit_path: %+v", res)
	}
}
