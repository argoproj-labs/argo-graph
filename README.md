# Argo Graph

**Argo Graph** is an open source container-native resource dependency analysis system for graphing Kubernetes resources and the relationships between them.

## How Does It Work

It watches for resources in one or more clusters labelled with `graph.argoproj.io/node`. If extracts the following annotations:

* `graph.argoproj.io/edges` - a comma-separated list of related nodes
* `graph.argoproj.io/label` - a label for the node

## Usage

In your  cluster:

```
kubectl create ns argo-graph
kubens argo-graph
go run ./cmd cluster add k3s-default
kubectl get secret clusters -o yaml ;# should show your cluster
```

Start database

```
docker run --rm -it -p 7000:8000 -p 7080:8080 -p 9080:9080 dgraph/standalone:latest
```

Start server:

```
make start
```

In a new terminal:

```
kubectl -n default delete pod -l graph.argoproj.io/node
kubectl -n default delete cm -l graph.argoproj.io/node
kubectl -n default apply -f examples/hello-world.yaml
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
