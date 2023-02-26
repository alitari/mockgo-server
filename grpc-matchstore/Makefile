MOCKGO_MODULE = grpc-matchstore
PROTO_CONTEXT = matchstore

include ../main.mk

.PHONY: env-module
env-module: env-lib

.PHONY: clean
clean: cleanexe cleancover cleanproto

.PHONY: build
build: clean gen-proto buildarchive

.PHONY: mod-dev
mod-dev: require-dep-mockgo-dev