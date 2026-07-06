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

.PHONY: build test vet clean preflight

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

# Pre-push gate (invoked by .githooks/pre-push): the same steps CI's check job
# runs — build, vet, test, and race-enabled internal tests — natively. Host-native
# `go build` (not the cross-compiling build target) because it mirrors CI.
preflight:
	go build ./...
	go vet ./...
	go test ./...
	go test -race ./internal/...

clean:
	rm -rf $(BINDIR)
