#!/bin/bash

set -x

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
REPO_ROOT="${SCRIPT_DIR}/.."
CLUSTER=${CLUSTER:-"acceptance"}
IMAGE=${IMAGE:-kind-cluster-controller:latest}

ensure_kind_cluster() {
  local cluster
  cluster="$1"
  if ! kind get clusters | grep -q "$cluster"; then
    kind create cluster --name "$cluster" --wait 5m --config "${REPO_ROOT}/tests/assets/kind-cluster-with-docker-sock-mount.yaml"
  fi
  kind export kubeconfig --name "$cluster" --kubeconfig "$HOME/.kube/$cluster.yml"
}

ensure_kind_cluster "$CLUSTER"
clusterctl init --kubeconfig "$HOME/.kube/$CLUSTER.yml"
kind load docker-image --name "$CLUSTER" "$IMAGE"
