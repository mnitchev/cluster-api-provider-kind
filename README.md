# Cluster API Provider Kind

Kubernetes-native declarative infrastructure for Kind. This is not a full implementation of the cluster-api specification and only supports a few Kind Config features. You can see what's supported in the [KindCluster spec](api/v1alpha3/kindcluster_types.go).

More information on implementing providers can be found in the [cluster-api book](https://cluster-api.sigs.k8s.io/user/concepts.html#infrastructure-provider).

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

## Presentation

Nodejs is required for the diagram generation used in the presentation. To install the npm package run:

```shell
npm install --global cli-diagram
```

Download the [slides](https://github.com/maaslalani/slides#installation) presentation tool and run (should be run from the repo root):

```shell
slides presentation/slides.md
```
## Known issues

Occasionally clusters will fail to create, but still be listable with kind. This is due to [this bug](https://github.com/kubernetes-sigs/kind/issues/2530).
