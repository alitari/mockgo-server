MOCKGO_MODULE = mockgo-redis
MOCKGO_VARIANT = redis

include ../main.mk

.PHONY: env-module
env-module: env-exe

.PHONY: clean
clean: helm-delete cleanexe cleancover

.PHONY: build
build: buildexe buildchecksum

.PHONY: run
run: buildexe runexe

.PHONY: dep-dev
dep-dev: require-dep-mockgo-dev require-dep-redis-kvstore-dev require-dep-redis-matchstore-dev

.PHONY: dep-release
dep-release: require-dep-mockgo-release require-dep-redis-kvstore-release require-dep-redis-matchstore-release


