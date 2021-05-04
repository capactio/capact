#!/usr/bin/env bash
#
# This script installs aws-for-fluent-bit Helm chart.
#

# standard bash error handling
set -o nounset # treat unset variables as an error and exit immediately.
set -o errexit # exit immediately when a command fails.
set -E         # needs to be set if we want the ERR trap

CURRENT_DIR=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)
readonly CURRENT_DIR

helm repo add eks-charts https://aws.github.io/eks-charts
helm repo update

helm upgrade aws-for-fluent-bit "eks-charts/aws-for-fluent-bit" \
    --install \
    --namespace="kube-system" \
    --version v0.1.7 \
    --values "${CURRENT_DIR}/values.yml" \
    --set "cloudWatch.logGroupName=/aws/eks/${CAPACT_NAME}/logs" \
    --set "cloudWatch.region=${CAPACT_REGION}" \
