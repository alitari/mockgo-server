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
      statusCode: 204 # [MANDATORY], http response code ( see RFC 7231)
      body: "hello" # [OPTIONAL], response body as string
      bodyFilename: "response.json" # [OPTIONAL], refers to a file which contains the response body
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

## let's get started with...

### binary excutable on your machine

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






