#!/bin/bash
set -eEu

# download capectl
curl --fail -Lo /usr/local/bin/capectl https://storage.googleapis.com/projectvoltron_ocftool/${capectl_version}/ocftool-linux-amd64
chmod +x /usr/local/bin/capectl
capectl --version

# download kubectl
curl --fail -Lo /usr/local/bin/kubectl "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl"
chmod +x /usr/local/bin/kubectl
kubectl version
