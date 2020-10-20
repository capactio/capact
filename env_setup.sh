#!/bin/bash
echo "GO_VERSION=^1.15.2" >>$GITHUB_ENV
echo "PROJECT_ID=projectvoltron" >>$GITHUB_ENV
echo "APPS=gateway k8s-engine och" >>$GITHUB_ENV

if [ "${GITHUB_EVENT_NAME}" = "push" ]
then
  echo "DOCKER_TAG=$(echo ${GITHUB_SHA:0:7})" >>$GITHUB_ENV
else
  PR_NUMBER=$(echo $GITHUB_REF | awk 'BEGIN { FS = "/" } ; { print $3 }')
  echo "DOCKER_TAG=PR-$(echo ${PR_NUMBER})" >>$GITHUB_ENV
fi

echo "TF_VAR_cluster_name=voltron-dev1" >>$GITHUB_ENV
echo "TF_VAR_location=europe-west3" >>$GITHUB_ENV
echo "SHA=d7d4c346aa37d5dbb6bcca7a8cfa0f0ebde54a7b" >>$GITHUB_ENV
