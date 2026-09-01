package ahoy

import "strings"

// scrubRemoteUserinfo drops any credential from a git remote URL before it
// enters RepoIdentity — the one value every registry sink and every JSON
// surface reads (GHSA-qc3w-8pv5-crc3). An operator who configured
// `https://user:token@github.com/owner/repo.git` as origin should not find the
// token at rest in ~/.abcd/history or on stdout, and scrubbing at the single
// derivation site means no downstream writer or renderer has to remember to.
//
// The rule distinguishes a credential from a route:
//   - scheme form (`scheme://[userinfo@]host/path`): a userinfo carrying a
//     password is dropped whole under any scheme; under http(s) a bare user is
//     dropped too, because the forge accepts a token as the user with no
//     password; under any other scheme (ssh, git+ssh) a bare user is the login
//     name the transport needs, so it stays;
//   - scp-like form (`[user@]host:path`): a user segment carrying a password is
//     dropped; `git@host:path` is an SSH login and stays.
//
// Everything else — a local path, an empty value — is returned unchanged.
func scrubRemoteUserinfo(s string) string {
	if i := strings.Index(s, "://"); i >= 0 {
		scheme := strings.ToLower(s[:i])
		rest := s[i+3:]
		authority := rest
		if slash := strings.IndexByte(rest, '/'); slash >= 0 {
			authority = rest[:slash]
		}
		at := strings.LastIndexByte(authority, '@')
		if at < 0 {
			return s
		}
		userinfo := authority[:at]
		hasPassword := strings.Contains(userinfo, ":")
		if !hasPassword && scheme != "http" && scheme != "https" {
			return s
		}
		return s[:i+3] + authority[at+1:] + rest[len(authority):]
	}
	at := strings.IndexByte(s, '@')
	if at < 0 || !strings.Contains(s[:at], ":") {
		return s
	}
	return s[at+1:]
}
