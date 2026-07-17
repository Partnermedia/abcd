package ahoy

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/REPPL/abcd-cli/internal/core/identity"
)

// stubPrompter answers Prompt from a fixed map (falling back to def) and Confirm
// with a fixed boolean. It is the seam for exercising interactive install paths.
type stubPrompter struct {
	answers map[string]string
	confirm bool
}

func (s stubPrompter) Confirm(string) bool { return s.confirm }

func (s stubPrompter) Prompt(key string, _ []string, def string) string {
	if v, ok := s.answers[key]; ok {
		return v
	}
	return def
}

// TestStepConfigValuesRejectsInvalidDocsTarget guards that a typo'd interactive
// docs_target answer is never persisted (which would plant markers in both
// files and re-emit the gap forever). stepConfigValues must return nil, exactly
// as it already does for an invalid visibility.
func TestStepConfigValuesRejectsInvalidDocsTarget(t *testing.T) {
	dir := t.TempDir()
	a := &applyCtx{
		cwd:      dir,
		approved: map[GapCategory]bool{ConfigChange: true},
		gapPresent: map[string]bool{
			"config.visibility_missing":     true,
			"config.docs_target_missing":    true,
			"config.oracle_backend_missing": true,
		},
		prompter: stubPrompter{answers: map[string]string{
			"visibility":  "private",
			"docs_target": "clade_md", // typo — not in docsTargetChoices
		}},
	}
	if cfg := a.stepConfigValues(); cfg != nil {
		t.Fatalf("stepConfigValues persisted an invalid docs_target: %+v", cfg)
	}
	if _, err := os.Stat(configPath(dir)); err == nil {
		t.Errorf("invalid docs_target was written to config.json")
	}
}

// TestStepConfigValuesRefusesToClobberMalformedConfig proves a malformed
// config.json is left untouched: rebuilding it from the four install values would
// destroy whatever the user had. stepConfigValues must return nil (partial) and
// the file bytes must be unchanged.
func TestStepConfigValuesRefusesToClobberMalformedConfig(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Dir(configPath(dir)), 0o755); err != nil {
		t.Fatal(err)
	}
	const malformed = "{ this is not valid json, but the user's file }\n"
	if err := os.WriteFile(configPath(dir), []byte(malformed), 0o644); err != nil {
		t.Fatal(err)
	}
	a := &applyCtx{
		cwd:      dir,
		approved: map[GapCategory]bool{ConfigChange: true},
		gapPresent: map[string]bool{
			"config.visibility_missing":     true,
			"config.docs_target_missing":    true,
			"config.oracle_backend_missing": true,
		},
		prompter: stubPrompter{answers: map[string]string{
			"visibility":     "private",
			"docs_target":    docsTargetDefault,
			"oracle_backend": oracleBackendDefault,
		}},
	}
	if cfg := a.stepConfigValues(); cfg != nil {
		t.Fatalf("stepConfigValues acted on a malformed config: %+v", cfg)
	}
	got, err := os.ReadFile(configPath(dir))
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != malformed {
		t.Errorf("malformed config.json was overwritten:\n got: %q\nwant: %q", got, malformed)
	}
}

// TestInstallAdoptsIdentityPinInteractively guards that the advisory identity
// pin can be adopted through a later interactive `ahoy install`, as the gap's
// fix hint advertises — the "already_up_to_date" early return must not short
// -circuit it (iss-62).
func TestInstallAdoptsIdentityPinInteractively(t *testing.T) {
	setupHermetic(t)
	t.Setenv("GIT_CONFIG_GLOBAL", os.DevNull)
	t.Setenv("GIT_CONFIG_SYSTEM", os.DevNull)
	repo := t.TempDir()
	idMustGit(t, repo, "init")
	idMustGit(t, repo, "config", "user.name", "Alex Reppel")
	idMustGit(t, repo, "config", "user.email", "alex@example.com")

	// First install (--yes) resolves the actionable gaps but deliberately leaves
	// the advisory identity pin un-adopted.
	if _, err := Install(repo, installOpts(), RefusingPrompter{}); err != nil {
		t.Fatal(err)
	}
	if _, ok, _ := identity.LoadPin(repo); ok {
		t.Fatal("precondition: --yes must not pin the current identity")
	}
	det, err := Detect(repo)
	if err != nil {
		t.Fatal(err)
	}
	if !hasGap(det.Gaps, "git_identity.unpinned") {
		t.Fatalf("expected advisory git_identity.unpinned gap after --yes install: %+v", det.Gaps)
	}
	if len(actionable(det.Gaps)) != 0 {
		t.Fatalf("precondition: repo should be otherwise clean, got %+v", actionable(det.Gaps))
	}

	// Interactive re-install with a confirming prompter must fall through the
	// early return and adopt the pin rather than reporting already_up_to_date.
	if _, err := Install(repo, InstallOptions{}, stubPrompter{confirm: true}); err != nil {
		t.Fatal(err)
	}
	got, ok, err := identity.LoadPin(repo)
	if err != nil || !ok {
		t.Fatalf("interactive install did not adopt the identity pin: ok=%v err=%v", ok, err)
	}
	if got.Name != "Alex Reppel" || got.Email != "alex@example.com" {
		t.Fatalf("wrong pin written: %+v", got)
	}
}
