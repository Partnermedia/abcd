package scanner

import (
	"math"
	"strings"
	"testing"
)

// The iss-96 verification corpus: what the TRANSCRIPT path's pattern set does,
// and does not, catch.
//
// Since session-end capture became automatic, every session's raw text is fed to
// `(*Scanner).ScanText` before it lands in the history store (internal/core/history
// Capture, stage one), so this set is the whole of that boundary's detection. The
// two tests below are a coverage PIN, not a behaviour change: they assert what the
// shipped patterns do today, so the tracked gap is evidence in the tree rather than
// a claim in a ledger entry.
//
// Test B is deliberately an assertion that nothing is found, and its reach is
// exactly the pattern set:
//
//   - it ALARMS for growth in the set this scanner runs — a charset/length
//     floor, a key-name context matcher, an entropy floor (the runtime-assembled
//     high-entropy specimen below exists so that a real entropy threshold cannot
//     land while every case here stays green), or any new DefaultPatterns entry
//     that reaches these shapes. That failure is the intended signal to re-point
//     the pin, not a regression;
//   - it CANNOT alarm for an opt-in external-scanner adapter run over this path.
//     Such an adapter never consults DefaultPatterns, so ScanText keeps returning
//     nothing and this pin stays green while coverage has in fact grown. Whoever
//     wires that route has to re-point the pin by hand; the pin alone is not
//     grounds to close iss-96.
//
// Every specimen is synthetic and self-evidently fake: no value here is, or has
// ever been, a credential. Identifiers are persona/reserved values only, and any
// value whose SHAPE is credential-like is assembled at run time (see phrase and
// entropySpecimen, and network_test.go's header for the discipline it follows).
//
// CROSS-REFERENCE: internal/core/history/history_test.go's
// TestCaptureStoresUnanchoredEntropyVerbatim carries the same specimen shapes at
// the STORE boundary (a separate package, so the set is duplicated rather than
// shared). The two sets are meant to move together — change one and change the
// other, or the ledger entry citing both stops describing the same residue. That
// intent is enforced rather than asked for: both sides assert the generated
// specimen equals the same golden value (entropySpecimenGolden here,
// storeEntropySpecimenGolden there), so a drifting copy fails on both.

// phrase assembles a nonsense passphrase from ordinary words. It exists so the
// joined value never appears as a literal in this repo: a credential-shaped run
// of characters beside a `password` keyword is what a history-wide secret scan
// fires on, and this repo's history is scanned on every push. The same reasoning
// makes network_test.go assemble every FLAGGED network specimen.
func phrase(words ...string) string { return strings.Join(words, "-") }

// specimenAlphabet is the alnum charset a prefix-less token is drawn from.
const specimenAlphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

// specimenSeed fixes the generator below. Changing it changes the specimen, so
// it is a literal here and nowhere else.
const specimenSeed = 0x5EED17

// lcg is a hand-rolled linear congruential generator (the Knuth/MMIX constants).
// It exists so the high-entropy specimen is DETERMINISTIC across runs, machines
// and Go releases without a credential-shaped literal ever entering this repo's
// tree or history: the value is computed at run time from a fixed seed, and the
// repo's own history scan sees only the seed and the alphabet. math/rand's
// top-level source is not used because its stream is a documented
// implementation detail; this one is fixed by the code beside it.
type lcg uint64

func (s *lcg) next() uint64 {
	*s = lcg(uint64(*s)*6364136223846793005 + 1442695040888963407)
	return uint64(*s)
}

// entropySpecimen assembles a 40-character alphanumeric value with genuinely
// high Shannon entropy — a deterministic shuffle of the alnum alphabet, taken
// without replacement, so all 40 characters are distinct and the per-character
// entropy is log2(40) ≈ 5.32 bits.
//
// It is the specimen the OTHER unanchored cases cannot be: FAKEfake00 repeated,
// deadbeef repeated and Zm9vYmFy repeated measure 3.12, 2.16 and 2.75 bits per
// character, all BELOW the ~3.5-bit floor an entropy detector conventionally
// uses. Without this case a real entropy threshold could land on the transcript
// path while every assertion in this file still passed — a silent pass, which is
// precisely what a pin exists to prevent.
func entropySpecimen() string {
	a := []byte(specimenAlphabet)
	s := lcg(specimenSeed)
	for i := len(a) - 1; i > 0; i-- {
		j := int(s.next() % uint64(i+1))
		a[i], a[j] = a[j], a[i]
	}
	return string(a[:40])
}

// entropySpecimenGolden is the exact value entropySpecimen returns. Two separate
// drifts are pinned by asserting it.
//
//   - DETERMINISM. A generator whose stream changed — a different seed, a
//     reordered shuffle, a Go release that altered integer behaviour — would very
//     likely still measure high entropy, so the entropy floor alone cannot see it.
//   - The CROSS-PACKAGE COPY. internal/core/history's storeEntropySpecimen is a
//     duplicate of this generator in another package (Go tests do not share
//     helpers across packages), and it asserts this same value. Each copy is
//     pinned to the same literal, so drift in either fails that side loudly and
//     its message names the other. The residual hole is a coordinated edit of
//     one side's seed and golden together: update both packages in one change.
//
// It is the one credential-SHAPED literal this file permits, and it is admitted
// by measurement rather than by assumption: a bare 40-character alphanumeric run
// with no key name and no `=`/`:` delimiter beside it does not trip gitleaks'
// generic-api-key rule, whose condition is keyword + delimiter + entropy >= 3.5
// (verified against the pinned 8.24.3 default rules, the same version this
// repo's history scan runs, over this exact declaration). That clean scan hangs
// on the keyword-free name: renaming this constant to embed a keyword (a
// ...Key/...Token/...Secret suffix) or moving the value beside a delimited key
// name would trip the rule and red the CI secret scan. It is a shuffle of the
// alphabet three lines above; it is not, and never was, a credential.
const entropySpecimenGolden = "fXbLtohSxAn9d1jv75E0eDkB6JHT8OzNWcVCQgsU"

// The low-entropy specimens, declared ONCE. TestTranscriptPathMissesUnanchoredEntropy
// scans these expressions and TestEntropySpecimenIsGenuinelyHighEntropy measures
// them, so the guard cannot drift away from the thing it guards: re-declaring the
// comparators inline in the self-check would let a corpus edit change what is
// scanned while the measurement went on describing the old value.
func lowEntropySecretKeyShape() string { return strings.Repeat("FAKEfake00", 4) }
func lowEntropyHexToken() string       { return strings.Repeat("deadbeef", 4) }
func lowEntropyBase64Token() string    { return strings.Repeat("Zm9vYmFy", 5) }

// shannonEntropy returns the per-character Shannon entropy of s in bits, the
// measure an entropy-based detector thresholds on.
func shannonEntropy(s string) float64 {
	if s == "" {
		return 0
	}
	counts := map[rune]int{}
	n := 0
	for _, r := range s {
		counts[r]++
		n++
	}
	h := 0.0
	for _, c := range counts {
		p := float64(c) / float64(n)
		h -= p * math.Log2(p)
	}
	return h
}

// TestEntropySpecimenIsGenuinelyHighEntropy guards the guard. entropySpecimen is
// only useful as a pin if it stays high-entropy: a future edit to the seed, the
// alphabet or the shuffle that quietly produced a repetitive value would restore
// the silent-pass gap it was added to close, and every other assertion would go
// on passing. The floor asserted here sits above any threshold a detector would
// plausibly use.
func TestEntropySpecimenIsGenuinelyHighEntropy(t *testing.T) {
	spec := entropySpecimen()
	if len(spec) != 40 {
		t.Fatalf("specimen must be 40 characters, got %d", len(spec))
	}
	// Never prints the specimen: a diagnostic in this file names the case, not
	// the value. See entropySpecimenGolden for what this catches that the
	// entropy floor below cannot.
	if spec != entropySpecimenGolden {
		t.Errorf("specimen no longer equals entropySpecimenGolden — the generator's stream has changed; internal/core/history pins the same value, so update BOTH copies or the two pins stop describing the same residue")
	}
	const floor = 4.5
	if h := shannonEntropy(spec); h < floor {
		t.Errorf("specimen entropy %.2f bits/char is below the %.1f floor — it can no longer alarm an entropy detector", h, floor)
	}
	// The low-entropy specimens are the contrast that makes the point: they are
	// the reason this case exists, so their measurement is asserted too. These are
	// the SAME expressions the corpus scans, not inline copies of them.
	for _, c := range []struct {
		name string
		text string
		max  float64
	}{
		{"repeated secret-key shape", lowEntropySecretKeyShape(), 3.5},
		{"repeated hex token", lowEntropyHexToken(), 3.5},
		{"repeated base64 token", lowEntropyBase64Token(), 3.5},
	} {
		if h := shannonEntropy(c.text); h >= c.max {
			t.Errorf("%s: entropy %.2f bits/char is no longer below %.1f — the high-entropy specimen is no longer the only case an entropy floor would catch", c.name, h, c.max)
		}
	}
}

// newTranscriptScanner builds the scanner the transcript capture path uses. New
// is given an EMPTY root, so the pattern set is the bundled default set with
// nothing merged on top of it — which is exactly what history.Capture gets in a
// repo carrying no .abcd/config/pii.json. (A repo that ships one gets a strictly
// wider set; this corpus pins the baseline, so an opt-in per-repo detector
// landing later cannot silently invalidate the comment.)
//
// The probed identity is replaced by a fixed persona so the corpus cannot depend
// on the machine running it: New probes the real $HOME and git config, and a bare
// `go test` on one workstation would otherwise see a different identity set from
// CI — neither of which is what this corpus is about.
//
// One scanner serves a whole corpus: the identity override is the only per-case
// state, and it is the same for every case.
func newTranscriptScanner(t *testing.T) *Scanner {
	t.Helper()
	sc, err := New(t.TempDir())
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if unavail, reason := sc.Unavailable(); unavail {
		t.Fatalf("scanner unavailable: %s", reason)
	}
	sc.identity = Identity{HomePath: "/home/alice", HomeUser: "alice"} // abcd-audit:allow
	return sc
}

// TestTranscriptPathCatchesAnchoredTokens pins the covered half: a token with a
// recognisable prefix, an absolute home path, and — since the network-identifier
// set was folded into DefaultPatterns — every one of the five network kinds. The
// network cases are the evidence that the transcript path inherits that set
// rather than a separate copy of the secret patterns.
func TestTranscriptPathCatchesAnchoredTokens(t *testing.T) {
	r := strings.Repeat
	sc := newTranscriptScanner(t)
	cases := []struct {
		name     string
		specimen string
		kind     string
	}{
		{"aws access key id (AKIA prefix)", "AWS_ACCESS_KEY_ID=AKIA" + r("F", 16), "token:aws_access_key"},
		{"github pat (ghp_ prefix)", "token=ghp_" + r("F", 36), "token:github_pat"},
		{"anthropic key (sk-ant- prefix)", "key: sk-ant-" + r("F", 40), "token:anthropic"},
		// Assembled from two halves for the same reason every other
		// credential-shaped value in this file is: the joined header is what a
		// history-wide secret scan keys on, and this corpus obeys the discipline
		// it pins. (The header is a literal in the PATTERN, which is where it
		// belongs; it need not be one here.)
		{"pem private key header", "-----BEGIN " + "PRIVATE KEY-----", "token:pem_private_key"},
		{"absolute home path of another user", "wrote /home/bob/notes.md", "home_path_other"}, // abcd-audit:allow
		{"absolute home path of the caller", "cd /home/alice/repo", "home_path_self"},         // abcd-audit:allow
		{"non-reserved IPv4 address", "ssh " + v4(10, 0, 0, 7), kindNetIPv4},
		{"non-reserved IPv6 address", "peer " + v6("fd00", "", "1"), kindNetIPv6},
		{"non-reserved MAC address", "hw addr " + mac(0x02, 0x1a, 0x2b, 0x3c, 0x4d, 0x5e), kindNetMAC},
		{"LAN hostname", "ping " + host("printer", "lan"), kindNetLANHost},
		{"device hostname", "backup from " + dash("zeta", "laptop"), kindNetDeviceHost},
	}
	for _, c := range cases {
		found := sc.ScanText(c.specimen, "transcript")
		if !hasKind(found, c.kind) {
			t.Errorf("%s: expected kind %q, got %v", c.name, c.kind, kindsOf(found))
			continue
		}
		t.Logf("CAUGHT   %-40s -> %s", c.name, c.kind)
	}
}

// TestTranscriptPathMissesUnanchoredEntropy pins the residue iss-96 tracks: a
// high-entropy value with NO recognisable prefix passes through the transcript
// path verbatim. Keyword context ("password:", "api_key =") does not help — the
// TOKEN patterns are prefix-anchored apart from `rp_session_key`, which keys on
// one literal JSON field name (`"sessionKey"`) rather than on a generic key-name
// class, and nothing in the set measures entropy. (The set as a whole is not
// uniformly prefix-anchored: the network kinds are an allowlist inversion, the
// identity kinds key on the probed identity, `pem_private_key` is a literal
// header, and `Pattern.SkipAt` already lets a pattern decide by what SURROUNDS a
// match. Those mechanisms simply do not reach an unlabelled token.)
//
// The specimens are the classes named in the ledger entry plus one genuinely
// high-entropy value, each written so no reader could mistake it for a real
// credential. The first case is an ANCHORED positive control: without it, an
// empty or mis-wired pattern set would satisfy every remaining assertion
// vacuously.
func TestTranscriptPathMissesUnanchoredEntropy(t *testing.T) {
	r := strings.Repeat
	sc := newTranscriptScanner(t)
	cases := []struct {
		name     string
		specimen string
		// control, when non-empty, marks a case that MUST be found, with the
		// kind it must carry. It guards the negative cases below from passing
		// because nothing is being scanned at all.
		control string
	}{
		{"positive control: anchored github pat", "token=ghp_" + r("F", 36), "token:github_pat"},
		// An AWS SECRET access key is 40 base64-ish characters with no prefix —
		// the counterpart of the AKIA id that IS caught above.
		{"bare 40-char secret-key shape", "AWS_SECRET_ACCESS_KEY=" + lowEntropySecretKeyShape(), ""},
		{"bare 40-char secret-key shape, no key name", lowEntropySecretKeyShape(), ""},
		{"bare passphrase", "password: " + phrase("correct", "horse", "battery", "staple", "9x"), ""},
		{"bare passphrase in a flag", "--password=" + phrase("correct", "horse", "battery", "staple", "9x"), ""},
		{"generic prefix-less hex token", "api_key = " + lowEntropyHexToken(), ""},
		{"generic prefix-less base64 token", "Authorization: Bearer " + lowEntropyBase64Token(), ""},
		// The specimens above are all REPETITIVE, and so sit below a
		// conventional entropy floor: they would not alarm this pin if an
		// entropy detector landed. This one would — see entropySpecimen.
		{"high-entropy 40-char token, no key name", entropySpecimen(), ""},
		{"high-entropy 40-char token after a key name", "api_key = " + entropySpecimen(), ""},
	}
	for _, c := range cases {
		found := sc.ScanText(c.specimen, "transcript")
		if c.control != "" {
			if !hasKind(found, c.control) {
				t.Errorf("%s: positive control not caught — expected kind %q, got %v; the negative cases below prove nothing until this passes", c.name, c.control, kindsOf(found))
				continue
			}
			t.Logf("CONTROL CAUGHT  %s -> %s", c.name, c.control)
			continue
		}
		if len(found) != 0 {
			// Both branches fail — the pin must stay loud — but they mean
			// different things, and a message that conflates them would credit
			// the wrong change. Only a TOKEN/SECRET kind is evidence that the
			// gap iss-96 tracks has closed; a widened network or identity
			// pattern incidentally matching a synthetic specimen is a fact about
			// the SPECIMEN, not about secret coverage.
			if secretKindsOf(found) != nil {
				// Not a failure of the product — a failure of this PIN's premise. If
				// detection has grown to cover the class, re-point the corpus (and the
				// ledger entry that cites it) rather than deleting the case.
				t.Errorf("%s: expected no finding (tracked gap), got secret kinds %v — token/secret coverage has changed; re-point this pin", c.name, secretKindsOf(found))
			} else {
				// This branch has TWO readings and must not steer to the wrong
				// one. A widened network or identity pattern is a fact about the
				// specimen. But a secret or entropy detector that lands under a
				// naming family secretKindsOf does not yet know (say
				// `secret:entropy`) arrives here too, and that is the gap CLOSING —
				// weakening the specimen would then hide the very coverage this pin
				// exists to notice.
				t.Errorf("%s: unexpected non-secret kind now matches this specimen: %v — if this is a network/identity pattern widening, adjust the specimen so it isolates the token/secret question again; but if the new kind IS a secret or entropy detector under a different naming family, this is coverage GROWTH: re-point this pin and teach secretKindsOf the new family rather than weakening the specimen", c.name, kindsOf(found))
			}
			continue
		}
		t.Logf("PASSED THROUGH  %s", c.name)
	}
}

// secretKindsOf returns the token/secret-family kinds among a finding slice, or
// nil when there are none. The family is the `token:` namespace plus
// `rp_session_key`, the one non-prefixed secret kind in the bundled set — the
// kinds whose growth is what iss-96 is about. Network and identity kinds are
// deliberately excluded: they are covered elsewhere and have their own pins.
//
// The family list is the pin's blind spot, so it must be MAINTAINED: a secret or
// entropy detector shipping under a new namespace (`secret:`, `entropy:`) would
// close the tracked gap while reading here as "not a secret kind". The caller's
// failure message says so; add the family here when that happens.
func secretKindsOf(f []Finding) []string {
	var out []string
	for _, fn := range f {
		if strings.HasPrefix(fn.Kind, "token:") || fn.Kind == "rp_session_key" {
			out = append(out, fn.Kind)
		}
	}
	return out
}

// kindsOf lists the kinds of a finding slice for a failure message. It never
// prints a matched value, so a diagnostic cannot echo the specimen; every
// diagnostic in this file names the CASE LABEL instead.
func kindsOf(f []Finding) []string {
	out := make([]string, 0, len(f))
	for _, fn := range f {
		out = append(out, fn.Kind)
	}
	return out
}
