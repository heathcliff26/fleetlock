#!/bin/bash

set -e

base_dir="$(dirname "${BASH_SOURCE[0]}" | xargs realpath | xargs dirname)"
dist_dir="${base_dir}/dist"
bin_dir="${base_dir}/bin"
name="fleetctl"

echo "Checking if goreleaser is installed:"
if command -v goreleaser &>/dev/null; then
    echo "goreleaser is installed"
    goreleaser="$(command -v goreleaser)"
else
    echo "goreleaser is not installed, downloading latest version..."
    LATEST="$(curl -sf https://goreleaser.com/static/latest)"
    [ -e "${bin_dir}" ] || mkdir "${bin_dir}"
    arch="$(uname -m)"
    [ "${arch}" != "aarch64" ] || arch="arm64"
    curl -SL -o "${bin_dir}/goreleaser.tar.gz" "https://github.com/goreleaser/goreleaser/releases/download/${LATEST}/goreleaser_$(uname -s)_${arch}.tar.gz"
    tar -xzf "${bin_dir}/goreleaser.tar.gz" -C "${bin_dir}" goreleaser
    rm "${bin_dir}/goreleaser.tar.gz"
    goreleaser="${bin_dir}/goreleaser"
fi

echo "Building releaser artifacts with goreleaser"
${goreleaser} release --skip=announce,archive,publish,validate --clean

echo "Moving release artifacts to top level of dist directory"
artifacts="$(cat "${dist_dir}/artifacts.json" | jq -r -c '.[]')"
echo "${artifacts}" | while read -r artifact; do
    if ! [[ "$(echo "${artifact}" | jq -r '.name')" =~ ^${name}(\.exe)?$ ]]; then
        continue
    fi

    path="${base_dir}/$(echo "${artifact}" | jq -r '.path')"
    goarch="$(echo "${artifact}" | jq -r '.goarch')"
    ext="$(echo "${artifact}" | jq -r '.extra.Ext')"

    mv "${path}" "${dist_dir}/${name}-${goarch}${ext}"

    path="$(dirname "${path}")"
    rm -r "${path}"
done


echo "Cleaning up dist directory"
rm -r "${dist_dir}/artifacts.json" "${dist_dir}/config.yaml" "${dist_dir}/metadata.json"
