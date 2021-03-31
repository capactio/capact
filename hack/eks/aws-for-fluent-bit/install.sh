#!/bin/bash
set -eEu

readonly CURRENT_DIR=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )

echo '\nInstalling aws-for-fluent-bit Helm chart...\n'

helm repo add eks-charts https://aws.github.io/eks-charts

helm upgrade aws-for-fluent-bit "eks-charts/aws-for-fluent-bit" \
    --install \
    --namespace="kube-system" \
    --version v0.1.6 \
    --values "${CURRENT_DIR}/values.yml" \
    --set "cloudWatch.logGroupName=/aws/eks/${CAPACT_NAME}/logs"\
    --set "cloudWatch.region=${CAPACT_REGION}"\
    --wait

echo "\nWaiting for aws-for-fluent-bit to be ready...\n"

kubectl wait --namespace kube-system \
  --for=condition=ready pod \
  --selector=app.kubernetes.io/name=aws-for-fluent-bit \
  --timeout=90s

echo "\naws-for-fluent-bit installed!\n"
