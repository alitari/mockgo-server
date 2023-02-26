MOCKGO_MODULE = mockgo-grpc
MOCKGO_VARIANT = grpc

include ../main.mk

.PHONY: env-module
env-module: env-exe

.PHONY: clean
clean: cleanexe

.PHONY: build
build: buildexe

.PHONY: run
run: buildexe runexe

.PHONY: mod-dev
mod-dev: require-dep-mockgo-dev require-dep-grpc-kvstore-dev require-dep-grpc-matchstore-dev