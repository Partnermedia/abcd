package site

// Principles — the one record family that carries no frontmatter.
//
// A principle is a markdown file whose name is its handle and whose H1 is its
// title; there is no id field to read and no lifecycle directory to sit in, so
// the lint engine's frontmatter scan cannot see one. They are read here instead
// — a directory listing and a first heading, through the same section walk every
// other page uses — and joined to the record graph as nodes of their own, so the
// dashboard can count them, the chart can draw them and each gets a page.
//
// The directory is DERIVED, never configured: it is `principles/` under the
// record root the lint configuration already names. A repository that keeps none
// yields none, and every page that would list them is omitted.

import (
	"os"
	"path"
	"sort"
	"strings"

	"github.com/Partnermedia/abcd/internal/core/lint"
	"github.com/Partnermedia/abcd/internal/fsutil"
)

// principlesDirName is the record format's own name for the store, and
// principleType the store name its records carry.
const (
	principlesDirName = "principles"
	principleType     = "principle"
)

// disciplinesLifecycle is the bucket an intent sits in when it states a rule
// that holds across the whole record rather than a change to ship.
const disciplinesLifecycle = "disciplines"

// maxPrincipleBytes bounds one principle read.
const maxPrincipleBytes = 1 << 20

// PrinciplesDir is the record's principle store, or "" where the lint
// configuration names no record root.
func PrinciplesDir(cfg lint.Config) string {
	if len(cfg.Roots) == 0 {
		return ""
	}
	return path.Join(cfg.Roots[0], principlesDirName)
}

// LoadPrinciples reads the principle store as record nodes.
//
// An absent directory is a state, not a fault: it yields no nodes and no error,
// and the foundations page is omitted rather than rendered empty.
func LoadPrinciples(repoRoot, dir string) ([]lint.RecordNode, error) {
	if dir == "" {
		return nil, nil
	}
	entries, err := os.ReadDir(joinRepo(repoRoot, dir))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var out []lint.RecordNode
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() || !strings.HasSuffix(name, ".md") {
			continue
		}
		// The store's own index is not one of its records.
		if strings.EqualFold(name, "README.md") {
			continue
		}
		rel := dir + "/" + name
		data, err := fsutil.ReadGuarded(joinRepo(repoRoot, rel), maxPrincipleBytes)
		if err != nil {
			return nil, err
		}
		// No lifecycle, no status. The store grades its records by neither, and
		// a word invented here to fill the column would be published as though
		// the record had declared it.
		id := strings.TrimSuffix(name, ".md")
		out = append(out, lint.RecordNode{
			ID: id, Type: principleType,
			Title: firstHeading(rel, string(data), id), Path: rel,
		})
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out, nil
}

// firstHeading is a document's H1, or the fallback where it carries none.
func firstHeading(rel, text, fallback string) string {
	body, consumed := StripFrontmatter(text)
	secs, err := Sections(rel, body, consumed)
	if err != nil {
		return fallback
	}
	for _, s := range secs {
		if s.Level > 0 && s.Title != "" {
			return s.Title
		}
	}
	return fallback
}
