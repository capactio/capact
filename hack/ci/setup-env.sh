#!/usr/bin/env bash

echo "Setting up CI environmental variables..."
export NAME="dev3"
export APPS="gateway k8s-engine och"
export TESTS="e2e"
export INFRAS="json-go-gen"

cat <<EOT >> "$GITHUB_ENV"
GO_VERSION=^1.15.2
SKIP_DEPS_INSTALLATION=false
PROJECT_ID=projectvoltron
BUCKET=projectvoltron_cluster_recreate
ELB_IP=35.228.223.55
IMAGE_TAG=adea064
GET_IP_SERVICE=ifconfig.me
NAMESPACE=voltron-system
HELM_TEST_TIMEOUT=10m
HELM_CHARTS_PATH=deploy/kubernetes/charts/voltron/charts
CERT_MGR_TIMEOUT=120s
TF_VAR_region=europe-north1
TF_VAR_cluster_name=voltron-${NAME}
TF_VAR_google_compute_network_name=vpc-network-${NAME}
TF_VAR_google_compute_subnetwork_name=subnetwork-${NAME}
TF_VAR_node_pool_name=node-pool-${NAME}
TF_VAR_google_compute_subnetwork_secondary_ip_range_name1=gke-pods-${NAME}
TF_VAR_google_compute_subnetwork_secondary_ip_range_name2=gke-services-${NAME}
COMPONENTS="gateway engine och-public och-local"
SERVICES="voltron"
CERT_MAX_AGE=85
EOT


if [ "${GITHUB_EVENT_NAME}" = "push" ]
then
  echo "DOCKER_TAG=${GITHUB_SHA:0:7}" >> "$GITHUB_ENV"
  echo "DOCKER_REPOSITORY=gcr.io/projectvoltron" >> "$GITHUB_ENV"
else
  PR_NUMBER=$(echo "$GITHUB_REF" | awk 'BEGIN { FS = "/" } ; { print $3 }')
  echo "DOCKER_TAG=PR-${PR_NUMBER}" >> "$GITHUB_ENV"
  echo "DOCKER_REPOSITORY=gcr.io/projectvoltron/pr" >> "$GITHUB_ENV"
fi


APPS=$(echo 'name=matrix::{"include":['; for APP in ${APPS}; do echo {\"APP\":\"${APP}\"},; done; echo ']}' |tr -d "\n")
export APPS=$(echo ${APPS} |sed 's/}, ]/} ]/g' )
echo "APPS=${APPS}" >>"$GITHUB_ENV"

TESTS=$(echo 'name=matrix::{"include":['; for TEST in ${TESTS}; do echo {\"TEST\":\"${TEST}\"},; done; echo ']}' |tr -d "\n")
export TESTS=$(echo ${TESTS} |sed 's/}, ]/} ]/g' )
echo "TESTS=${TESTS}" >>"$GITHUB_ENV"

INFRAS=$(echo 'name=matrix::{"include":['; for INFRA in ${INFRAS}; do echo {\"INFRA\":\"${INFRA}\"},; done; echo ']}' |tr -d "\n")
export INFRAS=$(echo ${INFRAS} |sed 's/}, ]/} ]/g')
echo "INFRAS=${INFRAS}" >>"$GITHUB_ENV"



