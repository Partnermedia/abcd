package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/REPPL/abcd-cli/internal/core/lifeboat"
)

// buildGraveyardLifeboat hand-builds a packed lifeboat directory carrying a
// parseable _provenance.json and the two layer-1/2 files. Plan does not yet emit
// the graveyard, so the surface test builds the lifeboat by hand and is
// independent of the parallel archaeology/abandoned work.
func buildGraveyardLifeboat(t *testing.T) string {
	t.Helper()
	dir := filepath.Join(t.TempDir(), "lifeboat")
	gy := filepath.Join(dir, "graveyard")
	if err := os.MkdirAll(gy, 0o755); err != nil {
		t.Fatal(err)
	}
	arch := lifeboat.Archaeology{SchemaVersion: 1, Findings: []lifeboat.Finding{
		{ID: "rev-9f3a1c2d4e5b", Signal: lifeboat.SignalRevert, Summary: "reverted", Evidence: []string{"Revert"}},
	}}
	aband := lifeboat.Abandoned{SchemaVersion: 1, Findings: []lifeboat.Finding{
		{ID: "adr-12", Signal: lifeboat.SignalSupersededADR, Summary: "superseded", Evidence: []string{"superseded_by: adr-31"}},
	}}
	writeJSON(t, filepath.Join(gy, "archaeology.json"), arch)
	writeJSON(t, filepath.Join(gy, "abandoned.json"), aband)
	prov := fmt.Sprintf(`{"schema_version":1,"generator":"test","source_name":"fix","manifest_sha256":%q}`,
		strings.Repeat("a", 64))
	if err := os.WriteFile(filepath.Join(dir, lifeboat.ProvenanceName), []byte(prov), 0o644); err != nil {
		t.Fatal(err)
	}
	return dir
}

func writeJSON(t *testing.T, p string, v any) {
	t.Helper()
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, append(b, '\n'), 0o644); err != nil {
		t.Fatal(err)
	}
}

func lessonsPayloadFile(t *testing.T, dir string, lessons ...lifeboat.Lesson) string {
	t.Helper()
	lf := lifeboat.LessonsFile{SchemaVersion: 1, Lessons: lessons}
	b, err := json.Marshal(lf)
	if err != nil {
		t.Fatal(err)
	}
	p := filepath.Join(dir, "lessons.input.json")
	if err := os.WriteFile(p, b, 0o644); err != nil {
		t.Fatal(err)
	}
	return p
}

// TestDisembarkGraveyardWritesLessons is the end-to-end wiring: the verb reads a
// lesson file, validates it against the packed lifeboat, and writes lessons.json;
// --json emits the LessonsResult.
func TestDisembarkGraveyardWritesLessons(t *testing.T) {
	dir := buildGraveyardLifeboat(t)
	lessons := lessonsPayloadFile(t, t.TempDir(),
		lifeboat.Lesson{ID: "les-engine-v1", Lesson: "engine retired", Confidence: lifeboat.ConfidenceHigh,
			Evidence: []string{"rev-9f3a1c2d4e5b"}})
	out := runCLI(t, "disembark", "graveyard", dir, "--lessons-json", lessons, "--json")
	var res lifeboat.LessonsResult
	if err := json.Unmarshal(out, &res); err != nil {
		t.Fatalf("graveyard --json is not a result: %v\n%s", err, out)
	}
	if res.Written != 1 {
		t.Errorf("Written = %d, want 1", res.Written)
	}
	if _, err := os.Stat(filepath.Join(dir, "graveyard", "lessons.json")); err != nil {
		t.Errorf("lessons.json not written: %v", err)
	}
}

// TestDisembarkGraveyardStdin proves the "-" stdin transport works.
func TestDisembarkGraveyardStdin(t *testing.T) {
	dir := buildGraveyardLifeboat(t)
	payload := `{"schema_version":1,"lessons":[{"id":"les-x","lesson":"y","confidence":"high","evidence":["adr-12"]}]}`
	out := runCLIStdin(t, payload, "disembark", "graveyard", dir, "--lessons-json", "-", "--json")
	var res lifeboat.LessonsResult
	if err := json.Unmarshal(out, &res); err != nil {
		t.Fatalf("not a result: %v\n%s", err, out)
	}
	if res.Written != 1 {
		t.Errorf("Written = %d, want 1", res.Written)
	}
}

// TestDisembarkGraveyardAllDroppedExitZero holds the exit-code contract: a batch
// where every entry is dropped still exits 0 (a drop is honest, not a failure).
func TestDisembarkGraveyardAllDroppedExitZero(t *testing.T) {
	dir := buildGraveyardLifeboat(t)
	lessons := lessonsPayloadFile(t, t.TempDir(),
		lifeboat.Lesson{ID: "les-guess", Lesson: "no cite", Confidence: lifeboat.ConfidenceHigh,
			Evidence: []string{"dead-ref"}})
	var stdout, stderr bytes.Buffer
	code := Run([]string{"disembark", "graveyard", dir, "--lessons-json", lessons, "--json"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("exit = %d, want 0\nstdout:%s\nstderr:%s", code, stdout.String(), stderr.String())
	}
	var res lifeboat.LessonsResult
	if err := json.Unmarshal(stdout.Bytes(), &res); err != nil {
		t.Fatalf("not a result: %v", err)
	}
	if res.Dropped != 1 || res.Written != 0 {
		t.Errorf("res = %+v; want 1 dropped, 0 written", res)
	}
}

// TestDisembarkGraveyardBadDir holds the error contract: a directory that is not
// a lifeboat exits 2 with a scrubbed message (no absolute path leak).
func TestDisembarkGraveyardBadDir(t *testing.T) {
	dir := t.TempDir() // no _provenance.json
	lessons := lessonsPayloadFile(t, dir,
		lifeboat.Lesson{ID: "les-x", Lesson: "y", Confidence: lifeboat.ConfidenceHigh, Evidence: []string{"adr-12"}})
	out, err := runCLIErr(t, "disembark", "graveyard", dir, "--lessons-json", lessons)
	if err == nil {
		t.Fatal("a non-lifeboat dir must fail")
	}
	if strings.Contains(string(out), filepath.Dir(dir)) {
		t.Errorf("error leaked an absolute path: %q", out)
	}
}

// TestDisembarkGraveyardRefusesSymlinkPayload holds the read trust guard: a
// symlinked lessons file is refused.
func TestDisembarkGraveyardRefusesSymlinkPayload(t *testing.T) {
	dir := buildGraveyardLifeboat(t)
	real := lessonsPayloadFile(t, t.TempDir(),
		lifeboat.Lesson{ID: "les-x", Lesson: "y", Confidence: lifeboat.ConfidenceHigh, Evidence: []string{"adr-12"}})
	link := filepath.Join(t.TempDir(), "link.json")
	if err := os.Symlink(real, link); err != nil {
		t.Skipf("symlink unsupported: %v", err)
	}
	if _, err := runCLIErr(t, "disembark", "graveyard", dir, "--lessons-json", link); err == nil {
		t.Fatal("a symlinked payload must be refused")
	}
}

// TestDisembarkGraveyardRequiresFlag holds the required-flag contract.
func TestDisembarkGraveyardRequiresFlag(t *testing.T) {
	dir := buildGraveyardLifeboat(t)
	if _, err := runCLIErr(t, "disembark", "graveyard", dir); err == nil {
		t.Fatal("missing --lessons-json must fail")
	}
}
