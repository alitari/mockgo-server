MOCKGO_MODULE ?= mockgo-standalone

.PHONY: clean
clean:
	$(MAKE) -C $(MOCKGO_MODULE) clean

.PHONY: build
build:
	$(MAKE) -C $(MOCKGO_MODULE) build

.PHONY: buildimage
buildimage:
	$(MAKE) -C $(MOCKGO_MODULE) buildimage

.PHONY: misspell
misspell:
	misspell -error .

.PHONY: gofmt
gofmt:
	$(MAKE) -C $(MOCKGO_MODULE) gofmt

.PHONY: vet
vet:
	$(MAKE) -C $(MOCKGO_MODULE) vet

.PHONY: ineffassign
ineffassign:
	$(MAKE) -C $(MOCKGO_MODULE) ineffassign

.PHONY: cover
cover:
	$(MAKE) -C $(MOCKGO_MODULE) cover

.PHONY: cover-html
cover-html:
	$(MAKE) -C $(MOCKGO_MODULE) cover-html
