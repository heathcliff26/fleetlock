###############################################################################
# BEGIN build-stage
# Compile the binary
FROM --platform=$BUILDPLATFORM docker.io/library/golang:1.22.2@sha256:c4fb952e712efd8f787bcd8e53fd66d1d83b7dc26adabc218e9eac1dbf776bdf AS build-stage

ARG BUILDPLATFORM
ARG TARGETARCH
ARG RELEASE_VERSION

WORKDIR /app

COPY vendor ./vendor
COPY go.mod go.sum ./
COPY cmd ./cmd
COPY pkg ./pkg
COPY hack ./hack

RUN --mount=type=bind,target=/app/.git,source=.git hack/build.sh

#
# END build-stage
###############################################################################

###############################################################################
# BEGIN test-stage
# Run the tests in the container
FROM build-stage AS test-stage

WORKDIR /app

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
