# mockgo-server

*mockgo-server* is a lightweight http server which can be used to mock http endpoints. *mockgo-server* is designed for horizontal scaling and feels at home in cloud environments like kubernetes.

## start

### binary excutable on your machine

```bash
MOCKGO_RELEASE_VERSION="v0.0.3"
MOCKGO_VARIANT="standalone" # or "grpc"
MOCKGO_OS="linux" # or "windows"
MOCKGO_ARCH="amd64" # or "arm64"
MOCKGO_NAME=mockgo-${MOCKGO_VARIANT}-${MOCKGO_OS}-${MOCKGO_ARCH}
wget https://github.com/alitari/mockgo-server/releases/download/${MOCKGO_RELEASE_VERSION}/${MOCKGO_NAME}.tgz
tar xvf ${MOCKGO_NAME}.tgz
# run set MOCK_DIR
MOCK_DIR=$(pwd)/test/main ./bin/${MOCKGO_NAME} 
```

### with docker

```bash
MOCKGO_RELEASE_VERSION="v0.0.3"
MOCKGO_VARIANT="standalone" # or "grpc"
MOCK_DIR=$(pwd)/test/main # path to you mock files
docker run -it -v ${MOCK_DIR}:/mocks -e MOCK_DIR=/mocks alitari/mockgo-${MOCKGO_VARIANT}:$MOCKGO_RELEASE_VERSION 
```

### on k8s with helm

```bash
helm repo add mockgo-server https://alitari.github.io/mockgo-server/
helm upgrade mymock  mockgo-server/mockgo-server --install
```
see [here](./deployments/helm/mockgo-server/README.md) for further helm configuration options.

## usage

### configuration model

