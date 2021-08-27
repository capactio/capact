#!/usr/bin/env bash
#
# This script deletes kind cluster. If cluster name is not provided, it uses development cluster name.
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

    export CLUSTER_TYPE=${CLUSTER_TYPE:-"kind"}
    export CLUSTER_NAME=${CLUSTER_NAME:-${DEV_CLUSTER_NAME}}
    export REPO_DIR=$REPO_ROOT_DIR
    capact::delete_cluster

    shout "Development local cluster deleted successfully."
}

main
