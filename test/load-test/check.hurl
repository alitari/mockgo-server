# get matches
GET http://{{mockgo_host}}/__/matchesCount/{{endpoint}}
accept: application/json
[BasicAuth]
mockgo: {{mockgo_password}}

HTTP/1.1 200
[Asserts]
body contains "{{matches_count}}"

# clean mismatches
GET http://{{mockgo_host}}/__/mismatchesCount
accept: application/json
[BasicAuth]
mockgo: {{mockgo_password}}

HTTP/1.1 200
[Asserts]
body contains "{{mismatches_count}}"
