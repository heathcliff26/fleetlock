#!/bin/bash

set -e

script_dir="$(dirname "${BASH_SOURCE[0]}" | xargs realpath)/.."

pushd "${script_dir}" >/dev/null

OUT_DIR="${script_dir}/coverprofiles"

if [ ! -d "${OUT_DIR}" ]; then
    mkdir "${OUT_DIR}"
fi

go test -coverprofile="${OUT_DIR}/cover.out" -coverpkg "./..." "./..."
go tool cover -html "${OUT_DIR}/cover.out" -o "${OUT_DIR}/index.html"
rm "${OUT_DIR}/cover.out"

popd >/dev/null
