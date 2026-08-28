# Secret-shaped fixtures are built at runtime

**The rule.** A test that must exercise the secret scanner, its adapters, or any
redaction path needs realistic secret-shaped values, but it never commits one
verbatim. The fixture is CONSTRUCTED AT RUNTIME — from a seed, a concatenation,
a generator — so no literal secret, real or fake, ever appears in source.

**Why.** The commit gate is a full-history secret scan (`gitleaks git`,
`fetch-depth: 0` in CI): it reads every commit's diff, not just the working
tree. A committed literal trips the scan on the commit that introduced it, and
that verdict is permanent — the only ways to clear it are a history rewrite
(which the never-force-push rule forbids) or a config allowlist (which widens the
blind spot the scanner exists to close). Constructing the value at runtime means
the scan has nothing to find, on any commit, forever, with no allowlist.

**The enforcing gate is already armed.** `gitleaks` in CI is the detector; this
principle is how a fixture complies with it, not a new gate to build. When it
fires on a fixture, the fix is to generate the value, never to allowlist the
file — an allowlist tells the scanner to stop looking exactly where a real leak
could later hide.

**The compliance primitive.** `internal/testsecret.Synthetic(seed, n)` returns a
deterministic, high-entropy, alphanumeric string shaped like a generic API key,
built at runtime. A fixture calls it (`secret := testsecret.Synthetic(96, 40)`)
instead of pasting a literal; the value is stable per (seed, n) so tests stay
reproducible. Extend that one generator rather than pasting a second literal.

**Scope and residue.** The rule binds new and edited fixtures. A literal already
merged into history cannot be un-committed without a rewrite, so a pre-existing
one stays covered by the narrow `.gitleaks.toml` path allowlist until the
scanner's pinned version would flag it; the allowlist is the fallback for
history, never the pattern for new code.
