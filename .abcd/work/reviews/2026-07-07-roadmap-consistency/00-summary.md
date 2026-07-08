# Roadmap-Consistency Review - Consolidated Summary

**Scope:** `.abcd/development/brief/`, `.abcd/development/roadmap/`, and `.abcd/development/intents/`, with emphasis on phase ownership, lifecycle state, release policy, and command-surface consistency.

**Method:** external consistency review of the design record. The review looked for contradictions between canonical brief files, roadmap phase documents, ADRs, and intent lifecycle directories.

---

## Verdict

The current design record has several consistency gaps that make the roadmap difficult to use as an implementation scheduler. The highest-risk issues are phase ownership for `/abcd:launch`, the release-version source of truth, and disagreement between intent lifecycle directories and phase membership.

## Findings

| Severity | Enhancement | Why it matters |
|---|---|---|
| **Critical** | Reconcile `/abcd:launch` phase ownership. `roadmap/phases/phase-1-ahoy.md` and `brief/06-delivery/01-build-sequence.md` say install + launch is the first milestone, but `brief/04-surfaces/04-launch.md` says full launch builds in Phase 5. | Implementers cannot know whether launch is Phase 1 MVP work or late round-trip work. |
| **Critical** | Reconcile release-version source of truth. `roadmap/README.md` says plugin releases are tracked by `.claude-plugin/plugin.json`; ADR-31 says release versions are derived from shipped intents and never authored in the working tree; `brief/04-surfaces/04-launch.md` still describes `--version` and phase-completion tiering. | Release automation will encode the wrong versioning model unless one policy wins. |
| **Critical** | Make phase membership and lifecycle directories agree. `intents/README.md` says `planned/` contains capabilities scoped into roadmap phases, but many planned intents do not appear in any phase `## Scope` (`itd-20`, `itd-24`, `itd-29`, `itd-46`, etc.). Conversely, draft `itd-43` is scoped into Phase 0/1 and still listed later-phase. | The roadmap cannot be used as a scheduler if "planned", "phased", and "draft" disagree. |
| **Major** | Fix "six planned phases" to match the seven phase docs. `brief/01-product/04-scope.md`, `brief/06-delivery/03-out-of-scope.md`, and `brief/06-delivery/README.md` say six phases / Phase 0-5, but `roadmap/phases/README.md` has Phase 0-6 with Phase 6 as lifeboat round-trip. | Readers will miss or misplace the final round-trip phase. |
| **Major** | Move disembark/embark status banners to Phase 6. `brief/04-surfaces/02-disembark.md` says Phase 4, `brief/04-surfaces/03-embark.md` says Phase 5, while `roadmap/phases/phase-6-lifeboat.md` owns both. | The core product surface is assigned to three different milestones. |
| **Major** | Update the brief's command inventory to include `/abcd` and `/abcd:reflect`, or mark them non-user-facing consistently. `brief/01-product/04-scope.md` lists seven user-facing commands; `brief/04-surfaces/README.md` lists nine. | Scope, docs, and command wiring will diverge around the top-level status board and reflection surface. |
| **Major** | Replace hand-maintained intent counts/lists in `brief/01-product/04-scope.md`. It still says "thirteen phased intents"; the phase docs and `planned/` corpus now describe a much larger set. | This repeats the drift the roadmap says it avoids by deriving counts from disk. |
| **Major** | Regenerate `brief/06-delivery/03-out-of-scope.md` from the actual draft corpus. It omits current draft files such as `itd-74` and `itd-75`, while including scoped `itd-43`; its derivation command also excludes only a stale phased-in set. | The canonical later-phase list is no longer canonical. |
| **Major** | Standardise the glossary/terminology path. Phase 3 still writes glossary output to `.abcd/development/foundation/terminology/`, while the current record uses `brief/glossary/` and ADR-30 names that as the bounded-context glossary home. | Lint, grill, and documentation will write/read different vocabulary stores. |
| **Major** | Align Phase 0's "no user-facing command" claim. `roadmap/phases/phase-0-substrate.md` explicitly says Phase 0 carries `/abcd:intent review`; `roadmap/phases/README.md` says Phase 0 has no user-facing command. | The phase principle is sound, but the wording currently contradicts the exception. |
| **Minor** | Fix agent count drift. `brief/05-internals/01-agents.md` declares 16 agents; `brief/01-product/01-press-release.md`, `brief/01-product/04-scope.md`, and `brief/05-internals/05-prompt-quality.md` still say 15. | This is easy to fix and prevents future generated prompt/test counts being wrong. |
| **Minor** | Resolve `/abcd:memory` shipped/planned state. Memory docs say spc-38/spc-39 shipped, but `itd-36` lives in `planned/` and `intents/README.md` says `shipped/` is empty until Go capabilities ship. | The shipped/planned model loses authority if product docs describe shipped capability outside `shipped/`. |
| **Minor** | Remove stale status narration from current-state brief files, or move it to spec/logbook records. Many files contain "spc-X shipped", "today only stubs ship", and historical notes inside canonical design contracts. | ADR-5 says the brief is current state; embedded delivery history will keep re-drifting. |
| **Nitpick** | Fix stale deep links to `brief/04-surfaces/05-intent.md` section 6. `brief/04-surfaces/README.md` and `brief/05-internals/01-agents.md` point to the reviewer as section 6, but the current section appears later in the file. | Low severity, but it undermines the brief's use as agent-loadable reference material. |
