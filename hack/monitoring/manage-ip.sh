#!/bin/bash
set -e

CLUSTER_NAME=${CLUSTER_NAME:-voltron-dev3}
REGION=${REGION:-europe-north1}

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
  authorized=$(gcloud container clusters describe "${CLUSTER_NAME}" --region "${REGION}" --format json \
    | jq -r '.masterAuthorizedNetworksConfig.cidrBlocks | .[]? | .cidrBlock' \
    | tr '\n' ',' \
    | sed 's/,$/\n/' \
    | xargs printf "%s,%s" "${added_ip}" \
    | sed 's/,$//g')

  echo "Setting authorized networks to ${authorized}..."

  gcloud container clusters update "${CLUSTER_NAME}" --region "${REGION}" \
    --enable-master-authorized-networks \
    --master-authorized-networks "${authorized}"
}

ip::remove() {
  local removed_ip
  local authorized

  removed_ip="$(ip::get)/32"
  authorized=$(gcloud container clusters describe "${CLUSTER_NAME}" --region "${REGION}" --format json \
    | jq -r '.masterAuthorizedNetworksConfig.cidrBlocks | .[]? | .cidrBlock' \
    | grep -v "${removed_ip}" \
    | tr '\n' ',' \
    | sed 's/,$/\n/')

  echo "Setting authorized networks to ${authorized}..."

  if [ -z "${authorized}" ]; then
    gcloud container clusters update "${CLUSTER_NAME}" --region "${REGION}" --no-enable-master-authorized-networks
    gcloud container clusters update "${CLUSTER_NAME}" --region "${REGION}" --enable-master-authorized-networks
  else
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
