# health
GET http://{{mockgo_host}}/__/health

HTTP/1.1 200
[Asserts]
bytes count == 0

# hello1
GET http://{{mockgo_host}}/hello1

HTTP/1.1 200
[Asserts]
header "Content-Type" exists
header "Content-Type" contains "text/plain"
jsonpath "$.hello" == "World"

# hello2
GET http://{{mockgo_host}}/hello2

HTTP/1.1 200
[Asserts]
header "Content-Type" exists
header "Content-Type" contains "application/text"
header "endpoint-id" == "hello2"
bytes count == 12

# hello3
GET http://{{mockgo_host}}/hello3/alex

HTTP/1.1 200
[Asserts]
header "Content-Type" exists
header "Content-Type" contains "application/text"
header "endpoint-id" == "hello3"
bytes count == 20

# hello4
GET http://{{mockgo_host}}/hello4/go/joe

HTTP/1.1 200
[Asserts]
header "Content-Type" exists
header "Content-Type" contains "application/json"
header "endpoint-id" == "hello4"
jsonpath "$.message" == "Hello world, joe !"

# hello5
GET http://{{mockgo_host}}/hello5/bill/foo

HTTP/1.1 200
[Asserts]
header "Content-Type" exists
header "Content-Type" contains "application/json"
header "endpoint-id" == "hello5"
jsonpath "$.message" == "Hello world, bill !"

# hello6
GET http://{{mockgo_host}}/hello6/bill/sends/joe

HTTP/1.1 200
[Asserts]
header "Content-Type" exists
header "Content-Type" contains "application/json"
header "endpoint-id" == "hello6"
jsonpath "$.bill" == "says 'hello' to joe !"

# hello7
GET http://{{mockgo_host}}/hello7/alex/dani/foo/gdd

HTTP/1.1 200
[Asserts]
header "Content-Type" exists
header "Content-Type" contains "application/json"
header "endpoint-id" == "hello7"
jsonpath "$.alex" == "says 'hello' to dani !"

# hello8
GET http://{{mockgo_host}}/hello8/it/does/not/matter/last

HTTP/1.1 202
[Asserts]
header "endpoint-id" == "hello8"
bytes count == 0

# hello6
GET http://{{mockgo_host}}/hello9/some/stuff/sender/alex/additional/stuff/receiver/dani

HTTP/1.1 200
[Asserts]
header "Content-Type" exists
header "Content-Type" contains "application/json"
header "endpoint-id" == "hello9"
jsonpath "$.alex" == "says 'hello' to dani !"

