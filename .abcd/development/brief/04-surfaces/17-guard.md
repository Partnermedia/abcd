# `/abcd:guard` — Shell-Hazard Guard

`/abcd:guard` decides whether a shell command is safe to run, against abcd's
hazard registry. It is **strictly read-only** — it performs zero writes, it only
answers. It is the execution-time half of itd-103; the teaching half is the rules
loader injecting the same registry's entries before shell-heavy work.

## Sub-verbs

> _Machine-checked (`surface_coverage`, spc-27): each row records the verb's
> adr-40 bucket (`lint` / `review` / `audit` / `gate`, or `—` for a
> non-assessment verb) and its existence (`shipped` / `staged`), verified
> against the committed command-tree snapshot in both directions._

| Verb | Bucket | Status |
|---|---|---|
| `check` | gate | shipped |
| `hook` | — | shipped |


- **`/abcd:guard check`** — evaluate one candidate command line and report the
  decision. The candidate comes from `--command "<line>"` or from stdin. The
  plugin command (`commands/guard.md`) passes it on **stdin**, inside a
  quoted-delimiter heredoc: a candidate interpolated into a double-quoted
  `--command` argument is expanded by the shell before the guard starts, so a
  command substitution inside it would run at check time — the one moment the
  check exists to prevent. `--command` stays for literals a human typed.
- **`abcd guard hook`** — the host adapter. It reads a pre-tool-use hook payload
  on stdin, applies the same decision, and maps it onto the host's block/allow
  protocol. It is invoked by `hooks/hooks.json`, not by hand.

Bare `abcd guard` prints command usage — it does **not** render a status board;
guard health is reported by `abcd ahoy`, which is where install/health state
lives for every other part of abcd too.

## The decision

Core (`internal/core/guard`) returns one of three verdicts, and the front doors
only format it:

| Verdict | `guard check` | `guard hook` |
|---|---|---|
| `allow` | exit 0 | exit 0, silent |
| `warn` | exit 0, warning rendered | exit 0, warning on stderr |
| `block` | exit 1, why + successor rendered | the host's blocking status, why + successor as the message |

A guard that cannot be **evaluated** — an unparsable command line, a registry
that will not load, a candidate too long to read, a registry switched off — is a
fourth case, and the two front doors part company on it deliberately:

- `guard check` exits **2**. A script that asked the guard a question must never
  read silence as clearance.
- `guard hook` exits **1** and warns loudly. Exit 1 rather than 0 is the whole
  point: a pre-tool-use hook that exits 0 has its stderr discarded, so the
  warning would exist and nobody would ever see it. Exit 1 is non-blocking, so
  the command still runs — a guard that cannot answer never stops a session.

## Fail-open-loud

The installed hook entry wraps the binary call in a shim. The binary's own three
statuses pass through untouched — 0 (allow), 1 (ran, could not decide, allowed
loudly), 2 (block). Anything else means the binary did not run at all: missing,
not executable, crashed, killed. The shim then allows the command and emits an
unmissable `UNGUARDED` warning of its own. A session is therefore never bricked
by a broken guard, and never silently unprotected either — and a binary that ran
and reported is never described as having failed to run.

The outside-the-session half is `abcd ahoy`'s `guard:` line, which reports the
three things that can independently be false: whether the hook is installed,
whether the binary it calls is reachable, and whether the registry loads. Each
also surfaces as a gap (`guard.hook_missing`, `guard.binary_unreachable`,
`guard.registry_unloadable`).

## Registry and overrides

The hazard entries are bundled in the binary and merged with a repo's
`.abcd/guard.json` — a dedicated file rather than a rules-loader domain, so a
rules kill switch can never silently disable a safety guard. An entry key
overrides one field or declares a new hazard; `{"disabled": true}` switches the
guard off.

There is **no flag, environment variable, or prompt** that disarms the guard for
a session: the file is the only route, so the change lands in a diff. What the
file is *not*, today, is verified as committed — `guard.Load` reads the working
tree, so an edit takes effect on the next command, before review. The mitigation
is loudness rather than enforcement: a disabled registry makes every command it
lets through carry an UNGUARDED warning naming the file, and `abcd ahoy` reports
`OFF`. Closing the gap properly (refusing a `disabled: true` that is not in
`HEAD`) is a core-side change to `guard.Load`, tracked as an issue.

## What an allow means

An allow means **no registry entry matched** — never that a command is safe. The
guard reads command names it can see in command position, so a hazard reached any
other way is not seen:

- a command string handed to an interpreter (`eval`, `sh -c`);
- one launched through a wrapper outside the known set (`sudo`, `doas`,
  `command`, `env`, `nohup`, `time`, `xargs`, `timeout`, `exec`);
- one launched through a known wrapper carrying a value-taking flag the matcher's
  per-wrapper table does not name: `sudo -u bob <hazard>` is seen, the bundled
  short form `sudo -Hu bob <hazard>` is not (`bob` is read as the command). The
  miss is a non-match, never a false block;
- an API path an entry names by its ROOT segment when the host serves that API
  under a prefix: a GitHub Enterprise Server install mounts the same endpoints
  under `/api/v3/`, so `gh api -X DELETE https://ghe.example/api/v3/repos/o/r`
  is not seen. The `https://api.github.com/…` URL form *is* — an operand is
  normalised to its path (scheme, host, query, fragment dropped) before the
  depth check. Matching a root segment wherever it appeared would close the
  remainder and falsely refuse `DELETE /teams/{id}/repos/{owner}/{repo}`, which
  removes a repository from a team and destroys nothing;
- one inside a backtick substitution (`$(…)` is followed, backticks are not);
- a dangerous form no entry describes.

A wrapper's own arguments are stepped over with it, including the mandatory
operand in `timeout DURATION COMMAND` (iss-148).

Coverage is what the registry names. The registry grows from reality: a
facilitator who sees something scary captures it, and recurring captures are
promoted into the bundled defaults through the admission gate.

## References

- Plugin command: [`commands/guard.md`](../../../../commands/guard.md)
- Spec: [`spc-16`](../../specs/closed/spc-16-abcd-teaches-repo-agents-the-shell-commands-they-must-never.md)
- Intent: [`itd-103`](../../intents/shipped/itd-103-abcd-teaches-repo-agents-the-shell-commands-they-must-never.md)
- Install/health surface: [`01-ahoy.md`](01-ahoy.md)
