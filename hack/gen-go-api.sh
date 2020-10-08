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

readonly CURRENT_DIR=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )
readonly REPO_ROOT_DIR=$(cd "${CURRENT_DIR}/.." && pwd)
readonly TMP_DIR=$(mktemp -d)

source "${CURRENT_DIR}/lib/utilities.sh" || { echo 'Cannot load CI utilities.'; exit 1; }
source "${CURRENT_DIR}/lib/deps_ver.sh" || { echo 'Cannot load dependencies versions.'; exit 1; }

VOLTRON_NAMESPACE="voltron"
VOLTRON_RELEASE_NAME="voltron"


main() {
    shout "Generating Go struct from OCF JSON Schemas..."
    OUTPUT="pkg/sdk/apis/0.0.1/types/types.go"

    pushd "${REPO_ROOT_DIR}"
    rm -f "$OUTPUT"

    docker run -v "$(PWD):/local" gcr.io/projectvoltron/infra/json-go-gen:0.1.0 -l go -s schema --package types \
      -S /local/ocf-spec/0.0.1/schema/common/metadata.json -S /local/ocf-spec/0.0.1/schema/common/json-schema-type.json \
      --src /local/ocf-spec/0.0.1/schema/interface.json \
      --src /local/ocf-spec/0.0.1/schema/implementation.json \
      --src /local/ocf-spec/0.0.1/schema/repo-metadata.json \
      --src /local/ocf-spec/0.0.1/schema/tag.json \
      --src /local/ocf-spec/0.0.1/schema/type.json \
      --src /local/ocf-spec/0.0.1/schema/type-instance.json \
      --src /local/ocf-spec/0.0.1/schema/vendor.json \
      -o "/local/$OUTPUT"


    popd
    shout "Generation completed successfully."
}

main
