#!/usr/bin/env bash
#
# This script generates gRPC + ProtoBuf Go code for client and server.
#
# Dependencies such as `protoc` binary can be installed on demand.

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

# protoc uses different naming pattern for binary than what we have in the utilities.sh
host::os() {
  local host_os
  case "$(uname -s)" in
    Darwin)
      host_os=osx
      ;;
    Linux)
      host_os=linux
      ;;
    *)
      echo "Unsupported host OS. Must be Linux or Mac OS X."
      exit 1
      ;;
  esac
  echo "${host_os}"
}

host::install::protoc() {
  shout "Install the protoc ${STABLE_PROTOC_VERSION} locally to a tempdir..."
  mkdir -p "${TMP_DIR}/bin"

  export PATH="${TMP_DIR}/bin:${PATH}"
  pushd "$TMP_DIR" >/dev/null

  os=$(host::os)
  arch=$(uname -m)
  version_without_v=${STABLE_PROTOC_VERSION#"v"}
  name="protoc-${version_without_v}-${os}-${arch}"

  # download the release
  curl -L -O "https://github.com/protocolbuffers/protobuf/releases/download/${STABLE_PROTOC_VERSION}/${name}.zip"

  # extract the archive
  unzip "${name}".zip

  popd >/dev/null

  echo -e "${GREEN}√ install protoc${NC}"
}

main() {
  if [[ "${SKIP_DEPS_INSTALLATION}" == "false" ]]; then
    host::install::protoc
  else
    echo "Skipping protoc installation cause SKIP_DEPS_INSTALLATION is set to true."
  fi

  shout "Generating Capact gRPC related resources..."

  readonly apiPaths=(
    "/pkg/hub/api/grpc"
  )

  for path in "${apiPaths[@]}"; do
    echo "- Processing ${path}..."
    pushd "${REPO_ROOT_DIR}$path" > /dev/null
    protoc -I=. --go_out=. --go-grpc_out=. ./*.proto
    popd > /dev/null
  done

  shout "Generation completed successfully."
}

main
