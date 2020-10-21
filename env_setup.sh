#!/bin/bash
echo "GO_VERSION=^1.15.2" >>$GITHUB_ENV
echo "PROJECT_ID=projectvoltron" >>$GITHUB_ENV


if [ "${GITHUB_EVENT_NAME}" = "push" ]
then
  echo "DOCKER_TAG=$(echo ${GITHUB_SHA:0:7})" >>$GITHUB_ENV
else
  PR_NUMBER=$(echo $GITHUB_REF | awk 'BEGIN { FS = "/" } ; { print $3 }')
  echo "DOCKER_TAG=PR-$(echo ${PR_NUMBER})" >>$GITHUB_ENV
fi

echo "TF_VAR_cluster_name=voltron-dev1" >>$GITHUB_ENV
echo "TF_VAR_location=europe-west3" >>$GITHUB_ENV
echo "" >>$GITHUB_ENV

export APPS="name=matrix::{\"include\":[{\"APP\":\"gateway\"},{\"APP\":\"k8s-engine\"},{\"APP\":\"och\"}]}"
