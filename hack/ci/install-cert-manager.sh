#!/usr/bin/env bash

readonly CURRENT_DIR=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )

helm repo add jetstack https://charts.jetstack.io
helm repo update

echo -e "\n- Installing Cert Manager Helm chart...\n"
helm install cert-manager jetstack/cert-manager --namespace cert-manager --create-namespace --version v1.0.4 --set installCRDs=true --wait

echo -e "\n- Waiting for Cert Manager to be ready...\n"
kubectl -n cert-manager rollout status deploy/cert-manager-webhook

sleep 60 # due to webhook not ready, see: https://github.com/jetstack/cert-manager/issues/1873#issuecomment-683142375
kubectl apply -f "${CURRENT_DIR}/terraform/yaml/cluster-issuer.yaml"
