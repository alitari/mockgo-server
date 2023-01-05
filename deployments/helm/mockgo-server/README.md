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
helm upgrade mockgo mockgo-server/mockgo-server --install

# cluster 
helm upgrade mockgocluster mockgo-server/mockgo-server --set cluster.enabled=true,cluster.replicas=3 --install
```

