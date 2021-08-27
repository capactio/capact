#!/usr/bin/env bash
#
# This script provisions testing environment using 'kind'(kubernetes-in-docker)
# and execute end-to-end Capact tests.
#
# It requires Docker to be installed.
# Dependencies such as Helm v3 and kind can be installed on demand.

# standard bash error handling
set -o nounset # treat unset variables as an error and exit immediately.
set -o errexit # exit immediately when a command fails.
set -E         # needs to be set if we want the ERR trap

CURRENT_DIR=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)
REPO_ROOT_DIR=$(cd "${CURRENT_DIR}/.." && pwd)
TMP_DIR=$(mktemp -d)
readonly CURRENT_DIR
readonly REPO_ROOT_DIR
readonly TMP_DIR

# shellcheck source=./hack/lib/utilities.sh
source "${CURRENT_DIR}/lib/utilities.sh" || { echo 'Cannot load CI utilities.'; exit 1; }
# shellcheck source=./hack/lib/const.sh
source "${CURRENT_DIR}/lib/const.sh" || { echo 'Cannot load constant values.'; exit 1; }

SKIP_DEPS_INSTALLATION=${SKIP_DEPS_INSTALLATION:-true}

DUMP_CLUSTER_INFO="${DUMP_CLUSTER_INFO:-false}"

HELM_TEST_TIMEOUT="15m"

cleanup() {
    if [[ "${DUMP_CLUSTER_INFO}" == true ]]; then
        shout '- Creating artifacts...'

        dump_cluster_info || true
    fi

    shout "Nodes description"
    kubectl describe nodes

    capact::delete_cluster || true

    rm -rf "${TMP_DIR}"
}

trap cleanup EXIT

capact::test::execute() {
    shout "- Executing e2e test..."
    helm test ${CAPACT_RELEASE_NAME} --namespace=${CAPACT_NAMESPACE} --timeout=${HELM_TEST_TIMEOUT} --logs
}

main() {
    shout "Starting integration test..."
    if [[ "${SKIP_DEPS_INSTALLATION}" == "false" ]]; then
        export INSTALL_DIR=${TMP_DIR}
        export HELM_VERSION=${STABLE_HELM_VERSION}
        host::install::helm
    else
        echo "Skipping kind and helm installation cause SKIP_DEPS_INSTALLATION is set to true."
    fi


    export REPO_DIR=$REPO_ROOT_DIR
    export CLUSTER_TYPE=${CLUSTER_TYPE:-"kind"}

    export KUBECONFIG="${TMP_DIR}/kubeconfig"
    export CLUSTER_NAME=${CLUSTER_NAME:-${KIND_CI_CLUSTER_NAME}}
    export HELM_VERSION=${STABLE_HELM_VERSION}
    capact::create_cluster

    # Cluster is already created, and all below operations are performed against that cluster,
    # so we should dump cluster info for debugging purpose in case of any error
    DUMP_CLUSTER_INFO=true


    if [[ "${BUILD_IMAGES:-"true"}" == "true" ]]; then
      export DOCKER_TAG=$RANDOM
      export DOCKER_REPOSITORY="local"
    fi

    export INCREASE_RESOURCE_LIMITS="false" # To comply with the default GitHub Actions Runner limits
    export CLUSTER_NAME="${CLUSTER_NAME}"
    export USE_TEST_SETUP="true"
    export PRINT_INSECURE_NOTES="false"
    capact::install

    capact::test::execute
    # Test completed successfully. We do not have to dump cluster info
    DUMP_CLUSTER_INFO=false
    shout "Integration test completed successfully."
}

main
