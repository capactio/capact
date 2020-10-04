#!/usr/bin/env bash
#
# This script rebuilds Docker iamges from sources and upgrades Voltron Helm chart installed on cluster.
#
# It requires Docker to be installed.

# standard bash error handling
set -o nounset # treat unset variables as an error and exit immediately.
set -o errexit # exit immediately when a command fails.
set -E         # needs to be set if we want the ERR trap

readonly CURRENT_DIR=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )
readonly REPO_ROOT_DIR=${CURRENT_DIR}/..
readonly TMP_DIR=$(mktemp -d)

source "${CURRENT_DIR}/lib/utilities.sh" || { echo 'Cannot load CI utilities.'; exit 1; }
source "${CURRENT_DIR}/lib/deps_ver.sh" || { echo 'Cannot load dependencies versions.'; exit 1; }


VOLTRON_NAMESPACE="voltron"
VOLTRON_RELEASE_NAME="voltron"

main() {
    shout "Update development local cluster..."

    export UPDATE=true DOCKER_TAG=devel DOCKER_PUSH_REPOSITORY="local" REPO_DIR=$REPO_ROOT_DIR
    voltron::install::from_sources

    shout "Devel local cluster updated successfully."
}

main
