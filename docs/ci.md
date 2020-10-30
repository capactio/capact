# CI/CD jobs documentation
This document elaborates on jobs created to automate the process of development/building  new releases.
It touches the matter of encrypting files in repo, which was a part of the task

### Entry information
In main foldera file called env_setup.sh exists. This setups the environment for every job.

## PR pipeline
The job is defined in file pr-build.yaml. This is executed on pull request while it is: opened, synchronize or reopened. The branch related is master.
It is executed while *.go or *.graphql file is commited.

It builds the services images and pushes them to appropriate repository and location i.e.
gcr.io/projectvoltron/pr/<service-name>:PR-<pr-number>
Additionally it executes all tests on related services.

## Build pipeline
The job is defined in file branch-build.yaml. This is executed on push to master branch.
It is executed while *.go or *.graphql file is commited to master branch.

It builds the services images and pushes them to appropriate repository and location i.e.
gcr.io/projectvoltron/<service-name>:<sha-number>
Additionally it executes all tests on related services and updates the existing cluster.

## Create cluster pipeline
The job is defined in file create_cluster.yaml. This is executed on manual trigger.

It builds the services images and pushes them to appropriate repository and location i.e.
gcr.io/projectvoltron/<service-name>:<image-tag>
As image tag it takes env variable ${IMAGE_TAG} and restores the repo state to provided SHA upon job execution.
Additionally it executes all tests on related services, creates new cluster according to provided values in env_setup.sh, install all dependencies on it and install services.

## Integration tests pipeline
The job is defined in file cluster_integration_tests.yaml. This is executed periodically according to cron settings set in job definition. This run just integration tests.

## Encrypting files
Files encryption is applied with git-crypt.
Currently it works for *.txt files put in data directory.
The definition is made in .gitattributes file.
Current setup is as follows.
```
*.txt filter=git-crypt diff=git-crypt
.gitattributes !filter !diff
```
It means in current setup every *.txt file pushed in data directory will be encrypted.
If you need to encrypt other files in other directory, you have to create there .gittatributes file and apply there proper rules like  for *.txt in given example. Pls. do not forget to add the last line as it prevents .gitattributes file to be encrypted.

Keys ......