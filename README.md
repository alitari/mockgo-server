# mockgo-server

*mockgo-server* is a lightweight http server which can be used to mock http endpoints. *mockgo-server* is designed for horizontal scaling and feels at home in cloud environments like kubernetes.

## installation

### helm

```bash
helm repo add mockgo-server https://alitari.github.io/mockgo-server/
helm upgrade mymock  mockgo-server/mockgo-server --install
```

see [here](./deployments/helm/mockgo-server/README.md) for further helm configuration options.

### binaries

[TODO]

