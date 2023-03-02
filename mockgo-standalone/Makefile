MOCKGO_MODULE = mockgo-standalone
MOCKGO_VARIANT = standalone

include ../main.mk

.PHONY: env-module
env-module: env-exe

.PHONY: clean
clean: helm-delete cleanexe cleancover

.PHONY: build
build: buildexe

.PHONY: run
run: buildexe runexe


.PHONY: dep-dev
dep-dev: require-dep-mockgo-dev