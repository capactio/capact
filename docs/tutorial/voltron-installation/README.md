# Voltron installation

This tutorial shows how to set up a private Google Kubernetes Engine (GKE) cluster with full Voltron installation. All core Voltron components are located in [`deploy/kubernetes/charts`](../../../deploy/kubernetes/charts). Additionally, Voltron uses [Cert Manager](https://github.com/jetstack/cert-manager/) to generate the certificate for the Voltron Gateway domain.


![overview](assets/overview.svg)

###  Prerequisites

* [Helm v3](https://helm.sh/docs/intro/install/) installed
* [`kubectl`](https://kubernetes.io/docs/tasks/tools/install-kubectl/) installed
* [`terraform`](https://learn.hashicorp.com/tutorials/terraform/install-cli) installed
* [`gcloud`](https://cloud.google.com/sdk/docs/install) installed
* Access to the `projectvoltron` GCP project/own domain and LoadBalancer IP 
   > **NOTE:** Setting the domain sections is tightly coupled with the `projectvoltron` GCP project. If you do not have access to it, you need to have your own domain managed by Google DNS and LoadBalancer IP. 

### Instructions

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
    export NAME="voltron-demo-v1"
    export REGION="europe-west2"
    ```
       
    **2. Create Terraform variables.**
       
    ```bash
    cat <<EOF > ./hack/ci/terraform/terraform.tfvars
    region="${REGION}"
    cluster_name="${NAME}"
    google_compute_network_name="vpc-network-${NAME}"
    google_compute_subnetwork_name="subnetwork-${NAME}"
    node_pool_name="node-pool-${NAME}"
    google_compute_subnetwork_secondary_ip_range_name1="gke-pods-${NAME}"
    google_compute_subnetwork_secondary_ip_range_name2="gke-services-${NAME}"
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
    gcloud container clusters get-credentials $NAME --region $REGION
    ```
    
    At this point, these are the only IP addresses that have access to the cluster control plane:
     - The primary range of **subnetwork-${NAME}**
     - The secondary range used for Pods
    
    **6. If you have your machine outside your VPC network, authorize it to access the public endpoint.**
       
    ```bash
    gcloud container clusters update $NAME --region $REGION \
        --enable-master-authorized-networks \
        --master-authorized-networks $(printf "%s/32" "$(curl ifconfig.me)")
    ```
     
    Now these are the only IP addresses that have access to the control plane:
     - The primary range of **subnetwork-${NAME}**
     - The secondary range used for Pods
     - Address ranges that you have authorized, for example `203.0.113.0/32`

1. Install Voltron.
    
    **1. Export Gateway password and domain name.**
    
    ```bash
    export DOMAIN="demo.cluster.projectvoltron.dev" # you can use your own domain if you have one
    export GATEWAY_PASSWORD=$(openssl rand -base64 32)
    ```
    
    **1. Install Cert Manager.**
       
    ```bash 
    ./hack/ci/install-cert-manager.sh
    ```

    **2. Install all Voltron components (Voltron core, Grafana, Prometheus, Neo4J, NGINX, Argo).**
       
    ```bash
    CUSTOM_VOLTRON_SET_FLAGS="--set global.domainName=$DOMAIN --set global.gateway.auth.password=$GATEWAY_PASSWORD" \
    DOCKER_REPOSITORY="gcr.io/projectvoltron" \
    OVERRIDE_DOCKER_TAG="76a84bf" \
    ./hack/ci/cluster-components-install-upgrade.sh
    ```
    
    >**NOTE:** This commands installs ingress which automatically creates a LoadBalancer. If you have your own LoadBalancer, you can use it by adding 
    > `CUSTOM_NGINX_SET_FLAGS="--set ingress-nginx.controller.service.loadBalancerIP={YOUR_LOAD_BALANCER_IP}"` to the above install command. In such a case, skip the next step.
 
    **3. Update the DNS record.**
       
    As the previous step created a LoadBalancer, you now need to create a DNS record for its external IP. 
    
    ```bash
    export EXTERNAL_PUBLIC_IP=$(kubectl get service ingress-nginx-controller -n ingress-nginx -o jsonpath="{.status.loadBalancer.ingress[0].ip}")
    export DNS_ZONE=cluster-voltron
    export GCP_PROJECT=projectvoltron
    gcloud dns --project=$GCP_PROJECT record-sets transaction start --zone=$DNS_ZONE
    gcloud dns --project=$GCP_PROJECT record-sets transaction add $EXTERNAL_PUBLIC_IP --name=\*.$DOMAIN. --ttl=60 --type=A --zone=$DNS_ZONE
    gcloud dns --project=$GCP_PROJECT record-sets transaction execute --zone=$DNS_ZONE
    ```

1. Get information about Voltron Gateway.

    To obtain Gatway URL and authorization information, run:
    
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
