#!/usr/bin/env bash

echo "Setting up CI environmental variables..."
export NAME="stage"

# LOAD_BALANCER_EXTERNAL_IP is a reserved IP in "External IP addresses" on GCP. It needs to be in the same region.
# Remember when changing LOAD_BALANCER_EXTERNAL_IP to update record A in the Cloud DNS for gateway
cat <<EOT >> "$GITHUB_ENV"
GO_VERSION=1.17.x
GOLANGCI_LINT_VERSION=v1.41.1
GOLANGCI_LINT_TIMEOUT=10m
SKIP_DEPS_INSTALLATION=false
PROJECT_ID=capact
RECREATE_CLUSTER_GCS_BUCKET=capact-stage-cluster-recreate
GET_IP_SERVICE=ifconfig.me
TF_VAR_region=europe-west1
TF_VAR_cluster_name=capact-${NAME}
TF_VAR_google_compute_network_name=vpc-network-${NAME}
TF_VAR_google_compute_subnetwork_name=subnetwork-${NAME}
TF_VAR_node_pool_name=node-pool-${NAME}
TF_VAR_google_compute_subnetwork_secondary_ip_range_name1=gke-pods-${NAME}
TF_VAR_google_compute_subnetwork_secondary_ip_range_name2=gke-services-${NAME}
LOAD_BALANCER_EXTERNAL_IP=34.78.197.21
CERT_MAX_AGE=85
CERT_NUMBER_TO_BACKUP=1
CERT_SERVICE_NAMESPACE=capact-system
EOT

if [ "${GITHUB_EVENT_NAME}" = "pull_request_target" ]
then
  echo "DOCKER_TAG=latest" >> "$GITHUB_ENV"
  echo "DOCKER_REPOSITORY=gcr.io/sm-cluster-dev/pr" >> "$GITHUB_ENV"
else
  echo "DOCKER_TAG=latest" >> "$GITHUB_ENV"
  echo "DOCKER_REPOSITORY=gcr.io/sm-cluster-dev" >> "$GITHUB_ENV"
fi

# TODO: Read components to build in automated way, e.g. from directory structure
cat <<EOT >>"$GITHUB_ENV"
APPS=name=matrix::{"include":[{"APP":"gateway"},{"APP":"k8s-engine"},{"APP":"hub-js"},{"APP":"argo-runner"},{"APP":"helm-runner"},{"APP":"populator"},{"APP":"terraform-runner"},{"APP":"gcplist-runner"},{"APP":"argo-actions"},{"APP":"gitlab-api-runner"},{"APP":"secret-storage-backend"},{"APP":"helm-storage-backend"},{"APP":"ti-value-fetcher"}]}
TESTS=name=matrix::{"include":[{"TEST":"e2e"}, {"TEST":"local-hub"}]}
INFRAS=name=matrix::{"include":[{"INFRA":"json-go-gen"},{"INFRA":"graphql-schema-linter"},{"INFRA":"jinja2"},{"INFRA":"merger"}]}
EOT
