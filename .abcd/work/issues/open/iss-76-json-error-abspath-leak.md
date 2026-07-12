---
schema_version: 1
id: "iss-76"
slug: "json-error-abspath-leak"
severity: "minor"
category: "bug"
source: "agent-finding"
found_during: "2026-07-12 /abcd:run iss-29 security review"
---

cli.Run now routes all command errors through the --json envelope, so any verb that returns a bare *fs.PathError (os.ReadFile/os.Open failures not wrapped by core) emits the absolute local path into machine JSON output, violating no-absolute-paths-in-machine-output. iss-29 sanitised the docs-lint config-load branch specifically; the systemic fix is to sanitise PathError-bearing errors at the Run() boundary (or audit which errors reach the envelope). Detector: a table test that runs each verb's known filesystem-error path under --json and asserts the envelope carries no absolute path. Pre-existing as stderr text; newly widened to --json by cli.Run.