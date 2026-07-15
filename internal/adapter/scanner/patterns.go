package scanner

import (
	"regexp"
	"strings"
)

// Pattern is one compiled secret regex plus its metadata. Skip is the RE2
// lookaround replacement: when non-nil it is called with the full match and,
// when it returns true, the match is discarded (mirrors the negative lookaheads
// the Python patterns used, which RE2 cannot express).
type Pattern struct {
	Name       string
	Kind       string
	Label      string
	Re         *regexp.Regexp
	Severity   Severity
	Skip       func(match string) bool // nil == no skip
	Suggestion string
}

// awsExample is the canonical AWS-docs example key, widely used as a test
// placeholder; it must never be flagged (ported neg-lookahead).
const awsExample = "AKIAIOSFODNN7EXAMPLE"

// rpRedactedPlaceholder is the sanitised RepoPrompt sessionKey value; a match
// carrying it is already redacted and must not be re-flagged.
const rpRedactedPlaceholder = "<RP-SESSION-UUID-REDACTED>"

// DefaultPatterns returns the bundled secret pattern set (spec §2.2), ported
// verbatim from scripts/abcd/defaults/pii.json. Every secret pattern is
// hard_fail and non-sanitisable. This set is the built-in baseline the merged
// config layers on top of (the Go analogue of the bundled defaults/pii.json).
func DefaultPatterns() []Pattern {
	p := []Pattern{
		{
			Name: "rp_session_key", Kind: "rp_session_key",
			Label: "RepoPrompt sessionKey UUID",
			// Match the whole key/value; the negative lookahead on the value is
			// ported as a Skip that discards the already-redacted placeholder.
			Re:         regexp.MustCompile(`"sessionKey"\s*:\s*"([^"]+)"`),
			Severity:   SeverityHardFail,
			Skip:       func(m string) bool { return strings.Contains(m, rpRedactedPlaceholder) },
			Suggestion: "RP local-workspace session token — remove or redact to placeholder",
		},
		{
			Name: "github_pat", Kind: "token:github_pat", Label: "GitHub PAT (ghp_)",
			Re: regexp.MustCompile(`\bghp_[A-Za-z0-9]{36,}\b`), Severity: SeverityHardFail,
			Suggestion: "DELETE AND ROTATE — never commit credentials",
		},
		{
			Name: "github_server_token", Kind: "token:github_server", Label: "GitHub server token (ghs_)",
			Re: regexp.MustCompile(`\bghs_[A-Za-z0-9]{36,}\b`), Severity: SeverityHardFail,
			Suggestion: "DELETE AND ROTATE",
		},
		{
			Name: "github_oauth", Kind: "token:github_oauth", Label: "GitHub OAuth (gho_)",
			Re: regexp.MustCompile(`\bgho_[A-Za-z0-9]{36,}\b`), Severity: SeverityHardFail,
			Suggestion: "DELETE AND ROTATE",
		},
		{
			Name: "github_user_token", Kind: "token:github_user", Label: "GitHub user token (ghu_)",
			Re: regexp.MustCompile(`\bghu_[A-Za-z0-9]{36,}\b`), Severity: SeverityHardFail,
			Suggestion: "DELETE AND ROTATE",
		},
		{
			Name: "github_refresh", Kind: "token:github_refresh", Label: "GitHub refresh token (ghr_)",
			Re: regexp.MustCompile(`\bghr_[A-Za-z0-9]{36,}\b`), Severity: SeverityHardFail,
			Suggestion: "DELETE AND ROTATE",
		},
		{
			// Fine-grained PAT, GitHub's default token type since 2022:
			// github_pat_ + 22 alnum + '_' + 59 alnum. The classic ghp_ pattern
			// above cannot match this prefix, so it needs its own entry.
			Name: "github_pat_finegrained", Kind: "token:github_pat_finegrained",
			Label:      "GitHub fine-grained PAT (github_pat_)",
			Re:         regexp.MustCompile(`\bgithub_pat_[A-Za-z0-9]{22}_[A-Za-z0-9]{59}\b`),
			Severity:   SeverityHardFail,
			Suggestion: "DELETE AND ROTATE — never commit credentials",
		},
		{
			// PEM private-key header. The scanner is line-oriented and the BEGIN
			// line is single-line and self-identifying, so matching the header
			// flags the block's presence (RSA/EC/DSA/OPENSSH/PGP/ENCRYPTED/plain).
			Name: "pem_private_key", Kind: "token:pem_private_key",
			Label:      "PEM private key header",
			Re:         regexp.MustCompile(`-----BEGIN (?:[A-Z0-9]+ )*PRIVATE KEY( BLOCK)?-----`),
			Severity:   SeverityHardFail,
			Suggestion: "DELETE AND ROTATE — private key material must never be committed",
		},
		{
			Name: "anthropic_key", Kind: "token:anthropic", Label: "Anthropic API key (sk-ant-)",
			Re: regexp.MustCompile(`\bsk-ant-[A-Za-z0-9_-]{40,}\b`), Severity: SeverityHardFail,
			Suggestion: "DELETE AND ROTATE",
		},
		{
			Name: "openai_project_key", Kind: "token:openai_project", Label: "OpenAI project key (sk-proj-)",
			Re: regexp.MustCompile(`\bsk-proj-[A-Za-z0-9_-]{40,}\b`), Severity: SeverityHardFail,
			Suggestion: "DELETE AND ROTATE",
		},
		{
			Name: "openai_service_account", Kind: "token:openai_svcacct", Label: "OpenAI service account key (sk-svcacct-)",
			Re: regexp.MustCompile(`\bsk-svcacct-[A-Za-z0-9_-]{40,}\b`), Severity: SeverityHardFail,
			Suggestion: "DELETE AND ROTATE",
		},
		{
			Name: "aws_access_key", Kind: "token:aws_access_key", Label: "AWS access key ID",
			// The neg-lookahead excluding the docs example is ported as a Skip.
			Re:         regexp.MustCompile(`\bAKIA[0-9A-Z]{16}\b`),
			Severity:   SeverityHardFail,
			Skip:       func(m string) bool { return m == awsExample },
			Suggestion: "DELETE AND ROTATE — also rotate corresponding secret in IAM",
		},
		{
			Name: "slack_token", Kind: "token:slack", Label: "Slack token (xox*)",
			Re: regexp.MustCompile(`\bxox[baprs]-[A-Za-z0-9-]{10,}\b`), Severity: SeverityHardFail,
			Suggestion: "DELETE AND ROTATE",
		},
		{
			Name: "google_api_key", Kind: "token:google_api", Label: "Google API key (AIza)",
			Re: regexp.MustCompile(`\bAIza[0-9A-Za-z_-]{35}\b`), Severity: SeverityHardFail,
			Suggestion: "DELETE AND ROTATE",
		},
		{
			Name: "stripe_live_key", Kind: "token:stripe_live", Label: "Stripe live key (sk_live_)",
			Re: regexp.MustCompile(`\bsk_live_[A-Za-z0-9]{20,}\b`), Severity: SeverityHardFail,
			Suggestion: "DELETE AND ROTATE — production credential",
		},
		{
			Name: "stripe_test_key", Kind: "token:stripe_test", Label: "Stripe test key (sk_test_)",
			Re: regexp.MustCompile(`\bsk_test_[A-Za-z0-9]{20,}\b`), Severity: SeverityHardFail,
			Suggestion: "Review — test keys are lower risk but still shouldn't be committed",
		},
		{
			Name: "jwt_shaped", Kind: "token:jwt_shaped", Label: "JWT-shaped token",
			Re:         regexp.MustCompile(`\beyJ[A-Za-z0-9_-]{10,}\.[A-Za-z0-9_-]{10,}\.[A-Za-z0-9_-]{10,}\b`),
			Severity:   SeverityHardFail,
			Suggestion: "Review — may be benign sample or real bearer token",
		},
	}
	return p
}

// defaultPatternFloors captures the built-in severity floor per bundled pattern
// name (used to clamp a config override that tries to downgrade one).
func defaultPatternFloors() map[string]Severity {
	floors := map[string]Severity{}
	for _, p := range DefaultPatterns() {
		floors[p.Name] = p.Severity
	}
	return floors
}
