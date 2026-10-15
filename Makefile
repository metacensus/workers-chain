.PHONY: help check proto mocks gen lint-proto-check docker-build test fmt vet \
	install-tools-gen install-tools-lint lint vuln release latest \
	fmt-check hooks generated-paths

.DEFAULT_GOAL := help

# Read out of go.mod so no copy can drift from what CI's setup-go uses.
GEN_TOOLCHAIN := go$(shell sed -n 's/^go //p' go.mod | head -1)
ifeq ($(GEN_TOOLCHAIN),go)
$(error could not read the Go version from go.mod; refusing to run unpinned)
endif
TOOL_PATH := PATH="$(CURDIR)/.bin:$$PATH"

GOLANGCI_LINT_VERSION := v2.14.0
GOVULNCHECK := golang.org/x/vuln/cmd/govulncheck@v1.8.0

help:
	@awk -F' — ' '/^## /{ sub(/^## /, ""); printf "  make %-22s %s\n", $$1, $$2 }' $(MAKEFILE_LIST)

## install-tools-gen — install buf and protoc plugins into .bin
install-tools-gen:
	@mkdir -p .bin
	@GOTOOLCHAIN=$(GEN_TOOLCHAIN) GOBIN=$(CURDIR)/.bin go install tool

## install-tools-lint — install golangci-lint into .bin
install-tools-lint: install-tools-gen
	@test -x $(CURDIR)/.bin/golangci-lint || \
		GOTOOLCHAIN=$(GEN_TOOLCHAIN) GOBIN=$(CURDIR)/.bin go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION)

## proto — regenerate protobuf code
proto: install-tools-gen
	@$(TOOL_PATH) buf generate

## mocks — regenerate mocks
mocks: install-tools-gen
	@GOWORK=off GOTOOLCHAIN=$(GEN_TOOLCHAIN) go generate -run mockery

## gen — proto + mocks
gen: proto mocks

## lint-proto-check — buf lint and format check
lint-proto-check: install-tools-gen
	@$(TOOL_PATH) buf lint
	@$(TOOL_PATH) buf format -d --exit-code

## fmt — rewrite with gofmt (skip gen/)
fmt:
	@gofmt -w $$(git ls-files '*.go' | grep -v '^gen/')

## fmt-check — fail if gofmt would change anything (skip gen/)
fmt-check:
	@test -z "$$(gofmt -l $$(git ls-files '*.go' | grep -v '^gen/') | tee /dev/stderr)" || (echo 'gofmt dirty; run make fmt' && exit 1)

## vet — go vet
vet:
	@GOTOOLCHAIN=$(GEN_TOOLCHAIN) go vet ./...

## test — unit tests
test:
	@GOTOOLCHAIN=$(GEN_TOOLCHAIN) go test ./...

## lint — proto lint, fmt check, vet, golangci-lint
lint: lint-proto-check install-tools-lint
	@test -z "$$(gofmt -l $$(git ls-files '*.go' | grep -v '^gen/') | tee /dev/stderr)"
	@GOTOOLCHAIN=$(GEN_TOOLCHAIN) go vet ./...
	@GOTOOLCHAIN=$(GEN_TOOLCHAIN) $(TOOL_PATH) golangci-lint run ./cmd/... ./lib/...

## vuln — govulncheck over the module
vuln:
	@GOTOOLCHAIN=$(GEN_TOOLCHAIN) go run $(GOVULNCHECK) ./...

## check — proto/mocks freshness, helm lint, lint, unit tests
check: lint
	@$(TOOL_PATH) buf generate
	@git diff --exit-code -- gen proto || (echo 'proto gen dirty; run make proto' && exit 1)
	@GOWORK=off GOTOOLCHAIN=$(GEN_TOOLCHAIN) go generate -run mockery
	@git diff --exit-code -- mocks || (echo 'mocks dirty; run make mocks' && exit 1)
	@bash -c 'for chart in helm/fabric/*; do \
		if [ -f "$$chart/Chart.yaml" ]; then \
			echo "  Linting $$chart..."; \
			if [ -f "$$chart/values.example.yaml" ]; then \
				helm lint "$$chart" -f "$$chart/values.example.yaml"; \
			else \
				helm lint "$$chart"; \
			fi; \
		fi; \
	done'
	@GOTOOLCHAIN=$(GEN_TOOLCHAIN) go test ./...

## docker-build — build the three worker images for this machine
docker-build:
	docker build -f cmd/workers/infra/Dockerfile -t metacensus/worker-infra:local .
	docker build -f cmd/workers/fabric/Dockerfile -t metacensus/worker-fabric:local .
	docker build -f cmd/workers/shard/Dockerfile -t metacensus/worker-fabric-shard:local .

## release — tag and push VERSION=x.y.z or the next TYPE=major|minor|patch; prompts unless YES=1
release:
	@set -e; \
	git fetch -q --tags origin main; \
	VERSION=$$(./scripts/version.sh "$(VERSION)" "$(TYPE)"); \
	TAG="v$$VERSION"; \
	if git ls-remote --exit-code --tags origin "refs/tags/$$TAG" >/dev/null 2>&1; then \
		echo "Error: Tag $$TAG already exists on origin"; \
		exit 1; \
	fi; \
	if [ -n "$$(git status --porcelain)" ]; then \
		echo "Error: working tree is dirty; commit or clean it before releasing"; \
		exit 1; \
	fi; \
	if ! git merge-base --is-ancestor HEAD origin/main; then \
		echo "Error: HEAD is not on origin/main; refusing to tag"; \
		exit 1; \
	fi; \
	echo "About to tag and push $$TAG, which publishes the three worker images."; \
	if [ "$(YES)" != "1" ]; then \
		if [ -t 0 ]; then \
			printf "Proceed? [y/N] "; read -r reply; \
			case "$$reply" in y|Y|yes|YES) ;; *) echo "Aborted."; exit 1;; esac; \
		else \
			echo "Refusing to release non-interactively; pass YES=1 if you mean it."; \
			exit 1; \
		fi; \
	fi; \
	git tag -a "$$TAG" -m "Release $$VERSION" && \
	git push origin "$$TAG" && \
	echo "Released: $$TAG"

## latest — print the most recent version tag
latest:
	@bash -c '. scripts/version.sh && v=$$(get_latest_version) && [ -n "$$v" ] && echo v$$v || true'

## hooks — opt in to the pre-commit and pre-push hooks (opt out: git config --unset core.hooksPath)
hooks:
	git config core.hooksPath .githooks
	@echo "core.hooksPath set to .githooks"

# What `gen` writes, named once so `check` and the pre-commit hook agree.
generated-paths:
	@echo gen proto/README.md mocks
