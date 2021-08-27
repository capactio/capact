# shellcheck shell=bash

#
# Library of useful utilities for CI purposes.
#

# To enforce readonly values for colors, shellcheck validation is disabled, as some of them may be not used (yet).

# shellcheck disable=SC2034
readonly RED='\033[0;31m'
# shellcheck disable=SC2034
readonly GREEN='\033[0;32m'
# shellcheck disable=SC2034
readonly INVERTED='\033[7m'
# shellcheck disable=SC2034
readonly NC='\033[0m' # No Color

# Prints first argument as header. Additionally prints current date.
shout() {
    echo -e "
#################################################################################################
# $(date)
# $1
#################################################################################################
"
}

dump_cluster_info() {
    LOGS_DIR=${ARTIFACTS:-./tmp}/logs
    mkdir -p "${LOGS_DIR}"

    echo "Dumping cluster info into ${LOGS_DIR}"
    kubectl cluster-info dump --all-namespaces --output-directory="${LOGS_DIR}"
}

# Installs kubebuilder dependency locally.
# Required envs:
#  - KUBEBUILDER_VERSION
#  - INSTALL_DIR
#
# usage: env INSTALL_DIR=/tmp KUBEBUILDER_VERSION=v0.4.0 host::install::kubebuilder
host::install::kubebuilder() {
  shout "Install the kubebuilder ${KUBEBUILDER_VERSION} locally to a tempdir..."

  export KUBEBUILDER_ASSETS="${INSTALL_DIR}/kubebuilder/bin"

  os=$(host::os)
  arch=$(host::arch)
  name="kubebuilder_${KUBEBUILDER_VERSION}_${os}_${arch}"

  pushd "${INSTALL_DIR}" || return

  # download the release
  curl -L -O "https://github.com/kubernetes-sigs/kubebuilder/releases/download/v${KUBEBUILDER_VERSION}/${name}.tar.gz"

  # extract the archive
  tar -zxvf "${name}".tar.gz
  mv "${name}" kubebuilder

  popd || return

  echo -e "${GREEN}√ install kubebuilder${NC}"
}

host::os() {
  local host_os
  case "$(uname -s)" in
    Darwin)
      host_os=darwin
      ;;
    Linux)
      host_os=linux
      ;;
    *)
      echo "Unsupported host OS. Must be Linux or Mac OS X."
      exit 1
      ;;
  esac
  echo "${host_os}"
}

host::arch() {
  local host_arch
  case "$(uname -m)" in
    x86_64*)
      host_arch=amd64
      ;;
    i?86_64*)
      host_arch=amd64
      ;;
    amd64*)
      host_arch=amd64
      ;;
    aarch64*)
      host_arch=arm64
      ;;
    arm64*)
      host_arch=arm64
      ;;
    arm*)
      host_arch=arm
      ;;
    ppc64le*)
      host_arch=ppc64le
      ;;
    *)
      echo "Unsupported host arch. Must be x86_64, arm, arm64, or ppc64le."
      exit 1
      ;;
  esac
  echo "${host_arch}"
}


# Installs helm dependencies locally.
# Required envs:
#  - HELM_VERSION
#  - INSTALL_DIR
#
# usage: env INSTALL_DIR=/tmp HELM_VERSION=v2.14.3 host::install::helm
host::install::helm() {
    mkdir -p "${INSTALL_DIR}/bin"
    export PATH="${INSTALL_DIR}/bin:${PATH}"

    pushd "${INSTALL_DIR}" || return

    shout "- Install helm ${HELM_VERSION} locally to a tempdir..."
    curl -fsSL -o "${INSTALL_DIR}"/get_helm.sh https://raw.githubusercontent.com/helm/helm/master/scripts/get-helm-3
    chmod 700 "${INSTALL_DIR}"/get_helm.sh
    env HELM_INSTALL_DIR="${INSTALL_DIR}/bin" ./get_helm.sh \
        --version "${HELM_VERSION}" \
        --no-sudo

    popd || return
}

#
# 'helm' functions
#
helm::version(){
  helm version --short -c | tr -d  'Client: '
}

#
# Capact functions
#

#  - CLUSTER_NAME
#  - CLUSTER_TYPE
#  - REPO_DIR
#  - MULTINODE_CLUSTER
capact::create_cluster() {
    shout "- Creating K8s cluster..."
    if [[ "${MULTINODE_CLUSTER:-"false"}" == "true" ]]; then
      CLUSTER_CONFIG_FLAG=--cluster-config="${REPO_DIR}/hack/cluster-config/kind/config-multinode.yaml"
    fi
    # shellcheck disable=SC2086
    capact::cli env create ${CLUSTER_TYPE} \
      --name="${CLUSTER_NAME}" \
      ${CLUSTER_CONFIG_FLAG:-} \
      --wait=5m
}

#  - CLUSTER_NAME
#  - CLUSTER_TYPE
capact::delete_cluster() {
    shout "- Deleting K8s cluster..."
    # shellcheck disable=SC2086
    capact::cli env delete ${CLUSTER_TYPE} --name="${CLUSTER_NAME}"
}


# Installs Capact charts. If they are already installed, it upgrades them.
#
# Required envs:
#  - DOCKER_REPOSITORY
#  - DOCKER_TAG
#  - REPO_DIR
#  - CAPACT_NAMESPACE
#  - CAPACT_VERSION
#  - CLUSTER_TYPE
#  - CLUSTER_NAME
#  - ENABLE_POPULATOR - if set to true then database populator will be enabled and it will populate database with manifests
#  - USE_TEST_SETUP - if set to true, then a test policy is configured
#  - INCREASE_RESOURCE_LIMITS - if set to true, then the components will use higher resource requests and limits
#  - HUB_MANIFESTS_SOURCE_REPO_REF - set this to override the Git branch from which the source manifests are populated
#  - HUB_MANIFESTS_SOURCE_REPO_URL - set this to override the Git URL from which the source manifests are populated
#  - CAPACT_HELM_REPO_URL - set this to override the Helm repo url used to install Capact.
#  - ENABLE_HOSTS_UPDATE - if set to true, /etc/hosts is updated
#  - ENABLE_ADDING_TRUSTED_CERT - if set to true, add Capact self-signed TLS certificate as trusted
capact::install() {
    pushd "${REPO_DIR}" || return

    export CAPACT_VERSION=${CAPACT_VERSION:-@local}
    export ENABLE_POPULATOR=${ENABLE_POPULATOR:-${CAPACT_ENABLE_POPULATOR}}
    export USE_TEST_SETUP=${USE_TEST_SETUP:-${CAPACT_USE_TEST_SETUP}}
    export INCREASE_RESOURCE_LIMITS=${INCREASE_RESOURCE_LIMITS:-${CAPACT_INCREASE_RESOURCE_LIMITS}}
    export PRINT_INSECURE_NOTES=${PRINT_INSECURE_NOTES:-"false"}
    export ENABLE_HOSTS_UPDATE=${ENABLE_HOSTS_UPDATE:-"true"}
    export ENABLE_ADDING_TRUSTED_CERT=${ENABLE_ADDING_TRUSTED_CERT:-"true"}
    export HUB_MANIFESTS_SOURCE_REPO_REF=${HUB_MANIFESTS_SOURCE_REPO_REF:-${CAPACT_HUB_MANIFESTS_SOURCE_REPO_REF}}
    export HUB_MANIFESTS_SOURCE_REPO_URL=${HUB_MANIFESTS_SOURCE_REPO_URL:-${CAPACT_HUB_MANIFESTS_SOURCE_REPO_URL}}
    export COMPONENTS="neo4j,ingress-nginx,argo,cert-manager,capact"
    export INGRESS_CONTROLLER_OVERRIDES=${INGRESS_CONTROLLER_OVERRIDES:=""}
    export CAPACT_OVERRIDES=${CAPACT_OVERRIDES:=""}
    export CAPACT_INSTALL_ADDITIONAL_OPTS=""

    CAPACT_OVERRIDES+=",global.containerRegistry.path=${DOCKER_REPOSITORY}"
    CAPACT_OVERRIDES+=",global.containerRegistry.overrideTag=${DOCKER_TAG}"
    CAPACT_OVERRIDES+=",hub-public.populator.enabled=${ENABLE_POPULATOR}"
    CAPACT_OVERRIDES+=",engine.testSetup.enabled=${USE_TEST_SETUP}"
    CAPACT_OVERRIDES+=",notes.printInsecure=${PRINT_INSECURE_NOTES}"

    if [[ "${DISABLE_KUBED_INSTALLATION:-"false"}" == "true" ]]; then
      shout "Skipping kubed installation cause DISABLE_KUBED_INSTALLATION is set to true."
    else
      COMPONENTS+=",kubed"
    fi

    if [[ "${DISABLE_MONITORING_INSTALLATION:-"false"}" == "true" ]]; then
      shout "Skipping monitoring installation cause DISABLE_MONITORING_INSTALLATION is set to true."
    else
      COMPONENTS+=",monitoring"
    fi

    if [ -n "${HUB_MANIFESTS_SOURCE_REPO_REF:-}" ]; then
      CAPACT_OVERRIDES+=",hub-public.populator.manifestsLocation.branch=${HUB_MANIFESTS_SOURCE_REPO_REF}"
    fi

    if [ -n "${HUB_MANIFESTS_SOURCE_REPO_URL:-}" ]; then
      CAPACT_OVERRIDES+=",hub-public.populator.manifestsLocation.repository=${HUB_MANIFESTS_SOURCE_REPO_URL}"
    fi

    if [[ "${BUILD_IMAGES:-"true"}" == "false" ]]; then
      BUILD_IMAGES_FLAG=--build-image=""
    fi

    if [ -n "${CAPACT_HELM_REPO_URL:-}" ]; then
      CAPACT_INSTALL_ADDITIONAL_OPTS="${CAPACT_INSTALL_ADDITIONAL_OPTS} --helm-repo-url=${CAPACT_HELM_REPO_URL}"
    fi

    # shellcheck disable=SC2086
    capact::cli install --verbose \
        --environment="${CLUSTER_TYPE}" \
        --name="${CLUSTER_NAME}" \
        --namespace="${CAPACT_NAMESPACE}" \
        --capact-overrides="${CAPACT_OVERRIDES}" \
        --ingress-controller-overrides="${INGRESS_CONTROLLER_OVERRIDES}" \
        --increase-resource-limits="${INCREASE_RESOURCE_LIMITS}" \
        --update-hosts-file="${ENABLE_HOSTS_UPDATE}" \
        --update-trusted-certs="${ENABLE_ADDING_TRUSTED_CERT}" \
        --install-component="${COMPONENTS}" \
        ${BUILD_IMAGES_FLAG:-} \
        --version="${CAPACT_VERSION}" \
        ${CAPACT_INSTALL_ADDITIONAL_OPTS}
}

# Required envs:
#  - REPO_DIR
capact::cli()  {
  local os
  local arch
  local default_path
  local cli

  os=$(host::os)
  arch=$(host::arch)
  default_path="${REPO_DIR}/bin/capact-${os}-${arch}"
  cli=${CAPACT_BINARY:-${default_path}}
  if [ ! -f "${cli}" ]; then
    shout "capact cli not found, trying to build one..."
    make build-tool-cli
    cli=${default_path}
  fi

  ${cli} "$@"
}

# Installs kind and helm dependencies locally.
# Required envs:
#  - MINIMAL_VERSION
#  - CURRENT_VERSION
#
# usage: env MINIMAL_VERSION=v3.3.4 CURRENT_VERSION=v2.16.9 capact::version_supported
capact::version_supported(){
  printf '%s\n%s\n' "$CURRENT_VERSION" "$MINIMAL_VERSION" | sort -rVC
}

capact::validate::tools() {
  shout "- Validating tools versions..."
  local current_helm_version
  local wrong_versions

  current_helm_version=$(helm::version)
  wrong_versions=false

  echo "Current helm version: $current_helm_version, recommended helm version: $STABLE_HELM_VERSION"

  if ! MINIMAL_VERSION="${STABLE_HELM_VERSION}" CURRENT_VERSION="${current_helm_version}" capact::version_supported; then
      wrong_versions=true
      echo "Unsupported helm version $current_helm_version. Must be at least $STABLE_HELM_VERSION"
  fi
  [ ${wrong_versions} == false ]
}
