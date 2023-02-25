include ../main.mk

PROTO_CONTEXT = kvstore

.PHONY: clean
clean: cleanexe cleancover cleanproto

.PHONY: build
build: clean gen-proto buildarchive

.PHONY: mod-dev
mod-dev: require-dep-mockgo-dev
