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
  local ADDED_IP="$(ip::get)/32"
  local AUTHORIZED=$(gcloud container clusters describe "${CLUSTER_NAME}" --region "${REGION}" --format json \
    | jq -r '.masterAuthorizedNetworksConfig.cidrBlocks | .[]? | .cidrBlock' \
    | tr '\n' ',' \
    | sed 's/,$/\n/' \
    | xargs printf "%s,%s" "${ADDED_IP}" \
    | sed 's/,$//g')

  echo "Setting authorized networks to $AUTHORIZED..."

  gcloud container clusters update "${CLUSTER_NAME}" --region "${REGION}" \
    --enable-master-authorized-networks \
    --master-authorized-networks "${AUTHORIZED}"
}

ip::remove() {
  local REMOVED_IP="$(ip::get)/32"
  local AUTHORIZED=$(gcloud container clusters describe "${CLUSTER_NAME}" --region "${REGION}" --format json \
    | jq -r '.masterAuthorizedNetworksConfig.cidrBlocks | .[]? | .cidrBlock' \
    | grep -v "${REMOVED_IP}" \
    | tr '\n' ',' \
    | sed 's/,$/\n/')

  echo "Setting authorized networks to $AUTHORIZED..."

  if [ -z "${AUTHORIZED}" ]; then
    gcloud container clusters update "${CLUSTER_NAME}" --region "${REGION}" --no-enable-master-authorized-networks
    gcloud container clusters update "${CLUSTER_NAME}" --region "${REGION}" --enable-master-authorized-networks
  else
    gcloud container clusters update "${CLUSTER_NAME}" --region "${REGION}" \
      --enable-master-authorized-networks \
      --master-authorized-networks "${AUTHORIZED}"
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
