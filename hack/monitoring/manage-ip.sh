#!/bin/bash
#
# This script manages authorized networks for private GKE cluster.

# standard bash error handling
set -o nounset # treat unset variables as an error and exit immediately.
set -o errexit # exit immediately when a command fails.
set -E         # needs to be set if we want the ERR trap

CLUSTER_NAME=${CLUSTER_NAME:-capact-dev}
REGION=${REGION:-europe-west1}

usage() {
  echo "usage: manage-ip.sh add|remove"
}

ip::get() {
  curl 'https://api.ipify.org'
}

ip::add() {
  local added_ip
  local authorized

  added_ip="$(ip::get)/32"
  authorized=$(gcloud container clusters describe "${CLUSTER_NAME}" --zone "${REGION}" \
    | yq r - 'masterAuthorizedNetworksConfig.cidrBlocks[*].cidrBlock')
  authorized=$(echo "${authorized}" \
    | tr ' ' ',' \
    | sed 's/^,//g;s/ //g')
  authorized=$(printf "%s,%s" "${authorized}" "${added_ip}" \
    | sed s/^,//g)

  echo "Setting authorized networks to ${authorized}..."

  gcloud container clusters update "${CLUSTER_NAME}" --region "${REGION}" \
    --enable-master-authorized-networks \
    --master-authorized-networks "${authorized}"
}

ip::remove() {
  local removed_ip
  local authorized

  removed_ip="$(ip::get)/32"

  authorized=$(gcloud container clusters describe "${CLUSTER_NAME}" --zone "${REGION}" \
    | yq r - 'masterAuthorizedNetworksConfig.cidrBlocks[*].cidrBlock' \
    | grep -v "${removed_ip}" || true)
  authorized=$(echo "${authorized}" \
    | tr '\n' ',' \
    | sed 's/,$/\n/')


  if [ -z "${authorized}" ]; then
    echo "Setting authorized networks to empty list..."
    gcloud container clusters update "${CLUSTER_NAME}" --region "${REGION}" --no-enable-master-authorized-networks
    gcloud container clusters update "${CLUSTER_NAME}" --region "${REGION}" --enable-master-authorized-networks
  else
    echo "Setting authorized networks to ${authorized}..."
    gcloud container clusters update "${CLUSTER_NAME}" --region "${REGION}" \
      --enable-master-authorized-networks \
      --master-authorized-networks "${authorized}"
  fi
}

case $1 in
  add)
    ip::add
    ;;

  remove)
    ip::remove
    ;;

  *)
    usage
    exit 1
  ;;
esac
