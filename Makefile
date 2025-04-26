SHELL := bash

REPOSITORY ?= localhost
CONTAINER_NAME ?= fleetlock
TAG ?= latest

# Build all binaries
build:

# Build fleetctl binary
build-fleetctl:
	hack/build.sh fleetctl

# Build fleetlock binary
build-fleetlock:
	hack/build.sh fleetlock

# Build the container image
image:
	podman build -t $(REPOSITORY)/$(CONTAINER_NAME):$(TAG) .

# Run unit-tests
test:
	go test -v -race -coverprofile=coverprofile.out -coverpkg "./pkg/..." ./cmd/... ./pkg/... ./tests/storage/...

# Run end-to-end tests
test-e2e:
	go test -count=1 -v ./tests/e2e/...

# Update dependencies
update-deps:
	hack/update-deps.sh

# Generate coverage profile
coverprofile:
	hack/coverprofile.sh

# Run linter
lint:
	golangci-lint
	golangci-lint run -v

# Format code
fmt:
	gofmt -s -w ./cmd ./pkg ./tests

# Generate manifests
manifests:
	hack/manifests.sh

# Validate that all generated files are up to date
validate:
	hack/validate.sh

# Scan code for vulnerabilities using gosec
gosec:
	gosec ./...

# Clean up generated files
clean:
	rm -rf bin manifests/release coverprofiles coverprofile.out logs tmp_fleetlock_image_fleetlock-e2e-*.tar

# Install golangci-lint
golangci-lint:
	go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest

# Show this help message
help:
	@echo "Available targets:"
	@echo ""
	@awk '/^#/{c=substr($$0,3);next}c&&/^[[:alpha:]][[:alnum:]_-]+:/{print substr($$1,1,index($$1,":")),c}1{c=0}' $(MAKEFILE_LIST) | column -s: -t
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
