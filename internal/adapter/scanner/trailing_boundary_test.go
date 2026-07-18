package scanner

import (
	"strings"
	"testing"
)

// TestSecretSurvivesUnderscoreSuffix is the attack-input test for the ASCII
// word-boundary redaction gap. A secret pattern whose charset is pure word
// characters ([A-Za-z0-9]) but excludes '_' cannot end at a trailing \b when the
// secret is immediately followed by '_': '_' is a word char, so <alnum>_ has no
// boundary, RE2 cannot shorten past it (every interior position is word/word),
// and the whole match fails — the secret is NOT redacted. A credential embedded
// in an identifier-like context (token=ghp_...._old, JSON key, concatenation)
// therefore survives into a committed, pushed transcript. Every hard-fail token
// must still be caught when suffixed with '_'.
func TestSecretSurvivesUnderscoreSuffix(t *testing.T) {
	r := strings.Repeat
	cases := []struct {
		kind   string
		secret string
	}{
		{"token:github_pat", "ghp_" + r("a", 36)},
		{"token:github_server", "ghs_" + r("b", 36)},
		{"token:github_oauth", "gho_" + r("c", 36)},
		{"token:github_user", "ghu_" + r("d", 36)},
		{"token:github_refresh", "ghr_" + r("e", 36)},
		{"token:github_pat_finegrained", "github_pat_" + r("a", 22) + "_" + r("b", 59)},
		{"token:aws_access_key", "AKIA" + r("A", 16)},
		{"token:stripe_live", "sk_live_" + r("a", 24)},
		{"token:stripe_test", "sk_test_" + r("a", 24)},
		// Slack's charset [A-Za-z0-9-] also excludes '_', so it had the same gap.
		{"token:slack", "xoxb-" + r("a", 16)},
	}
	for _, c := range cases {
		// Bare (control): must be detected.
		if !hasKind(scanLine(c.secret), c.kind) {
			t.Errorf("%s: bare secret not detected (%q)", c.kind, c.secret)
		}
		// Underscore suffix: the secret is still present and must still be caught.
		suffixed := c.secret + "_old"
		if !hasKind(scanLine(suffixed), c.kind) {
			t.Errorf("%s: secret survived an underscore suffix unredacted (%q)", c.kind, suffixed)
		}
		// A word char that IS in the class (alnum) simply extends the token — also
		// caught (redacting a superset is safe).
		if !hasKind(scanLine(c.secret+"AAAA"), c.kind) {
			t.Errorf("%s: secret not detected with an alnum suffix", c.kind)
		}
	}
}

// TestDashEndingSecretStillCaught covers the other trailing-\b failure mode (the
// original google_api_key axis, extended): a mixed-class token ([A-Za-z0-9_-])
// whose match ends in '-' at the minimum length, followed by a non-word char.
// The last matched char '-' is non-word and the next char is non-word, so a
// trailing \b (between two non-word chars) can never hold and RE2 cannot shorten
// below the minimum — the secret is missed. Dropping the trailing \b closes it.
func TestDashEndingSecretStillCaught(t *testing.T) {
	r := strings.Repeat
	cases := []struct {
		kind   string
		secret string
	}{
		{"token:anthropic", "sk-ant-" + r("a", 39) + "-"},   // 40 body chars ending in '-'
		{"token:openai_project", "sk-proj-" + r("a", 39) + "-"},
		{"token:openai_svcacct", "sk-svcacct-" + r("a", 39) + "-"},
	}
	for _, c := range cases {
		if !hasKind(scanLine(c.secret+" trailing"), c.kind) {
			t.Errorf("%s: a minimum-length secret ending in '-' was missed before a non-word char (%q)", c.kind, c.secret)
		}
	}
}
