#!/usr/bin/env bash
set -e

NAMESPACE=argo

minikube start --driver=hyperkit

echo "Creating namespace '${NAMESPACE}'..."
kubectl create ns ${NAMESPACE} || true

echo "Waiting for default ServiceAccount in namespace '${NAMESPACE}'..."
while ! kubectl get sa default -n ${NAMESPACE}
do
  sleep 1
done

echo "Install Argo with MinIO in '${NAMESPACE}' namespace"
kubectl apply -n ${NAMESPACE} -f https://raw.githubusercontent.com/argoproj/argo/stable/manifests/quick-start-minimal.yaml
