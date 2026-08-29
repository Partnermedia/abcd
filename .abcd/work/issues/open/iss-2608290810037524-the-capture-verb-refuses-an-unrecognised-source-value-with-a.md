---
schema_version: 1
id: "iss-2608290810037524"
slug: "the-capture-verb-refuses-an-unrecognised-source-value-with-a"
severity: "nitpick"
category: "ux"
source: "agent-observation"
found_during: "intent-implementation-run"
found_at: "internal/surface/cli"
---

The capture verb refuses an unrecognised source value with a message that names neither the offending flag nor the accepted set, and blames the wrong layer: an invalid source produces a malformed-frontmatter error quoting the value, when the value came from a command-line flag and never reached any frontmatter the caller wrote. The accepted values are discoverable only by grepping existing records. The flag's help text lists no enumeration either. The same shape likely applies to category and to any other closed-set flag on this path. A closed set should be named in the refusal and in the help text.