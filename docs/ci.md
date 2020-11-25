# CI/CD jobs documentation
This document elaborates on jobs created to automate the process of development/building  new releases.
It touches the matter of encrypting files in repo, which was a part of the task

### Entry information
In folder  /hack/ci file called [setup-env.sh](../hack/ci/setup-env.sh)  exists. This setups the environment for every job.

## PR pipeline
The job is defined in file [pr-build.yaml](../.github/workflows/pr-build.yaml). This is executed on pull request while it is: opened, synchronize or reopened. The branch related is master.
It is executed while *.go or *.graphql file is commited.

It builds the services images and pushes them to appropriate repository and location i.e.
gcr.io/projectvoltron/pr/\<service-name\>:PR-\<pr-number\>  
Additionally it executes all tests on related services.

## Build pipeline
The job is defined in file [branch-build.yaml](../.github/workflows/branch-build.yaml). This is executed on push to master branch.
It is executed while *.go or *.graphql file is commited to master branch.

It builds the services images and pushes them to appropriate repository and location i.e.
gcr.io/projectvoltron/\<service-name\>:\<sha-number\>  
Additionally it executes all tests on related services and updates the existing cluster.

## Create cluster pipeline
The job is defined in file [create_cluster.yaml](../.github/workflows/create_cluster.yaml). This is executed on manual trigger. https://github.blog/changelog/2020-07-06-github-actions-manual-triggers-with-workflow_dispatch/  
In case the cluster is recreated within the same region and the same name: cluster,router & vpc should be deleted.

It uses already existing images. 
As image tag it takes env variable ${IMAGE_TAG} and restores the repo state to provided SHA upon job execution.  
Additionally it executes all tests on related services, creates new cluster according to provided values in above mentioned setup-env.sh, install all dependencies on it and install services.

## Integration tests pipeline
The job is defined in file [cluster_integration_tests.yaml](../.github/workflows/cluster_integration_tests.yaml). This is executed periodically according to cron settings set in job definition. This run just integration tests.  
This is executed on master branch.

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

Currently a file [decrypt.yaml](../.github/workflows/decrypt.yaml) shows how to decrypt an encrypted file in ci/cd.

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
          . ./hack/ci/setup-env.sh

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
Pls. remember to update .setup-env.sh file accordingly.

### Let's encrypt certificates
Currently the CI/CD building cluster creates both Certificates Issuers (stag & prod) for Let's Encrypt.  
Current setup is that the certificates are issued with STAG LE. To switch it to PROD in every values.yaml file you have to adapt .ingress.annotations.issuer and set it to letsencrypt (it is letsencrypt-stag 
currently).

## Repo secrets and serivce accounts.
Following secrets are defined in repo and are used:  
GCR_CREDS - used for pushing data to GCR.  
GKE_CREDS - used for GKE cluster creation and management.  
GIT_CRYPT_KEY - is a symetric key used to decrypt files encrypted with git crypt.

For the GCR service account one needs at least following permissions to push the image to registry -->https://cloud.google.com/container-registry/docs/access-control#permissions_and_roles  

For GKE service account following:
```
Role	                Title	        Description	Lowest resource
roles/container.admin	Kubernetes    Provides access to full management of clusters and their Kubernetes  
                      Engine        API objects.   
                      Admin
                                    To set a service account on nodes, you must also grant the Service Account User role (roles/iam.serviceAccountUser)                       
```





