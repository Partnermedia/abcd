package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestDocsLintRenderSanitisesConfigFields pins that the findings renderer
// sanitises every config-derived field. A banned token's id is free text from
// the committed config — a trust boundary (LoadConfig's own contract) — so a
// hostile clone must not be able to put a raw terminal escape on the finding
// line through it.
func TestDocsLintRenderSanitisesConfigFields(t *testing.T) {
	repo := t.TempDir()
	t.Chdir(repo)
	if err := os.MkdirAll(filepath.Join(repo, ".abcd"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(repo, "docs"), 0o755); err != nil {
		t.Fatal(err)
	}
	cfg := `{
	  "roots": ["docs"],
	  "banned_tokens": [
	    {"id": "evil\u001b[2Jtoken", "pattern": "forbidden-word", "message": "no", "severity": "warn", "successor": "allowed-word", "allow_context": ["nowhere-real"]}
	  ]
	}`
	if err := os.WriteFile(filepath.Join(repo, ".abcd", "docs-lint.json"), []byte(cfg), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(repo, "docs", "page.md"), []byte("uses the forbidden-word here\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	var stdout, stderr bytes.Buffer
	Run([]string{"docs", "lint"}, &stdout, &stderr)
	out := stdout.String() + stderr.String()
	if !strings.Contains(out, "evil") {
		t.Fatalf("expected the finding to render (config id present), got:\n%s", out)
	}
	if strings.ContainsRune(out, 0x1b) {
		t.Fatalf("a raw ESC from the config's token id reached the terminal:\n%q", out)
	}
}
