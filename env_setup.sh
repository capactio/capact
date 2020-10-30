#!/bin/bash
echo "GO_VERSION=^1.15.2" >>$GITHUB_ENV
echo "PROJECT_ID=projectvoltron" >>$GITHUB_ENV
echo "LINT_TIMEOUT=2m" >>$GITHUB_ENV


if [ "${GITHUB_EVENT_NAME}" = "push" ]
then
  echo "DOCKER_TAG=$(echo ${GITHUB_SHA:0:7})" >>$GITHUB_ENV
else
  PR_NUMBER=$(echo $GITHUB_REF | awk 'BEGIN { FS = "/" } ; { print $3 }')
  echo "DOCKER_TAG=PR-$(echo ${PR_NUMBER})" >>$GITHUB_ENV
  echo "DOCKER_PUSH_REPOSITORY=gcr.io/projectvoltron/pr" >>$GITHUB_ENV
fi

echo APPS="name=matrix::{\"include\":[{\"APP\":\"gateway\"},{\"APP\":\"k8s-engine\"},{\"APP\":\"och\"}]}" >>$GITHUB_ENV
echo TESTS="name=matrix::{\"include\":[{\"TEST\":\"e2e\"}]}" >>$GITHUB_ENV
echo INFRAS="name=matrix::{\"include\":[{\"INFRA\":\"json-go-gen\"}]}" >>$GITHUB_ENV

#Create & upgrade cluster related
#IMAGE_TAG is the TAG which you assign when you recreate the cluster
export NAME=dev3
export REGION=europe-north1
export BUCKET=projectvoltron_le
export ELB_IP='35.228.223.55'
echo "IMAGE_TAG=adea064" >>$GITHUB_ENV
echo "TF_VAR_region=${REGION}" >>$GITHUB_ENV
echo "TF_VAR_cluster_name=voltron-${NAME}" >>$GITHUB_ENV
echo "TF_VAR_google_compute_network_name=vpc-network-${NAME}" >>$GITHUB_ENV
echo "TF_VAR_google_compute_subnetwork_name=subnetwork-${NAME}" >>$GITHUB_ENV
echo "TF_VAR_node_pool_name=node-pool-${NAME}" >>$GITHUB_ENV
echo "TF_VAR_google_compute_subnetwork_secondary_ip_range_name1=gke-pods-${NAME}" >>$GITHUB_ENV
echo "TF_VAR_google_compute_subnetwork_secondary_ip_range_name2=gke-services-${NAME}" >>$GITHUB_ENV
echo "GET_IP_SERVICE=icanhazip.com" >>$GITHUB_ENV
echo "NAMESPACE=voltron" >>$GITHUB_ENV
echo "SERVICES=gateway engine och-public och-local" >>$GITHUB_ENV
echo "HELM_TEST_TIMEOUT=10m"  >>$GITHUB_ENV
echo "CERT_MGR_TIMEOUT=120" >>$GITHUB_ENV
echo "BUCKET=${BUCKET}" >>$GITHUB_ENV
echo "ELB_IP=${ELB_IP}" >>$GITHUB_ENV