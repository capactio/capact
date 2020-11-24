#!/bin/bash
sudo snap install yq
IP_ADDED_JOB=$(curl ${GET_IP_SERVICE})
echo "export IP_ADDED_JOB=${IP_ADDED_JOB}" >job_ip.sh
AUTHORIZED=$(gcloud container clusters describe ${TF_VAR_cluster_name} --zone ${TF_VAR_region} |yq r - 'masterAuthorizedNetworksConfig.cidrBlocks[*].cidrBlock')
AUTHORIZED=$(echo ${AUTHORIZED} | tr ' ' ',' | sed 's/^,//g;s/ //g')
AUTHORIZED=$(printf "%s,%s/32" "${AUTHORIZED}" "${IP_ADDED_JOB}" |sed s/^,//g)
gcloud container clusters update ${TF_VAR_cluster_name} --zone ${TF_VAR_region} --enable-master-authorized-networks \
--master-authorized-networks ${AUTHORIZED}