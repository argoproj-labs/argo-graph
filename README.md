# Argo Graph

POC for graphing resources across clusters.

## Proposal

* [Proposal](https://docs.google.com/document/d/15H09bsRvdyIAPSUjsnlmqzu-z3N9ZfLnYb-CEJ8uBaM/edit)

## Usage

Create at least two clusters, one to run in, one or more to monitor,

```
k3d create --name other --api-port 7443
export KUBECONFIG=$(k3d get-kubeconfig):$(k3d get-kubeconfig --name other)
kubectx ;# should list two or more clusters
```

In your default cluster:

```
kubectl create ns argo-graph
kubens argo-graph
go run ./cmd cluster add other
kubectl get secret clusters -o yaml ;# should show your cluster
```

Start database

```
docker run --rm -it -p 7000:7000 -p 7080:7080 -p 9080:9080 dgraph/standalone:latest
```

Start server:

```
make start
```

In a new terminal:

```
export KUBECONFIG=$(k3d get-kubeconfig):$(k3d get-kubeconfig --name other)
kubectx other
kubens default
kubectl delete pod -l graph.argoproj.io/vertex
kubectl delete cm -l graph.argoproj.io/vertex
kubectl apply -f examples/hello-world.yaml
```

Open localhost:5678.

## Developer

You probably want to run the UI:

```
yarn --cwd ui start
```

Open localhost:8080.

## Resources

* [Dagre](https://github.com/dagrejs/dagre/wiki)