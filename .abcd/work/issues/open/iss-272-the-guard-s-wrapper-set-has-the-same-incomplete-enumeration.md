---
schema_version: 1
id: "iss-272"
slug: "the-guard-s-wrapper-set-has-the-same-incomplete-enumeration"
severity: "minor"
category: "observation"
source: "user-observation"
found_during: "manual-capture"
---

The guard's WRAPPER set has the same incomplete-enumeration shape gh-297 fixed for interpreters. gh-297 made the execute-a-string INTERPRETER family one shared predicate (sh, bash, dash, zsh, ksh, mksh, ash). The wrappers map in internal/core/guard/match.go is a separate list of commands that merely exec another, and it omits several that are equally ordinary: nice (and 'nice -n 5'), setsid, flock, stdbuf, ionice, unshare, chroot, and busybox sh -c / busybox ash -c. Each is a silent allow today for every bundled blocker -- and equally so before gh-297 with plain sh, so none is introduced by that change. Also absent from BOTH sets: su -c '<hazard>', runuser -c '<hazard>' and script -c '<hazard>' /dev/null, which are execute-a-string interpreters in exactly the sense gh-297's rationale describes (each runs its -c operand with the grammar the tokenizer already parses) but are not shells, so they belong in neither list as currently shaped. The realistic ones are nice, setsid, busybox sh and su -c. Out of scope by contrast: ssh host '<cmd>' and docker run img sh -c '<cmd>' execute on a different host or namespace. Note for whoever takes this: gh-297's comment now says explicitly that 'nowhere left to widen' covers the interpreter set only, not the wrapper set, so the claim and this gap do not contradict each other. Found by the security review of gh-297.