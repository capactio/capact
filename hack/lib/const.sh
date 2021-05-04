# shellcheck shell=bash
# shellcheck disable=SC2034

#
# Dependencies
#

# Upgrade binary versions in a controlled fashion
# along with the script contents (config, flags...)
readonly STABLE_KUBERNETES_VERSION=v1.19.1
readonly STABLE_KIND_VERSION=v0.9.0
readonly STABLE_HELM_VERSION=v3.3.4
readonly STABLE_CONTROLLER_GEN_VERSION=v0.5.0
readonly STABLE_KUBEBUILDER_VERSION=2.3.1
readonly STABLE_GQLGEN_VERSION=v0.13.0

#
# Kubernetes installation
#

readonly CAPACT_NAMESPACE="capact-system"
readonly CAPACT_RELEASE_NAME="capact"
readonly KIND_DEV_CLUSTER_NAME="kind-dev-capact"
readonly KIND_CI_CLUSTER_NAME="kind-ci-capact"

#
# OCF
#

readonly DEFAULT_OCF_VERSION="0.0.1"

#
# Infra
#

readonly JSON_GO_GEN_IMAGE_TAG="PR-274"
readonly GRAPHQL_SCHEMA_LINTER_IMAGE_TAG="PR-274"

#
# Development
#

readonly CAPACT_ENABLE_POPULATOR="true"
readonly CAPACT_USE_TEST_SETUP="false"

#
# Cluster Configuration
#

readonly CAPACT_INCREASE_RESOURCE_LIMITS="true"
