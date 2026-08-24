package site

// The one git pass.
//
// Every date the site shows about a record — when it was written, when it
// entered the lifecycle directory it sits in now, when it was last touched — is
// a fact about git history, and history is the expensive thing to read. So it is
// read ONCE: a single `git log --reverse --name-status` walk, oldest commit
// first, replaying every add, rename and delete to reconstruct each file's
// biography. Per-file `git log` calls would be the obvious alternative and would
// multiply one process by the size of the corpus.
//
// Rename detection is what makes the middle date possible at all: the record
// moves a file between lifecycle directories (drafts → planned → shipped) rather
// than editing a status field, so "the day this intent shipped" IS "the day this
// path's directory last changed", and only `-M` can see that.

import (
	"strings"

	"github.com/intentdriven/abcd/internal/gitutil"
)

// maxLogBytes bounds the history read. The walk is one process over the whole
// history, so it is bounded rather than trusted: a repository large enough to
// exceed this deserves a refusal it can see, not an unbounded allocation.
const maxLogBytes = 64 << 20

// FileDates is one file's biography, each date as YYYY-MM-DD.
type FileDates struct {
	// Created is the day the file first appeared, under any path.
	Created string `json:"created"`
	// Entered is the day it arrived in the directory it occupies now. For a
	// file that never moved, it equals Created.
	Entered string `json:"entered"`
	// Touched is the day the last commit changed it.
	Touched string `json:"touched"`
}

// History is the result of the one pass.
type History struct {
	// Files maps a repo-relative path (as of HEAD) to its dates.
	Files map[string]FileDates
	// First is the day of the oldest commit; Last the day of the newest.
	First string
	Last  string
	// Commits is the number of commits the walk saw.
	Commits int
}

// fileHist is a path's biography while the walk is still replaying it.
type fileHist struct {
	created string
	entered string
	touched string
	dir     string
	alive   bool
}

// LoadHistory replays the repository's history and returns every live file's
// dates. A directory that is not a git repository yields an empty history and no
// error: an unversioned tree is a state the site renders around, not a fault.
func LoadHistory(repoRoot string) (History, error) {
	h := History{Files: map[string]FileDates{}}
	if !gitutil.InRepo(repoRoot) {
		return h, nil
	}
	// `--diff-merges=first-parent` is load-bearing, not a refinement. Without it
	// git prints NO file lines for a merge commit, so a file whose content
	// entered the trunk only through a merge — a conflict resolution, a rename
	// settled while merging — is invisible to the whole walk and comes out with
	// no dates at all. That is not a rounding error in a published date; it is a
	// record the site cannot place in time, and this repository has two of them.
	//
	// The walk still visits every commit rather than only the trunk, so a file
	// written on a branch keeps the day it was WRITTEN as its creation date; the
	// merge then reads as the day it last changed, which is what landing on the
	// trunk is.
	out, err := gitutil.RunCapped(repoRoot, maxLogBytes,
		"log", "--reverse", "--name-status", "-M", "--diff-merges=first-parent",
		"--date=short", "--pretty=format:%x00%ad")
	if err != nil {
		return History{}, err
	}

	live := map[string]*fileHist{}
	date := ""
	for _, line := range strings.Split(out, "\n") {
		if strings.HasPrefix(line, "\x00") {
			date = strings.TrimSpace(line[1:])
			if date == "" {
				continue
			}
			if h.First == "" {
				h.First = date
			}
			h.Last = date
			h.Commits++
			continue
		}
		if line == "" || date == "" {
			continue
		}
		fields := strings.Split(line, "\t")
		if len(fields) < 2 {
			continue
		}
		status := fields[0]
		switch {
		case strings.HasPrefix(status, "R") && len(fields) >= 3:
			rename(live, fields[1], fields[2], date)
		case strings.HasPrefix(status, "D"):
			if f, ok := live[fields[1]]; ok {
				f.alive = false
				f.touched = date
			}
		default:
			touch(live, fields[1], date)
		}
	}

	for path, f := range live {
		if !f.alive {
			continue
		}
		h.Files[path] = FileDates{Created: f.created, Entered: f.entered, Touched: f.touched}
	}
	return h, nil
}

// touch records an add or a modification.
func touch(live map[string]*fileHist, path, date string) {
	f, ok := live[path]
	if !ok {
		f = &fileHist{created: date, entered: date, dir: dirOf(path)}
		live[path] = f
	}
	// A path deleted and later re-added keeps its original creation date — the
	// record's own view is that the file came back, not that it was born twice —
	// but re-entered its directory on the day it returned.
	if !f.alive {
		f.entered = date
	}
	f.alive = true
	f.touched = date
}

// rename moves a biography from one path to another, resetting the middle date
// only when the DIRECTORY changed: a file renamed inside its lifecycle bucket
// did not change lifecycle state.
func rename(live map[string]*fileHist, from, to, date string) {
	f, ok := live[from]
	if !ok {
		f = &fileHist{created: date, entered: date}
	}
	delete(live, from)
	if dirOf(to) != f.dir {
		f.entered = date
	}
	f.dir = dirOf(to)
	f.touched = date
	f.alive = true
	live[to] = f
}

// dirOf is the slash-separated parent directory of a repo-relative path.
func dirOf(path string) string {
	if i := strings.LastIndex(path, "/"); i >= 0 {
		return path[:i]
	}
	return ""
}

// EnteredBucket is the day a record arrived in the lifecycle directory it sits
// in now — for a shipped intent, the day it shipped. It is the same fact as
// FileDates.Entered, named for the question the site actually asks of it.
func (h History) EnteredBucket(path string) string {
	return h.Files[path].Entered
}

// EffectiveDate is the date a record is placed by: the date it declares in its
// own frontmatter where it carries one, else the day its file first appeared.
// It is the ordering key of both chart arrangements, so it lives here with the
// history rather than being re-derived per consumer.
func (h History) EffectiveDate(path, frontmatterDate string) string {
	if frontmatterDate != "" {
		return frontmatterDate
	}
	return h.Files[path].Created
}

// HeadCommit is the short SHA at HEAD, or "" outside a repository. The footer
// prints it, so a build from a detached tree still says which bytes it read.
func HeadCommit(repoRoot string) string {
	if !gitutil.InRepo(repoRoot) {
		return ""
	}
	out, err := gitutil.Run(repoRoot, "rev-parse", "--short", "HEAD")
	if err != nil {
		return ""
	}
	return strings.TrimSpace(out)
}
