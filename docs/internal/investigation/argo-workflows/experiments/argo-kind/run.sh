#!/usr/bin/env bash
set -e

SCRIPT_PATH=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )
NAMESPACE=argo

kind create cluster

echo "Creating namespace '${NAMESPACE}'..."
kubectl create ns ${NAMESPACE} || true

echo "Waiting for default ServiceAccount in namespace '${NAMESPACE}'..."
while ! kubectl get sa default -n ${NAMESPACE}
do
  sleep 1
done

echo "Install Argo with MinIO in '${NAMESPACE}' namespace"
kubectl apply -n ${NAMESPACE} -f "${SCRIPT_PATH}/argo-minimal-kind.yaml"
