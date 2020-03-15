# Argo Graph

POC for graphing resources across clusters.

## Proposal

* [Proposal](https://docs.google.com/document/d/15H09bsRvdyIAPSUjsnlmqzu-z3N9ZfLnYb-CEJ8uBaM/edit)

## Usage

```
make start
```

```
cd ui
yarn start
```

```
kubectl delete pod -l argoproj.io/vertex
kubectl delete svc -l argoproj.io/vertex
kubectl apply -f examples/hello-world.yaml
```

Open localhost:8080.


## Resources

* [Dagre](https://github.com/dagrejs/dagre/wiki)