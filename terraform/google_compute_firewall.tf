# # resource "google_compute_firewall" "gcf-default-allow-http" {
# #   name    = "gke-allow-http-${var.cluster-name}"
# #   project = var.project
# #   network = google_compute_network.gcn_vpc.name

# #   allow {
# #     protocol = "tcp"
# #     ports    = ["80"]
# #   }


# # }

resource "google_compute_firewall" "gcf-default-allow-http-and-https" {
  name    = "gke-allow-http-s-${var.cluster_name}"
  project = var.project
  network = google_compute_network.gcn_vpc.name

  allow {
    protocol = "tcp"
    ports    = ["80","443"]
  }
}

# resource "google_compute_firewall" "gcf-default-allow-all-exp01" {
#   name    = "gke-allow-all-${var.cluster-name}"
#   project = var.project
#   network = google_compute_network.gcn_vpc.name
#   source_ranges = ["192.168.10.0/24"]
# }

# #ok
resource "google_compute_firewall" "gcf-allow-tcp" {
  name    = "gke-allow-master-${var.cluster_name}"
  project = var.project
  network = google_compute_network.gcn_vpc.name
  source_ranges = ["172.16.10.0/28"]

  allow {
    protocol = "tcp"
    ports    = ["443","10250", "80", "8443"]
  }
}

# # resource "google_compute_firewall" "gcf-default-allow-ssh" {
# #   name    = "gcf-allow-ssh-${var.cluster-name}"
# #   project = var.project
# #   network = google_compute_network.gcn_vpc.name

# #   allow {
# #     protocol = "tcp"
# #     ports    = ["22"]
# #   }


# # }

# #ok
# resource "google_compute_firewall" "gcf-allow-all" {
#   name    = "gcf-allow-all-${var.cluster-name}"
#   project = var.project
#   network = google_compute_network.gcn_vpc.name
#   source_ranges = ["10.10.16.0/20"]

#   allow {
#     protocol = "tcp"
#   }
#   allow {
#     protocol = "icmp"
#   }
#   allow {
#     protocol = "udp"
#   }
#   allow {
#     protocol = "esp"
#   }
#   allow {
#     protocol = "ah"
#   }
#   allow {
#     protocol = "sctp"
#   }
# }