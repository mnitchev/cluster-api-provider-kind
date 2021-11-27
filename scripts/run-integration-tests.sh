#!/bin/bash

set -x

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
REPO_ROOT="${SCRIPT_DIR}/.."
CLUSTER=${CLUSTER:-"integration"}

ensure_kind_cluster() {
  local cluster
  cluster="$1"
  if ! kind get clusters | grep -q "$cluster"; then
    current_cluster="$(kubectl config current-context)" || true
    kind create cluster --name "$cluster" --wait 5m
    if [[ -n "$current_cluster" ]]; then
      kubectl config use-context "$current_cluster"
    fi
  fi
  kind export kubeconfig --name "$cluster" --kubeconfig "$HOME/.kube/$cluster.yml"
}

ensure_kind_cluster "$CLUSTER"
clusterctl init --kubeconfig "$HOME/.kube/$CLUSTER.yml"

"${REPO_ROOT}/bin/kustomize" build "${REPO_ROOT}/config/crd" | kubectl apply -f

ginkgo -p -r -randomizeAllSpecs --randomizeSuites tests/integration
