#!/usr/bin/env bash
#
# This is a helper script for Helm Chart releasing.
# Set MASTER_BUILD to `true` to change the Helm chart version to commit SHA and push them to the master branch.
#

set -o nounset
set -o errexit
set -o pipefail

readonly CURRENT_DIR=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)
readonly REPO_ROOT_DIR=$(cd "${CURRENT_DIR}/.." && pwd)
readonly DEPLOY_CHARTS_DIR="${REPO_ROOT_DIR}/deploy/kubernetes/charts"

readonly CR_PACKAGE_PATH="${REPO_ROOT_DIR}/tmp/charts"
readonly CAPACTIO_OFFICIAL_BUCKET="capactio-awesome-charts"
readonly CAPACTIO_MASTER_BUCKET="capactio-master-charts"

readonly charts=(
  "argo"
  "ingress-nginx"
  "kubed"
  "monitoring"
  "neo4j"
  "voltron"
)

# shellcheck source=./hack/lib/utilities.sh
source "${CURRENT_DIR}/lib/utilities.sh" || { echo 'Cannot load CI utilities.'; exit 1; }

setChartVersionToCommitSHA() {
  readonly version="${GITHUB_SHA:0:7}"

  for chart in "${charts[@]}"; do
    sed -i.bak "/^version: / s/$/-${version}/" "${DEPLOY_CHARTS_DIR}/${chart}/Chart.yaml"
  done
}

function wereChartsModifed() {
  readonly DIFF=$(git diff HEAD^ HEAD -- "${DEPLOY_CHARTS_DIR}")
  if [ "${DIFF:-}" = "" ]; then
    return 1
  fi
  return 0
}

main() {
  if wereChartsModifed; then
    shout "Changes detected in ${DEPLOY_CHARTS_DIR}. Starting Helm chart releasing..."
  else
    shout "No changes detected in ${DEPLOY_CHARTS_DIR}. Skipping Helm chart releasing."
    exit 0
  fi

  local CAPACTIO_BUCKET="${CAPACTIO_OFFICIAL_BUCKET}"
  if [ "${MASTER_BUILD:-}" = "true" ]; then
    CAPACTIO_BUCKET="${CAPACTIO_MASTER_BUCKET}"
    setChartVersionToCommitSHA
  fi

  readonly CAPACTIO_REPO_URL=https://storage.googleapis.com/${CAPACTIO_BUCKET}

  mkdir -p "${CR_PACKAGE_PATH}"
  pushd "${CR_PACKAGE_PATH}"

  # Copy old index
  gsutil cp gs://${CAPACTIO_BUCKET}/index.yaml .

  for chart in "${charts[@]}"; do
    # Currently, we execute this method on locally and committed charts already has .tgz with dependent chart.
    # It is less robust but decrease CI pipeline time.
    # If enabled we also need to add `helm repo add ...` for each dependency.
    # helm dep build "${DEPLOY_CHARTS_DIR}/${chart}"

    helm package "${DEPLOY_CHARTS_DIR}/${chart}"
  done

  helm repo index --url "${CAPACTIO_REPO_URL}" --merge ./index.yaml .
  gsutil -m rsync ./ gs://"${CAPACTIO_BUCKET}"/

  # By default Google sets `cache-control: public, max-age=3600`.
  # We need to change to ensure the file is not cached by http clients
  # and we get rid of `chart version X.Y.Z not could in repository` errors.
  # source: https://cloud.google.com/storage/docs/caching#performance_considerations
  gsutil setmeta -h "Cache-Control: no-cache, no-store" gs://"${CAPACTIO_BUCKET}"/index.yaml

  popd
  rm -rf "${CR_PACKAGE_PATH}"
}

main
