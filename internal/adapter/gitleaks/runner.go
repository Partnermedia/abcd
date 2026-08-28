package gitleaks

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/intentdriven/abcd/internal/fsutil"
)

// execRunner is the production Runner: it writes the text to a private temp
// directory and runs `gitleaks dir` over it with the default ruleset, reading
// back the JSON report. This mirrors the invocation iss-96's measurement used
// (gitleaks 8.24.3, `gitleaks dir`, default rules). It is the ONE place that
// spawns a process; every other path through this package is pure.
type execRunner struct{}

// Run scans text and returns gitleaks' raw JSON report bytes. gitleaks exits
// non-zero when it finds a leak, so --exit-code 0 makes a found leak a success
// and reserves a non-zero exit for a genuine tool failure. The report is read
// from a file rather than stdout so banner/log noise cannot corrupt the JSON.
func (execRunner) Run(ctx context.Context, binPath, text string) ([]byte, error) {
	dir, err := os.MkdirTemp("", "abcd-gitleaks-")
	if err != nil {
		return nil, fmt.Errorf("temp dir: %w", err)
	}
	defer os.RemoveAll(dir)

	src := filepath.Join(dir, "transcript.txt")
	if err := os.WriteFile(src, []byte(text), 0o600); err != nil {
		return nil, fmt.Errorf("write scan input: %w", err)
	}
	report := filepath.Join(dir, "report.json")

	ctx, cancel := context.WithTimeout(ctx, runTimeout)
	defer cancel()

	// #nosec G204 -- binPath is resolved by the adapter (a configured regular
	// file or an exec.LookPath result), not attacker-controlled; the scan input
	// is a file path we just created, never the transcript content.
	cmd := exec.CommandContext(ctx, binPath,
		"dir", dir,
		"--report-format", "json",
		"--report-path", report,
		"--no-banner",
		"--exit-code", "0",
	)
	if out, err := cmd.CombinedOutput(); err != nil {
		return nil, fmt.Errorf("gitleaks run: %w (%s)", err, string(out))
	}

	data, err := fsutil.ReadGuarded(report, maxReportBytes)
	if err != nil {
		if os.IsNotExist(err) {
			// gitleaks writes no report file when it finds nothing; treat as empty.
			return nil, nil
		}
		return nil, fmt.Errorf("read report: %w", err)
	}
	return data, nil
}

// maxReportBytes caps the report we read back (trust boundary on a file gitleaks
// wrote into a temp dir we own — a bound, not an expected size).
const maxReportBytes = 8 * 1024 * 1024
