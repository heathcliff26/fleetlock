SHELL := bash

REPOSITORY ?= localhost
CONTAINER_NAME ?= fleetlock
TAG ?= latest

default: build

build:
	hack/build.sh

image:
	podman build -t $(REPOSITORY)/$(CONTAINER_NAME):$(TAG) .

test:
	go test -v -race ./...

update-deps:
	hack/update-deps.sh

coverprofile:
	hack/coverprofile.sh

lint:
	golangci-lint run -v

.PHONY: \
	default \
	build \
	image \
	test \
	update-deps \
	coverprofile \
	lint \
	$(NULL)
