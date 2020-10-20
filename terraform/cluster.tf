resource "google_container_cluster" gcc-cls {
  name               = var.cluster-name
  location           = var.location
  project            = var.project
  network            = google_compute_network.gcn_vpc.name
  subnetwork         = google_compute_subnetwork.gcs_compute_subnetwork.name
  initial_node_count = 1
  monitoring_service = "monitoring.googleapis.com/kubernetes"
  logging_service    = "logging.googleapis.com/kubernetes" 
  remove_default_node_pool = true

    ip_allocation_policy {
      cluster_secondary_range_name = var.google_compute_subnetwork_secondary_ip_range_name1
      services_secondary_range_name = var.google_compute_subnetwork_secondary_ip_range_name2
    }

  private_cluster_config {
    enable_private_nodes = true
    enable_private_endpoint = false
    master_ipv4_cidr_block = var.google_container_cluster_private_cluster_config_master_ipv4_cidr_block
  }

  cluster_autoscaling {
    enabled = true
    resource_limits {
      resource_type = "cpu"
      minimum = "1"
      maximum = "2"
    }
    resource_limits {
      resource_type = "memory"
      minimum = "6"
      maximum = "12"
    }
  }

  maintenance_policy {
    recurring_window {
    start_time = "2019-08-01T02:00:00Z"
    end_time = "2019-08-01T06:00:00Z"
      recurrence = "FREQ=WEEKLY;BYDAY=MO,TU,WE,TH,FR"
  }
}

  master_auth {
    username = ""
    password = ""

    client_certificate_config {
      issue_client_certificate = false
    }
  }


  node_config {
#    oauth_scopes = [
#      "https://www.googleapis.com/auth/logging.write",
#      "https://www.googleapis.com/auth/monitoring",
#    ]
    machine_type       = var.machine_type

    metadata = {
      disable-legacy-endpoints = "true"
    }

  }

  timeouts {
    create = "30m"
    update = "40m"
  }
  master_authorized_networks_config {

  }
}

resource "google_container_node_pool" "gcnp-container-node-pool" {
  name       = var.node-pool-name
  location   = var.location
  cluster    = google_container_cluster.gcc-cls.name
  node_count = 1
  project    = var.project

  autoscaling {
    min_node_count = 1
    max_node_count = var.autoscaling_max_node_count
  }
  node_config {
    preemptible  = var.preemptible
    machine_type = var.machine_type
    disk_type    = var.disk_type
    image_type   = var.image_type

    metadata = {
      disable-legacy-endpoints = "true"
    }

    oauth_scopes = [
      "https://www.googleapis.com/auth/cloud-platform",
    ]
  }
}



