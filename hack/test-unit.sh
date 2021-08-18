#!/usr/bin/env bash
#
# This script executes unit tests against whole Capact codebase.
#
# It requires Go to be installed.

# standard bash error handling
set -o nounset # treat unset variables as an error and exit immediately.
set -o errexit # exit immediately when a command fails.
set -E         # needs to be set if we want the ERR trap

CURRENT_DIR=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)
REPO_ROOT_DIR=$(cd "${CURRENT_DIR}/.." && pwd)
readonly CURRENT_DIR
readonly REPO_ROOT_DIR

# shellcheck source=./hack/lib/utilities.sh
source "${CURRENT_DIR}/lib/utilities.sh" || { echo 'Cannot load CI utilities.'; exit 1; }

pushd "${REPO_ROOT_DIR}" >/dev/null

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
  if ! go test -race -coverprofile="${REPO_ROOT_DIR}/coverage.txt" ./...;
 then
    echo -e "${RED}✗ go test\n${NC}"
    exit 1
  else
    echo -e "${GREEN}√ go test${NC}"
  fi
}

function test::ocf_spec() {
  shout "? OCF spec test "

  # Check if tests passed
  if ! go test --tags=ocfmanifests ./ocf-spec/...;
 then
    echo -e "${RED}✗ OCF spec\n${NC}"
    exit 1
  else
    echo -e "${GREEN}√ OCF spec${NC}"
  fi
}

function main() {
  print_info

  test::go_modules

  test::unit

  test::ocf_spec
}

main
