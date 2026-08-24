package cli

import (
	"fmt"

	"github.com/intentdriven/abcd/internal/core/ahoy"
	"github.com/intentdriven/abcd/internal/core/vintage"
	"github.com/intentdriven/abcd/internal/termsafe"
)

// stalenessNotice is the session-start notice for a dogfood binary that trails
// (or cannot be located against) its own source checkout tip: "" when the binary
// is fresh, or when cwd is not abcd's source checkout. It is git-only and never
// touches the network — the disk-only posture of adr-38.
func stalenessNotice(cwd, exePath string) string {
	return formatStalenessNotice(ahoy.DogfoodStaleness(cwd), exePath)
}

// formatStalenessNotice renders the notice from a report, split out so the
// fire-on-stale/unknown and silent-on-fresh behaviour is testable without the
// running binary's own build info. The binary path is sanitised like every
// other untrusted string this repo renders to a terminal.
func formatStalenessNotice(rep ahoy.DogfoodReport, exePath string) string {
	if !rep.IsDogfood || rep.Outcome == vintage.Fresh {
		return ""
	}
	bin := termsafe.Sanitize(exePath)
	if bin == "" {
		bin = "the abcd binary"
	}
	if rep.Outcome == vintage.Stale {
		return fmt.Sprintf("abcd: %s was built from commit %s but this checkout is at %s — the binary is behind its own source. Rebuild it with `make build`.",
			bin, shortSHA(rep.Revision), shortSHA(rep.Tip))
	}
	// Unknown: a modified (dirty) rebuild — the revision no longer identifies
	// what is running, so it cannot be reported as up to date.
	return fmt.Sprintf("abcd: %s is a modified (dirty) rebuild atop commit %s, so its vintage is unknown and may not match this checkout at %s. Rebuild it cleanly with `make build`.",
		bin, shortSHA(rep.Revision), shortSHA(rep.Tip))
}
