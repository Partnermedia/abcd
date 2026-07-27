# `/abcd:guard` — Shell-Hazard Guard

`/abcd:guard` decides whether a shell command is safe to run, against abcd's
hazard registry. It is **strictly read-only** — it performs zero writes, it only
answers. It is the execution-time half of itd-103; the teaching half is the rules
loader injecting the same registry's entries before shell-heavy work.

## Sub-verbs

- **`/abcd:guard check`** — evaluate one candidate command line and report the
  decision. The candidate comes from `--command "<line>"` or from stdin. The
  plugin command (`commands/abcd/guard.md`) passes it on **stdin**, inside a
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

A guard that cannot be **evaluated** is a fourth case, and the two front doors
part company on it deliberately:

- `guard check` exits **2**. A script that asked the guard a question must never
  read silence as clearance.
- `guard hook` exits **0** and warns loudly. A guard that cannot answer must
  never be the reason a session stops.

## Fail-open-loud

The installed hook entry wraps the binary call in a shim: if the abcd binary is
missing, not executable, crashes, or is killed, the shim allows the command and
emits an unmissable `UNGUARDED` warning. Only the binary's own exit statuses —
allow and block — propagate. A session is therefore never bricked by a broken
guard, and never silently unprotected either.

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
- one launched through a wrapper outside the small known set (`sudo`, `doas`,
  `command`, `env`, `nohup`, `time`) — `xargs`, `timeout`, `exec` are not in it;
- one inside a backtick substitution (`$(…)` is followed, backticks are not);
- a dangerous form no entry describes.

Coverage is what the registry names. The registry grows from reality: a
facilitator who sees something scary captures it, and recurring captures are
promoted into the bundled defaults through the admission gate.

## References

- Plugin command: [`commands/abcd/guard.md`](../../../../commands/abcd/guard.md)
- Spec: [`spc-16`](../../specs/open/spc-16-abcd-teaches-repo-agents-the-shell-commands-they-must-never.md)
- Intent: [`itd-103`](../../intents/planned/itd-103-abcd-teaches-repo-agents-the-shell-commands-they-must-never.md)
- Install/health surface: [`01-ahoy.md`](01-ahoy.md)
