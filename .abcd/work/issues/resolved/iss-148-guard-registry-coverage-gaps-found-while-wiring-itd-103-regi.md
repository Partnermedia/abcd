---
schema_version: 1
id: "iss-148"
slug: "guard-registry-coverage-gaps-found-while-wiring-itd-103-regi"
severity: "minor"
category: "observation"
source: "user-observation"
found_during: "manual-capture"
resolution: "All four gaps closed: the +refspec force-in-disguise as a new bundled entry (new arg_prefixes pattern field); xargs, timeout and exec added to the matcher's wrappers set, with timeout's mandatory DURATION operand stepped over too; a wrapper's own flags stepped over off an explicit per-wrapper value-flag table, so sudo -u bob <hazard> no longer defangs an entry; and backtick substitution followed like $( … ), at parity with the parens handling that already followed it."
impact: fix
---

guard registry coverage gaps found while wiring itd-103 (registry content, not matching semantics): a push whose refspec carries a leading plus is a force in disguise and no entry describes it; xargs, timeout and exec are absent from the matcher's wrappers set, so a hazard launched through one of them is not seen; a backtick command substitution is not followed, while the dollar-paren form is; and a wrapper that IS in the set defangs an entry the moment it carries its own flags, because only the wrapper name is stepped over — `sudo <hazard>` is seen, `sudo -u bob <hazard>` is not, and the same holds for `env -i` and `time -p`. That last one is the sharpest: it turns an entry the registry does describe into an allow with one extra token, and it is the only item here that a facilitator would reasonably assume was covered. Candidates for the admission gate as the registry grows from reality; the wrapper-flag item is matcher-side (`commandOf` in internal/core/guard/match.go), not registry content.