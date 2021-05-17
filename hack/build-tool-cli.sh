#!/usr/bin/env bash
#
# This script build the Capact CLI.
#
# Optional envs:
#  - UPX_ON

# standard bash error handling
set -o nounset # treat unset variables as an error and exit immediately.
set -o errexit # exit immediately when a command fails.
set -E         # needs to be set if we want the ERR trap

UPX_ON="${UPX_ON:-false}"
CURRENT_DIR=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)
REPO_ROOT_DIR=$(cd "${CURRENT_DIR}/.." && pwd)
readonly CURRENT_DIR
readonly REPO_ROOT_DIR
readonly UPX_ON

ARCHs=${CLI_ARCH:-"amd64"}
OSes=${CLI_OS:-"linux darwin windows"}

CAPACT_CMD_CLI_PKG="capact.io/capact/cmd/cli"
readonly CAPACT_CMD_CLI_PKG

# TODO: got to get version?
GO_BUILD_VERSION_LDFLAGS="
  -X ${CAPACT_CMD_CLI_PKG}/cmd.Version=$(git describe --tags)
  -X ${CAPACT_CMD_CLI_PKG}/cmd.Revision=${GITHUB_SHA:-$(git rev-parse HEAD)}
  -X ${CAPACT_CMD_CLI_PKG}/cmd.BuildUser=${GITHUB_ACTOR:-${USER:-}}
  -X ${CAPACT_CMD_CLI_PKG}/cmd.BuildDate=$(date +"%Y%m%d-%T")
  -X ${CAPACT_CMD_CLI_PKG}/cmd.Branch=$(git rev-parse --abbrev-ref HEAD)
"
readonly GO_BUILD_VERSION_LDFLAGS

main () {
  for ARCH in $ARCHs; do
    for OS in $OSes; do
      echo "- Building Capact CLI binary [OS: ${OS} ARCH: ${ARCH}]..."
      binary="bin/capact-$OS-$ARCH"

      GOOS=$OS GOARCH=$ARCH go build -ldflags "-s -w ${GO_BUILD_VERSION_LDFLAGS}" -o "$binary" "${REPO_ROOT_DIR}/cmd/cli/main.go"
      if ${UPX_ON} ; then
        upx -9 "$binary"
      fi
    done
  done
}

main
