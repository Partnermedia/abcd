package scanner

import (
	"os"
	"os/exec"
	"regexp"
	"strings"
)

// Identity is the caller's runtime identity, probed from git config and the
// environment. Its matchers are built at scan time; empty fields disable the
// corresponding kind.
type Identity struct {
	GitUserName       string
	GitUserEmail      string
	GitRemoteUsername string
	HomePath          string
	HomeUser          string
}

// Built-in identity kinds.
const (
	kindHomeSelf   = "home_path_self"
	kindHomeOther  = "home_path_other"
	kindRealEmail  = "real_email"
	kindRealName   = "real_name"
	kindGithubUser = "github_username"
	kindLocalUser  = "local_username"
)

// DefaultIdentitySeverities is the built-in severity floor per identity kind
// (ported from pii.py DEFAULT_IDENTITY_SEVERITIES). A config override may raise
// but never lower these.
func DefaultIdentitySeverities() map[string]Severity {
	return map[string]Severity{
		kindHomeSelf:   SeverityHardFail,
		kindHomeOther:  SeverityWarn,
		kindRealEmail:  SeverityHardFail,
		kindRealName:   SeverityHardFail,
		kindGithubUser: SeverityWarn,
		kindLocalUser:  SeverityHardFail,
	}
}

// ProbeIdentity gathers the caller's identity from git config and $HOME,
// best-effort: any probe that fails leaves its field empty. repoRoot scopes the
// git config reads so a per-repo user.name/email is honoured.
func ProbeIdentity(repoRoot string) Identity {
	var id Identity
	git := func(args ...string) string {
		full := append([]string{"-C", repoRoot}, args...)
		cmd := exec.Command("git", full...)
		out, err := cmd.Output()
		if err != nil {
			return ""
		}
		return strings.TrimSpace(string(out))
	}
	id.GitUserName = git("config", "--get", "user.name")
	id.GitUserEmail = git("config", "--get", "user.email")
	if remote := git("config", "--get", "remote.origin.url"); remote != "" {
		if m := githubRemoteRe.FindStringSubmatch(remote); m != nil {
			id.GitRemoteUsername = m[1]
		}
	}
	home := os.Getenv("HOME")
	if home == "" {
		if h, err := os.UserHomeDir(); err == nil {
			home = h
		}
	}
	if home != "" {
		id.HomePath = strings.TrimRight(home, "/")
		if i := strings.LastIndex(id.HomePath, "/"); i >= 0 {
			id.HomeUser = id.HomePath[i+1:]
		}
	}
	return id
}

var (
	// GitHub username inside a remote URL (https or ssh form).
	githubRemoteRe = regexp.MustCompile(`github\.com[:/]([A-Za-z0-9-]+)/`)
	// Generic home path — \b is RE2-safe; the trailing boundary is a Go predicate.
	genericHomeRe = regexp.MustCompile(`\b(?:/Users/[A-Za-z0-9._-]+|/home/[A-Za-z0-9._-]+)`)
	// Loose URL span (scheme to whitespace/quote/closing).
	urlSpanRe = regexp.MustCompile(`(?:https?://|git@|ftp://|ssh://)[^\s"'` + "`" + `)>\]<]+`)
	// A git noreply email is not a leak.
	noreplyRe = regexp.MustCompile(`(?i)@users\.noreply\.github\.com$`)
)

// homeBoundary is the trailing-boundary set for a home-path match (ported from
// the Python lookahead [/\s"'`)\]\}<,;:]).
func homeBoundary(r rune) bool {
	switch r {
	case '/', '"', '\'', '`', ')', ']', '}', '<', ',', ';', ':':
		return true
	}
	return r == ' ' || r == '\t' || r == '\n' || r == '\r' || r == '\f' || r == '\v'
}

// identityMatchers holds the per-scan compiled identity regexes.
type identityMatchers struct {
	id           Identity
	homeSelf     *regexp.Regexp
	email        *regexp.Regexp
	name         *regexp.Regexp
	github       *regexp.Regexp
	localBare    *regexp.Regexp
	localEncoded string // path-encoded username (dots->hyphens); boundary checked in Go
	nameEqGithub bool
}

func newIdentityMatchers(id Identity) identityMatchers {
	m := identityMatchers{id: id}
	if id.HomePath != "" {
		m.homeSelf = regexp.MustCompile(regexp.QuoteMeta(id.HomePath))
	}
	if id.GitUserEmail != "" {
		m.email = regexp.MustCompile(regexp.QuoteMeta(id.GitUserEmail))
	}
	if n := strings.TrimSpace(id.GitUserName); len(n) >= 3 {
		m.name = regexp.MustCompile(`(?i)\b` + regexp.QuoteMeta(id.GitUserName) + `\b`)
	}
	if id.GitRemoteUsername != "" {
		m.github = regexp.MustCompile(`\b` + regexp.QuoteMeta(id.GitRemoteUsername) + `\b`)
	}
	if id.HomeUser != "" {
		m.localBare = regexp.MustCompile(`\b` + regexp.QuoteMeta(id.HomeUser) + `\b`)
		if enc := strings.ReplaceAll(id.HomeUser, ".", "-"); enc != id.HomeUser {
			m.localEncoded = enc
		}
	}
	m.nameEqGithub = id.GitUserName != "" && id.GitRemoteUsername != "" &&
		strings.EqualFold(id.GitUserName, id.GitRemoteUsername)
	return m
}

// span is a half-open byte interval on a line.
type span struct{ start, end int }

func inAnySpan(pos int, spans []span) bool {
	for _, s := range spans {
		if s.start <= pos && pos < s.end {
			return true
		}
	}
	return false
}

// urlSpans returns the URL-like spans on a line.
func urlSpans(line string) []span {
	var out []span
	for _, loc := range urlSpanRe.FindAllStringIndex(line, -1) {
		out = append(out, span{loc[0], loc[1]})
	}
	return out
}

// findings scans one line for identity-derived matches, applying every ported
// suppression, and returns findings tagged with the merged identity severities.
func (m identityMatchers) findings(line string, lineno int, id2sev map[string]Severity, file string) []Finding {
	var out []Finding
	sevFor := func(kind string) Severity {
		if s, ok := id2sev[kind]; ok {
			return s
		}
		return defaultPatternSeverity
	}
	add := func(kind string, col int, matched, suggested string) {
		out = append(out, Finding{
			File: file, Line: lineno, Column: col, Kind: kind,
			Severity: sevFor(kind), Snippet: snippet(line), Matched: matched,
			Suggested: suggested, line: line,
		})
	}

	urls := urlSpans(line)

	// home_path_self — accept only when the trailing rune is a boundary/EOL.
	if m.homeSelf != nil {
		for _, loc := range m.homeSelf.FindAllStringIndex(line, -1) {
			if !trailingBoundaryOK(line, loc[1]) {
				continue
			}
			add(kindHomeSelf, loc[0]+1, line[loc[0]:loc[1]], "~")
		}
	}
	// home_path_other — a generic /Users|/home path that is not the caller's own.
	for _, loc := range genericHomeRe.FindAllStringIndex(line, -1) {
		if !trailingBoundaryOK(line, loc[1]) {
			continue
		}
		matched := line[loc[0]:loc[1]]
		if m.homeSelf != nil && m.homeSelf.MatchString(matched) {
			continue
		}
		add(kindHomeOther, loc[0]+1, matched, "(remove or relativise — third-party path)")
	}
	// real_email — skip the noreply form.
	if m.email != nil {
		for _, loc := range m.email.FindAllStringIndex(line, -1) {
			matched := line[loc[0]:loc[1]]
			if noreplyRe.MatchString(matched) {
				continue
			}
			add(kindRealEmail, loc[0]+1, matched, "<github-userid>@users.noreply.github.com or remove")
		}
	}
	// real_name — suppress inside URL spans and when it equals the github username.
	if m.name != nil && !m.nameEqGithub {
		for _, loc := range m.name.FindAllStringIndex(line, -1) {
			if inAnySpan(loc[0], urls) {
				continue
			}
			add(kindRealName, loc[0]+1, line[loc[0]:loc[1]], "(remove or replace with persona)")
		}
	}
	// github_username — suppress inside URL spans.
	if m.github != nil {
		for _, loc := range m.github.FindAllStringIndex(line, -1) {
			if inAnySpan(loc[0], urls) {
				continue
			}
			add(kindGithubUser, loc[0]+1, line[loc[0]:loc[1]], "(review — may be intentional in repo URL contexts)")
		}
	}
	// local_username — suppress inside home/generic-home/email/URL spans.
	if m.localBare != nil {
		supp := m.localSuppressionSpans(line, urls)
		emit := func(loc []int) {
			if inAnySpan(loc[0], supp) {
				return
			}
			add(kindLocalUser, loc[0]+1, line[loc[0]:loc[1]],
				"(local machine username; replace with [USERNAME] or remove)")
		}
		for _, loc := range m.localBare.FindAllStringIndex(line, -1) {
			emit(loc)
		}
		if m.localEncoded != "" {
			for _, loc := range encodedMatches(line, m.localEncoded) {
				emit(loc)
			}
		}
	}
	return out
}

// localSuppressionSpans returns spans where a local-username match is not a
// standalone leak (own home path, any generic home path, the exact email, URLs).
func (m identityMatchers) localSuppressionSpans(line string, urls []span) []span {
	spans := append([]span(nil), urls...)
	if m.homeSelf != nil {
		for _, loc := range m.homeSelf.FindAllStringIndex(line, -1) {
			spans = append(spans, span{loc[0], loc[1]})
		}
	}
	for _, loc := range genericHomeRe.FindAllStringIndex(line, -1) {
		spans = append(spans, span{loc[0], loc[1]})
	}
	if m.email != nil {
		for _, loc := range m.email.FindAllStringIndex(line, -1) {
			spans = append(spans, span{loc[0], loc[1]})
		}
	}
	return spans
}

// encodedMatches finds the path-encoded username with the ported custom
// boundary: preceded by start-of-string or a non-[A-Za-z0-9.] rune (the RE2
// lookbehind replacement) and followed by EOL or a non-[A-Za-z0-9.] rune.
func encodedMatches(line, encoded string) [][]int {
	var out [][]int
	from := 0
	for {
		i := strings.Index(line[from:], encoded)
		if i < 0 {
			break
		}
		start := from + i
		end := start + len(encoded)
		if boundaryBefore(line, start) && boundaryAfter(line, end) {
			out = append(out, []int{start, end})
		}
		from = start + 1
	}
	return out
}

func isUsernameWordRune(r byte) bool {
	return r == '.' || (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9')
}

func boundaryBefore(line string, pos int) bool {
	if pos == 0 {
		return true
	}
	return !isUsernameWordRune(line[pos-1])
}

func boundaryAfter(line string, pos int) bool {
	if pos >= len(line) {
		return true
	}
	return !isUsernameWordRune(line[pos])
}

// trailingBoundaryOK reports whether the rune at byte offset end is a home-path
// boundary or the line ends there.
func trailingBoundaryOK(line string, end int) bool {
	if end >= len(line) {
		return true
	}
	return homeBoundary(rune(line[end]))
}

// snippet is the trimmed line capped at 200 bytes.
func snippet(line string) string {
	s := strings.TrimSpace(line)
	if len(s) > 200 {
		return s[:200]
	}
	return s
}
