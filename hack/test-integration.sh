#!/usr/bin/env bash
#
# This script provisions testing environment using 'kind'(kubernetes-in-docker)
# and execute end-to-end Voltron tests.
#
# It requires Docker to be installed.
# Dependencies such as Helm v3 and kind can be installed on demand.

# standard bash error handling
set -o nounset # treat unset variables as an error and exit immediately.
set -o errexit # exit immediately when a command fails.
set -E         # needs to be set if we want the ERR trap

readonly CURRENT_DIR=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )
readonly REPO_ROOT_DIR=${CURRENT_DIR}/..
readonly TMP_DIR=$(mktemp -d)

# shellcheck source=./hack/lib/utilities.sh
source "${CURRENT_DIR}/lib/utilities.sh" || { echo 'Cannot load CI utilities.'; exit 1; }
# shellcheck source=./hack/lib/const.sh
source "${CURRENT_DIR}/lib/const.sh" || { echo 'Cannot load constant values.'; exit 1; }

SKIP_DEPS_INSTALLATION=${SKIP_DEPS_INSTALLATION:-true}

DUMP_CLUSTER_INFO="${DUMP_CLUSTER_INFO:-false}"

HELM_TEST_TIMEOUT="10m"

cleanup() {
    if [[ "${DUMP_CLUSTER_INFO}" == true ]]; then
        shout '- Creating artifacts...'

        export DUMP_NAMESPACE=${VOLTRON_NAMESPACE}
        dump_logs || true
    fi

    kind::delete_cluster || true

    rm -rf "${TMP_DIR}"
}

trap cleanup EXIT

voltron::test::execute() {
    shout "- Executing e2e test..."
    helm test ${VOLTRON_RELEASE_NAME} --namespace=${VOLTRON_NAMESPACE} --timeout=${HELM_TEST_TIMEOUT} --logs
}

main() {
    shout "Starting integration test..."

    if [[ "${SKIP_DEPS_INSTALLATION}" == "false" ]]; then
        export INSTALL_DIR=${TMP_DIR}
        export KIND_VERSION=${STABLE_KIND_VERSION}
        export HELM_VERSION=${STABLE_HELM_VERSION}
        host::install::kind
        host::install::helm
    else
        echo "Skipping kind and helm installation cause SKIP_DEPS_INSTALLATION is set to true."
    fi

    export REPO_DIR=$REPO_ROOT_DIR

    export KUBERNETES_VERSION=${KUBERNETES_VERSION:-${STABLE_KUBERNETES_VERSION}}
    export KUBECONFIG="${TMP_DIR}/kubeconfig"
    export KIND_CLUSTER_NAME=${KIND_CLUSTER_NAME:-${KIND_CI_CLUSTER_NAME}}
    kind::create_cluster

    # Cluster is already created, and all below operation are performed against that cluster,
    # so we should dump cluster info for debugging purpose in case of any error
    DUMP_CLUSTER_INFO=true

    if [[ "${BUILD_IMAGES:-"true"}" == "true" ]]; then
      export DOCKER_TAG=$RANDOM
      export DOCKER_REPOSITORY="local"
      voltron::update::images_on_kind
    fi

    export CLUSTER_TYPE="KIND"
    export MOCK_GRAPHQL=${MOCK_GRAPHQL:-${VOLTRON_MOCK_GRAPHQL}}
    voltron::install::charts

    voltron::test::execute
    # Test completed successfully. We do not have to dump cluster info
    DUMP_CLUSTER_INFO=false
    shout "Integration test completed successfully."
}

main
