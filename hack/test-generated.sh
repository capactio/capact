#!/usr/bin/env bash
#
# This is a helper script for validating if generators were executed and results were committed.
#

set -o nounset
set -o errexit
set -o pipefail

CURRENT_DIR=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)
REPO_ROOT_DIR=$(cd "${CURRENT_DIR}/.." && pwd)
readonly CURRENT_DIR
readonly REPO_ROOT_DIR


# shellcheck source=./hack/lib/utilities.sh
source "${CURRENT_DIR}/lib/utilities.sh" || { echo 'Cannot load CI utilities.'; exit 1; }

capact::generate() {
  pushd "$REPO_ROOT_DIR"
  make generate
  popd
}

git::detect_dirty_state() {
  shout "- Checking for modified files..."

  # The porcelain format is used because it guarantees not to change in a backwards-incompatible
  # way between Git versions or based on user configuration.
  # source: https://git-scm.com/docs/git-status#_porcelain_format_version_1
  if [[ -n "$(git status --porcelain)" ]]; then
      echo "Detected modified files:"
      git status --porcelain

      echo "
      Run:
          make generate
      in the root of the repository and commit changes.
      "
      exit 1
  else
      echo -e "${GREEN}√ No issues detected. Have a nice day :-)${NC}"
  fi
}

main() {
  capact::generate
  git::detect_dirty_state
}

main
