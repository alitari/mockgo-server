# mockgo-server

<img src="./assets/mockgo-icon.png" alt="drawing" width="200"/>

[![CII Best Practices](https://bestpractices.coreinfrastructure.org/projects/6619/badge)](https://bestpractices.coreinfrastructure.org/projects/6619)
[![Build](https://github.com/alitari/mockgo-server/workflows/Workflow/badge.svg?branch=master)](https://github.com/alitari/mockgo-server/actions?workflow=Workflow&branch=master)
[![Go Report Card](https://goreportcard.com/badge/github.com/alitari/mockgo-server/mockgo)](https://goreportcard.com/report/github.com/alitari/mockgo-server/mockgo)
[![codecov](https://codecov.io/gh/alitari/mockgo-server/branch/master/graph/badge.svg?token=WRJ9KFCJFI)](https://codecov.io/gh/alitari/mockgo-server)
[![Go Reference](https://pkg.go.dev/badge/github.com/alitari/mockgo-server/mockgo.svg)](https://pkg.go.dev/github.com/alitari/mockgo-server/mockgo)

![Docker Image Size (latest semver)](https://img.shields.io/docker/image-size/alitari/mockgo-standalone?label=image%20mockgo-standalone)
![Docker Image Size (latest semver)](https://img.shields.io/docker/image-size/alitari/mockgo-grpc?label=image%20mockgo-grpc)



*mockgo-server* is a lightweight http server which can be used to mock http endpoints. *mockgo-server* is designed for horizontal scaling and feels at home in cloud environments like [kubernetes](https://kubernetes.io/) and [knative](https://knative.dev/). The main features are:

- **Simplicity** : easy configuration with human readable yaml files with reasonable defaults
- **Scalability** : mockgo-server is designed for horizontal scaling. Therefore it can be used in environments with high http traffic ( e.g. for performance/load tests )
- **Flexibility** : sometimes requirements for mocking requests goes beyond static responses. With templating mechanism and state management it is possible to implement particular logic for building responses.

## TL;DR

See [Examples.md](./Examples.md).

## variants

*mockgo-server* is built with different variants which have their own use cases. The following table gives an overview:

| variant             | can scale | persistence | use this when you ...                                                                         |
|---------------------|-----------|-------------|-----------------------------------------------------------------------------------------------|
| `mockgo-standalone` | no        | no          | want a simple setup and you don't need to scale, usually this variant is the starting point   |
| `mockgo-grpc`       | yes       | no          | have to deal with high incoming traffic, but you don't need durability for states and results |
| `mockgo-redis`      | yes       | yes         | have to deal with high incoming traffic and you want a durable storage for states and results or if you want to use mockgo as serverless deployment in knative |


## install on kubernetes

Kubernetes is the first class environment for *mockgo-server*. So, if you have kubernetes cluster available, you can start right away with [helm](https://helm.sh/) charts:

```bash
helm repo add mockgo-server https://alitari.github.io/mockgo-server/
helm upgrade mymock mockgo-server/mockgo-server --install
```
see [here](./deployments/helm/mockgo-server/README.md) for further helm configuration options.

## install on knative

If you have a [knative](https://knative.dev/) cluster available, you can start right away with [kn](https://knative.dev/docs/client/install-kn/)

```bash
cat <<EOF > kn-mock.yaml
endpoints:
- request:
    path: "/hello-kn"
  response:
    statusCode: 200
    body: "Hello from knative"
EOF
kubectl create configmap mockgo-config --from-file=kn-mock.yaml
kn service create mockgo --image=alitari/mockgo-standalone:v1.3.0 --mount /mockdir=cm:mockgo-config --env MOCK_DIR=/mockdir --env MOCK_PORT=8080
MOCKGO_URL=$(kn service describe mockgo -o url)
curl -v $MOCKGO_URL/hello-kn
```

See [here](./knative.md) for a example with redis.


## install on local machine

Define the environment :

```bash
MOCKGO_RELEASE_VERSION="v1.3.0"
MOCKGO_VARIANT="standalone" # or "grpc"
MOCKGO_OS="linux" # or "windows"
MOCKGO_ARCH="amd64" # or "arm64"
MOCKGO_NAME=mockgo-${MOCKGO_VARIANT}-${MOCKGO_OS}-${MOCKGO_ARCH}
```

### option 1: download binary excutable

```bash
wget https://github.com/alitari/mockgo-server/releases/download/$MOCKGO_RELEASE_VERSION/$MOCKGO_NAME
chmod +x $MOCKGO_NAME
# MOCK_DIR is the path for looking for mockfiles, per default this are files with names matching "*-mock.*"
MOCK_DIR=$(pwd)/test/mocks ./$MOCKGO_NAME
```

### option 2: go install

```bash
go install github.com/alitari/mockgo-server/mockgo-$MOCKGO_VARIANT/starter/mockgo@latest
MOCK_DIR=$(pwd)/test/mocks mockgo
```

### option 3: run with docker

```bash
MOCK_DIR=$(pwd)/test/mocks
docker run -it -v $MOCK_DIR:/mocks -e MOCK_DIR=/mocks alitari/mockgo-$MOCKGO_VARIANT:$MOCKGO_RELEASE_VERSION 
```

## mockfiles and endpoints

A mockgo-server configuration consist of one or multiple files in yaml format, the so called *mockfiles*.
Each *mockfile* contains one or multiple *endpoints*.
The *endpoint* defines criteria which qualify the endpoint to serve an incoming request. This process is called *matching*.
The second configuration part of an *endpoint* is the definition of the http response. 

```yaml
endpoints:
  - id: "id" # [OPTIONAL] unique string to identify endpoint
    prio: 1  # [OPTIONAL] integer to define precedence of endpoints if more than one endpoint matches
    request: # request defines the matching
      host: "alexkrieg.com" # [OPTIONAL], match to http host
      method: "POST" # [OPTIONAL], match to http method, default is "GET"
      path: "/mypath" # [MANDATORY], match to http request path
      query: # [OPTIONAL] for matching, every key value pair must be part of the http query parameters of the incoming request
        firstQueryParam: value1
        secondQueryParam: value2
      headers: # [OPTIONAL] for matching, every key value pair must be part of the http header values of the incoming request
        Content-Type: "application/json"
        Myheader: myheaderValue
      body: "^{.*}$" # [OPTIONAL] regular expression which match to the request body
    response: # defines the response
      statusCode: 204 # [OPTIONAL], http response code ( see RFC 7231), defaults to "200"
      body: "hello" # [OPTIONAL], response body as string, templates can be used
      bodyFilename: "response.json" # [OPTIONAL], refers to a file which contains the response body, templates can be used in the file
      headers: | # [OPTIONAL],multiline string in form of key: value, templates can be used
        Content-Type: "application/text"
```

## path matching

The form of a http request can be described as a sequence of *pathsegments* which are separated through a `/`. In order to achieve flexibility matching the path of an http request, there are 3 special symbols which can be used for the *pathsegment* defining the `request.path`:

1. `*` every *pathsegment* on this position matches
2. `{pathvariable}` every *pathsegment* on this position matches, the value of the *pathsegment* is available as template variable with name `pathvariable`
3. `**` one or more arbitrary *pathsegments* match on this position

### examples

| `request.path`      | incoming request path | match                                                        |
|---------------------|-----------------------|--------------------------------------------------------------|
| `/foo/*`            | `/bar`                | ➖                                                            |
| `/foo/*`            | `/foo`                | ➖                                                            |
| `/foo/*`            | `/foo/bar`            | ✅                                                            |
| `/foo/*`            | `/foo/bar/foo`        | ➖                                                            |
| `/foo/*/bar`        | `/foo/bar`            | ➖                                                            |
| `/foo/*/bar`        | `/foo/bar/bar`        | ✅                                                            |
| `/foo/{foo}/bar`    | `/foo/bar/bar`        | ✅ `{{ RequestPathParams.foo }}` resolves to `bar`            |
| `/foo/*/{bar}`      | `/foo/bar/bar/foo`    | ➖                                                            |
| `/foo/*/{bar}`      | `/foo/bar/bar`        | ✅ `{{ RequestPathParams.bar }}` resolves to `bar`            |
| `/foo/**`           | `/foo/`               | ➖                                                            |
| `/foo/**`           | `/foo/bar`            | ✅                                                            |
| `/foo/**`           | `/foo/bar/1/2/3`      | ✅                                                            |
| `/foo/**/{bar}`     | `/foo/bar/1`          | ✅ matches but `{{ RequestPathParams.bar }}` does not resolve |
| `/foo/**/foo/{bar}` | `/foo/bar/foo`        | ➖                                                            |
| `/foo/**/foo/{bar}` | `/foo/bar/1/2/foo/3`  | ✅ `{{ RequestPathParams.bar }}` resolves to `3`              |

## creating dynamic responses with go templates

In order to implement program logic, you can use [*go templates*](https://blog.gopheracademy.com/advent-2017/using-go-templates/) for creating dynamic responses.
The templating mechanism is available for the `response.body` and `response.headers` attributes and provides following additional functions for:

- common template functionality through the [sprig library](https://masterminds.github.io/sprig/)
- access to the incoming request through the `Request` variables
- using the key-value store

### access attributes of the incoming request

| template variable     | type                     |
|-----------------------|--------------------------|
| `RequestPathParams`   | `map[string]string`      |
| `RequestUrl`          | `string`                 |
| `RequestPath`         | `string`                 |
| `RequestHost`         | `string`                 |
| `RequestBody`         | `string`                 |
| `RequestBodyJSONData` | `map[string]interface{}` |

### access the key-value store

There are 2 parameters which identifies a value in the key-value store:
- `store` is a string which identifies the store
- `key` is a string which identifies the value in the store

| template function name | signature                                    | description                                                           |
|------------------------|----------------------------------------------|-----------------------------------------------------------------------|
| `kvStoreGet`           | `func(store, key string) interface{}`        | get the value from store `store`, which is stored under the key `key` |
| `kvStorePut`           | `func(store, key string, value interface{})` | store value `value` under the key `key` in the store `store`          |
| `kvStoreRemove`        | `func(store, key string)`                    | remove the value under the key `key` from store `store`               |

## mockgo-server api

The *mockgo-server* holds multiple kinds of state: 
- current configuration (mock endpoints, logging , etc.)
- storage of incoming requests
- key-value store

All states can be accessed through an [REST](https://en.wikipedia.org/wiki/Representational_state_transfer) api. 
The api is secured with basic auth which can be configured with the environment variables `API_USERNAME` and `API_PASSWORD`.
In order to avoid conflicts with the mock endpoints, the api is exposed under the path prefix `/__`.
This can be changed with the environment variable `API_PATH_PREFIX`.

### configuration api

| method  | path         | description                             |
|---------|--------------|-----------------------------------------|
| `POST`  | `/__/reload` | reload the mock files from the mock dir |

### matching api

The request storage has a limited capacity which can be configured with `MATCHES_CAPACITY`.

| method   | path                            | description                                                                                          |
|----------|---------------------------------|------------------------------------------------------------------------------------------------------|
| `GET`    | `/__/matches/{endpointId}`      | returns all requests which matched to an endpoint                                                    |
| `GET`    | `/__/matchesCount/{endpointId}` | returns the count of requests which matched to an endpoint, is not limited through capacity          |
| `GET`    | `/__/mismatches`                | returns all requests which didn't match to an endpoint                                               |
| `GET`    | `/__/mismatchesCount`           | returns the count of all requests which didn't match to an endpoint, is not limited through capacity |
| `DELETE` | `/__/matches/{endpointId}`      | deletes storage of all requests which matched to an endpoint                                         |
| `DELETE` | `/__/mismatches`                | deletes storage of all requests which didn't match to an endpoint                                    |

### key-value store api

Using the path `/__/kvstore/{store}/{key}` you can access the key-value store with the following methods:

- `PUT` stores the content of the request body under the key `key` in the store `store`
- `GET` returns the content which is stored under the key `key` in the store `store`
- `DELETE` removes the content which is stored under the key `key` in the store `store`

## prometheus metrics

The *mockgo-server* exposes besides of the standard go metrics the following application specific counters:

- `matches{"endpoint":"<endpointId>"}`: Number of matches of an endpoint
- `mismatches`: Number of requests which did not match to on endpoint

## using config reload feature

For local development it is useful to have a way to reload the mock files without restarting the *mockgo-server*. This can be achieved by sending a `POST` request to the reload endpoint. This functionality can be combined with a file watcher to automatically reload the mock files when they change. The script `scripts/watchmocks.sh` implements this functionality. 

```bash
# watch for changes in test/main and reload the mock files
./scripts/watchmocks.sh test/main
```

## contribute

Learn how to contribute [here](./contribute.md)
