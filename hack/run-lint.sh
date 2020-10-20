#!/usr/bin/env bash
#
# This scripts runs linters to ensure the correctness of the Voltron Go codebase.
#
# Golangci-lint dependencies can be installed on-demand.

# standard bash error handling
set -o nounset # treat unset variables as an error and exit immediately.
set -o errexit # exit immediately when a command fails.
set -E         # needs to be set if we want the ERR trap

readonly CURRENT_DIR=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )
readonly ROOT_PATH=$(cd "${CURRENT_DIR}/.." && pwd)
readonly GOLANGCI_LINT_VERSION="v1.31.0"
readonly TMP_DIR=$(mktemp -d)

SKIP_DEPS_INSTALLATION=${SKIP_DEPS_INSTALLATION:-true}

# shellcheck source=./hack/lib/utilities.sh
source "${CURRENT_DIR}/lib/utilities.sh" || { echo 'Cannot load CI utilities.'; exit 1; }

cleanup() {
    rm -rf "${TMP_DIR}"
}

trap cleanup EXIT

host::install::golangci() {
    mkdir -p "${TMP_DIR}/bin"
    export PATH="${TMP_DIR}/bin:${PATH}"

    shout "Install the golangci-lint ${GOLANGCI_LINT_VERSION} locally to a tempdir..."
    curl -sfSL -o "${TMP_DIR}/golangci-lint.sh" https://install.goreleaser.com/github.com/golangci/golangci-lint.sh
    chmod 700 "${TMP_DIR}/golangci-lint.sh"

    "${TMP_DIR}/golangci-lint.sh" -b "${TMP_DIR}/bin" ${GOLANGCI_LINT_VERSION}

    echo -e "${GREEN}√ install golangci-lint${NC}"
}

golangci::run_checks() {
  if [ -z "$(command -v golangci-lint)" ]; then
    echo "golangci-lint not found locally. Execute script with env variable SKIP_DEPS_INSTALLATION=false"
    exit 1
  fi

  shout "Run golangci-lint checks"

  # shellcheck disable=SC2046
  golangci-lint run $(golangci::fix_if_requested) "${ROOT_PATH}/..."

  echo -e "${GREEN}√ run golangci-lint${NC}"
}

golangci::fix_if_requested() {
  if [[ "${LINT_FORCE_FIX:-x}" == "true" ]]; then
    echo --fix
  fi
}

docker::run_dockerfile_checks() {
  shout "Run hadolint Dockerfile checks"
  docker run --rm -i hadolint/hadolint < "${ROOT_PATH}/Dockerfile"
  echo -e "${GREEN}√ run hadolint${NC}"
}


# In the future we can add support for auto fix: https://github.com/koalaman/shellcheck/issues/1220
shellcheck::run_checks() {
  shout "Run shellcheck checks"

  docker run --rm -v "$ROOT_PATH":/mnt koalaman/shellcheck:stable -x ./hack/*.sh
  echo -e "${GREEN}√ run shellcheck${NC}"
}

main() {
  if [[ "${SKIP_DEPS_INSTALLATION}" == "false" ]]; then
      host::install::golangci
  else
      echo "Skipping golangci-lint installation cause SKIP_DEPS_INSTALLATION is set to true."
  fi

  golangci::run_checks

  docker::run_dockerfile_checks

  shellcheck::run_checks
}

main
