package positioning

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// containmentCanary stands in for whatever a hostile repo is actually aiming at
// — an SSH key, a credentials file, a host token. If it ever reaches a Report, a
// Block, or an error message, content from outside the repository escaped.
const containmentCanary = "SUPER-SECRET-CONTAINMENT-CANARY"

// hostileRepo builds the attack every lexical path check misses: a repository
// that COMMITS a symlinked directory (git mode 120000) and then names a file
// through it. Nothing in the configured path contains ".." or a leading "/", so
// fsutil.ValidRelPath accepts it and filepath.Join resolves it straight through
// the link to a file the repository does not own.
//
// Both real shapes are built: an absolute target (`etclink -> /etc`) and a
// relative one (`home -> ../../..`), which is the form that reaches a developer's
// dotfiles from an arbitrary clone location.
//
// It returns the repo root and the outside directory the link lands in.
func hostileRepo(t *testing.T, linkName string, absolute bool, outsideFiles map[string]string) (repo, outside string) {
	t.Helper()
	base := t.TempDir()
	outside = filepath.Join(base, "outside")
	if err := os.MkdirAll(outside, 0o755); err != nil {
		t.Fatal(err)
	}
	for rel, content := range outsideFiles {
		p := filepath.Join(outside, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	repo = filepath.Join(base, "repo")
	if err := os.MkdirAll(repo, 0o755); err != nil {
		t.Fatal(err)
	}
	target := "../outside" // a relative ancestor escape, resolved from repo/
	if absolute {
		target = outside
	}
	if err := os.Symlink(target, filepath.Join(repo, linkName)); err != nil {
		t.Fatal(err)
	}
	return repo, outside
}

// writeInto drops files into an existing directory (writeRepo makes its own).
func writeInto(t *testing.T, dir string, files map[string]string) {
	t.Helper()
	for rel, content := range files {
		p := filepath.Join(dir, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
}

const containmentBlock = `# Identity

## Identity (canonical)

- **Title:** Contained
- **Tagline:** A tagline no surface carries.
`

// TestCheckRefusesASurfaceReachedThroughASymlinkedAncestor is the leak this
// package's containment exists to stop: a cloned repo declares a surface behind
// a committed symlinked directory, and the check reads a file outside the repo
// and quotes it verbatim into the Report — which `abcd lint` prints and
// `abcd identity --json` emits whole.
func TestCheckRefusesASurfaceReachedThroughASymlinkedAncestor(t *testing.T) {
	for _, tc := range []struct {
		name     string
		link     string
		absolute bool
	}{
		{"absolute target", "etclink", true},
		{"relative ancestor target", "home", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			repo, _ := hostileRepo(t, tc.link, tc.absolute, map[string]string{
				"secret.txt": containmentCanary + "\n",
			})
			writeInto(t, repo, map[string]string{"IDENTITY.md": containmentBlock})

			cfg := Config{
				SchemaVersion: 1,
				Block:         BlockLocation{File: "IDENTITY.md", Heading: "Identity (canonical)"},
				Surfaces: []Surface{{
					ID:       "leak",
					Files:    []string{tc.link + "/secret.txt"},
					Kind:     KindRegexp,
					Patterns: []string{`(?s)^(.*)$`},
					Requires: []string{"tagline"},
				}},
			}
			// The lexical guard accepts the path: it has no ".." and no leading
			// "/", so containment cannot rest on it.
			if err := cfg.Validate(); err != nil {
				t.Fatalf("fixture is not the attack under test (config refused lexically): %v", err)
			}

			rep, err := Check(repo, cfg)
			if err == nil {
				t.Fatalf("Check read through a symlinked ancestor instead of refusing; report = %+v", rep)
			}
			if !errors.Is(err, ErrSurfaceUnreadable) {
				t.Fatalf("err = %v, want ErrSurfaceUnreadable", err)
			}
			assertNoCanary(t, "Check error", err.Error())
			assertNoCanary(t, "Check report", fmt.Sprintf("%+v", rep))

			// The proposal path re-reads the surface file; it must refuse too.
			prop, perr := Propose(repo, cfg)
			if perr == nil {
				t.Fatalf("Propose read through a symlinked ancestor; proposal = %+v", prop)
			}
			assertNoCanary(t, "Propose proposal", fmt.Sprintf("%+v", prop))
		})
	}
}

// TestParseBlockRefusesABlockReachedThroughASymlinkedAncestor closes the same
// hole on the canon: a registry that points the block at a path behind a
// committed symlink must not parse a file outside the repository as the repo's
// own identity.
func TestParseBlockRefusesABlockReachedThroughASymlinkedAncestor(t *testing.T) {
	repo, _ := hostileRepo(t, "home", false, map[string]string{
		"IDENTITY.md": "## Identity (canonical)\n\n- **Title:** " + containmentCanary +
			"\n- **Tagline:** " + containmentCanary + "\n",
	})
	loc := BlockLocation{File: "home/IDENTITY.md", Heading: "Identity (canonical)"}

	b, err := ParseBlock(repo, loc)
	if err == nil {
		t.Fatalf("ParseBlock read through a symlinked ancestor; block = %+v", b)
	}
	if !errors.Is(err, ErrBlockFileMissing) {
		t.Fatalf("err = %v, want ErrBlockFileMissing", err)
	}
	assertNoCanary(t, "ParseBlock error", err.Error())
	assertNoCanary(t, "ParseBlock block", fmt.Sprintf("%+v", b))
}

// TestLoadConfigRefusesARegistryReachedThroughASymlinkedAncestor: the registry
// itself sits at a fixed path, but a hostile repo can commit `.abcd` as a
// symlink and have the loader read a registry it does not own.
func TestLoadConfigRefusesARegistryReachedThroughASymlinkedAncestor(t *testing.T) {
	repo, _ := hostileRepo(t, ".abcd", false, map[string]string{
		"positioning.json": `{"schema_version":1,"block":{"file":"IDENTITY.md","heading":"` +
			containmentCanary + `"}}`,
	})

	cfg, ok, err := LoadConfig(repo)
	if err == nil {
		t.Fatalf("LoadConfig read a registry through a symlinked ancestor (ok=%v, cfg=%+v)", ok, cfg)
	}
	if ok {
		t.Error("LoadConfig reported the registry as adopted despite refusing it")
	}
	assertNoCanary(t, "LoadConfig error", err.Error())
	assertNoCanary(t, "LoadConfig config", fmt.Sprintf("%+v", cfg))
}

// TestInitRefusesToWriteThroughASymlinkedAncestor: onboarding writes two files,
// and the write path needs the same containment as the read path — a location
// behind a committed symlink must not land the scaffolded block outside the
// repository.
func TestInitRefusesToWriteThroughASymlinkedAncestor(t *testing.T) {
	repo, outside := hostileRepo(t, "home", false, nil)

	res, err := Init(repo, InitRequest{
		Title:    "Contained",
		Tagline:  "A tagline.",
		Location: BlockLocation{File: "home/IDENTITY.md", Heading: "Identity (canonical)"},
	})
	if err == nil {
		t.Fatalf("Init wrote through a symlinked ancestor; result = %+v", res)
	}
	if _, serr := os.Lstat(filepath.Join(outside, "IDENTITY.md")); serr == nil {
		t.Fatal("Init created a file outside the repository root")
	}
}

func assertNoCanary(t *testing.T, what, s string) {
	t.Helper()
	if strings.Contains(s, containmentCanary) {
		t.Fatalf("%s carries content from outside the repository root: %s", what, s)
	}
}
