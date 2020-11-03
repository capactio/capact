#!/usr/bin/env bash
#
# This script generates the K8s related resources
# such as manifests (CRD, RBAC etc.) and code (DeepCopy, DeepCopyInto etc.)
#
# Dependencies such as `controller-gen` binary can be installed on demand.

# standard bash error handling
set -o nounset # treat unset variables as an error and exit immediately.
set -o errexit # exit immediately when a command fails.
set -E         # needs to be set if we want the ERR trap

readonly CURRENT_DIR=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)
readonly REPO_ROOT_DIR=$(cd "${CURRENT_DIR}/.." && pwd)
readonly UMBRELLA_CHART="${REPO_ROOT_DIR}/deploy/kubernetes/voltron"
readonly TMP_DIR=$(mktemp -d)

SKIP_DEPS_INSTALLATION=${SKIP_DEPS_INSTALLATION:-true}

# shellcheck source=./hack/lib/utilities.sh
source "${CURRENT_DIR}/lib/utilities.sh" || { echo 'Cannot load CI utilities.' exit 1; }
# shellcheck source=./hack/lib/deps_ver.sh
source "${CURRENT_DIR}/lib/deps_ver.sh" || { echo 'Cannot load CI utilities.' exit 1; }

cleanup() {
    rm -rf "${TMP_DIR}"
}

trap cleanup EXIT

host::install::controller-gen() {
    shout "Install the controller-gen ${STABLE_CONTROLLER_GEN_VERSION} locally to a tempdir..."
    mkdir -p "${TMP_DIR}/bin"

    export PATH="${TMP_DIR}/bin:${PATH}"
    export GOBIN="${TMP_DIR}/bin"

    pushd "$TMP_DIR" >/dev/null

    go mod init tmp
    go get sigs.k8s.io/controller-tools/cmd/controller-gen@STABLE_CONTROLLER_GEN_VERSION

    popd >/dev/null

    echo -e "${GREEN}√ install controller-gen${NC}"
}

main() {
  if [[ "${SKIP_DEPS_INSTALLATION}" == "false" ]]; then
    host::install::controller-gen
  else
    echo "Skipping controller-gen installation cause SKIP_DEPS_INSTALLATION is set to true."
  fi

  shout "Generating Kubernetes related resources..."

  CRDS_OUTPUT="${UMBRELLA_CHART}/crds"
  RBAC_OUTPUT="${UMBRELLA_CHART}/charts/engine/templates"

  controller-gen object crd:trivialVersions=true rbac:roleName=k8s-engine-role \
    paths="$REPO_ROOT_DIR/..." \
    output:crd:artifacts:config="$CRDS_OUTPUT" \
    output:rbac:artifacts:config="$RBAC_OUTPUT"

  echo "CRDs manifests saved in $CRDS_OUTPUT"
  echo "RBAC manifest saved in $RBAC_OUTPUT"

  shout "Generation completed successfully."
}

main
