
# borrowed from https://github.com/technosophos/helm-template

HELM_VERSION := 3.17.2
VERSION := $(shell grep -o 'PluginVersion = ".*"' pkg/unittest/version.go | cut -d'"' -f2)
DIST := ./_dist
LDFLAGS := "-X main.version=${VERSION} -extldflags '-static'"
DOCKER ?= helmunittest/helm-unittest
PROJECT_DIR := $(shell pwd)
TEST_NAMES ?=basic \
	failing-template \
	full-snapshot \
	global-double-setting \
	library-chart \
	nested_glob \
	with-document-select \
	with-files \
	with-helm-tests \
	with-k8s-fake-client \
	with-post-renderer \
	with-samenamesubsubcharts \
	with-schema \
	with-subchart \
	with-subfolder \
	with-subsubcharts

.PHONY: help
help:
	@grep -E '^[a-zA-Z0-9_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

.PHONY: plugin-dir
plugin-dir:
	$(eval HELM_3_PLUGINS := $(shell helm env | grep HELM_PLUGINS | cut -d '=' -f 2 | tr -d '"'))
	$(eval HELM_PLUGIN_DIR := $(HELM_3_PLUGINS)/helm-unittest)

.PHONY: install
install: bootstrap build plugin-dir
	mkdir -p $(HELM_PLUGIN_DIR)
	cp untt $(HELM_PLUGIN_DIR)
	cp plugin.yaml $(HELM_PLUGIN_DIR)

.PHONY: install-dbg
install-dbg: bootstrap build-debug plugin-dir
	mkdir -p $(HELM_PLUGIN_DIR)
	cp untt-dbg $(HELM_PLUGIN_DIR)
	cp plugin-dbg.yaml $(HELM_PLUGIN_DIR)/plugin.yaml

.PHONY: hookInstall
hookInstall: bootstrap build

.PHONY: unittest
unittest: ## Run unit tests
	go test ./... -v -cover

.PHONY: test-coverage
test-coverage: build ## Test coverage with open report in default browser
	@go test -cover -coverprofile=cover.out -v ./...
	@go tool cover -html=cover.out

.PHONY: build-debug
build-debug: ## Compile packages and dependencies with debug flag
	go build -o untt-dbg -gcflags "all=-N -l" ./cmd/helm-unittest

.PHONY: plugin-yaml
plugin-yaml: ## Write plugin version to plugin.yaml
	sed "s/__VERSION__/${VERSION}/g" plugin-tpl.yaml > plugin.yaml


.PHONY: build
build: plugin-yaml ## Compile packages and dependencies
	go build -o untt -ldflags $(LDFLAGS) ./cmd/helm-unittest

.PHONY: build-amd64
build-amd64: plugin-yaml ## Compile packages and dependencies, pinned to amd64 for the docker image
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o untt -ldflags $(LDFLAGS) ./cmd/helm-unittest

.PHONY: dist
dist: plugin-yaml
	mkdir -p $(DIST)
	CGO_ENABLED=0 GOOS=linux GOARCH=ppc64le go build -o untt -ldflags $(LDFLAGS) ./cmd/helm-unittest
	tar -zcvf $(DIST)/helm-unittest-linux-ppc64le-$(VERSION).tgz untt README.md LICENSE plugin.yaml
	CGO_ENABLED=0 GOOS=linux GOARCH=s390x go build -o untt -ldflags $(LDFLAGS) ./cmd/helm-unittest
	tar -zcvf $(DIST)/helm-unittest-linux-s390x-$(VERSION).tgz untt README.md LICENSE plugin.yaml
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o untt -ldflags $(LDFLAGS) ./cmd/helm-unittest
	tar -zcvf $(DIST)/helm-unittest-linux-arm64-$(VERSION).tgz untt README.md LICENSE plugin.yaml
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o untt -ldflags $(LDFLAGS) ./cmd/helm-unittest
	tar -zcvf $(DIST)/helm-unittest-linux-amd64-$(VERSION).tgz untt README.md LICENSE plugin.yaml
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -o untt -ldflags $(LDFLAGS) ./cmd/helm-unittest
	tar -zcvf $(DIST)/helm-unittest-macos-amd64-$(VERSION).tgz untt README.md LICENSE plugin.yaml
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -o untt -ldflags $(LDFLAGS) ./cmd/helm-unittest
	tar -zcvf $(DIST)/helm-unittest-macos-arm64-$(VERSION).tgz untt README.md LICENSE plugin.yaml
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o untt.exe -ldflags $(LDFLAGS) ./cmd/helm-unittest
	tar -zcvf $(DIST)/helm-unittest-windows-amd64-$(VERSION).tgz untt.exe README.md LICENSE plugin.yaml
	tar -zcvf $(DIST)/helm-unittest-windows_nt-amd64-$(VERSION).tgz untt.exe README.md LICENSE plugin.yaml
	shasum -a 256 -b $(DIST)/* > $(DIST)/helm-unittest-checksum.sha

.PHONY: bootstrap
bootstrap:

.PHONY: dockerdist
dockerdist:
	./docker-build.sh

.PHONY: go-dependency
dependency: ## Dependency maintanance
	go get -u ./...
	go mod tidy

.PHONY: dockerimage
dockerimage: build-amd64 plugin-yaml ## Build docker image
	docker build --no-cache --build-arg HELM_VERSION=$(HELM_VERSION) --build-arg BUILDPLATFORM=amd64 -t $(DOCKER):$(VERSION) -f AlpineTest.Dockerfile .

.PHONY: test-docker
test-docker: dockerimage ## Execute 'helm unittests' in container
	@for f in $(TEST_NAMES); do \
		echo "running helm unit tests in folder '$(PROJECT_DIR)/test/data/v3/$${f}'"; \
		docker run \
			--platform linux/amd64 \
			-v $(PROJECT_DIR)/test/data/v3/$${f}:/apps:z \
			--rm  $(DOCKER):$(VERSION) -f tests/*.yaml .;\
	done

.PHONY: go-lint
go-lint: ## Execute golang linters
	gofmt -l -s -w .
	golangci-lint run --timeout=30m --fix ./...
