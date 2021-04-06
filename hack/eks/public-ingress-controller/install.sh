#!/usr/bin/env bash
#
# This script installs public ingress-nginx Helm chart.
#

# standard bash error handling
set -o nounset # treat unset variables as an error and exit immediately.
set -o errexit # exit immediately when a command fails.
set -E         # needs to be set if we want the ERR trap

readonly CURRENT_DIR=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )
readonly REPO_ROOT_DIR=${CURRENT_DIR}/../../..
readonly K8S_DEPLOY_DIR="${REPO_ROOT_DIR}/deploy/kubernetes"

helm upgrade public-ingress-nginx "${K8S_DEPLOY_DIR}/charts/ingress-nginx" \
    --install \
    --create-namespace \
    --namespace="public-ingress-nginx" \
    --values "${CURRENT_DIR}/values.yml" \
    --wait

kubectl wait --namespace public-ingress-nginx \
  --for=condition=ready pod \
  --selector=app.kubernetes.io/component=controller \
  --timeout=90s
