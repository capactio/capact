#!/usr/bin/env bash
#
# This script generates the Go struct from the JSON Schemas for the OCF Manifest.
# The quicktype is used for that purpose.
#
# It requires Docker to be installed.

# standard bash error handling
set -o nounset # treat unset variables as an error and exit immediately.
set -o errexit # exit immediately when a command fails.
set -E         # needs to be set if we want the ERR trap

readonly CURRENT_DIR=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)
readonly REPO_ROOT_DIR=$(cd "${CURRENT_DIR}/.." && pwd)

# shellcheck source=./hack/lib/utilities.sh
source "${CURRENT_DIR}/lib/utilities.sh" || { echo 'Cannot load CI utilities.'; exit 1; }
# shellcheck source=./hack/lib/const.sh
source "${CURRENT_DIR}/lib/const.sh" || { echo 'Cannot load constant values.' exit 1; }

OCF_VERSION="${OCF_VERSION:-$DEFAULT_OCF_VERSION}"

REPORT_FILE_DIR="${REPO_ROOT_DIR}/tmp"
REPORT_FILENAME="${REPORT_FILE_DIR}/gen_go_api_issues.txt"
KNOWN_VIOLATION_FILENAME="${CURRENT_DIR}/gen_go_api_issue_exceptions.txt"

check_for_unknown_issues() {
  shout "Checking for unknown generate issues..."

  if ! diff -u "${REPORT_FILENAME}" "${KNOWN_VIOLATION_FILENAME}"; then
    echo "Error:
    API rules check failed. Reported violations \"${REPORT_FILENAME}\" differ from known violations \"${KNOWN_VIOLATION_FILENAME}\"."
    
    diff -u "${REPORT_FILENAME}" "${KNOWN_VIOLATION_FILENAME}"

    echo "Please fix API source file if new violation is detected, or update known violations \"${KNOWN_VIOLATION_FILENAME}\" if existing violation is being fixed."
    exit 1
  fi

  echo -e "${GREEN}√ No issues detected. Have a nice day :-)${NC}"
}

gen_go_api_from_ocf_specs() {
  shout "Generating Go struct from OCF JSON Schemas..."
  OUTPUT="pkg/sdk/apis/${OCF_VERSION}/types/types.gen.go"
  mkdir -p "${REPORT_FILE_DIR}"

  pushd "${REPO_ROOT_DIR}"
  rm -f "$OUTPUT"

  docker pull "${JSON_GO_GEN_IMAGE}"
  docker run -v "${PWD}:/local" "${JSON_GO_GEN_IMAGE}" -l go -s schema --package types \
    --additional-schema "/local/ocf-spec/${OCF_VERSION}/schema/common/metadata.json" \
    --additional-schema "/local/ocf-spec/${OCF_VERSION}/schema/common/metadata-tags.json" \
    --additional-schema "/local/ocf-spec/${OCF_VERSION}/schema/common/json-schema-type.json" \
    --src "/local/ocf-spec/${OCF_VERSION}/schema/interface.json" \
    --src "/local/ocf-spec/${OCF_VERSION}/schema/implementation.json" \
    --src "/local/ocf-spec/${OCF_VERSION}/schema/repo-metadata.json" \
    --src "/local/ocf-spec/${OCF_VERSION}/schema/tag.json" \
    --src "/local/ocf-spec/${OCF_VERSION}/schema/type.json" \
    --src "/local/ocf-spec/${OCF_VERSION}/schema/type-instance.json" \
    --src "/local/ocf-spec/${OCF_VERSION}/schema/vendor.json" \
    -o "/local/$OUTPUT" 2> "${REPORT_FILENAME}"

  popd
  shout "Generation completed successfully."
}

main() {
  gen_go_api_from_ocf_specs
  check_for_unknown_issues
}

main
