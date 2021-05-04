#!/usr/bin/env bash
#
# This script build the Capact CLI.
#

# standard bash error handling
set -o nounset # treat unset variables as an error and exit immediately.
set -o errexit # exit immediately when a command fails.
set -E         # needs to be set if we want the ERR trap

CURRENT_DIR=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)
REPO_ROOT_DIR=$(cd "${CURRENT_DIR}/.." && pwd)
readonly CURRENT_DIR
readonly REPO_ROOT_DIR

ARCHs=${CLI_ARCH:-"amd64"}
OSes=${CLI_OS:-"linux darwin windows"}

main () {
  for ARCH in $ARCHs; do
    for OS in $OSes; do
      echo "- Building Capact CLI binary [OS: ${OS} ARCH: ${ARCH}]..."
      binary="bin/capact-$OS-$ARCH"

      GOOS=$OS GOARCH=$ARCH go build -ldflags "-s -w" -o "$binary" "${REPO_ROOT_DIR}/cmd/cli/main.go"
    done
  done
}

main
