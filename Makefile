SHELL := bash

REPOSITORY ?= localhost
CONTAINER_NAME ?= fleetlock
TAG ?= latest

build: build-fleetctl build-fleetlock

build-fleetctl:
	hack/build.sh fleetctl

build-fleetlock:
	hack/build.sh fleetlock

image:
	podman build -t $(REPOSITORY)/$(CONTAINER_NAME):$(TAG) .

test:
	go test -v -race -coverprofile=coverprofile.out -coverpkg "./pkg/..." ./cmd/... ./pkg/... ./tests/storage/...

test-e2e:
	go test -count=1 -v ./tests/e2e/...

update-deps:
	hack/update-deps.sh

coverprofile:
	hack/coverprofile.sh

lint: golangci-lint
	golangci-lint run -v

fmt:
	gofmt -s -w ./cmd ./pkg ./tests

manifests:
	hack/manifests.sh

validate:
	hack/validate.sh

clean:
	rm -rf bin manifests/release coverprofiles coverprofile.out logs tmp_fleetlock_image_fleetlock-e2e-*.tar

golangci-lint:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

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
	clean \
	golangci-lint \
	$(NULL)
