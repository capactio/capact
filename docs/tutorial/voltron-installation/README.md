# Voltron installation

This tutorial shows how to install all Voltron components on GKE private cluster. All core Voltron components are located in [`deploy/kubernetes/charts`](../../../deploy/kubernetes/charts). Additionally, [Cert Manager](https://github.com/jetstack/cert-manager/) is used for generating certificate for Voltron Gateway domain.

###  Prerequisites

* Install [Helm v3](https://helm.sh/docs/intro/install/)
* Install [`kubectl`](https://kubernetes.io/docs/tasks/tools/install-kubectl/)
* Access to Google Cloud Platform 

### Goal

This instruction will guide you through the installation of Voltron on the private GKE cluster. Setting domain sections are tightly coupled with the `projectvoltron` GCP project. If you do not have access to it you need to have our own domain managed by GCP and dedicated LoadBalancer IP. 

1. Clone the `master` branch from the `go-voltron` repository:

	```bash
	git clone --depth 1 --branch master https://github.com/Project-Voltron/go-voltron.git
	cd ./go-voltron
	```

1. Generate Service Account for Terraform:

	* Open [https://console.cloud.google.com](https://console.cloud.google.com) and select your project.
   
   	* On the left pane, go to **Identity** and select **Service accounts**.
   
   	* Click **Create service account**, name your account, and click **Create**.
   
   	* Set `Compute Network Admin`, `Compute Security Admin`, `Kubernetes Engine Admin`, `Service Account User` roles.
   
   	* Click **Create key** and choose `JSON` as a key type.
   
   	* Save the `JSON` file.
   
   	* Click **Done**.

1. Create a GKE private cluster:

    **Export GKE cluster name**

    > **NOTE:** To reduce latency when working with a cluster, select region based on your location.
    
    ```bash
    export NAME=voltron-demo-v1
    export REGION="europe-west2"
    export DOMAIN="demo.cluster.projectvoltron.dev" # you can use your own domain if you have one.
    ```

    **Create Terraform variables**

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

    **Initialize Terraform working directory**

    ```bash
    terraform -chdir=hack/ci/terraform/ init
    ```

    **Create GKE cluster**

    > **NOTE:** This takes around 10 minutes to finish.

    ```bash
    GOOGLE_APPLICATION_CREDENTIALS={PATH_TO_SA_JSON_FILE} \
    terraform -chdir=hack/ci/terraform/ apply
    ```

    **Fetch GKE credentials**
    
    ```bash
    gcloud container clusters get-credentials $NAME --region $REGION
    ```
    
    At this point, these are the only IP addresses that have access to the cluster control plane:

    - The primary range of "subnetwork-${NAME}".
    - The secondary range used for Pods.

    Suppose you have your machine, outside of your VPC network. To authorize your machine to access the public endpoint, run:

    ```bash
    gcloud container clusters update $NAME --region $REGION \
        --enable-master-authorized-networks \
        --master-authorized-networks $(printf "%s/32" "$(curl ifconfig.me)")
    ```

    Now these are the only IP addresses that have access to the control plane:

    - The primary range of "subnetwork-${NAME}".
    - The secondary range used for Pods.
    - Address ranges that you have authorized, for example, 203.0.113.0/32.


1. Install Voltron:

    **Install Cert Manager**

    ```bash 
    ./hack/ci/install-cert-manager.sh
    ```

    **Install all Voltron components (Voltron core, Grafana, Prometheus, Neo4J, NGINX, Argo)**

    ```bash
    CUSTOM_VOLTRON_SET_FLAGS="--set global.domainName=$DOMAIN" \
    DOCKER_REPOSITORY="gcr.io/projectvoltron" \
    OVERRIDE_DOCKER_TAG="76a84bf" \
    ./hack/ci/cluster-components-install-upgrade.sh
    ```

    >**NOTE:** This commands installs ingress which automatically creates LoadBalancer. If you have LoadBalancer you can use it by adding 
    > CUSTOM_NGINX_SET_FLAGS="--set ingress-nginx.controller.service.loadBalancerIP={YOUR_LOAD_BALANCER_IP}" to the above install command. If own LoadBalance is used, skip next step.
    
    1. Update DNS record:
    >**NOTE:** As the previous step created LoadBalance, now we need to create DNS for its external IP. 
    
    ```bash
    export EXTERNAL_PUBLIC_IP=$(kubectl get service ingress-nginx-controller -n ingress-nginx -o jsonpath="{.status.loadBalancer.ingress[0].ip}")
    gcloud dns --project=projectvoltron record-sets transaction start --zone=cluster-voltron
    gcloud dns --project=projectvoltron record-sets transaction add $EXTERNAL_PUBLIC_IP --name=\gateway.$DOMAIN. --ttl=60 --type=A --zone=$DNS_ZONE
    gcloud dns --project=projectvoltron record-sets transaction execute --zone=cluster-voltron
    ```
