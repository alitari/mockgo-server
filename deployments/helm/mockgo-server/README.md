# mockgo-server helm charts

## add repo

```bash
helm repo add mockgo-server https://alitari.github.io/mockgo-server/
```

## configuration

see [`values.yaml`](./values.yaml)

## examples

```bash
# default standalone
helm upgrade mymock  mockgo-server/mockgo-server --install

# cluster 
helm upgrade mymockcluster mockgo-server/mockgo-server --set image=alitari/mockgo-grpc:latest,cluster.enabled=true,cluster.replicas=3 --install
```

