#!/usr/bin/env bash
#
# This script provisions testing environment for local Hub using docker-compose
# and execute local Hub tests.
#
# It requires Docker to be installed.

# standard bash error handling
set -o nounset # treat unset variables as an error and exit immediately.
set -o errexit # exit immediately when a command fails.
set -E         # needs to be set if we want the ERR trap

CURRENT_DIR=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)
REPO_ROOT_DIR=$(cd "${CURRENT_DIR}/.." && pwd)

# shellcheck source=./hack/lib/utilities.sh
source "${CURRENT_DIR}/lib/utilities.sh" || { echo 'Cannot load CI utilities.'; exit 1; }
# shellcheck source=./hack/lib/const.sh
source "${CURRENT_DIR}/lib/const.sh" || { echo 'Cannot load constant values.'; exit 1; }

build_images() {
      export DOCKER_TAG=$RANDOM
      export DOCKER_REPOSITORY="local"
    
      cd ${REPO_ROOT_DIR} && make build-test-image-localhub
      cd ${REPO_ROOT_DIR} && make build-app-image-hub-js
      cd ${REPO_ROOT_DIR} && make build-app-image-secret-storage-backend
}

main() {
    if [[ "${BUILD_IMAGES:-"true"}" == "true" ]]; then
        build_images
    fi
    
    docker-compose -f "${REPO_ROOT_DIR}/test/localhub/docker-compose.yml" up --exit-code-from tests
}

main
