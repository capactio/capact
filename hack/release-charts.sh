#!/usr/bin/env bash
#
# This is a helper script for Helm Chart releasing.
#

set -o nounset
set -o errexit
set -o pipefail

readonly CURRENT_DIR=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )
readonly REPO_ROOT_DIR=$(cd "${CURRENT_DIR}/.." && pwd)

main() {
  readonly CR_PACKAGE_PATH="${REPO_ROOT_DIR}/tmp/charts"
  readonly CAPACTIO_BUCKET="capactio-awesome-charts"
  readonly CAPACTIO_REPO_URL=https://${CAPACTIO_BUCKET}.storage.googleapis.com/

  readonly charts=(
    "argo"
    "ingress-nginx"
    "kubed"
    "monitoring"
    "neo4j"
    "voltron"
  )

  mkdir -p ${CR_PACKAGE_PATH}
  pushd ${CR_PACKAGE_PATH}

  # Copy old index
  gsutil cp gs://${CAPACTIO_BUCKET}/index.yaml .

  for chart in "${charts[@]}"; do
      # Currently, we execute this method on locally and committed charts already has .tgz with dependent chart.
      # It is less robust but decrease CI pipeline time.
      # If enabled we also need to add `helm repo add ...` for each dependency.
      # helm dep build "${REPO_ROOT_DIR}/deploy/kubernetes/charts/${chart}"

      helm package "${REPO_ROOT_DIR}/deploy/kubernetes/charts/${chart}"
  done

  helm repo index --url "${CAPACTIO_REPO_URL}" --merge ./index.yaml .
  gsutil -m rsync ./ gs://"${CAPACTIO_BUCKET}"/

  popd
  rm -rf ${CR_PACKAGE_PATH}
}

main
