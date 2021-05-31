# Introduction

This tutorial showcases how to install a set of productivity applications on AWS EKS cluster with single AWS RDS instance.

The productivity stack consists of the following applications:
- Atlassian Jira
- Atlassian Crowd
- Atlassian Confluence
- Atlassian Bitbucket
- Rocket.chat

## Diagram

![productivity-stack-diagram](./assets/productivity-stack-diagram.svg)

## Prerequisites

- AWS account with **AdministratorAccess** permissions
- Capact cluster provisioned using [EKS Installation tutorial](../../installation/aws-eks.md)

## Steps

1. [Configure Cluster Policy to prefer AWS solutions](./1-cluster-policy-configuration.md)
1. [Provision AWS RDS for PostgreSQL](./2-aws-rds-provisioning.md)
1. [Install Crowd](./3-crowd-installation.md)
1. [Install Bitbucket](./4-bitbucket-installation.md)
1. [Install Jira](./5-jira-installation.md)
1. [Install Confluence](./6-confluence-installation.md)
1. [Install RocketChat](./7-rocket-chat-installation.md)
