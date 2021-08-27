# shellcheck shell=bash
# shellcheck disable=SC2034

#
# Dependencies
#

# Upgrade binary versions in a controlled fashion
# along with the script contents (config, flags...)
readonly STABLE_CONTROLLER_GEN_VERSION=v0.6.2
readonly STABLE_KUBEBUILDER_VERSION=2.3.2
readonly STABLE_GQLGEN_VERSION=v0.13.0
readonly STABLE_HELM_VERSION=v3.6.3

#
# Kubernetes installation
#

readonly CAPACT_NAMESPACE="capact-system"
readonly CAPACT_RELEASE_NAME="capact"
readonly DEV_CLUSTER_NAME="dev-capact"
readonly KIND_CI_CLUSTER_NAME="kind-ci-capact"

#
# OCF
#

readonly DEFAULT_OCF_VERSION="0.0.1"

#
# Infra
#

readonly JSON_GO_GEN_IMAGE_TAG="PR-310"
readonly GRAPHQL_SCHEMA_LINTER_IMAGE_TAG="PR-310"

#
# Development
#

readonly CAPACT_ENABLE_POPULATOR="true"
readonly CAPACT_USE_TEST_SETUP="false"

#
# Cluster Configuration
#

readonly CAPACT_INCREASE_RESOURCE_LIMITS="true"
readonly CAPACT_HUB_MANIFESTS_SOURCE_REPO_URL="github.com/capactio/hub-manifests"
# The git ref to checkout. It can point to a commit SHA, a branch name, or a tag.
# If you want to use your forked version, remember to update CAPACT_HUB_MANIFESTS_SOURCE_REPO_URL  respectively.
readonly CAPACT_HUB_MANIFESTS_SOURCE_REPO_REF="main"
