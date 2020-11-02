# CI/CD jobs documentation
This document elaborates on jobs created to automate the process of development/building  new releases.
It touches the matter of encrypting files in repo, which was a part of the task

### Entry information
In main folder file called env_setup.sh exists. This setups the environment for every job.

## PR pipeline
The job is defined in file pr-build.yaml. This is executed on pull request while it is: opened, synchronize or reopened. The branch related is master.
It is executed while *.go or *.graphql file is commited.

It builds the services images and pushes them to appropriate repository and location i.e.
gcr.io/projectvoltron/pr/<<service-name>>:PR-<<pr-number>>
Additionally it executes all tests on related services.

## Build pipeline
The job is defined in file branch-build.yaml. This is executed on push to master branch.
It is executed while *.go or *.graphql file is commited to master branch.

It builds the services images and pushes them to appropriate repository and location i.e.
gcr.io/projectvoltron/<<service-name>>:<<sha-number>>
Additionally it executes all tests on related services and updates the existing cluster.

## Create cluster pipeline
The job is defined in file create_cluster.yaml. This is executed on manual trigger.

It builds the services images and pushes them to appropriate repository and location i.e.
gcr.io/projectvoltron/<<service-name>>:<<image-tag>>
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

To decrypt the data locally you need to use either symetric key or I need to add your keys.
The procedure of decrypting files and working in team with that is perfectly described here -->https://buddy.works/guides/git-crypt#working-in-team-with-git-crypt

Currently a file decrypt.yaml shows how to decrypt an encrypted file.

## Cookbook new pipeline related.
To create a new pipeline you must follow the rules of syntax related to GitHub Actions.
In short take one of the current job defined in .github/workflows in *.yaml. Rename it. Update the:
.name
.jobs.<jobs-name>
.jobs.<jobs-name>.name
.jobs.<jobs-name>.step.name
Update the steps accordingly to actions you want to execute.

The following steps are necessarry to checkout the code, setup go environment, authorize to GCR and GKE in case they are necesary.
```
    steps:    
      #Checkout code & GCR authorization
      - name: Checkout code
        uses: actions/checkout@v2
      - uses: GoogleCloudPlatform/github-actions/setup-gcloud@master
        with:
          export_default_credentials: true
          service_account_key: ${{ secrets.GCR_CREDS }}

      - name: Setup env
        run: |
          . ./env_setup.sh

      - name: Setup golang
        uses: actions/setup-go@v2
        with:
          go-version: ${{env.GO_VERSION}}

      #Checkout coud & GKE authorization
      - name: Checkout code
        uses: actions/checkout@v2
      - uses: GoogleCloudPlatform/github-actions/setup-gcloud@master
        with:
          service_account_key: ${{ secrets.GKE_CREDS }}
          export_default_credentials: true
```
Pls. remember to update env_setup.sh file accordingly.

### Let's encrypt certificates
Currently the CI/CD building cluster creates both Certificates Issuers (stag & prod) for Let's Encrypt. 
Current setup is that the certificates are issued with STAG LE. To switch it to PROD in every values.yaml file you have to adapt .ingress.annotations.issuer and set it to letsencrypt (it is letsencrypt-stag currently).


