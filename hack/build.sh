#!/bin/bash

set -e

base_dir="$(dirname "${BASH_SOURCE[0]}" | xargs realpath)/.."

bin_dir="${base_dir}/bin"

GOOS="${GOOS:-$(go env GOOS)}"
GOARCH="${GOARCH:-$(go env GOARCH)}"
CC="${CC:-$(go env CC)}"

GO_LD_FLAGS="${GO_LD_FLAGS:-"-s -extldflags=-static"}"
GO_TAGS="${GO_TAGS:-"sqlite_omit_load_extension"}"

if [ "${RELEASE_VERSION}" != "" ]; then
    echo "Building release version ${RELEASE_VERSION}"
    GO_LD_FLAGS+=" -X github.com/heathcliff26/fleetlock/pkg/version.version=${RELEASE_VERSION}"
fi

if [ "${GO_CC_ZIG}" == "true" ] && [ "${CC}" == "$(go env CC)" ]; then
    echo "Using zig as C compiler"
    CC="zig cc"
    if [ "${GOARCH}" == "arm64" ]; then
        CC="${CC} -target aarch64-linux"
    else
        CC="${CC} -target x86_64-linux"
    fi
fi

output_name="${bin_dir}/fleetlock"
if [ "${GOOS}" == "windows" ]; then
    output_name="${output_name}.exe"
fi

pushd "${base_dir}" >/dev/null

echo "Building $(basename "${output_name}")"
GOOS="${GOOS}" GOARCH="${GOARCH}" CGO_ENABLED=1 CC="${CC}" go build -ldflags="${GO_LD_FLAGS}" -tags "${GO_TAGS}" -o "${output_name}" ./cmd/...
