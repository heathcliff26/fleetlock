###############################################################################
# BEGIN build-stage
# Compile the binary
FROM --platform=$BUILDPLATFORM docker.io/library/golang:1.22.5@sha256:fcae9e0e7313c6467a7c6632ebb5e5fab99bd39bd5eb6ee34a211353e647827a AS build-stage

ARG BUILDPLATFORM
ARG TARGETARCH
ARG RELEASE_VERSION

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
# BEGIN combine-stage
# Combine all outputs, to enable single layer copy for the final image
FROM scratch AS combine-stage

COPY --from=build-stage /app/bin/fleetlock /

COPY --from=build-stage /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

#
# END combine-stage
###############################################################################

###############################################################################
# BEGIN final-stage
# Create final docker image
FROM scratch AS final-stage

WORKDIR /

COPY --from=combine-stage / /

USER 1001

ENTRYPOINT ["/fleetlock"]

#
# END final-stage
###############################################################################
