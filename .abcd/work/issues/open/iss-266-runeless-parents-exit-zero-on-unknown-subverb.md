---
schema_version: 1
id: "iss-266"
slug: "runeless-parents-exit-zero-on-unknown-subverb"
severity: "minor"
category: "ux"
source: "impl-review"
found_during: "spc-30 build"
found_at: "internal/surface/cli/cli.go"
---

cobra parents with no RunE are not Runnable, so cobra prints help and exits 0 WITHOUT validating args: an unknown sub-verb reads as success to a script. Confirmed on 'abcd docs nonsense' and 'abcd guard nonsense' (both exit 0); 'abcd memory nonsense' and 'abcd history nonsense' exit 1 because those parents have a RunE. spc-30 fixed the disembark parent (a retired sub-verb spelling must refuse, not silently succeed) by giving it a RunE that prints help, so cobra.NoArgs then refuses a stray token at exit 2. The same one-line fix is owed to every other RunE-less parent — sweep them so no unknown sub-verb anywhere exits 0.