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

import (
	"encoding/binary"
	"fmt"
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

// svgSizeRe matches the width/height attributes the stylesheet replaces.
var svgSizeRe = regexp.MustCompile(`\s(width|height)="\d+"`)

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
	data, err := fsutil.ReadGuarded(joinRepo(a.repoRoot, rel), maxAssetBytes)
	if err != nil {
		return "", fmt.Errorf("%s:%d: image %q (%s): %w", at.Path, at.Line, src, rel, err)
	}

	name := path.Base(rel)
	if strings.HasSuffix(name, ".svg") {
		svg := svgSizeRe.ReplaceAllString(string(data), "")
		stem := strings.TrimSuffix(name, ".svg")
		return `<span class="svgasset ` + escapeAttr(stem) + `" data-asset="` + escapeAttr(rel) + `">` + svg + `</span>`, nil
	}

	dst := assetOutDir + "/" + name
	if prev, ok := a.copies[rel]; ok {
		dst = prev
	} else {
		for other, taken := range a.copies {
			if taken == dst && other != rel {
				// Two different files with the same basename would silently
				// overwrite one another in the flat output directory.
				return "", fmt.Errorf("%s:%d: assets %q and %q share a file name; the output tree is flat", at.Path, at.Line, other, rel)
			}
		}
		a.copies[rel] = dst
	}
	size := ""
	if w, h, ok := pngSize(data); ok {
		size = fmt.Sprintf(` width="%d" height="%d"`, w, h)
	}
	return `<img src="` + escapeAttr(dst) + `" alt="` + escapeAttr(alt) + `"` + size + ` loading="lazy">`, nil
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
