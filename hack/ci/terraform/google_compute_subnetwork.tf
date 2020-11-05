resource "google_compute_subnetwork" "gcs_compute_subnetwork"{
  name   = var.google_compute_subnetwork_name
  ip_cidr_range = var.google_compute_subnetwork_ip_cidr_range
  project = var.project
  region        = var.region
  network       = google_compute_network.gcn_vpc.id
  private_ip_google_access = true
  secondary_ip_range {
    range_name = var.google_compute_subnetwork_secondary_ip_range_name1   
    ip_cidr_range = var.google_compute_subnetwork_secondary_ip_range_cidr1
  }

  secondary_ip_range {
    range_name = var.google_compute_subnetwork_secondary_ip_range_name2  
    ip_cidr_range = var.google_compute_subnetwork_secondary_ip_range_cidr2
  }
}


