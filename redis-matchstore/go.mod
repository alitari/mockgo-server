module github.com/alitari/mockgo-server/redis-matchstore

go 1.20

require (
	github.com/alicebob/miniredis/v2 v2.30.2
	github.com/alitari/mockgo-server/mockgo v0.0.0-00010101000000-000000000000
	github.com/go-redis/redismock/v9 v9.0.3
	github.com/google/uuid v1.3.0
	github.com/redis/go-redis/v9 v9.0.3
	github.com/stretchr/testify v1.8.1
)

require (
	github.com/alicebob/gopher-json v0.0.0-20200520072559-a9ecdc9d1d3a // indirect
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/go-http-utils/headers v0.0.0-20181008091004-fed159eddc2a // indirect
	github.com/gorilla/mux v1.8.0 // indirect
	github.com/kr/pretty v0.1.0 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/yuin/gopher-lua v1.1.0 // indirect
	gopkg.in/check.v1 v1.0.0-20190902080502-41f04d3bba15 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/alitari/mockgo-server/mockgo => ../mockgo
