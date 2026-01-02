
# borrowed from https://github.com/technosophos/helm-template

PLUGIN_EMAIL := "helmunittest@gmail.com"
HELM_VERSION := 4.0.4
VERSION := $(shell sed -n -e 's/version:[ "]*\([^"]*\).*/\1/p' plugin.yaml)
BUILD := ./_build
DIST := ./_dist
LDFLAGS := "-X github.com/helm-unittest/helm-unittest/internal/build.version=${VERSION} -extldflags '-static'"
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

.PHONY: build
build: ## Compile packages and dependencies
	go build -o untt -ldflags $(LDFLAGS) ./cmd/helm-unittest

.PHONY: build-amd64
build-amd64: ## Compile packages and dependencies, pinned to amd64 for the docker image
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o untt -ldflags $(LDFLAGS) ./cmd/helm-unittest

.PHONY: dist
dist: ## Build distribution packages, expect to have helm 4 installed.
	mkdir -p $(BUILD)
	mkdir -p $(DIST)

	CGO_ENABLED=0 GOOS=linux GOARCH=ppc64le go build -o untt-linux-ppc64le -ldflags $(LDFLAGS) ./cmd/helm-unittest
	tar -zcvf $(DIST)/helm-unittest-linux-ppc64le-$(VERSION).tgz untt-linux-ppc64le README.md LICENSE plugin.yaml
	mv untt-linux-ppc64le $(BUILD)/
	
	CGO_ENABLED=0 GOOS=linux GOARCH=s390x go build -o untt-linux-s390x -ldflags $(LDFLAGS) ./cmd/helm-unittest
	tar -zcvf $(DIST)/helm-unittest-linux-s390x-$(VERSION).tgz untt-linux-s390x README.md LICENSE plugin.yaml
	mv untt-linux-s390x $(BUILD)/

	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o untt-linux-arm64 -ldflags $(LDFLAGS) ./cmd/helm-unittest
	tar -zcvf $(DIST)/helm-unittest-linux-arm64-$(VERSION).tgz untt-linux-arm64 README.md LICENSE plugin.yaml
	mv untt-linux-arm64 $(BUILD)/

	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o untt-linux-amd64 -ldflags $(LDFLAGS) ./cmd/helm-unittest
	tar -zcvf $(DIST)/helm-unittest-linux-amd64-$(VERSION).tgz untt-linux-amd64 README.md LICENSE plugin.yaml
	mv untt-linux-amd64 $(BUILD)/

	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -o untt-macos-amd64 -ldflags $(LDFLAGS) ./cmd/helm-unittest
	tar -zcvf $(DIST)/helm-unittest-macos-amd64-$(VERSION).tgz untt-macos-amd64 README.md LICENSE plugin.yaml
	mv untt-macos-amd64 $(BUILD)/

	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -o untt-macos-arm64 -ldflags $(LDFLAGS) ./cmd/helm-unittest
	tar -zcvf $(DIST)/helm-unittest-macos-arm64-$(VERSION).tgz untt-macos-arm64 README.md LICENSE plugin.yaml
	mv untt-macos-arm64 $(BUILD)/

	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o untt-windows-amd64.exe -ldflags $(LDFLAGS) ./cmd/helm-unittest
	tar -zcvf $(DIST)/helm-unittest-windows-amd64-$(VERSION).tgz untt-windows-amd64.exe README.md LICENSE plugin.yaml
	mv untt-windows-amd64.exe $(BUILD)/

	cp -f README.md $(BUILD)
	cp -f LICENSE $(BUILD)
	cp -f plugin.yaml $(BUILD)
	cp -f install-binary.sh $(BUILD)
	cp -f install-binary.ps1 $(BUILD)
	chmod +x $(BUILD)/install-binary.ps1
	
	helm plugin package $(BUILD) --key $(PLUGIN_EMAIL) --keyring ./.secring.gpg --passphrase-file ./.passphrase --sign --destination $(DIST)
	rm -f .secring.gpg
	rm -f .passphrase

	shasum -a 256 -b $(DIST)/* > $(DIST)/helm-unittest-checksum.sha

.PHONY: sign
sign-dist: ## Sign distribution packages
	@for f in $$(ls $(DIST)/*.* 2>/dev/null); do \
		echo "signing $$f"; \
		gpg --detach-sign --armor --output $$f.asc $$f; \
	done

.PHONY: bootstrap
bootstrap:

.PHONY: go-dependency
dependency: ## Dependency maintanance
	go get -u ./...
	go mod tidy

.PHONY: dockerimage-alpine
dockerimage-alpine: build-amd64 ## Build docker image
	docker build --no-cache --build-arg HELM_VERSION=$(HELM_VERSION) --build-arg BUILDPLATFORM=amd64 -t $(DOCKER):$(VERSION) -f AlpineTest.Dockerfile .

.PHONY: dockerimage-fedora
dockerimage-fedora: build-amd64 ## Build docker image
	docker build --no-cache --build-arg HELM_VERSION=$(HELM_VERSION) --build-arg BUILDPLATFORM=amd64 -t $(DOCKER):$(VERSION) -f FedoraTest.Dockerfile .

.PHONY: test-docker-alpine
test-docker-alpine: dockerimage-alpine ## Execute 'helm unittests' in container
	@for f in $(TEST_NAMES); do \
		echo "running helm unit tests in folder '$(PROJECT_DIR)/test/data/v3/$${f}'"; \
		docker run \
			--platform linux/amd64 \
			-v $(PROJECT_DIR)/test/data/v3/$${f}:/apps:z \
			--rm  $(DOCKER):$(VERSION) -f tests/*.yaml .;\
	done

.PHONY: test-docker-fedora
test-docker-fedora: dockerimage-fedora ## Execute 'helm unittests' in container
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
