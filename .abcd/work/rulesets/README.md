# Branch rulesets — committed mirror

The two rulesets applied to this repository's default branch, mirrored here in
normalised form (name, target, enforcement, conditions, bypass actors, rules —
no volatile ids or timestamps) so every gate claim in the contribution
documents has a source of truth in the tree rather than only a web console.

- `main-protection.json` — deletion/force-push protection, the seven required
  status checks (strict), and the merge queue. **No bypass actors**: nobody
  merges past a red required check.
- `main-review.json` — the pull-request rule requiring a code-owner review on
  changes to the publish surface (`.github/CODEOWNERS`). Repository admins may
  bypass, which is why it is a separate ruleset: a bypass covers every rule in
  its ruleset, so the review requirement and the required checks must never
  share one.

Refresh after any settings change (ids: 19969675, 21045110):

```sh
gh api repos/Partnermedia/abcd/rulesets/<id> \
  | jq -S '{name, target, enforcement, conditions, bypass_actors, rules}'
```

The automated drift check — the live rulesets diffed against these files —
is itd-92's remit; until it ships, this mirror is refreshed by hand in the
same change as the settings edit it records.
