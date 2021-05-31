#  Capact CI and CD

> ⚠️ **DEPRECATION:** The [gcr.io/projectvoltron](https://gcr.io/projectvoltron) registry is deprecated. We support this registry in the read-only mode until November 2021. At that time this registry will likely be garbage collected and no longer available. New Docker images are pushed to [ghcr.io/capactio](https://github.com/orgs/capactio/packages?ecosystem=container).

This document describes jobs created to automate the process of testing, building, and deploying newly merged functionality.

##  Table of Contents

<!-- toc -->

- [Overview](#overview)
- [Repository secrets](#repository-secrets)
- [Pipelines](#pipelines)
  * [Pull request](#pull-request)
  * [Main branch](#main-branch)
  * [Recreate a long-running cluster](#recreate-a-long-running-cluster)
    + [Let's encrypt certificates](#lets-encrypt-certificates)
  * [Execute integration tests on a long-running cluster](#execute-integration-tests-on-a-long-running-cluster)
- [Accessing encrypted files on CI](#accessing-encrypted-files-on-ci)
- [Add a new pipeline](#add-a-new-pipeline)

<!-- tocstop -->

##  Overview

For all our CI/CD jobs, we use [GitHub Actions](https://docs.github.com/en/free-pro-team@latest/actions). Our workflows are defined in the [`.github/workflows`](https://github.com/capactio/capact/tree/main/.github/workflows) directory. All scripts used for the CI/CD purpose are defined in the [`/hack/ci/`](https://github.com/capactio/capact/tree/main/hack/ci) directory. For example, the [`/hack/ci/setup-env.sh`](https://github.com/capactio/capact/tree/main/hack/ci/setup-env.sh) file has defined all environment variables used for every pipeline job.

##  Repository secrets

All sensitive data is stored in [GitHub secrets](https://docs.github.com/en/free-pro-team@latest/actions/reference/encrypted-secrets). As a result, we can access them in each workflow executed on our pipeline.

The following secrets are defined:

| Secret name                       | Description                                                                                                                                                                                                                                        |
|-----------------------------------|----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| **CAPACT_GCS_CREDS**              | Holds credentials which allow CI jobs to manage GCS bucket. Has the `roles/storage.objectAdmin` role.                                                                                                                                              |
| **CAPACT_GKE_CREDS**              | Holds credentials which allow CI jobs to create and manage the GKE private cluster. Has the `Compute Admin`, `Compute Network Admin`, `Kubernetes Engine Admin`, `Kubernetes Engine Cluster Admin`, `Service Account User`, `Storage Admin` roles. |
| **STAGE_CAPACT_GATEWAY_PASSWORD** | Holds the Gateway password for the long-running cluster.                                                                                                                                                                                           |
| **GIT_CRYPT_KEY**                 | Holds a symmetric key used to decrypt files encrypted with [git crypt](https://github.com/AGWA/git-crypt).                                                                                                                                         |

##  Pipelines

###  Pull request

<p align="center"><img alt="ci-pr-build" src="./assets/ci-pr-build.svg" /></p>

The job is defined in the [`pr-build.yaml`](https://github.com/capactio/capact/tree/main/.github/workflows/pr-build.yaml) file. It runs on pull requests created to the `main` branch.

Steps:

1. Lint and test submitted code.
1. Check documentation if the `*.md` files were modified. 
1. Run integration tests.
1. Build Docker images for applications, tests and infra tools, and push them to [ghcr.io/capactio](https://github.com/orgs/capactio/packages?ecosystem=container) using this pattern: `ghcr.io/capactio/pr/{service_name}:PR-{pr_number}`.

### Main branch

> **NOTE:** To reduce the CI build time, we disable the `entry-tests`, `build-tools` and `integration-tests` jobs. They will be enabled when the project is open-sourced.

<p align="center"><img alt="ci-default-branch-build" src="./assets/ci-default-branch-build.svg" /></p>

The job is defined in the [`.github/workflows/branch-build.yaml`](https://github.com/capactio/capact/tree/main/.github/workflows/branch-build.yaml) file. It runs on every new commit pushed to the `main` branch but skips execution for files which do not affect the building process, e.g. documentation.

Steps:

1. Lint and test code.
1. Build Docker images for applications, tests and infra tools, and push them to [ghcr.io/capactio](https://github.com/orgs/capactio/packages?ecosystem=container) using this pattern: `ghcr.io/capactio/{service_name}:{first_7_chars_of_commit_sha}`.
1. If [Capact Helm Charts](https://github.com/capactio/capact/tree/main/deploy/kubernetes/charts) were changed:
   1. Change **version** in all `Chart.yaml` to `{current_version}-{first_7_chars_of_commit_sha}`.
   1. Package and push charts to the [`capactio-latest-charts`](https://storage.googleapis.com/capactio-latest-charts) GCS.   
2. Update the existing long-running cluster via [CLI](../cli/commands/capact_upgrade.md).
1. Delete all Actions which are in the `SUCCEEDED` phase and whose names have the `capact-upgrade-` prefix. 
1. If any step failed, send a Slack notification.

###  Recreate a long-running cluster

<p align="center"><img alt="ci-recreate-cluster" src="./assets/ci-recreate-cluster.svg" /></p>

The job is defined in the [`.github/workflows/recreate_cluster.yaml`](https://github.com/capactio/capact/tree/main/.github/workflows/recreate_cluster.yaml) file. It is executed on a manual trigger using the [`workflow_dispatch`](https://github.blog/changelog/2020-07-06-github-actions-manual-triggers-with-workflow_dispatch/) event. It uses already existing images available in the [ghcr.io/capactio](https://github.com/orgs/capactio/packages?ecosystem=container) registry. As a result, you need to provide a git SHA from which the cluster should be recreated. Optionally, you can override the Docker image version used via the **DOCKER_TAG** parameter.

> **CAUTION:** This job removes the old GKE cluster.

####  Let's encrypt certificates

The `recreate` job checks if the certificate exists in the GCS bucket. If it does, it downloads it and checks if the certificate is still valid. If it's valid, it copies it to a long-running cluster, otherwise the job creates the Let's Encrypt certificates using Cert Manager and backs it up to the dedicated GCS bucket. By doing so, we ensure that we do not hit the quotas defined on the Let's Encrypt side.

###  Execute integration tests on a long-running cluster

<p align="center"><img alt="ci-integration-tests" src="./assets/ci-integration-tests.svg" /></p>

The job is defined in the [`.github/workflows/cluster_integration_tests.yaml`](https://github.com/capactio/capact/tree/main/.github/workflows/cluster_integration_tests.yaml) file. It runs periodically according to cron defined in the job definition. It executes integration tests using the `helm test` command.

##  Accessing encrypted files on CI

The sensitive data that needs to be accessed on a pipeline, such as overrides for passwords, certificates etc., must be stored in the [`hack/ci/sensitive-data`](https://github.com/capactio/capact/tree/main/hack/ci/sensitive-data) directory. Files in that folder are encrypted using [git crypt](https://github.com/AGWA/git-crypt), which you should install and configure on your local machine. Currently, it works for `*.txt` files put in this directory, but this can be changed in the `.gitattributes` file.

The demo setup is as follows:

```bash
*.txt filter=git-crypt diff=git-crypt
.gitattributes !filter !diff
```

It means that every `*.txt` file in the `hack/ci/sensitive-data` directory is encrypted before being push to a git repository. If you need to encrypt other files in a different directory, you have to create there a `.gittatributes` file with proper rules. Do not forget to add the `.gitattributes !filter !diff` statement as it prevents encryption for the `.gitattributes` file.

To decrypt the data locally, you must either use a symmetric key or add the GPG key. The procedure of decrypting files and working with it in a team is described [here](https://buddy.works/guides/git-crypt#working-in-team-with-git-crypt).

Currently, [`decrypt.yaml`](https://github.com/capactio/capact/tree/main/.github/workflows/decrypt.yaml) shows how to decrypt a file on CI.

##  Add a new pipeline

To create a new pipeline you must follow the rules of the syntax used by GitHub Actions. The new workflow must be defined in the [`.github/workflows`](https://github.com/capactio/capact/tree/main/.github/workflows) directory. All scripts for CI/CD purposes must be defined in the [`/hack/ci/`](https://github.com/capactio/capact/tree/main/hack/ci) directory.

The following steps show how to checkout the code, set up the Go environment, and authorize to GHCR and GKE in case they are necessary.

```yaml
    steps:    
      - name: Checkout code
        uses: actions/checkout@v2

      - name: Authorize to GHCR
        run: echo "${{ secrets.GITHUB_TOKEN }}" | docker login ghcr.io -u ${{ github.actor }} --password-stdin

      - name: Authorize to GKE
        uses: GoogleCloudPlatform/github-actions/setup-gcloud@master
        with:
          service_account_key: ${{ secrets.CAPACT_GKE_CREDS }}
          export_default_credentials: true

      - name: Setup env
        run: |
          . ./hack/ci/setup-env.sh

      - name: Setup Go
        uses: actions/setup-go@v2
        with:
          go-version: ${{env.GO_VERSION}}
```
