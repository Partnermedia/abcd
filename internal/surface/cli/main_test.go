package cli

import (
	"os"
	"testing"

	"github.com/Partnermedia/abcd/internal/core/ahoy"
	"github.com/Partnermedia/abcd/internal/core/vintage"
)

// TestMain defaults the ahoy build-vintage seam to a fresh, determinable,
// non-dogfood value for the whole package. runCLI executes the commands
// in-process through this unstamped go-test binary, so without the override
// every `ahoy install` exercised here would hit the itd-111 unknown-vintage
// refusal.
func TestMain(m *testing.M) {
	ahoy.SetCurrentVintageForTest(func() vintage.Current {
		return vintage.Current{Revision: "testvintage", Known: true}
	})
	os.Exit(m.Run())
}
