# Capact release process

This document describes Capact release process. Currently, it consists of a set of manual steps, however in future it will be automated.

## Table of contents

<!-- toc -->

- [Prerequisites](#prerequisites)
- [Steps](#steps)
  * [Export environmental variables](#export-environmental-variables)
  * [Create a pre-release pull request](#create-a-pre-release-pull-request)
  * [Create release branch](#create-release-branch)
    + [Release Helm charts and binaries](#release-helm-charts-and-binaries)
  * [Create new git tag and GitHub release](#create-new-git-tag-and-github-release)
- [Create GitHub release](#create-github-release)

<!-- tocstop -->

## Prerequisites

- [yq](https://github.com/mikefarah/yq) v3

## Steps

### Export environmental variables

Export environmental variable with your version:
    
Use [Semantic Versioning 2.0.0](https://semver.org/spec/v2.0.0.html) to specify the next Capact release.

```bash
export RELEASE_VERSION={major}.{minor}.{patch} 
export RELEASE_BRANCH=release-{major}.{minor} 
```

For example, in case of the `0.3.0` release, it would be:

```bash
export RELEASE_VERSION=0.3.0 
export RELEASE_BRANCH=release-0.3
```

### Create a pre-release pull request

1. Checkout the destination branch for the pull request.

    - For major and minor release versions, set the destination branch to `master`. 
    - For patch releases, set the destination to corresponding release branch. For example, for `0.3.1` release, checkout the `release-0.3` branch.

    ```bash
    git checkout {destination-branch}
    ```

1. Create and checkout new branch:
    
   ```bash
   git checkout -b prepare-${RELEASE_VERSION}
   ```   

1. Modify `.github/workflows/branch-build.yaml` and append new branch to `branches`:

    ```bash
    yq --style double w -i .github/workflows/branch-build.yaml 'on.push.branches[+]' "${RELEASE_BRANCH}"
    ```

1. Change versions of all Helm charts:

   ```bash
   DEPLOY_DIR=deploy/kubernetes/charts
   for d in ${DEPLOY_DIR}/*/ ; do
     sed -i.bak "s/^version: .*/version: ${RELEASE_VERSION}/g" "${d}/Chart.yaml"
   done
   ```

1. Change CLI version:

    ```bash
   sed -i.bak "s/Version = .*/Version = \"${RELEASE_VERSION}\"/g" "internal/cli/info.go"
   ```
   
1. Commit the changes and push the branch to origin.
    
    ```bash
    git add .
    git commit -m "Prepare ${RELEASE_VERSION} release"
    git push -u origin prepare-${RELEASE_VERSION}
    ```
    
1. Create the pull request from the branch.
   
   - In the pull request description, write the GitHub release notes that will be posted with the release to review.
   - As the pull request target branch, pick the proper destination branch from the first step of this section.
    
1. Merge the pull request.
    
### Create release branch

If you release major or minor version, create a dedicated release branch.

1. Checkout the destination branch and pull the latest changes:

    ```bash
    git checkout {destination_branch}
    git pull
    ```

1. Create new branch:
   
    ```bash
    git checkout -b ${RELEASE_BRANCH}
    ```

#### Release Helm charts and binaries

1. Get the latest commit short hash on the destination branch:
    
   ```bash
     export CAPACT_IMAGE_TAG=$(git rev-parse --short HEAD | sed 's/.$//')
   ```  

   > **NOTE:** It will be used as a Docker image tag for the release Helm charts. Make sure all the component images with this tag have been built on CI.  

1. Replace default tag for Capact chart:

    ```bash
    sed -i.bak "s/overrideTag: \"latest\"/overrideTag: \"${CAPACT_IMAGE_TAG}\"/g" "deploy/kubernetes/charts/capact/values.yaml"
    ```

1. Replace Populator target branch from `master` to the release branch:
  
   ```bash
   sed -i.bak "s/branch: master/branch: ${RELEASE_BRANCH}/g" "deploy/kubernetes/charts/capact/charts/och-public/values.yaml"
   ```

1. Review and commit the changes:

   ```bash
   git add .
   git commit -m "Set fixed Capact image tag and Populator source branch"
   ```

1. Release Helm charts:
   
    ```bash
    ./hack/release-charts.sh
    ```

1. Release tools binaries:
   
    ```bash
    make build-all-tools-prod
    export CAPACT_BINARIES_BUCKET=capactio_binaries
    gsutil -m cp ./bin/* gs://${CAPACT_BINARIES_BUCKET}/v0.0.0/
    ```

### Create new git tag and GitHub release

1. Create new tag and push it and release branch to origin:

    > **NOTE**: Git tag is in a form of SemVer version with `v` prefix, such as `v0.3.0`.
   
    ```bash
    git tag v${RELEASE_VERSION} HEAD
    ```
   
1. Review the changes you've made and push the release branch with tag to origin:

   ```bash
   git push -u origin ${RELEASE_BRANCH}
   git push origin v${RELEASE_VERSION}
   ```

## Create GitHub release
    
1. Navigate to [New GitHub release](https://github.com/Project-Voltron/go-voltron/releases/new) page.
1. Copy the release notes from the pull request created in first step.
1. Create new GitHub release with the copied notes.

