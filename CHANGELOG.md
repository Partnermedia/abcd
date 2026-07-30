# Changelog

All notable changes to abcd are recorded here. The format follows
[Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and abcd
uses [Semantic Versioning](https://semver.org/spec/v2.0.0.html) with a
leading `v`.

Before v1.0.0, minor releases may make breaking changes; each one is
called out in a **Breaking** section.

## [Unreleased]

### Added

- **A citations family in `abcd docs lint`, with zero network in the gate**
  (itd-101, spc-17). Cited references rot silently — pages retitle, URLs
  redirect, whole platforms announce their own shutdown — but a gate that dials
  out to notice is a gate that flakes. So the checking splits in two. The lint
  side, landing here, reads only committed markdown, committed config, and a
  committed baseline: `citation_footnotes` holds a page's footnote markers and
  definitions in bijection (an unreferenced definition counts, being the way a
  reference page quietly stops meaning what it says); `citation_crosswalk_rows`
  requires every crosswalk table row to carry a footnote; `citation_url_syntax`
  checks cited URLs and DOIs are well-formed; and `citation_source_policy`
  refuses aggregator domains named in config — a list that ships empty, because
  naming one is a project's editorial policy and never something the gate
  invents on its behalf. A page's citations are its footnote definitions, not
  its prose, so ordinary body links stay `links_resolve`'s business.
  `citation_baseline` enforces the committed record offline — no cited URL
  without an entry, none recorded broken, none whose recorded final address has
  drifted from what the page cites, and a 180-day staleness warning. Each rule
  is opt-in per repo, and the whole family is inert until configured.
- **`abcd docs cite refresh` — the one verb that fetches, so the gate never has
  to** (itd-101, spc-17). It collects every cited URL through the same collector
  the lint uses, gives each exactly one bounded attempt, and rewrites the
  committed baseline. It never retries — a run's cost cannot become a function of
  how many links are failing, nor a burst against a struggling host — and it never
  reads a response body, because liveness is a property of the status line and a
  citation to a huge file should cost a response header, not a download. Sources
  that refuse automated fetchers (401, 403, 406, 429) are **not** recorded as
  broken: a refusal says the fetcher may not look, which is a different fact from
  the source being gone, so those URLs get no invented entry and are printed as a
  manual checklist naming exactly where each is cited. A current human-verified
  receipt is preserved verbatim and not even re-requested; once it ages past the
  staleness threshold it is re-checked like any other, because human and machine
  verifications share one clock. Receipts for addresses the documentation no
  longer cites are dropped.
- **`abcd docs cite confirm` — recording that a human cleared a queued link.**
  Name the URLs, or pass `--receipt` with a receipt file; both assemble the same
  schema, so the generated checklist page that lands later is a different producer
  of one input rather than a second pathway. Only URLs the documentation actually
  cites can be confirmed, and one bad line refuses the whole receipt rather than
  half-applying it. The receipt records **that** a human verified a citation and
  **when**, never how — the schema declares no such field and decoding rejects
  unknown keys outright.
- **The citation baseline surfaces where maintainers already look.** `abcd ahoy`
  reports its coverage and age on one line in any repo that has armed the rule;
  `abcd launch --dry-run` gains a `citation-baseline` gate that names entries
  approaching the staleness blocker while still letting a release cut, and refuses
  on ones that are overdue, broken, or unreceipted. `abcd docs lint
  --release-gate` promotes an overdue citation from a warning to a blocker — the
  flag is the trust root, so a repository cannot defang its own release by editing
  a committed config, and an ordinary commit is never blocked by the calendar.
  The flag is built and tested but **not yet passed by `release.yml`**, which
  still runs the plain `abcd docs lint`; wiring it into the release workflow is a
  CI change needing its own sign-off, and until it lands the 365-day threshold
  warns at release time rather than blocking.
- **The citation gate is armed for this repository.** 50 of the 51 URLs cited
  under `docs/` carry a receipt, and four citations that had silently drifted
  behind redirects now name the address they actually resolve to — rot nobody
  would have caught by reading. The fifty-first answers HTTP 403 to every
  automated fetcher and is waiting in the manual queue, so `abcd docs lint`
  reports it as unreceipted until a maintainer opens it and runs `abcd docs cite
  confirm`. That report is the mechanism working: no receipt is ever written on
  a human's behalf.
- **A schema-versioned citation baseline at `.abcd/citations-baseline.json`.**
  Per cited URL it records the final resolved address, when it was last checked,
  the outcome, and whether verification was automatic or manual with its date.
  It records nothing about *how* a human verified, and cannot be made to: the
  schema declares no such field, and loading rejects unknown keys outright, so a
  hand-added `method` or transcript is a refusal rather than a quietly-kept
  note. Manual entries age on the same clock as automatic ones, so a human
  confirmation buys no exemption from going stale.
- **`abcd launch scaffold` — the changelog-driven release-gate scaffolder**
  (itd-93, spc-14). Writes the fixed release machinery into a managed repo that
  lacks it: `.github/workflows/release.yml` (verify → build → publish, the verify
  gate armed against the reviewed **content** commit so the first public release
  cannot hit the receipt-vs-tag self-reference), `.github/workflows/auto-release.yml`
  (newest dated CHANGELOG heading → tag that commit → call `release.yml`), and the
  adr-37 runbook — wired to the repo's own default branch and Go version,
  `GITHUB_TOKEN`-only and injection-safe. The workflows ship from a **single
  embedded template** that abcd-cli's own release workflows are regenerated from
  (self-scaffold parity, proven by a byte-exact test), so every abcd release
  exercises the exact machinery a managed repo receives. The scaffolded
  `release.yml` carries a `workflow_dispatch` **rehearsal** that arms the full gate
  against a simulated changelog roll and reviewed-content commit and publishes
  nothing — a green rehearsal is the runbook precondition for the first real
  release. A bare repo with no semantic detector degrades cleanly to the
  deterministic gates and a generic build. The verb is idempotent and fail-safe: a
  re-run on current machinery is a no-op, a hand-edited file is refused rather than
  clobbered (unless `--confirm`), and it refuses rather than half-writing.
- **A public terminology crosswalk at `docs/reference/terminology.md`** (itd-100).
  One reference page maps 26 established agentic-AI terms — protocols, the core
  loop, context, safety, governance, operations — to abcd's position on each:
  USES (naming the native verb or principle), ADAPTS (the sharper native name and
  why), REJECTS (with the recorded reason), or WATCHING (with the record id).
  Every established definition carries a footnote citation to a primary source
  (specification, standards body, DOI-bearing paper, or origin engineering doc);
  every abcd claim is grounded in the committed record. Vendor names appear only
  inside citation footnotes, keeping the page body host-agnostic.

- **`abcd ahoy install --dev` — a track-latest dogfood install mode** (iss-75).
  Normal `ahoy install` symlinks the pinned built binary, so tracking live
  development meant hand-rolling a `~/.local/bin/abcd` wrapper that ran
  `go build -C <repo> && exec` on every call. That manual workaround now dies:
  `--dev` installs a shim at the same `PATH` target that rebuilds abcd from the
  source tip on every invocation and execs the fresh binary. A broken build fails
  loudly and never execs a stale binary (loud-staging). `abcd ahoy` status reports
  the mode as `install: dev (tip build)`, detected from the installed shim itself
  (never recorded in the tracked repo config, so it can never go stale), so a dev
  install is never invisible. Installing over an existing
  install applies-as-update in either direction — `--dev` replaces the pinned
  symlink with the shim, a plain re-install restores the symlink — and a foreign
  occupant is still never clobbered.
- **Record-id minting now sees every branch, and a spec-id uniqueness lint closes
  the class** (iss-115, iss-120). Sequential ids (`iss-N`, `itd-N`, `spc-N`) were
  minted from the local working tree only, so two branches cut from the same base
  silently minted the same next id — invisible on each branch and surfacing only
  at merge. Minting now folds in the highest id committed on every local and
  remote-tracking branch (a single canonical refs-union scan), so once one branch
  commits an id, the other mints past it. When git cannot be read over a present
  repository the mint degrades to working-tree-only and says so loudly on stderr
  (never a silent fallback); a directory that is not a repository has no branches
  to collide with and mints quietly. The residual window — two branches that both
  mint before either commits — is caught by the record-lint uniqueness rules on
  the merged pull request, which now cover spec ids too: the new `spec_id_unique`
  rule flags every file claiming a duplicate `spc-N`, mirroring the existing
  `issue_id_unique` and intent-id guards.
- **`abcd capture resolve` and `abcd intent "<text>"` can now stamp a product
  `impact`** (iss-117). A resolved issue and a shipped intent are in the release
  set, so the `issue_impact_valid` and `intent_impact_valid` record-lint blockers
  require a valid `impact` on those records — but the verbs that mint them had no
  way to set one, so the tool's own path produced records its own gates rejected.
  `capture resolve` now takes a mandatory `--impact <additive|breaking|fix|internal>`
  (there is no default: an absent or misspelled value is refused, not guessed),
  and `abcd intent "<text>"` takes an optional `--impact <additive|breaking|fix>`
  that is stamped onto the seeded draft and travels unchanged through planning to
  `shipped/`. `internal` is rejected on an intent (a press-release-first intent is
  user-facing by definition). `capture wontfix` is unchanged — a non-action ships
  nothing, so `wontfix/` carries no impact.

### Fixed

- **`ahoy install` writes a `.gitignore` block that matches the tier layout it
  documents** (iss-169). The managed block ignored a root-level `.work/` under
  both visibilities — a path the three-tier layout does not have — while never
  naming `.abcd/.work.local/`, the one tier that must be gitignored. So a fresh
  install ignored nothing that existed and left the local-ephemeral tier tracked:
  precisely the state `abcd audit`'s `three-tier-layout` rule then reports as an
  error, the installer and the auditor disagreeing about the same convention. The
  block now follows the brief's visibility table exactly — a private repo ignores
  `.abcd/.work.local/` alone, because the rest of the `.abcd/` namespace is meant
  to be committed; a public repo ignores `.abcd/` outright, one switch with no
  per-subdirectory exceptions, plus the legacy root-level `memory/` snapshot. An
  existing repo carrying the old block needs no migration step: the block reads
  as drift, `abcd ahoy` reports it, and one apply replaces it, leaving the
  repository's own ignore rules untouched. A repository still on the historical
  root-level `.work/` layout stops ignoring `.work/` when the block refreshes;
  the layout migration in `/abcd:prepare-this-repo` is the path off that state.
- **The `PII` rules domain fires on network work, and says never to commit a
  hostname or an address** (iss-156). Its recall keywords were the vocabulary of
  credentials — secret, token, credential, pii, redact, hostname, email — so a
  session investigating a mesh VPN's reachability, a firewall rule, or an
  address in a network config matched nothing and the hook injected nothing.
  The keywords now cover that context — `ip`, `ips`, `ipv4`, `ipv6`, `vpn`,
  `tailscale`, `tailnet`, `wireguard`, `firewall`, `network`, `reachability`,
  `reachable`, `unreachable`, `dns`, `ssh`, `subnet`, plus the `mac address`
  phrase alias — and
  one new rule line forbids committing hostnames, IP or MAC addresses, and other
  live network identifiers: redact or omit them, and reach for a reserved
  documentation value (RFC 5737, RFC 3849, RFC 2606, RFC 7042 — the same set the
  audit privacy-hygiene rule cites) only where an illustrative
  example is actually needed. That rule previously existed only in a repo's own
  instructions file, which meant the one discipline this domain most needed to
  state was the one it never said — an agent could recall `PII` and still read
  nothing about the machine identifiers in front of it. Recall matching is
  word-bounded already, so the bare `ip` keyword matches a standalone token and
  never the middle of a word such as "script" or "zip"; the plural, the
  version-qualified forms and the `-able` adjective need their own entries
  because the stemmer has a three-character floor and does not bridge
  `-ability` to `-able`. `mac` is an alias phrase rather than a bare keyword so
  that an Apple Mac does not recall the domain.
- **`abcd audit` flags local-tier artefacts sitting in a committed tier**
  (iss-155). The `three-tier-layout` rule verified that the committed tiers
  exist and that `.abcd/.work.local/` is gitignored, but never the reverse
  containment: a `NEXT.md`, `scratch/` or `logs/` placed directly in
  `.abcd/work/` or `.abcd/development/` passed clean — which is exactly how a
  handover file carrying machine-local detail rides a committed tier into
  public history. Those three names are the local-ephemeral tier's
  conventional contents, so their presence in a committed tier is now an
  error, one finding per misplacement, each naming the offending path and
  fixing it with a move to `.abcd/.work.local/`. Presence is checked on the
  filesystem, like the tiers themselves: an untracked `NEXT.md` in
  `.abcd/work/` is one `git add -A` from being committed, and the audit
  should say so before that happens, not after.
- **Network identifiers are detected, redacted and audited — one pattern set,
  three surfaces** (iss-154, iss-157, iss-125, iss-153). The scanner carried
  token-shaped secrets and identity matchers but nothing for addresses or
  hostnames, so a lifeboat pack or launch bundle carrying a tailnet address and
  device names shipped clean, a stored transcript kept a LAN hostname verbatim,
  and `abcd audit` passed the whole class in silence. The detector is an
  allowlist inversion rather than a leak-recognition heuristic: because every
  illustrative identifier in a committed file comes from a reserved range, the
  small closed set of values that are ALLOWED is what a pattern can recognise,
  and everything outside it is a finding. Allowed are the documentation ranges
  (RFC 5737, RFC 3849, RFC 2606/6761, RFC 7042), device names derived from the
  persona registry, and the values that name no individual host at all —
  loopback, unspecified, netmasks, masked CIDR prefixes, and the IANA
  special-use ranges for link-local, multicast, benchmarking, protocol
  assignments and the NAT64 well-known prefix. Flagged is what identifies
  private topology: private ranges, CGNAT/tailnet, IPv6 unique-local, 6to4,
  public unicast, LAN suffixes (`.local`, `.lan`, `.fritz.box`) and
  device-hostname shapes. The set is built once and folded into the scanner's
  defaults, so Stage-1 redaction and the launch/lifeboat scan inherit it and the
  audit rule consults exactly the same patterns — the surfaces cannot disagree
  about what a leak is. Addresses are hard_fail; the two hostname shapes are
  warn, and the audit surface carries that split through rather than flattening
  it. A line carrying `abcd-audit:allow` stays exempt, so a deliberately
  illustrative value needs no weakening of the patterns. An identifier is masked
  WHOLE — to a readable placeholder in redacted text, and fully starred in a
  serialised finding — never to the head-and-tail fingerprint a credential gets:
  on a MAC or a hostname that fingerprint preserves the vendor bytes and the
  head, which is enough to re-identify the machine the masking was meant to hide.
  The two hostname shapes tell a host from source code, because Stage-1
  redaction rewrites every finding and a false positive there corrupts a stored
  transcript: a mixed-case suffix where code puts a value (`zone := time.Local`),
  a selector position (an assignment target, a block head, a call) and a
  determiner before a device noun ("a synology-nas") are all read as code or
  prose rather than machines. Mixed case on its own exempts nothing — a shifted
  key in a command or a URL would otherwise be a bypass anyone could type — and a
  fixture host must be spelled the way the persona registry spells it, in lower
  case, because a capitalised given name in front of a device noun is how macOS
  names a real person's machine. A digest is told from an address by the length
  of the colon-separated run around it, counting only the groups that carry hex:
  an address is eight of them and may sit beside one more (the port a tool
  prints), while the shortest fingerprint is sixteen. A
  repo that wants the hostname shapes to block raises their severity in
  `.abcd/config/pii.json`, and `abcd audit` reads that merged set, so the
  override reaches the surface that reports it. The transcript store's
  verification rescan refuses the write on ANY surviving identifier, not only a
  hard_fail one, so a warn-severity hostname cannot reach disk in silence, and
  the refusal it prints names the surviving kinds rather than a severity it does
  not gate on.
- **`/Users/Shared` and `/Users/Guest` stop reading as usernames** (iss-153).
  privacy-hygiene flagged every segment under a `/Users` root, so the macOS
  system directories that live there were reported as though they named a
  person, taxing product code that legitimately writes to one. The exemption is
  narrow: it is scoped to the `/Users` root, it covers the system directory
  itself and not what sits beneath it (a name nested under one — `Shared/<user>`
  — is still a user, and still flags), a segment that merely begins with a
  system-directory name is still a leak, and an empty or dots-only segment in
  between (`Shared/../<user>`) restores no shield. The audit rule holds those
  terms for both spellings it matches, the POSIX one and the Windows one — and
  because Windows accepts either separator within one path, a name written after
  a forward slash inside a Windows path is read as the nested name it is; the
  scanner's own identity matcher shares the allowlist for the POSIX paths it
  recognises, which are the only ones it has ever matched.
- **The disembark probe's recursive file walk is bounded per directory, opens
  each child in O(1), and skips the common ecosystems' dependency trees**
  (iss-112, iss-114, iss-116). The walk now reads every directory with a bounded
  `ReadDir` (the same 50 000-entry guard `ListDir` uses), so a single directory
  of millions of entries can no longer balloon memory before the file cap
  applies. It holds a sub-root per directory (`os.Root.OpenRoot`) instead of
  re-resolving every path from the containment root one component at a time, so a
  deep tree costs O(entries) rather than O(entries × depth) — a 48 000-directory
  depth-30 tree walks in ~1.4 s where the old walk took ~7 s, and the cost is now
  independent of depth. The `os.Root` containment guarantee is unchanged: a
  symlink is still refused rather than followed out of the tree. The skip set
  widens beyond Node and Go to the common dependency, cache, and build-output
  trees — Python (`.venv`, `venv`, `.tox`, `__pycache__`), Rust and generic
  build output (`target`, `build`, `dist`), and CocoaPods (`Pods`) — so a
  vendored `TODO` is no longer cited as the project's own open question, and a
  large dot-prefixed dependency tree can no longer exhaust the walk cap before
  the project's own `src/` is reached.
- **The open-questions marker scan no longer reads documentation about markers as
  open questions** (iss-111). The pattern that grounds `evidence/open-questions`
  admitted a bare uppercase `TODO`/`FIXME` followed by whitespace, so on a
  repository that documents its own conventions every prose mention of a marker
  was cited as a work marker — 14 such false positives across the durable record
  (`.abcd/development/`) on this repository, all documentation, none a real
  marker, down to 3 after the fix (each an irreducible prose quotation of the
  literal `TODO:` form). `TODO` and `FIXME`
  now require a trailing `:` or `(` (the conventional `TODO:` / `TODO(alice):`
  spellings), which is how genuine markers are almost always written; `XXX`,
  `HACK`, and `BUG` still admit their conventional bare form, because they are
  rarely written as bare words in prose and carry no measured false-positive
  cost.
- **Concurrent runs can no longer drop a repo registration or delete a
  just-committed issue file** (iss-101, iss-102). Two `abcd ahoy install` runs
  from different worktrees shared one `~/.abcd/history/index.json`, and its
  registration was an unlocked load-modify-write: atomic rename kept the file
  intact but the last writer clobbered the other's update, silently erasing a
  repo entry or a re-founding lineage link. The history registry now serializes
  its load-modify-write behind an inter-process lock and re-loads inside it, so
  concurrent registrations compose instead of overwriting; the store bootstrap
  creates `index.json` with an exclusive create, so exactly one racing run seeds
  it. The re-founding lineage confirmation is still asked before the lock is
  taken — never across an interactive prompt — and the state it validated is
  re-checked under the lock, surfacing a conflict rather than writing a link the
  user approved against a stale index. Separately, the capture ledger's orphan
  sweep and its commit write now take the same ledger lock, closing a window in
  which a capture stalled more than sixty seconds could have its committed issue
  file swept away after the capture reported success. The inter-process lock is a
  single shared primitive; the capture allocator and the history registry both
  route through it.

## [0.4.1] - 2026-07-28

### Added

- **`abcd identity` — one canonical identity block, and every surface held to
  it** (itd-102, spc-19). A project's positioning fragments silently: the README
  strapline, the package or plugin manifest description, and the conventions
  file's opening are edited at different moments until three surfaces tell three
  stories (the recorded iss-143 drift, which is this check's acceptance corpus).
  A repo now records one canonical markdown block — `- **Title:**`,
  `- **Tagline:**`, and an optional, wrappable `- **Pitch:**` under a recorded
  heading — and `.abcd/positioning.json` records only where that block lives, how
  loudly the family reports, and which surfaces render from it. Markdown stays
  the single source of truth. The registry is data, not branches: a surface names
  candidate files (the first present wins, so one entry covers several manifest
  formats), a locator (a capture-group regexp or a top-level JSON field), the
  block fields it must carry, and the template a proposal renders from; the three
  defaults are the README strapline, the manifest description, and the
  conventions opening, and a declared list replaces them so nothing is registered
  silently. A new `identity-positioning` rule runs the check on **every**
  `abcd audit`, naming the file, the exact drifted line, and the canonical line it
  should carry — warn-tier by default (it highlights, it never gates) and
  upgradeable per-repo to blocker. `abcd identity` renders the block and each
  surface's verdict; `abcd identity render` prints the proposed correction as a
  unified diff and **writes nothing** (autonomous rewriting is permanently out of
  scope — adopting a proposal is always the maintainer's move); `abcd identity
  init` records the block and the pointer at onboarding, adopting an existing
  block rather than re-interviewing over it. `/abcd:prepare-this-repo` gains the
  interview, detect-first. Every path the registry names is read and written
  inside an OS-enforced containment root, so an audited repo that commits a
  symlinked directory cannot make the check read a file the repository does not
  own and quote it into the report.

- **`/abcd:ideate` — the idea-admission gauntlet** (itd-104, spc-18). A big,
  unproven idea can be put through three legs before it becomes a record entry:
  primary-source research (each load-bearing claim checked against its **primary**
  source, never a secondary citation), a grill against the existing record, and an
  adversarial review that is fresh-context, off-policy, and receives the idea as an
  artefact of unknown authorship. The legs are host work; `abcd ideate record
  <idea-slug> --verdict-json <file|->` is the deterministic frame that validates
  them and writes the durable verdict — the dated record under
  `.abcd/development/research/` plus one pointer line in `.abcd/work/DECISIONS.md`.
  The verdict is recorded **whether the idea survives or dies**, and its rejected
  alternatives may be empty only behind an explicit marker, because silence and
  "nothing was weighed" read the same to a session tempted to re-propose the idea.
  Every record-grill hit is **cited by id and proved to resolve** in the
  repository: an id naming no record refuses the whole verdict and names the id.
  The legs travel as an ordered array so "three legs, in order" is checked rather
  than assumed, and refusals are whole-document — nothing is written unless
  everything validates. Ideate is **optional and never a gate**: the `intent` and
  `capture` routing help names it for big unproven ideas, and nothing requires it
  or warns when it is skipped.

- **`abcd guard` — the shell-hazard guard, wired into a live session** (itd-103,
  spc-16). `abcd guard check --command "<line>"` evaluates one candidate command
  against the hazard registry and reports allow, warn, or block: a blocker exits
  1 and answers with the plain-language why and the safe successor, a warn exits
  0 with the warning rendered, and a guard that cannot be evaluated at all exits
  2 rather than letting silence read as clearance. `abcd guard hook` is the host
  adapter: it reads a pre-tool-use hook payload, refuses a matching command with
  the successor as the block message, and lets everything else through. The
  installed hook entry wraps the binary call so a missing or broken abcd **fails
  open, loudly** — the command runs and the session carries an unmissable
  UNGUARDED warning, never a silent no-op and never a stuck session. `abcd ahoy`
  gains a `guard:` line reporting the three things that can independently be
  false — hook installed, binary reachable, registry loadable — plus a
  deliberately disabled registry, so a broken guard is visible from outside the
  session too. Every unguarded state is loud, including the deliberate one: a
  registry switched off in `.abcd/guard.json` makes each command it lets through
  carry the warning, so "off" can never pass for "clear". Per-repo overrides live
  in that one file and nowhere else — no flag, environment variable, or prompt
  disarms the guard for a session, so the change lands in a diff. An allow means
  no entry matched, never that a command is safe: a hazard reached another way —
  a string handed to an interpreter (`eval`, `sh -c`), a launcher the guard does
  not step over, a backtick substitution, or a form no entry describes — is not
  seen. Coverage is what the registry names, and the command reference says so.
  A registry that is switched off answers `abcd guard check` with a fault rather
  than a clearance, so a script using the verb as a gate cannot be waved through
  by an edit to `.abcd/guard.json`.
- **`abcd launch scaffold` — the changelog-driven release-gate scaffolder**
  (itd-93, spc-14). Writes the fixed release machinery into a managed repo that
  lacks it: `.github/workflows/release.yml` (verify → build → publish, the verify
  gate armed against the reviewed **content** commit so the first public release
  cannot hit the receipt-vs-tag self-reference), `.github/workflows/auto-release.yml`
  (newest dated CHANGELOG heading → tag that commit → call `release.yml`), and the
  adr-37 runbook — wired to the repo's own default branch and Go version,
  `GITHUB_TOKEN`-only and injection-safe. The workflows ship from a **single
  embedded template** that abcd-cli's own release workflows are regenerated from
  (self-scaffold parity, proven by a byte-exact test), so every abcd release
  exercises the exact machinery a managed repo receives. The scaffolded
  `release.yml` carries a `workflow_dispatch` **rehearsal** that arms the full gate
  against a simulated changelog roll and reviewed-content commit and publishes
  nothing — a green rehearsal is the runbook precondition for the first real
  release. A bare repo with no semantic detector degrades cleanly to the
  deterministic gates and a generic build. The verb is idempotent and fail-safe: a
  re-run on current machinery is a no-op, a hand-edited file is refused rather than
  clobbered (unless `--confirm`), and it refuses rather than half-writing.
- **A public terminology crosswalk at `docs/reference/terminology.md`** (itd-100).
  One reference page maps 26 established agentic-AI terms — protocols, the core
  loop, context, safety, governance, operations — to abcd's position on each:
  USES (naming the native verb or principle), ADAPTS (the sharper native name and
  why), REJECTS (with the recorded reason), or WATCHING (with the record id).
  Every established definition carries a footnote citation to a primary source
  (specification, standards body, DOI-bearing paper, or origin engineering doc);
  every abcd claim is grounded in the committed record. Vendor names appear only
  inside citation footnotes, keeping the page body host-agnostic.

- **`abcd ahoy install --dev` — a track-latest dogfood install mode** (iss-75).
  Normal `ahoy install` symlinks the pinned built binary, so tracking live
  development meant hand-rolling a `~/.local/bin/abcd` wrapper that ran
  `go build -C <repo> && exec` on every call. That manual workaround now dies:
  `--dev` installs a shim at the same `PATH` target that rebuilds abcd from the
  source tip on every invocation and execs the fresh binary. A broken build fails
  loudly and never execs a stale binary (loud-staging). `abcd ahoy` status reports
  the mode as `install: dev (tip build)`, detected from the installed shim itself
  (never recorded in the tracked repo config, so it can never go stale), so a dev
  install is never invisible. Installing over an existing
  install applies-as-update in either direction — `--dev` replaces the pinned
  symlink with the shim, a plain re-install restores the symlink — and a foreign
  occupant is still never clobbered.
- **Record-id minting now sees every branch, and a spec-id uniqueness lint closes
  the class** (iss-115, iss-120). Sequential ids (`iss-N`, `itd-N`, `spc-N`) were
  minted from the local working tree only, so two branches cut from the same base
  silently minted the same next id — invisible on each branch and surfacing only
  at merge. Minting now folds in the highest id committed on every local and
  remote-tracking branch (a single canonical refs-union scan), so once one branch
  commits an id, the other mints past it. When git cannot be read over a present
  repository the mint degrades to working-tree-only and says so loudly on stderr
  (never a silent fallback); a directory that is not a repository has no branches
  to collide with and mints quietly. The residual window — two branches that both
  mint before either commits — is caught by the record-lint uniqueness rules on
  the merged pull request, which now cover spec ids too: the new `spec_id_unique`
  rule flags every file claiming a duplicate `spc-N`, mirroring the existing
  `issue_id_unique` and intent-id guards.
- **`abcd capture resolve` and `abcd intent "<text>"` can now stamp a product
  `impact`** (iss-117). A resolved issue and a shipped intent are in the release
  set, so the `issue_impact_valid` and `intent_impact_valid` record-lint blockers
  require a valid `impact` on those records — but the verbs that mint them had no
  way to set one, so the tool's own path produced records its own gates rejected.
  `capture resolve` now takes a mandatory `--impact <additive|breaking|fix|internal>`
  (there is no default: an absent or misspelled value is refused, not guessed),
  and `abcd intent "<text>"` takes an optional `--impact <additive|breaking|fix>`
  that is stamped onto the seeded draft and travels unchanged through planning to
  `shipped/`. `internal` is rejected on an intent (a press-release-first intent is
  user-facing by definition). `capture wontfix` is unchanged — a non-action ships
  nothing, so `wontfix/` carries no impact.

### Fixed

- **Untrusted prose can no longer open raw HTML in a record abcd writes.** The
  shared cleaner every host-delegated ingest routes through — the release
  changelog ingest, the lifeboat synthesis writers, and the ideate verdict —
  neutralised newlines and HTML comment markers but left a bare tag intact. In
  CommonMark a `<` followed by a letter, `/`, `!`, or `?` opens an HTML block, and
  several of those block types run to the end of the document: a single
  `<script>` in one model-supplied field made every later section of the record
  render as inert text inside an unclosed element, while a forged section above it
  rendered normally — so an artefact whose whole value is that a later session
  trusts it could hide its own evidence. Every tag opener is now broken apart in
  the one canonical primitive, so the fix lands at all three boundaries at once.
  The neutralisation runs **after** terminal sanitisation, which closes a second
  hole: sanitising substitutes `?` for a masked rune, so `<` before an escape byte
  previously became `<?` — a processing instruction — after the old ordering.
- **The disembark probe's recursive file walk is bounded per directory, opens
  each child in O(1), and skips the common ecosystems' dependency trees**
  (iss-112, iss-114, iss-116). The walk now reads every directory with a bounded
  `ReadDir` (the same 50 000-entry guard `ListDir` uses), so a single directory
  of millions of entries can no longer balloon memory before the file cap
  applies. It holds a sub-root per directory (`os.Root.OpenRoot`) instead of
  re-resolving every path from the containment root one component at a time, so a
  deep tree costs O(entries) rather than O(entries × depth) — a 48 000-directory
  depth-30 tree walks in ~1.4 s where the old walk took ~7 s, and the cost is now
  independent of depth. The `os.Root` containment guarantee is unchanged: a
  symlink is still refused rather than followed out of the tree. The skip set
  widens beyond Node and Go to the common dependency, cache, and build-output
  trees — Python (`.venv`, `venv`, `.tox`, `__pycache__`), Rust and generic
  build output (`target`, `build`, `dist`), and CocoaPods (`Pods`) — so a
  vendored `TODO` is no longer cited as the project's own open question, and a
  large dot-prefixed dependency tree can no longer exhaust the walk cap before
  the project's own `src/` is reached.
- **The open-questions marker scan no longer reads documentation about markers as
  open questions** (iss-111). The pattern that grounds `evidence/open-questions`
  admitted a bare uppercase `TODO`/`FIXME` followed by whitespace, so on a
  repository that documents its own conventions every prose mention of a marker
  was cited as a work marker — 14 such false positives across the durable record
  (`.abcd/development/`) on this repository, all documentation, none a real
  marker, down to 3 after the fix (each an irreducible prose quotation of the
  literal `TODO:` form). `TODO` and `FIXME`
  now require a trailing `:` or `(` (the conventional `TODO:` / `TODO(alice):`
  spellings), which is how genuine markers are almost always written; `XXX`,
  `HACK`, and `BUG` still admit their conventional bare form, because they are
  rarely written as bare words in prose and carry no measured false-positive
  cost.
- **Concurrent runs can no longer drop a repo registration or delete a
  just-committed issue file** (iss-101, iss-102). Two `abcd ahoy install` runs
  from different worktrees shared one `~/.abcd/history/index.json`, and its
  registration was an unlocked load-modify-write: atomic rename kept the file
  intact but the last writer clobbered the other's update, silently erasing a
  repo entry or a re-founding lineage link. The history registry now serializes
  its load-modify-write behind an inter-process lock and re-loads inside it, so
  concurrent registrations compose instead of overwriting; the store bootstrap
  creates `index.json` with an exclusive create, so exactly one racing run seeds
  it. The re-founding lineage confirmation is still asked before the lock is
  taken — never across an interactive prompt — and the state it validated is
  re-checked under the lock, surfacing a conflict rather than writing a link the
  user approved against a stale index. Separately, the capture ledger's orphan
  sweep and its commit write now take the same ledger lock, closing a window in
  which a capture stalled more than sixty seconds could have its committed issue
  file swept away after the capture reported success. The inter-process lock is a
  single shared primitive; the capture allocator and the history registry both
  route through it.

## [0.4.0] - 2026-07-22

### Breaking

- **The record-lint banlist now requires a machine-readable successor and
  context.** Every `banned_tokens` entry must declare a non-empty `successor`
  (what to use instead of the retired token) and a non-empty `allow_context`
  (where the token is legitimately allowed); a config whose entry omits either
  is rejected at load rather than lints with a replacement that lived only in
  prose. Each finding auto-cites its successor, so the reader is told the
  replacement inline. The bundled record-lint and docs-lint banlists carry the
  new fields.

### Added

- **Generated CLI reference with a drift gate** (iss-47). `docs/reference/cli/`
  now holds `commands.md`, a per-command Markdown reference walked deterministically
  from the Cobra command tree (`cli.GenerateReference`) — usage line, summary, and
  flags for every user-facing verb, with the operator-internal hook subtree omitted.
  Refresh it with `go generate ./internal/surface/cli`; a `go test` drift gate
  regenerates the tree and fails the build whenever the committed page and the tree
  disagree, so the reference can never silently go stale. The walker is hand-rolled
  over stdlib and the existing Cobra dependency — no new module dependency.
- **`agent-observation` is now a valid `--source` value for `abcd capture`.**
  An autonomous run's self-observation had no honest surfacing channel and was
  reusing `agent-finding`. `agent-observation` parallels it ("an agent
  observed") without being tied to one run mode (iss-57).

- **`issue_id_unique` record-lint rule — a duplicate issue id is a blocker.**
  The lint pass now scans the issue ledger's three status directories
  (`open/`, `resolved/`, `wontfix/`) and flags any `iss-N` id claimed by two or
  more files, mirroring the existing intent-id uniqueness check (both share one
  `validateIDUnique` primitive). The capture allocator already rejects a
  duplicate on the reservation path, but a hand-added issue file that bypassed
  it slipped straight through until now; this is the backstop that catches it.
  Every file in a colliding set is flagged, since the linter cannot know which
  claimant is authoritative.
- **`abcd intent ready <itd-N>` — the implement-readiness gate.** A read-only
  verb reporting whether an intent may be implemented now: planned
  (directory-as-truth), enumerable Acceptance Criteria, a bidirectional spec
  link, and a spec body written past its minted stub. Every check carries a
  reason and, when failing, the exact remedy command. The exit code is the
  machine seam: 0 ready, 1 not ready (the rendered report is the output), 2
  structural fault — so an autonomous run can gate on it (step 0 of the run
  protocol) instead of re-deriving readiness from prose, and an unplanned
  intent is refused rather than improvised against.
- **`/abcd:intent` plugin command surface** (`commands/abcd/intent.md`),
  covering the full verb family — status, quoted-text create, `ready`, `plan`,
  `link`, `review`/`ingest` — plus the host-run planning interview an unready
  intent is routed to: the human confirms the press release, resolves open
  questions, and accepts, edits, or strikes every acceptance criterion before
  `abcd intent plan` is run as their sign-off act.

### Fixed

- **The published plugin now ships its agents and hooks.** The launch payload's
  include list named neither the `agents/` nor the `hooks/` surface directory,
  so the released bundle omitted both — the plugin installed without its agents
  and without its hook wiring. Both directories are added to the payload
  includes, and a bundle-completeness check now fails if any auto-discovered
  plugin-surface directory present on disk is neither included nor explicitly
  excluded with a reason, so a newly-added surface can never be silently dropped.
- **The history-store bootstrap error names the verb that exists.** When a
  transcript capture found the store's owned directories absent, the preflight
  error told the user to "run `abcd install`" — a verb that does not exist. The
  remediation now names `abcd ahoy install`, the verb that actually bootstraps
  the store (iss-58).
- **`abcd ahoy` no longer overclaims `managed-repo` for a stray `.abcd/`
  directory (iss-88).** Folder classification treated the mere presence of an
  `.abcd/` directory as a strong managed signal, so a repo with an unregistered,
  markerless `.abcd/` reported `managed-repo`. Only index registration or a
  marker block now promotes a folder to managed; a stray `.abcd/` reports
  `unmanaged-repo` (or `unmanaged-folder` outside a git repo).
- **The identity pin round-trips through the self-contained commit guard.** The
  pin was stored with Go's default JSON encoder, which escapes `&`, `<`, `>` (and
  always `"`, `\`, control characters); the pre-commit identity guard reads the
  raw bytes between the quotes with a naive parse, so an escaped value never
  matched `git config` and fail-closed a correctly configured identity (e.g. a
  `user.name` of `Marks & Spencer`). The pin is now marshalled without HTML
  escaping so `&`/`<`/`>` are stored literally, and the characters a parse can
  never read back (`"`, `\`, control) are refused at pin time with a clear remedy
  — keeping the commit guard zero-dependency rather than delegating it to a
  possibly-stale binary.
- **`abcd ahoy install` now honours an explicit config-value flag on an
  already-configured repo (apply-as-update).** Passing `--visibility`,
  `--docs-target`, `--oracle-backend`, or `--scan-deep` on a repo whose value
  was already set and valid was silently dropped: the persisted value
  short-circuited the install before the override was consulted. An
  explicitly-passed flag whose value differs from the persisted one now
  overwrites it, echoes the change (`changed: visibility: private -> public`),
  and — for visibility and docs-target — refreshes the `.gitignore` block and
  marker files so nothing is left inconsistent. A re-install with no such flag
  is still an exact no-op and never clobbers a valid value.
- **GL002 no longer fires a spurious blocker when a line has a closed inline-code
  span before a stray backtick.** `stripInlineCode` restored the entire original
  line whenever it reached end-of-line with an unpaired backtick, un-masking an
  enforced synonym that sat inside an earlier, correctly-closed span. It now
  blanks matched backtick pairs only and leaves a trailing unpaired backtick (and
  its tail) literal, so the earlier span stays masked (iss-106). Full CommonMark
  double-backtick span parsing remains out of scope.
- **`abcd intent "<text>"` no longer files a draft from a mistyped subcommand.**
  A near-miss for an intent subverb (`intent paln`, `intent lnk itd-5`) is
  refused with a did-you-mean and writes nothing, mirroring `abcd capture`'s
  guard; a genuine prose title still files. The shared typo heuristic is now
  record-id aware (`iss`/`itd`/`spc`), which also sharpens `abcd capture`.

## [0.3.0] - 2026-07-18

### Security

- **The redaction scanner no longer lets a secret survive a trailing underscore.**
  Nine hard-fail token patterns (GitHub `ghp_`/`ghs_`/`gho_`/`ghu_`/`ghr_` and
  fine-grained PATs, AWS `AKIA`, Stripe `sk_live_`/`sk_test_`) used a pure
  alphanumeric charset closed by a `\b` word boundary. Because `_` is itself a
  word character, a credential immediately followed by `_` (a JSON key, a
  concatenation, `token=ghp_..._old`) had no boundary and slipped through
  unredacted into the stored transcript. The trailing `\b` is dropped (matching
  the existing Google-key fix); the leading boundary, prefix, and minimum length
  keep the match precise. The same fix extends to Slack `xox` tokens (whose
  charset also excludes `_`) and the Anthropic/OpenAI `sk-ant-`/`sk-proj-`/
  `sk-svcacct-` keys (a minimum-length key ending in `-`), so no token pattern
  relies on a trailing word boundary.
- **Untrusted file reads are guarded against symlink-follow and unbounded
  reads.** Reads of content that can originate outside the local worktree — the
  sources-index registry, packed-lifeboat layer files, and the CLI JSON operands
  (`disembark coverage`, the memory `--pages-json`/`--page-json` transport, and
  the lesson/synthesis payloads) — now route through a single guarded primitive
  (`O_NOFOLLOW` + regular-file on the open fd + size cap, in one call). Previously
  some followed a symlink to its target's content or read an endless/oversized
  file unbounded, and a symlinked registry surfaced a raw, path-leaking error;
  all are refused consistently, closing a class of `lstat`→`read` swap windows.
- **The last three ahoy reads route through the guarded primitive, closing a
  residual read-time TOCTOU.** The hook-manifest check and the two `.gitignore`
  reads (one at the attacker-influenced working-directory boundary) each ran a
  separate `lstat`, regular-file, and size-cap check and then a distinct
  `os.ReadFile`, so a type or symlink swap in the window between the check and the
  read was not refused on the descriptor that was read. All three now read through
  the single guarded open (`O_NOFOLLOW` + regular-file on the open fd + size cap),
  and every structured signal — the manifest's reason strings and the
  `.gitignore` overwrite refusals — is preserved unchanged.
- **Machine (`--json`) output no longer leaks an absolute local path.** The
  error-only path scrub covered failures but not the success envelope, so
  `capture`, `capture resolve`/`wontfix`, and `capture list` echoed the absolute
  ledger path in their `path` field; a successful `memory ingest` recorded the
  absolute (and, for a `~/…` source, home-rooted) source path in its
  `citation.origin`; and a `memory ingest` failure on a source outside the
  working directory embedded an absolute source path the scrub could not reach.
  Every such locator is now rendered repo-relative — in machine output and in the
  persisted citation alike — so no verb emits a developer-identity path into
  machine output.

### Added

- **A one-line, checksum-verified installer in the README.** The command
  detects OS/architecture, downloads the binary and `checksums.txt` from the
  latest GitHub Release, verifies the binary's SHA-256 against the manifest
  fail-closed (a mismatch — or a binary the manifest does not list — refuses
  to install), and installs to `/usr/local/bin`. The README also documents
  the inspect-first manual equivalent.

## [0.2.0] - 2026-07-17

### Security

- **Two more git call sites are isolated — finishing the env-inheritance sweep.**
  An ultracode sweep-completeness pass found the two the earlier round missed.
  `identity.gitConfig` (the commit-identity gate) read `user.name`/`user.email`
  with the ambient environment, so an injected `GIT_CONFIG_*` could forge — or an
  inherited `GIT_DIR` redirect — the very identity the gate verifies; it now uses
  `gitutil.ScrubbedEnv` (keeping global config, like the identity probe).
  `capture.discoverRepoRoot` ran `git rev-parse --show-toplevel` with the ambient
  environment, so an inherited `GIT_WORK_TREE` redirected repo-root discovery — and
  thus where the issue ledger is read and written — at an attacker-chosen tree; it
  now uses `gitutil.IsolatedEnv`.
- **The embark/pack path guards every read of an arbitrary target repo.** Five
  reads on the lifeboat embark/pack path — `embark`'s target `CLAUDE.md`
  (`embarkMarker`), the target record compare (`classifyEmbark`), the coverage
  handoff (`readCoverageHandoff`), the pinned provenance (`readProvenance`), and the
  destination-gate provenance probe (`isAbcdLifeboat`) — used a plain `os.ReadFile`
  (three of them after a separate `Lstat`, a TOCTOU window; one checking the size
  only *after* reading it all). The source is an arbitrary, possibly hostile repo,
  so a symlinked or device/endless file could be followed or read unbounded. All
  now route through `fsutil.ReadGuarded` (`O_NOFOLLOW` + regular-file on the open fd
  + size cap), closing the swap window and the resource-exhaustion path in one call.
- **The identity probe and the ahoy git helper ignore inherited `GIT_DIR`/
  `GIT_WORK_TREE` and injected `GIT_CONFIG_*` — completing the `IsolatedEnv`
  sweep.** Two git call sites still ran with the ambient environment. Worse of the
  two: `scanner.ProbeIdentity` reads the caller's `user.name`/`user.email` to
  build the hard-fail identity-redaction matchers, so an injected
  `GIT_CONFIG_COUNT`/`GIT_CONFIG_KEY_*` (as a CI/agent sandbox can export) forged a
  fake identity and the caller's *real* name/email then sailed through the ship
  gate and the transcript sanitiser unredacted. It now runs with a new
  `gitutil.ScrubbedEnv` — the repo-selection and config-injection vars stripped but
  global config kept, since the caller's identity legitimately lives there and full
  isolation would blind the probe. `ahoy.runGit` (which derives the root-commit SHA
  and origin URL that key the cross-repo history registry) now uses the fully
  isolated `gitutil.IsolatedEnv`, so an inherited `GIT_DIR` can no longer register
  one repo's transcripts under another's immutable key.
- **The secret scanner detects a Google API key whose 35th character is `-`.** The
  `\bAIza…{35}\b` pattern is fixed-length and its class includes `-`, so a key
  ending in `-` had no shorter match to satisfy the trailing ASCII `\b` and the
  `hard_fail` secret slipped both the launch gate and the transcript sanitiser. The
  trailing `\b` is dropped (the `AIza` prefix and fixed length still bound it).
- **`home_path_self` redaction is case-insensitive.** The overlapping-secret and
  Unicode-boundary hunt added `(?i)` to the email/name/github identity matchers but
  left the caller's own home path case-sensitive, so on a case-folding filesystem a
  differently-cased spelling of `$HOME` — the same directory on disk — escaped the
  hard-fail `home_path_self` gate. It now folds case like its siblings.
- **Untrusted repo content is sanitised on every terminal render path, not just
  the lifeboat report.** The escape/C1 sanitiser that guarded `disembark`'s human
  report is now a shared `internal/termsafe` primitive, and the other render paths
  that printed repository-derived text raw — `abcd audit`, `docs lint`, `capture
  list` skip rows, the `intent`/`spec` boards, `memory` (bare + `ask`) — route
  through it. A crafted commit subject, file path, error string, or memory-page
  summary can no longer inject an ANSI escape to recolour or corrupt the report.
  The primitive also now masks the bidirectional-override and zero-width
  ("Trojan Source") classes, so untrusted text cannot visually reorder or hide
  characters so the rendered line differs from the bytes. JSON output is unaffected.
- **The identity pin is written atomically.** `identity.WritePin` persisted
  `.abcd/config/identity.json` with a plain in-place `os.WriteFile` — the fifth
  writer the iss-32 atomic-write consolidation missed (the guard flags only
  divergent atomic primitives, not a non-atomic one). It truncated the pin before
  rewriting, so a crash mid-write left a corrupt or empty gate config, and it
  followed a symlink at the path. It now routes through the canonical
  `fsutil.WriteFileAtomic` (temp + fchmod + fsync + rename + parent fsync).

- **The secret scanner now detects GitHub fine-grained PATs (`github_pat_…`) and
  PEM private-key headers.** Neither was in the bundled pattern set, so a
  current-generation GitHub token (GitHub's default since 2022) or a committed
  private key passed both the `abcd launch` ship gate and the history-transcript
  sanitiser unflagged.
- **Write-time transcript redaction no longer leaks raw secret bytes from
  overlapping matches.** `scanner.Redact` masked by longest-first substring
  replacement, so two partially-overlapping secret spans (e.g. an `sk-ant-` key
  running into a JWT) left the shorter token's tail verbatim, and the fail-closed
  re-scan could not catch the now-truncated remainder. Redaction is now by
  authoritative byte span (overlap bytes forced to `*`), matching the serializer.
- **History capture fails closed when the per-repo `pii.json` is broken.**
  `Scanner.ScanText`/`Redact` could not signal the degraded (`unavailable`) state
  the way `ScanBundle` does, so a malformed config silently redacted transcripts
  with a weakened pattern set and still reported success. A new `Unavailable()`
  accessor is now consulted before capture.
- **A per-repo config can no longer neuter a bundled detector by replacing its
  regex.** The severity floor clamped an override's severity but not its regex, so
  swapping a bundled pattern's regex for a never-match one disabled detection at
  full `hard_fail` severity. Bundled regexes are now immutable (a config may only
  raise severity / adjust label; new detection must use a new pattern name), and
  the config merge is all-or-nothing.
- **The launch bundle no longer ships denied namespaces or gitignored files
  reached through a symlink.** A symlink to (or into) the repo root let a
  dereferenced walk descend into `.git/**` and `.abcd/**`, and a symlink whose
  target was gitignored shipped the ignored content under the symlink's benign
  name. The structural deny is now re-applied to real paths at every level of a
  dereferenced walk, and the ignore probe covers the symlink target.
- **Command-error output no longer leaks a path equal to `$HOME` itself.** The
  redactor only rewrote paths *under* a root, so a message or `PathError` naming
  exactly the home directory slipped through — and its base segment is the
  username. `record-lint` likewise printed raw `*os.PathError` config-load paths;
  both now scrub the home/root prefix.
- **The launch bundle's gitignore exclusion fails closed on a git failure.** It
  delegated to a fail-open probe, so if git errored the exclusion pass admitted
  every gitignored file (typically `.env`-style secrets). A launch-local strict
  probe now distinguishes "nothing ignored" from a real git failure and rejects
  the affected candidates on failure; a plain non-git directory still resolves.
- **Git queries ignore inherited `GIT_DIR`/`GIT_WORK_TREE`/`GIT_INDEX_FILE` and
  config-injection env vars.** An inherited repo-selection variable could redirect
  abcd's isolated git queries at a different repository; those variables are now
  stripped before every invocation.
- **Transcript snippets no longer leak the head/tail of a short multi-byte
  identity value.** `sealLine` measured its fingerprint threshold in bytes while
  `maskSecret` used runes, so a short non-ASCII identity value kept visible
  characters; the two are now both rune-based.
- **`WriteFileAtomic` sets permissions on the open descriptor** (`fchmod`) rather
  than by name after close, closing a TOCTOU symlink-swap window; and
  `WriteFileAtomicPreserveMode` no longer silently widens an existing file to
  `0644` on a transient stat error (it fails closed).
- **The PII scanner detects real names, emails, and third-party home paths that
  RE2's ASCII `\b` was silently skipping.** A `hard_fail` `real_name` whose first
  or last character is non-ASCII (accented, CJK, Cyrillic) never matched — the
  name shipped unredacted; the `real_email`/`github_username` matchers were
  case-sensitive, so a case variant of the caller's own address slipped the gate;
  and `home_path_other` never fired at a realistic boundary (line start, after a
  space or `=`), so third-party home paths published verbatim. All three now use
  Unicode-aware, case-folding boundary predicates. Relatedly, a warn-level
  `home_path_other` span no longer suppresses a `hard_fail` `local_username`
  finding underneath it (which downgraded a username leak out of the ship gate).
- **The privacy-hygiene audit flags a bare home path with no trailing separator.**
  `absPathRe` required a path component *after* the username, so the leak itself —
  `HOME=/home/alice` at end of line — was never caught. The trailing separator is <!-- abcd-audit:allow -->
  now optional (matching the Windows branch).
- **The release-bundle gitignore probe ignores inherited `GIT_DIR`/`GIT_WORK_TREE`
  and config-injection env vars.** The strict probe appended to `os.Environ()`, so
  an inherited repo-selection variable could redirect it at a different repository
  and make a gitignored secret read as "not ignored" — promoting it into the
  release. It now runs with the same scrubbed environment as every other isolated
  git call (also applied to release-tag retention).
- **Terminal-report sanitisation strips the C1 control range (`0x80`–`0x9F`).**
  It masked C0 controls and DEL but let U+009B (CSI) through — an 8-bit terminal
  acts on it exactly like `ESC[`, reopening the escape-injection path from
  untrusted commit subjects/refs that masking `ESC` alone was meant to close.
- **The lifeboat pack overlap gate is case-insensitive on case-folding
  filesystems.** On macOS's default filesystem a differently-cased destination
  inside the source (`.../REPO/lifeboat` vs source `.../repo`) computed as an
  out-of-tree sibling and slipped the gate, so the pack wrote into the source
  tree.
- **The graveyard probe rejects option-like git refs from a hostile repo.** The
  lifeboat probe (whose stated threat model is hostile/archived repositories) fed
  branch names and an `origin/HEAD`-derived default branch straight to `git
  merge-base`/`branch`/`rev-list` as positional args; a crafted ref such as `-x`
  (written into `.git/refs`) parsed as a flag. Repo-derived refs beginning with
  `-` are now refused before reaching git — closing the argument-injection vector
  before a future richer subcommand makes it exploitable.
- **Memory-store reads are guarded against symlink and size attacks.** The sources
  registry (`.sources_index.json`) and each memory page were read with a bare
  `os.ReadFile` — no size cap, following symlinks — inside the repo working tree
  (a trust boundary) on every `abcd memory` verb, some under the store lock. A
  committed symlink to `/dev/zero` could OOM or hang the CLI. Both now route
  through a shared `fsutil.ReadGuarded` (`O_NOFOLLOW` + regular-file check on the
  open fd + byte cap).

### Fixed

- **`abcd install` no longer deletes a user's `.gitignore` content after an
  orphan `# BEGIN ABCD` fence.** An unbalanced BEGIN with no matching END made
  the rewriter drop every line to end-of-file, taking the user's own ignore rules
  with it. An unmatched BEGIN is now dropped alone; the content after it is
  preserved (mirroring the stray-END policy).
- **`git check-ignore` no longer inverts the answer for force-added files.**
  Dropping `--no-index` means a tracked file that matches a `.gitignore` pattern
  is correctly reported not-ignored, so the privacy audit stops flagging a
  force-added `DECISIONS.md` and the bundler stops dropping force-added files.
- **Recall keyword matching now handles inflected forms.** The stemmer could not
  round-trip e-drop (`merging`→`merge`) or doubled-consonant (`committing`→
  `commit`) inflections, and multi-word aliases bypassed stemming entirely, so
  guardrail domains like `COMMITTING` silently failed to match common phrasings.
- **Frontmatter and intent records agree on delimiter handling.** An unclosed
  frontmatter block is now treated as no-frontmatter (it previously harvested body
  prose as top-level fields), and the intent writer tolerates a trailing-space
  `---` delimiter exactly as the reader does.
- **YAML frontmatter flow-map keys are quoted.** A citation key containing a YAML
  metacharacter previously corrupted the record or broke the write→read round-trip.
- **Concurrent `abcd intent plan` runs mint distinct spec ids** (an exclusive
  advisory lock now guards the id-scan-and-write), and history capture attributes
  a byte-identical transcript from a distinct session to its own record instead of
  the first session's.
- **Several confident-but-wrong diagnostics are now guarded:** `abcd audit
  --root <missing>` and `disembark coverage <non-report>` return a usage error
  instead of fabricated findings; the privacy-hygiene rule surfaces an error
  rather than reporting clean when it cannot read the repo; brittle-reference
  linting skips fenced code blocks; and Cobra usage errors exit `2`, not `1`
  (which `abcd audit`'s tri-state reserves for "warnings only").
- **Robustness:** a malformed glob in the include config is a preflight error
  instead of a panic; an overflowing manifest pointer index is rejected instead
  of panicking; the intent identity gate compares the author git will actually
  stamp (honouring `GIT_AUTHOR_*`); inline-list items with quoted commas round-trip
  faithfully; and the lifeboat probe's tier gate matches what its adapters read.
- **The modular-rules loader resolves `.abcd/rules.json` from the repo root, not
  the current directory.** Run from any subdirectory, `abcd` looked up the
  per-repo overrides — and the kill switch — under the subdirectory, found
  nothing, and silently injected the default ruleset a repo had disabled. The
  loader now walks up to the nearest `.abcd` directory.
- **`abcd install` no longer rebuilds a malformed `.abcd/config.json` from
  scratch,** which destroyed whatever the user had. A JSON parse error is now
  respected: the file is left untouched and the install reports partial.
- **The issue-ledger reader rejects malformed records it used to accept
  silently:** a duplicate top-level (or nested) frontmatter key — where the reader
  kept the last value but a status transition rewrote the first — and a
  non-string `resolved_by` sub-value that validated clean then dropped to `""` on
  read.
- **The disembark voyage ledger logs SHA-256-format repositories.** Its root-SHA
  key accepted only a 40-char SHA-1; a 64-char SHA-256 root was rejected and the
  pack silently went unlogged.
- **The history transcript store accepts a SHA-256 root key too.** The same
  40-char-only assumption in `history.store`'s `rootSHARe` (a sibling of the voyage
  key above) made `history capture`/`list`/`read` all fail for a repo in git's
  SHA-256 object format — the ahoy layer derives the 64-char root SHA, but history
  refused it, so no session was ever stored. The key now accepts 40 or 64 hex.
- **Spec id minting cannot wrap to a negative id.** `specNum` discarded the
  `strconv.Atoi` overflow error, keeping the clamped `MaxInt64`, so an over-int64
  spec number made `NextID` compute `max+1` and mint `spc--9223372036854775808`. An
  unparseable/over-range number is now treated as no reservation.
- **Release-tag retention ignores prerelease/build tags.** `Tag()` renders the
  core `MAJOR.MINOR.PATCH` only, so a real tag `v1.2.3-rc1` surfaced in the plan
  as a phantom `v1.2.3` and collapsed against the real release; prerelease/build
  tags are now excluded (retention operates on release cores).
- **Distinct deleted paths key distinct graveyard findings.** The id cleaner
  *deleted* spaces and control characters, so two paths differing only in
  whitespace collided onto one finding id and one shadowed the other; the
  transform is now injective (percent-encoding), leaving ordinary paths unchanged.
- **The lifeboat pack destination gate treats an `ENOTDIR` stat as "absent"**
  (a prefix component being a file) rather than an uninterpretable error that
  refused a writable destination.
- **A relative PATH symlink is resolved against the symlink's own directory,**
  not the process working directory — so `abcd ahoy` no longer reports a bogus
  "foreign symlink" gap (and uninstall no longer refuses to remove a link it
  owns) for a correct relative install such as `/usr/local/bin/abcd ->
  ../lib/abcd/abcd` when run from another directory.
- **`source.classes` on a memory page is validated as a set, not an ordered
  list.** The same classes declared in a different order from their first
  appearance in `sources[]` were rejected, contradicting the schema's (and the
  error message's) set semantics.

### Changed

- **The coverage report is now schema v2.** Each brief section carries a `kind`
  (`extractable` — a source or a better adapter could ground it, so a blank is
  coverage debt — versus `human-owned` — a question only a person can answer, so a
  blank is not a failure), the durable form of the M2 cross-repo gate decision
  (adr-36). A blank additionally carries a `resolution` (`open`/`answered`/
  `deferred`) and, once answered, an authored `answer` whose provenance is a
  person and a date rather than a file it did not come from. `abcd disembark
  coverage` still refuses a report from a newer schema with an upgrade message.

### Added

- **`abcd intent "<text>"` files a new draft from quoted text — a symmetric
  create path.** Typing `abcd intent "I want users to feel X"` mints the next
  `itd-N` under the intent-store lock and writes
  `.abcd/development/intents/drafts/itd-N-<slug>.md`, seeded from the text with the
  canonical draft frontmatter and a minimal, lint-valid body skeleton — no `new`
  sub-verb required, mirroring `abcd capture "<text>"`. The old `abcd intent new
  "<text>"` still works as a backwards-compatible alias but prints a deprecation
  warning on stderr naming the quoted-text shape. Bare `abcd intent` stays
  read-only status + help and mutates nothing. Both ledgers' bare-form help now
  carry a one-line decision rule (nitpick/observation -> capture; user-facing
  change to ship -> intent). The `/abcd:capture promote` flow hands the issue text
  to this create path (itd-46).
- **`GL002` — a glossary-driven forbidden-synonym gate for the record lint.**
  The lint now reads each glossary term's `forbidden_synonyms` and flags an
  *enforced* synonym used as a standalone word in live prose, so terminology
  drift is caught by a detector instead of by eye (itd-43). Enforcement is a
  deliberate subset (`epic` first): most forbidden synonyms are common English
  words whose false-positive rate would sink the gate, and each enforced word
  must be one the glossary actually forbids. Matching uses explicit Unicode word
  boundaries — not the ASCII-only regexp `\b` — and skips code spans, YAML
  frontmatter, dated/historical records, and the glossary term files themselves.
- **Bare `abcd ahoy` now names the next step for the folder it classified.** An
  unmanaged git repo report points at `/abcd:ahoy install` as the way to adopt
  it, and a plain (non-git) folder report states there is nothing to act on —
  the read-only classification never mutates either (itd-40).
- **Synthesis over the record — `abcd disembark principles`, `press-release`,
  and `oracle`.** Three post-pack verbs interpret a packed lifeboat, each in
  one of two self-recorded modes. Without a payload they run **deterministic
  mode**: principles distilled evidence-only from the packed ADRs'
  Decision/Consequences bullets, the press release composed from the brief's
  own page (or the spine, or an honest placeholder), and the oracle scoring
  mechanically — a failed manifest verification is a `MAJOR_RETHINK` verdict,
  not an error; more blanks than grounded sections is `NEEDS_WORK`; a healthy,
  verified lifeboat ships `SHIP` — the first code home of abcd's registered
  review-verdict vocabulary. With `--*-json <file|->` they ingest a
  host-delegated agent's output behind the same trust guards as an intent
  verdict, under cite-or-be-dropped (a principle or oracle finding citing no
  live record id, graveyard finding, or packed path is dropped and reported; a
  press release citing nothing resolvable is refused whole). The binary stamps
  the oracle's attestation fields itself, so a model cannot fabricate a
  manifest hash. All synthesis artifacts live outside `manifest_sha256`, are
  fully replaced per run, and carry no wall-clock — the audit is keyed by the
  lifeboat's own manifest hash. The four agents (`principle-distiller`,
  `graveyard-interpreter`, `press-release-composer`, `lifeboat-oracle`) ship
  under `agents/` with itd-5's prompt discipline: versioned prompts in the 0.x
  calibration band, `reads_untrusted_input` declared, and an injection-canary
  fixture each (itd-88, adr-35).

- **`abcd embark` — a lifeboat comes ashore.** `embark probe <lifeboat> [target]`
  is the read-only reconciliation: it refuses a lifeboat whose provenance schema
  is newer than the binary (with an upgrade message), re-hashes every archived
  file against the pinned `manifest_sha256` so a tampered or truncated lifeboat
  is caught before anything is read in anger, and reports — in one bulk report,
  not a per-file barrage — every conflict with the target repository. `embark
  from <lifeboat> [target]` is the write path: it refuses entirely on any
  conflict (identical bytes are an idempotent skip, and a re-embark is a no-op),
  writes only the record families (ADRs, issues, intents, specs) through
  `os.Root` containment plus independent path validation — two layers, so a bug
  in one is not an escape — and never copies lifeboat prose into `CLAUDE.md`:
  it re-injects the *current* abcd marker block instead. The rendered result
  leads with the coverage report's blanks and their questions — the handoff to
  the human who must answer them. The packer now also carries the spec store
  (`rescue/specs/`), and every lifeboat's provenance records
  `record_manifest_sha256`, the seal over exactly the record-derived families
  that must survive a round-trip byte-for-byte: pack → embark → re-pack
  reproduces it, and embarking into a byte-copy of the source reproduces the
  full original manifest hash (itd-88, adr-35; closure re-scope in the
  2026-07-16 decision log).

- **The graveyard — what the project abandoned, in three strictly-ordered
  layers.** Every packed lifeboat now carries `graveyard/archaeology.json`
  (deterministic git archaeology: reverted commits, branches never merged into
  the default branch ranked by divergence age, paths deleted after substantial
  history, dependencies adopted then dropped, wholesale-rewrite commits — pure
  evidence, no interpretation, from any git repo) and `graveyard/abandoned.json`
  (what the record itself declared dead: superseded intents and ADRs, wontfix
  issues with their reasons, each ADR's Alternatives-Considered options, and
  rejected options named in the decision log). A new
  `abcd disembark graveyard <lifeboat-dir> --lessons-json <file|->` verb ingests
  a host-delegated interpretation over those two layers into
  `graveyard/lessons.json` under a **cite-or-be-dropped** validator: every
  lesson must cite live layer-1/2 evidence ids or it is dropped (reported, never
  fatal), low-confidence lessons are quarantined under
  `graveyard/low-confidence/` instead of the main file, and the untrusted
  payload is read behind the same trust guards as an intent verdict (size cap,
  no symlinks, unknown fields refused, schema version gated). Each ingest fully
  replaces the prior interpretation, so a promoted or later-dropped lesson
  leaves nothing stale behind. The validator — not the model's good intentions
  — is the difference between a graveyard and a séance (itd-88, adr-35).

- **`abcd disembark probe <repo>` — a read-only coverage probe over any
  repository.** It walks a repo without touching it and reports, per brief
  section, whether a lifeboat could ground it: `grounded` / `partial` / `blank`,
  with the tier it was grounded from, a confidence, and the evidence cited. A
  blank is a first-class result — it carries what abcd searched and the question
  a human must answer, so the report is a to-do list, not a shrug. Adapters
  degrade across three tiers — git (any repo), conventions (README, docs,
  CHANGELOG, manifests, ADRs wherever they live), and abcd-native (`.abcd/`) — so
  a richer repo grounds more, and the `graveyard` section grounds from git
  history alone (reverts, deleted files, dependency churn). Every read is
  contained to the repo (`os.Root`), bounded, and non-blocking, and the source
  tree is byte-identical afterwards — the probe never writes to a source. A
  companion `abcd disembark coverage <report.json>...` reduces several probe
  reports to one section×repo table with an always-blank verdict per section:
  the delta between a record-rich repo and a git-only one is what keeping a
  record is worth, legible as a number. Both are read-only operator verbs (no
  `/abcd:disembark` command surface yet); the packer that writes a lifeboat is a
  later milestone (itd-88, adr-35). The dependency-manifest detector spans Go,
  Node, Rust, Python (pip/poetry/pdm/uv/pipenv), Ruby, and PHP, so a real project
  is not reported as having no dependencies merely because the probe did not know
  its packaging tool (found probing a Python/uv repo in the M2 cross-repo run).

- **`abcd disembark plan <repo>` — a dry run of the packer.** It shows the
  complete file set a lifeboat pack would write — brief citation maps for the
  grounded sections, `coverage.json`/`coverage.md`, verbatim copies of the ADRs
  and the issue ledger, the rescue spine (the intent corpus where one exists, a
  git-derived summary where it does not), and a `_provenance.json` carrying a
  pinned `manifest_sha256` over every other file — and writes **nothing**. Plan
  and the eventual packer are one code path, so the dry run cannot describe a pack
  a real pack would not perform; a re-plan of an unchanged source is byte-for-byte
  identical (the manifest carries no timestamp). `--json` emits the manifest
  (paths, sizes, and the hash — never file content). Still a read-only operator
  verb (no `/abcd:disembark` command surface yet); the destination write path is
  a later milestone (itd-88, adr-35).

- **`abcd disembark pack <repo> <dest>` — writes a lifeboat out-of-tree, and the
  `/abcd:disembark` command surface.** It writes the planned file set to `<dest>`
  and never to the source (a test hashes the source tree before and after).
  Everything that stops a pack destroying real work is enforced: a **destination
  safety gate** refuses unless `<dest>` is absent, an empty directory, or an
  existing lifeboat abcd produced (it carries a parseable `_provenance.json`) —
  and refuses a symlinked destination, one inside a `.git/` directory, or one that
  overlaps the source tree. The planned bytes are **secret-scanned before any
  write** and a hard-fail refuses the whole pack — a secret is fixed at source,
  never redacted into the artefact. Files are written into a staging directory
  through `os.Root` (no crafted path or symlink escapes it) and renamed into
  place, so a crash leaves staging, never a half-lifeboat; `_provenance.json` is
  written last. Any abcd marker block in a copied record is stripped so embarking
  the lifeboat cannot plant a stale rules-loader. Each pack appends one line to an
  append-only voyage ledger at `~/.abcd/voyage/<source-root-sha>/disembark/history.jsonl`,
  keyed on the source's root-commit SHA and carrying the manifest hash. `--json`
  emits the result (destination, file/byte counts, hash, voyage status).

- **Session transcripts are captured automatically when a session ends.** A
  `SessionEnd` hook now runs `abcd hook session-end`, which redacts the session
  transcript through the existing two-stage, fail-closed scanner and files it in
  the local per-repo store — no flag to pass, no command to type. `abcd history
  list` shows the records. It is wired to `SessionEnd` (which fires once when a
  session terminates) rather than `Stop` (which fires once per assistant turn) —
  the transcript grows through a session, so a `Stop`-wired capture would store a
  fresh, larger superset every turn. The store has existed since the native transcript
  corpus landed (adr-29) and was **called by nothing**: `history.Capture` was
  built, correct, and unused, so no session was ever stored. That gap was the one
  cost on the board that could not be recovered later — a session that ends
  without being captured cannot be reconstructed by any amount of future work.
  The hook is operator-internal, never blocks the host, and always exits `0`: a
  malformed payload, a missing or non-regular `transcript_path`, a hostile
  session id, or a directory that is not a git repo each capture nothing, say why
  on stderr, and exit cleanly — a `Stop` hook that errors or hangs would wedge
  the session, which is strictly worse than a missed transcript. Re-capture is
  idempotent (a `Stop` hook may fire more than once per session), the transcript
  open is non-blocking so a FIFO cannot hang the hook, and nothing is ever
  written to stdout. It needed a new verb because `history capture` cannot be
  wired to a `Stop` hook: from stdin it *requires* `--session <id>`, and a `Stop`
  hook delivers its session id inside a JSON payload (itd-89, adr-29).

- **A session that starts in a repo where abcd is not installed now says so.** A
  `SessionStart` hook runs `abcd hook session-start`, which — when the current
  repo is a git repository whose transcript store has not been bootstrapped —
  prints a one-line notice telling the user their sessions will not be captured
  and how to fix it (`abcd ahoy install`). Without it the automatic-capture hook
  above fails silently: the plugin is enabled, the user assumes their transcript
  corpus is accruing, and it is not. The notice rides `SessionStart`'s visible
  channel (stderr on a non-zero exit) and never blocks the session; every case
  that is *not* a bootstrappable-store problem — a non-git cwd, a malformed or
  empty payload, an already-installed store — stays completely silent and exits
  `0` (iss-95, itd-89).

- **`abcd audit` — a read-only repo-conformance check.** One command reports
  whether a repository follows the working conventions: the three-tier `.abcd/`
  layout, an `AGENTS.md` router, decisions durable in a committed
  `.abcd/work/DECISIONS.md`, docs currency (reusing the docs-lint engine where
  `docs/` exists), and privacy hygiene (no absolute local paths in committed
  files, waivable per line with `abcd-audit:allow`). It runs against any repo
  given only a working directory, prints a grouped human report with a fix per
  gap or machine JSON (`--json`, stable rule ids, `{ "findings": [] }` when
  clean), and exits with a tri-state code — `0` clean, `1` warnings only, `2`
  any error — so it gates CI as well as onboarding. It answers a different
  question from `abcd ahoy doctor`: `doctor` is tool-setup health, `audit` is
  repo conformance. `/abcd:prepare-this-repo` now runs `abcd audit` for its
  Phase 2 gap report instead of hand-auditing (itd-85, iss-86).

### Fixed

- **`--json` and stderr command errors no longer leak the developer's home or
  working-directory paths.** `cli.Run` routes every command error through the
  machine envelope, so identity-bearing paths reached it three ways: an
  `os.PathError`/`os.LinkError` (e.g. `memory ask --page-json` on a missing file),
  a path `fmt`-formatted into a core error (e.g. `capture` on a symlinked ledger
  dir), and a custom error type (e.g. history's home-rooted store path). The
  `Run()` boundary now redacts the working-directory and home roots (to `.` and
  `~`) and reduces any remaining `PathError`/`LinkError` path to its base name.
  Generalises the per-branch fix made in `iss-29` (iss-76). A verb echoing a
  user-supplied absolute path outside both roots is out of scope, tracked in
  `iss-81`.
- The `intent_lifecycle` record-lint rule now **blocks duplicate intent ids**.
  Id allocators are branch-local — parallel agents on separate branches each
  scan for `max + 1` and mint the same id — so two intents both claimed
  `itd-82` and both merged with every gate green. The rule flags *every* file in
  a colliding set, not just one: the linter cannot know which claimant is
  authoritative, and flagging a single file would imply the others are fine. The
  collision itself is resolved (the later claimant renumbered to `itd-83`); the
  underlying minting scheme is tracked as `iss-80`.
- `memory ingest --keep-original` writes the stored source copy through the
  canonical `fsutil.WriteFileAtomic` (temp + fsync + **chmod + parent-directory
  fsync**) instead of an inline temp+rename that omitted both — the fifth
  divergent atomic write the `iss-32` consolidation left untouched. The
  one-canonical-primitive detector now also flags inline `os.O_EXCL`+`os.Rename`
  sequences, not just named primitives (iss-79).

### Added

- **Four reviewer agents ship with the plugin**: `abcd:ruthless-reviewer`
  (correctness, resource handling, error paths, dead code),
  `abcd:security-reviewer` (adversarial review of a trust boundary),
  `abcd:docs-currency-reviewer` (every user-facing claim verified against the
  code), and `abcd:sota-researcher` (evidence-tiered state-of-the-art research).
  Every repo with the abcd plugin enabled gets the same review bar, versioned in
  the repo rather than in a per-machine harness config. Each renders a binary
  verdict, and every finding it emits must carry a concrete failure scenario —
  the LLM-judge calibration discipline (itd-81).
- **Intent-fidelity review** (itd-80): the ship move now emits a report-only
  fidelity-review receipt, and `abcd intent review ingest --verdict-json <path>`
  applies the host-produced verdict back onto the record. When `abcd spec close`
  ships a linked intent (`planned/ → shipped/`), it parks a deterministic OWED
  receipt marker in the intent's `## Audit Notes` and writes an ephemeral review
  request under `.abcd/.work.local/reviews/` (gitignored); the emit is
  non-fatal, so a failure never un-ships the intent. `abcd intent review ingest`
  validates an untrusted intent-fidelity verdict JSON fail-closed (schema,
  in-enum verdicts, cited evidence, and each `criterion_id` bound to an actual
  Acceptance-Criteria bullet), then either replaces the OWED stub with the
  rendered per-criterion verdicts and honoured/diverged/missing audit
  (`INGESTED`, idempotent — a re-ingest is a no-op) or quarantines a bad payload
  (`DEAD_LETTER`: all criteria `INCONCLUSIVE`, raw payload retained) — never a
  partial application. Bare `abcd intent review <itd-N>` re-emits a shipped
  intent's request. The single source of truth is the intent file's Audit Notes;
  there is no side receipt store.

- The **intent lifecycle** verbs `abcd intent` and `abcd spec` (itd-80), the
  front doors onto the native intent store (`internal/core/intent`). Bare
  `abcd intent` renders a read-only lifecycle summary (intent counts by bucket,
  spec counts by status, and the linked intent↔spec pairs); `abcd intent plan
  <itd-N>` mints a native spec for a draft intent that carries a non-empty
  `## Acceptance Criteria` section (the itd-1 gate), writes both sides of the
  bidirectional link (the spec's `intent: itd-N` and the intent's
  `spec_id: spc-N` plus a default `kind: standalone`), and moves the intent
  `drafts/ → planned/` — fail-closed, so every intermediate on-disk state stays
  valid under the `intent_lifecycle` record-lint rule. `abcd intent link <itd-N>
  <spc-N>` retroactively links a planned intent to an existing spec, refusing a
  spec that realises a different intent. Bare `abcd spec` renders the spec-store
  status; `abcd spec close <spc-N>` moves a spec `open/ → closed/` (the
  lifecycle reconcile that trails a close lands in a later phase). The
  frontmatter line-scanner shared by these stores now lives in
  `internal/core/frontmatter`.
- The **modular rules loader** core and its `abcd rules [domain]` verb (itd-3,
  phases 1 + 3). `internal/core/rules` holds binary-bundled default rule domains
  (COMMITTING, DOCUMENTATION, ROADMAP, ISSUES, INTENTS, LIFEBOAT, PII, and
  OPINIONS — whose rules point at the canonical conventions under
  `.abcd/development/principles/` rather than copying them) merged
  with an optional per-repo `.abcd/rules.json` override (per-field domain
  override, sticky kill switch), with word-bounded recall matching (including a
  conservative suffix stemmer so `commits`/`issues` recall their keyword),
  `*<DOMAIN>` star-commands, and per-domain dedup signatures. Bare `abcd rules` renders the
  active rule set; a positional `DOMAIN` (case-insensitive) scopes to one; a
  malformed `rules.json` fails closed. A Claude Code prompt-router hook
  (`abcd hook prompt-router` / `prompt-router-reset`, operator-internal) injects
  the matched rules just-in-time on `UserPromptSubmit` with per-session
  signature dedup, clears the ledger on a `SessionStart`/`PreCompact` reset
  (event-driven refresh; a large fixed-N counter is only a backstop), and is
  fail-closed and non-blocking — a malformed payload, unreadable `rules.json`,
  or state error injects nothing and logs out-of-band, never wedging a session.
  The `hooks/hooks.json` manifest wiring lands with ahoy in the next phase.
- A `surface_coverage` record-lint rule (iss-35): the deterministic half of the
  brief↔surface cross-check. It reads the plugin surface
  (`rules.surface_coverage.commands_dir`, `skills_dir` — outside the lint roots)
  and the brief's surface registry table (`rules.surface_coverage.registry`, by
  convention `.abcd/development/brief/04-surfaces/README.md`), and asserts three
  invariants: every real surface has a registry row; every row marked `shipped`
  in the registry's **Status** column has a backing surface while every `staged`
  row (a design target) has none; and every row's status is `shipped` or
  `staged`. The bare `/abcd` top-level is binary-backed and exempt from the file
  check. Chapter-link resolution stays with `links_resolve`; the semantic half —
  each row's prose vs. binary behaviour — stays a release-gate agent check.
- A managed-repo **git-identity gate** (iss-62): a repo can pin its expected
  commit identity in `.abcd/config/identity.json`, and every commit is checked
  against it. `ahoy doctor` reports a divergence (a repo-local override that
  differs from the pin, or an unset identity) or an un-pinned repo; `ahoy
  install` adopts the gate by pinning the current git identity; `ahoy
  identity-check` exits non-zero on a mismatch; and the `pre-commit` hook
  fail-closes so a stray identity (e.g. a sandbox default) is caught at commit
  time rather than discovered later. A repo with no pin is unaffected.
- A `context_status_free` record-lint rule: the shared orientation file
  (`rules.context_status_free.target`, by convention `.abcd/work/CONTEXT.md`)
  must carry no phase/status claims — status is read live from the CLI and
  the ledger, never hand-written into orientation docs. Patterns are
  configurable (`rules.context_status_free.patterns`) with sensible defaults;
  lines matching inside fenced code blocks are skipped.

- A `/abcd:prepare-this-repo` command — audits the current repository against
  the abcd record and adopts the three-tier `.abcd/` layout, a marked
  working-conventions section in `AGENTS.md`, and the commit gates; an interim
  bridge until repos are managed directly. Owned repos only (it refuses
  elsewhere), and it migrates the older root-level `.work/` scaffold layout
  with explicit sign-off.
- `/abcd:consult` and `/abcd:ingest` commands — consult the user-level sources
  corpus (confidential entries are never cited or named in public artifacts)
  and ingest a URL or document into it with extracted reference metadata,
  keywords, and a text-quality check. Both are thin fronts on the corpus's own
  tooling and stop gracefully when no corpus exists.
- A `persona_registry` record-lint rule: press-release quote attributions
  (`said <Name>,`) must name a persona from the registry file the rule's
  `registry` key points at; unknown names are blocker findings. Configured
  per repo in `record-lint.json`; the historical record is skipped via the
  standard content-drift exemptions.
- `abcd capture --blocked-by <iss-N,…>` records typed dependency edges on a new
  issue, and `capture list` / the status board now render a derived-priority
  view: unblocked issues first, then by severity, with blocked rows annotated
  `[blocked-by iss-N,…]`. There is no stored priority — the ordering is a
  read-time projection, so resolving a blocker re-prioritises its dependents
  automatically.
- A store-contract README for the issue ledger (`.abcd/work/issues/README.md`).

### Changed

- `abcd intent plan` seeds a new native spec with a clear author-guidance
  placeholder in its `## Summary`, rather than a bare `TODO` (iss-68).
- `abcd spec close <spc-N>` now reconciles the linked intent (itd-80): it moves
  the intent `planned/ → shipped/` and then closes the spec, so one command
  completes the lifecycle transition. It is fail-closed (a missing/empty intent
  link, a non-existent or ambiguously-linked intent, bidirectional drift, or an
  intent in an unexpected bucket refuses with no partial move) and idempotent (a
  re-run on an already-shipped intent / already-closed spec is a clean no-op).
  The intent's `## Audit Notes` are left untouched. A new `spec_lifecycle`
  record-lint rule mirrors `intent_lifecycle` on the spec side: every spec under
  `specs/{open,closed}/` must carry a well-formed `id`/`slug`/`intent` link whose
  named intent EXISTS and points back at this spec (bidirectional agreement).
- The issue ledger moved from `.abcd/development/activity/issues` to
  `.abcd/work/issues` (the committed shared-working tier).
- The atomic-write and real-directory primitives are consolidated onto
  `internal/fsutil` (iss-32): the ahoy, capture, and memory store writers no
  longer keep their own divergent temp-file+rename copies. Two observable
  effects of routing through the canonical primitive: the ahoy and capture
  writers now fsync the parent directory after the rename (a crash-durability
  strengthening they previously lacked), and memory pages are written at a
  fixed `0644` (an explicit chmod, where the old writer left the mode subject to
  the process umask). A `TestNoNonCanonicalAtomicWritePrimitives` guard keeps a
  fifth copy from reappearing.

### Removed

- The `created` and `updated` frontmatter fields on issues. Git is the canonical
  source of an issue's timeline; the ledger no longer duplicates it.

### Fixed

- **Launch dogfood gate — identity false positive and resolver race** (iss-31).
  The secret/PII scanner no longer hard-fails on a system path such as
  `/dev/null` when the machine username collides with a system directory name
  (e.g. a user called `dev`): a local-username match is suppressed only when it
  is the top segment of an absolute system path, so genuine username leaks
  (nested under a home root, or bare) are still caught. The launch bundle's
  compiled-glob cache is now guarded by a mutex, removing a data race when the
  transport-agnostic core resolves bundles concurrently.
- **Memory-ingest boundary — partial-failure reporting and CRLF parity**
  (iss-30, continued). When `abcd memory ingest --keep-original` fails to store
  the original *after* the pages and registry are durably written, it no longer
  reports total failure: the successful ingest is reported (pages listed) with a
  warning and a non-zero exit, and the failure message names only the
  repo-relative store location — no absolute path, in text or `--json`. CRLF
  documents now split identically to their LF form (`splitFileFrontmatter`
  normalises line endings like its sibling parsers), so a `\r\n` closing `---`
  delimiter is no longer rejected and hashes/summaries no longer silently
  degrade.
- **Fail-closed capture surface** (iss-29). A mistyped `capture` subcommand
  (e.g. `abcd capture resovle iss-1 …`) is no longer swallowed as free text and
  filed as a new issue; it is refused with a did-you-mean and writes nothing.
  Errors requested with `--json` are now emitted as a `{"error": …}` envelope
  rather than raw Go text, and `abcd docs lint` with a missing or unreadable
  config reports a clean, repo-relative diagnostic instead of a raw file error
  that leaked the absolute config path.
- `abcd` status now reports `IsGitRepo` correctly in a linked git worktree or a
  submodule, where `.git` is a regular gitfile rather than a directory (iss-72).
- `abcd intent plan` now refuses an `## Acceptance Criteria` section with no
  top-level `-`/`*` bullet, matching the ingest gate — an intent can no longer be
  planned into a state where every fidelity verdict dead-letters for having zero
  positional criteria. The intent template's Audit Notes placeholder is cleared
  when the first review block lands, so a populated audit carries no stale "Empty"
  claim (iss-67).
- The frontmatter scanner (`internal/core/frontmatter`, used by `abcd intent`/
  `spec` and record-lint) now tolerates a trailing space or tab on the `---`
  delimiters; previously a `--- ` closing delimiter went unrecognised and every
  body line after it was misread as a frontmatter field. `record-lint` no longer
  keeps a divergent copy of the scanner — it routes through the canonical one and
  inherits this fix (iss-69).

### Security

- **Memory-ingest fetch/read hardening** (iss-30). `abcd memory ingest` now treats
  a non-2xx HTTP response as an error instead of storing the 404/500 error page as
  source content; the SSRF guard additionally rejects NAT64 (`64:ff9b::/96`) and
  6to4 (`2002::/16`) IPv6 addresses that embed a metadata/loopback/private IPv4; a
  local source file is size-capped like the URL path; and a `~user` path is left
  literal rather than being mangled into `home`+`user`.
- **Spec-store hardening** (iss-68). The spec-store reader now opens a file once
  with `O_NOFOLLOW`+`O_NONBLOCK` and validates the file descriptor before reading,
  closing a symlink-swap window (and never blocking on a FIFO leaf). `NextID` fails
  closed on an intent `spec_id` that carries no parseable reservation number (e.g.
  `spc-` with no digits) instead of silently dropping it from the id-reservation
  scan (which could hand out a colliding id); a `spc-N` or `spc-N-<slug>` form still
  reserves N, consistent with record-lint. (The leaf-only ancestor-symlink guard
  and the atomic-rename clobber check are documented as accepted under the
  trusted-worktree model.)
- **Release receipt-gate hardening** (iss-70). The `receipt_gate` record-lint
  rule now binds each semantic-pass receipt to the gate it attests: a receipt
  satisfies a required gate only when its `policy.detector` equals that gate name,
  not merely when a `<gate>.json` file exists. This closes a hole where one
  genuine PROMOTE receipt copied across every gate's path satisfied them all.
  Arming (`record-lint --release-gate`) now treats the caller's required-gate list
  as authoritative even when empty — an argless arming clears the gates and fails
  closed rather than inheriting the committer-editable in-tree list. The
  `gate_lockstep` workflow parser no longer mistakes a nested `with: name:` for a
  step name. (The receipt-gate remains disabled outside release time.)
- **Secret-scanner serialisation hardening** (iss-65). A serialized scan finding's
  snippet now masks *every* secret on its source line, not only the finding's own
  token — two secrets sharing a line (a minified `.env`, collapsed JSON) no longer
  leak each other into the `abcd launch --json` report. The content sniff no longer
  misclassifies a valid UTF-8 file as binary when a multibyte rune straddles the
  8 KB boundary (which would have skipped scanning it), and a bundle file that
  cannot be read is now surfaced in `unscanned` rather than silently dropped.
- **Issue-ledger transition hardening** (iss-71). `abcd capture resolve`/`wontfix`
  now run their find→move under the same ledger lock id allocation uses, so two
  concurrent conflicting transitions on one issue can no longer land it in two
  status directories at once. A migrator-supplied `ForceID` is validated against
  the `iss-N` shape before any path is built, so a traversal id cannot touch the
  filesystem outside the ledger.
- **Rules-loader trust hardening** (iss-66). The per-repo `.abcd/rules.json` is now
  opened once with `O_NOFOLLOW` and validated on that file descriptor, closing a
  Lstat-then-read window where the file could be swapped for a symlink. The
  prompt-router's per-session dedup state moved off the world-writable shared temp
  dir to the per-user cache dir (`ABCD_RULES_STATE_DIR` still overrides), so a local
  co-tenant can no longer pre-create the predictable state path to suppress rule
  injection.

## [v0.1.0] - 2026-07-07

First tagged milestone: the Go rebuild through Phase 2. abcd is a single,
host-agnostic Go binary that is also a plugin for compatible agent harnesses, holding all
behaviour in a transport-agnostic `internal/core` behind a Cobra CLI front door and
a markdown plugin surface that shells out to it.

### Added

- Phase 0 scaffold: Go module (`github.com/REPPL/abcd-cli`), a
  transport-agnostic `internal/core`, a Cobra CLI front door (`abcd` status
  board and `abcd version`), the plugin surface, and the design record carried
  forward as the build specification.
- Phase 1 — install and launch. `abcd ahoy` installs abcd into a repo
  (folder-kind detection, visibility-driven gitignore, idempotent marker blocks in
  CLAUDE.md/AGENTS.md). `abcd launch --dry-run` renders a curated release bundle
  that excludes `.abcd/**` by default-deny, running a native secret + PII scanner,
  strict SemVer, marketplace-lockstep anti-drift, and newest-per-line retention over
  the bundle.
- Phase 2 — native capture substrates. `abcd history` is a SHA-keyed, redacted,
  gitignored transcript store (`list`, `show`, and a fail-closed `capture` write
  path); `abcd capture` is a directory-as-status issue ledger; `abcd memory`
  provides deterministic ingest / ask / lint.
- `abcd docs lint` (itd-60 layer 1) — a deterministic docs-currency gate over
  `docs/` and the repo root: change-narration in a doc body, a broken relative
  link, or a stray top-level markdown file each fails the gate.
- `record-lint` — a deterministic drift gate for the `.abcd/development` design
  record (banned tokens, git-metadata, link resolution, intent lifecycle), wired
  blocking into CI and the pre-push preflight.
- Derived-versioning design record (intent itd-73 + ADR-31): the release version
  is derived from intents' declared impact, never hand-authored. The derivation
  itself lands in a later phase.
