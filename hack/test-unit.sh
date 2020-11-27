#!/usr/bin/env bash
#
# This script executes unit tests against whole Voltron codebase.
#
# It requires Go to be installed.

# standard bash error handling
set -o nounset # treat unset variables as an error and exit immediately.
set -o errexit # exit immediately when a command fails.
set -E         # needs to be set if we want the ERR trap

readonly CURRENT_DIR=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)
readonly ROOT_PATH=${CURRENT_DIR}/..

# shellcheck source=./hack/lib/utilities.sh
source "${CURRENT_DIR}/lib/utilities.sh" || { echo 'Cannot load CI utilities.'; exit 1; }

pushd "${ROOT_PATH}" >/dev/null

# Exit handler. This function is called anytime an EXIT signal is received.
# This function should never be explicitly called.
function _trap_exit() {
  popd >/dev/null
}
trap _trap_exit EXIT

function print_info() {
  echo -e "${INVERTED}"
  echo "USER: ${USER}"
  echo "PATH: ${PATH}"
  echo "GOPATH: ${GOPATH:-"unknown"}"
  echo -e "${NC}"
}

function test::go_modules() {
  shout "? go mod tidy"
  go mod tidy

  # check, if cleaned go.mod or go.sum are not is git stage
  # grep fails, when didn't match anything, so "|| true"o is used to supress the error code
  STATUS=$(git status --porcelain go.mod go.sum | grep -E '^ M' || true )
  if [ -n "$STATUS" ]; then
    echo -e "${RED}✗ go mod tidy modified go.mod and/or go.sum${NC}"
    exit 1
  else
    echo -e "${GREEN}√ go mod tidy${NC}"
  fi
}


function test::unit() {
  shout "? go test"

  # Check if tests passed
  if ! go test -race -coverprofile="${ROOT_PATH}/coverage.txt" ./...;
 then
    echo -e "${RED}✗ go test\n${NC}"
    exit 1
  else
    echo -e "${GREEN}√ go test${NC}"
  fi
}

function main() {
  print_info

  test::go_modules

  test::unit
}

main
