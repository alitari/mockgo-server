MOCKGO_MODULE = redis-kvstore

include ../main.mk

.PHONY: env-module
env-module: env-lib

.PHONY: clean
clean: cleancover

.PHONY: build
build: buildarchive

.PHONY: dep-dev
dep-dev: require-dep-mockgo-dev

.PHONY: dep-release
dep-release: require-dep-mockgo-release
