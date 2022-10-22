# mockgo-server

*mockgo-server* is a lightweight http server which can be used to mock http endpoints. *mockgo-server* is designed for horizontal scaling and feels at home in cloud environments like kubernetes.

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

| Variable | type |
| -------- | ---- |
| `RequestPathParams` | `   map[string]string`|
| `RequestUrl` | `          string`|
| `RequestUser` | `         string`|
| `RequestPath` | `         string`|
| `RequestHost` | `         string`|
| `RequestBody` | `         string`|
| `RequestBodyJsonData` | ` map[string]interface{}`|

The [Sprig library](http://masterminds.github.io/sprig/) for useful functions is available. In order to manage a state you can use a **key-value store** :
| Function name | signature | description |
| ------------- | --------- | ----------- |
| `kvStoreGet` | func(key string) interface{} | get the value stored under this key |
| `kvStorePut` | func(key string, value string) string | store a value with this key |
| `kvStoreAdd` | func(key, path, value string) string | modify a value with an [json patch](https://jsonpatch.com/) "add" operation |
| `kvStoreRemove` | func(key, path string) string | modify a value with an [json patch](https://jsonpatch.com/) "remove" operation |
| `kvStoreLookup` | func(key, jsonPath string) interface{} | get a value with a [json path](https://goessner.net/articles/JsonPath/) expression |

### examples

#### templating request attributes

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

#### kvstore and sprig

Create a file ( e.g. `people-mock.yaml` with this content:

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

## mockgo-server api

[TODO]

## let's get started with...

### downloading binary excutable

```bash
MOCKGO_RELEASE_VERSION="v0.0.3"
MOCKGO_VARIANT="standalone" # or "grpc"
MOCKGO_OS="linux" # or "windows"
MOCKGO_ARCH="amd64" # or "arm64"
MOCKGO_NAME=mockgo-${MOCKGO_VARIANT}-${MOCKGO_OS}-${MOCKGO_ARCH}
wget https://github.com/alitari/mockgo-server/releases/download/${MOCKGO_RELEASE_VERSION}/${MOCKGO_NAME}.tgz
tar xvf ${MOCKGO_NAME}.tgz
# MOCK_DIR is the path for looking for mockfiles, per default this are files with names matching "*-mock.*"
MOCK_DIR=$(pwd)/test/main ./bin/${MOCKGO_NAME} 
```

### building binary executable

```bash
git clone https://github.com/alitari/mockgo-server.git
cd mockgo-server
# go version must be >= 1.19
go version
MOCKGO_VARIANT="standalone" # or "grpc"
MOCKGO_OS="linux" # or "windows"
MOCKGO_ARCH="amd64" # or "arm64"
MOCKGO_NAME=mockgo-${MOCKGO_VARIANT}-${MOCKGO_OS}-${MOCKGO_ARCH}
./scripts/go-build-mockgo-${MOCKGO_VARIANT}.sh $MOCKGO_OS $MOCKGO_ARCH
# MOCK_DIR is the path for looking for mockfiles, per default this are files with names matching "*-mock.*"
MOCK_DIR=$(pwd)/test/main ./bin/${MOCKGO_NAME}
```

### with docker

```bash
MOCKGO_RELEASE_VERSION="v0.0.3"
MOCKGO_VARIANT="standalone" # or "grpc"
MOCK_DIR=$(pwd)/test/main # MOCK_DIR is the path for looking for mockfiles, per default this are files with names matching "*-mock.*"
docker run -it -v ${MOCK_DIR}:/mocks -e MOCK_DIR=/mocks alitari/mockgo-${MOCKGO_VARIANT}:$MOCKGO_RELEASE_VERSION 
```

### on k8s with helm

```bash
helm repo add mockgo-server https://alitari.github.io/mockgo-server/
helm upgrade mymock  mockgo-server/mockgo-server --install
```
see [here](./deployments/helm/mockgo-server/README.md) for further helm configuration options.

