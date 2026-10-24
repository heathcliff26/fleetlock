###############################################################################
# BEGIN build-stage
# Compile the binary
FROM --platform=$BUILDPLATFORM docker.io/library/golang:1.27.3 AS build-stage

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
FROM docker.io/library/alpine:3.24.3@sha256:be2656061b536c1883c4034790e21d0a09c5aaaf51968b89368a3c0371ea2c9a AS final-stage

LABEL   org.opencontainers.image.authors="heathcliff@heathcliff.eu"

WORKDIR /

COPY --from=build-stage /app/bin/fleetlock /

USER nobody:nobody

ENTRYPOINT ["/fleetlock"]

#
# END final-stage
###############################################################################
