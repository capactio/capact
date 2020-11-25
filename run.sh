#!/bin/bash

# TODO: Remove the file

make build-app-image-helm-runner

kind load docker-image gcr.io/projectvoltron/helm-runner:latest --name kind-dev-voltron

kubectl delete job helm-runner-example
sleep 3

echo "# Creating resources..."
kubectl apply -f testing.yaml

echo "# Waiting for pod ready..."
kubectl wait --for=condition=ready pod --selector=helm-runner-example=true

echo "# Watching logs from Runner..."
kubectl logs -l helm-runner-example=true -f -c runner

echo "# Printing logs from printer..."
kubectl logs -l helm-runner-example=true -f -c printer
