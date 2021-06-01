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

1. Kubevela uses OAM, declarative API, to describe application, its configuration and dependencies. Capact is using workflow-based approach. We believe that it makes Capact more flexible. Especially for day 2 operations.
For example, you may create an advanced workflow for taking a backup. Where you first pause an application, create volume snapshot, create DB snapshot, copy snapshot to s3 bucket and rotate the previous backups.

1. Capact is using [Policies](./policy-configuration.md) to select implementation when deploying and managing application. For example, you may use SQL database in your application. For local development, you may choose to use PostgreSQL installed by Helm, but for production you will be using RDS. This again gives you more flexibility.

### Helm

Helm is a package manager for Kubernets. Capact is using [Helm runner](../cmd/helm-runner/README.md) to install applications in Kubernets. Capact goes beyond Kubernetes and can deploy and manage diverse workloads like AWS RDS or EC2 instances. In some way Capact extends a Helm.

For example, if you are deploying an application which is using a database you may use RDS PostgreSQL and pass required values to the Helm chart. We are doing it in many of our examples. See our [Productivity stack installation](./tutorial/productivity-stack-installation/README.md) tutorial for more details.

### Operator Framework

Operator Framework is a toolkit to manage Kubernetes application. It makes it easy to do day1 and day2 operations. Capact being a glue connecting different tools could use Operators to manage Kubernets applications. For now, we are using only Helm, but in future we could also support Operator Framework.

### Terraform, Ansible, Chef, etc.

All this tools, of course, differ a lot, but in general, all are used to describe and enforce the desired state of the environment.
Capact builds on top of them, these tools can be used as a part of a workflow and even can be mixed. We already have [Terraform](../cmd/terraform-runner/README.md) runner. Ansible and other runners are also possible.

For example, when running Capact manifests using Terraform, you can deploy AWS RDS and EKS. Then, using Helm, you can deploy Mattermost in your Kubernetes cluster. All this can be done in one manifest. See our [Mattermost example](./tutorial/mattermost-installation/README.md) for more details.


### Pulumi

Pulumi is similar to Terraform but instead of having its configuration language you can use languages like Python, Go, JavaScript and others.
This mean that we could have a Pulumi runner in Capact and use it in our Manifests.

## Where did the name Capact come from?

Capact comes from the word capabilities. We wanted to use a noun which shows that this tool has infinite possibilities ;)
