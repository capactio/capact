#!/bin/bash
set -o nounset # treat unset variables as an error and exit immediately.
set -o errexit # exit immediately when a command fails.

apt-get update
apt-get install -y pass unzip

# install awscli v2
curl "https://awscli.amazonaws.com/awscli-exe-linux-x86_64.zip" -o "awscliv2.zip"
unzip awscliv2.zip
./aws/install

rm -rf awscliv2.zip aws

# download capact
curl --fail -Lo /usr/local/bin/capact https://storage.googleapis.com/capactio_binaries/${capact_cli_version}/capact-linux-amd64
chmod +x /usr/local/bin/capact
capact --version

# enable capact autocompletion
echo "source <(capact completion bash)" >> /home/${capact_user}/.bashrc

# download kubectl
curl --fail -Lo /usr/local/bin/kubectl "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl"
chmod +x /usr/local/bin/kubectl
kubectl version --client

# enable kubectl autocompletion and set aliases
cat <<EOT >> /home/${capact_user}/.bashrc
source <(kubectl completion bash)
alias k=kubectl
alias kc=kubectl
complete -F __start_kubectl k
EOT

# download helm
curl --fail -Lo helm.tar.gz "https://get.helm.sh/helm-v3.5.3-linux-amd64.tar.gz"
tar xfz helm.tar.gz
cp linux-amd64/helm /usr/local/bin
rm -rf helm.tar.gz linux-amd64
helm version

# enable helm autocompletion
echo "source <(helm completion bash)" >> /home/${capact_user}/.bashrc

# download argo
curl -sLO "https://github.com/argoproj/argo/releases/download/v2.12.11/argo-linux-amd64.gz"
gunzip argo-linux-amd64.gz
chmod +x argo-linux-amd64
mv ./argo-linux-amd64 /usr/local/bin/argo
argo version

# enable argo autocompletion
echo "source <(argo completion bash)" >> /home/${capact_user}/.bashrc
