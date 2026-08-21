package recordid

// mint.go is the WRITE side of this package's id space: the native
// timestamp-numeric mint (adr-45; mechanics fixed by spc-33).
//
// A native id is <family>-<yymmddHHMMSS><rrrr>: a 12-digit UTC second stamp and
// a 4-digit uniform random suffix — 16 digits, matching the [0-9]+ grammar every
// id consumer already parses. The mint never looks at any existing maximum
// (adr-45 ruling 2): time orders the ids, entropy separates two minters landing
// in the same second on branches that share no filesystem, and the ledger's
// O_EXCL reservation resolves same-ledger clashes by redrawing (spc-33 ruling
// 2). The clock and entropy are injectable seams so the same-instant and race
// acceptance tests are deterministic.

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"io"
	"regexp"
	"time"
)

// stampLayout renders the mint instant as yymmddHHMMSS, always in UTC — a
// zone-local stamp would let two minters in different timezones invert the
// global time-order the timestamp exists to provide.
const stampLayout = "060102150405"

// suffixMod is the size of the suffix space: 4 uniform decimal digits (spc-33
// ruling 1 — sized for the cross-branch same-second coincidence, the one case
// no lock can arbitrate).
const suffixMod = 10000

// suffixRejectAbove is the largest multiple of suffixMod representable in a
// 16-bit draw; draws at or above it are rejected and redrawn so the modulo
// fold introduces no bias.
const suffixRejectAbove = 60000

// mintFamilyRe bounds the family tag: it is spliced into an id and then a
// filename, so nothing beyond lowercase letters may pass.
var mintFamilyRe = regexp.MustCompile(`^[a-z]+$`)

// Minter mints native timestamp-numeric record ids. Both fields are seams for
// tests; the zero value is the production configuration (time.Now and
// crypto/rand).
type Minter struct {
	// Now supplies the mint instant; nil means time.Now.
	Now func() time.Time
	// Entropy supplies the suffix draw; nil means crypto/rand.Reader.
	Entropy io.Reader
}

// Mint returns a fresh native id for family (e.g. "iss"): the family tag, the
// UTC second stamp, and a uniform 4-digit suffix. It reads nothing — no ledger,
// no refs, no maximum — so two minters can never converge on one id by sharing
// a stale view; only a same-second suffix coincidence remains, which is the
// armed uniqueness detectors' residue to assert (adr-45 ruling 5).
func (m Minter) Mint(family string) (string, error) {
	if !mintFamilyRe.MatchString(family) {
		return "", fmt.Errorf("record-id family must be lowercase letters, got %q", family)
	}
	now := m.Now
	if now == nil {
		now = time.Now
	}
	entropy := m.Entropy
	if entropy == nil {
		entropy = rand.Reader
	}
	suffix, err := uniformSuffix(entropy)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s-%s%04d", family, now().UTC().Format(stampLayout), suffix), nil
}

// uniformSuffix draws a uniform value in [0, suffixMod) by rejection sampling
// over 16-bit reads: a draw in the truncated top band would fold unevenly under
// modulo, so it is discarded and drawn again. A short or failing entropy read
// is an error — never a partial or zeroed suffix.
func uniformSuffix(entropy io.Reader) (int, error) {
	var buf [2]byte
	for {
		if _, err := io.ReadFull(entropy, buf[:]); err != nil {
			return 0, fmt.Errorf("record-id entropy read failed: %w", err)
		}
		v := int(binary.BigEndian.Uint16(buf[:]))
		if v >= suffixRejectAbove {
			continue
		}
		return v % suffixMod, nil
	}
}
