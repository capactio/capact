# Frequently asked questions

<!-- toc -->

- [How does it compare to](#how-does-it-compare-to)
  * [Kubevela with Crossplane](#kubevela-with-crossplane)
  * [Helm](#helm)
  * [Operator Framework](#operator-framework)
  * [Terraform, Ansible, Chef, etc.](#terraform-ansible-chef-etc)
  * [Pulumi](#pulumi)
- [Where did the name Capact come from](#where-did-the-name-capact-come-from)

<!-- tocstop -->

## How do I start using Capact?

Have a look at our [tutorials](./tutorial/README.md). You can learn there how to deploy Capact on [AWS](./tutorial/capact-installation/aws-eks.md) and [GCP](./tutorial/capact-installation/gcp-gke.md). You can learn [How to use it](./tutorial/mattermost-installation/README.md) and how to create [a new content](./tutorial/content-creation/README.md) for Capact.

## How does it compare to

### Kubevela with Crossplane

Kubevela is a tool to describe and ship applications. Kubevela's core is built on top of Crossplane and it uses Open Application Model (OAM). 
There are many similarities between Capact and Kubevela. Both can deploy and manage diverse workload types such as container, databases, EC2 instances across hybrid environments. Both solutions work as glue which can connect different tools like Terraform and Helm.

There are two main differences between them:

1. Kubevela uses declarative API to describe application, its configuration and dependencies. Capact is using workflow-based approach. We believe that it makes Capact more flexible, especially for day-2 operations.
    For example, with Capact you can create an advanced workflow for doing a backup. In the workflow first you pause an application, create volume snapshot, create DB snapshot, copy snapshot to S3 bucket and rotate the previous backups.

1. Capact has interchangeable dependencies as a built-in feature. Dependencies are described using Interfaces. You can configure Implementation preferences for any Interface  with [Policies](./policy-configuration.md) to select Implementation based on a Interface while managing applications. For example, if your application depends on SQL database, for local development, you can prefer to use in-cluster PostgreSQL installed by Helm, but for production environment you prefer managed solution such as AWS RDS.

### Helm

Helm is a package manager for Kubernetes. Capact uses [Helm runner](https://github.com/capactio/capact/tree/main/cmd/helm-runner/README.md) to install applications in Kubernetes. Capact goes beyond Kubernetes and can deploy and manage diverse workloads like AWS RDS or EC2 instances. In a way, Capact extends Helm.

For example, if you are deploying an application which is using a database you may use RDS PostgreSQL and pass required values to the Helm chart. We are doing it in many of our examples. See our [Productivity stack installation](./tutorial/productivity-stack-installation/README.md) tutorial for more details.

### Operator Framework

Operator Framework is a toolkit to manage Kubernetes application. It makes it easy to do day-1 and day-2 operations. Capact is a glue connecting different tools and it could use Operators to manage Kubernetes applications. For now, we have a runner that supports Helm from the Kubernetes ecosystem, but in future we could also support Operator Framework.

### Terraform, Ansible, Chef, etc.

In general, all these tools are used to describe and enforce the desired state of the environment. As Capact is a layer above of the tools, they can be used as a part of a Capact workflow and even can be mixed. We already have [Terraform runner](https://github.com/capactio/capact/tree/main/cmd/terraform-runner/README.md). Ansible and other runners are also possible.

For example, when running Capact manifests, you can deploy AWS RDS and EKS defined as Terraform modules. Then, using Helm runner, you can deploy Mattermost in your Kubernetes cluster. All this can be done in one OCF manifest. See our [Mattermost example](./tutorial/mattermost-installation/README.md) for more details.

### Pulumi

Capact supports Terraform runner to use Terraform modules to manage your infrastucture. Pulumi is similar to Terraform, however, Instead of having its configuration language, you can use programming languages like Python, Go, JavaScript and others. There could be a Pulumi runner for Capact, which allows you to use Pulimi content in OCF manifests.

## What is the origin of name "Capact"?

Capact is a combination of two shortened words: "**cap**ability" and "**act**ion". Capact makes it easy to manage system capabilities via running Actions. Once you start using Capact, the possibilities are virtually endless.