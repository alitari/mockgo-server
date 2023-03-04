# matches delete
DELETE http://{{mockgo_host}}/__/matches/hello1
[BasicAuth]
mockgo: asecretPassword
[Options]
retry: true


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

# matches count
GET http://{{mockgo_host}}/__/matchesCount/hello1
Accept: application/json
[BasicAuth]
mockgo: asecretPassword

HTTP/1.1 200
[Asserts]
header "Content-Type" exists
header "Content-Type" == "text/plain; charset=utf-8"
body == "1"

# clean mismatches
DELETE http://{{mockgo_host}}/__/mismatches
[BasicAuth]
mockgo: asecretPassword

HTTP/1.1 200
[Asserts]
bytes count == 0

# not found request
GET http://{{mockgo_host}}/doesnotexist

HTTP/1.1 404
[Asserts]
header "Content-Type" exists
header "Content-Type" contains "text/plain"

# mismatches count
GET http://{{mockgo_host}}/__/mismatchesCount
Accept: application/json
[BasicAuth]
mockgo: asecretPassword

HTTP/1.1 200
[Asserts]
body == "1"