package cli

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestJSONErrorEnvelopeNoAbsolutePathLeak is the iss-76 detector: cli.Run routes
// every command error through the --json envelope (and the stderr text line), so
// any verb whose error chain carries an *os.PathError / *os.LinkError would leak
// its absolute local path into machine output. The sanitisation belongs at the
// Run() boundary — iss-29 fixed only the docs-lint config-load branch. This is a
// per-verb table: each row drives a real verb into a filesystem error and asserts
// the envelope carries no absolute path but keeps the file's base name for context.
func TestJSONErrorEnvelopeNoAbsolutePathLeak(t *testing.T) {
	repo := t.TempDir()
	t.Chdir(repo)

	// An absolute path guaranteed to fail os.ReadFile with a *PathError.
	absMissing := filepath.Join(repo, "no-such-dir", "page.json")

	cases := []struct {
		name string
		args []string
		base string // basename the sanitised message should retain
	}{
		{
			name: "memory ask --page-json missing file",
			args: []string{"memory", "ask", "q", "--file-back", "--page-json", absMissing, "--json"},
			base: "page.json",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			code := Run(tc.args, &stdout, &stderr)
			if code == 0 {
				t.Fatalf("expected a non-zero exit; stdout=%q stderr=%q", stdout.String(), stderr.String())
			}
			var env struct {
				Error string `json:"error"`
			}
			if err := json.Unmarshal(stderr.Bytes(), &env); err != nil {
				t.Fatalf("--json error not JSON-shaped: %v\nstderr: %q", err, stderr.String())
			}
			if strings.Contains(env.Error, repo) {
				t.Fatalf("envelope leaked the absolute path %q:\n%s", repo, env.Error)
			}
			if !strings.Contains(env.Error, tc.base) {
				t.Fatalf("envelope dropped all file context (want basename %q):\n%s", tc.base, env.Error)
			}
		})
	}
}

// TestCaptureSymlinkErrorNoPathLeak reproduces the security-review finding that
// the fix must also cover: capture's allocator embeds an ABSOLUTE ledger path via
// fmt.Errorf("%w … %s") — not an os.PathError — so a typed walk misses it. A
// symlinked issues/open dir trips the guard and its error reaches the --json
// envelope; it must not leak the developer's absolute repo path.
func TestCaptureSymlinkErrorNoPathLeak(t *testing.T) {
	repo := t.TempDir()
	t.Chdir(repo)
	issues := filepath.Join(repo, ".abcd", "work", "issues")
	if err := os.MkdirAll(issues, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(t.TempDir(), filepath.Join(issues, "open")); err != nil {
		t.Fatal(err)
	}

	var stdout, stderr bytes.Buffer
	code := Run([]string{"capture", "a defect", "--json"}, &stdout, &stderr)
	if code == 0 {
		t.Fatalf("expected a non-zero exit for a symlinked issues/open; stdout=%q stderr=%q", stdout.String(), stderr.String())
	}
	var env struct {
		Error string `json:"error"`
	}
	if err := json.Unmarshal(stderr.Bytes(), &env); err != nil {
		t.Fatalf("--json error not JSON-shaped: %v\nstderr: %q", err, stderr.String())
	}
	if strings.Contains(env.Error, repo) {
		t.Fatalf("capture envelope leaked the absolute repo path %q:\n%s", repo, env.Error)
	}
}

// homeRootedErr is a custom error type (like history.StorePathError) that renders
// a home-dir-rooted absolute path — the class a typed PathError walk cannot see.
type homeRootedErr struct{ path string }

func (e homeRootedErr) Error() string { return "history: store unreadable (" + e.path + ")" }

// TestScrubPaths is the unit-level detector for the Run()-boundary sanitiser. It
// removes local-identity paths reaching machine output three ways — os.PathError/
// os.LinkError (→ base name), and cwd- or home-rooted paths embedded by fmt or a
// custom error type (→ "." / "~") — while leaving relative and non-path errors
// untouched.
func TestScrubPaths(t *testing.T) {
	abs := filepath.Join(os.TempDir(), "secret", "config.json")
	rel := filepath.Join("rel", "config.json")
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatal(err)
	}

	cases := []struct {
		name        string
		err         error
		wantAbsent  string // must NOT appear in the scrubbed message
		wantPresent []string
	}{
		{
			name:        "fmt-embedded cwd path (capture class)",
			err:         fmt.Errorf("path unsafe: allocator lock path is a symlink: %s", filepath.Join(cwd, ".abcd", "work", "issues", ".iss-alloc.lock")),
			wantAbsent:  cwd,
			wantPresent: []string{".iss-alloc.lock"},
		},
		{
			name:        "custom-type home path (history class)",
			err:         homeRootedErr{path: filepath.Join(home, ".abcd", "history", "x")},
			wantAbsent:  home,
			wantPresent: []string{".abcd", "history: store unreadable"},
		},
		{
			name:        "bare PathError",
			err:         &os.PathError{Op: "open", Path: abs, Err: fs.ErrNotExist},
			wantAbsent:  abs,
			wantPresent: []string{"open", "config.json", fs.ErrNotExist.Error()},
		},
		{
			name:        "wrapped PathError keeps context",
			err:         fmt.Errorf("cannot read --page-json: %w", &os.PathError{Op: "open", Path: abs, Err: fs.ErrPermission}),
			wantAbsent:  abs,
			wantPresent: []string{"cannot read --page-json", "config.json"},
		},
		{
			name:        "LinkError both paths",
			err:         &os.LinkError{Op: "rename", Old: abs, New: abs + ".tmp", Err: fs.ErrExist},
			wantAbsent:  abs,
			wantPresent: []string{"rename", "config.json"},
		},
		{
			name:        "joined errors both scrubbed",
			err:         errors.Join(&os.PathError{Op: "open", Path: abs, Err: fs.ErrNotExist}, fmt.Errorf("and: %w", &os.PathError{Op: "stat", Path: abs, Err: fs.ErrPermission})),
			wantAbsent:  abs,
			wantPresent: []string{"config.json"},
		},
		{
			name:        "relative PathError left intact",
			err:         &os.PathError{Op: "open", Path: rel, Err: fs.ErrNotExist},
			wantAbsent:  "\x00-never-", // sentinel: nothing to strip
			wantPresent: []string{rel},
		},
		{
			name:        "non-path error untouched",
			err:         errors.New("plain failure"),
			wantAbsent:  "\x00-never-",
			wantPresent: []string{"plain failure"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := scrubPaths(tc.err)
			if strings.Contains(got, tc.wantAbsent) {
				t.Fatalf("scrubPaths kept the absolute path %q: %s", tc.wantAbsent, got)
			}
			for _, want := range tc.wantPresent {
				if !strings.Contains(got, want) {
					t.Fatalf("scrubPaths dropped %q: %s", want, got)
				}
			}
		})
	}
}
