
# borrowed from https://github.com/technosophos/helm-template

HELM_HOME ?= $(shell helm home)
HELM_PLUGIN_DIR ?= $(HELM_HOME)/plugins/helm-unittest
VERSION := $(shell sed -n -e 's/version:[ "]*\([^"]*\).*/\1/p' plugin.yaml)
DIST := ./_dist
LDFLAGS := "-X main.version=${VERSION} -extldflags '-static'"
DOCKER ?= "quintush/helm-unittest"

.PHONY: install
install: bootstrap build
	cp untt $(HELM_PLUGIN_DIR)
	cp plugin.yaml $(HELM_PLUGIN_DIR)

.PHONY: hookInstall
hookInstall: bootstrap build

.PHONY: unittest
unittest:
	go test ./unittest/... -v -cover

.PHONY: build
build: unittest
	go build -o untt -ldflags $(LDFLAGS) ./main.go

.PHONY: dist
dist:
	mkdir -p $(DIST)
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o untt -ldflags $(LDFLAGS) ./main.go
	tar -zcvf $(DIST)/helm-unittest-linux-arm64-$(VERSION).tgz untt README.md LICENSE plugin.yaml
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o untt -ldflags $(LDFLAGS) ./main.go
	tar -zcvf $(DIST)/helm-unittest-linux-amd64-$(VERSION).tgz untt README.md LICENSE plugin.yaml
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -o untt -ldflags $(LDFLAGS) ./main.go
	tar -zcvf $(DIST)/helm-unittest-macos-amd64-$(VERSION).tgz untt README.md LICENSE plugin.yaml
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o untt.exe -ldflags $(LDFLAGS) ./main.go
	tar -zcvf $(DIST)/helm-unittest-windows-amd64-$(VERSION).tgz untt.exe README.md LICENSE plugin.yaml
	shasum -a 256 -b $(DIST)/* > $(DIST)/helm-unittest-checksum.sha

.PHONY: bootstrap
bootstrap:

.PHONY: dockerdist
dockerdist:
	./docker-build.sh

.PHONY: dockerimage
dockerimage:
	docker build -t $(DOCKER):$(VERSION) .
