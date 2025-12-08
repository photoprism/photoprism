#!/usr/bin/env bash

# Installs additional admin tools for managing Kubernetes deployments on Linux.
# bash <(curl -s https://raw.githubusercontent.com/photoprism/photoprism/develop/scripts/dist/install-k8s-tools.sh)

set -eux;

# Is Go installed?
if ! command -v go &> /dev/null
then
    echo "Go must be installed to run this."
    exit 1
fi

echo "Installing K9s for managing Kubernetes clusters...."
sudo GOBIN="/usr/local/bin" go install github.com/derailed/k9s@latest

echo "Installing Kustomize for Kubernetes configuration management...."
sudo GOBIN="/usr/local/bin" go install sigs.k8s.io/kustomize/kustomize/v5@latest
