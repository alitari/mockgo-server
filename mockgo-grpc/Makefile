include ../main.mk

MOCKGO_VARIANT = grpc

.PHONY: clean
clean: cleanexe

.PHONY: build
build: buildexe

.PHONY: run
run: buildexe runexe

.PHONY: mod-dev
mod-dev: require-dep-mockgo-dev require-dep-grpc-kvstore-dev require-dep-grpc-matchstore-dev