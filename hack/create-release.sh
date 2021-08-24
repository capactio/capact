#!/bin/bash

set -e

release::set_capact_images_in_charts() {
  local -r image_tag="$1"
  sed -i.bak "s/overrideTag: \"latest\"/overrideTag: \"${image_tag}\"/g" "deploy/kubernetes/charts/capact/values.yaml"
}

release::set_hub_manifest_source_branch() {
  local -r branch="$1"
  sed -i.bak "s/branch: main/branch: ${branch}/g" "deploy/kubernetes/charts/capact/charts/hub-public/values.yaml"
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
# GIT_TAG_REF - Git tag ref in format 'refs/tags/vx.y.z.'
[ -z "${RELEASE_VERSION}" ] && echo "Need to set RELEASE_VERSION" && exit 1;

SOURCE_BRANCH="$(git rev-parse --abbrev-ref HEAD)"
RELEASE_VERSION_MAJOR_MINOR="$(echo "${RELEASE_VERSION}" | sed -E 's/([0-9]+\.[0-9])\.[0-9]/\1/g')"
RELEASE_BRANCH="release-${RELEASE_VERSION_MAJOR_MINOR}"

main() {
  local -r capact_image_tag=$(git rev-parse --short HEAD | sed 's/.$//')

  git checkout -B "${RELEASE_BRANCH}"

  release::set_capact_images_in_charts "${capact_image_tag}"
  release::set_hub_manifest_source_branch "${RELEASE_BRANCH}"
  release::make_release_commit "${RELEASE_VERSION}" "${RELEASE_BRANCH}"
}

main
