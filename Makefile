MOCKGO_MODULE ?= mockgo-standalone

.PHONY: env
env:
	$(MAKE) -C $(MOCKGO_MODULE) env-global

.PHONY: env-module
env-module:
	$(MAKE) -C $(MOCKGO_MODULE) env-module

.PHONY: clean
clean:
	$(MAKE) -C $(MOCKGO_MODULE) clean

.PHONY: build
build:
	$(MAKE) -C $(MOCKGO_MODULE) build

.PHONY: run
run:
	$(MAKE) -C $(MOCKGO_MODULE) run

.PHONY: builddocker
builddocker:
	$(MAKE) -C $(MOCKGO_MODULE) builddocker

.PHONY: pushdocker
pushdocker:
	$(MAKE) -C $(MOCKGO_MODULE) pushdocker

.PHONY: rundocker
rundocker:
	$(MAKE) -C $(MOCKGO_MODULE) rundocker

.PHONY: cover
cover:
	$(MAKE) -C $(MOCKGO_MODULE) cover

.PHONY: cover-html
cover-html:
	$(MAKE) -C $(MOCKGO_MODULE) cover-html

.PHONY: vulncheck
vulncheck:
	$(MAKE) -C $(MOCKGO_MODULE) vulncheck

.PHONY: helm-deploy
helm-deploy:
	$(MAKE) -C $(MOCKGO_MODULE) helm-deploy

.PHONY: helm-delete
helm-delete:
	$(MAKE) -C $(MOCKGO_MODULE) helm-delete

.PHONY: kind-delete
kind-delete:
	$(MAKE) -C $(MOCKGO_MODULE) kind-delete

.PHONY: clean-acctest
clean-acctest:
	$(MAKE) -C $(MOCKGO_MODULE) clean-acctest

.PHONY: acctest
acctest:
	$(MAKE) -C $(MOCKGO_MODULE) acctest

.PHONY: loadtest
loadtest:
	$(MAKE) -C $(MOCKGO_MODULE) loadtest

.PHONY: dep-dev
dep-dev:
	$(MAKE) -C $(MOCKGO_MODULE) dep-dev

.PHONY: dep-release
dep-release:
	$(MAKE) -C $(MOCKGO_MODULE) dep-release

.PHONY: mod-release
mod-release:
	$(MAKE) -C $(MOCKGO_MODULE) mod-release

.PHONY: tidy
tidy:
	$(MAKE) -C $(MOCKGO_MODULE) mod-tidy