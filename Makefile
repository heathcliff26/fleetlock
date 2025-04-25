SHELL := bash

REPOSITORY ?= localhost
CONTAINER_NAME ?= fleetlock
TAG ?= latest

build: ## Build all binaries

build-fleetctl: ## Build fleetctl binary
	hack/build.sh fleetctl

build-fleetlock: ## Build fleetlock binary
	hack/build.sh fleetlock

image: ## Build the container image
	podman build -t $(REPOSITORY)/$(CONTAINER_NAME):$(TAG) .

test: ## Run unit-tests
	go test -v -race -coverprofile=coverprofile.out -coverpkg "./pkg/..." ./cmd/... ./pkg/... ./tests/storage/...

test-e2e: ## Run end-to-end tests
	go test -count=1 -v ./tests/e2e/...

update-deps: ## Update dependencies
	hack/update-deps.sh

coverprofile: ## Generate coverage profile
	hack/coverprofile.sh

lint: ## Run linter
	golangci-lint
	golangci-lint run -v

fmt: ## Format code
	gofmt -s -w ./cmd ./pkg ./tests

manifests: ## Generate manifests
	hack/manifests.sh

validate: ## Validate that all generated files are up to date
	hack/validate.sh

gosec: ## Scan code for vulnerabilities using gosec
	gosec ./...

clean: ## Clean up generated files
	rm -rf bin manifests/release coverprofiles coverprofile.out logs tmp_fleetlock_image_fleetlock-e2e-*.tar

golangci-lint: ## Install golangci-lint
	go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest

help: ## Show this help message
	@echo "Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?##' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "%-20s %s\n", $$1, $$2}'
	@echo ""
	@echo "Run 'make <target>' to execute a specific target."

.PHONY: \
	default \
	build \
	image \
	test \
	test-e2e \
	update-deps \
	coverprofile \
	lint \
	fmt \
	manifests \
	validate \
	gosec \
	clean \
	golangci-lint \
	help \
	$(NULL)
