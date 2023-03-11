MOCKGO_MODULE = grpc-matchstore
PROTO_CONTEXT = matchstore

include ../main.mk

.PHONY: env-module
env-module: env-lib

.PHONY: clean
clean: cleanexe cleancover cleanproto

.PHONY: build
build: gen-proto buildarchive

.PHONY: dep-dev
dep-dev: require-dep-mockgo-dev

.PHONY: dep-release
dep-release: require-dep-mockgo-release