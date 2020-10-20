cluster-name = "dev01"
location = "europe-west1"
node-pool-name = "dev01-node-pool"
google_compute_network_name = "dev01-vpc-network"
google_compute_subnetwork_name = "dev01-subnetwork"
#google_compute_subnetwork_ip_cidr_range = "192.168.10.0/24"
google_compute_subnetwork_secondary_ip_range_name1 = "gke-dev01-pods" 
#google_compute_subnetwork_secondary_ip_range_cidr1 = "10.10.16.0/20" 
google_compute_subnetwork_secondary_ip_range_name2 = "gke-dev01-services" 
#google_compute_subnetwork_secondary_ip_range_cidr2 = "10.10.0.0/23" 
#google_container_cluster_private_cluster_config_master_ipv4_cidr_block = "172.16.10.0/28" 
project = "development-gg"
machine_type = "n1-standard-2"
lp-name = "lpdev01.zipzero.com"
preemptible = "true"
autoscaling_max_node_count = 2
disk_type = "pd-standard"
image_type = "COS"




