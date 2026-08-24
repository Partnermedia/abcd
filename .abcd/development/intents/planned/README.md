# Planned

Committed capabilities awaiting their Go build — some scheduled into a roadmap
phase, others committed but not yet sequenced (the two axes are orthogonal, per
[adr-34](../../decisions/adrs/0034-lifecycle-and-scheduling-orthogonal.md)). An
intent's `spec_id` is `null` until the native spec layer schedules it (Phase 4),
then points at a `spc-N`; bundle members share one `spec_id` with their
bundle-mates. An intent moves on to `shipped/` automatically when its linked spec
closes.

See the [intents index](../README.md) for the corpus listing and the
[intent surface contract](../../brief/04-surfaces/05-intent.md) for the full
lifecycle.
