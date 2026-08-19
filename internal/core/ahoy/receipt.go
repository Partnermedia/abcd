package ahoy

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/Partnermedia/abcd/internal/fsutil"
)

// note records one written artefact on the install receipt.
//
// It is the SINGLE seam through which anything reaches InstallResult.Writes, and
// it renders what it is handed identity-free before recording it (iss-177). An
// apply step therefore passes the absolute path it actually wrote and formats
// nothing itself: a step added later cannot put a home directory or a username
// on a receipt a user pastes into an issue, because the scrub is not a thing a
// step has to remember to do.
func (a *applyCtx) note(path string) {
	a.writes = append(a.writes, receiptPath(a.cwd, path))
}

// receiptPath renders one receipt entry with the two roots that carry developer
// identity — the repo under installation and the home directory — removed. They
// are the same two roots the CLI's error scrub redacts, through the same
// primitive (fsutil.RedactRoot), so an install's receipt and an install's error
// message cannot disagree about what is safe to print:
//
//   - a path under the repo becomes repo-relative (".abcd/config.json");
//   - a path under the home directory becomes "~/…"
//     ("~/.abcd/history/index.json") — the user-scope store's absolute form is
//     exactly the leak, since its second segment IS the username;
//   - anything else is left alone, save for a redaction of either root found
//     EMBEDDED in it. A system location such as /usr/local/bin/abcd names no
//     developer, and this deliberately matches the error scrub's stated limit
//     rather than being a universal absolute-path scrub. The embedded pass is
//     what covers a note that is a sentence rather than a bare path (the
//     dependency fix-hint line, and whatever a later step composes).
func receiptPath(cwd, p string) string {
	if p == "" {
		return p
	}
	home, homeErr := os.UserHomeDir()
	if filepath.IsAbs(p) {
		if under(cwd, p) {
			return filepath.ToSlash(fsutil.RepoRel(cwd, p))
		}
		if homeErr == nil && under(home, p) {
			rel := filepath.ToSlash(fsutil.RepoRel(home, p))
			if rel == "." {
				return "~" // the home directory itself, not "~/."
			}
			return "~/" + rel
		}
	}
	p = fsutil.RedactRoot(p, cwd, ".")
	if homeErr == nil {
		p = fsutil.RedactRoot(p, home, "~")
	}
	return p
}

// under reports whether p lies at or inside the absolute directory root. It is
// lexical: both sides come from the same source in every caller (the cwd Install
// absolutised, or the HOME the store itself reads), so resolving symlinks here
// would compare a resolved path against an unresolved root and answer no.
func under(root, p string) bool {
	if root == "" || !filepath.IsAbs(root) {
		return false
	}
	rel, err := filepath.Rel(root, p)
	if err != nil {
		return false
	}
	return rel != ".." && !strings.HasPrefix(rel, ".."+string(os.PathSeparator))
}
