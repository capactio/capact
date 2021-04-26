#!/usr/bin/env bash
#
# This script makes Grafana accessible on local port.

# standard bash error handling
set -o nounset # treat unset variables as an error and exit immediately.
set -o errexit # exit immediately when a command fails.
set -E         # needs to be set if we want the ERR trap

CURRENT_DIR=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)
#
# shellcheck source=./hack/lib/const.sh
source "${CURRENT_DIR}/../lib/const.sh" || { echo 'Cannot load constant values.'; exit 1; }

USERNAME=$(kubectl -n "${CAPACT_NAMESPACE}" get secrets monitoring-grafana -ojsonpath="{.data.admin-user}" | base64 -d)
PASSWORD=$(kubectl -n "${CAPACT_NAMESPACE}" get secrets monitoring-grafana -ojsonpath="{.data.admin-password}" | base64 -d)

echo "URL: http://localhost:3000"
echo "Username: $USERNAME"
echo "Password: $PASSWORD"

kubectl -n "${CAPACT_NAMESPACE}" port-forward svc/monitoring-grafana 3000:80
