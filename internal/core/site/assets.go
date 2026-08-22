package site

// Pictures.
//
// adr-47 decision 2: every picture is a committed asset under `docs/assets/img/`,
// referenced from a docs page like any other image, and **the build never
// draws**. So this file does two things and refuses a third. An SVG is INLINED,
// because its colours are written as `var(--token, fallback)` and inlining is
// what lets them follow the reader's theme; its width and height attributes are
// stripped so the stylesheet sizes it. A raster is COPIED VERBATIM into the
// output tree and referenced — no re-encoding, because optimisation would mean a
// dependency and a decision nobody has made yet.
//
// An image the page names but the repository does not carry is a build error.
// The alternative is a broken image on a published page, which nobody notices
// until a reader does.
//
// Two refusals guard the pair of holes an image reference opens, and both are
// about the same thing: this file turns repository CONTENT into published
// OUTPUT, and repository content is written by whoever can land a pull request.
//
//  1. A reference is a PATH, and a path read out of prose will name whatever it
//     is pointed at. `![x](../../.git/config)` is a perfectly ordinary-looking
//     image reference that publishes git's config — which on a CI runner carries
//     the checkout token — into a public directory, under a name nobody reads
//     twice, in a tree git never reports as changed. So a reference must resolve
//     inside the asset root and carry a picture's extension, and anything else
//     is refused by name.
//
//  2. An SVG is INLINED, which is the whole point of committing drawings as SVG
//     — but inlining means its bytes become markup. A `<script>` element or an
//     `on*` attribute inside one is executable code on the site's own origin,
//     reached with no click, in a file that is not code and that no reviewer
//     reads as code. So an SVG is parsed and held to an allowlist of the
//     elements and attributes a drawing is made of.

import (
	"encoding/binary"
	"encoding/xml"
	"fmt"
	"io"
	"path"
	"regexp"
	"sort"
	"strings"

	"github.com/Partnermedia/abcd/internal/fsutil"
)

// maxAssetBytes bounds one asset read.
const maxAssetBytes = 16 << 20

// assetOutDir is where copied rasters land in the output tree, and the prefix
// the page references them by. It is relative, so the page works from any mount
// point.
const assetOutDir = "assets/img"

// assetRootPrefix is the only directory a picture may come from. adr-47
// decision 2 says every picture is a committed asset under `docs/assets/img/`,
// referenced from a docs page like any other image; this is that sentence made
// enforceable.
const assetRootPrefix = "docs/assets/"

// pictureExts is what a picture's file name ends in. It is a closed list rather
// than a "not one of these dangerous ones" list, because the dangerous set is
// the whole rest of the repository.
var pictureExts = map[string]bool{
	".svg": true, ".png": true, ".jpg": true, ".jpeg": true,
	".gif": true, ".webp": true, ".avif": true,
}

// svgSizeRe matches the width/height attributes the stylesheet replaces.
var svgSizeRe = regexp.MustCompile(`\s(width|height)="\d+"`)

// svgElements is what a drawing is made of: shapes, the scaffolding that
// positions and paints them, and text. Anything else — a script, a stylesheet,
// an embedded document, a foreign namespace — is refused, because a drawing
// does not need it and an attack does.
var svgElements = map[string]bool{
	"svg": true, "g": true, "defs": true, "symbol": true, "title": true, "desc": true,
	"path": true, "rect": true, "circle": true, "ellipse": true, "line": true,
	"polyline": true, "polygon": true, "text": true, "tspan": true,
	"clipPath": true, "mask": true, "marker": true, "pattern": true, "use": true,
	"linearGradient": true, "radialGradient": true, "stop": true,
	// <image> is allowed only because the committed drawings place raster
	// panels inside them; its href is held to an embedded raster (below), never
	// to a fetch and never to another SVG.
	"image": true,
}

// embeddedRasterPrefixes are the only `data:` payloads an <image> may carry: a
// self-contained raster. `data:image/svg+xml` is deliberately absent — that is
// a document, not a picture, and nesting one inside a drawing reopens exactly
// the surface this check exists to close.
var embeddedRasterPrefixes = []string{
	"data:image/png;base64,", "data:image/jpeg;base64,", "data:image/gif;base64,",
	"data:image/webp;base64,", "data:image/avif;base64,",
}

// svgHrefAttrs are the attributes that point somewhere. Every one of them is
// held to a same-document fragment: a drawing refers to its own defs and to
// nothing else, so an absolute or relative URL here is either an external fetch
// from a page that makes none, or a scheme that executes.
var svgHrefAttrs = map[string]bool{"href": true, "xlink:href": true, "src": true}

// checkInlinableSVG parses an SVG and refuses anything a drawing does not need.
// It parses rather than greps: entities, CDATA, comments and attribute quoting
// are exactly what a substring check gets wrong, and an XML decoder gets right.
func checkInlinableSVG(svg, rel string) error {
	bad := func(format string, args ...any) error {
		return fmt.Errorf("%s: %s — an inlined drawing becomes markup in the published page, so it is held to the elements and attributes a drawing is made of",
			rel, fmt.Sprintf(format, args...))
	}
	dec := xml.NewDecoder(strings.NewReader(svg))
	// Entity expansion is how an XML parser is made to read files it was never
	// given. A drawing needs no custom entity, so the map stays empty and any
	// reference to one is an error from the decoder itself.
	dec.Strict = true
	dec.Entity = map[string]string{}
	for {
		tok, err := dec.Token()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return bad("is not well-formed XML (%v)", err)
		}
		switch t := tok.(type) {
		case xml.ProcInst:
			if strings.EqualFold(t.Target, "xml") {
				continue // the XML declaration itself
			}
			return bad("carries a processing instruction (<?%s?>)", t.Target)
		case xml.Directive:
			// `<!DOCTYPE … [<!ENTITY x SYSTEM "file:///etc/passwd">]>` is the
			// classic external-entity read, and no drawing has a doctype.
			return bad("carries a document-type declaration")
		case xml.StartElement:
			name := t.Name.Local
			if !svgElements[name] {
				return bad("uses <%s>, which is not one of the drawing elements", name)
			}
			for _, a := range t.Attr {
				if err := checkSVGAttr(a, name, bad); err != nil {
					return err
				}
			}
		}
	}
}

// checkSVGAttr holds one attribute to the allowlist rules.
func checkSVGAttr(a xml.Attr, element string, bad func(string, ...any) error) error {
	local := a.Name.Local
	full := local
	if a.Name.Space != "" {
		// The decoder resolves a prefix to its namespace URL; for the reader's
		// sake report the conventional spelling.
		if strings.Contains(a.Name.Space, "xlink") {
			full = "xlink:" + local
		}
	}
	if len(local) >= 2 && strings.EqualFold(local[:2], "on") {
		return bad("sets %s on <%s>, which is an event handler", strings.ToLower(full), element)
	}
	if svgHrefAttrs[full] || svgHrefAttrs[local] {
		v := strings.TrimSpace(a.Value)
		if strings.HasPrefix(v, "#") {
			return nil
		}
		if element == "image" {
			for _, p := range embeddedRasterPrefixes {
				if strings.HasPrefix(v, p) {
					return nil
				}
			}
			return bad("points %s on <image> at %q; an embedded panel is a self-contained raster (data:image/png|jpeg|gif|webp|avif;base64,…)",
				full, clip(v))
		}
		return bad("points %s on <%s> at %q; a drawing refers only to its own definitions (#id)",
			full, element, clip(v))
	}
	return nil
}

// assetPipe resolves image references and records what the build must copy.
type assetPipe struct {
	repoRoot string
	// copies maps a repo-relative source to its output-relative destination.
	copies map[string]string
}

func newAssetPipe(repoRoot string) *assetPipe {
	return &assetPipe{repoRoot: repoRoot, copies: map[string]string{}}
}

// Copies lists the rasters to copy, sorted, as (source, destination) pairs.
func (a *assetPipe) Copies() [][2]string {
	out := make([][2]string, 0, len(a.copies))
	for src, dst := range a.copies {
		out = append(out, [2]string{src, dst})
	}
	sort.Slice(out, func(i, j int) bool { return out[i][0] < out[j][0] })
	return out
}

// render turns one image reference into HTML. pageDir is the repo-relative
// directory of the page the reference was written in, so `../assets/img/x.png`
// resolves the way a reader on GitHub sees it.
func (a *assetPipe) render(pageDir, src, alt string, at Source) (string, error) {
	if strings.HasPrefix(src, "http://") || strings.HasPrefix(src, "https://") || strings.HasPrefix(src, "data:") {
		return "", &UnsupportedError{at.Path, at.Line, "remote image",
			"every picture is a committed asset under docs/assets/img/; " + quote(src) + " is fetched at render time"}
	}
	rel := path.Clean(path.Join(pageDir, src))
	if !fsutil.ValidRelPath(rel) {
		return "", fmt.Errorf("%s:%d: image %q resolves outside the repository", at.Path, at.Line, src)
	}
	// The build PUBLISHES what it reads here, so what may be read is a closed
	// set: a picture, under the asset root, and nothing else. Without this a
	// reference in prose reaches `.git/config`, `.env`, or the gitignored local
	// tier, and copies it into a directory served to the public.
	if !strings.HasPrefix(rel, assetRootPrefix) {
		return "", fmt.Errorf("%s:%d: image %q resolves to %q, outside %s — every picture is a committed asset under the asset root",
			at.Path, at.Line, src, rel, assetRootPrefix)
	}
	ext := strings.ToLower(path.Ext(rel))
	if !pictureExts[ext] {
		return "", fmt.Errorf("%s:%d: image %q is a %q file, which is not a picture the build publishes",
			at.Path, at.Line, src, ext)
	}
	data, err := fsutil.ReadGuarded(joinRepo(a.repoRoot, rel), maxAssetBytes)
	if err != nil {
		return "", fmt.Errorf("%s:%d: image %q (%s): %w", at.Path, at.Line, src, rel, err)
	}

	name := path.Base(rel)
	if ext == ".svg" {
		if err := checkInlinableSVG(string(data), rel); err != nil {
			return "", fmt.Errorf("%s:%d: %w", at.Path, at.Line, err)
		}
		svg := svgSizeRe.ReplaceAllString(string(data), "")
		stem := strings.TrimSuffix(name, ".svg")
		return `<span class="svgasset ` + escapeAttr(stem) + `" data-asset="` + escapeAttr(rel) + `">` + svg + `</span>`, nil
	}

	dst := assetOutDir + "/" + name
	if prev, ok := a.copies[rel]; ok {
		dst = prev
	} else {
		// Two different files with the same basename would silently overwrite one
		// another in the flat output directory. The clash is reported against the
		// LOWEST-sorting other source rather than whichever the map happened to
		// yield first, so two identical builds name the same pair.
		clash := ""
		for other, taken := range a.copies {
			if taken == dst && other != rel && (clash == "" || other < clash) {
				clash = other
			}
		}
		if clash != "" {
			return "", fmt.Errorf("%s:%d: assets %q and %q share a file name; the output tree is flat", at.Path, at.Line, clash, rel)
		}
		a.copies[rel] = dst
	}
	size := ""
	if w, h, ok := pngSize(data); ok {
		size = fmt.Sprintf(` width="%d" height="%d"`, w, h)
	}
	// The href is root-absolute, because the same asset is referenced from every
	// depth the site serves: the landing page at `/`, and a record's page at
	// `/record/adr/adr-1/`. An output-relative src resolves only at the root and
	// 404s everywhere else — a picture missing on hundreds of pages and present
	// on the one page anybody previews.
	return `<img src="/` + escapeAttr(dst) + `" alt="` + escapeAttr(alt) + `"` + size + ` loading="lazy">`, nil
}

// pngSize reads a PNG's pixel dimensions out of its IHDR chunk, so the page can
// reserve the right box and not reflow when the image arrives. A format this
// does not understand simply gets no attributes.
func pngSize(data []byte) (int, int, bool) {
	const header = "\x89PNG\r\n\x1a\n"
	if len(data) < 24 || string(data[:8]) != header || string(data[12:16]) != "IHDR" {
		return 0, 0, false
	}
	w := binary.BigEndian.Uint32(data[16:20])
	h := binary.BigEndian.Uint32(data[20:24])
	if w == 0 || h == 0 {
		return 0, 0, false
	}
	return int(w), int(h), true
}
