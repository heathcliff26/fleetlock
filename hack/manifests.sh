#!/bin/bash

set -e

base_dir="$(dirname "${BASH_SOURCE[0]}" | xargs realpath)/.."

export REPOSITORY="${REPOSITORY:-ghcr.io/heathcliff26}"
export TAG="${TAG:-latest}"
export FLEETLOCK_NAMESPACE="${FLEETLOCK_NAMESPACE:-fleetlock}"

output_dir="${base_dir}/manifests/release"
deployment_file="${output_dir}/deployment.yaml"

if [[ "${RELEASE_VERSION}" != "" ]] && [[ "${TAG}" == "latest" ]]; then
    TAG="${RELEASE_VERSION}"
fi

[ ! -d "${output_dir}" ] && mkdir "${output_dir}"

echo "Creating manifest from helm chart"
cat > "${deployment_file}" <<EOF
---
apiVersion: v1
kind: Namespace
metadata:
EOF
echo "  name: ${FLEETLOCK_NAMESPACE}" >> "${deployment_file}"

helm template "${base_dir}/manifests/helm" \
    --debug \
    --set fullnameOverride=fleetlock \
    --set image.repository="${REPOSITORY}/fleetlock" \
    --set image.tag="${TAG}" \
    --set ingress.enabled=true \
    --name-template fleetlock \
    --namespace "${FLEETLOCK_NAMESPACE}" \
    | grep -v '# Source: fleetlock/templates/' \
    | grep -v 'helm.sh/chart: fleetlock' \
    | grep -v 'app.kubernetes.io/managed-by: Helm' > "${deployment_file}.tmp"

echo "Patching version labels"
yq -i '
    (
        .metadata.labels."app.kubernetes.io/version" = env(TAG)
    ) |
    (
        select(.spec.template.metadata.labels)
        | .spec.template.metadata.labels."app.kubernetes.io/version" = env(TAG)
    ) // .
' "${deployment_file}.tmp"

cat "${deployment_file}.tmp" >> "${deployment_file}"
rm "${deployment_file}.tmp"

echo "Wrote manifests to ${output_dir}"

if [ "${TAG}" == "latest" ]; then
    echo "Tag is latest, syncing manifests with examples"
    cp "${output_dir}"/*.yaml "${base_dir}/examples/"
fi
