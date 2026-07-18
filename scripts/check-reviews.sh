#!/usr/bin/env bash
# Deterministic gate for the .abcd/work/reviews/ charter (RD001-RD003).
#
# Stopgap: enforces the machine-checkable half of the reviews-folder charter
# until these codes are implemented in `internal/core/lint` (Go). The semantic
# half (provenance discriminator, "not a shadow backlog") is not mechanisable
# and stays a convention. See:
#   .abcd/development/brief/05-internals/06-lint.md  (RD family)
#   .abcd/work/reviews/README.md                     (the charter)
set -euo pipefail
cd "$(git rev-parse --show-toplevel)"

ROOT=".abcd/work/reviews"
fail=0
note() { echo "  $1" >&2; }

[ -d "$ROOT" ] || { echo "check-reviews: no $ROOT — nothing to check"; exit 0; }

# RD001 — review directory shape: each directory under reviews/ is named
# <YYYY-MM-DD>-<kebab-scope> and carries a 00-summary.md. (reviews/README.md,
# the charter itself, sits at the root and is exempt.)
for d in "$ROOT"/*/; do
  [ -d "$d" ] || continue
  base="$(basename "$d")"
  # Semantic-gate receipt directories are sha-keyed (.abcd/work/reviews/<40-hex>/
  # <gate>.json, iss-35) — a distinct artifact class from the dated human-review
  # dirs this charter governs, with their own integrity check (the receipt_gate
  # record-lint rule + release.yml attestation). Exempt them from RD001's
  # <date>-<scope>/00-summary.md shape.
  printf '%s' "$base" | grep -Eq '^[0-9a-f]{40}$' && continue
  printf '%s' "$base" | grep -Eq '^[0-9]{4}-[0-9]{2}-[0-9]{2}-[a-z0-9]+(-[a-z0-9]+)*$' \
    || { note "RD001 $d — directory name must be <YYYY-MM-DD>-<kebab-scope>"; fail=1; }
  [ -f "${d}00-summary.md" ] \
    || { note "RD001 $d — missing required 00-summary.md"; fail=1; }
done

# Review files live inside dated dirs (depth >=2); the root README is mutable.
# (while-read, not mapfile — portable to the bash 3.2 that macOS ships.)
files=()
while IFS= read -r line; do files+=("$line"); done < <(find "$ROOT" -mindepth 2 -type f -name '*.md' | sort)

for f in "${files[@]:-}"; do
  [ -n "$f" ] || continue
  # RD002 — append-only: no post-creation modify/rename in committed history.
  [ -z "$(git log --format=%H --diff-filter=MR -- "$f" 2>/dev/null)" ] \
    || { note "RD002 $f — review files are append-only; edited after creation"; fail=1; }
  # RD003 — path hygiene: repo-relative only, no absolute personal paths.
  if grep -nE '/Users/|/home/[a-z]|C:\\Users' "$f" >/dev/null 2>&1; then
    note "RD003 $f — contains an absolute personal path (use repo-relative)"; fail=1
  fi
done

if [ "$fail" -ne 0 ]; then
  echo "check-reviews: FAILED — reviews-charter discipline (RD001-RD003)" >&2
  exit 1
fi
echo "check-reviews: OK — $ROOT (${#files[@]} review files), RD001-RD003 clean"
