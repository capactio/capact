#!/usr/bin/env bash
#
# This script build the ocftool CLI. Runs the generate command to ensure OCF schemas are embedded.
#

# standard bash error handling
set -o nounset # treat unset variables as an error and exit immediately.
set -o errexit # exit immediately when a command fails.
set -E         # needs to be set if we want the ERR trap

readonly CURRENT_DIR=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)
readonly REPO_ROOT_DIR=$(cd "${CURRENT_DIR}/.." && pwd)

ARCHs=${CLI_ARCH:-"amd64"}
OSes=${CLI_OS:-"linux darwin windows"}

main () {
  for ARCH in $ARCHs; do
    for OS in $OSes; do
      echo "- Building ocftool binary [OS: ${OS} ARCH: ${ARCH}]..."
      binary="bin/ocftool-$OS-$ARCH"

      GOOS=$OS GOARCH=$ARCH go build -ldflags "-s -w" -o "$binary" "${REPO_ROOT_DIR}/cmd/ocftool/main.go"
    done
  done
}

main
