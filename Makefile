BINARY_NAME=obs-cli
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS := -s -w -X github.com/andy-neoaira/obs-cli/cmd.ldflagsVersion=$(VERSION)

install-hooks:
	git config core.hooksPath .githooks

build-all:
	GOOS=darwin GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o bin/darwin/${BINARY_NAME}
	GOOS=darwin GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o bin/darwin-arm64/${BINARY_NAME}
	GOOS=linux GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o bin/linux/${BINARY_NAME}
	GOOS=linux GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o bin/linux-arm64/${BINARY_NAME}
	GOOS=windows GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o bin/windows/${BINARY_NAME}.exe
	GOOS=windows GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o bin/windows-arm64/${BINARY_NAME}.exe

clean-all:
	go clean
	rm -f bin/darwin/${BINARY_NAME}
	rm -f bin/darwin-arm64/${BINARY_NAME}
	rm -f bin/linux/${BINARY_NAME}
	rm -f bin/linux-arm64/${BINARY_NAME}
	rm -f bin/windows/${BINARY_NAME}.exe
	rm -f bin/windows-arm64/${BINARY_NAME}.exe

test:
	go test ./...

test-search-content:
	go test ./pkg/actions -run TestSearchNotesContentWithOptions -v

test-coverage:
	go test ./... -coverprofile=coverage.out

license-check:
	./scripts/license-check.sh

	# Release automation
# Usage: make release VERSION=v0.2.2
release:
ifndef VERSION
	$(error VERSION is not set. Usage: make release VERSION=v0.2.2)
endif
	@echo "Starting release process for $(VERSION)..."
	@# Update version in root.go
	@perl -pi -e 's/Version: "v[0-9]+\.[0-9]+\.[0-9]+"/Version: "$(VERSION)"/' cmd/root.go
	@echo "✓ Updated version in root.go to $(VERSION)"
	@# Build all binaries
	@$(MAKE) build-all
	@echo "✓ Built binaries for all platforms"
	@# Git operations
	@git add cmd/root.go
	@git commit -m "chore: bump version to $(VERSION)"
	@git tag $(VERSION)
	@git push origin main
	@git push origin $(VERSION)
	@echo "✓ Release $(VERSION) complete!"

# Quick release (interactive version bump)
release-patch:
	@$(eval CURRENT_VERSION := $(shell grep 'Version:' cmd/root.go | sed 's/.*"v\([0-9]*\.[0-9]*\.[0-9]*\)".*/\1/'))
	@$(eval NEW_VERSION := $(shell echo $(CURRENT_VERSION) | awk -F. '{print "v" $$1 "." $$2 "." $$3+1}'))
	@$(MAKE) release VERSION=$(NEW_VERSION)

release-minor:
	@$(eval CURRENT_VERSION := $(shell grep 'Version:' cmd/root.go | sed 's/.*"v\([0-9]*\.[0-9]*\.[0-9]*\)".*/\1/'))
	@$(eval NEW_VERSION := $(shell echo $(CURRENT_VERSION) | awk -F. '{print "v" $$1 "." $$2+1 ".0"}'))
	@$(MAKE) release VERSION=$(NEW_VERSION)

release-major:
	@$(eval CURRENT_VERSION := $(shell grep 'Version:' cmd/root.go | sed 's/.*"v\([0-9]*\.[0-9]*\.[0-9]*\)".*/\1/'))
	@$(eval NEW_VERSION := $(shell echo $(CURRENT_VERSION) | awk -F. '{print "v" $$1+1 ".0.0"}'))
	@$(MAKE) release VERSION=$(NEW_VERSION)
