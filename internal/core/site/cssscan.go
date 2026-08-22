package site

// The stylesheet reader the static mobile checks consult.
//
// AC 7's static half is a claim about the SERVED page: an image can shrink, a
// wide element can scroll. Half of that lives in the markup and half in the
// stylesheet the markup links, so asserting it from the HTML alone would be
// asserting half a fact. This reads the other half.
//
// It is not a CSS engine and does not try to be. It reads declaration blocks,
// descends into at-rules, and answers three questions: which selectors are
// given a scrolling overflow, whether `img` is given a max-width, and how wide
// the widest column the design offers is. Everything it cannot resolve it
// declines to credit, so an unreadable rule makes a check STRICTER rather than
// weaker — a stylesheet this cannot understand fails the page rather than
// passing it.

import (
	"regexp"
	"strconv"
	"strings"
)

// styleSheet is what one linked stylesheet says about narrow viewports.
type styleSheet struct {
	// OverflowElements are bare element selectors given a scrolling overflow.
	OverflowElements map[string]bool
	// OverflowClasses are single-class selectors given a scrolling overflow.
	OverflowClasses map[string]bool
	// ImageMaxWidth reports whether `img` is given a max-width.
	ImageMaxWidth bool
	// ContentColumnPx is the widest column the sheet declares in pixels — the
	// width a picture would take if the max-width rule were ever dropped.
	ContentColumnPx int
}

var (
	cssCommentRe  = regexp.MustCompile(`(?s)/\*.*?\*/`)
	cssMaxWidthRe = regexp.MustCompile(`(?i)^\s*max-width\s*:\s*(\d+)px\s*$`)
	cssClassRe    = regexp.MustCompile(`^\.([A-Za-z_][A-Za-z0-9_-]*)$`)
	cssElementRe  = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9]*$`)
)

// parseStylesheet reads one stylesheet.
func parseStylesheet(src string) *styleSheet {
	s := &styleSheet{OverflowElements: map[string]bool{}, OverflowClasses: map[string]bool{}}
	s.scan(cssCommentRe.ReplaceAllString(src, " "))
	return s
}

// scan walks a run of rules. An at-rule whose body holds rules of its own — a
// media query, a supports block — is descended into, so a rule that only exists
// at one width is still read.
func (s *styleSheet) scan(src string) {
	for i := 0; i < len(src); {
		open := strings.IndexByte(src[i:], '{')
		if open < 0 {
			return
		}
		open += i
		selector := strings.TrimSpace(src[i:open])
		end := matchBrace(src, open)
		if end < 0 {
			return
		}
		body := src[open+1 : end]
		if strings.HasPrefix(selector, "@") {
			if strings.ContainsRune(body, '{') {
				s.scan(body)
			}
		} else {
			s.rule(selector, body)
		}
		i = end + 1
	}
}

// matchBrace returns the index of the '}' closing the '{' at open.
func matchBrace(src string, open int) int {
	depth := 0
	for i := open; i < len(src); i++ {
		switch src[i] {
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				return i
			}
		}
	}
	return -1
}

// rule reads one selector list and its declarations.
func (s *styleSheet) rule(selectorList, body string) {
	scrolls := false
	for _, decl := range strings.Split(body, ";") {
		prop, value, ok := strings.Cut(decl, ":")
		if !ok {
			continue
		}
		prop = strings.ToLower(strings.TrimSpace(prop))
		value = strings.ToLower(strings.TrimSpace(value))
		switch prop {
		case "overflow", "overflow-x":
			// `hidden` clips the content rather than letting a reader reach it,
			// which is not a wide element made readable on a phone.
			if value == "auto" || value == "scroll" {
				scrolls = true
			}
		}
		if m := cssMaxWidthRe.FindStringSubmatch(decl); m != nil {
			if n, err := strconv.Atoi(m[1]); err == nil && n > s.ContentColumnPx {
				s.ContentColumnPx = n
			}
		}
	}

	for _, sel := range strings.Split(selectorList, ",") {
		sel = strings.TrimSpace(sel)
		// Only a SINGLE compound selector is credited. `.a .b { overflow-x: auto }`
		// gives the rule to a `.b` inside an `.a`, and this reader cannot tell
		// whether the element it is looking at is inside one — so it declines,
		// and the page has to say so in a way that is plainly true.
		if sel == "" || strings.ContainsAny(sel, " >+~:[") {
			continue
		}
		switch {
		case cssElementRe.MatchString(sel):
			if scrolls {
				s.OverflowElements[strings.ToLower(sel)] = true
			}
			if strings.EqualFold(sel, "img") && strings.Contains(strings.ToLower(body), "max-width") {
				s.ImageMaxWidth = true
			}
		case cssClassRe.MatchString(sel):
			if scrolls {
				s.OverflowClasses[cssClassRe.FindStringSubmatch(sel)[1]] = true
			}
		}
	}
}
