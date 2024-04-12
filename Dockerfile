###############################################################################
# BEGIN build-stage
# Compile the binary
FROM --platform=$BUILDPLATFORM ghcr.io/heathcliff26/go-fyne-ci:1.22.2 AS build-stage

ARG BUILDPLATFORM
ARG TARGETARCH
ARG RELEASE_VERSION

# Cross-compile using zig
ENV GO_CC_ZIG=true

WORKDIR /app

COPY vendor ./vendor
COPY go.mod go.sum ./
COPY cmd ./cmd
COPY pkg ./pkg
COPY hack ./hack

RUN --mount=type=bind,target=/app/.git,source=.git GOOS=linux GOARCH="${TARGETARCH}" hack/build.sh

#
# END build-stage
###############################################################################

###############################################################################
# BEGIN test-stage
# Run the tests in the container
FROM docker.io/library/golang:1.22.2@sha256:450e3822c7a135e1463cd83e51c8e2eb03b86a02113c89424e6f0f8344bb4168 AS test-stage

WORKDIR /app

COPY --from=build-stage /app /app

RUN go test -v ./...

#
# END test-stage
###############################################################################

###############################################################################
# BEGIN final-stage
# Create final docker image
FROM scratch AS final-stage

WORKDIR /

COPY --from=test-stage /app/bin/fleetlock /

USER 1001

ENTRYPOINT ["/fleetlock"]

#
# END final-stage
###############################################################################
