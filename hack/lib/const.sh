#!/usr/bin/env bash

#
# Dependencies
#

# Upgrade binary versions in a controlled fashion
# along with the script contents (config, flags...)
export STABLE_KUBERNETES_VERSION=v1.19.1
export STABLE_KIND_VERSION=v0.9.0
export STABLE_HELM_VERSION=v3.3.4
export STABLE_CONTROLLER_GEN_VERSION=v0.4.0
export STABLE_KUBEBUILDER_VERSION=2.3.1
export STABLE_GQLGEN_VERSION=v0.13.0

#
# Kubernetes installation
#

export VOLTRON_NAMESPACE="voltron-system"
export VOLTRON_RELEASE_NAME="voltron"
export KIND_DEV_CLUSTER_NAME="kind-dev-voltron"
export KIND_CI_CLUSTER_NAME="kind-ci-voltron"

#
# OCF
#

export DEFAULT_OCF_VERSION="0.0.1"

#
# Infra
#

export JSON_GO_GEN_IMAGE_VERSION="0.1.0"
export GRAPHQL_SCHEMA_LINTER_IMAGE_VERSION="0.1.0"
