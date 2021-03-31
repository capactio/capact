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

  state_bucket="${TERRAFORM_STATE_BUCKET:-capact-terraform-states}" # TODO change this name
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
      -var "domain_name"=${CAPACT_DOMAIN_NAME}
  popd
}

readonly CURRENT_DIR=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )

export CAPACT_NAME="${CAPACT_NAME}"
export CAPACT_REGION="${CAPACT_REGION}"
export CAPACT_DOMAIN_NAME="${CAPACT_DOMAIN_NAME}"

capact::aws::terraform::apply
