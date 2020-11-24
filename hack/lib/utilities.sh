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

dump_logs() {
    LOGS_DIR=${ARTIFACTS:-./tmp}/logs
    mkdir -p "${LOGS_DIR}"

    echo "Dumping logs from namespace ${DUMP_NAMESPACE} into ${LOGS_DIR}"
    kubectl cluster-info dump --namespace="${DUMP_NAMESPACE}" --output-directory="${LOGS_DIR}"
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
      --config "${REPO_DIR}/hack/kind/config.yaml" \
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

# Required envs:
#  - DOCKER_REPOSITORY
#  - DOCKER_TAG
#  - REPO_DIR
#  - VOLTRON_NAMESPACE
#  - VOLTRON_RELEASE_NAME
#  - CLUSTER_TYPE
#
# Optional envs:
#  - UPDATE - if specified then, Helm charts are updated
voltron::install::charts() {
    readonly K8S_DEPLOY_DIR="${REPO_DIR}/deploy/kubernetes"

    shout "- Applying Voltron CRDs..."
    kubectl apply -f "${K8S_DEPLOY_DIR}"/crds

    voltron::install::ingress_controller

    voltron::install::argo

    if [[ "${DISABLE_MONITORING_INSTALLATION:-"false"}" == "true" ]]; then
      shout "Skipping monitoring installation cause DISABLE_MONITORING_INSTALLATION is set to true."
    else
      voltron::install::monitoring
    fi

    shout "- Installing Voltron Helm chart from sources [wait: true]..."
    echo -e "- Using DOCKER_REPOSITORY=$DOCKER_REPOSITORY and DOCKER_TAG=$DOCKER_TAG\n"

    helm "$(voltron::install::detect_command)" "${VOLTRON_RELEASE_NAME}" "${K8S_DEPLOY_DIR}/charts/voltron" \
        --create-namespace \
        --namespace="${VOLTRON_NAMESPACE}" \
        --set global.containerRegistry.path="$DOCKER_REPOSITORY" \
        --set global.containerRegistry.overrideTag="$DOCKER_TAG" \
        --wait
}

voltron::install::monitoring() {
    # not waiting as Helm Charts installation takes additional ~3 minutes. To proceed further we need only monitoring CRDs.
    shout "- Installing monitoring Helm chart [wait: false]..."
    helm "$(voltron::install::detect_command)" monitoring "${K8S_DEPLOY_DIR}/charts/monitoring" \
        --create-namespace \
        --namespace="monitoring"
}

voltron::install::ingress_controller() {
    # waiting as admission webhooks server is required to be available during further installation steps
    shout "- Installing Ingress NGINX Controller Helm chart [wait: true]..."

    if [[ "${CLUSTER_TYPE}" == "KIND" ]]; then
      readonly INGRESS_CTRL_OVERRIDES="${REPO_DIR}/hack/kind/overrides.ingress-nginx.yaml"
      echo -e "- Applying overrides from ${INGRESS_CTRL_OVERRIDES}\n"
    else # currently, only KIND needs custom settings
      readonly INGRESS_CTRL_OVERRIDES=""
    fi

    helm "$(voltron::install::detect_command)" ingress-nginx "${K8S_DEPLOY_DIR}/charts/ingress-nginx" \
        --create-namespace \
        --namespace="ingress-nginx" \
        -f "${INGRESS_CTRL_OVERRIDES}" \
        --wait

    echo -e "\n- Waiting for Ingress Controller to be ready...\n"
    kubectl wait --namespace ingress-nginx \
      --for=condition=ready pod \
      --selector=app.kubernetes.io/component=controller \
      --timeout=90s
}

voltron::install::argo() {
    # not waiting as other components do not need it during installation
    shout "- Installing Argo Helm chart [wait: false]..."

    if [[ "${CLUSTER_TYPE}" == "KIND" ]]; then
      readonly ARGO_OVERRIDES="${REPO_DIR}/hack/kind/overrides.argo.yaml"
      echo -e "- Applying overrides from ${ARGO_OVERRIDES}\n"
    else # currently, only KIND needs custom settings
      readonly ARGO_OVERRIDES=""
    fi

    helm "$(voltron::install::detect_command)" argo "${K8S_DEPLOY_DIR}/charts/argo" \
        --create-namespace \
        --namespace="argo" \
        -f "${ARGO_OVERRIDES}"
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
  CERT_PATH="${REPO_DIR}/hack/kind/${CERT_FILE}"
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

voltron::install::detect_command() {
  if [[ "${UPDATE:-x}" == "true" ]]; then
    echo "upgrade"
    return
  fi
  echo "install"
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
