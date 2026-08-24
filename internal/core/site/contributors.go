package site

// Authorship, as the repository itself records it.
//
// Two facts are published and they are different facts. WHO wrote the code is
// `git shortlog` folded through `.mailmap` — the authors of record, humans
// responsible for the work. WHAT assisted is the `Assisted-by:` trailer, which
// this repository requires on every AI-assisted commit and requires as an
// explicit `None` on every human-only one, so that silence is never mistaken for
// a declaration. Presenting the second as authorship would be exactly the claim
// the trailer convention exists to refuse.
//
// The bots-and-tools row is derived, not listed. An author is a tool when its
// name carries the forge's own `[bot]` suffix, or when the repository's own
// trailers name it as an assisting vendor — which is what a pre-policy commit
// authored by the tool looks like from here. Deriving it means a second tool
// that ever lands a commit appears in the right row without an edit, and means
// no vendor name is written into this file.

import (
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/intentdriven/abcd/internal/gitutil"
)

// maxShortlogBytes bounds the authorship reads.
const maxShortlogBytes = 4 << 20

var (
	shortlogRe = regexp.MustCompile(`^\s*(\d+)\s+(.*?)\s*<([^>]*)>\s*$`)
	botNameRe  = regexp.MustCompile(`\[bot\]$`)
)

// Author is one authorship line.
//
// The address is deliberately absent. `git shortlog -sne` yields it and the
// mailmap folds it, so it is read — but the site is published, and republishing
// a contributor's address on a web page is a harvesting surface the record
// never asked for. The name is the attribution; the address stays in git.
type Author struct {
	Name    string `json:"name"`
	Commits int    `json:"commits"`
	// Profile is the author's forge profile URL, derived only from a forge
	// noreply address — which carries no real mailbox, only the public
	// username — so exporting it republishes nothing the rule above protects.
	// An author whose address is not a noreply simply has none.
	Profile string `json:"profile,omitempty"`
	// email is read so identities fold correctly and sort stably, and is never
	// exported.
	email string
}

// noreplyRe matches GitHub's noreply commit addresses, both forms:
// `77722411+user@users.noreply.github.com` and the older
// `user@users.noreply.github.com`. The pattern itself names the forge, so the
// derivation needs no forge configuration and degrades by absence everywhere
// else (itd-140).
var noreplyRe = regexp.MustCompile(`(?i)^(?:\d+\+)?([a-z\d](?:[a-z\d-]*[a-z\d])?)@users\.noreply\.github\.com$`)

// profileURL is the forge profile a noreply address names, or "".
func profileURL(email string) string {
	m := noreplyRe.FindStringSubmatch(email)
	if m == nil {
		return ""
	}
	return "https://github.com/" + m[1]
}

// ModelTally is one distinct `Assisted-by:` value and how often it appears.
type ModelTally struct {
	Model   string `json:"model"`
	Commits int    `json:"commits"`
}

// Authorship is the whole picture: who authored, what assisted, and the commits
// that declared no assistance at all.
type Authorship struct {
	// Commits is the total number of commits in the history walked.
	Commits int `json:"commits"`
	// Authored is Commits less the merges: the commits a person actually wrote,
	// and the only honest denominator for a disclosure rate.
	Authored int `json:"authored"`
	// Merges is what was set aside to get there. It is published rather than
	// quietly subtracted, because a denominator that changed without saying so
	// is how the rate went wrong in the first place.
	Merges int `json:"merges"`
	// Humans are the authors of record, mailmap-folded, most commits first.
	Humans []Author `json:"humans"`
	// Bots are the forge bots and tool-authored commits, kept in a separate row
	// so a reader never has to guess which lines are people.
	Bots []Author `json:"bots"`
	// AssistedCommits is how many AUTHORED commits declare assistance — the
	// commit-level count, and the numerator of the disclosure rate.
	AssistedCommits int `json:"assisted_commits"`
	// Assisted is the number of `Assisted-by:` trailer OCCURRENCES across
	// authored commits. It is what the per-model tally sums to; it is not a
	// number of commits, and a commit naming two models counts twice here and
	// once in AssistedCommits. Rendering this one as a count of commits is the
	// defect that published a disclosure rate well below the truth.
	Assisted int `json:"assisted"`
	// MultiTrailerCommits is how many AUTHORED commits declare more than one
	// model. It is counted per COMMIT rather than derived as Assisted minus
	// AssistedCommits: that subtraction is a surplus of occurrences, and a
	// single commit naming three models would report two — publishing an
	// occurrence figure under a label that says commits, which is the exact
	// conflation this type exists to keep apart.
	MultiTrailerCommits int `json:"multi_trailer_commits"`
	// DeclaredNone is the number of AUTHORED commits declaring
	// `Assisted-by: None` — work no tool touched, saying so.
	DeclaredNone int `json:"declared_none"`
	// Undeclared is the number of AUTHORED commits carrying no trailer at all.
	// It is published because an absent trailer and a forgotten one are the
	// same bytes, and the honest number is the one that says how much of the
	// history predates the convention. Merges are excluded: nobody wrote them,
	// so nothing was forgotten.
	Undeclared int `json:"undeclared"`
	// ByModel tallies each distinct declared value by OCCURRENCE, most first.
	ByModel []ModelTally `json:"by_model"`
}

// noneDeclaration is the trailer value a human-only commit carries. It is the
// only accepted non-vendor value, so it can be compared for exactly.
const noneDeclaration = "None"

// LoadAuthorship reads the authorship facts out of git. A directory that is not
// a repository yields a zero Authorship and no error.
func LoadAuthorship(repoRoot string) (Authorship, error) {
	// Empty rather than absent: a page that renders "no bots" from an empty list
	// and nothing at all from a null is a page with two ways to say one thing.
	a := Authorship{Humans: []Author{}, Bots: []Author{}, ByModel: []ModelTally{}}
	if !gitutil.InRepo(repoRoot) {
		return a, nil
	}

	// The parent list comes back with the trailers because a MERGE is not a
	// commit anyone wrote: the forge creates it, no convention asks it to
	// declare anything, and counting merges as undeclared work buries the real
	// number under them.
	trailers, err := gitutil.RunCapped(repoRoot, maxShortlogBytes,
		"log", "--pretty=format:%x00%p%x1e%(trailers:key=Assisted-by,valueonly,separator=%x1f)")
	if err != nil {
		return Authorship{}, err
	}
	vendors := map[string]bool{}
	tally := map[string]int{}
	// One record per commit, each opened by the NUL the format writes; the first
	// field of the split is the text before the first record and is not a commit.
	records := strings.Split(trailers, "\x00")
	if len(records) > 0 {
		records = records[1:]
	}
	for _, rec := range records {
		a.Commits++
		parents, rest, _ := strings.Cut(rec, "\x1e")
		if len(strings.Fields(parents)) > 1 {
			a.Merges++
			continue
		}
		a.Authored++
		declared, assisted, models := false, false, 0
		for _, v := range strings.Split(rest, "\x1f") {
			v = strings.TrimSpace(v)
			if v == "" {
				continue
			}
			declared = true
			if v == noneDeclaration {
				// Counted, never charted: a declaration of NO assistance in a
				// tally of what assisted would make the bars sum past their own
				// total. It is stated separately, beneath the chart.
				continue
			}
			tally[v]++
			a.Assisted++
			assisted = true
			models++
			vendors[strings.SplitN(v, ":", 2)[0]] = true
		}
		if models > 1 {
			a.MultiTrailerCommits++
		}
		switch {
		case assisted:
			a.AssistedCommits++
		case declared:
			a.DeclaredNone++
		default:
			a.Undeclared++
		}
	}

	for model, n := range tally {
		a.ByModel = append(a.ByModel, ModelTally{Model: model, Commits: n})
	}
	sort.Slice(a.ByModel, func(i, j int) bool {
		if a.ByModel[i].Commits != a.ByModel[j].Commits {
			return a.ByModel[i].Commits > a.ByModel[j].Commits
		}
		return a.ByModel[i].Model < a.ByModel[j].Model
	})

	shortlog, err := gitutil.RunCapped(repoRoot, maxShortlogBytes, "shortlog", "-sne", "HEAD")
	if err != nil {
		return Authorship{}, err
	}
	for _, line := range strings.Split(shortlog, "\n") {
		m := shortlogRe.FindStringSubmatch(line)
		if m == nil {
			continue
		}
		n, err := strconv.Atoi(m[1])
		if err != nil {
			continue
		}
		au := Author{Name: m[2], Commits: n, Profile: profileURL(m[3]), email: m[3]}
		if botNameRe.MatchString(au.Name) || vendors[au.Name] {
			a.Bots = append(a.Bots, au)
			continue
		}
		a.Humans = append(a.Humans, au)
	}
	sortAuthors(a.Humans)
	sortAuthors(a.Bots)
	return a, nil
}

// sortAuthors orders an authorship column: most commits first, then by name, so
// the page is a function of the history and not of git's output order.
func sortAuthors(as []Author) {
	sort.Slice(as, func(i, j int) bool {
		if as[i].Commits != as[j].Commits {
			return as[i].Commits > as[j].Commits
		}
		if as[i].Name != as[j].Name {
			return as[i].Name < as[j].Name
		}
		return as[i].email < as[j].email
	})
}
