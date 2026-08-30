---
schema_version: 1
id: "iss-2608301308367566"
slug: "record-stores-values-are-not-containment-checked-so-a-commit"
severity: "major"
category: "security"
source: "user-observation"
found_during: "itd-189-round-2-security"
found_at: "internal/core/lint/config.go"
---

record_stores values are not containment-checked so a committed config walks the record gate outside the repository

Found by the round-2 adversarial security review of build/itd-189.
PRE-EXISTING -- reproduces identically on 932629f9 and on main.

A PR that edits `.abcd/record-lint.json` to `"adr": "../outside/decisions"`
reaches `filepath.Join(repoRoot, ...)` at schema.go:1007 with no `..` or
`IsAbs` rejection, and the gate then reads and echoes `.md` frontmatter from
outside the checkout. Reproduced on both HEAD and base:

    ../outside/decisions/0001-outside-secret.md:2: [BLOCKER record_schema]
    filename claims id 'adr-1' but frontmatter declares 'adr-77'

`validateRecordStores` checks only prefix membership and inter-store layout,
not containment.

Mitigating factor, and why this is major rather than critical: the edit is
glaringly visible in the PR diff, and a config change is exactly the thing
review looks at. This is a defence-in-depth gap, not an invisible one.

Remedy: reject any `record_stores` value that is absolute or that escapes the
repo root after `filepath.Clean`.

Sibling: iss-2608301203521317 (the store walk's raw os.ReadFile). Both are the
same underlying shape -- the GATE reads attacker-influenceable paths without
the guard the repo already owns -- and `fsutil.ReadGuarded` closes the read
half of both at one call site.
