#!/usr/bin/env bash
#
# This script provisions testing environment for local Hub using docker-compose
# and execute tests.
#
# It requires Docker to be installed.

# standard bash error handling
set -o nounset # treat unset variables as an error and exit immediately.
set -o errexit # exit immediately when a command fails.

CURRENT_DIR=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)
REPO_ROOT_DIR=$(cd "${CURRENT_DIR}/.." && pwd)

# shellcheck source=./hack/lib/utilities.sh
source "${CURRENT_DIR}/lib/utilities.sh" || { echo 'Cannot load CI utilities.'; exit 1; }
# shellcheck source=./hack/lib/const.sh
source "${CURRENT_DIR}/lib/const.sh" || { echo 'Cannot load constant values.'; exit 1; }

build_images() {
  export DOCKER_TAG=$RANDOM
  export DOCKER_REPOSITORY="local"

  local make_targets=("build-test-image-local-hub" "build-app-image-hub-js" "build-app-image-secret-storage-backend")
  for make_target in "${make_targets[@]}"
  do
    cd "${REPO_ROOT_DIR}" && make "${make_target}"
  done
}

main() {
  if [[ "${BUILD_IMAGES:-"true"}" == "true" ]]; then
    build_images
  fi
    
  docker-compose -f "${REPO_ROOT_DIR}/test/local-hub/docker-compose.yml" up --exit-code-from tests --force-recreate
}

cleanup() {
  docker-compose -f "${REPO_ROOT_DIR}/test/local-hub/docker-compose.yml" down
}

trap cleanup EXIT

main
