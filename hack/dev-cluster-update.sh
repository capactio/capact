#!/usr/bin/env bash
#
# This script rebuilds Docker images from sources and upgrades Capact Helm chart installed on cluster.
#

# standard bash error handling
set -o nounset # treat unset variables as an error and exit immediately.
set -o errexit # exit immediately when a command fails.
set -E         # needs to be set if we want the ERR trap

CURRENT_DIR=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)
REPO_ROOT_DIR=$(cd "${CURRENT_DIR}/.." && pwd)
readonly CURRENT_DIR
readonly REPO_ROOT_DIR


# shellcheck source=./hack/lib/utilities.sh
source "${CURRENT_DIR}/lib/utilities.sh" || { echo 'Cannot load CI utilities.'; exit 1; }
# shellcheck source=./hack/lib/const.sh
source "${CURRENT_DIR}/lib/const.sh" || { echo 'Cannot load constant values.'; exit 1; }

main() {
    shout "Updating development local cluster..."

    export DOCKER_TAG=dev-$RANDOM
    export DOCKER_REPOSITORY="local"
    export BUILD_IMAGES="true"
    export REPO_DIR=$REPO_ROOT_DIR
    export CLUSTER_NAME=${CLUSTER_NAME:-${DEV_CLUSTER_NAME}}
    export CLUSTER_TYPE="kind"
    export PRINT_INSECURE_NOTES="true"
    capact::install

    shout "Development local cluster updated successfully."
}

main
