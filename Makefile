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

compatibility-check:
	./scripts/compatibility-check.sh

test-search-content:
	go test ./pkg/actions -run TestSearchNotesContentWithOptions -v

test-coverage:
	go test ./... -coverprofile=coverage.out

format-check:
	./scripts/gofmt-check.sh

naming-check:
	./scripts/naming-check.sh
	./scripts/naming-check.sh --self-test

schema-check:
	./scripts/schema-check.sh

coverage-check:
	./scripts/coverage-check.sh

build-check:
	./scripts/build-check.sh

rc-smoke:
	./scripts/rc-smoke.sh

license-check:
	./scripts/license-check.sh

skill-lint:
	./scripts/lint-skills.sh

skill-evals:
	./scripts/run-skill-evals.sh

release-check: format-check naming-check compatibility-check
	go vet ./...
	go test ./...
	go test -race ./...
	$(MAKE) coverage-check
	$(MAKE) schema-check
	$(MAKE) build-check
	$(MAKE) license-check
	$(MAKE) skill-lint
	$(MAKE) skill-evals
	$(MAKE) rc-smoke

	# Release automation
# Usage: make release VERSION=v1.0.0-rc.1
release:
ifndef VERSION
	$(error VERSION is not set. Usage: make release VERSION=v1.0.0-rc.1)
endif
	@SKILL_EVAL_CLI_VERSION="$(VERSION)" $(MAKE) release-check
	@$(MAKE) build-all
	@echo "✓ Built binaries for all platforms"
	@echo "✓ Candidate $(VERSION) validated; tag creation and publication require separate approval"
