.PHONY: help check proto mocks gen lint-proto lint-proto-check docker-build test fmt vet

help:
	@echo "Targets:"
	@echo "  make check         # proto lint, gen freshness, helm lint, fmt, vet, unit tests"
	@echo "  make proto         # regenerate protobuf code"
	@echo "  make mocks         # regenerate mocks"
	@echo "  make gen           # proto + mocks"
	@echo "  make docker-build  # build the three worker images"
	@echo "  make test          # unit tests"

# Tooling pinned via go.mod tool directives (see install-tools-gen pattern from infra).
GEN_TOOLCHAIN := go$(shell sed -n 's/^go //p' go.mod | head -1)
TOOL_PATH := PATH="$(CURDIR)/.bin:$$PATH"

GOLANGCI_LINT_VERSION := v2.5.0

install-tools-gen:
	@mkdir -p .bin
	@GOTOOLCHAIN=$(GEN_TOOLCHAIN) GOBIN=$(CURDIR)/.bin go install tool

install-tools-lint: install-tools-gen
	@test -x $(CURDIR)/.bin/golangci-lint || \
		GOTOOLCHAIN=$(GEN_TOOLCHAIN) GOBIN=$(CURDIR)/.bin go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION)

proto: install-tools-gen
	@$(TOOL_PATH) buf generate

mocks: install-tools-gen
	@GOWORK=off GOTOOLCHAIN=$(GEN_TOOLCHAIN) go generate -run mockery

gen: proto mocks

lint-proto-check: install-tools-gen
	@$(TOOL_PATH) buf lint
	@$(TOOL_PATH) buf format -d --exit-code

fmt:
	@gofmt -w $$(git ls-files '*.go' | grep -v '^gen/')

vet:
	@GOTOOLCHAIN=$(GEN_TOOLCHAIN) go vet ./...

test:
	@GOTOOLCHAIN=$(GEN_TOOLCHAIN) go test ./...

check: lint-proto-check install-tools-lint
	@$(TOOL_PATH) buf generate
	@git diff --exit-code -- gen proto || (echo 'proto gen dirty; run make proto' && exit 1)
	@GOWORK=off GOTOOLCHAIN=$(GEN_TOOLCHAIN) go generate -run mockery
	@git diff --exit-code -- mocks || (echo 'mocks dirty; run make mocks' && exit 1)
	@for chart in helm/fabric/*; do \
		if [ -f "$$chart/Chart.yaml" ]; then \
			echo "  Linting $$chart..."; \
			if [ -f "$$chart/values.example.yaml" ]; then \
				helm lint "$$chart" -f "$$chart/values.example.yaml"; \
			else \
				helm lint "$$chart"; \
			fi; \
		fi; \
	done
	@test -z "$$(gofmt -l $$(git ls-files '*.go' | grep -v '^gen/') | tee /dev/stderr)"
	@GOTOOLCHAIN=$(GEN_TOOLCHAIN) go vet ./...
	@GOTOOLCHAIN=$(GEN_TOOLCHAIN) $(TOOL_PATH) golangci-lint run ./cmd/... ./lib/...
	@GOTOOLCHAIN=$(GEN_TOOLCHAIN) go test ./...

docker-build:
	docker build -f cmd/workers/infra/Dockerfile -t metacensus/infra-worker:local .
	docker build -f cmd/workers/fabric/Dockerfile -t metacensus/fabric-worker:local .
	docker build -f cmd/workers/shard/Dockerfile -t metacensus/shard-worker:local .
