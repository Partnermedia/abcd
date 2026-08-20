# `/abcd:update` — Complete a Chosen Update

`/abcd:update` completes what `version --check` reports: it fetches the named
release (or resolves the latest, naming the tag before acting), verifies the
platform binary against the same release's `checksums.txt`, and swaps the
PATH-installed copy atomically. The verb is the explicit ask — abcd never
checks for or applies updates on its own
([adr-38](../../decisions/adrs/0038-implicit-checks-are-disk-only.md)); this
verb and `version --check` are the only two paths to the release origin, each
only when invoked ([itd-130](../../intents/planned/itd-130-abcd-update-completes-a-chosen-update-in-one-verb-it-fetches.md), spc-32).

## Behaviour

```bash
abcd update [tag] [--yes] [--json]
```

The dispatch is keyed on what actually runs — the first `abcd` PATH occupant,
classified by the same ownership predicate detection and install use. Only a
regular file proven to be abcd's by provenance (its digest appears in a
published release's `checksums.txt`) is ever swapped; every other shape is a
loud refusal naming its remedy: a plugin-root binary (the plugin update owns
it — itd-108's one-cut coherence), the track-latest dev shim, a stranded
owned entry (`ahoy install` heals it), a Homebrew Cellar-resolved install
(`brew upgrade abcd`), a foreign occupant, or an empty PATH.

The transport is pinned: no proxy or CA overrides from the environment (set
ones are ignored and named in the receipt), redirects only onto the release
origin's own hosts, every hop re-checked under the urlguard policy. The swap
is atomic in the target's directory; a failed download or verification
leaves no partial file. Progress renders on a TTY only; the receipt (origin,
tag, digest, old→new) prints in both modes.

## References

- Plugin command: [`commands/update.md`](../../../../commands/update.md)
- Intent / spec: [itd-130](../../intents/planned/itd-130-abcd-update-completes-a-chosen-update-in-one-verb-it-fetches.md) / [spc-32](../../specs/open/spc-32-abcd-update-completes-a-chosen-update-in-one-verb-it-fetches.md)
- Staleness check it completes: [`12-version.md`](12-version.md)
