package site

// The composition manifest, `.abcd/site.json`.
//
// The manifest names WHERE every block of the site comes from and carries no
// prose of its own (adr-47 decision 2). It is decoded with unknown fields
// REFUSED, at every depth: a key the binary does not read is a composition
// instruction somebody wrote and the build silently ignored, which is the exact
// failure mode a manifest exists to prevent. The same strictness is what makes
// the file a contract rather than a suggestion.

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/Partnermedia/abcd/internal/fsutil"
)

// ManifestRelPath is the committed composition manifest.
const ManifestRelPath = ".abcd/site.json"

// maxManifestBytes bounds the manifest read. It is configuration, not content.
const maxManifestBytes = 256 * 1024

// ErrManifestInvalid is returned for a manifest the build cannot act on.
var ErrManifestInvalid = errors.New("site: manifest is invalid")

// The layouts the composer implements. A manifest naming any other layout is
// refused at load rather than at render, so the failure names the manifest.
const (
	LayoutCardsFromH2  = "cards-from-h2"
	LayoutLeadInCards  = "lead-in-cards"
	LayoutProse        = "prose"
	LayoutInstall      = "install"
	figureFirstImage   = "first-image"
	featureShippedPR   = "shipped-intent-press-release"
	featurePickNewest  = "newest-with-audit-MET"
	featurePartPR      = "press-release"
	featurePartFirstAC = "first-acceptance-criterion"
)

// Manifest is `.abcd/site.json`.
type Manifest struct {
	SchemaVersion int          `json:"schema_version"`
	Purpose       string       `json:"purpose"`
	Identity      BlockRef     `json:"identity"`
	UIStrings     string       `json:"ui_strings"`
	Home          Home         `json:"home"`
	Record        RecordOpts   `json:"record"`
	Docs          DocsRefs     `json:"docs"`
	RecordPages   RecordPages  `json:"record_pages"`
	Checks        ManifestGate `json:"checks"`
}

// BlockRef selects a span of a file by heading.
type BlockRef struct {
	File    string `json:"file"`
	Heading string `json:"heading"`
}

// Home is the landing page's composition.
type Home struct {
	Hero     Hero      `json:"hero"`
	Chapters []Chapter `json:"chapters"`
}

// Hero names the page the headline, lede and figure come from.
type Hero struct {
	Page   string `json:"page"`
	Figure string `json:"figure"`
}

// Chapter is one lettered section of the landing page.
type Chapter struct {
	Letter         string   `json:"letter"`
	Page           string   `json:"page"`
	Layout         string   `json:"layout"`
	Feature        *Feature `json:"feature,omitempty"`
	Icons          string   `json:"icons,omitempty"`
	TablePortraits string   `json:"table_portraits,omitempty"`
	Figure         *Figure  `json:"figure,omitempty"`
	Lead           string   `json:"lead,omitempty"`
	Release        *Release `json:"release,omitempty"`
	After          string   `json:"after,omitempty"`
	Tabs           string   `json:"tabs,omitempty"`
	Left           []string `json:"left,omitempty"`
}

// Feature declares the quoted record block a chapter closes with.
type Feature struct {
	Kind  string   `json:"kind"`
	Pick  string   `json:"pick"`
	Parts []string `json:"parts"`
}

// Figure declares a chapter's lifted illustration.
type Figure struct {
	Kind           string `json:"kind"`
	LabelsFromPage bool   `json:"labels-from-page,omitempty"`
}

// Release names where the install chapter reads version and asset names from.
type Release struct {
	From   string `json:"from"`
	Assets string `json:"assets"`
}

// RecordOpts carries the working-tier publication opt-in (adr-32, itd-140).
type RecordOpts struct {
	IssueLedger bool `json:"issue_ledger"`
}

// DocsRefs names the documentation pages the site links its docs surface to.
type DocsRefs struct {
	Index string `json:"index"`
	CLI   string `json:"cli"`
}

// RecordPages carries the explorer pages' selectors.
type RecordPages struct {
	Contributors Contributors `json:"contributors"`
}

// Contributors names the attribution policy span the contributors page quotes.
type Contributors struct {
	Policy PolicyRef `json:"policy"`
}

// PolicyRef selects a part of a span of a file by heading.
type PolicyRef struct {
	File    string `json:"file"`
	Heading string `json:"heading"`
	Part    string `json:"part"`
}

// ManifestGate is the check declaration. The gates themselves are a later
// slice; the manifest records which of them this repo arms.
type ManifestGate struct {
	EveryTextNodeHasSource        bool   `json:"every_text_node_has_source"`
	DocsLintOnRenderedText        bool   `json:"docs_lint_on_rendered_text"`
	CommandSnippetsPinnedToCLIRef bool   `json:"command_snippets_pinned_to_cli_reference"`
	UnresolvedReferenceBaseline   string `json:"unresolved_reference_baseline"`
}

// LoadManifest reads and validates the composition manifest at repoRoot.
func LoadManifest(repoRoot string) (Manifest, error) {
	data, err := fsutil.ReadGuarded(joinRepo(repoRoot, ManifestRelPath), maxManifestBytes)
	if err != nil {
		return Manifest{}, err
	}
	var m Manifest
	dec := json.NewDecoder(strings.NewReader(string(data)))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&m); err != nil {
		return Manifest{}, fmt.Errorf("%w: %s: %v", ErrManifestInvalid, ManifestRelPath, err)
	}
	if err := m.validate(); err != nil {
		return Manifest{}, err
	}
	return m, nil
}

// validate refuses a manifest the composer cannot act on. Every refusal names
// the key, because the reader's next move is an edit to this file.
func (m Manifest) validate() error {
	bad := func(format string, args ...any) error {
		return fmt.Errorf("%w: %s: %s", ErrManifestInvalid, ManifestRelPath, fmt.Sprintf(format, args...))
	}
	if m.SchemaVersion != 1 {
		return bad("schema_version is %d, want 1", m.SchemaVersion)
	}
	if m.Identity.File == "" || m.Identity.Heading == "" {
		return bad("identity.file and identity.heading are both required — the hero renders from the Identity block")
	}
	if !fsutil.ValidRelPath(m.Identity.File) {
		return bad("identity.file %q is not a repo-relative path", m.Identity.File)
	}
	if m.UIStrings == "" || !fsutil.ValidRelPath(m.UIStrings) {
		return bad("ui_strings %q is not a repo-relative path", m.UIStrings)
	}
	if m.Home.Hero.Page == "" || !fsutil.ValidRelPath(m.Home.Hero.Page) {
		return bad("home.hero.page %q is not a repo-relative path", m.Home.Hero.Page)
	}
	if m.Home.Hero.Figure != "" && m.Home.Hero.Figure != figureFirstImage {
		return bad("home.hero.figure %q is not a figure rule (want %q)", m.Home.Hero.Figure, figureFirstImage)
	}
	if len(m.Home.Chapters) == 0 {
		return bad("home.chapters is empty — the landing page is composed of chapters")
	}
	letters := map[string]bool{}
	for i, ch := range m.Home.Chapters {
		where := fmt.Sprintf("home.chapters[%d]", i)
		if ch.Letter == "" {
			return bad("%s.letter is empty", where)
		}
		if letters[ch.Letter] {
			return bad("%s.letter %q is used twice — a chapter's letter is its identity", where, ch.Letter)
		}
		letters[ch.Letter] = true
		if !fsutil.ValidRelPath(ch.Page) {
			return bad("%s.page %q is not a repo-relative path", where, ch.Page)
		}
		switch ch.Layout {
		case LayoutCardsFromH2, LayoutLeadInCards, LayoutProse, LayoutInstall:
		default:
			return bad("%s.layout %q is not a layout the composer implements (%s)", where, ch.Layout,
				strings.Join([]string{LayoutCardsFromH2, LayoutLeadInCards, LayoutProse, LayoutInstall}, ", "))
		}
		if ch.Layout == LayoutInstall && ch.Lead == "" {
			return bad("%s.lead is required for the install layout — it names the section whose sub-headings become the CLI tabs", where)
		}
		if ch.Feature != nil {
			if ch.Feature.Kind != featureShippedPR {
				return bad("%s.feature.kind %q is not a feature the composer implements (want %q)", where, ch.Feature.Kind, featureShippedPR)
			}
			if ch.Feature.Pick != featurePickNewest {
				return bad("%s.feature.pick %q is not a pick rule the composer implements (want %q)", where, ch.Feature.Pick, featurePickNewest)
			}
			for _, p := range ch.Feature.Parts {
				if p != featurePartPR && p != featurePartFirstAC {
					return bad("%s.feature.parts names %q, which is not a quotable part (%s, %s)", where, p, featurePartPR, featurePartFirstAC)
				}
			}
		}
		if ch.Figure != nil && ch.Figure.Kind != figureFirstImage {
			return bad("%s.figure.kind %q is not a figure rule (want %q)", where, ch.Figure.Kind, figureFirstImage)
		}
	}
	return nil
}

// FeatureWants reports whether a chapter's feature block quotes the named part.
func (f *Feature) FeatureWants(part string) bool {
	if f == nil {
		return false
	}
	for _, p := range f.Parts {
		if p == part {
			return true
		}
	}
	return false
}
