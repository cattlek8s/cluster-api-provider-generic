# cluster-api-provider-generic

An experimental provider which uses webhooks as an extension mechanism for
allocating machine infrastructure while sharing software provisioning.

# Developer

## Build artifacts

```bash
export IMG=<CONTAINER_REPOSITORY_URL>
make docker-build
make docker-push
```

## Deploy controller

```bash
kind create cluster
export KUBECONFIG="$(kind get kubeconfig-path --name="1")"
make deploy
```

## Test controller

```bash
kubectl apply -f config/samples/cluster_v1alpha1_cluster.yaml
kubectl apply -f config/samples/cluster_v1alpha1_machine.yaml
kubectl logs -n cluster-api-provider-generic-system cluster-api-provider-generic-controller-manager-0 -c manager -f
```
