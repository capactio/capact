#
# Library of useful utilities for CI purposes.
#
readonly RED='\033[0;31m'
readonly GREEN='\033[0;32m'
readonly INVERTED='\033[7m'
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
    mkdir -p ${LOGS_DIR}

    echo "Dumping logs from namespace ${DUMP_NAMESPACE} into ${LOGS_DIR}"
    kubectl cluster-info dump --namespace=${DUMP_NAMESPACE} --output-directory=${LOGS_DIR}
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
  tar -zxvf ${name}.tar.gz
  mv ${name} kubebuilder

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
    curl -fsSL -o ${INSTALL_DIR}/get_helm.sh https://raw.githubusercontent.com/helm/helm/master/scripts/get-helm-3
    chmod 700 ${INSTALL_DIR}/get_helm.sh
    env HELM_INSTALL_DIR="${INSTALL_DIR}/bin" ./get_helm.sh \
        --version ${HELM_VERSION} \
        --no-sudo

    popd || return
}

#
# 'kind'(kubernetes-in-docker) functions
#
# Required environments variables for all 'kind' commands:
# - KIND_CLUSTER_NAME

kind::create_cluster() {
    shout "- Create k8s cluster..."
    kind create cluster --name=${KIND_CLUSTER_NAME} --image="kindest/node:${KUBERNETES_VERSION}" --wait=5m
}

kind::delete_cluster() {
    kind delete cluster --name=${KIND_CLUSTER_NAME}
}

# Arguments:
#   $1 - list of docker images in format "name:tag [name:tag...]"
kind::load_images() {
  for id in $1; do
    kind load docker-image $id --name=${KIND_CLUSTER_NAME}
  done
}


#
# Docker functions
#

# Arguments:
#   $1 - reference filter
docker::list_images() {
  names=$(docker image ls --filter=reference="$1" --format="{{.Repository}}:{{.Tag}}")

  echo $names
}

# Arguments:
#   $1 - image list
docker::delete_images() {
  docker rmi $1
}

#
# Voltron functions
#

# Required envs:
#  - DOCKER_PUSH_REPOSITORY
#  - DOCKER_TAG
#  - REPO_DIR
#  - VOLTRON_NAMESPACE
#  - VOLTRON_RELEASE_NAME
#
# Optional envs:
#  - UPDATE - if specified then, Helm charts are updated
voltron::install::from_sources() {
    readonly K8S_DEPLOY_DIR="${REPO_DIR}/deploy/kubernetes"

    pushd "${REPO_DIR}" || return

    shout "- Building Voltron image from sources..."
    make build-all-images

    REFERENCE_FILTER="$DOCKER_PUSH_REPOSITORY/*:$DOCKER_TAG"
    shout "- Loading Voltron image into kind cluster... [reference filter: $REFERENCE_FILTER]"
    names=$(docker::list_images "$REFERENCE_FILTER")
    kind::load_images "$names"

    shout "- Deleting local Docker Voltron images..."
    docker::delete_images "$names"

    shout "- Applying Voltron CRDs..."
    kubectl apply -f "${K8S_DEPLOY_DIR}"/crds

    shout "- Installing Voltron Helm chart from sources..."
    helm "$(voltron::install::detect_command)" "${VOLTRON_RELEASE_NAME}" "${K8S_DEPLOY_DIR}"/chart \
        --create-namespace \
        --namespace=${VOLTRON_NAMESPACE} \
        --set global.containerRegistry.path=$DOCKER_PUSH_REPOSITORY \
        --set global.containerRegistry.overrideTag=$DOCKER_TAG \
        --wait

    popd || return
}

voltron::install::detect_command() {
  if [[ "${UPDATE:-x}" == "true" ]]; then
    echo "upgrade"
    return
  fi
  echo "install"
}
