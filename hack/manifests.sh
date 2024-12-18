#!/bin/bash

set -e

base_dir="$(dirname "${BASH_SOURCE[0]}" | xargs realpath)/.."

export REPOSITORY="${REPOSITORY:-ghcr.io/heathcliff26}"
export TAG="${TAG:-latest}"
export FLEETLOCK_NAMESPACE="${FLEETLOCK_NAMESPACE:-fleetlock}"

output_dir="${base_dir}/manifests/release"

if [[ "${RELEASE_VERSION}" != "" ]] && [[ "${TAG}" == "latest" ]]; then
    TAG="${RELEASE_VERSION}"
fi

[ ! -d "${output_dir}" ] && mkdir "${output_dir}"

echo "Creating deployment.yaml"
envsubst < "${base_dir}/manifests/base/deployment.yaml.template" > "${output_dir}/deployment.yaml"

echo "Wrote manifests to ${output_dir}"

if [ "${TAG}" == "latest" ]; then
    echo "Tag is latest, syncing manifests with examples"
    cp "${output_dir}"/*.yaml "${base_dir}/examples/"
fi
