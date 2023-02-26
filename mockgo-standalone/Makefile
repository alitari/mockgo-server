MOCKGO_MODULE = mockgo-standalone
MOCKGO_VARIANT = standalone

include ../main.mk

.PHONY: env-module
env-module: env-exe

.PHONY: clean
clean: cleanexe cleancover

.PHONY: build
build: buildexe

.PHONY: run
run: buildexe runexe


.PHONY: mod-dev
mod-dev: require-dep-mockgo-dev