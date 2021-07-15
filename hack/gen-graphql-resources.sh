#!/usr/bin/env bash
#
# This script generates the GraphQL related resources using qglgen
#
# Dependencies such as `gqlgen` binary can be installed on demand.

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

SKIP_DEPS_INSTALLATION=${SKIP_DEPS_INSTALLATION:-true}

# shellcheck source=./hack/lib/utilities.sh
source "${CURRENT_DIR}/lib/utilities.sh" || { echo 'Cannot load CI utilities.' exit 1; }
# shellcheck source=./hack/lib/const.sh
source "${CURRENT_DIR}/lib/const.sh" || { echo 'Cannot load constant values.' exit 1; }

cleanup() {
  rm -rf "${TMP_DIR}"
}

trap cleanup EXIT

host::install::gqlgen() {
  shout "Install the gqlgen ${STABLE_GQLGEN_VERSION} locally to a tempdir..."
  mkdir -p "${TMP_DIR}/bin"

  export PATH="${TMP_DIR}/bin:${PATH}"
  export GOBIN="${TMP_DIR}/bin"

  pushd "$TMP_DIR" >/dev/null

  go mod init tmp
  go install github.com/99designs/gqlgen@$STABLE_GQLGEN_VERSION

  popd >/dev/null

  echo -e "${GREEN}√ install gqlgen${NC}"
}

main() {
  if [[ "${SKIP_DEPS_INSTALLATION}" == "false" ]]; then
    host::install::gqlgen
  else
    echo "Skipping gqlgen installation cause SKIP_DEPS_INSTALLATION is set to true."
  fi

  shout "Generating Volron GraphQL related resources..."

  readonly apiPaths=(
    "/pkg/engine/api/graphql"
    "/pkg/hub/api/graphql/public"
    "/pkg/hub/api/graphql/local"
  )

  for path in "${apiPaths[@]}"; do
    echo "- Processing ${path}..."
    pushd "${REPO_ROOT_DIR}$path" > /dev/null
    gqlgen generate --verbose --config ./config.yaml
    popd > /dev/null
  done

  shout "Generation completed successfully."
}

main
