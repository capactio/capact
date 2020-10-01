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

source "${CURRENT_DIR}/lib/utilities.sh" || { echo 'Cannot load CI utilities.'; exit 1; }
source "${CURRENT_DIR}/lib/deps_ver.sh" || { echo 'Cannot load dependencies versions.'; exit 1; }

SKIP_DEPS_INSTALLATION=${SKIP_DEPS_INSTALLATION:-true}

DUMP_CLUSTER_INFO="${DUMP_CLUSTER_INFO:-false}"

VOLTRON_NAMESPACE="voltron"
VOLTRON_RELEASE_NAME="voltron"

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

install::cluster::voltron() {
    readonly DOCKER_TAG=$RANDOM
    readonly DOCKER_PUSH_REPOSITORY="local"

    pushd "${REPO_ROOT_DIR}"
    shout "- Building Voltron image from sources..."
    env DOCKER_PUSH_REPOSITORY=$DOCKER_PUSH_REPOSITORY DOCKER_TAG=$DOCKER_TAG \
        make build-all-images

    REFERENCE_FILTER="$DOCKER_PUSH_REPOSITORY/*:$DOCKER_TAG"
    shout "- Loading Voltron image into kind cluster... [reference filter: $REFERENCE_FILTER]"
    names=$(docker::list_images "$REFERENCE_FILTER")
    kind::load_images "$names"

    shout "- Deleting local Docker Voltron images..."
    docker::delete_images "$names"

    shout "- Installing Voltron via helm chart from sources..."
    helm install ${VOLTRON_RELEASE_NAME} ./deploy/kubernetes/voltron \
        --create-namespace \
        --namespace=${VOLTRON_NAMESPACE} \
        --set global.containerRegistry.path=$DOCKER_PUSH_REPOSITORY \
        --set global.containerRegistry.overrideTag=$DOCKER_TAG \
        --wait

    popd
}

test::execute() {
    shout "- Executing e2e test..."
    pushd "${REPO_ROOT_DIR}/test/e2e/"
    helm test ${VOLTRON_RELEASE_NAME} --namespace=${VOLTRON_NAMESPACE} --timeout=${HELM_TEST_TIMEOUT} --logs
    popd
}

main() {
    shout "Starting integration test..."

    if [[ "${SKIP_DEPS_INSTALLATION}" == "" ]]; then
        export INSTALL_DIR=${TMP_DIR} KIND_VERSION=${STABLE_KIND_VERSION} HELM_VERSION=${STABLE_HELM_VERSION}
        install::local::kind
        install::local::helm
    else
        echo "Skipping kind and helm installation cause SKIP_DEPS_INSTALLATION is set to true."
    fi

    export KUBERNETES_VERSION=${KUBERNETES_VERSION:-${STABLE_KUBERNETES_VERSION}} KUBECONFIG="${TMP_DIR}/kubeconfig"
    kind::create_cluster

    # Cluster is already created, and all below operation are performed against that cluster,
    # so we should dump cluster info for debugging purpose in case of any error
    DUMP_CLUSTER_INFO=true

    install::cluster::voltron
    test::execute

    # Test completed successfully. We do not have to dump cluster info
    DUMP_CLUSTER_INFO=false
    shout "Integration test completed successfully."
}

main
