MOCKGO_MODULE = grpc-kvstore
PROTO_CONTEXT = kvstore

include ../main.mk

.PHONY: env-module
env-module: env-lib

.PHONY: clean
clean: cleancover cleanproto

.PHONY: build
build: clean gen-proto buildarchive

.PHONY: mod-dev
mod-dev: require-dep-mockgo-dev
