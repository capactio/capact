resource "google_compute_firewall" "gcf-default-allow-http-and-https" {
  name    = "gke-allow-http-s-${var.cluster_name}"
  project = var.project
  network = google_compute_network.gcn_vpc.name

  allow {
    protocol = "tcp"
    ports    = ["80", "443"]
  }
}

resource "google_compute_firewall" "gcf-allow-tcp" {
  name          = "gke-allow-master-${var.cluster_name}"
  project       = var.project
  network       = google_compute_network.gcn_vpc.name
  source_ranges = ["172.16.10.0/28"]

  allow {
    protocol = "tcp"
    ports    = ["443", "10250", "80", "8443"]
  }
}
