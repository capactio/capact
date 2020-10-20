resource "google_compute_router" "gcr_router" {
  project = var.project
  name    = var.cluster-name
  network = google_compute_network.gcn_vpc.name
  
}

resource "google_compute_router_nat" "gcrn_nat" {
  name                               = var.cluster-name
  project                            = var.project
  router                             = google_compute_router.gcr_router.name
  region                             = google_compute_router.gcr_router.region
  source_subnetwork_ip_ranges_to_nat = "ALL_SUBNETWORKS_ALL_IP_RANGES"
  nat_ip_allocate_option             = "AUTO_ONLY"
}