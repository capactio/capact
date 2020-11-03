resource "google_compute_network" "gcn_vpc" {
  name   = var.google_compute_network_name
  project = var.project
  auto_create_subnetworks = false
}
