#!/usr/bin/env bash
#
# This script build the populator CLI.
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

main () {
  for ARCH in $ARCHs; do
    for OS in $OSes; do
      echo "- Building populator binary [OS: ${OS} ARCH: ${ARCH}]..."
      binary="bin/populator-$OS-$ARCH"

      GOOS=$OS GOARCH=$ARCH go build -ldflags "-s -w" -o "$binary" "${REPO_ROOT_DIR}/cmd/populator/main.go"
      if ${UPX_ON} ; then
        upx -9 "$binary"
      fi
    done
  done
}

main
