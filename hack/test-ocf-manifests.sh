#!/usr/bin/env bash
#
# This is a helper script for validating if the OCF manifests are valid against OCF specification.
#

set -o nounset
set -o errexit
set -o pipefail

CURRENT_DIR=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)
REPO_ROOT_DIR=$(cd "${CURRENT_DIR}/.." && pwd)
readonly CURRENT_DIR
readonly REPO_ROOT_DIR
cd "${REPO_ROOT_DIR}"

# shellcheck source=./hack/lib/utilities.sh
source "${CURRENT_DIR}/lib/utilities.sh" || { echo 'Cannot load CI utilities.'; exit 1; }

shout "Validating OCF examples..."
for pathPrefix in ocf-spec/*/examples ; do
    echo "- Testing examples in ${pathPrefix}..."
    go test -v --tags=ocfmanifests "./${pathPrefix}/..."
done

shout "Validating Hub content..."
go test -v --tags=ocfmanifests och-content/manifests_test.go
