#!/bin/bash

set -e

base_dir="$(dirname "${BASH_SOURCE[0]}" | xargs realpath)/.."

bin_dir="${base_dir}/bin"

target="${1}"

GOOS="${GOOS:-$(go env GOOS)}"
GOARCH="${GOARCH:-$(go env GOARCH)}"

GO_LD_FLAGS="${GO_LD_FLAGS:-"-s"}"

output_name="${bin_dir}/${target}"

if [ "${2}" != "" ]; then
    output_name="${bin_dir}/${2}"
fi

if [ "${GOOS}" == "windows" ]; then
    output_name="${output_name}.exe"
fi

pushd "${base_dir}" >/dev/null

echo "Building $(basename "${output_name}")"
GOOS="${GOOS}" GOARCH="${GOARCH}" CGO_ENABLED=0 go build -ldflags="${GO_LD_FLAGS}" -o "${output_name}" "./cmd/${target}"
