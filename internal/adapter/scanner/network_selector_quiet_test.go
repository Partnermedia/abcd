package scanner

import (
	"strings"
	"testing"
)

// TestSnakeCaseSelectorsStayQuiet pins the iss-358 wontfix decision: the
// hostname patterns keep their trailing word boundary, so a snake_case
// selector whose leading segments happen to spell a LAN suffix is NEVER a
// finding. The alternative — dropping the boundary so `printer.local_backup`
// is caught — was built, adversarially reviewed, and reverted: it flagged
// `stream.local_addr()` (Rust), `args.local_rank` (Python) and
// `storage.local_key` (JS), and Stage-1 redaction rewrites every finding into
// the history store, the one irreversible artefact abcd writes. A warn-tier
// miss on an underscore-suffixed hostname is the accepted cost.
func TestSnakeCaseSelectorsStayQuiet(t *testing.T) {
	for _, line := range []string{
		"let a = stream.local_addr()?;",
		"rank = args.local_rank",
		"const p = storage.local_key;",
		"cfg.local_path = 1",
		"see " + host("scans", "printer", "lan") + "_2026.log", // the accepted miss
	} {
		var hits []Finding
		for _, f := range scanNet(line) {
			if strings.Contains(f.Kind, "net:") {
				hits = append(hits, f)
			}
		}
		if len(hits) != 0 {
			t.Errorf("%q flagged: %v — a snake_case selector must never reach the irreversible redaction path", line, hits)
		}
	}
}
