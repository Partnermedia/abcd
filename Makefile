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
LDFLAGS := -s -w$(if $(VERSION), -X github.com/REPPL/abcd-cli/internal/core.Version=$(VERSION),)

.PHONY: build test vet clean preflight lint-reviews record-lint

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

vet:
	go vet ./...

# Deterministic gate for the .abcd/work/reviews/ charter (RD001-RD003) — a
# stopgap until these codes land in internal/core/lint. Needs full git history
# (RD002 is append-only over committed history), which the local pre-push hook has.
lint-reviews:
	@bash scripts/check-reviews.sh

# Deterministic drift gate for the .abcd/development design record (first slice
# of internal/core/lint). Blocking: any record drift (stale tool names, dropped
# concepts, lifecycle or reference breakage) fails preflight and CI.
record-lint:
	@go run ./cmd/record-lint

# Pre-push gate (invoked by .githooks/pre-push): the same steps CI's check job
# runs — build, vet, test, and race-enabled internal tests — natively, plus the
# reviews-charter discipline. Host-native `go build` (not the cross-compiling
# build target) because it mirrors CI.
preflight: lint-reviews record-lint
	go build ./...
	go vet ./...
	go test ./...
	go test -race ./internal/...

clean:
	rm -rf $(BINDIR)
