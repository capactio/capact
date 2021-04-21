#!/usr/bin/env bash
#
# This script installs aws-efs-csi-driver Helm chart.
#

# standard bash error handling
set -o nounset # treat unset variables as an error and exit immediately.
set -o errexit # exit immediately when a command fails.
set -E         # needs to be set if we want the ERR trap

CURRENT_DIR=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)
readonly CURRENT_DIR

helm repo add aws-efs-csi-driver https://kubernetes-sigs.github.io/aws-efs-csi-driver/
helm repo update

helm upgrade aws-efs-csi-driver aws-efs-csi-driver/aws-efs-csi-driver \
    --install \
    --namespace="kube-system" \
    --version v0.1.6 \
    --set "serviceAccount.controller.create=false" \
    --set "serviceAccount.controller.name=efs-csi-controller-sa" \
    --wait

