BINARY := abcd
BINDIR := bin
TARGETS := darwin/arm64 darwin/amd64 linux/arm64 linux/amd64
# Version stamped into `abcd version`. Defaults to the in-source "dev" value; the
# release build passes the git tag (VERSION=vX.Y.Z). SemVer, v-prefixed.
VERSION ?=
# -s -w strips the symbol table and DWARF debug info; -X stamps the version.
# -trimpath (in the build recipe) rewrites absolute source paths to module paths
# so no local filesystem path is embedded — a smaller, path-clean binary suitable
# for public distribution.
LDFLAGS := -s -w$(if $(VERSION), -X github.com/intentdriven/abcd/internal/core.Version=$(VERSION),)

.PHONY: build test vet clean preflight lint-reviews lint-issues record-lint docs-lint site-render smoke \
	check-attribution scaffold-sync scaffold-sync-check

# Cross-compile every supported target to bin/abcd-<goos>-<arch>.
# Pass VERSION=vX.Y.Z to stamp the version (release builds); omit for a dev build.
build:
	@mkdir -p $(BINDIR)
	@for target in $(TARGETS); do \
		goos=$${target%/*}; goarch=$${target#*/}; \
		out=$(BINDIR)/$(BINARY)-$$goos-$$goarch; \
		echo "building $$out"; \
		GOOS=$$goos GOARCH=$$goarch go build -trimpath -ldflags "$(LDFLAGS)" -o $$out ./cmd/abcd || exit 1; \
	done

test:
	go test ./...

# Self-discovering smoke harness (evals/): build the binary, walk the Cobra tree,
# run every command's --help + the read-only verbs. Behind the `smoke` build tag so
# it stays out of the unit-test lane; run explicitly here, in CI's smoke job, and in
# the release verify gate.
smoke:
	go test -tags smoke ./evals/...

vet:
	go vet ./...

# Deterministic gate for the .abcd/work/reviews/ charter (RD001-RD003) — a
# stopgap until these codes land in internal/core/lint. Needs full git history
# (RD002 is append-only over committed history): on a shallow checkout the
# script refuses (exit 2) rather than pass vacuously with nothing covered.
lint-reviews:
	@bash scripts/check-reviews.sh

# AI-attribution gate (AGENTS.md § Attribution). Checks the commit trailers on
# this branch against the default branch; the pull-request BODY half runs only in
# CI, where the body exists. Not a preflight prerequisite — preflight guards a
# push, and the body it must agree with is not written until the PR is opened.
check-attribution:
	@bash scripts/check-attribution.sh commits origin/main HEAD
	@bash scripts/check-attribution-cases.sh

# Deterministic drift gate for the .abcd/development design record (first slice
# of internal/core/lint). Blocking: any record drift (stale tool names, dropped
# concepts, lifecycle or reference breakage) fails preflight and CI.
record-lint:
	@go run ./cmd/record-lint

# Deterministic issue-resolution gate (iss-2608241347321757). RS001: a
# `Resolves: iss-N` trailer must be accompanied by that record entering a terminal
# folder (resolved/ or wontfix/) in the same change, so resolution lands INSIDE
# the fix and no post-merge step
# exists to forget. RS002/RS003: a resolved_by.commit sha must name a commit
# that is actually reachable — `--commit` is shape-checked only, and the repo
# allows squash and rebase merges, either of which rewrites a cited branch sha
# out of existence. The cases run first: a gate nobody has watched fail is an
# enforcement claim with no evidence behind it. Needs full git history, like
# lint-reviews: on a shallow checkout the scripts refuse (exit 2) rather than
# report unfetched commits as ledger violations.
lint-issues:
	@bash scripts/check-issue-resolution-cases.sh
	@bash scripts/check-issue-resolution.sh ledger HEAD
	@bash scripts/check-issue-resolution.sh commits origin/main HEAD

# Deterministic docs-currency gate (itd-60): the same internal/core/lint engine,
# driven over docs/ and the repo root via the transport-agnostic `abcd docs lint`
# verb. Blocking: change-narration in a doc body, a broken relative link, or a
# stray root markdown file fails preflight and CI.
docs-lint:
	@go run ./cmd/abcd docs lint

# Site-render gate (iss-2608241845109280). The site renders the RECORD as well as
# docs/ — 852 records at the 0.6.4 cut — and the renderer supports a fixed
# markdown subset, refusing anything it would otherwise pass through unrendered.
# So a record is a site input, and a malformed one breaks the render. Nothing
# caught that before this target existed: the four lint gates read records but
# never render them, and site-screenshots.yml is path-filtered to the generator
# and docs/, so a records-only pull request changed the rendered page and
# triggered no audit. It reached a release, where `release.yml`'s site job failed
# AFTER the binaries had published.
#
# Builds into a throwaway directory, then runs `site check` over it before
# discarding it. Build alone answers "does it render"; the publish gates
# (provenance, hero, banned-tokens, snippets, baseline, mobile, figure-labels)
# are what release.yml's post-publish site job runs, so a change that renders but
# trips a check gate would otherwise pass every pre-publish gate and fail only
# AFTER the binaries ship. `site check` needs no mkdocs — it excludes docs/ — so a
# plain build output suffices. Roughly ten seconds, almost all of it writing files.
site-render:
	@rm -rf .abcd/.work.local/scratch/site-render-check
	@go run ./cmd/abcd site build --out .abcd/.work.local/scratch/site-render-check >/dev/null
	@go run ./cmd/abcd site check --out .abcd/.work.local/scratch/site-render-check >/dev/null
	@rm -rf .abcd/.work.local/scratch/site-render-check
	@echo "site-render: the record and docs render and pass the site gates"

# Propagate the pinned action refs in the committed release workflows back into
# the scaffold templates they were rendered from (iss-209). Dependabot only ever
# edits the rendered workflow, which breaks the self-scaffold parity test; this
# is the one-way fix — re-rendering the template over the workflow would revert
# the bump instead.
#
# Nothing in CI calls either target. The drift is GATED by `go test`
# (TestSelfScaffoldParity and TestSyncRepoPinsIsCleanToday, both under
# preflight), which is what fails a bump that only landed on the workflow;
# these are the human's read-only look at it and the one-command fix.
scaffold-sync:
	@go run ./cmd/scaffold-sync

scaffold-sync-check:
	@go run ./cmd/scaffold-sync -check

# Pre-push gate (invoked by .githooks/pre-push): the four lint gates
# (lint-reviews, lint-issues, record-lint, docs-lint) plus the site-render gate
# as prerequisites, then build, vet, test,
# and race-enabled internal tests natively. CI's check job runs those same four
# Go steps plus a `gofmt -l .` format gate this target does not, so run gofmt
# separately before pushing. Host-native `go build` (not the cross-compiling
# build target) because it mirrors CI.
preflight: lint-reviews lint-issues record-lint docs-lint site-render
	go build ./...
	go vet ./...
	go test ./...
	go test -race ./internal/...

clean:
	rm -rf $(BINDIR)
