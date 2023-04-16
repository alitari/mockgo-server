MAIN_DIR = cmd/mockgo-gen
MOCKGO_MODULE = mockgo-gen
MOCKGO_VARIANT = gen
MOCKGO_HOST = none

include ../main.mk

.PHONY: env-module
env-module: env-exe

.PHONY: clean
clean: cleanexe cleancover

.PHONY: build
build: buildexe buildchecksum

.PHONY: run
run: buildexe runexe

.PHONY: dep-dev
dep-dev: require-dep-mockgo-dev

.PHONY: dep-release
dep-release: require-dep-mockgo-release