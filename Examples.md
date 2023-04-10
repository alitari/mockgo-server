# mockgo examples

The following examples assume that you have `mockgo-standalone` binary in your path.
See [install on local machine](./README.md#install-on-local-machine) for more details.

## a minimal template

```bash
mkdir minimal-example && cd minimal-example
# create a mockfile
cat <<EOF > minimal-mock.yaml
endpoints:
- request:
    path: "/minimal"
  response:
    statusCode: 204
EOF
# start server, per default mockgo-server looks in the current dir for files with names matching "*-mock.*"
mockgo-standalone

# in a new shell create requests 
curl http://localhost:8081/minimal
# no match 404
curl http://localhost:8081/wrong
```

## using request attributes

```bash
mkdir request-attributes-example && cd request-attributes-example
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
        Header1: "{{ .RequestURL }}"
EOF

# start server, per default mockgo-server looks in the current dir for files with names matching "*-mock.*"
mockgo-standalone

# ok response
curl -v http://localhost:8081/statusCode/200 -d "Alex"
# internal server error
curl -v http://localhost:8081/statusCode/500 -d "An error"
```
## using template functions and key-value store

```bash
mkdir template-functions-example && cd template-functions-example
cp ../test/main/people-mock.yaml .

# per default mockgo-server looks in the current dir for files with names matching "*-mock.*"
mockgo-standalone

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
curl -v http://localhost:8081/getPeople/children
```

