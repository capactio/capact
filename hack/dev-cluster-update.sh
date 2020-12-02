#!/usr/bin/env bash
#
# This script rebuilds Docker images from sources and upgrades Voltron Helm chart installed on cluster.
#

# standard bash error handling
set -o nounset # treat unset variables as an error and exit immediately.
set -o errexit # exit immediately when a command fails.
set -E         # needs to be set if we want the ERR trap

readonly CURRENT_DIR=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )
readonly REPO_ROOT_DIR=${CURRENT_DIR}/..

# shellcheck source=./hack/lib/utilities.sh
source "${CURRENT_DIR}/lib/utilities.sh" || { echo 'Cannot load CI utilities.'; exit 1; }
# shellcheck source=./hack/lib/const.sh
source "${CURRENT_DIR}/lib/const.sh" || { echo 'Cannot load constant values.'; exit 1; }

main() {
    shout "Updating development local cluster..."

    export DOCKER_TAG="dev-$RANDOM"
    export DOCKER_REPOSITORY="local"
    export REPO_DIR=$REPO_ROOT_DIR
    export KIND_CLUSTER_NAME=${KIND_CLUSTER_NAME:-${KIND_DEV_CLUSTER_NAME}}
    export CLUSTER_TYPE="KIND"
    voltron::update::images_on_kind

    export MOCK_OCH_GRAPHQL=${MOCK_OCH_GRAPHQL:-${VOLTRON_MOCK_OCH_GRAPHQL}}
    export MOCK_ENGINE_GRAPHQL=${MOCK_ENGINE_GRAPHQL:-${VOLTRON_MOCK_ENGINE_GRAPHQL}}
    voltron::install_upgrade::charts

    shout "Development local cluster updated successfully."
}

main
