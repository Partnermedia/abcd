// Package guard is abcd's transport-agnostic shell-hazard registry (itd-103,
// spc-16). It owns one registry of dangerous command patterns and the
// deterministic decision that a harness hook calls at execution time: given a
// candidate command line it returns allow, warn, or block — never writing to
// stdout, never exiting, never knowing a transport. Front doors under
// internal/surface format the decision, and the fail-open-loud behaviour on a
// broken guard lives in the hook shim, not here.
//
// The model is a set of binary-bundled default entries (embedded from
// defaults/guard.json) merged with an optional per-repo
// <repoRoot>/.abcd/guard.json override. Each entry declares an id, a
// command-position pattern over shell tokens, a tier (blocker/warn), the safe
// successor, a plain-language why, and the known-bad/known-good fixtures that
// prove it. Bundled entries are held to an admission gate (fixtures present,
// known-good at least 40% of the corpus, 100% true-negative rate) enforced by
// TestBundledEntriesPassAdmissionGate, so an entry cannot ship without proof.
//
// Matching is shell-token-aware and applies in COMMAND POSITION only: the
// candidate is tokenised with shell quoting honoured, so a hazard named inside a
// quoted string argument (an incident capture whose text mentions
// "cd scratch && rm -rf *") never fires. Quoting affects tokenisation, not
// argument semantics — `git push '--force'` still fires, because the shell hands
// the same argv either way.
package guard
