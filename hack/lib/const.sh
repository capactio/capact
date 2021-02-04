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
readonly STABLE_CONTROLLER_GEN_VERSION=v0.4.0
readonly STABLE_KUBEBUILDER_VERSION=2.3.1
readonly STABLE_GQLGEN_VERSION=v0.13.0

#
# Kubernetes installation
#

readonly VOLTRON_NAMESPACE="voltron-system"
readonly VOLTRON_RELEASE_NAME="voltron"
readonly KIND_DEV_CLUSTER_NAME="kind-dev-voltron"
readonly KIND_CI_CLUSTER_NAME="kind-ci-voltron"

#
# OCF
#

readonly DEFAULT_OCF_VERSION="0.0.1"

#
# Infra
#

readonly JSON_GO_GEN_IMAGE_VERSION="0.1.1"
readonly GRAPHQL_SCHEMA_LINTER_IMAGE_VERSION="0.1.0"

#
# Development
#

readonly VOLTRON_MOCK_OCH_GRAPHQL="false"
readonly VOLTRON_MOCK_ENGINE_GRAPHQL="false"
readonly VOLTRON_ENABLE_POPULATOR="true"
