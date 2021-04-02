#!/usr/bin/env bash
set -eEu

readonly CURRENT_DIR=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )

helm repo add jetstack https://charts.jetstack.io
helm repo update

echo -e "\n- Installing Cert Manager Helm chart...\n"
#shellcheck disable=SC2140
helm upgrade cert-manager jetstack/cert-manager \
  --install \
  --namespace cert-manager \
  --create-namespace \
  --version v1.0.4 \
  --values "${CURRENT_DIR}/values.yml" \
  --set installCRDs=true \
  --set serviceAccount.annotations."eks\.amazonaws\.com/role-arn"="${CERT_MANAGER_ROLE_ARN}" \
   --wait

echo -e "\n- Waiting for Cert Manager to be ready...\n"
kubectl -n cert-manager rollout status deploy/cert-manager-webhook

sleep 60 # due to webhook not ready, see: https://github.com/jetstack/cert-manager/issues/1873#issuecomment-683142375

< "${CURRENT_DIR}/cluster-issuer.yaml" \
  sed "s/{{REGION}}/${CAPACT_REGION}/g" \
  | sed "s/{{HOSTED_ZONE_ID}}/${CAPACT_HOSTED_ZONE_ID}/g" \
  | kubectl apply -f -

echo -e "\n- Cert Manager installed!\n"