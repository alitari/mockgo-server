# mockgo-gen

mockgo-gen is a tool to generate mock objects from [OpenAPI specifications](https://swagger.io/specification/).

## Installation

### go install

```bash
go install github.com/alitari/mockgo-server/mockgo-gen/cmd/mockgo-gen
```

### build from source

```bash
 make build MOCKGO_MODULE=mockgo-gen
```

## Usage

### from local file

```bash
# create mock file from OpenAPI spec
mockgo-gen --logLevel=2 --o test/mockgo-gen/apiSpec-mock.yaml test/mockgo-gen/apiSpec.yaml
```
### from url

```bash
# create mock file from OpenAPI spec
mockgo-gen --logLevel=2 --o test/mockgo-gen/petstore-mock.yaml https://petstore.swagger.io/v2/swagger.json
```


`
# run mock server watching for changes in the mock file
make build && ./scripts/watchmocks.sh test/mockgo-gen
```

