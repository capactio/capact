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

# Installs kind dependency locally.
# Required envs:
#  - KIND_VERSION
#  - INSTALL_DIR
#
# usage: env INSTALL_DIR=/tmp KIND_VERSION=v0.4.0 host::install::kind
host::install::kind() {
    mkdir -p "${INSTALL_DIR}/bin"
    export PATH="${INSTALL_DIR}/bin:${PATH}"

    pushd "${INSTALL_DIR}" || return

    os=$(host::os)
    arch=$(host::arch)

    shout "- Install kind ${KIND_VERSION} locally to a tempdir..."

    curl -sSLo kind "https://github.com/kubernetes-sigs/kind/releases/download/${KIND_VERSION}/kind-${os}-${arch}"
    chmod +x kind
    mv kind "${INSTALL_DIR}/bin"

    popd || return
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


# Installs kind and helm dependencies locally.
# Required envs:
#  - HELM_VERSION
#  - INSTALL_DIR
#
# usage: env INSTALL_DIR=/tmp HELM_VERSION=v2.14.3 host::install::kind
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
# 'kind'(kubernetes-in-docker) functions
#
# Required environments variables for all 'kind' commands:
#  - KIND_CLUSTER_NAME
#  - REPO_DIR
kind::create_cluster() {
    shout "- Creating K8s cluster..."
    kind create cluster \
      --name="${KIND_CLUSTER_NAME}" \
      --image="kindest/node:${KUBERNETES_VERSION}" \
      --config="${REPO_DIR}/hack/cluster-config/kind/config.yaml" \
      --wait=5m
}

kind::delete_cluster() {
    shout "- Deleting K8s cluster..."
    kind delete cluster --name="${KIND_CLUSTER_NAME}"
}

# Arguments:
#   $1 - list of docker images in format "name:tag [name:tag...]"
kind::load_images() {
  for id in $1; do
    kind load docker-image "$id" --name="${KIND_CLUSTER_NAME}"
  done
}

kind::version() {
  echo "v$(kind version -q)"
}

#
# 'helm' functions
#
helm::version(){
  helm version --short -c | tr -d  'Client: '
}


#
# Docker functions
#

# Arguments:
#   $1 - reference filter
docker::list_images() {
  names=$(docker image ls --filter=reference="$1" --format="{{.Repository}}:{{.Tag}}")

  echo "$names"
}

# Arguments:
#   $1 - image list
docker::delete_images() {
  # images have to be passed as multiple arguments
  # shellcheck disable=SC2206
  images=($1)
  docker rmi "${images[@]}"
}

#
# Voltron functions
#

# Installs Voltron charts. If they are already installed, it upgrades them.
#
# Required envs:
#  - DOCKER_REPOSITORY
#  - DOCKER_TAG
#  - REPO_DIR
#  - VOLTRON_NAMESPACE
#  - VOLTRON_RELEASE_NAME
#  - CLUSTER_TYPE
#  - MOCK_ENGINE_GRAPHQL - if set to true then predifined values are used in engine graphql
#  - ENABLE_POPULATOR - if set to true then database populator will be enabled and it will populate database with manifests
#  - USE_TEST_SETUP - if set to true, then the OCH manifests are populated from `test/och-content` and a test policy is configured
#  - INCREASE_RESOURCE_LIMITS - if set to true, then the components will use higher resource requests and limits
voltron::install_upgrade::charts() {
    readonly K8S_DEPLOY_DIR="${REPO_DIR}/deploy/kubernetes"
    readonly CLUSTER_CONFIG_DIR="${REPO_DIR}/hack/cluster-config"
    readonly KIND_CONFIG_DIR="${CLUSTER_CONFIG_DIR}/kind"

    export MOCK_ENGINE_GRAPHQL=${MOCK_ENGINE_GRAPHQL:-${VOLTRON_MOCK_ENGINE_GRAPHQL}}
    export ENABLE_POPULATOR=${ENABLE_POPULATOR:-${VOLTRON_ENABLE_POPULATOR}}
    export USE_TEST_SETUP=${USE_TEST_SETUP:-${VOLTRON_USE_TEST_SETUP}}
    export INCREASE_RESOURCE_LIMITS=${INCREASE_RESOURCE_LIMITS:-${VOLTRON_INCREASE_RESOURCE_LIMITS}}

    # TODO: Prepare overrides for Github Actions CI and use the "higher resource requests and limits" overrides by default in charts
    if [[ "${INCREASE_RESOURCE_LIMITS}" == "true" ]]; then
      shout "Using higher resource requests and limits from ${CLUSTER_CONFIG_DIR}"
    fi

    shout "- Applying Voltron CRDs..."
    kubectl apply -f "${K8S_DEPLOY_DIR}"/crds

    voltron::install_upgrade::neo4j

    voltron::install_upgrade::ingress_controller

    voltron::install_upgrade::argo

     if [[ "${DISABLE_KUBED_INSTALLATION:-"false"}" == "true" ]]; then
      shout "Skipping kubed installation cause DISABLE_KUBED_INSTALLATION is set to true."
    else
      voltron::install_upgrade::kubed
      voltron::synchronize::minio_secret
    fi

    if [[ "${DISABLE_MONITORING_INSTALLATION:-"false"}" == "true" ]]; then
      shout "Skipping monitoring installation cause DISABLE_MONITORING_INSTALLATION is set to true."
    else
      voltron::install_upgrade::monitoring
    fi

    shout "- Installing Voltron Helm chart from sources [wait: true]..."
    echo -e "- Using DOCKER_REPOSITORY=$DOCKER_REPOSITORY and DOCKER_TAG=$DOCKER_TAG\n"

    if [[ "${CLUSTER_TYPE}" == "KIND" ]]; then
      readonly VOLTRON_OVERRIDES="${KIND_CONFIG_DIR}/overrides.voltron.yaml"
      echo -e "- Applying overrides from ${VOLTRON_OVERRIDES}\n"
    else # currently, only KIND needs custom settings
      readonly VOLTRON_OVERRIDES=""
    fi

    if [[ "${INCREASE_RESOURCE_LIMITS}" == "true" ]]; then
      readonly VOLTRON_RESOURCE_OVERRIDES="${CLUSTER_CONFIG_DIR}/overrides.voltron.higher-res-limits.yaml"
      echo -e "- Applying overrides from ${VOLTRON_RESOURCE_OVERRIDES}\n"
    else
      readonly VOLTRON_RESOURCE_OVERRIDES=""
    fi

    if [ "${USE_TEST_SETUP}" == "true" ]; then
      readonly VOLTRON_TEST_SETUP_OVERRIDES="${CLUSTER_CONFIG_DIR}/overrides.voltron.test-setup.yaml"
      echo -e "- Applying overrides from ${VOLTRON_TEST_SETUP_OVERRIDES}\n"
    else
      readonly VOLTRON_TEST_SETUP_OVERRIDES=""
    fi

    # CUSTOM_VOLTRON_SET_FLAGS cannot be quoted
    # shellcheck disable=SC2086
    helm upgrade "${VOLTRON_RELEASE_NAME}" "${K8S_DEPLOY_DIR}/charts/voltron" \
        --install \
        --create-namespace \
        --namespace="${VOLTRON_NAMESPACE}" \
        --set global.containerRegistry.path="${DOCKER_REPOSITORY}" \
        --set global.containerRegistry.overrideTag="${DOCKER_TAG}" \
        --set global.mockEngineGraphQL="${MOCK_ENGINE_GRAPHQL}" \
        --set och-public.populator.enabled="${ENABLE_POPULATOR}" \
        ${CUSTOM_VOLTRON_SET_FLAGS:-}  \
        -f "${VOLTRON_TEST_SETUP_OVERRIDES}" \
        -f "${VOLTRON_OVERRIDES}" \
        -f "${VOLTRON_RESOURCE_OVERRIDES}" \
        --wait
}

voltron::install_upgrade::monitoring() {
    # not waiting as Helm Charts installation takes additional ~3 minutes. To proceed further we need only monitoring CRDs.
    shout "- Installing monitoring Helm chart [wait: false]..."
    helm upgrade monitoring "${K8S_DEPLOY_DIR}/charts/monitoring" \
        --install \
        --create-namespace \
        --namespace="monitoring"
}

voltron::install_upgrade::kubed() {
    # not waiting as it is not needed.
    shout "- Installing kubed Helm chart [wait: false]..."
    helm upgrade kubed "${K8S_DEPLOY_DIR}/charts/kubed" \
        --install \
        --create-namespace \
        --namespace="kubed"
}

voltron::install_upgrade::neo4j() {
    shout "- Installing Neo4j Helm chart..."

    if [[ "${INCREASE_RESOURCE_LIMITS}" == "true" ]]; then
      readonly NEO4J_RESOURCE_OVERRIDES="${CLUSTER_CONFIG_DIR}/overrides.neo4j.higher-res-limits.yaml"
      echo -e "- Applying overrides from ${NEO4J_RESOURCE_OVERRIDES}\n"
    else
      readonly NEO4J_RESOURCE_OVERRIDES=""
    fi

    helm upgrade neo4j "${K8S_DEPLOY_DIR}/charts/neo4j" \
        --install \
        --create-namespace \
        --namespace="neo4j" \
        -f "${NEO4J_RESOURCE_OVERRIDES}" \
        --wait

    echo -e "\n- Waiting for Neo4j database to be ready...\n"
    kubectl wait --namespace neo4j \
      --for=condition=ready pod \
      --selector=app.kubernetes.io/component=core \
      --timeout=300s
}

voltron::install_upgrade::ingress_controller() {
    # waiting as admission webhooks server is required to be available during further installation steps
    shout "- Installing Ingress NGINX Controller Helm chart [wait: true]..."

    if [[ "${CLUSTER_TYPE}" == "KIND" ]]; then
      readonly INGRESS_CTRL_OVERRIDES="${KIND_CONFIG_DIR}/overrides.ingress-nginx.yaml"
      echo -e "- Applying overrides from ${INGRESS_CTRL_OVERRIDES}\n"
    else # currently, only KIND needs custom settings
      readonly INGRESS_CTRL_OVERRIDES=""
    fi

    # CUSTOM_NGINX_SET_FLAGS cannot be quoted
    # shellcheck disable=SC2086
    helm upgrade ingress-nginx "${K8S_DEPLOY_DIR}/charts/ingress-nginx" \
        --install \
        --create-namespace \
        --namespace="ingress-nginx" \
        -f "${INGRESS_CTRL_OVERRIDES}" \
        ${CUSTOM_NGINX_SET_FLAGS:-} \
        --wait

    echo -e "\n- Waiting for Ingress Controller to be ready...\n"
    kubectl wait --namespace ingress-nginx \
      --for=condition=ready pod \
      --selector=app.kubernetes.io/component=controller \
      --timeout=90s
}

voltron::install_upgrade::argo() {
    # not waiting as other components do not need it during installation
    shout "- Installing Argo Helm chart [wait: false]..."

    helm upgrade argo "${K8S_DEPLOY_DIR}/charts/argo" \
        --install \
        --create-namespace \
        --namespace="argo"
}

voltron::synchronize::minio_secret() {
  echo "Annotating Minio secret to be synchronized across all namespaces..."
  kubectl annotate secret -n argo argo-minio kubed.appscode.com/sync="" --overwrite
}

# Updates /etc/hosts with all Voltron subdomains.
host::update::voltron_hosts() {
  shout "- Updating /etc/hosts..."
  readonly DOMAIN="voltron.local"
  readonly VOLTRON_HOSTS=("gateway")

  LINE_TO_APPEND="127.0.0.1 $(printf "%s.${DOMAIN} " "${VOLTRON_HOSTS[@]}")"
  HOSTS_FILE="/etc/hosts"

  grep -qxF -- "$LINE_TO_APPEND" "${HOSTS_FILE}" || (echo "$LINE_TO_APPEND" | sudo tee -a "${HOSTS_FILE}" > /dev/null)
}

# Sets self-signed wildcard TLS certificate as trusted
#
# Required envs:
#  - REPO_DIR
host::install:trust_self_signed_cert() {
  shout "- Trusting self-signed TLS certificate if not already trusted..."
  CERT_FILE="voltron.local.crt"
  CERT_PATH="${REPO_DIR}/hack/cluster-config/kind/${CERT_FILE}"
  OS="$(host::os)"

  echo "Certificate path: ${CERT_PATH}"
  echo "Detected OS: ${OS}"

  case $OS in
    'linux')
      if diff "${CERT_PATH}" "/usr/local/share/ca-certificates/${CERT_FILE}"; then
        echo "Certificate is already trusted."
        return
      fi

      sudo cp "${CERT_PATH}" "/usr/local/share/ca-certificates"
      sudo update-ca-certificates
      ;;
    'darwin')
      if security verify-cert -c "${CERT_PATH}"; then
        echo "Certificate is already trusted."
        return
      fi

      sudo security add-trusted-cert -d -r trustRoot -k /Library/Keychains/System.keychain "${CERT_PATH}"
      ;;
    *)
      echo "Please manually set the certificate ${CERT_PATH} as trusted for your OS."
      ;;
  esac
}

# Required envs:
#  - DOCKER_REPOSITORY
#  - DOCKER_TAG
#  - REPO_DIR
voltron::update::images_on_kind() {
    pushd "${REPO_DIR}" || return

    shout "- Building Voltron apps and tests images from sources..."
    make build-all-apps-images build-all-tests-images

    REFERENCE_FILTER="$DOCKER_REPOSITORY/*:$DOCKER_TAG"
    shout "- Loading Voltron image into kind cluster... [reference filter: $REFERENCE_FILTER]"
    names=$(docker::list_images "$REFERENCE_FILTER")
    kind::load_images "$names"

    shout "- Deleting local Docker Voltron images..."
    docker::delete_images "$names"

    popd || return
}

# Installs kind and helm dependencies locally.
# Required envs:
#  - MINIMAL_VERSION
#  - CURRENT_VERSION
#
# usage: env MINIMAL_VERSION=v3.3.4 CURRENT_VERSION=v2.16.9 voltron::version_supported
voltron::version_supported(){
  printf '%s\n%s\n' "$CURRENT_VERSION" "$MINIMAL_VERSION" | sort -rVC
}

voltron::validate::tools() {
  shout "- Validating tools versions..."
  local current_kind_version
  local current_helm_version
  local wrong_versions

  current_kind_version=$(kind::version)
  current_helm_version=$(helm::version)
  wrong_versions=false

  echo "Current kind version: $current_kind_version, recommended kind version: $STABLE_KIND_VERSION"
  echo "Current helm version: $current_helm_version, recommended helm version: $STABLE_HELM_VERSION"

  if ! MINIMAL_VERSION="${STABLE_KIND_VERSION}" CURRENT_VERSION="${current_kind_version}" voltron::version_supported; then
    wrong_versions=true
    echo "Unsupported kind version $current_kind_version. Must be at least $STABLE_KIND_VERSION"
  fi
  if ! MINIMAL_VERSION="${STABLE_HELM_VERSION}" CURRENT_VERSION="${current_helm_version}" voltron::version_supported; then
      wrong_versions=true
      echo "Unsupported helm version $current_helm_version. Must be at least $STABLE_HELM_VERSION"
  fi
  [ ${wrong_versions} == false ]
}
