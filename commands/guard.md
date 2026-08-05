---
name: guard
description: Check a shell command against abcd's hazard registry before it runs, by invoking the abcd binary. Read-only; performs zero writes.
argument-hint: "[check <command> | hook]"
---

# `/abcd:guard` shell-hazard check

Decide whether a shell command is safe to run, using abcd's hazard registry —
the bundled hazard entries merged with this repo's `.abcd/guard.json`. This
command performs **zero writes**.

## `check` — decide one command

Pass the candidate on **stdin**, inside a quoted-delimiter heredoc:

```bash
abcd guard check --json <<'ABCD_GUARD_EOF'
<the command line, verbatim, on one or more lines>
ABCD_GUARD_EOF
```

Never interpolate the candidate into `--command "…"`. The shell expands a
double-quoted argument before `abcd` ever starts, so a candidate containing
`$(…)` or a backtick would **execute at check time** — the exact moment the check
exists to prevent — and an embedded `"` would break the quoting outright. The
quoted delimiter (`<<'ABCD_GUARD_EOF'`, quotes included) switches expansion off,
so the candidate reaches the guard as written. Use `--command` only for a literal
you typed yourself.

Then report the JSON to the user:

- `verdict` — `allow`, `warn`, or `block`.
- `entry_id`, `tier` — which hazard matched, and how severe it is.
- `why` — the plain-language reason, written for a non-expert.
- `successor` — the safe form to run instead.
- `matches` — every entry the command tripped, blockers first.

Exit codes: `0` for allow and warn, `1` for a block, `2` when the guard could
not be evaluated at all (an unparsable command line, or a `.abcd/guard.json`
that does not load). Treat `2` as a fault to report, never as a clearance.

On a `block`, do not run the command. Tell the user the `why`, then run the
`successor` instead — the refusal is the lesson, so pass it on in full. On a
`warn`, the command may run; surface the warning first so the user can stop it.

## `hook` — the host adapter

```bash
abcd guard hook
```

Reads a host pre-tool-use hook payload on stdin and applies the same decision
before a shell command executes. It is invoked by the plugin's hook manifest,
not by hand; a blocker returns the host's blocking status with the successor and
the why as the message, and a warn or an allow lets the command run.

Anything the adapter cannot turn into a decision — an unreadable payload, a tool
call that is not a shell command, a registry that does not load — allows the
command and warns loudly. A guard that cannot answer never stops a session, and
is never silently absent.

## Registry and overrides

The bundled hazards ship inside the binary. A repo overrides them in one file in
the repository, `.abcd/guard.json`: add an entry key to change one field (for
example `{"tier": "warn"}`) or to declare a new hazard, and set
`{"disabled": true}` to switch the guard off entirely.

There is no flag, environment variable, or prompt that turns the guard off for a
session — the file is the only route, so the change lands in a diff someone
reviews. Two things follow, and both must be said plainly if a user asks. The
file is read from the working tree, so an edit takes effect on the very next
command, before anyone has reviewed it. And a repo whose guard is switched off is
an unguarded session: every command it lets through carries an UNGUARDED warning
naming the file, so the state cannot pass unnoticed.

**Never write `.abcd/guard.json` on your own initiative.** Disabling or
retiering a hazard is the user's decision to make and to review.

### What an allow does and does not mean

An allow means **no registry entry matched**. It is never a statement that a
command is safe. The guard reads command names it can see in command position, so
a hazard reached any other way is not seen: a command string handed to an
interpreter (`eval`, `sh -c`), one launched through a wrapper outside the known
set, one launched through a known wrapper carrying a value-taking flag the guard
does not name (`sudo -u bob <hazard>` is seen; the bundled short form
`sudo -Hu bob <hazard>` is not), one whose API path an entry names by its ROOT
segment but the host serves under a prefix (a GitHub Enterprise Server install
mounts the same endpoints under `/api/v3/`; the `https://api.github.com/…` URL
form **is** read), one inside a backtick substitution, or a dangerous form no
entry describes.
Coverage is what the registry names. Say exactly this if a user asks about
coverage — never that the guard cleared the command.

A candidate too long to read is refused (exit 2), not answered on the part that
fitted.

To check whether the guard is actually armed in this repo, run `abcd ahoy` and
read its `guard:` line: it reports whether the hook is installed, whether the
binary is reachable, and whether the registry loads.

If the `abcd` binary is not on `PATH`, fall back to `go run ./cmd/abcd guard
check --json` from the repo root, with the same quoted-delimiter heredoc on
stdin, or run `go run ./cmd/abcd ahoy install` to put a binary on `PATH`.

**User input:** $ARGUMENTS
