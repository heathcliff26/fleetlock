#!/bin/bash

set -e

base_dir="$(dirname "${BASH_SOURCE[0]}" | xargs realpath | xargs dirname)"

pushd "${base_dir}" > /dev/null

if [ -z "${RELEASE_VERSION}" ]; then
    git_version="$(git describe --tags --always --dirty)"
    RELEASE_VERSION="${git_version}"
fi

export RELEASE_VERSION

popd > /dev/null
