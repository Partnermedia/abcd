---
schema_version: 1
id: "iss-266"
slug: "runeless-parents-exit-zero-on-unknown-subverb"
severity: "minor"
category: "ux"
source: "impl-review"
found_during: "spc-30 build"
found_at: "internal/surface/cli/cli.go"
resolution: "Every cobra parent is now runnable, so its declared cobra.NoArgs refuses an unknown sub-verb at exit 2 instead of printing help at exit 0. Six parents fixed (docs, docs cite, guard, embark, ideate, hook); the two existing inline fixes consolidated onto a shared helpRunE. Held tree-wide by TestEveryParentIsRunnable, which walks the live command tree rather than a hand-kept list."
impact: fix
---

cobra parents with no RunE are not Runnable, so cobra prints help and exits 0 WITHOUT validating args: an unknown sub-verb reads as success to a script. Confirmed on 'abcd docs nonsense' and 'abcd guard nonsense' (both exit 0); 'abcd memory nonsense' and 'abcd history nonsense' exit 1 because those parents have a RunE. spc-30 fixed the disembark parent (a retired sub-verb spelling must refuse, not silently succeed) by giving it a RunE that prints help, so cobra.NoArgs then refuses a stray token at exit 2. The same one-line fix is owed to every other RunE-less parent — sweep them so no unknown sub-verb anywhere exits 0.