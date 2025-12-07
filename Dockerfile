###############################################################################
# BEGIN build-stage
# Compile the binary
FROM --platform=$BUILDPLATFORM docker.io/library/golang:1.25.5 AS build-stage

ARG BUILDPLATFORM
ARG TARGETARCH

WORKDIR /app

COPY . ./

RUN GOOS=linux GOARCH="${TARGETARCH}" hack/build.sh fleetlock

#
# END build-stage
###############################################################################

###############################################################################
# BEGIN final-stage
# Create final docker image
FROM docker.io/library/alpine:3.23.0@sha256:51183f2cfa6320055da30872f211093f9ff1d3cf06f39a0bdb212314c5dc7375 AS final-stage

LABEL   org.opencontainers.image.authors="heathcliff@heathcliff.eu"

WORKDIR /

COPY --from=build-stage /app/bin/fleetlock /

USER 1001

ENTRYPOINT ["/fleetlock"]

#
# END final-stage
###############################################################################
