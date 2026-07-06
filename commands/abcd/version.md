---
name: version
description: Print the installed abcd version by invoking the abcd binary. Read-only.
---

# `/abcd:version`

Report the installed abcd version. This command performs **zero writes**.

Run:

```bash
abcd version --json
```

Then tell the user the `name` and `version` from the JSON. If the `abcd` binary
is not on `PATH`, fall back to `go run ./cmd/abcd version --json` from the repo
root, or tell the user to build it with `make build`.

**User input:** $ARGUMENTS
