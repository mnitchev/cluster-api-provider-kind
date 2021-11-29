# Cluster API Provider Kind

Kubernetes-native declarative infrastructure for Kind.

## Installation
First install [kind](https://kind.sigs.k8s.io/docs/user/quick-start/#installation) and [clusterctl](https://cluster-api.sigs.k8s.io/user/quick-start.html#install-clusterctl).
To install a management cluster named `management-cluster` simply run:
```shell
make deploy-management-cluster
```
If you wish to name your cluster something different export or add the `CLUSTER` environment variable.
To create a cluster apply (optionally modify) the `tests/assets/clusters.yaml` file:
```shell
kubectl apply -f tests/assets/clusters.yaml
```
