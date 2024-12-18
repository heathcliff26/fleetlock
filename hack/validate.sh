#!/bin/bash

set -e

script_dir="$(dirname "${BASH_SOURCE[0]}" | xargs realpath)"

echo "Check if source code is formatted"
make fmt
rc=0
git update-index --refresh && git diff-index --quiet HEAD -- || rc=1
if [ $rc -ne 0 ]; then
    echo "FATAL: Need to run \"make fmt\""
    exit 1
fi

echo "Check if the example manifests are up to date"
export TAG="latest"
export RELEASE_VERSION=""
"${script_dir}/manifests.sh"
rc=0
git diff-index -I "kubernetesVersion: v1.*" --quiet HEAD -- || rc=1
if [ $rc -ne 0 ]; then
    echo "FATAL: Need to run \"make manifests\" and update the examples with the result"
    exit 1
fi
