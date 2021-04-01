#!/bin/bash
set -eEu

readonly CURRENT_DIR=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )
readonly REPO_ROOT_DIR=${CURRENT_DIR}/../../..
readonly K8S_DEPLOY_DIR="${REPO_ROOT_DIR}/deploy/kubernetes"

echo -e "\n- Installing Cert Manager Helm chart...\n"

helm upgrade public-ingress-nginx "${K8S_DEPLOY_DIR}/charts/ingress-nginx" \
    --install \
    --create-namespace \
    --namespace="public-ingress-nginx" \
    --values "${CURRENT_DIR}/values.yml" \
    --wait

echo -e "\n- Waiting for public Ingress Controller to be ready...\n"
kubectl wait --namespace public-ingress-nginx \
  --for=condition=ready pod \
  --selector=app.kubernetes.io/component=controller \
  --timeout=90s

echo -e "\n- Public ingress controller installed!\n"
