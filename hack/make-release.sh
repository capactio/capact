#!/bin/bash

set -e

CURRENT_DIR=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)
REPO_ROOT_DIR=$(cd "${CURRENT_DIR}/.." && pwd)
readonly CURRENT_DIR
readonly REPO_ROOT_DIR


release::update_helm_charts_version() {
  local -r release_version="$1"
  local -r deploy_dir="${REPO_ROOT_DIR}/deploy/kubernetes/charts"

  for d in "${deploy_dir}"/*/ ; do
    sed -i.bak "s/^version: .*/version: ${release_version}/g" "${d}/Chart.yaml"
  done
}

release::make_prepare_release_commit() {
  local -r version="$1"
  local -r branch="$2"

  git add .
  git commit -m "Prepare ${version} release"
  git push origin "${branch}"
}

release::set_capact_images_in_charts() {
  local -r image_tag="$1"
  sed -E -i.bak "s/overrideTag: \".+\"/overrideTag: \"${image_tag}\"/g" "${REPO_ROOT_DIR}/deploy/kubernetes/charts/capact/values.yaml"
}

release::set_hub_manifest_source_branch() {
  local -r branch="$1"
  sed -E -i.bak "s/branch: .+/branch: ${branch}/g" "${REPO_ROOT_DIR}/deploy/kubernetes/charts/capact/charts/hub-public/values.yaml"
}

release::make_release_commit() {
  local -r version="$1"
  local -r release_branch="$2"
  local -r tag="v${version}"

  git add .
  git commit -m "Set fixed Capact image tag and Populator source branch"
  git tag "${tag}"
  git push origin "${release_branch}"
  git push origin "${tag}"
}

# required inputs:
# RELEASE_VERSION - new version in SemVer format: x.y.z
[ -z "${RELEASE_VERSION}" ] && echo "Need to set RELEASE_VERSION" && exit 1;

SOURCE_BRANCH="$(git rev-parse --abbrev-ref HEAD)"
RELEASE_VERSION_MAJOR_MINOR="$(echo "${RELEASE_VERSION}" | sed -E 's/([0-9]+\.[0-9])\.[0-9]/\1/g')"
RELEASE_BRANCH="release-${RELEASE_VERSION_MAJOR_MINOR}"

main() {
  release::update_helm_charts_version "${RELEASE_VERSION}"
  release::make_prepare_release_commit "${RELEASE_VERSION}" "${SOURCE_BRANCH}"

  local -r revision=$(git rev-parse --short HEAD)
  local -r capact_image_tag="${revision:0:7}"
  git checkout -B "${RELEASE_BRANCH}"

  release::set_capact_images_in_charts "${capact_image_tag}"
  release::set_hub_manifest_source_branch "${RELEASE_BRANCH}"
  release::make_release_commit "${RELEASE_VERSION}" "${RELEASE_BRANCH}"
}

main
