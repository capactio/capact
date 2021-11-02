#!/usr/bin/env bash
#
# This script installs public ingress-nginx Helm chart.
#

# standard bash error handling
set -o nounset # treat unset variables as an error and exit immediately.
set -o errexit # exit immediately when a command fails.
set -E         # needs to be set if we want the ERR trap

CURRENT_DIR=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)
REPO_ROOT_DIR=$(cd "${CURRENT_DIR}/../../.." && pwd)
readonly CURRENT_DIR
readonly REPO_ROOT_DIR
readonly K8S_DEPLOY_DIR="${REPO_ROOT_DIR}/deploy/kubernetes"

helm upgrade public-ingress-nginx "${K8S_DEPLOY_DIR}/charts/ingress-controller" \
    --install \
    --namespace="capact-system" \
    --values "${CURRENT_DIR}/values.yml" \
    --wait

kubectl wait --namespace capact-system \
  --for=condition=ready pod \
  --selector=app.kubernetes.io/component=controller \
  --timeout=90s
