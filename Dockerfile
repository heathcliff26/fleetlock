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
FROM docker.io/library/alpine:3.24.4@sha256:f59abe656b692d901e184be3445e084786602108599d2a51d229c0f5a48c71bf AS final-stage

LABEL   org.opencontainers.image.authors="heathcliff@heathcliff.eu"

WORKDIR /

COPY --from=build-stage /app/bin/fleetlock /

USER nobody:nobody

ENTRYPOINT ["/fleetlock"]

#
# END final-stage
###############################################################################
