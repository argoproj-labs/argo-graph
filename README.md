# Argo Graph

POC for graphing resources across clusters.

## Proposal

* [Proposal](https://docs.google.com/document/d/15H09bsRvdyIAPSUjsnlmqzu-z3N9ZfLnYb-CEJ8uBaM/edit)

## Usage

Create at least two clusters, one to run in, one or more to monitor,

```
k3d create --name other --api-port 7443
```

```
make start
```

```
kubectl delete pod -l argoproj.io/vertex
kubectl delete svc -l argoproj.io/vertex
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