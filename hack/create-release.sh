#!/bin/bash

set -e

release::append_branch_to_ci() {
  local -r release_branch="$1"
  yq e -i ".on.push.branches |= . + [\"${release_branch}\"]" ".github/workflows/branch-build.yaml"
}

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

  git add .
  git commit -m "Set fixed Capact image tag and Populator source branch"
  git tag v${version} HEAD

  git push origin ${release_branch}
  git push origin v${version}
}

# required inputs:
# RELEASE_VERSION - new version in semver format: x.y.z

RELEASE_VERSION="${RELEASE_VERSION}"

SOURCE_BRANCH="$(git rev-parse --abbrev-ref HEAD)"
RELEASE_VERSION_MAJOR_MINOR="$(echo "${RELEASE_VERSION}" | sed -E 's/([0-9]+\.[0-9])\.[0-9]/\1/g')"
RELEASE_BRANCH="release-${RELEASE_VERSION_MAJOR_MINOR}"

main() {
  git config --global user.email "bot@capact.io"
  git config --global user.name "Capact Bot"

  if [ "${SOURCE_BRANCH}" == "main" ]; then
    release::append_branch_to_ci "${RELEASE_BRANCH}"
  fi
  
  release::update_helm_charts_version "${RELEASE_VERSION}"
  release::update_cli_version "${RELEASE_VERSION}"
  release::make_prepare_release_commit "${RELEASE_VERSION}" "${SOURCE_BRANCH}"
  
  if [ "${SOURCE_BRANCH}" == "main" ]; then
    git checkout -b "${RELEASE_BRANCH}"
  fi
  
  local -r capact_image_tag=$(git rev-parse --short HEAD | sed 's/.$//')
  release::set_capact_images_in_charts "${capact_image_tag}"
  release::set_hub_manifest_source_branch "${RELEASE_BRANCH}"
  
  release::make_release_commit "${RELEASE_VERSION}" "${RELEASE_BRANCH}"
}

main