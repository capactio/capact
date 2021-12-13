#!/usr/bin/env bash
# shellcheck disable=SC2154

sudo snap install yq --channel=v4/stable
gcloud config set project "${PROJECT_ID}"
AUTHORIZED=$(gcloud container clusters describe "${TF_VAR_cluster_name}" --zone "${TF_VAR_region}" \
| yq e '.masterAuthorizedNetworksConfig.cidrBlocks[].cidrBlock' - | grep -v "${IP_ADDED_JOB}") || true
AUTHORIZED=$(echo "${AUTHORIZED}" | tr '\n' ',' | sed 's/,$//') || true

if [ -z "${AUTHORIZED}" ]
  then
    gcloud container clusters update "${TF_VAR_cluster_name}" --zone "${TF_VAR_region}" --no-enable-master-authorized-networks
    gcloud container clusters update "${TF_VAR_cluster_name}" --zone "${TF_VAR_region}" --enable-master-authorized-networks
  else
    gcloud container clusters update "${TF_VAR_cluster_name}" --zone "${TF_VAR_region}" --enable-master-authorized-networks \
    --master-authorized-networks "${AUTHORIZED}"
fi
