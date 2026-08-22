// Package term holds the canonical terminal-capability primitives: the
// colour-mode ladder and the TTY check (adr-49, brief invariant 13). It is
// deliberately banner-independent — the bare-invocation banner (itd-112) is
// its first consumer and the styled grill (itd-110) its declared second; a
// surface that wants decoration resolves its mode here rather than minting a
// parallel copy.
package term

import (
	"os"
	"strings"
)

// ColorMode is one rung of the colour ladder.
type ColorMode int

const (
	// Mono means no colour at all: decoration degrades to plain glyphs,
	// never to blank output.
	Mono ColorMode = iota
	// Ansi16 is the 16-colour floor for colour-capable terminals.
	Ansi16
	// Ansi256 is the xterm-256 palette.
	Ansi256
	// TrueColor is 24-bit colour, rendered straight from hex.
	TrueColor
)

// ResolveColorMode resolves the ladder for stdout decoration. Precedence,
// descending: the surface's --no-color flag; the NO_COLOR convention
// (present AND non-empty — presence alone does not count); TERM dumb or
// unset (cron, bare containers) forces Mono even when COLORTERM leaks
// through a multiplexer; COLORTERM truecolor/24bit; a 256color TERM; else
// the 16-colour floor.
func ResolveColorMode(getenv func(string) string, noColorFlag bool) ColorMode {
	if noColorFlag {
		return Mono
	}
	if getenv("NO_COLOR") != "" {
		return Mono
	}
	switch t := getenv("TERM"); {
	case t == "" || t == "dumb":
		return Mono
	}
	switch strings.ToLower(getenv("COLORTERM")) {
	case "truecolor", "24bit":
		return TrueColor
	}
	if strings.Contains(getenv("TERM"), "256color") {
		return Ansi256
	}
	return Ansi16
}

// IsTerminal reports whether f is an interactive character device. This is
// the one canonical check; call sites that used to hand-roll the Stat/
// ModeCharDevice test delegate here.
func IsTerminal(f *os.File) bool {
	if f == nil {
		return false
	}
	fi, err := f.Stat()
	return err == nil && fi.Mode()&os.ModeCharDevice != 0
}

// UTF8Locale reports whether the locale advertises UTF-8. Block art (half
// and shade blocks) assumes UTF-8; without it a decorated surface renders
// its text lines only. Checked in POSIX order: LC_ALL, LC_CTYPE, LANG.
func UTF8Locale(getenv func(string) string) bool {
	for _, k := range []string{"LC_ALL", "LC_CTYPE", "LANG"} {
		if v := getenv(k); v != "" {
			v = strings.ToLower(v)
			return strings.Contains(v, "utf-8") || strings.Contains(v, "utf8")
		}
	}
	// No locale variables at all: modern terminals default to UTF-8, but a
	// bare environment is exactly where art least belongs. Stay honest.
	return false
}
