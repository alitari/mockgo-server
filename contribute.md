# contribute

## general

Development guides are applied on a linux amd64 system. Please understand that my support capatibilties are limited regarding other setups. Vscode was used as IDE, however this repo does not contain any relevant files regarding IDE usage.

## tools

As [go-workspaces](https://go.dev/blog/get-familiar-with-workspaces) are used development requires **go version 1.19++**. For kubernetes related work i recommend using [minikube](https://minikube.sigs.k8s.io/docs/), [helm](https://helm.sh/) and [skaffold](https://skaffold.dev/).

## grpc

As the [grpc]() variant is using [protobuf](https://developers.google.com/protocol-buffers), you need to install the protobuf compiler and its go tools:

```bash
sudo apt update -y
sudo apt install -y protobuf-compiler
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
```

## building testing 

For building and testing you can use the commands which are defined in the [github actions for a pull request](.github/workflows/onpr.yml).
E.g. build a the mockgo-server grpc variant for linux amd64: `./scripts/go-build-mockgo-grpc.sh linux amd64`

## releasing

### github and dockerhub

```bash
./scripts/github-create-release.sh <release tag> # semver is used e.g. ( "v0.0.1")
```
### helm

```bash
git clone https://github.com/alitari/mockgo-server.git -b gh-pages
cd mockgo-server
# define tag/branch with arg
./build.sh master
# push new helm chart tgz in gh-pages branch
```
