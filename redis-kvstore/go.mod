module github.com/alitari/mockgo-server/redis-kvstore

go 1.20

require (
	github.com/alicebob/miniredis/v2 v2.30.2
	github.com/go-redis/redismock/v9 v9.0.3
	github.com/redis/go-redis/v9 v9.0.3
	github.com/stretchr/testify v1.8.1
)

require (
	github.com/alicebob/gopher-json v0.0.0-20200520072559-a9ecdc9d1d3a // indirect
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/yuin/gopher-lua v1.1.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/alitari/mockgo-server/mockgo => ../mockgo