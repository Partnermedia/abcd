package launch

import (
	"errors"
	"strings"
	"testing"
)

// TestSvgPayloadSecretRefuses is the GHSA-5mmm-3whv-3rqp SVG axis: a secret in
// an included, extension-"skipped" text asset (docs/assets/img/x.svg) must be
// scanned, hard-fail, drive DryRun.WouldRefuseOn non-empty, and block Ship. The
// old skip-by-extension shipped the raw bytes unscanned.
func TestSvgPayloadSecretRefuses(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, ".abcd/config/launch-payload.json", `{"includes": ["docs"]}`)
	// FAKE token shape only: ghp_ + 36 chars, matching \bghp_[A-Za-z0-9]{36,}.
	token := "ghp_" + strings.Repeat("a", 36)
	writeFile(t, root, "docs/assets/img/x.svg",
		`<svg xmlns="http://www.w3.org/2000/svg"><!-- `+token+` --></svg>`+"\n")

	report, err := DryRun(DryRunRequest{RepoRoot: root, Version: "1.0.0"})
	if err != nil {
		t.Fatalf("dry-run must return nil error on a finding, got %v", err)
	}
	if report.Scan.HardFails == 0 {
		t.Fatalf("secret in the SVG asset was not caught by the scan: %+v", report.Scan)
	}
	if len(report.WouldRefuseOn) == 0 {
		t.Errorf("WouldRefuseOn must be non-empty for a secret in a payload SVG")
	}
	if report.WouldPublish {
		t.Errorf("WouldPublish must be false with a payload secret")
	}

	ship, err := Ship(ShipRequest{RepoRoot: root, Version: "1.0.0"})
	if !errors.Is(err, ErrShipBlocked) {
		t.Fatalf("ship must return ErrShipBlocked for a secret in a payload SVG, got %v", err)
	}
	if !ship.Blocked || ship.WouldPublish {
		t.Errorf("ship must be blocked and not would-publish: %+v", ship)
	}
}

// TestUnscannedPayloadRefuses is the GHSA-5mmm-3whv-3rqp binary-smuggle axis: a
// NUL-prefixed file whose extension is NOT on the reviewed skip list is
// classified unscannable and surfaced in Unscanned; the launch gate must fail
// closed on it rather than let "some other file scanned" count as coverage.
func TestUnscannedPayloadRefuses(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, ".abcd/config/launch-payload.json", `{"includes": ["commands"]}`)
	// A leading NUL makes an otherwise-.md file read as binary → Unscanned.
	writeFile(t, root, "commands/clean.md", "wholly clean documentation\n")
	writeFile(t, root, "commands/smuggle.md", "\x00ghp_"+strings.Repeat("a", 40))

	report, err := DryRun(DryRunRequest{RepoRoot: root, Version: "1.0.0"})
	if err != nil {
		t.Fatalf("dry-run must return nil error, got %v", err)
	}
	var sawUnscanned bool
	for _, p := range report.Scan.Unscanned {
		if p == "commands/smuggle.md" {
			sawUnscanned = true
		}
	}
	if !sawUnscanned {
		t.Fatalf("unscannable payload file must be surfaced in Unscanned: %+v", report.Scan)
	}
	var sawRefusal bool
	for _, r := range report.WouldRefuseOn {
		if strings.Contains(r, "unscanned") {
			sawRefusal = true
		}
	}
	if !sawRefusal {
		t.Errorf("WouldRefuseOn must fail closed on an unscanned payload file, got %v", report.WouldRefuseOn)
	}

	ship, err := Ship(ShipRequest{RepoRoot: root, Version: "1.0.0"})
	if !errors.Is(err, ErrShipBlocked) {
		t.Fatalf("ship must be blocked on an unscanned payload file, got %v", err)
	}
	if ship.WouldPublish {
		t.Errorf("ship must not would-publish with an unscanned payload file: %+v", ship)
	}
}
