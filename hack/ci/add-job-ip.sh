#!/usr/bin/env bash
# shellcheck disable=SC2154

sudo snap install yq --channel=v3/stable
IP_ADDED_JOB=$(printf "%s/32" "$(curl "${GET_IP_SERVICE}")")
AUTHORIZED=$(gcloud container clusters describe "${TF_VAR_cluster_name}" --zone "${TF_VAR_region}" |yq r - 'masterAuthorizedNetworksConfig.cidrBlocks[*].cidrBlock')
AUTHORIZED=$(echo "${AUTHORIZED}" | tr ' ' ',' | sed 's/^,//g;s/ //g')
AUTHORIZED=$(printf "%s,%s" "${AUTHORIZED}" "${IP_ADDED_JOB}" |sed s/^,//g)
gcloud container clusters update "${TF_VAR_cluster_name}" --zone "${TF_VAR_region}" --enable-master-authorized-networks \
--master-authorized-networks "${AUTHORIZED}"
echo "::set-output name=JOB_IP::$IP_ADDED_JOB"
