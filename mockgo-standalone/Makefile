include ../main.mk

MOCKGO_VARIANT = standalone

.PHONY: clean
clean: cleanexe cleancover

.PHONY: build
build: buildexe

.PHONY: run
run: buildexe runexe

.PHONY: mod-dev
mod-dev: require-dep-mockgo-dev