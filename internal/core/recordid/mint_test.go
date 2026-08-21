package recordid

import (
	"bytes"
	"io"
	"regexp"
	"strings"
	"testing"
	"time"
)

// fixedClock returns a Now seam pinned to t.
func fixedClock(t time.Time) func() time.Time {
	return func() time.Time { return t }
}

// entropyBytes returns an Entropy seam that yields exactly b then EOF.
func entropyBytes(b ...byte) io.Reader {
	return bytes.NewReader(b)
}

// TestMintFormat pins the native id shape (spc-33 ruling 1): a family tag, the
// 12-digit UTC yymmddHHMMSS stamp from the injected clock, and a 4-digit
// zero-padded uniform suffix — 16 digits total.
func TestMintFormat(t *testing.T) {
	clock := fixedClock(time.Date(2026, 8, 20, 11, 42, 7, 0, time.UTC))
	// 0x0315 = 789 -> suffix "0789" (zero-padded to fixed width).
	m := Minter{Now: clock, Entropy: entropyBytes(0x03, 0x15)}
	id, err := m.Mint("iss")
	if err != nil {
		t.Fatalf("Mint: %v", err)
	}
	if want := "iss-2608201142070789"; id != want {
		t.Fatalf("Mint = %q, want %q", id, want)
	}
	if !regexp.MustCompile(`^iss-[0-9]{16}$`).MatchString(id) {
		t.Fatalf("minted id %q does not match the 16-digit shape", id)
	}
}

// TestMintStampIsUTC pins the stamp to UTC (spc-33): a clock reading in a
// non-UTC zone must not shift the stamp — global time-order across minters in
// different timezones is the point of the timestamp.
func TestMintStampIsUTC(t *testing.T) {
	zone := time.FixedZone("UTC+5", 5*3600)
	clock := fixedClock(time.Date(2026, 8, 20, 23, 30, 0, 0, zone)) // 18:30 UTC
	m := Minter{Now: clock, Entropy: entropyBytes(0x00, 0x00)}
	id, err := m.Mint("iss")
	if err != nil {
		t.Fatalf("Mint: %v", err)
	}
	if !strings.HasPrefix(id, "iss-260820183000") {
		t.Fatalf("Mint = %q, want the 18:30 UTC stamp, not the zone-local one", id)
	}
}

// TestMintSameInstantDiffers is itd-114's first acceptance criterion at the
// seam: two minters whose injected clocks read the same instant produce
// different ids, because independent entropy separates them.
func TestMintSameInstantDiffers(t *testing.T) {
	instant := time.Date(2026, 8, 20, 11, 42, 7, 0, time.UTC)
	a := Minter{Now: fixedClock(instant), Entropy: entropyBytes(0x00, 0x2A)}
	b := Minter{Now: fixedClock(instant), Entropy: entropyBytes(0x11, 0x11)}
	idA, err := a.Mint("iss")
	if err != nil {
		t.Fatalf("Mint a: %v", err)
	}
	idB, err := b.Mint("iss")
	if err != nil {
		t.Fatalf("Mint b: %v", err)
	}
	if idA == idB {
		t.Fatalf("same-instant mints collided: %q", idA)
	}
	if idA[:len("iss-260820114207")] != idB[:len("iss-260820114207")] {
		t.Fatalf("same-instant mints disagree on the stamp: %q vs %q", idA, idB)
	}
}

// TestMintFamilyGeneric pins itd-114's rollout criterion: a second family is a
// configuration of the same seam — the identical call mints itd and spc ids.
func TestMintFamilyGeneric(t *testing.T) {
	clock := fixedClock(time.Date(2026, 8, 20, 11, 42, 7, 0, time.UTC))
	for _, family := range []string{"itd", "spc"} {
		m := Minter{Now: clock, Entropy: entropyBytes(0x00, 0x07)}
		id, err := m.Mint(family)
		if err != nil {
			t.Fatalf("Mint(%q): %v", family, err)
		}
		if want := family + "-2608201142070007"; id != want {
			t.Fatalf("Mint(%q) = %q, want %q", family, id, want)
		}
	}
}

// TestMintRejectsBadFamily: the family tag reaches a filename, so anything
// outside the lowercase-letter grammar is refused before it can touch a path.
func TestMintRejectsBadFamily(t *testing.T) {
	m := Minter{Now: fixedClock(time.Unix(0, 0)), Entropy: entropyBytes(0x00, 0x00)}
	for _, bad := range []string{"", "Iss", "iss2", "../evil", "iss-"} {
		if id, err := m.Mint(bad); err == nil {
			t.Fatalf("Mint(%q) = %q, want a refusal", bad, id)
		}
	}
}

// TestMintSuffixRejectionSampling pins the uniform draw (spc-33 ruling 1): a
// 16-bit draw at or above the largest multiple of 10000 is rejected and drawn
// again, never folded by modulo — so the suffix distribution carries no bias.
func TestMintSuffixRejectionSampling(t *testing.T) {
	clock := fixedClock(time.Date(2026, 8, 20, 11, 42, 7, 0, time.UTC))
	// 0xFFFF (65535) must be rejected; the retry draw 0x1770 (6000) is taken.
	m := Minter{Now: clock, Entropy: entropyBytes(0xFF, 0xFF, 0x17, 0x70)}
	id, err := m.Mint("iss")
	if err != nil {
		t.Fatalf("Mint: %v", err)
	}
	if want := "iss-2608201142076000"; id != want {
		t.Fatalf("Mint = %q, want %q (the 65535 draw must be rejected, not folded)", id, want)
	}
}

// TestMintEntropyExhaustion: an entropy source that cannot supply a full draw
// is an error, never a partial or zeroed suffix.
func TestMintEntropyExhaustion(t *testing.T) {
	m := Minter{Now: fixedClock(time.Unix(0, 0)), Entropy: entropyBytes(0x01)}
	if id, err := m.Mint("iss"); err == nil {
		t.Fatalf("Mint with exhausted entropy = %q, want an error", id)
	}
}

// TestMintZeroValueUsesProductionSeams: the zero-value Minter mints with the
// real clock and crypto entropy — the production configuration needs no wiring.
func TestMintZeroValueUsesProductionSeams(t *testing.T) {
	id, err := Minter{}.Mint("iss")
	if err != nil {
		t.Fatalf("zero-value Mint: %v", err)
	}
	if !regexp.MustCompile(`^iss-[0-9]{16}$`).MatchString(id) {
		t.Fatalf("zero-value Mint = %q, want the 16-digit shape", id)
	}
}
