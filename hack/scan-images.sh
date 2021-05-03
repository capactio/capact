#!/bin/bash
# shellcheck disable=SC2034
#
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

returnCode=0

# TODO: pick these up dynamically
imageArray=( "e2e-test" "argo-actions" "terraform-runner" "populator" "cloudsql-runner" "helm-runner" "argo-runner" "k8s-engine" "gateway" )
for image in "${imageArray[@]}"
do
	currentReturn=0
    docker scan "gcr.io/projectvoltron/${image}" --file="${DEFAULT_DOCKERFILE}" || returnCode=1
done

# TODO: pick these up dynamically
imageArray=( "jinja2" "graphql-schema-linter" "json-go-gen" )
for image in "${imageArray[@]}"
do
    docker scan "gcr.io/projectvoltron/infra/${image}" --file="${REPO_ROOT_DIR}/hack/images/${image}/Dockerfile" || returnCode=1
done


# Other Docker Images
image="och-js"
docker scan "gcr.io/projectvoltron/${image}" --file="${REPO_ROOT_DIR}/${image}/Dockerfile" || returnCode=1

[ ${returnCode} -eq 0 ] && echo "No vulnerabilities found." || echo "Vulnerabilies found"
exit ${returnCode}