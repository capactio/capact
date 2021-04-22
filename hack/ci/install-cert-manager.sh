#!/usr/bin/env bash

CURRENT_DIR=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)
readonly CURRENT_DIR

sleep 60 # due to webhook not ready, see: https://github.com/jetstack/cert-manager/issues/1873#issuecomment-683142375
kubectl apply -f "${CURRENT_DIR}/terraform/yaml/cluster-issuer.yaml"
