#!/usr/bin/env bash
#
# This scripts runs linters to ensure the correctness of the Capact Go codebase.
#
# Golangci-lint dependencies can be installed on-demand.

# standard bash error handling
set -o nounset # treat unset variables as an error and exit immediately.
set -o errexit # exit immediately when a command fails.
set -E         # needs to be set if we want the ERR trap

CURRENT_DIR=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)
REPO_ROOT_DIR=$(cd "${CURRENT_DIR}/.." && pwd)
TMP_DIR=$(mktemp -d)
readonly TMP_DIR
readonly CURRENT_DIR
readonly REPO_ROOT_DIR
readonly GOLANGCI_LINT_VERSION="v1.31.0"

LINT_TIMEOUT=${LINT_TIMEOUT:-5m}
SKIP_DEPS_INSTALLATION=${SKIP_DEPS_INSTALLATION:-true}

# shellcheck source=./hack/lib/utilities.sh
source "${CURRENT_DIR}/lib/utilities.sh" || { echo 'Cannot load CI utilities.'; exit 1; }
# shellcheck source=./hack/lib/const.sh
source "${CURRENT_DIR}/lib/const.sh" || { echo 'Cannot load constant values.' exit 1; }

GRAPHQL_SCHEMA_LINTER_IMAGE="gcr.io/projectvoltron/pr/infra/graphql-schema-linter:${GRAPHQL_SCHEMA_LINTER_IMAGE_TAG}"

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
  golangci-lint run --timeout="${LINT_TIMEOUT}" $(golangci::fix_if_requested) "${REPO_ROOT_DIR}/..."

  echo -e "${GREEN}√ run golangci-lint${NC}"
}

golangci::fix_if_requested() {
  if [[ "${LINT_FORCE_FIX:-x}" == "true" ]]; then
    echo --fix
  fi
}

dockerfile::run_checks() {
  shout "Run hadolint Dockerfile checks"
  docker run --rm -i hadolint/hadolint < "${REPO_ROOT_DIR}/Dockerfile"
  echo -e "${GREEN}√ run hadolint${NC}"
}

shellcheck::files_to_check() {
  pushd "$REPO_ROOT_DIR" > /dev/null
  paths=$(find . \( -path ./och-js/node_modules -o -path ./hack/eks/terraform/.terraform -o -path ./tmp \) -prune -false -o -name '*.sh')
  popd > /dev/null

  echo "$paths"
}

# In the future we can add support for auto fix: https://github.com/koalaman/shellcheck/issues/1220
shellcheck::run_checks() {
  shout "Run shellcheck checks"

  # shellcheck disable=SC2046
  docker run --rm -v "$REPO_ROOT_DIR":/mnt -w /mnt koalaman/shellcheck:stable -x $(shellcheck::files_to_check)
  echo -e "${GREEN}√ run shellcheck${NC}"
}

graphql::run_checks() {
  shout "Run graphql-schema-linter checks"

  docker run --rm -v "$REPO_ROOT_DIR":/repo -w=/repo "${GRAPHQL_SCHEMA_LINTER_IMAGE}" \
    --src ./pkg/engine/api/graphql/schema.graphql \
    --src ./och-js/graphql/public/schema.graphql \
    --src ./och-js/graphql/local/schema.graphql \
    --linter-args "-c ./ --format compact"
  echo -e "${GREEN}√ run graphql-schema-linter${NC}"
}

main() {
  if [[ "${SKIP_DEPS_INSTALLATION}" == "false" ]]; then
      host::install::golangci
  else
      echo "Skipping golangci-lint installation cause SKIP_DEPS_INSTALLATION is set to true."
  fi

  golangci::run_checks

  dockerfile::run_checks

  shellcheck::run_checks

  graphql::run_checks
}

main
