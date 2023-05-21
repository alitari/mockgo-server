# Main targets for a Go app project

ccred := $(shell echo -e "\033[0;31m")
ccyellow := $(shell echo -e "\033[0;33m")
ccend := $(shell echo -e "\033[0m")

OS = $(shell uname | tr A-Z a-z)
ARCH = $(shell uname -m)

# golang
GO_VERSION = $(shell go version | awk '{sub(/^go/, "", $$3);print $$3}')
ifeq ($(GO_VERSION),)
GO_VERSION = $(ccred)"not installed$(ccend)"
endif
CGO_ENABLED ?= 0
MAIN_DIR ?= cmd/mockgo
BUILD_DIR = cmd/bin

GIT_TAG ?= $(shell git fetch --all --tags && git tag | tail -1)
GIT_COMMIT_HASH ?= $(shell git rev-parse --short HEAD 2>/dev/null)

MODULE_DIR ?= $(shell pwd)
PROJECT_DIR ?= "$(shell dirname "${MODULE_DIR}")"

MOCKGO_OS = ${OS}
MOCKGO_ARCH ?= amd64
MOCKGO_RELEASE ?= latest
ifeq ($(MOCKGO_RELEASE),)
MOCKGO_RELEASE = latest
endif
ifeq ($(MOCKGO_RELEASE),latest)
IS_REAL_RELEASE = "false"
else
IS_REAL_RELEASE = "true"
endif


GOCOVERMODE ?= atomic
GOCYCLO_LIMIT ?= 15

DOCKER_VERSION = $(shell docker version --format '{{.Server.Version}}')
ifeq ($(DOCKER_VERSION),)
DOCKER_VERSION = $(ccred)"not installed$(ccend)"
endif
MOCKGO_IMAGE_REGISTRY ?= localhost:5001
MOCKGO_IMAGE_REPO ?= alitari


KUBECTL_VERSION = $(shell kubectl version --client=true --short=true 2>/dev/null | awk '{print $$3}')
ifeq ($(KUBECTL_VERSION),)
KUBECTL_VERSION = $(ccred)"not installed$(ccend)"
endif

HELM_VERSION = $(shell helm version --short)
ifeq ($(HELM_VERSION),)
HELM_VERSION = $(ccred)"not installed$(ccend)"
endif
HELM_DEPLOY_TIMEOUT ?= 300s
HELM_DEPLOY_WAIT ?= 60s

KIND_VERSION ?= $(shell kind version | awk '{print $$2}')
ifeq ($(KIND_VERSION),)
KIND_VERSION = $(ccred)"not installed$(ccend)"
endif
KIND_CLUSTER_CONFIG ?= $(PROJECT_DIR)/deployments/kind/cluster.yaml
KIND_CLUSTER_RUNNING ?= $(shell kind get clusters -q | grep -q mockgo && echo "true" || echo "false")
KIND_CLUSTER_WAIT ?= 60s

LOCAL_REGISTRY_NAME ?= kind-registry
LOCAL_REGISTRY_PORT ?= 5001
LOCAL_REGISTRY_RUNNING = $(shell docker ps -a | grep -q $(LOCAL_REGISTRY_NAME) && echo "true" || echo "false")

HURL_VERSION ?= $(shell hurl --version 2> /dev/null | head -1 | awk '{print $$2}')
ifeq ($(HURL_VERSION),)
HURL_VERSION = $(ccred)"not installed$(ccend)"
endif


MOCKGO_DEPLOYED ?= $(shell helm --namespace mockgo list -q 2> /dev/null | grep -q mockgo-${MOCKGO_VARIANT} && echo "true" || echo "false")

CLUSTER_IP ?= 127.0.0.1
MOCKGO_HOST ?= mockgo-$(MOCKGO_VARIANT).$(CLUSTER_IP).nip.io

.PHONY: env-global
env-global:
	@echo "------------------ workstation --------------------"
	@echo "OS-ARCH:               ${OS}-${ARCH}"
	@echo "------------------- golang ------------------------"
	@echo "GO_VERSION:            ${GO_VERSION}"
	@echo "CGO_ENABLED:           ${CGO_ENABLED}"
	@echo "MAIN_DIR:              ${MAIN_DIR}"
	@echo "BUILD_DIR:             ${BUILD_DIR}"
	@echo "------------------- git ---------------------------"
	@echo "GIT_TAG:               ${GIT_TAG}"
	@echo "GIT_COMMIT_HASH:       ${GIT_COMMIT_HASH}"
	@echo "------------------- project -----------------------"
	@echo "PROJECT_DIR:           ${PROJECT_DIR}"
	@echo "MOCKGO_OS:             ${MOCKGO_OS}"
	@echo "MOCKGO_ARCH:           ${MOCKGO_ARCH}"
	@echo "MOCKGO_RELEASE:        ${MOCKGO_RELEASE}"
	@echo "IS_REAL_RELEASE:       ${IS_REAL_RELEASE}"
	@echo "------------------- unit test --------------------------"
	@echo "GOCOVERMODE:           ${GOCOVERMODE}"
	@echo "GOCYCLO_LIMIT:         ${GOCYCLO_LIMIT}"
	@echo "------------------- docker ------------------------"
	@echo "DOCKER_VERSION:        ${DOCKER_VERSION}"
	@echo "MOCKGO_IMAGE_REGISTRY: ${MOCKGO_IMAGE_REGISTRY}"
	@echo "MOCKGO_IMAGE_REPO:     ${MOCKGO_IMAGE_REPO}"
	@echo "--------------------- k8s -------------------------"
	@echo "KUBECTL_VERSION:       ${KUBECTL_VERSION}"
	@echo "LOCAL_REGISTRY:        ${LOCAL_REGISTRY_NAME}:${LOCAL_REGISTRY_PORT}"
	@echo "LOCAL_REGISTRY_RUNNING:${LOCAL_REGISTRY_RUNNING}"
	@echo "KIND_VERSION:          ${KIND_VERSION}"
	@echo "KIND_CLUSTER_CONFIG:   ${KIND_CLUSTER_CONFIG}"
	@echo "KIND_CLUSTER_RUNNING:  ${KIND_CLUSTER_RUNNING}"
	@echo "KIND_CLUSTER_WAIT:     ${KIND_CLUSTER_WAIT}"
	@echo "HELM_VERSION:          ${HELM_VERSION}"
	@echo "HELM_DEPLOY_TIMEOUT:   ${HELM_DEPLOY_TIMEOUT}"
	@echo "HELM_DEPLOY_WAIT:      ${HELM_DEPLOY_WAIT}"
	@echo "---------------- acceptance test ------------------"
	@echo "HURL_VERSION:          ${HURL_VERSION}"
	@echo "---------------------------------------------------"

.PHONY: env-exe
env-exe:
	@echo "------- MOCKGO MODULE: $(MOCKGO_MODULE) -----------"
	@echo "MOCKGO_VARIANT:        ${MOCKGO_VARIANT}"
	@echo "MOCKGO_DEPLOYED:       ${MOCKGO_DEPLOYED}"
	@echo "MOCKGO_HOST:           ${MOCKGO_HOST}"

.PHONY: env-lib
env-lib:
	@echo "------- MOCKGO MODULE: $(MOCKGO_MODULE) -----------"
	@echo "PROTO_CONTEXT:         ${PROTO_CONTEXT}"

.PHONY: cleanexe
cleanexe:
	rm -rf $(BUILD_DIR)/

.PHONY: cleancover
cleancover:
	rm -f cover.out

cleanproto: 
	rm -f $(PROTO_CONTEXT)/*.pb.go

.PHONY: gen-proto
gen-proto:
	protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative $(PROTO_CONTEXT)/$(PROTO_CONTEXT).proto
	@echo "generated files:"
	@ls -l ./$(PROTO_CONTEXT)/*.pb.go


.PHONY: buildexe
buildexe:
	@sed -i "s/const versionTag = .*/const versionTag = \"$(MOCKGO_RELEASE)\"/g" $(MAIN_DIR)/main.go
	CGO_ENABLED=$(CGO_ENABLED) GOOS=linux GOARCH=amd64 go build $(GOARGS) -o $(BUILD_DIR)/mockgo-$(MOCKGO_VARIANT)-linux-amd64 $(MAIN_DIR)/main.go
	CGO_ENABLED=$(CGO_ENABLED) GOOS=linux GOARCH=arm64 go build -o $(BUILD_DIR)/mockgo-$(MOCKGO_VARIANT)-linux-arm64 $(MAIN_DIR)/main.go
	CGO_ENABLED=$(CGO_ENABLED) GOOS=windows GOARCH=amd64 go build -o $(BUILD_DIR)/mockgo-$(MOCKGO_VARIANT)-windows-amd64 $(MAIN_DIR)/main.go
	sha256sum $(BUILD_DIR)/mockgo-$(MOCKGO_VARIANT)-*
	@echo "files:"
	@ls -l $(BUILD_DIR)/mockgo-$(MOCKGO_VARIANT)-*

.PHONY: buildchecksum
buildchecksum:
	sha256sum $(BUILD_DIR)/mockgo-$(MOCKGO_VARIANT)-linux-amd64 > $(BUILD_DIR)/mockgo-$(MOCKGO_VARIANT)-linux-amd64.sha256
	sha256sum $(BUILD_DIR)/mockgo-$(MOCKGO_VARIANT)-linux-arm64 > $(BUILD_DIR)/mockgo-$(MOCKGO_VARIANT)-linux-arm64.sha256
	sha256sum $(BUILD_DIR)/mockgo-$(MOCKGO_VARIANT)-windows-amd64 > $(BUILD_DIR)/mockgo-$(MOCKGO_VARIANT)-linux-windows.sha256

.PHONY: runexe
runexe: buildexe
	./$(BUILD_DIR)/mockgo-$(MOCKGO_VARIANT)-$(MOCKGO_OS)-$(MOCKGO_ARCH) $(RUN_OPTIONS)

.PHONY: buildarchive
buildarchive:
	CGO_ENABLED=$(CGO_ENABLED) GOOS=$(MOCKGO_OS) GOARCH=$(MOCKGO_ARCH) go build --buildmode=archive ./...

.PHONY: builddocker
builddocker: buildexe buildchecksum
	docker build --build-arg RELEASE=$(MOCKGO_RELEASE)-$(GIT_COMMIT_HASH) -f build/mockgo-$(MOCKGO_VARIANT).Dockerfile . -t $(MOCKGO_IMAGE_REGISTRY)/$(MOCKGO_IMAGE_REPO)/mockgo-$(MOCKGO_VARIANT):$(MOCKGO_RELEASE) $(DOCKER_BUILD_OPTIONS)

.PHONY: pushdocker
pushdocker: builddocker
	docker push $(MOCKGO_IMAGE_REGISTRY)/$(MOCKGO_IMAGE_REPO)/mockgo-$(MOCKGO_VARIANT):$(MOCKGO_RELEASE)
	docker tag $(MOCKGO_IMAGE_REGISTRY)/$(MOCKGO_IMAGE_REPO)/mockgo-$(MOCKGO_VARIANT):$(MOCKGO_RELEASE) $(MOCKGO_IMAGE_REGISTRY)/$(MOCKGO_IMAGE_REPO)/mockgo-$(MOCKGO_VARIANT):latest
	docker push $(MOCKGO_IMAGE_REGISTRY)/$(MOCKGO_IMAGE_REPO)/mockgo-$(MOCKGO_VARIANT):latest

.PHONY: rundocker
rundocker: builddocker
	docker run $(MOCKGO_IMAGE_REGISTRY)/$(MOCKGO_IMAGE_REPO)/mockgo-$(MOCKGO_VARIANT):$(MOCKGO_RELEASE) $(DOCKER_RUN_OPTIONS)

.PHONY: gofmt
gofmt:
	gofmt -s -w .

.PHONY: vet
vet:
	go vet ./...

.PHONY: ineffassign
ineffassign:
	ineffassign ./...

.PHONY: gocyclo
gocyclo:
	gocyclo -ignore ".*\\.pb\\.go$$" -over $(GOCYCLO_LIMIT) .

.PHONY: golint
golint:
	golint -set_exit_status ./...

cover.out: gofmt vet ineffassign gocyclo golint
	go clean -testcache
	CGO_ENABLED=$(CGO_ENABLED) go test $(GOARGS) -coverprofile=cover-temp.out -covermode=$(GOCOVERMODE) ./...
	@cat cover-temp.out | grep -v ".pb.go" > cover.out
	@rm cover-temp.out

.PHONY: cover
cover: cover.out
	@go tool cover -func=cover.out

.PHONY: cover-html
cover-html: cover.out
	@go tool cover -html=cover.out

.PHONY: vulncheck
vulncheck:
	govulncheck ./...

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
	sleep $(KIND_CLUSTER_WAIT)
	docker network connect "kind" "$(LOCAL_REGISTRY_NAME)"
	kubectl apply -f $(PROJECT_DIR)/deployments/kind/local-registry-configmap.yaml
endif

.PHONY: kind-delete
kind-delete: local-registry-remove
ifeq ($(KIND_CLUSTER_RUNNING), true)
	kind delete cluster --name mockgo
endif

.PHONY: helm-deploy
helm-deploy: kind pushdocker
	helm dependency build $(PROJECT_DIR)/deployments/helm/mockgo-server
	helm upgrade --install mockgo-$(MOCKGO_VARIANT) $(PROJECT_DIR)/deployments/helm/mockgo-server \
	--namespace mockgo --create-namespace -f $(PROJECT_DIR)/deployments/helm/$(MOCKGO_VARIANT)-values.yaml \
	--wait --timeout $(HELM_DEPLOY_TIMEOUT) --atomic
	sleep $(HELM_DEPLOY_WAIT)

.PHONY: helm-delete
helm-delete:
ifeq ($(MOCKGO_DEPLOYED), true)
	helm delete mockgo-$(MOCKGO_VARIANT) --namespace mockgo
endif

.PHONY: clean-acctest
clean-acctest:
	rm -rf $(PROJECT_DIR)/reports/hurl/mockgo-$(MOCKGO_VARIANT)

.PHONY: acctest
acctest: helm-deploy
	mkdir -p $(PROJECT_DIR)/reports/hurl/mockgo-$(MOCKGO_VARIANT)
	hurl $(PROJECT_DIR)/test/acceptance-test/hello.hurl $(PROJECT_DIR)/test/acceptance-test/matches.hurl --variable mockgo_host=$(MOCKGO_HOST) --test --report-html $(PROJECT_DIR)/reports/hurl/mockgo-$(MOCKGO_VARIANT)

.PHONY: loadtest
loadtest: helm-deploy
	./scripts/load-test.sh 100 10 2 mockgo-$(MOCKGO_VARIANT).127.0.0.1.nip.io GET /hello1 hello1 

.PHONY: drop-dep-mockgo
drop-dep-mockgo:
	go mod edit -droprequire github.com/alitari/mockgo-server/mockgo
	go mod edit -dropreplace github.com/alitari/mockgo-server/mockgo

.PHONY: require-dep-mockgo-dev
require-dep-mockgo-dev: drop-dep-mockgo
	go mod edit -replace=github.com/alitari/mockgo-server/mockgo=../mockgo

.PHONY: require-dep-mockgo-release
require-dep-mockgo-release: drop-dep-mockgo
	go mod edit -require=github.com/alitari/mockgo-server/mockgo@$(MOCKGO_RELEASE)


.PHONY: mod-release
mod-release:
ifeq ($(IS_REAL_RELEASE), "true")
	git add -A
	git commit -m "🔖 Setup dependencies for module $(MOCKGO_MODULE) release-$(MOCKGO_RELEASE)" --allow-empty
	git push --set-upstream origin "release-$(MOCKGO_RELEASE)"
	@echo "tagging mockgo module with $(MOCKGO_MODULE)/$(MOCKGO_RELEASE) ..."
	git tag -a "$(MOCKGO_MODULE)/$(MOCKGO_RELEASE)" -m "🔖 Tag mockgo module with $(MOCKGO_MODULE)/$(MOCKGO_RELEASE)"
	git push origin "$(MOCKGO_MODULE)/$(MOCKGO_RELEASE)"
	GOPROXY=proxy.golang.org go list -m "github.com/alitari/mockgo-server/$(MOCKGO_MODULE)@$(MOCKGO_RELEASE)"
else
	@echo "not a real release, you must set MOCKGO_RELEASE to a real release version with semantic versioning"
endif

.PHONY: mod-tidy
mod-tidy:
	go mod tidy
