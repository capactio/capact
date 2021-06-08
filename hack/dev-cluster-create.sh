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

    export KUBERNETES_VERSION=${KUBERNETES_VERSION:-${STABLE_KUBERNETES_VERSION}}
    export KIND_CLUSTER_NAME=${KIND_CLUSTER_NAME:-${KIND_DEV_CLUSTER_NAME}}
    kind::create_cluster

    export DOCKER_TAG=dev
    export DOCKER_REPOSITORY="local"
    export CLUSTER_TYPE="KIND"
    capact::update::images_on_kind
#    export DOCKER_REPOSITORY="ghcr.io/capactio"
#    export DOCKER_TAG="669208d"
    capact::install_upgrade::charts

    if [[ "${DISABLE_HOSTS_UPDATE:-"false"}" == "true" ]]; then
      shout "Skipping updating /etc/hosts cause DISABLE_HOSTS_UPDATE is set to true."
    else
      host::update::capact_hosts
    fi

    if [[ "${DISABLE_ADDING_TRUSTED_CERT:-"false"}" == "true" ]]; then
      shout "Skipping setting self-signed TLS certificate as trusted cause DISABLE_ADDING_TRUSTED_CERT is set to true."
    else
      host::install:trust_self_signed_cert
    fi

    shout "Development local cluster created successfully."
}

main

retry() {
  echo "Waiting for all Thar worker nodes to become 'Ready' in ${CLUSTER_NAME} cluster"
  KUBECTL="kubectl --kubeconfig ${KUBECONFIG_FILE}"
  MAX_ATTEMPTS=30
  attempts=0
  while true; do
    ((attempts+=1))
    if [ "${attempts}" -gt ${MAX_ATTEMPTS} ]; then
      echo "* Retry limit (${MAX_ATTEMPTS}) reached!" >&2
      exit 1
    fi
    sleep 5
    nodes=$(${KUBECTL} get nodes --no-headers)
    exit_on_error ${?} "* Failed to get node information for ${CLUSTER_NAME} cluster"

    found=$(${KUBECTL} get nodes --no-headers -o name | wc -l)
    ready=$(echo -n "${nodes}" | grep -c -w "Ready")
    echo "ready: ${ready}"

    if [ "${found}" -eq "${actual_num_nodes}" ] && [ "${ready}" -eq "${actual_num_nodes}" ]; then
      break
    fi
  done
}
