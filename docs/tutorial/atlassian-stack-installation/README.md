# Atlassian stack installation

This tutorial showcases how to install Atlassian stack on AWS EKS cluster with single AWS RDS instance.

![Capact_Atlassian](./assets/Capact_Atlassian.svg)

## Prerequisites

- AWS account with **AdministratorAccess** permissions
- Capact cluster provisioned using [EKS Installation tutorial](../eks-installation/README.md)

## Steps

1. [Configure Cluster Policy to prefer AWS solutions](./1-cluster-policy-configuration.md)
1. [Provision AWS RDS for PostgreSQL](./2-aws-rds-provisioning.md)
1. [Install Crowd](./3-crowd-installation.md)
1. [Provision Bitbucket](./4-bitbucket-installation.md)
1. [Provision Jira](./5-jira-provisioning.md)
1. [Install Confluence](./6-confluence-installation.md)
1. [Install RocketChat](./7-rocket-chat-installation.md)
