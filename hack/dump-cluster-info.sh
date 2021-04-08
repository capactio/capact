#!/usr/bin/env bash
#
# This script dumps the cluster logs.
#

# standard bash error handling
set -o nounset # treat unset variables as an error and exit immediately.
set -o errexit # exit immediately when a command fails.
set -E         # needs to be set if we want the ERR trap

readonly CURRENT_DIR=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)

# shellcheck source=./hack/lib/utilities.sh
source "${CURRENT_DIR}/lib/utilities.sh" || { echo 'Cannot load CI utilities.'; exit 1; }

shout '- Dumping cluster info...'
dump_cluster_info || true
shout '- Cluster info dumped'
