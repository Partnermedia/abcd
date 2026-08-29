// Package gitleaks is an OPT-IN external-scanner adapter for the transcript
// redaction path (iss-96). It is off by default and adds nothing — no
// dependency invoked, no process spawned, no cost — until a repo asks for it by
// dropping an enabled .abcd/config/gitleaks.json. This realises abcd's
// host-delegated boundary (AGENTS.md: "native/CLI/API/MCP oracles are opt-in
// adapters"): the always-on native scanner (internal/adapter/scanner) stays the
// default, and a repo that wants deeper coverage over the unstructured prose of
// a captured transcript arms this adapter, which shells out to a gitleaks binary
// and folds its findings into the same redaction the native scanner already
// runs.
//
// Reach and cost. The native pattern set is prefix-anchored and misses
// unanchored, labelled, high-entropy values in prose (iss-96's residue).
// gitleaks' generic-api-key rule reaches keyword+delimiter+entropy values that
// the native set does not. Its findings AUGMENT native redaction — they are
// masked by the same scanner.Redact discipline and counted in the record's
// audit counters — they never replace it.
//
// Fail-closed, never a silent no-op. When a repo has opted in but the gitleaks
// binary is not found (not on PATH and no valid configured path), Scan returns
// ErrConfiguredNotFound rather than quietly skipping the deeper scan: a repo
// that armed the adapter must not believe it is covered when it is not. The
// history store surfaces that error and refuses the write, exactly as it fails
// closed on a degraded native scanner.
//
// Testability. The binary lookup and the process execution are both injected
// (Adapter.LookPath, Adapter.Runner), so the gate, the loud-stage and the
// augmentation path are all exercised with a fake — no real gitleaks binary is
// required to test this package.
package gitleaks

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/intentdriven/abcd/internal/adapter/scanner"
	"github.com/intentdriven/abcd/internal/fsutil"
)

// configRelPath is the per-repo opt-in config, a sibling of the native
// scanner's .abcd/config/pii.json.
var configRelPath = filepath.Join(".abcd", "config", "gitleaks.json")

// maxConfigBytes caps the per-repo opt-in config (trust boundary), matching the
// native scanner's own override cap.
const maxConfigBytes = 256 * 1024

// runTimeout bounds a single gitleaks invocation so an opted-in repo cannot be
// wedged by a hung binary (the SessionEnd hook holds the history repo lock
// across the capture — the same concern that guards the native config read).
const runTimeout = 30 * time.Second

// ErrConfiguredNotFound is returned when the repo opted in but no gitleaks
// binary could be located. It is deliberately loud: the message names the
// opt-in so an operator sees "gitleaks configured but not found" rather than a
// silent skip.
var ErrConfiguredNotFound = errors.New("gitleaks configured but not found")

// Config is the on-disk opt-in shape (.abcd/config/gitleaks.json). Absent file
// means not opted in (the default).
type Config struct {
	SchemaVersion int    `json:"schema_version"`
	Enabled       bool   `json:"enabled"`
	Path          string `json:"path,omitempty"` // optional explicit binary path
}

// Runner executes gitleaks over text and returns its raw JSON report bytes. It
// is the single seam that touches a process, injected so tests need no binary.
type Runner interface {
	Run(ctx context.Context, binPath, text string) ([]byte, error)
}

// Adapter augments native redaction with gitleaks findings. Construct it with
// NewDefault for the production wiring (real PATH lookup, real exec runner), or
// build one directly with a fake LookPath/Runner in tests.
type Adapter struct {
	// LookPath resolves a bare binary name to a full path (defaults to
	// exec.LookPath). It returns an error when the binary is absent.
	LookPath func(string) (string, error)
	// Runner executes the located binary over the text.
	Runner Runner
}

// NewDefault returns the production adapter: exec.LookPath for discovery and a
// real gitleaks shell-out for execution.
func NewDefault() *Adapter {
	return &Adapter{LookPath: exec.LookPath, Runner: execRunner{}}
}

// LoadConfig reads the per-repo opt-in config with the same trust-boundary
// discipline the native scanner applies to pii.json: an os.Root-contained,
// size-capped, symlink-refusing read. An ABSENT config is the default-off case
// and returns a disabled Config with no error — the whole point of opt-in is
// that a repo which did nothing pays nothing. A PRESENT but unreadable or
// invalid config fails closed with an error, so a repo that tried to arm the
// adapter and got it wrong is told, not silently left unprotected.
func LoadConfig(repoRoot string) (Config, error) {
	root, err := os.OpenRoot(repoRoot)
	if err != nil {
		return Config{}, fmt.Errorf("gitleaks: repo root cannot be opened for contained reads: %w", err)
	}
	defer root.Close()
	data, err := fsutil.ReadGuardedInRoot(root, configRelPath, maxConfigBytes)
	if err != nil {
		if os.IsNotExist(err) {
			return Config{}, nil // not opted in — the default
		}
		return Config{}, fmt.Errorf("gitleaks: per-repo config unreadable (%s): %w", configRelPath, err)
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return Config{}, fmt.Errorf("gitleaks: per-repo config is not valid JSON (%s): %w", configRelPath, err)
	}
	return cfg, nil
}

// Scan is the transcript-path entry point the history store calls. It loads the
// per-repo config and delegates to the default adapter. When the repo has NOT
// opted in it returns (nil, nil) having invoked nothing.
func Scan(repoRoot, text, logical string) ([]scanner.Finding, error) {
	cfg, err := LoadConfig(repoRoot)
	if err != nil {
		return nil, err
	}
	return NewDefault().Augment(context.Background(), cfg, text, logical)
}

// Augment scans text with gitleaks when cfg opts in, returning findings to fold
// into the native redaction. Behaviour by state:
//
//   - not opted in (cfg.Enabled false): returns (nil, nil), invokes NOTHING —
//     no lookup, no process. This is the gate that keeps the default path free.
//   - opted in, binary absent: returns ErrConfiguredNotFound (loud-stage).
//   - opted in, binary present: runs it and converts its findings.
func (a *Adapter) Augment(ctx context.Context, cfg Config, text, logical string) ([]scanner.Finding, error) {
	if !cfg.Enabled {
		return nil, nil
	}
	bin, err := a.resolveBinary(cfg)
	if err != nil {
		return nil, err
	}
	raw, err := a.Runner.Run(ctx, bin, text)
	if err != nil {
		return nil, fmt.Errorf("gitleaks: scan failed: %w", err)
	}
	reports, err := parseReport(raw)
	if err != nil {
		return nil, fmt.Errorf("gitleaks: cannot parse report: %w", err)
	}
	return toFindings(text, logical, reports), nil
}

// resolveBinary finds the gitleaks binary. A configured path wins when it names
// an existing regular file; otherwise the bare name is looked up on PATH. An
// unresolvable binary on an opted-in repo is ErrConfiguredNotFound — never a
// silent skip.
func (a *Adapter) resolveBinary(cfg Config) (string, error) {
	if p := strings.TrimSpace(cfg.Path); p != "" {
		if fi, err := os.Stat(p); err == nil && fi.Mode().IsRegular() {
			return p, nil
		}
		return "", fmt.Errorf("%w: configured path %q is not an executable file", ErrConfiguredNotFound, p)
	}
	bin, err := a.LookPath("gitleaks")
	if err != nil {
		return "", fmt.Errorf("%w: not on PATH and no path configured", ErrConfiguredNotFound)
	}
	return bin, nil
}

// report is the subset of a gitleaks JSON finding this adapter consumes. Secret
// is the raw credential value gitleaks isolated; Match is the wider surrounding
// text. RuleID names the rule (e.g. generic-api-key).
type report struct {
	RuleID string `json:"RuleID"`
	Secret string `json:"Secret"`
	Match  string `json:"Match"`
}

// parseReport decodes gitleaks' JSON report. An empty report (no findings) is
// valid and yields no reports.
func parseReport(raw []byte) ([]report, error) {
	trimmed := strings.TrimSpace(string(raw))
	if trimmed == "" || trimmed == "null" {
		return nil, nil
	}
	var reports []report
	if err := json.Unmarshal([]byte(trimmed), &reports); err != nil {
		return nil, err
	}
	return reports, nil
}

// toFindings converts gitleaks reports into scanner.Findings positioned against
// the scanned text, so scanner.Redact masks them exactly as it masks a native
// secret (byte-span fingerprint). It locates each secret by CONTENT — searching
// the text for the reported value rather than trusting gitleaks' line/column
// indexing — which keeps the byte column exact for sealLine and is robust
// across gitleaks versions. Every line carrying the value gets a finding, so a
// value repeated on several lines is redacted on each.
//
// A single-line credential (generic-api-key, the rule iss-96's measurement
// showed doing the work) is fully covered. A multi-line secret (a PEM block) is
// skipped here — it is already native-covered by token:pem_private_key — so this
// adapter never has to reason about spans that cross a line boundary.
func toFindings(text, logical string, reports []report) []scanner.Finding {
	lines := strings.Split(text, "\n")
	var out []scanner.Finding
	// One Finding per (line, column, value) span: two reports of the same
	// value each locate every occurrence, so the same span must not be
	// reported once per report.
	type span struct {
		line, col int
		value     string
	}
	seen := map[span]bool{}
	for _, r := range reports {
		needle := r.Secret
		if needle == "" {
			needle = r.Match
		}
		if needle == "" || strings.ContainsAny(needle, "\n\r") {
			continue
		}
		kind := "gitleaks:" + r.RuleID
		if r.RuleID == "" {
			kind = "gitleaks:generic"
		}
		for i, ln := range lines {
			// Every occurrence on the line, not the first: gitleaks reports a
			// secret echoed twice on one line (a request with its response, a
			// retry log) as two reports, and a search that never advances past
			// its first hit positions both on the first occurrence, leaving the
			// second one verbatim after redaction.
			for start := 0; start < len(ln); {
				idx := strings.Index(ln[start:], needle)
				if idx < 0 {
					break
				}
				col := start + idx
				start = col + len(needle)
				if seen[span{i, col, needle}] {
					continue
				}
				seen[span{i, col, needle}] = true
				out = append(out, scanner.Finding{
					File:     logical,
					Line:     i + 1,
					Column:   col + 1, // 1-based byte column, matching sealLine
					Kind:     kind,
					Severity: scanner.SeverityHardFail,
					Matched:  needle,
					Snippet:  ln,
				})
			}
		}
	}
	return out
}
