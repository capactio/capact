#!/bin/bash
#General  settings
export "GO_VERSION=^1.15.2"
export "PROJECT_ID=projectvoltron"
export "LINT_TIMEOUT=2m"
export "SKIP_DEPS_INSTALLATION=false"
echo "GO_VERSION=${GO_VERSION}" >>$GITHUB_ENV
echo "PROJECT_ID=${PROJECT_ID}" >>$GITHUB_ENV
echo "LINT_TIMEOUT=${LINT_TIMEOUT}" >>$GITHUB_ENV
echo "SKIP_DEPS_INSTALLATION=${SKIP_DEPS_INSTALLATION}" >>$GITHUB_ENV

#Setup docker image tag upon event
if [ "${GITHUB_EVENT_NAME}" = "push" ]
then
  echo "DOCKER_TAG=$(echo ${GITHUB_SHA:0:7})" >>$GITHUB_ENV
else
  PR_NUMBER=$(echo $GITHUB_REF | awk 'BEGIN { FS = "/" } ; { print $3 }')
  echo "DOCKER_TAG=PR-$(echo ${PR_NUMBER})" >>$GITHUB_ENV
  echo "DOCKER_PUSH_REPOSITORY=gcr.io/projectvoltron/pr" >>$GITHUB_ENV
fi

APPS="gateway k8s-engine och"
APPS=$(echo 'name=matrix::{"include":['; for APP in ${APPS}; do echo {\"APP\":\"${APP}\"},; done; echo ']}' |tr -d "\n")
export APPS=$(echo ${APPS} |sed 's/}, ]/} ]/g' )
echo "APPS=${APPS}" >>$GITHUB_ENV

TESTS="e2e"
TESTS=$(echo 'name=matrix::{"include":['; for TEST in ${TESTS}; do echo {\"TEST\":\"${TEST}\"},; done; echo ']}' |tr -d "\n")
export TESTS=$(echo ${TESTS} |sed 's/}, ]/} ]/g' )
echo "TESTS=${TESTS}" >>$GITHUB_ENV

INFRAS="json-go-gen"
INFRAS=$(echo 'name=matrix::{"include":['; for INFRA in ${INFRAS}; do echo {\"INFRA\":\"${INFRA}\"},; done; echo ']}' |tr -d "\n")
export INFRAS=$(echo ${INFRAS} |sed 's/}, ]/} ]/g' )
echo "INFRAS=${INFRAS}" >>$GITHUB_ENV

#Create & upgrade cluster related
#IMAGE_TAG is the TAG which you assign when you recreate the cluster
export NAME="dev3"
export REGION="europe-north1"
export BUCKET="projectvoltron_le"
export ELB_IP='35.228.223.55'
export IMAGE_TAG="adea064"
export GET_IP_SERVICE="ifconfig.me"
export NAMESPACE="voltron"
export SERVICES="gateway engine och-public och-local"
export HELM_TEST_TIMEOUT="10m"
export CERT_MGR_TIMEOUT="120"


echo "IMAGE_TAG=${IMAGE_TAG}" >>$GITHUB_ENV
echo "TF_VAR_region=${REGION}" >>$GITHUB_ENV
echo "TF_VAR_cluster_name=voltron-${NAME}" >>$GITHUB_ENV
echo "TF_VAR_google_compute_network_name=vpc-network-${NAME}" >>$GITHUB_ENV
echo "TF_VAR_google_compute_subnetwork_name=subnetwork-${NAME}" >>$GITHUB_ENV
echo "TF_VAR_node_pool_name=node-pool-${NAME}" >>$GITHUB_ENV
echo "TF_VAR_google_compute_subnetwork_secondary_ip_range_name1=gke-pods-${NAME}" >>$GITHUB_ENV
echo "TF_VAR_google_compute_subnetwork_secondary_ip_range_name2=gke-services-${NAME}" >>$GITHUB_ENV
echo "GET_IP_SERVICE=${GET_IP_SERVICE}" >>$GITHUB_ENV
echo "NAMESPACE=${NAMESPACE}" >>$GITHUB_ENV
echo "SERVICES=${SERVICES}" >>$GITHUB_ENV
echo "HELM_TEST_TIMEOUT=${HELM_TEST_TIMEOUT}"  >>$GITHUB_ENV
echo "CERT_MGR_TIMEOUT=${CERT_MGR_TIMEOUT}" >>$GITHUB_ENV
echo "BUCKET=${BUCKET}" >>$GITHUB_ENV
echo "ELB_IP=${ELB_IP}" >>$GITHUB_ENV