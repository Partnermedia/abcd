# Prefer the experiment to the inference

**The rule.** When a claim about a system's behaviour can be settled by running
it, and running it is cheap, run it before asserting the claim. Reasoning from
configuration files, manifests and workflow definitions produces a hypothesis,
never a finding — and a hypothesis stated as a fix is a false record with the
author's confidence attached.

**Why.** Configuration is a description of intent, and the gap between intent
and behaviour is exactly where defects live. Reading the files tells you what
someone meant; only running it tells you what happens.

The 2026-08-23 release drew the line empirically. A site deploy failed with
`CLOUDFLARE_API_TOKEN` unset, and two diagnoses were reasoned from the
repository and both asserted as fixes. First, that the production secrets had
never been created: false, they existed and were correctly scoped hours before
the failure, and acting on it nearly created a credential that was not needed.
Second, that `secrets: inherit` was missing from the workflow call: necessary,
insufficient, and shipped in v0.6.3 as a completed fix, so a published release
carried a CHANGELOG entry titled "The release chain's site deploy receives its
credentials" for a deploy that still failed.

The experiment was one command, available from the first failure, and settled it
in a minute: dispatch the same workflow directly rather than through its caller.
Same job, same environment, same secrets, opposite outcome — which localised the
defect to the call path, something no amount of reading the files had produced
in two attempts (iss-2608231912566984).

**Bounds.**

- "Cheap" is the operative test. Where an experiment needs a public release, a
  destructive change, or hours of setup, inference is the honest instrument and
  the claim is labelled as inference.
- The bar rises with the claim. Reporting a hypothesis as a hypothesis is always
  fine. Asserting a fix, resolving a record, or writing a CHANGELOG entry is a
  claim about behaviour, and behaviour is what experiments establish.
- A second inference after a first one proved wrong is the strongest signal to
  stop and run something. Being wrong once means the model of the system is
  wrong, and the same model produced the next guess.
- This does not license changing production to see what happens. The experiment
  must be one whose failure costs no more than the ignorance it removes.

**Promotion.** No mechanical check distinguishes a tested claim from a reasoned
one; a record that carried its evidence — the command run and its result,
alongside the assertion — would promote this. `enforcement-claims-are-facts`
covers the neighbouring case where a gate's existence is asserted without
checking the gate fires; this one covers asserting a behaviour without running
it. Related: `reality-is-filable`, `fix-the-detector`.
