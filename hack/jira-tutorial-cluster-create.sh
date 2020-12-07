#!/usr/bin/env bash
#
# This script provisions demo cluster for Jira installation using 'kind'(kubernetes-in-docker)
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
    shout "Starting Voltron local cluster..."

    export REPO_DIR=$REPO_ROOT_DIR

    voltron::validate::tools

    export KUBERNETES_VERSION=${KUBERNETES_VERSION:-${STABLE_KUBERNETES_VERSION}}
    export KIND_CLUSTER_NAME="jira-tutorial"
    kind::create_cluster

    export DOCKER_TAG="1863a6f"
    export DOCKER_REPOSITORY="gcr.io/projectvoltron"
    export CLUSTER_TYPE="KIND"
    export MOCK_OCH_GRAPHQL=true
    export MOCK_ENGINE_GRAPHQL=false
    export DISABLE_MONITORING_INSTALLATION=true
    voltron::install_upgrade::charts

    if [[ "${DISABLE_HOSTS_UPDATE:-"false"}" == "true" ]]; then
      shout "Skipping updating /etc/hosts cause DISABLE_HOSTS_UPDATE is set to true."
    else
      host::update::voltron_hosts
    fi

    if [[ "${DISABLE_ADDING_TRUSTED_CERT:-"false"}" == "true" ]]; then
      shout "Skipping setting self-signed TLS certificate as trusted cause DISABLE_ADDING_TRUSTED_CERT is set to true."
    else
      host::install:trust_self_signed_cert
    fi

    shout "Voltron local cluster created successfully."
}

main
