# mockgo-server

[![Build](https://github.com/alitari/mockgo-server/workflows/OnMaster/badge.svg?branch=master)](https://github.com/alitari/mockgo-server/actions?workflow=OnMaster&branch=master)
[![Total Downloads](https://img.shields.io/github/downloads/alitari/mockgo-server/total.svg)](https://github.com/alitari/mockgo-server/releases)
[![DockerHub Pulls](https://img.shields.io/docker/pulls/alitari/mockgo-standalone.svg)](https://hub.docker.com/r/alitari/mockgo-standalone)
[![DockerHub Pulls](https://img.shields.io/docker/pulls/alitari/mockgo-grpc.svg)](https://hub.docker.com/r/alitari/mockgo-grpc)
[![codebeat badge](https://codebeat.co/badges/48cb2798-51ac-4ae3-8d9d-c41087c43c13)](https://codebeat.co/projects/github-com-alitari-mockgo-server-master)
[![CII Best Practices](https://bestpractices.coreinfrastructure.org/projects/6619/badge)](https://bestpractices.coreinfrastructure.org/projects/6619)


*mockgo-server* is a lightweight http server which can be used to mock http endpoints. *mockgo-server* is designed for horizontal scaling and feels at home in cloud environments like [kubernetes](https://kubernetes.io/). The main software design principles are:

- **Simplicity** : easy configuration with human readable yaml files with reasonable defaults
- **Scalability** : mockgo-server is designed for horizontal scaling. Therefore it can be used in environments with high http traffic ( e.g. for performance/load tests )
- **Flexibility** : sometimes requirements for mocking requests goes beyond static responses. With templating mechanism and state management it is possible to implement particular logic for building responses.


## install

Kubernetes is the first class environment for *mockgo-server*. So, if you have kubernetes cluster available, you can start right away with [helm](https://helm.sh/) charts:

```bash
helm repo add mockgo-server https://alitari.github.io/mockgo-server/
helm upgrade mymock mockgo-server/mockgo-server --install
```
see [here](./deployments/helm/mockgo-server/README.md) for further helm configuration options.

If you prefer installing on a/your workstation, define first the environment :

```bash
MOCKGO_RELEASE_VERSION="v0.0.5"
MOCKGO_VARIANT="standalone" # or "grpc"
MOCKGO_OS="linux" # or "windows"
MOCKGO_ARCH="amd64" # or "arm64"
MOCKGO_NAME=mockgo-${MOCKGO_VARIANT}-${MOCKGO_OS}-${MOCKGO_ARCH}
```

Now, you have the following options:

### downloading binary excutable

```bash
wget https://github.com/alitari/mockgo-server/releases/download/${MOCKGO_RELEASE_VERSION}/${MOCKGO_NAME}.tgz
tar xvf ${MOCKGO_NAME}.tgz
# MOCK_DIR is the path for looking for mockfiles, per default this are files with names matching "*-mock.*"
MOCK_DIR=$(pwd) ./bin/${MOCKGO_NAME} 
```

### building binary executable

```bash
git clone https://github.com/alitari/mockgo-server.git -b $MOCKGO_RELEASE_VERSION
cd mockgo-server
# go version must be >= 1.19
go version
./scripts/go-build-mockgo-${MOCKGO_VARIANT}.sh $MOCKGO_OS $MOCKGO_ARCH
# MOCK_DIR is the path for looking for mockfiles, per default this are files with names matching "*-mock.*"
MOCK_DIR=$(pwd) ./bin/${MOCKGO_NAME}
```

### run with docker

```bash
MOCK_DIR=$(pwd) # MOCK_DIR is the path for looking for mockfiles, per default this are files with names matching "*-mock.*"
docker run -it -v ${MOCK_DIR}:/mocks -e MOCK_DIR=/mocks alitari/mockgo-${MOCKGO_VARIANT}:$MOCKGO_RELEASE_VERSION 
```

## mockfiles and endpoints

A mockgo-server configuration consist of one or multiple files in yaml format, the so called *mockfiles*. Each *mockfile* contains one or multiple *endpoints*. The *endpoint* defines criterias which qualify the endpoint to serve an incoming request. This process is called *matching*. The second configuration part of an *endpoint* is the defintion of the http response. 

```yaml
endpoints:
  - id: "id" # [OPTIONAL] unique string to identify endpoint
    prio: 1  # [OPTIONAL] integer to define precedence of endpoints if more than one endpoint matches
    request: # request defines the matching
      scheme: "https" # [OPTIONAL], match to http schema, possible values: "http" or "https"
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
      body: "hello" # [OPTIONAL], response body as string
      bodyFilename: "response.json" # [OPTIONAL], refers to a file which contains the response body
      headers: | # [OPTIONAL],multiline string in form of key: value, templates can be used
        Content-Type: "application/text"
```

### example

```bash
# create a mockfile
cat <<EOF > minimal-mock.yaml
endpoints:
- request:
    path: "/minimal"
  response:
    statusCode: 204
EOF
# per default mockgo-server looks in the current dir for files with names matching "*-mock.*"
./bin/mockgo-standalone-linux-amd64

# match
curl http://localhost:8081/minimal
# no match 404
curl http://localhost:8081/wrong
```

## path matching

The form of a http request can be described as a sequence of *pathsegments* which are separated through a `/`. In order to achieve flexibility matching the path of an http request, there are 3 special symbols which can be used for the *pathsegment* defining the `request.path`:

1. `*` every *pathsegment* on this position matches
2. `{pathvariable}` every *pathsegment* on this position matches, the value of the *pathsegment* is available as template variable with name `pathvariable`
3. `**` one or more arbitrary *pathsegments* match on this position

### examples

| `request.path` | incoming request path | match |
| -------------- | --------------------- | ----- |
| `/foo/*` | `/bar` | ➖ |
| `/foo/*` | `/foo` | ➖ |
| `/foo/*` | `/foo/bar` | ✅ |
| `/foo/*` | `/foo/bar/foo` | ➖ |
| `/foo/*/bar` | `/foo/bar` | ➖ |
| `/foo/*/bar` | `/foo/bar/bar` | ✅ |
| `/foo/{foo}/bar` | `/foo/bar/bar` | ✅ `{{ RequestPathParams.foo }}` resolves to `bar` |
| `/foo/*/{bar}` | `/foo/bar/bar/foo` | ➖ |
| `/foo/*/{bar}` | `/foo/bar/bar` | ✅ `{{ RequestPathParams.bar }}` resolves to `bar` |
| `/foo/**` | `/foo/` | ➖ |
| `/foo/**` | `/foo/bar` | ✅ |
| `/foo/**` | `/foo/bar/1/2/3` | ✅ |
| `/foo/**/{bar}` | `/foo/bar/1` | ✅ matches but `{{ RequestPathParams.bar }}` does not resolve |
| `/foo/**/foo/{bar}` | `/foo/bar/foo` | ➖ |
| `/foo/**/foo/{bar}` | `/foo/bar/1/2/foo/3` | ✅ `{{ RequestPathParams.bar }}` resolves to `3` |

## creating dynamic responses with go templates

[*go templates*](https://blog.gopheracademy.com/advent-2017/using-go-templates/) can be used for creating dynamic responses. You can refer to following attributes of the incoming request:

| variable | type |
| -------- | ---- |
| `RequestPathParams` | `   map[string]string`|
| `RequestUrl` | `          string`|
| `RequestUser` | `         string`|
| `RequestPath` | `         string`|
| `RequestHost` | `         string`|
| `RequestBody` | `         string`|
| `RequestBodyJsonData` | ` map[string]interface{}`|

### example

```bash
cat <<EOF > template-mock.yaml
endpoints:
  - id: "statusCode"
    request:
      method: "POST"
      path: /statusCode/{statusCode}
    response:
      statusCode: "{{.RequestPathParams.statusCode }}"
      body: |-
        {{ .RequestBody -}}
      headers: |
        Header1: "{{ .RequestUrl }}"
EOF
# per default mockgo-server looks in the current dir for files with names matching "*-mock.*"
./bin/mockgo-standalone-linux-amd64
# ok response
curl -v http://localhost:8081/statusCode/200 -d "Alex"
# internal server error
curl -v http://localhost:8081/statusCode/500 -d "An error"
```

In order to use a state when creating responses, you have access to a **key-value store** in your template:

| function name | signature | description |
| ------------- | --------- | ----------- |
| `kvStoreGet` | `func(key string) interface{}` | get the value stored under this key |
| `kvStorePut` | `func(key string, value string) string` | store a value with this key |
| `kvStoreAdd` | `func(key, path, value string) string` | modify a value with an [json patch](https://jsonpatch.com/) "add" operation |
| `kvStoreRemove` | `func(key, path string) string` | modify a value with an [json patch](https://jsonpatch.com/) "remove" operation |
| `kvStoreLookup` | `func(key, jsonPath string) interface{}` | get a value with a [json path](https://goessner.net/articles/JsonPath/) expression |

See extensive example, how to use it.


## mockgo-server api

The *mockgo-server* stores 2 kinds of state. The first one is the storage of incoming requests. The second state is the build-in *key-value store* which can be utilized for creating dynamic responses. Both states can be checked and changed through an [REST](https://en.wikipedia.org/wiki/Representational_state_transfer) api. The api is secured with basic auth which can be configured with the environment variables `API_USERNAME` and `API_PASSWORD`.

### matching api

The request storage has a limited capacity which can be configured with `MATCHES_CAPACITY`.

| method | path  | description |
| ------ | ----- | ----------- |
| `GET` | `/matches/{endpointId}` | returns all requests which matched to an endpoint |
| `GET` | `/matchesCount/{endpointId}` | returns the count of requests which matched to an endpoint, is not limited through capacity |
| `GET` | `/mismatches` | returns all requests which didn't match to an endpoint |
| `GET` | `/mismatchesCount` | returns the count of all requests which didn't match to an endpoint, is not limited through capacity |
| `DELETE` | `/matches/{endpointId}` | deletes storage of all requests which matched to an endpoint |
| `DELETE` | `/mismatches` | deletes storage of all requests which didn't match to an endpoint |

### key-value store api

| method | path  | description |
| ------ | ----- | ----------- |
| `PUT`  | `/kvstore/{key}`| store content of request body under a key |
| `GET` | `/kvstore/{key}` | get content of key |
| `POST` | `/kvstore/{key}/add` | add content to kvstore with a [json patch](https://jsonpatch.com/) "add" operation. The json format of the request payload is `{ "path": json path, "value": json value }` |
| `POST` | `/kvstore/{key}/remove` | remove content to kvstore with a [json patch](https://jsonpatch.com/) "remove" operation. The json format of the request payload is `{ "path": json path }` |


## An extensive example

Create a file ( e.g. `people-mock.yaml`) with this content:

```yaml
endpoints:
  - id: "addPeople"
    request:
      method: "PUT"
      path: '/storePeople'
    response:
      statusCode: |-
        {{ $payload := .RequestBodyJsonData -}}
        {{ if and $payload.name $payload.age -}}
          200
        {{- else -}}
          400
        {{- end -}}
      body: |-
        {{ $payload := .RequestBodyJsonData -}}
        {{ if and $payload.name $payload.age -}}
          {{ if gt ( int $payload.age) 17 -}}
            {{ kvStoreAdd "people" "/adults/-" .RequestBody -}}
            stored '{{ .RequestBody }}'' as adult
          {{- else -}}
            {{ kvStoreAdd "people" "/childs/-" .RequestBody -}}
            stored '{{ .RequestBody }}' as child
          {{- end -}}
        {{- end -}}
  - id: "getPeople"
    request:
      method: "GET"
      path: '/getPeople/{category}'
    response:
      statusCode: |-
        {{ $category := .RequestPathParams.category -}}
        {{ if or ( eq $category "childs") ( eq $category "adults") -}}
          200
        {{- else -}}
          400
        {{- end -}}
      body: |-
        {{ $category := .RequestPathParams.category -}}
        {{ if or ( eq $category "childs") ( eq $category "adults") -}}
          {{ kvStoreLookup "people" ( printf "$.%s" $category ) | toPrettyJson -}}
        {{- end -}}
```

```bash
# per default mockgo-server looks in the current dir for files with names matching "*-mock.*"
./bin/mockgo-standalone-linux-amd64

# setup kvstore with api
curl -v -u mockgo:password -H "Content-Type: application/json" -X PUT http://localhost:8081/__/kvstore/people -d '{ "adults": [], "childs": [] }'

# payload must have right attributes, -> bad request
curl -v -X PUT http://localhost:8081/storePeople -d '{ "name": "Dani" }'

# store some adults
curl -v -X PUT http://localhost:8081/storePeople -d '{ "name": "Alex", "age": 55 }'
curl -v -X PUT http://localhost:8081/storePeople -d '{ "name": "Dani", "age": 45 }'

# store some childs
curl -v -X PUT http://localhost:8081/storePeople -d '{ "name": "Klara", "age": 16 }'

# get adults
curl -v http://localhost:8081/getPeople/adults

# get childs
curl -v http://localhost:8081/getPeople/childs
```

## contribute

Learn how to contribute [here](./contribute.md)
