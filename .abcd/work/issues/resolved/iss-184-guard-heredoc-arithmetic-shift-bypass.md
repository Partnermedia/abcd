---
schema_version: 1
id: "iss-184"
slug: "guard-heredoc-arithmetic-shift-bypass"
severity: "critical"
category: "bug"
source: "agent-finding"
found_during: "bug-hunt loop round 1"
found_at: "internal/core/guard/tokenize.go:152"
resolution: "Fixed at root cause: skipHeredocBodies now signals when a pending here-document never finds its terminating delimiter line, and tokenize turns that into ErrUnparsableCommand instead of silently consuming the rest of the input. Covers both the identifier-operand arithmetic-shift misparse and genuinely unterminated heredocs (same underlying gap). Guard fails open LOUDLY (per the documented itd-103 contract) on the unparsable command instead of silently allowing it. Detector: internal/core/guard/tokenize_test.go TestArithmeticShiftByIdentifierIsRejected + TestTokenizeRejectsUnterminatedHeredoc, both watched failing before the fix and passing after."
impact: fix
---

guard tokenizer heredoc misparse: an unquoted arithmetic left-shift with an identifier operand (e.g. `$((1<<shift))`) is misparsed as a here-document start by isDelimStart (tokenize.go:206, accepts any word starting with a letter/underscore) plus readHeredocDelim (tokenize.go:249, terminates the delimiter word on ')'). skipHeredocBodies (tokenize.go:264) then scans for a line equal to the bogus delimiter, finds none, and silently swallows every subsequent line of the command before it reaches command position. A dangerous command on a later line (git push --force, rm -rf, gh repo delete, etc.) is never seen by Registry.Check and the guard returns VerdictAllow with no error — a silent fail-open, not the loud could-not-answer path the shim reserves for unparsable input. Confirmed live via both CLI front doors (internal/surface/cli/guard.go: guard check and guard hook). The literal-digit form (1<<20) is unaffected; only identifier/variable operands trip it. Reproducing test: internal/core/guard/tokenize_test.go, TestArithmeticShiftByIdentifierIsNotAHeredoc.