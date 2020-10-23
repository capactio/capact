#!/usr/bin/env bash
#
# This script provisions a development environment using 'kind'(kubernetes-in-docker)
# Add cluster config to a file specified by KUBECONFIG env variable.
#
# It requires Docker to be installed.

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
    shout "Starting development local cluster..."

    export KUBERNETES_VERSION=${KUBERNETES_VERSION:-${STABLE_KUBERNETES_VERSION}}
    export KIND_CLUSTER_NAME=${KIND_CLUSTER_NAME:-${KIND_DEV_CLUSTER_NAME}}
    kind::create_cluster

    export DOCKER_TAG=dev
    export DOCKER_PUSH_REPOSITORY="local"
    export REPO_DIR=$REPO_ROOT_DIR
    voltron::install::from_sources

    shout "Development local cluster created successfully."
}

main
