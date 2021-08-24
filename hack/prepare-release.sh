#!/bin/bash

set -e

release::update_helm_charts_version() {
  local -r release_version="$1"
  local -r deploy_dir=deploy/kubernetes/charts

  for d in "${deploy_dir}"/*/ ; do
    sed -i.bak "s/^version: .*/version: ${release_version}/g" "${d}/Chart.yaml"
  done
}

release::update_cli_version() {
  local -r release_version="$1"
  sed -i.bak "s/Version = .*/Version = \"${release_version}\"/g" "internal/cli/info.go"
}

release::make_prepare_release_commit() {
  local -r version="$1"
  local -r branch="$2"

  git add .
  git commit -m "Prepare ${version} release"
  git push origin "${branch}"
}

# required inputs:
# RELEASE_VERSION - new version in semver format: x.y.z
[ -z "${RELEASE_VERSION}" ] && echo "Need to set RELEASE_VERSION" && exit 1;

SOURCE_BRANCH="$(git rev-parse --abbrev-ref HEAD)"

main() {
  release::update_helm_charts_version "${RELEASE_VERSION}"
  release::update_cli_version "${RELEASE_VERSION}"
  release::make_prepare_release_commit "${RELEASE_VERSION}" "${SOURCE_BRANCH}"
}

main
