package scanner

import (
	"net/netip"
	"regexp"
	"strings"
)

// Network-identifier detection, built as an ALLOWLIST INVERSION.
//
// Every legitimate illustrative identifier in a committed file comes from a
// reserved documentation range (the examples-use-reserved-identifiers
// principle): RFC 5737 for IPv4, RFC 3849 for IPv6, RFC 2606/6761 for names,
// RFC 7042 for MACs, and the persona registry for device names. That convention
// is what makes detection cheap: instead of trying to recognise a real address —
// impossible without knowing the network — the detector recognises the small,
// closed set of values that are ALLOWED and flags everything else. Private, LAN
// and CGNAT/tailnet addresses are therefore findings, not exceptions: they are
// exactly the class the 2026-07-29 field incident leaked.
//
// This set is the ONE canonical primitive. It is folded into DefaultPatterns, so
// every scanner consumer inherits it (launch dry-run, lifeboat pack, history
// Stage-1 redaction), and the audit privacy-hygiene rule consults it directly
// rather than keeping a second copy.

// Built-in network kinds.
const (
	kindNetIPv4       = "net:ipv4"
	kindNetIPv6       = "net:ipv6"
	kindNetMAC        = "net:mac"
	kindNetLANHost    = "net:lan_hostname"
	kindNetDeviceHost = "net:device_hostname"
)

var (
	// RFC 5737 IPv4 documentation blocks — the only addresses an example may use.
	docIPv4 = []netip.Prefix{
		netip.MustParsePrefix("192.0.2.0/24"),
		netip.MustParsePrefix("198.51.100.0/24"),
		netip.MustParsePrefix("203.0.113.0/24"),
	}
	// RFC 3849 IPv6 documentation prefix.
	docIPv6 = netip.MustParsePrefix("2001:db8::/32")

	// A dotted quad. Octet-range validation is done by netip in the skip, not by
	// the regex: a regex that spells out 0-255 is unreadable and no more correct.
	ipv4Re = regexp.MustCompile(`\b\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}\b`)

	// An IPv6 candidate: a hextet followed by at least two more colon-separated
	// groups, so a "::" scope resolution (std::string) and a clock time (12:34:56)
	// cannot reach the skip as addresses. The leading \b is load-bearing — without
	// it the tail of an identifier ("std" -> "d::") parses as a valid address.
	// Correctness is netip's job; the regex only bounds the candidate.
	ipv6Re = regexp.MustCompile(`\b[0-9A-Fa-f]{1,4}(?::[0-9A-Fa-f]{0,4}){2,7}`)

	// A MAC address in either separator style.
	macRe = regexp.MustCompile(`\b[0-9A-Fa-f]{2}(?::[0-9A-Fa-f]{2}){5}\b|\b[0-9A-Fa-f]{2}(?:-[0-9A-Fa-f]{2}){5}\b`)

	// A hostname under a LAN/private suffix: mDNS (.local), the plain .lan
	// convention, and the home-router default (.fritz.box, whose leading label is
	// optional because the bare name is itself the host). At least one label must
	// precede .local/.lan, so a bare "~/.local" directory cannot match.
	lanHostRe = regexp.MustCompile(`(?i)\b(?:(?:[a-z0-9](?:[a-z0-9-]*[a-z0-9])?\.)+(?:local|lan)|(?:[a-z0-9](?:[a-z0-9-]*[a-z0-9])?\.)*fritz\.box)\b`)

	// A device hostname: <name>-<device noun>. The noun set is deliberately narrow
	// — it excludes "server", "host", "machine", "router" and "box", which are
	// overwhelmingly software-role words in a source tree ("mcp-server",
	// "prompt-router", "state-machine") and would drown the signal. A
	// persona-derived name (alice-laptop, carol-server) is allowed by the skip
	// below whether or not its noun is in this set.
	deviceHostRe = regexp.MustCompile(`(?i)\b[a-z0-9][a-z0-9-]*-(?:laptop|macbook|mbp|imac|thinkpad|desktop|workstation|netbook|chromebook|nas)\b`)
)

// personaNames is the persona registry's given-name sequence
// (.abcd/development/personas.json). A device name derived from a persona is a
// fixture host by construction, so it is allowed. The list is embedded rather
// than read at run time because the registry lives under .abcd/development/,
// which is excluded from the released artifact.
var personaNames = map[string]bool{
	"alice": true, "bob": true, "carol": true, "dave": true, "eve": true,
	"frank": true, "grace": true, "henry": true, "iris": true, "jack": true,
	"kira": true, "liam": true, "maya": true, "nia": true,
}

// nonHostLabels are names that carry a LAN suffix but name no host. abcd's own
// local tier directory (.abcd/.work.local) appears as the bare string
// "work.local" in product code and its tests, so the canonical set names it once
// here rather than making every occurrence carry a waiver — the iss-153 lesson
// (a detector must not tax legitimate product code) applied to hostnames.
var nonHostLabels = map[string]bool{"work.local": true}

// nonUserHomeSegments are the well-known macOS directories that live under
// /Users but name no user. Flagging them as usernames (iss-153) forced waivers
// onto product code that legitimately writes to /Users/Shared.
var nonUserHomeSegments = map[string]bool{
	"shared": true, "guest": true, "public": true,
}

// IsNonUserHomeSegment reports whether seg, the segment immediately after a
// /Users home root, names a system directory rather than a user. The comparison
// is case-insensitive because a case-folding filesystem resolves /users/shared
// to the same directory.
func IsNonUserHomeSegment(seg string) bool {
	return nonUserHomeSegments[strings.ToLower(seg)]
}

// NetworkPatterns returns the canonical network-identifier pattern set. It is
// exported so the audit privacy-hygiene rule can run exactly these patterns over
// committed files without duplicating them, and it is included in
// DefaultPatterns so every scanner consumer gets the same detection.
//
// Severity: addresses (IPv4/IPv6/MAC) are hard_fail — the reserved ranges make a
// non-reserved address an authoring error with near-zero ambiguity. The two
// hostname patterns are shape heuristics rather than range checks, so they are
// warn: Stage-1 redaction rewrites them either way (Redact acts on every
// finding), and a repo that wants them blocking can raise the severity through
// .abcd/config/pii.json, which may raise but never lower a floor.
func NetworkPatterns() []Pattern {
	return []Pattern{
		{
			Name: "net_ipv4", Kind: kindNetIPv4, Label: "non-reserved IPv4 address",
			Re: ipv4Re, Severity: SeverityHardFail,
			Skip: func(m string) bool { return allowedIPv4(m) },
			SkipAt: func(line string, start, end int) bool {
				return insideLongerDottedRun(line, start, end) || cidrPrefixDeclaration(line, start, end)
			},
			Suggestion: "replace with an RFC 5737 documentation address (192.0.2.x, 198.51.100.x, 203.0.113.x)",
		},
		{
			Name: "net_ipv6", Kind: kindNetIPv6, Label: "non-reserved IPv6 address",
			Re: ipv6Re, Severity: SeverityHardFail,
			Skip:       func(m string) bool { return allowedIPv6(m) },
			SkipAt:     cidrPrefixDeclaration,
			Suggestion: "replace with an RFC 3849 documentation address (2001:db8::/32)",
		},
		{
			Name: "net_mac", Kind: kindNetMAC, Label: "non-reserved MAC address",
			Re: macRe, Severity: SeverityHardFail,
			Skip:       func(m string) bool { return allowedMAC(m) },
			Suggestion: "replace with an RFC 7042 documentation MAC (00:00:5E:00:53:00-FF)",
		},
		{
			Name: "net_lan_hostname", Kind: kindNetLANHost, Label: "LAN hostname",
			Re: lanHostRe, Severity: SeverityWarn,
			Skip:       func(m string) bool { return personaDerivedHost(m) },
			SkipAt:     dottedFileOrDirectory,
			Suggestion: "replace with a reserved name (example.com, host.test) or a persona-derived fixture host",
		},
		{
			Name: "net_device_hostname", Kind: kindNetDeviceHost, Label: "device hostname",
			Re: deviceHostRe, Severity: SeverityWarn,
			Skip:       func(m string) bool { return personaDerivedHost(m) },
			Suggestion: "replace with a persona-derived fixture host (alice-laptop, bob-desktop)",
		},
	}
}

// allowedIPv4 reports whether a dotted quad may appear in committed content: a
// reserved documentation address, loopback, the unspecified address, or a
// netmask. A candidate netip cannot parse is not an address at all (a version
// string, a padded octet), so it is allowed rather than guessed at.
func allowedIPv4(m string) bool {
	addr, err := netip.ParseAddr(m)
	if err != nil || !addr.Is4() {
		return true
	}
	if addr.IsLoopback() || addr.IsUnspecified() {
		return true
	}
	for _, p := range docIPv4 {
		if p.Contains(addr) {
			return true
		}
	}
	return isNetmask4(addr)
}

// isNetmask4 reports whether an address is a contiguous subnet mask
// (255.255.255.0, 255.255.0.0, 255.255.255.255). A mask is a shape, not a host:
// it identifies nobody, and it is the commonest dotted quad in network prose.
func isNetmask4(addr netip.Addr) bool {
	b := addr.As4()
	if b[0] != 255 {
		return false
	}
	v := uint32(b[0])<<24 | uint32(b[1])<<16 | uint32(b[2])<<8 | uint32(b[3])
	inv := ^v
	return inv&(inv+1) == 0 // ones followed by zeros, no gaps
}

// allowedIPv6 mirrors allowedIPv4 for IPv6: the RFC 3849 documentation prefix,
// loopback (::1) and the unspecified address (::) are allowed. A 4-in-6 mapped
// address is judged by its IPv4 half so ::ffff:192.0.2.1 stays allowed while
// ::ffff:<private> does not.
func allowedIPv6(m string) bool {
	addr, err := netip.ParseAddr(m)
	if err != nil || addr.Is4() {
		return true // not an address, or a quad the IPv4 pattern already judged
	}
	if addr.Is4In6() {
		return allowedIPv4(addr.Unmap().String())
	}
	if addr.IsLoopback() || addr.IsUnspecified() {
		return true
	}
	return docIPv6.Contains(addr)
}

// allowedMAC reports whether a MAC sits in the RFC 7042 documentation range
// 00:00:5E:00:53:00-FF.
func allowedMAC(m string) bool {
	n := strings.ToLower(strings.ReplaceAll(m, "-", ":"))
	return strings.HasPrefix(n, "00:00:5e:00:53:")
}

// cidrPrefixDeclaration suppresses a match that is the base address of a CIDR
// prefix written immediately after it — "fc00::/7", "100.64.0.0/10",
// "64:ff9b::/96". A prefix names a RANGE, not a host: it identifies nobody, and
// range notation is how the reserved and special-use blocks are documented in
// the first place, including by this detector's own design record. The
// suppression is deliberately narrow: the address must be the MASKED BASE of the
// stated length, so "10.0.0.1/24" (and any URL path that merely begins with a
// digit) is still a finding, and a single-host prefix (/32, /128) is never
// exempt.
func cidrPrefixDeclaration(line string, start, end int) bool {
	if end >= len(line) || line[end] != '/' {
		return false
	}
	i := end + 1
	for i < len(line) && isASCIIDigit(line[i]) {
		i++
	}
	if i == end+1 || i-(end+1) > 3 {
		return false
	}
	p, err := netip.ParsePrefix(line[start:i])
	if err != nil {
		return false
	}
	if p.Bits() == p.Addr().BitLen() {
		return false // /32 or /128 names one host, not a range
	}
	return p.Masked() == p
}

// personaDerivedHost reports whether a host or device name is built from the
// persona registry — alice-laptop, bob-desktop, maya-workstation.lan. The first
// hyphen-separated token of the first label is the persona.
func personaDerivedHost(m string) bool {
	if nonHostLabels[strings.ToLower(m)] {
		return true
	}
	label := m
	if i := strings.IndexByte(label, '.'); i >= 0 {
		label = label[:i]
	}
	if i := strings.IndexByte(label, '-'); i >= 0 {
		label = label[:i]
	}
	return personaNames[strings.ToLower(label)]
}

// insideLongerDottedRun suppresses a quad that is part of a longer dotted number
// run (a four-part version, "1.2.3.4.5"): the leading \b already excludes a
// digit neighbour, so only a '.' on either side can extend the run.
func insideLongerDottedRun(line string, start, end int) bool {
	if start > 0 && line[start-1] == '.' {
		return true
	}
	return end+1 < len(line) && line[end] == '.' && isASCIIDigit(line[end+1])
}

// dottedFileOrDirectory suppresses a LAN-suffix match that is really a filename
// or dot-directory rather than a host: ".work.local/" (the local tier),
// "settings.local.json", "DECISIONS.local.md". A leading '.' means the label
// chain began as a dotfile; a trailing '.' followed by a label means a further
// extension. Sentence-final punctuation ("ping printer.local.") is not
// suppressed, because the '.' there is not followed by a label character.
func dottedFileOrDirectory(line string, start, end int) bool {
	if start > 0 && line[start-1] == '.' {
		return true
	}
	return end+1 < len(line) && line[end] == '.' && isHostLabelByte(line[end+1])
}

func isASCIIDigit(b byte) bool { return b >= '0' && b <= '9' }

func isHostLabelByte(b byte) bool {
	return isASCIIDigit(b) || (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z')
}
