# Main targets for a Go app project
#
# A Self-Documenting Makefile: http://marmelab.com/blog/2016/02/29/auto-documented-makefile.html

OS = $(shell uname | tr A-Z a-z)
ARCH = $(shell uname -m)

# Build variables
BUILD_DIR ?= bin
MAIN_DIR ?= cmd
GIT_TAG ?= $(shell git describe --tags --exact-match 2>/dev/null || git symbolic-ref -q --short HEAD)
GIT_COMMIT_HASH ?= $(shell git rev-parse --short HEAD 2>/dev/null)
DATE_FMT = +%FT%T%z
ifdef SOURCE_DATE_EPOCH
    BUILD_DATE ?= $(shell date -u -d "@$(SOURCE_DATE_EPOCH)" "$(DATE_FMT)" 2>/dev/null || date -u -r "$(SOURCE_DATE_EPOCH)" "$(DATE_FMT)" 2>/dev/null || date -u "$(DATE_FMT)")
else
    BUILD_DATE ?= $(shell date "$(DATE_FMT)")
endif

export CGO_ENABLED ?= 0
ifeq (${VERBOSE}, 1)
ifeq ($(filter -v,${GOARGS}),)
	GOARGS += -v
endif
TEST_FORMAT = short-verbose
endif

IGNORE_GOLANG_VERSION ?= 0
GOLANG_VERSION ?= 1.20.1
DOCKER_VERSION ?= 20.10.7

MOCKGO_VARIANT ?= standalone
MOCKGO_OS= ${OS}
MOCKGO_ARCH ?= amd64
MOCKGO_RELEASE_TAG ?= latest

MOCKGO_IMAGE_REGISTRY ?= docker.io
MOCKGO_IMAGE_REPOSITORY ?= alitari

GOCYCLO_LIMIT ?= 15

GOCOVERMODE ?= atomic

.PHONY: goversion
goversion:
ifneq (${IGNORE_GOLANG_VERSION}, 1)
	@printf "${GOLANG_VERSION}\n$$(go version | awk '{sub(/^go/, "", $$3);print $$3}')" | sort -t '.' -k 1,1 -k 2,2 -k 3,3 -g | head -1 | grep -q -E "^${GOLANG_VERSION}$$" || (printf "Required Go version is ${GOLANG_VERSION}\nInstalled: `go version`" && exit 1)
endif

.PHONY: dockerversion
dockerexists:
	@command -v docker >/dev/null 2>&1 || (printf "Docker is required to build this project" && exit 1)

.PHONY: env
env:

	@echo "--------------------------------------------------"
	@echo "GIT_COMMIT_HASH:    ${GIT_COMMIT_HASH}"
	@echo "GIT_TAG:            ${GIT_TAG}"
	@echo "GOPATH:             ${GOPATH}"
	@echo "BUILD_DIR:          ${BUILD_DIR}"
	@echo "OS:                 ${OS}"
	@echo "GOARGS:             ${GOARGS}"
	@echo "MOCKGO_OS:          ${MOCKGO_OS}"
	@echo "MOCKGO_ARCH:        ${MOCKGO_ARCH}"
	@echo "MOCKGO_RELEASE_TAG: ${MOCKGO_RELEASE_TAG}"
	@echo "CGO_ENABLED:        ${CGO_ENABLED}"
	@echo "MOCKGO_VARIANT:     ${MOCKGO_VARIANT}"
	@echo "MOCKGO_MODULE:      ${MOCKGO_MODULE}"
	@echo "--------------------------------------------------"

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
buildexe: goversion
	@sed -i "s/const versionTag = .*/const versionTag = \"$(MOCKGO_RELEASE_TAG)\"/g" $(MAIN_DIR)/main.go
	CGO_ENABLED=$(CGO_ENABLED) GOOS=$(MOCKGO_OS) GOARCH=$(MOCKGO_ARCH) go build -C $(MAIN_DIR) $(GOARGS) -o ../$(BUILD_DIR)/mockgo-$(MOCKGO_VARIANT)-$(MOCKGO_OS)-$(MOCKGO_ARCH)
	@echo "executable file:"
	@ls -l $(BUILD_DIR)/mockgo-$(MOCKGO_VARIANT)-$(MOCKGO_OS)-$(MOCKGO_ARCH)

.PHONY: buildarchive
buildarchive: goversion
	CGO_ENABLED=$(CGO_ENABLED) GOOS=$(MOCKGO_OS) GOARCH=$(MOCKGO_ARCH) go build --buildmode=archive ./...
	

.PHONY: builddocker
builddocker: dockerexists
	docker build --build-arg RELEASE=$(MOCKGO_RELEASE_TAG)-$(GIT_COMMIT_HASH) -f build/mockgo-$(MOCKGO_VARIANT).Dockerfile . -t $(MOCKGO_IMAGE_REGISTRY)/$(MOCKGO_IMAGE_REPOSITORY)/mockgo-$(MOCKGO_VARIANT):$(MOCKGO_RELEASE_TAG) --no-cache 

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
	gocyclo -ignore '.*\.pb\.go$' -over $(GOCYCLO_LIMIT) .

.PHONY: golint
golint:
	golint -set_exit_status ./...

cover.out: goversion env
	CGO_ENABLED=$(CGO_ENABLED) go test $(GOARGS) -coverprofile=cover-temp.out -covermode=$(GOCOVERMODE) ./...
	@cat cover-temp.out | grep -v ".pb.go" > cover.out
	@rm cover-temp.out

.PHONY: cover
cover: cover.out
	@go tool cover -func=cover.out

.PHONY: cover-html
cover-html: cover.out
	@go tool cover -html=cover.out
