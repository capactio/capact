#!/bin/bash
# Scan docker images using snyk for vulnerabilities
# NOTE: Assumes you have docker w/ snyk installed and that you accepted the license. Not tested on CI

# standard bash error handling
set -o nounset # treat unset variables as an error and exit immediately.
set -o errexit # exit immediately when a command fails.
set -E         # needs to be set if we want the ERR trap

CURRENT_DIR=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)
REPO_ROOT_DIR=$(cd "${CURRENT_DIR}/.." && pwd)
DEFAULT_DOCKERFILE="${REPO_ROOT_DIR}/Dockerfile"
readonly CURRENT_DIR
readonly REPO_ROOT_DIR
readonly DEFAULT_DOCKERFILE

DOCKER_REPOSITORY="${DOCKER_REPOSITORY:-ghcr.io/capactio}"

# shellcheck source=./hack/lib/utilities.sh
source "${CURRENT_DIR}/lib/utilities.sh" || { echo 'Cannot load CI utilities.'; exit 1; }

returnCode=0

# TODO: pick these up dynamically
imageArray=( "e2e-test" "argo-actions" "terraform-runner" "populator" "cloudsql-runner" "helm-runner" "argo-runner" "k8s-engine" "gateway" )
for image in "${imageArray[@]}"
do
    docker scan "${DOCKER_REPOSITORY}/${image}" --file="${DEFAULT_DOCKERFILE}" || returnCode=1
done

# TODO: pick these up dynamically
imageArray=( "jinja2" "graphql-schema-linter" "json-go-gen" )
for image in "${imageArray[@]}"
do
    docker scan "${DOCKER_REPOSITORY}/infra/${image}" --file="${REPO_ROOT_DIR}/hack/images/${image}/Dockerfile" || returnCode=1
done


# Other Docker Images
image="och-js"
docker scan "${DOCKER_REPOSITORY}/${image}" --file="${REPO_ROOT_DIR}/${image}/Dockerfile" || returnCode=1

if [ ${returnCode} -eq 0 ]; then
  shout "No vulnerabilities found."
else
  shout "Vulnerabilities found. See the logs for details."
fi
exit ${returnCode}
