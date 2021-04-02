#!/bin/bash
set -eEu

# Required envs:
# CAPACT_NAME - Prefix added to every resource name
# CAPACT_REGION - AWS region, where to deploy the infrastructure
# CAPACT_DOMAIN_NAME - Domain name, under which Capact will be available
capact::aws::terraform::apply() {
  local state_bucket
  local state_key
  local state_region

  state_bucket="${TERRAFORM_STATE_BUCKET}"
  state_key="${CAPACT_NAME}.tfstate"
  state_region="${CAPACT_REGION}"

  pushd "${CURRENT_DIR}/terraform"
    terraform init \
      -backend-config="bucket=${state_bucket}" \
      -backend-config="key=${state_key}" \
      -backend-config="region=${state_region}"

    terraform apply -auto-approve -no-color \
      -var "namespace=${CAPACT_NAME}" \
      -var "region=${CAPACT_REGION}" \
      -var "domain_name=${CAPACT_DOMAIN_NAME}"

    readonly tf_output=$(terraform output -json)

    echo "${tf_output}" | jq -r '.bastion_public_ip.value' > "${CONFIG_DIR}/bastion_public_ip"
    echo "${tf_output}" | jq -r '.bastion_ssh_private_key.value' > "${CONFIG_DIR}/bastion_ssh_private_key"
    chmod 400 "$CONFIG_DIR/bastion_ssh_private_key"

    echo "${tf_output}" | jq -r '.eks_kubeconfig.value' > "$CONFIG_DIR/eks_kubeconfig"
    chmod 600 "$CONFIG_DIR/eks_kubeconfig"

    echo "${tf_output}" | jq -r '.route53_zone_name_servers.value' > "$CONFIG_DIR/route53_zone_name_servers"
    echo "${tf_output}" | jq -r '.route53_zone_id.value' > "$CONFIG_DIR/route53_zone_id"
    echo "${tf_output}" | jq -r '.cert_manager_irsa_role_arn.value' > "${CONFIG_DIR}/cert_manager_role_arn"
  popd
}

capact::aws::install::fluent_bit() {
  "${CURRENT_DIR}"/aws-for-fluent-bit/install.sh
}

capact::aws::install::capact() {
  "${CURRENT_DIR}"/cluster-components-install-upgrade.sh
}

capact::aws::install::cert_manager() {
  "${CURRENT_DIR}"/cert-manager/install.sh
}

capact::aws::install::public_ingress_controller() {
  "${CURRENT_DIR}"/public-ingress-controller/install.sh
}

capact::aws::register_dnses() {
  echo "- Adding DNS entries to Route53..."
  local internal_lb_fqdn
  for _ in $(seq 6); do
    internal_lb_fqdn="$(kubectl -n ingress-nginx get svc ingress-nginx-controller '-ojsonpath={.status.loadBalancer.ingress[].hostname}')"
    if [ -n "${internal_lb_fqdn}" ]; then
      break
    fi
    sleep 10
  done
  if [ -z "${internal_lb_fqdn}" ]; then
    echo "Timout waiting for internal load balancer to be provisioned."
    exit 1
  fi

  local external_lb_fqdn
  for _ in $(seq 6); do
    external_lb_fqdn="$(kubectl -n public-ingress-nginx get svc public-ingress-nginx-controller '-ojsonpath={.status.loadBalancer.ingress[].hostname}')"
    if [ -n "${external_lb_fqdn}" ]; then
      break
    fi
    sleep 10
  done
  if [ -z "${external_lb_fqdn}" ]; then
    echo "Timout waiting for public load balancer to be provisioned."
    exit 1
  fi

  local internal_lb_hosted_zone
  internal_lb_hosted_zone="$(aws elb describe-load-balancers \
    | jq -r ".LoadBalancerDescriptions[] \
      | select(.DNSName == \"${internal_lb_fqdn}\") \
      | .CanonicalHostedZoneNameID")"

  local external_lb_hosted_zone
  external_lb_hosted_zone="$(aws elb describe-load-balancers \
    | jq -r ".LoadBalancerDescriptions[] \
      | select(.DNSName == \"${external_lb_fqdn}\") \
      | .CanonicalHostedZoneNameID")"

  local changes
  changes="
{
  \"Changes\":[{
    \"Action\": \"UPSERT\",
    \"ResourceRecordSet\": {
      \"Name\":\"gateway.${CAPACT_DOMAIN_NAME}\",
      \"Type\": \"A\",
      \"AliasTarget\": {
        \"DNSName\":\"${internal_lb_fqdn}\",
        \"HostedZoneId\":\"${internal_lb_hosted_zone}\",
        \"EvaluateTargetHealth\": false
      }
    }
  },
  {
    \"Action\": \"UPSERT\",
    \"ResourceRecordSet\": {
      \"Name\":\"*.${CAPACT_DOMAIN_NAME}\",
      \"Type\": \"A\",
      \"AliasTarget\": {
        \"DNSName\":\"${external_lb_fqdn}\",
        \"HostedZoneId\":\"${external_lb_hosted_zone}\",
        \"EvaluateTargetHealth\": false
      }
    }
  }]
}
"

  aws route53 change-resource-record-sets \
    --hosted-zone-id "${CAPACT_HOSTED_ZONE_ID}" \
    --change-batch "${changes}"
  echo "- Added DNS entries added"
}

capact::aws::configure_bastion() {
  # upload kubeconfig
  ssh -i config/bastion_ssh_private_key ec2-user@"$(cat config/bastion_public_ip)" 'mkdir -p $HOME/.kube'
  scp -i config/bastion_ssh_private_key config/eks_kubeconfig ec2-user@"$(cat config/bastion_public_ip)":.kube/config
}

readonly CURRENT_DIR=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )
readonly CONFIG_DIR="${CURRENT_DIR}/config"
rm -rf "${CONFIG_DIR}"
mkdir -p "${CONFIG_DIR}"

export CAPACT_NAME="${CAPACT_NAME}"
export CAPACT_REGION="${CAPACT_REGION}"
export CAPACT_DOMAIN_NAME="${CAPACT_DOMAIN_NAME}"
export DOCKER_TAG="${CAPACT_DOCKER_TAG}"
export DOCKER_REPOSITORY="${CAPACT_DOCKER_REPOSITORY:-gcr.io/projectvoltron}"

capact::aws::terraform::apply

export KUBECONFIG="${CONFIG_DIR}/eks_kubeconfig"

CAPACT_HOSTED_ZONE_ID=$(cat "${CONFIG_DIR}/route53_zone_id")
CERT_MANAGER_ROLE_ARN=$(cat "${CONFIG_DIR}/cert_manager_role_arn")
CUSTOM_VOLTRON_SET_FLAGS="--set global.domainName=${CAPACT_DOMAIN_NAME}"

export CAPACT_HOSTED_ZONE_ID
export CERT_MANAGER_ROLE_ARN
export CUSTOM_VOLTRON_SET_FLAGS

capact::aws::install::fluent_bit
capact::aws::install::capact
capact::aws::install::public_ingress_controller
capact::aws::install::cert_manager
capact::aws::register_dnses
capact::aws::configure_bastion

echo -e "\n -------------------- Installation completed! -------------------"
echo -e "Kubeconfig and SSH key to the bastion are available in ${CONFIG_DIR} directory"
