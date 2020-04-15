
# borrowed from https://github.com/technosophos/helm-template

HELM_HOME ?= $(shell helm home)
HELM_PLUGIN_DIR ?= $(HELM_HOME)/plugins/helm-unittest
VERSION := $(shell sed -n -e 's/version:[ "]*\([^"]*\).*/\1/p' plugin.yaml)
DIST := $(CURDIR)/_dist
LDFLAGS := "-X main.version=${VERSION} -extldflags '-static'"
DOCKER ?= "lrills/helm-unittest"

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
	sha256sum $(DIST)/helm-unittest-linux-arm64-$(VERSION).tgz > $(DIST)/helm-unittest-linux-arm64-$(VERSION).tgz.sha
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o untt -ldflags $(LDFLAGS) ./main.go
	tar -zcvf $(DIST)/helm-unittest-linux-amd64-$(VERSION).tgz untt README.md LICENSE plugin.yaml
	sha256sum $(DIST)/helm-unittest-linux-amd64-$(VERSION).tgz > $(DIST)/helm-unittest-linux-amd64-$(VERSION).tgz.sha
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -o untt -ldflags $(LDFLAGS) ./main.go
	tar -zcvf $(DIST)/helm-unittest-macos-amd64-$(VERSION).tgz untt README.md LICENSE plugin.yaml
	sha256sum $(DIST)/helm-unittest-macos-amd64-$(VERSION).tgz > $(DIST)/helm-unittest-macos-amd64-$(VERSION).tgz.sha
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o untt.exe -ldflags $(LDFLAGS) ./main.go
	tar -zcvf $(DIST)/helm-unittest-windows-amd64-$(VERSION).tgz untt.exe README.md LICENSE plugin.yaml
	sha256sum $(DIST)/helm-unittest-windows-amd64-$(VERSION).tgz > $(DIST)/helm-unittest-windows-amd64-$(VERSION).tgz.sha

.PHONY: bootstrap
bootstrap:

dockerimage:
	docker build -t $(DOCKER):$(VERSION) .
