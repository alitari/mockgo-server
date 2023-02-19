MOCKGO_MODULE ?= mockgo-standalone



LOCAL_REGISTRY_NAME ?= kind-registry
LOCAL_REGISTRY_PORT ?= 5001
LOCAL_REGISTRY_RUNNING = $(shell docker ps -a | grep -q $(LOCAL_REGISTRY_NAME) && echo "true" || echo "false")

KIND_CLUSTER_CONFIG ?= deployments/kind/cluster.yaml
KIND_CLUSTER_RUNNING ?= $(shell kind get clusters -q | grep -q mockgo && echo "true" || echo "false")

CLUSTER_IP ?= 127.0.0.1

.PHONY: env
env:
	$(MAKE) -C $(MOCKGO_MODULE) env

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
builddocker: build
	$(MAKE) -C $(MOCKGO_MODULE) builddocker

pushdocker: builddocker
	$(MAKE) -C $(MOCKGO_MODULE) pushdocker

.PHONY: rundocker
rundocker: builddocker
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

.PHONY: local-registry
local-registry:
ifeq ($(LOCAL_REGISTRY_RUNNING), false)
	docker run -d --restart=always -p "127.0.0.1:$(LOCAL_REGISTRY_PORT):5000" --name "$(LOCAL_REGISTRY_NAME)" registry:2
endif

.PHONY: local-registry-remove
local-registry-remove:
ifeq ($(LOCAL_REGISTRY_RUNNING), true)
	docker rm -f "$(LOCAL_REGISTRY_NAME)"
endif

.PHONY: kind
kind: local-registry
ifeq ($(KIND_CLUSTER_RUNNING), false)
	kind create cluster  --name mockgo --config $(KIND_CLUSTER_CONFIG)
	kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/main/deploy/static/provider/kind/deploy.yaml
	docker network connect "kind" "$(LOCAL_REGISTRY_NAME)"
	kubectl apply -f deployments/kind/local-registry-configmap.yaml
endif

.PHONY: kind-delete
kind-delete: local-registry-remove
ifeq ($(KIND_CLUSTER_RUNNING), true)
	kind delete cluster --name mockgo
endif

.PHONY: helm-deploy
helm-deploy: kind pushdocker
	$(MAKE) -C $(MOCKGO_MODULE) helm-deploy

.PHONY: helm-delete
helm-delete:
	$(MAKE) -C $(MOCKGO_MODULE) helm-delete

.PHONY: tavernbuild
tavernbuild:
	$(MAKE) -C $(MOCKGO_MODULE) tavernbuild

.PHONY: tavern
tavern: tavernbuild
	$(MAKE) -C $(MOCKGO_MODULE) tavern
