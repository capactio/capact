# Capact installation on Google Cloud Platform

This tutorial shows how to set up a private Google Kubernetes Engine (GKE) cluster with full Capact installation. All core Capact components are located in [`deploy/kubernetes/charts`](../../../deploy/kubernetes/charts). Additionally, Capact uses [Cert Manager](https://github.com/jetstack/cert-manager/) to generate the certificate for the Capact Gateway domain.

![overview](assets/overview.svg)

## Table of Contents

<!-- toc -->

- [Prerequisites](#prerequisites)
- [Instructions](#instructions)
  * [Create GKE private cluster](#create-gke-private-cluster)
  * [Install Capact](#install-capact)
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

[![asciicast](https://asciinema.org/a/395877.svg)](https://asciinema.org/a/395877)

### Create GKE private cluster 

1. Check the latest stable Capact release on [Capact GitHub releases](https://github.com/capactio/capact/releases) page.
   
   Note the git tag for the latest version (for example, `v0.3.0`).

1. Clone the `capact` repository for a specific tag noted from previous version.
	
	```bash
	git clone --depth 1 --branch {tag} https://github.com/capactio/capact.git
	cd ./capact
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
    export CLUSTER_NAME="capact-demo-v1"
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

### Install Capact

This guide explains how to deploy Capact on a cluster using your own domain.

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
   export DOMAIN="$CLUSTER_NAME.$(echo $DNS_NAME | sed 's/\.$//')" # e.g. `export DOMAIN="demo.cluster.capact.dev"`
   export GATEWAY_PASSWORD=$(openssl rand -base64 32)
   ```

1. Install all Capact components (Capact core, Grafana, Prometheus, Neo4J, NGINX, Argo, Cert Manager)
   
   ```bash
   CUSTOM_CAPACT_SET_FLAGS="--set global.domainName=$DOMAIN --set global.gateway.auth.password=$GATEWAY_PASSWORD" \
   DOCKER_REPOSITORY="gcr.io/projectvoltron" \
   OVERRIDE_DOCKER_TAG="ce38bb3" \
   ./hack/ci/cluster-components-install-upgrade.sh
   ```

   >**NOTE:** This command installs ingress which automatically creates a LoadBalancer. If you have your own LoadBalancer, you can use it by adding 
   > `CUSTOM_NGINX_SET_FLAGS="--set ingress-nginx.controller.service.loadBalancerIP={YOUR_LOAD_BALANCER_IP}"` to the above install command. If your domain points to your LoadBalance IP, skip the next step.

   >**NOTE:** To install different Capact version, change `OVERRIDE_DOCKER_TAG` to different Docker image tag. 

1. Update the DNS record
   
   As the previous step created a LoadBalancer, you now need to create a DNS record for its external IP. 
   
   ```bash
   export EXTERNAL_PUBLIC_IP=$(kubectl get service ingress-nginx-controller -n capact-system -o jsonpath="{.status.loadBalancer.ingress[0].ip}")
   gcloud dns --project=$GCP_PROJECT record-sets transaction start --zone=$DNS_ZONE
   gcloud dns --project=$GCP_PROJECT record-sets transaction add $EXTERNAL_PUBLIC_IP --name=\*.$DOMAIN. --ttl=60 --type=A --zone=$DNS_ZONE
   gcloud dns --project=$GCP_PROJECT record-sets transaction execute --zone=$DNS_ZONE
   ```

1. Check if everything is set up correctly and your domain points to LoadBalancer IP. Run:

   ```bash
   nslookup gateway.$DOMAIN
   ```
   
1. Get information about Capact Gateway.

   To obtain Gateway URL and authorization information, run:
   
   ```bash
   helm get notes -n capact-system capact    
   ```
   
   Example output:
   ```bash
   Thank you for installing Capact components.
   
   Here is the list of exposed services:
   - Gateway GraphQL Playground: https://gateway.demo.cluster.capact.dev
   
   Use the following header configuration in the Gateway GraphQL Playground:
   
    {
      "Authorization": "Basic Z3JhcGhxbDpBbjR4YzQwb1M3MEllRnVkd0owcE9Bb2UxU3hVWWJ2a1dxNS8zZVRJZnJNPQ=="
    }
   ```

   **✨ Now you are ready to start a journey with the Capact project. Check out our [Mattermost installation tutorial](../mattermost-installation/README.md)!**

### Clean-up

When you are done, you can simply remove the whole infrastructure via Terraform:

```bash
GOOGLE_APPLICATION_CREDENTIALS={PATH_TO_SA_JSON_FILE} \
terraform -chdir=hack/ci/terraform/ destroy
```

Additionally, you can remove the Google DNS Zone if not needed. In the **Network** section navigate to **Network Services**, click **Cloud DNS**, select your zoned and click trash icon.


### Change the source of OCH manifests

By default, the OCH manifests are synchronized with the `och-content` directory from the `capact` repository on a specific release branch. You can change that by overriding **MANIFEST_PATH** environment variable for **och-public** Deployment.
     
For example, to use the `main` branch as a source of OCH manifests, run:
   
```bash
export SSH_KEY="LS0tLS1CRUdJTiBPUEVOU1NIIFBSSVZBVEUgS0VZLS0tLS0KYjNCbGJuTnphQzFyWlhrdGRqRUFBQUFBQkc1dmJtVUFBQUFFYm05dVpRQUFBQUFBQUFBQkFBQUJsd0FBQUFkemMyZ3RjbgpOaEFBQUFBd0VBQVFBQUFZRUFtYUxFMi85eW5Da1pDbWs5VU1tZnhxUUR5WlJHZi91RlZ5Tzg1a2FLKzVOTlVzUWs2RHFaCk5ZOVpuOGg5SkE2dFlnMWtRRTVadmhJUmxqUjViU1FJcmYySnR3bWV4SzMzdDJQTzk2blE3ZUlobkp4RmtvenZaenJxRkIKSkVhWmZZSkdWSDZNRlI4YjMzeWpOZmtxQW5lM0N2UkFsZXBITTErbVZIMFQxOGtSeTJXekRsTVlxRXlyT2JoUG8wdS9DSQpraGVpdkV5OTd5UXJiT0VmcFZGaEZlVG9Cd09hdjdWTUs5M1IzSm9TVFBQVzBncG9keFI3NG9HRHR6eWtCbW1WMHFuazljCkJpMXBxcFJCcUYyaS9ka1U4NUN6N1Azbk1ZaW9xV2pKcUNGUEY3VVZSWDRuNUtJeDRmK1pnZnRhWk04NkFzRmpGRVA3RnUKdXJVdmswdW1PTnU4QzBDeHpobFBTNENLN1NUWVphSmRyaW9Ibm0vL1VWaEIyL1pJVVcyN2VrSWVqRWhMNDBiZ0JScmZBLwplaTBnOHJ3eFVvOEhLbVY5TlVKdi9JYVh1Q3RINk1peUZFeVU3cTEyZ3IveGR1czN5RlJUakFxRERvRTgyV2ZsOGtQSTJsCmY4OTdpbDliZ1A3K0ZsRmFxYk5RK20wazhLWVRyUzk3dkFjbFB1ak5BQUFGaUJoNkQvUVllZy8wQUFBQUIzTnphQzF5YzIKRUFBQUdCQUptaXhOdi9jcHdwR1FwcFBWREpuOGFrQThtVVJuLzdoVmNqdk9aR2l2dVRUVkxFSk9nNm1UV1BXWi9JZlNRTwpyV0lOWkVCT1diNFNFWlkwZVcwa0NLMzlpYmNKbnNTdDk3ZGp6dmVwME8zaUlaeWNSWktNNzJjNjZoUVNSR21YMkNSbFIrCmpCVWZHOTk4b3pYNUtnSjN0d3IwUUpYcVJ6TmZwbFI5RTlmSkVjdGxzdzVUR0toTXF6bTRUNk5MdndpSklYb3J4TXZlOGsKSzJ6aEg2VlJZUlhrNkFjRG1yKzFUQ3ZkMGR5YUVrenoxdElLYUhjVWUrS0JnN2M4cEFacGxkS3A1UFhBWXRhYXFVUWFoZApvdjNaRlBPUXMrejk1ekdJcUtsb3lhZ2hUeGUxRlVWK0orU2lNZUgvbVlIN1dtVFBPZ0xCWXhSRCt4YnJxMUw1TkxwampiCnZBdEFzYzRaVDB1QWl1MGsyR1dpWGE0cUI1NXYvMUZZUWR2MlNGRnR1M3BDSG94SVMrTkc0QVVhM3dQM290SVBLOE1WS1AKQnlwbGZUVkNiL3lHbDdnclIraklzaFJNbE82dGRvSy84WGJyTjhoVVU0d0tndzZCUE5sbjVmSkR5TnBYL1BlNHBmVzREKwovaFpSV3FtelVQcHRKUENtRTYwdmU3d0hKVDdvelFBQUFBTUJBQUVBQUFHQVpQWnVZQlRUQUlTUmpDSDB4VzU2clZPRG1hCmp6VzQreTVMejdtbWlwVlFKTVFpUGNEVERWRmptS01GTFV5aWxMRDdDMVBQMUFSSVFqUW81aGJiUE1jR3E1WWF2VXhuTjgKNHV1WVMzRXhkK0t2Sy9nV1VHU0Z2MVVjRnV5YVFMb2t0R1pLaDA3anh2V01MVGp0aWJIdHdGVWhHSmovdFJweFVvZlVWbApFTjExOERCNUp1Uzh3M3orMlFPaWNqR0k3TmNSUlBRV2M5T2phT3d4SitkV214WDIzNmZRR0ZaSTZEN0IvdGxnYzZGNC8yCmtEbCt4U0tVOXhrZHNnRUwyWnJ3Y2J1elRna0tLbzdRQXhJaWNNbDY1b0lKTnB5ZlNSNWFIeWJ1RkF6RVlKbFluQ3F5UWkKdWpaSVJvQ2psMjNtK1RSOUJJUHhPcXM3SkZZYW4wY0I5ODAvQXpBN0VFS3Y0TW0rVkk0bjdyMFN1T3pwMmJYc1d2aVI2bQpxNkpkOVBTNXB0QVdLWlg1Sm9lTlNDZERIREhOK2dyMnoyV0VXYS9GcTlDSW9kQWE4cWR4cXJha3pvYllkVHp6QjJNK2JLCmxFMFRqWGdsc3JTV0xNL1dPVEc0ODYybUo0L2hockJ3ZEJDb0FUZ1lGbmdxMWVmU3FuQTMrREw2cXA1VEZTZ2JPQkFBQUEKd1FDenNuNVp0cDYwZkNOaTBEaWttamZMdHpFVXNIYW9GbTBuSFcydVN4SXQwUXJ3aWF3cktla0RHeUcxdjVIVEE0RFVBOApDNnpJbnR4NklidmZPZ0loZFZBRFJaZ2tqU0w1T3kydGhuSllneDZLRVdIYmZ5ME1heXlZUmdCN2dYdzFpcVdKY1dZVnQzCndhOGowR0k1WkM2UmJ0YStMSWpJQm5HSkxVbVJ0L0Z4ZDMzKzB5Ym5xS0hITnB2eUpWenRnUlBSM0JYK1grUGhNZE93VFgKbGgzSU9ZM3l6K1NabEdZcWdSaDMybXc3eGNJK2hGaCt4Rk9LaTJHdDdPMGg1bmVwTUFBQURCQU1sNFNNNGdVb1ZxMDhYbgo5N3JxdHhaR1FKUExmVythWHVERkFiNzNCSHVLZmIzOWNma2ZqR3dSWERYMUZGclgyWGRPUVZIdzR6ank2Skk0MWJrNUhNCmUyTkFvYUNkM0RMRXZhQkc0aWlNOVBabGJtdlVmbVBTNGhtQldoVklzdUtDc1JaakZCdk41NzdsWnhUTFBTK0pZSTg0aEsKNVdrczUwTHlyRXc2aUZ1OUJlSXgrRHUycnNmL2xqQ1dHNDBoVHJBckRoNEw1R2ZUa0MvTDY3L25zYldlR2ZFQ0dUQis4aApNNVcwZm9wSlRiSUZTaTFWOGp6NENJRDE2NXd3SUZwUUFBQU1FQXd6Z1lRYzBNSW9nbDJJanMrUEZ2aFluak93QUU2V2RwCllxQnZidE5OWlgwQUVDVkpSSkNkVktvZEFVZXdhU2NnY0Q2RDQ5L2lUVnF1Uk8wY0Y2RThwZFF5MXBkVlBwNXJVSkx0MUYKZndoMTlha2drWkRDSzR3TVEwZndVNGZmbmZGbHJBUFVmVkQ1VkNHSHphUTNjSmdkbVFCWTJLWE9iNlNZRmJmcWdTZFVDSwoxakpWUmF1TWlxaFM4QS9mT29waXR2QlBtQXlvUy9xV0JHTDA5SVVtNkovNEtqSTJpTmhaNm9zdWg4L0dsT3dudlgwUGRGCjRvYUpWTDR3WnYxLzRKQUFBQUVuUmxZVzB0WkdWMlFHTmhjR0ZqZEM1cGJ3PT0KLS0tLS1FTkQgT1BFTlNTSCBQUklWQVRFIEtFWS0tLS0tCg=="
export BRANCH_NAME=main
kubectl set env deployment/capact-och-public -n capact-system --containers="och-public-populator" MANIFESTS_PATH="git@github.com:capactio/capact.git?sshkey=${SSH_KEY}&ref=${BRANCH_NAME}"
```

> **NOTE:** The `sshkey` parameter is a Base64 encoded private key used by populator to download manifests. It has read only access.
> **NOTE:** Populator uses [go-getter](https://github.com/hashicorp/go-getter) so different sources are supported e.g. S3, GCS, etc.
