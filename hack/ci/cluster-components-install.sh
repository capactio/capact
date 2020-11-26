#!/usr/bin/env bash
#
# This script install Voltron components on a cluster.
#

# standard bash error handling
set -o nounset # treat unset variables as an error and exit immediately.
set -o errexit # exit immediately when a command fails.
set -E         # needs to be set if we want the ERR trap

readonly CURRENT_DIR=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )
readonly REPO_ROOT_DIR=${CURRENT_DIR}/../..

# shellcheck source=./hack/lib/utilities.sh
source "${CURRENT_DIR}/../lib/utilities.sh" || { echo 'Cannot load CI utilities.'; exit 1; }
# shellcheck source=./hack/lib/const.sh
source "${CURRENT_DIR}/../lib/const.sh" || { echo 'Cannot load constant values.'; exit 1; }

main() {
    export REPO_DIR=$REPO_ROOT_DIR
    export CLUSTER_TYPE="GKE"
    export DOCKER_TAG=${OVERRIDE_DOCKER_TAG:-${DOCKER_TAG}}
    export MOCK_GRAPHQL=${MOCK_GRAPHQL:-${VOLTRON_MOCK_GRAPHQL}}
    voltron::install::charts
}

main
