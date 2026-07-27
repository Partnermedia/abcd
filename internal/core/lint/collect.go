package lint

// The citation collector: the one place that answers "what does this repo cite?"
//
// The gate and the refresh verb must agree on that set exactly. If the refresh
// fetched a different set — a second scraper with its own idea of where a
// citation lives — every disagreement would surface as a permanent lint finding
// nobody could clear: a URL the gate demands a receipt for and the refresh never
// visits. So the walk lives here once, LintAt's citation family and
// `abcd docs cite refresh` both read it, and the two cannot drift.
//
// It is a thin composition of the primitives the lint already uses
// (markdownFiles, fenceMask, parseCitations) rather than new parsing.

import (
	"path/filepath"
	"sort"
	"strings"

	"github.com/REPPL/abcd-cli/internal/fsutil"
)

// citationPageSizeLimit caps one page read. Containment answers "does this path
// land inside the repo"; it says nothing about how big what it lands on is, and
// this collector's output is fetched over the network — so the read is bounded
// too. A documentation page is prose; eight mebibytes is far beyond any real one.
const citationPageSizeLimit = 8 << 20

// CitationSite is one place a URL is cited: a repo-relative page and a 1-based
// line. A URL cited from several pages carries one site per occurrence, so the
// manual queue can tell the maintainer where to look.
type CitationSite struct {
	File string `json:"file"`
	Line int    `json:"line"`
}

// CitedURL is one distinct cited address and everywhere it is cited.
type CitedURL struct {
	URL   string         `json:"url"`
	Sites []CitationSite `json:"sites"`
}

// CollectCitedURLs walks cfg.Roots and returns the repo's distinct cited URLs,
// sorted by address so a refresh run and its baseline diff are deterministic.
//
// "Cited" means what the lint means by it: a URL on a footnote-definition line
// (including its wrapped continuation lines), never one in body prose. A missing
// root is skipped rather than an error, matching Lint.
func CollectCitedURLs(cfg Config, repoRoot string) ([]CitedURL, error) {
	sites := map[string][]CitationSite{}
	for _, root := range cfg.Roots {
		// Containment matters HERE in a way it does not for a reporting lint.
		// What this collector returns is fetched over the network and then
		// persisted into the committed baseline — a file the workflow expects the
		// maintainer to push. A `roots` of "../private" in the committed config
		// would therefore fetch every URL in a sibling directory and publish the
		// list. The lint's own walk is a separate, pre-existing path.
		if err := containedRepoPath(root); err != nil {
			return nil, &configError{"roots entry " + quote(root) + " " + err.Error() +
				"; the citation collector reads only inside the repository"}
		}
		rootAbs := filepath.Join(repoRoot, root)
		if err := resolvedInsideRoot(repoRoot, rootAbs); err != nil {
			return nil, &configError{"roots entry " + quote(root) + " " + err.Error() +
				"; the citation collector reads only inside the repository"}
		}
		mdFiles, err := markdownFiles(rootAbs)
		if err != nil {
			return nil, err
		}
		for _, fileAbs := range mdFiles {
			// Containing the ROOT is not enough. WalkDir yields a symlinked .md
			// as an ordinary file and os.ReadFile follows it, so a committed
			// `docs/leak.md -> ../private/notes.md` sits inside a perfectly
			// contained root and still drags an outside file's citations into
			// the fetch and then into the committed baseline. Containment is
			// about where a path LANDS, so an in-repo symlink still resolves.
			realPath, err := containedRealPath(repoRoot, fileAbs)
			if err != nil {
				return nil, &configError{"cited page " + quote(repoRel(repoRoot, fileAbs)) + " " + err.Error() +
					"; the citation collector reads only inside the repository"}
			}
			// ReadGuarded, not os.ReadFile: the regular-file check refuses a
			// device and the cap bounds the read, neither of which containment
			// can supply — a committed page that is simply enormous lands inside
			// the repo and passes every path check. It opens the RESOLVED path,
			// because O_NOFOLLOW would otherwise refuse the legitimate in-repo
			// bridge that containment just approved.
			content, err := fsutil.ReadGuarded(realPath, citationPageSizeLimit)
			if err != nil {
				return nil, err
			}
			rel := repoRel(repoRoot, fileAbs)
			lines := strings.Split(string(content), "\n")
			page := parseCitations(rel, lines, fenceMask(lines))
			for _, ref := range page.urls {
				site := CitationSite{File: ref.file, Line: ref.line}
				// A URL repeated on one line (a source and its mirror) is one
				// site, not two — the queue would otherwise print it twice.
				if existing := sites[ref.url]; len(existing) > 0 && existing[len(existing)-1] == site {
					continue
				}
				sites[ref.url] = append(sites[ref.url], site)
			}
		}
	}

	out := make([]CitedURL, 0, len(sites))
	for u, s := range sites {
		out = append(out, CitedURL{URL: u, Sites: s})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].URL < out[j].URL })
	return out, nil
}
