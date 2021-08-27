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

CURRENT_DIR=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)
REPO_ROOT_DIR=$(cd "${CURRENT_DIR}/.." && pwd)
readonly CURRENT_DIR
readonly REPO_ROOT_DIR

# shellcheck source=./hack/lib/utilities.sh
source "${CURRENT_DIR}/lib/utilities.sh" || { echo 'Cannot load CI utilities.'; exit 1; }
# shellcheck source=./hack/lib/const.sh
source "${CURRENT_DIR}/lib/const.sh" || { echo 'Cannot load constant values.'; exit 1; }

main() {
    shout "Starting development local cluster..."

    export REPO_DIR=$REPO_ROOT_DIR

    capact::validate::tools

    export CLUSTER_TYPE=${CLUSTER_TYPE:-"kind"}
    export CLUSTER_NAME=${CLUSTER_NAME:-${DEV_CLUSTER_NAME}}
    capact::create_cluster

    export DOCKER_TAG=dev
    export DOCKER_REPOSITORY="local"
    export CLUSTER_NAME="${CLUSTER_NAME}"
    export PRINT_INSECURE_NOTES="true"
    shout "Installing Capact on development local cluster..."
    capact::install

    helm -n capact-system get notes capact

    shout "Development local cluster created successfully."
}

main
