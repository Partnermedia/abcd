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

The bundled hazards ship inside the binary. A repo overrides them in a committed,
reviewable file, `.abcd/guard.json`: add an entry key to change one field (for
example `{"tier": "warn"}`) or to declare a new hazard, and set
`{"disabled": true}` to switch the guard off entirely. There is no in-session
override — turning a guard off is a change someone can review.

Command strings passed to `eval` or `sh -c` are not parsed, so a hazard hidden
inside one is not seen. Say so if a user asks about coverage.

To check whether the guard is actually armed in this repo, run `abcd ahoy` and
read its `guard:` line: it reports whether the hook is installed, whether the
binary is reachable, and whether the registry loads.

If the `abcd` binary is not on `PATH`, fall back to `go run ./cmd/abcd guard
check --json` from the repo root, with the same quoted-delimiter heredoc on
stdin, or tell the user to build it with `make build`.

**User input:** $ARGUMENTS
