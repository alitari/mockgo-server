# clean matches
DELETE http://{{mockgo_host}}/__/matches/{{endpoint}}
[BasicAuth]
mockgo: {{mockgo_password}}

HTTP/1.1 200
[Asserts]
bytes count == 0

# clean mismatches
DELETE http://{{mockgo_host}}/__/mismatches
[BasicAuth]
mockgo: {{mockgo_password}}

HTTP/1.1 200
[Asserts]
bytes count == 0
