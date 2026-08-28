// Package testsecret builds synthetic, secret-shaped strings at RUNTIME for
// tests that must exercise the secret scanner and its adapters. No literal
// secret is ever committed to source, so the full-history secret scan
// (gitleaks git, fetch-depth 0 in CI) has nothing to find — see the principle
// secret-shaped-fixtures-at-runtime. A committed literal, even a fake one,
// trips the scan on the commit that added it and cannot be removed without a
// history rewrite, so the fixture is generated instead.
package testsecret

// alphabet is the base62 set a generic-API-key-shaped token draws from.
const alphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"

// hexAlphabet is the lower-case hex set the other opaque-identifier spellings
// draw from — a hex session id and the runs a UUID is assembled out of.
const hexAlphabet = "0123456789abcdef"

// Synthetic returns a deterministic length-n high-entropy alphanumeric string
// shaped like a generic API key, built at runtime from seed. It is stable for a
// given (seed, n) so tests are reproducible, and it never appears verbatim in
// source. n is clamped to at least 1.
func Synthetic(seed uint64, n int) string {
	return walk(seed, n, alphabet)
}

// SyntheticHex returns a deterministic length-n lower-case hex string built at
// runtime from seed — the other spelling an opaque identifier takes (a hex
// session id, and the runs a UUID is assembled out of). Same contract as
// Synthetic: stable for a given (seed, n), never verbatim in source, n clamped
// to at least 1.
func SyntheticHex(seed uint64, n int) string {
	return walk(seed, n, hexAlphabet)
}

// walk is the one generator both spellings share: a cheap deterministic
// high-entropy step over the given alphabet.
func walk(seed uint64, n int, set string) string {
	if n < 1 {
		n = 1
	}
	b := make([]byte, n)
	x := seed ^ 0x9e3779b97f4a7c15
	for i := range b {
		// splitmix-style step: a cheap deterministic high-entropy walk.
		x = x*6364136223846793005 + 1442695040888963407
		b[i] = set[(x>>40)%uint64(len(set))]
	}
	return string(b)
}
