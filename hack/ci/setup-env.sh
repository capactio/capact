#!/usr/bin/env bash

echo "Setting up CI environmental variables..."
export NAME="dev3"

cat <<EOT >> "$GITHUB_ENV"
GO_VERSION=^1.15.2
SKIP_DEPS_INSTALLATION=false
PROJECT_ID=projectvoltron
RECREATE_CLUSTER_GCS_BUCKET=projectvoltron_cluster_recreate
GET_IP_SERVICE=ifconfig.me
TF_VAR_region=europe-north1
TF_VAR_cluster_name=voltron-${NAME}
TF_VAR_google_compute_network_name=vpc-network-${NAME}
TF_VAR_google_compute_subnetwork_name=subnetwork-${NAME}
TF_VAR_node_pool_name=node-pool-${NAME}
TF_VAR_google_compute_subnetwork_secondary_ip_range_name1=gke-pods-${NAME}
TF_VAR_google_compute_subnetwork_secondary_ip_range_name2=gke-services-${NAME}
CERT_MAX_AGE=85
CERT_NUMBER_TO_BACKUP=1
CERT_SERVICE_NAMESPACE=voltron-system
EOT


if [ "${GITHUB_EVENT_NAME}" = "pull_request" ]
then
  PR_NUMBER=$(echo "$GITHUB_REF" | awk 'BEGIN { FS = "/" } ; { print $3 }')
  echo "DOCKER_TAG=PR-${PR_NUMBER}" >> "$GITHUB_ENV"
  echo "DOCKER_REPOSITORY=gcr.io/projectvoltron/pr" >> "$GITHUB_ENV"
else
  echo "DOCKER_TAG=${GITHUB_SHA:0:7}" >> "$GITHUB_ENV"
  echo "DOCKER_REPOSITORY=gcr.io/projectvoltron" >> "$GITHUB_ENV"
fi

function returnInfraMatrixIfNeeded() {
  git diff --name-only HEAD^ HEAD > changes.txt
  while IFS= read -r file; do
    if [[ $file == hack/images/* ]]; then
      # TODO: jinja2 is a Voltron Action move it to a separate directory or create a new repo for it
      echo 'INFRAS=name=matrix::{"include":[{"INFRA":"json-go-gen"},{"INFRA":"graphql-schema-linter"},{"INFRA":"jinja2"}]}'
      break
    fi
  done <changes.txt

  rm changes.txt
}

function returnOCHJSIfNeeded() {
  git diff --name-only HEAD^ HEAD > changes.txt
  while IFS= read -r file; do
    if [[ $file == och-js/* ]]; then
      echo '{"APP":"och-js"},'
      break
    fi
  done <changes.txt

  rm changes.txt
}

# TODO: Read components to build in automated way, e.g. from directory structure
cat <<EOT >>"$GITHUB_ENV"
APPS=name=matrix::{"include":[{"APP":"gateway"},{"APP":"k8s-engine"},$(returnOCHJSIfNeeded){"APP":"argo-runner"},{"APP":"helm-runner"},{"APP":"cloudsql-runner"},{"APP":"populator"},{"APP":"terraform-runner"},{"APP":"argo-actions"}]}
TESTS=name=matrix::{"include":[{"TEST":"e2e"}]}
TOOLS=name=matrix::{"include":[{"TOOL":"ocftool"}]}
$(returnInfraMatrixIfNeeded)
EOT
