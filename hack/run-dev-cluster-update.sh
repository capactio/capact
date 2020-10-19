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

VOLTRON_NAMESPACE="voltron"
VOLTRON_RELEASE_NAME="voltron"

main() {
    shout "Update development local cluster..."

    export UPDATE=true DOCKER_TAG="dev-$RANDOM" DOCKER_PUSH_REPOSITORY="local" REPO_DIR=$REPO_ROOT_DIR KIND_CLUSTER_NAME="kind-dev-voltron"
    voltron::install::from_sources

    shout "Development local cluster updated successfully."
}

main
