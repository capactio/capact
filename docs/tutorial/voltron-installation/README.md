# Voltron installation

This tutorial shows how to set up a private Google Kubernetes Engine (GKE) cluster with full Voltron installation. All core Voltron components are located in [`deploy/kubernetes/charts`](../../../deploy/kubernetes/charts). Additionally, Voltron uses [Cert Manager](https://github.com/jetstack/cert-manager/) to generate the certificate for the Voltron Gateway domain.


![overview](assets/overview.svg)

## Table of Contents

<!-- toc -->

- [Prerequisites](#prerequisites)
- [Instructions](#instructions)
  * [Create GKE private cluster](#create-gke-private-cluster)
  * [Install Voltron](#install-voltron)
  * [Clean-up](#clean-up)
  * [Change the source of OCH manifests](#change-the-source-of-och-manifests)

<!-- tocstop -->

##  Prerequisites

* [Helm v3](https://helm.sh/docs/intro/install/) installed
* [`kubectl`](https://kubernetes.io/docs/tasks/tools/install-kubectl/) installed
* [`terraform`](https://learn.hashicorp.com/tutorials/terraform/install-cli) installed
* [`gcloud`](https://cloud.google.com/sdk/docs/install) installed
* A domain for your Google Kubernetes Engine (GKE) cluster
* Google Cloud Platform (GCP) project with Kubernetes Engine API enabled


## Instructions

### Create GKE private cluster 

1. Clone the `master` branch from the `go-voltron` repository.
	
	```bash
	git clone --depth 1 --branch master https://github.com/Project-Voltron/go-voltron.git
	cd ./go-voltron
	```
 
1. Generate a Service Account for Terraform.

    1. Open [https://console.cloud.google.com](https://console.cloud.google.com) and select your project.
    2. In the left pane, go to **Identity** and select **Service accounts**.
    3. Click **Create service account**, name your account, and click **Create**.
    4. Assign the `Compute Network Admin`, `Compute Security Admin`, `Kubernetes Engine Admin`, `Service Account User` roles.
    5. Click **Create key** and choose `JSON` as the key type.
    6. Save the `JSON` file.
    7. Click **Done**.
   	
1. Create a private GKE cluster.
    
    **1. Export the GKE cluster name and region.**
       
    > **NOTE:** To reduce latency when working with a cluster, select the region based on your location.
    
    ```bash
    export CLUSTER_NAME="voltron-demo-v1"
    export REGION="europe-west2"
    ```
       
    **2. Create Terraform variables.**
       
    ```bash
    cat <<EOF > ./hack/ci/terraform/terraform.tfvars
    region="${REGION}"
    cluster_name="${CLUSTER_NAME}"
    google_compute_network_name="vpc-network-${CLUSTER_NAME}"
    google_compute_subnetwork_name="subnetwork-${CLUSTER_NAME}"
    node_pool_name="node-pool-${CLUSTER_NAME}"
    google_compute_subnetwork_secondary_ip_range_name1="gke-pods-${CLUSTER_NAME}"
    google_compute_subnetwork_secondary_ip_range_name2="gke-services-${CLUSTER_NAME}"
    EOF
    ```
       
    **3. Initialize your Terraform working directory.**
       
    ```bash
    terraform -chdir=hack/ci/terraform/ init
    ```
       
    **4. Create a GKE cluster.**
       
    > **NOTE:** This takes around 10 minutes to finish.
    
    ```bash
    GOOGLE_APPLICATION_CREDENTIALS={PATH_TO_SA_JSON_FILE} \
    terraform -chdir=hack/ci/terraform/ apply
    ```
    
    **5. Fetch GKE credentials.**
       
    ```bash
    gcloud container clusters get-credentials $CLUSTER_NAME --region $REGION
    ```
    
    At this point, these are the only IP addresses that have access to the cluster control plane:
     - The primary range of **subnetwork-${CLUSTER_NAME}**
     - The secondary range used for Pods
    
    **6. If you have your machine outside your VPC network, authorize it to access the public endpoint.**
       
    ```bash
    gcloud container clusters update $CLUSTER_NAME --region $REGION \
        --enable-master-authorized-networks \
        --master-authorized-networks $(printf "%s/32" "$(curl ifconfig.me)")
    ```
     
    Now these are the only IP addresses that have access to the control plane:
     - The primary range of **subnetwork-${CLUSTER_NAME}**
     - The secondary range used for Pods
     - Address ranges that you have authorized, for example `203.0.113.0/32`

### Install Voltron

This guide explains how to deploy Voltron on a cluster using your own domain.

>**TIP:** Get a free domain for your cluster using services like [freenom.com](https://www.freenom.com) or similar.

1. Delegate the management of your domain to Google Cloud DNS 

   If your domain is not managed by GCP DNS, follow below steps:
  
   1. Export the project name, the domain name, and the DNS zone name as environment variables. Run these commands:
   
      ```bash
      export GCP_PROJECT={YOUR_GCP_PROJECT} # e.g. projectvoltron
      export DNS_NAME={YOUR_ZONE_DOMAIN} # your custom domain, e.g. dogfooddemo.ga. 
      export DNS_ZONE={YOUR_DNS_ZONE} # e.g. own-domain
      ```

   2. Create a DNS-managed zone in your Google project. Run:
   
      ```bash
      gcloud dns --project=$GCP_PROJECT managed-zones create $DNS_ZONE --description= --dns-name=$DNS_NAME
      ```
   
      Alternatively, create the DNS-managed zone through the GCP UI. In the **Network** section navigate to **Network Services**, click **Cloud DNS**, and select **Create Zone**.
   
   3. Delegate your domain to Google name servers.
   
      - Get the list of the name servers from the zone details. This is a sample list:
    
        ```bash
        ns-cloud-b1.googledomains.com.
        ns-cloud-b2.googledomains.com.
        ns-cloud-b3.googledomains.com.
        ns-cloud-b4.googledomains.com.
        ```
    
      - Set up your domain to use these name servers.
   
   4. Check if everything is set up correctly and your domain is managed by Google name servers. Run:
   
      ```bash
      host -t ns $DNS_NAME
      ```
     
      A successful response returns the list of the name servers you fetched from GCP.
      
      >**NOTE:** It may take a few minutes before the DNS is updated.

1. Export Gateway password and domain name

   ```bash
   export DOMAIN="$CLUSTER_NAME.$(echo $DNS_NAME | sed 's/\.$//')" # eg. `export DOMAIN="demo.cluster.projectvoltron.dev"`
   export GATEWAY_PASSWORD=$(openssl rand -base64 32)
   ```

1. Install Cert Manager
   
   ```bash 
   ./hack/ci/install-cert-manager.sh
   ```

1. Install all Voltron components (Voltron core, Grafana, Prometheus, Neo4J, NGINX, Argo)
   
   ```bash
   CUSTOM_VOLTRON_SET_FLAGS="--set global.domainName=$DOMAIN --set global.gateway.auth.password=$GATEWAY_PASSWORD" \
   DOCKER_REPOSITORY="gcr.io/projectvoltron" \
   OVERRIDE_DOCKER_TAG="ce38bb3" \
   ./hack/ci/cluster-components-install-upgrade.sh
   ```

   >**NOTE:** This command installs ingress which automatically creates a LoadBalancer. If you have your own LoadBalancer, you can use it by adding 
   > `CUSTOM_NGINX_SET_FLAGS="--set ingress-nginx.controller.service.loadBalancerIP={YOUR_LOAD_BALANCER_IP}"` to the above install command. If your domain points to your LoadBalance IP, skip the next step.

   >**NOTE:** To install different Voltron version, change `OVERRIDE_DOCKER_TAG` to different Docker image tag. 

1. Update the DNS record
   
   As the previous step created a LoadBalancer, you now need to create a DNS record for its external IP. 
   
   ```bash
   export EXTERNAL_PUBLIC_IP=$(kubectl get service ingress-nginx-controller -n ingress-nginx -o jsonpath="{.status.loadBalancer.ingress[0].ip}")
   gcloud dns --project=$GCP_PROJECT record-sets transaction start --zone=$DNS_ZONE
   gcloud dns --project=$GCP_PROJECT record-sets transaction add $EXTERNAL_PUBLIC_IP --name=\*.$DOMAIN. --ttl=60 --type=A --zone=$DNS_ZONE
   gcloud dns --project=$GCP_PROJECT record-sets transaction execute --zone=$DNS_ZONE
   ```

1. Check if everything is set up correctly and your domain points to LoadBalancer IP. Run:

   ```bash
   nslookup gateway.$DOMAIN
   ```
   
1. Get information about Voltron Gateway.

   To obtain Gateway URL and authorization information, run:
   
   ```bash
   helm get notes -n voltron-system voltron    
   ```
   
   Example output:
   ```bash
   Thank you for installing Voltron components.
   
   Here is the list of exposed services:
   - Gateway GraphQL Playground: https://gateway.demo.cluster.projectvoltron.dev
   
   Use the following header configuration in the Gateway GraphQL Playground:
   
    {
      "Authorization": "Basic Z3JhcGhxbDpBbjR4YzQwb1M3MEllRnVkd0owcE9Bb2UxU3hVWWJ2a1dxNS8zZVRJZnJNPQ=="
    }
   ```

   Now you are ready to start a journey with the Voltron project. Check out our [Jira installation tutorial](../jira-installation/README.md)!

### Clean-up

When you are done, you can simply remove the whole infrastructure via Terraform:

```bash
GOOGLE_APPLICATION_CREDENTIALS={PATH_TO_SA_JSON_FILE} \
terraform -chdir=hack/ci/terraform/ destroy
```

Additionally, you can remove the Google DNS Zone if not needed. In the **Network** section navigate to **Network Services**, click **Cloud DNS**, select your zoned and click trash icon.


### Change the source of OCH manifests

By default, the OCH manifests are synchronized with the `och-content` directory from the `go-voltron` repository. You can change that by overriding **MANIFEST_PATH** environment variable for **och-public** Deployment.
     
For example, to use manifests Voltron `release-0.2` branch, run:
   
```bash
export SSH_KEY="LS0tLS1CRUdJTiBPUEVOU1NIIFBSSVZBVEUgS0VZLS0tLS0KYjNCbGJuTnphQzFyWlhrdGRqRUFBQUFBQkc1dmJtVUFBQUFFYm05dVpRQUFBQUFBQUFBQkFBQUFNd0FBQUF0emMyZ3RaVwpReU5UVXhPUUFBQUNDSTNEQmpXQjhTT1cvS3pjTWwyWm1ZTmtKRXZQcW5QRDY1RGVhQ0lrMDZjZ0FBQUpBVHIyOHJFNjl2Ckt3QUFBQXR6YzJndFpXUXlOVFV4T1FBQUFDQ0kzREJqV0I4U09XL0t6Y01sMlptWU5rSkV2UHFuUEQ2NURlYUNJazA2Y2cKQUFBRURJTk51VHg3dG1oUVY1dFFxRFVpbDVxOTZwWFE2ak5TdnR2bEVBRmowWUNJamNNR05ZSHhJNWI4ck53eVhabVpnMgpRa1M4K3FjOFBya041b0lpVFRweUFBQUFESE5zYlVCemJHMHRaR1ZzYkFFPQotLS0tLUVORCBPUEVOU1NIIFBSSVZBVEUgS0VZLS0tLS0K"
export BRANCH_NAME=release-0.2
kubectl set env deployment/voltron-och-public -n voltron-system --containers="och-public-populator" MANIFESTS_PATH="git@github.com:Project-Voltron/go-voltron.git?sshkey=${SSH_KEY}&ref=${BRANCH_NAME}"
```

> **NOTE:** The `sshkey` parameter is a Base64 encoded private key used by populator to download manifests. It has read only access.
> **NOTE:** Populator uses [go-getter](https://github.com/hashicorp/go-getter) so different sources are supported e.g. S3, GCS, etc.
