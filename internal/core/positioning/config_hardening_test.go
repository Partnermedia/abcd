package positioning

import (
	"errors"
	"fmt"
	"strings"
	"testing"
)

// validSurface returns a well-formed surface with the given id and files, so a
// hardening test varies only the field it means to prove.
func validSurface(id string, files ...string) Surface {
	return Surface{
		ID:       id,
		Files:    files,
		Kind:     KindRegexp,
		Patterns: []string{"(x)"},
		Requires: []string{"tagline"},
	}
}

func validConfig(surfaces ...Surface) Config {
	return Config{
		SchemaVersion: 1,
		Block:         BlockLocation{File: "b.md", Heading: "H"},
		Surfaces:      surfaces,
	}
}

// TestValidateRefusesSurfaceUnderGitDir pins that a surface candidate inside
// .git is refused: ValidRelPath alone accepts ".git/config" (it is clean,
// relative, and inside the containment root), which would let abcd identity
// --json quote a credential-bearing remote URL out of .git/config into its
// output (iss-150).
func TestValidateRefusesSurfaceUnderGitDir(t *testing.T) {
	refused := []string{
		".git/config",
		".git",
		".git/hooks/pre-commit",
		".GIT/config", // case-folding filesystems reach the same .git
	}
	for _, f := range refused {
		t.Run(f, func(t *testing.T) {
			err := validConfig(validSurface("s", f)).Validate()
			if err == nil {
				t.Fatalf("Validate accepted a surface file under .git: %q", f)
			}
			if !errors.Is(err, ErrConfigInvalid) {
				t.Fatalf("err = %v, want ErrConfigInvalid", err)
			}
		})
	}

	// A path that merely begins with the bytes ".git" but is a different
	// directory (".github/…", ".gitignore") is legitimate and must still pass.
	for _, f := range []string{".github/workflows/ci.yml", ".gitignore", "README.md"} {
		t.Run("allow_"+f, func(t *testing.T) {
			if err := validConfig(validSurface("s", f)).Validate(); err != nil {
				t.Fatalf("Validate refused a legitimate surface file %q: %v", f, err)
			}
		})
	}
}

// TestValidateBoundsSurfaceCount pins that a registry declaring more than the
// fixed surface cap is refused, so one audit run over a hostile repo cannot be
// made to hold an unbounded multiple of the per-surface 1 MiB read cap
// (iss-149).
func TestValidateBoundsSurfaceCount(t *testing.T) {
	atCap := make([]Surface, 0, maxSurfaces)
	for i := 0; i < maxSurfaces; i++ {
		atCap = append(atCap, validSurface(fmt.Sprintf("s%d", i), "README.md"))
	}
	if err := validConfig(atCap...).Validate(); err != nil {
		t.Fatalf("Validate refused a registry at the cap (%d surfaces): %v", maxSurfaces, err)
	}

	overCap := append(atCap, validSurface("over", "README.md"))
	err := validConfig(overCap...).Validate()
	if err == nil {
		t.Fatalf("Validate accepted a registry over the cap (%d surfaces)", len(overCap))
	}
	if !errors.Is(err, ErrConfigInvalid) {
		t.Fatalf("err = %v, want ErrConfigInvalid", err)
	}
	if !strings.Contains(err.Error(), "surface") {
		t.Errorf("refusal should name the surface bound, got %v", err)
	}
}
