---
schema_version: 1
id: "iss-2609012040003576"
slug: "ghsa-5wx3-2c86-fjpx-draft-advisory-severity-medium-cwe-636-n"
severity: "major"
category: "security"
source: "review-followup"
found_during: "autonomous-run-2026-09-01"
origin: researcher-authored
production_mode: hand-written
found_at: "internal/core/guard/tokenize.go"
---

GHSA-5wx3-2c86-fjpx (draft advisory, severity medium; CWE-636, Not Failing Securely): the wired PreToolUse guard hook fails open on two tokenizer error states that every shell an agent runs under executes. `tokenize` (internal/core/guard/tokenize.go) returns ErrUnparsableCommand for a backslash as the last byte of the line and for a here-document body that never reaches its delimiter line, and `newGuardHookCommand` (internal/surface/cli/guard.go) maps every Check error to exit 1 'NOT CHECKED … runs UNGUARDED', so `git push --force origin main \` and `git push --force origin main <<EOF` followed by a newline run without a verdict — bash 3.2 and 5.3, zsh and dash all execute both (the trailing backslash is dropped by bash 3.2/zsh, kept as a literal word by bash 5.3/dash; the empty heredoc body is recovered silently). A third shape shares the root cause: spaced arithmetic `echo $(( x << y ))` on one line and a hazard on the next is misread as a heredoc, because the `<<` branch tests only the byte immediately after the delimiter word for a paren and `followedByParen` sees the space, so the hazard line is swallowed as body and the same error returns. The sibling `shellInspect` (internal/core/guard/payload.go) maps the inner tokenizer's error to a warn, so `sh -c 'git push --force origin main \'` is a loud-warn allow through the same states. The fix must give the two bash-executable states a VERDICT in core rather than an error: arm 1 parses a trailing backslash as bash 3.2 does (dropped — the reading under which the hazard runs, so the fail-safe one) and yields a precise Tier-1 verdict; arm 2 turns an unterminated heredoc body into a segment flag that Check folds into a synthetic fail-closed block (the 2dbb0ca2 brace-group precedent: an error would route to the hook's fail-open), never a remap of the hook's error path, because the unterminated-quote class no shell runs must stay the loud fail-open of spc-16 / iss-269; prerequisite in the same change, the `<<` branch skips blanks before the paren test so `$(( x << y ))` classifies as arithmetic. The hook then blocks both spellings with the entry that names them, `echo hello \` allows, `rm -rf "unterminated` still fails open loud, and the repo corpus holds.
