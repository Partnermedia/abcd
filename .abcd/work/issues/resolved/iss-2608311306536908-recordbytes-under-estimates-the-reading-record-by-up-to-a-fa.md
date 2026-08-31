---
schema_version: 1
id: "iss-2608311306536908"
slug: "recordbytes-under-estimates-the-reading-record-by-up-to-a-fa"
severity: "minor"
category: "bug"
source: "agent-finding"
found_during: "manual-capture"
origin: researcher-authored
production_mode: hand-written
resolution: "recordBytes counts every value double, an exact bound rather than a guess: the record's scalar writer escapes two characters at one byte each and the hidden-rune encoder emits neither. The envelope allowance is widened for the redactor, and the residue is stated."
impact: fix
resolved_by:
  intent: "itd-185"
  spec: "spc-63"
---

recordBytes under-estimates the reading record by up to a factor of two: it is measured before the record writer escapes backslashes and double quotes, so an item body of quote characters lands a record roughly twice the measured size and past issueschema.RecordReadLimit. The envelope allowance is also measured before the ledger redactor runs, which can lengthen text when a placeholder is longer than the secret it replaces.

## Grounds

- pursued: TestEscapingCannotPushARecordPastTheLimit lands a body of double quotes that a single-counted estimate passes, and every written record is asserted under the family limit; a record past the limit appearing again would show this wrong
